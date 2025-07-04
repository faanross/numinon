package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"

	"github.com/gorilla/websocket"
)

func WSHandler(w http.ResponseWriter, r *http.Request) {

	// Upgrade, extract Agent ID, determine protocol
	conn, agentID, agentProtocol, err := _wsUpgradeAndRegister(w, r)
	if err != nil {
		log.Printf("|❗ERR WS HANDLER| Connection setup failed for %s: %v", r.RemoteAddr, err)
		return
	}

	// Clean-up connection once closed
	defer func() {
		log.Printf("|🔌 WS HANDLER| Closing WebSocket connection for AgentID: %s (RemoteAddr: %s, Proto: %s).",
			agentID, conn.RemoteAddr(), agentProtocol)
		conn.Close()
	}()

	// This is the classic WS pattern - an infinite for loop
	// This will be called from both client and server side
	// This "locks" them into the bidirectional push state

	for {
		messageType, messageBytes, readErr := conn.ReadMessage()

		if readErr != nil {
			if websocket.IsUnexpectedCloseError(readErr, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("|❗ERR WS HANDLER| AgentID: %s - WebSocket unexpected close error: %v", agentID, readErr)
			} else if e, ok := readErr.(*websocket.CloseError); ok && (e.Code == websocket.CloseNormalClosure || e.Code == websocket.CloseGoingAway) {
				log.Printf("|🔌 WS HANDLER| AgentID: %s - WebSocket closed gracefully by peer or self: %v", agentID, readErr)
			} else {
				log.Printf("|❗ERR WS HANDLER| AgentID: %s - WebSocket read error: %v", agentID, readErr)
			}
			break
		}

		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			if procErr := _wsProcessIncomingMessage(conn, agentID, agentProtocol, messageBytes); procErr != nil {
				log.Printf("|❗ERR WS HANDLER| AgentID: %s - Error processing incoming message: %v. Continuing read loop.", agentID, procErr)
			}
		} else {
			log.Printf("|ℹ️ WS HANDLER| AgentID: %s - Received WebSocket control/unhandled message type: %d.", agentID, messageType)
		}
	}
}

// _wsUpgradeAndRegister handles WebSocket upgrade, AgentID extraction, and protocol determination,
func _wsUpgradeAndRegister(w http.ResponseWriter, r *http.Request) (
	conn *websocket.Conn,
	agentID string,
	agentProtocol config.AgentProtocol,
	err error,
) {
	// (1) UPGRADE FROM HTTP/1.1 to WS/S
	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("|_DBG WS HANDLER| Upgrade failed for %s: %v", r.RemoteAddr, err)
		return nil, "", "", err
	}

	// (2) EXTRACT THE AGENT ID
	agentID = r.Header.Get("Agent-ID")

	// (3) DETERMINE PROTOCOL
	// The presence/absence of TLS will determine whether we treat this as WS or WSS

	if r.TLS == nil {
		agentProtocol = config.WebsocketClear
	} else {
		agentProtocol = config.WebsocketSecure
	}

	log.Printf("|_DBG WS HANDLER| AgentID: %s (Proto: %s) WebSocket connection established from %s.", agentID, agentProtocol, conn.RemoteAddr())

	return conn, agentID, agentProtocol, nil
}

// _wsProcessIncomingMessage handles a single data message (Text or Binary) received from the agent.
func _wsProcessIncomingMessage(conn *websocket.Conn, agentID string, agentProtocol config.AgentProtocol, messageBytes []byte) error {
	log.Printf("|_DBG WS HANDLER| AgentID: %s - Processing message (%d bytes)", agentID, len(messageBytes))

	// 1. Unmarshal the raw JSON into our AgentTaskResult struct
	var result models.AgentTaskResult
	if err := json.Unmarshal(messageBytes, &result); err != nil {
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

	return nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Right now we are indiscriminately upgrading WS connection
	// Later, we def should discriminate!
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
