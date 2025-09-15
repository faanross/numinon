package router

import (
	"encoding/json"
	"fmt"
	"github.com/faanross/numinon/internal/agent/config"
	"github.com/faanross/numinon/internal/models"
	"github.com/faanross/numinon/internal/taskbroker"
	"github.com/faanross/numinon/internal/tracker"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// WSPusher allows tasks to be pushed immediately from broker (operator layer) to intended agent
var WSPusher *taskbroker.WebSocketTaskPusher

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

	// Register with task pusher
	if WSPusher != nil {
		WSPusher.RegisterConnection(agentID, conn)
		defer WSPusher.UnregisterConnection(agentID) // Clean up on disconnect
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
