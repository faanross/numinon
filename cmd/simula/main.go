package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
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
	playbookPath = flag.String("playbook", "internal/simula/playbooks/example.yaml", "Path to the simulation playbook YAML file")
	logFilePath  = flag.String("log-file", "simulation_run.log", "Optional: Path to save the simulation log file")
	agentID      = flag.String("agent", "your-default-agent-id-here", "Target Agent ID for the simulation")
	serverURL    = flag.String("server", "ws://localhost:8080/client", "C2 server WebSocket API URL")
)

func main() {
	flag.Parse()

	// Configure logging to go to both console and a file
	if *logFilePath != "" {
		logFile, err := os.OpenFile(*logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer logFile.Close()
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	}

	// Load the playbook
	simPlan, err := loadPlaybook(*playbookPath)
	if err != nil {
		log.Fatalf("Failed to load playbook: %v", err)
	}

	// Connect to the C2 server
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	conn, err := connectToC2()
	if err != nil {
		log.Fatalf("Failed to connect to C2 server: %v", err)
	}
	defer conn.Close()

	// Run the simulation
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
		}

		// The Action Dispatcher
		switch step.Action {
		case "sleep":
			d, err := time.ParseDuration(step.Duration)
			if err != nil {
				log.Printf("|❗ERR| Invalid duration '%s': %v. Skipping sleep.", step.Duration, err)
				continue
			}
			log.Printf("Action: sleep for %v...", d)
			time.Sleep(d)
		case "run_cmd", "upload", "download", "enumerate", "shellcode", "morph", "hop":
			handleTaskingAction(conn, step)
		default:
			log.Printf("|⚠️ WARN| Unknown action '%s' in playbook. Skipping.", step.Action)
		}
	}
}

// handleTaskingAction is the "brain" of the simulator. It resolves the action
// (either from a category or explicit args) and sends the request.
// **Modified File: punkin_instigator/tools/simulator/main.go**

// (This is the updated handleTaskingAction function. The rest of the file remains the same.)

// handleTaskingAction is the "brain" of the simulator. It resolves the action
// (either from a category or explicit args) and sends the request.
func handleTaskingAction(conn *websocket.Conn, step plan.SimulationStep) {
	var specificArgs interface{}
	var apiAction string

	// --- Resolve Action Arguments (Category vs. Explicit) ---
	log.Printf("Action: %s (Category: '%s')", step.Action, step.Category)
	switch step.Action {
	case "run_cmd":
		apiAction = string(clientapi.ActionTaskAgentRunCmd)
		var cmd string
		// --- CORRECTED LOGIC ---
		// Use the new, more granular categories we defined.
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
		// --- END CORRECTION ---

		if cmd == "" && step.Args == nil {
			log.Printf("|❗ERR| No valid category or explicit args for run_cmd step '%s'. Skipping.", step.Name)
			return
		}

		if cmd != "" {
			log.Printf("Selected random command from category '%s': %s", step.Category, cmd)
			specificArgs = models.RunCmdArgs{CommandLine: cmd}
		} else {
			// Fallback to explicit args if no category is matched
			var args models.RunCmdArgs
			json.Unmarshal(step.Args, &args)
			specificArgs = args
			log.Printf("Using explicit args for run_cmd: '%s'", args.CommandLine)
		}

	case "upload":
		apiAction = string(clientapi.ActionTaskAgentUploadFile)
		if step.Category == "tooling" {
			preset := scenarios.ToolingUploads[rand.Intn(len(scenarios.ToolingUploads))]
			sourcePath := filepath.Join("tools", "simulator", "dummy_files", preset.DummyFileName)
			destDir := scenarios.GetRandom(preset.PlausibleTargetDirs)
			log.Printf("Selected upload preset: '%s' to '%s'", preset.DummyFileName, destDir)

			content, err := ioutil.ReadFile(sourcePath)
			if err != nil {
				log.Printf("|❗ERR| Failed to read dummy file %s: %v", sourcePath, err)
				return
			}
			hasher := sha256.New()
			hasher.Write(content)
			specificArgs = models.UploadArgs{
				TargetDirectory: destDir, TargetFilename: preset.DummyFileName,
				FileContentBase64: base64.StdEncoding.EncodeToString(content), ExpectedSha256: hex.EncodeToString(hasher.Sum(nil)),
			}
		} else {
			var args models.UploadArgs
			json.Unmarshal(step.Args, &args)
			specificArgs = args
		}

	case "download":
		apiAction = string(clientapi.ActionTaskAgentDownloadFile)
		var path string
		if step.Category == "system_files" {
			path = scenarios.GetRandom(scenarios.ExfilSystemFiles)
		}
		if step.Category == "user_files" {
			path = scenarios.GetRandom(scenarios.ExfilUserFiles)
		}

		if path == "" && step.Args == nil {
			log.Printf("|❗ERR| No valid category or explicit args for download step '%s'. Skipping.", step.Name)
			return
		}

		if path != "" {
			log.Printf("Selected download path from category '%s': '%s'", step.Category, path)
			specificArgs = models.DownloadArgs{SourceFilePath: path}
		} else {
			var args models.DownloadArgs
			json.Unmarshal(step.Args, &args)
			specificArgs = args
			log.Printf("Using explicit args for download: '%s'", args.SourceFilePath)
		}

	case "enumerate":
		apiAction = string(clientapi.ActionTaskAgentEnumerateProcs)
		var procName string
		if step.Category == "security_products" {
			procName = scenarios.GetRandom(scenarios.SecurityProductProcesses)
		}
		if step.Category == "remote_access" {
			procName = scenarios.GetRandom(scenarios.RemoteAccessProcesses)
		}

		if procName == "" && step.Args == nil {
			log.Printf("No category matched for enumerate step '%s'. Assuming enumeration of all processes.", step.Name)
			specificArgs = models.EnumerateArgs{ProcessName: ""}
		} else if procName != "" {
			log.Printf("Selected process to enumerate from category '%s': '%s'", step.Category, procName)
			specificArgs = models.EnumerateArgs{ProcessName: procName}
		} else {
			var args models.EnumerateArgs
			json.Unmarshal(step.Args, &args)
			specificArgs = args
			log.Printf("Using explicit args for enumerate: '%s'", args.ProcessName)
		}

	case "morph", "hop", "shellcode": // These actions require explicit args
		if step.Action == "morph" {
			apiAction = string(clientapi.ActionTaskAgentMorph)
		}
		if step.Action == "hop" {
			apiAction = string(clientapi.ActionTaskAgentHop)
		}
		if step.Action == "shellcode" {
			apiAction = string(clientapi.ActionTaskAgentExecuteShellcode)
		}
		log.Printf("Using explicit args from playbook for '%s' action.", step.Action)
		json.Unmarshal(step.Args, &specificArgs)
	}

	if specificArgs == nil {
		log.Printf("|❗ERR| Could not determine arguments for step '%s'. Skipping.", step.Name)
		return
	}

	// --- Build and Send Request (same as before) ---
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

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Printf("|⚠️ WARN| Did not receive confirmation for task: %v", err)
	} else {
		log.Printf("Received server confirmation.")
		_ = msg
	}
}
