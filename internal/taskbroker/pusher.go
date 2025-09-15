package taskbroker

import (
	"encoding/json"
	"fmt"
	"github.com/faanross/numinon/internal/models"
	"github.com/faanross/numinon/internal/taskmanager"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

// WebSocketTaskPusher manages WebSocket connections and pushes tasks to them.
// Think of this as the "instant notification service" for connected agents.
//
// When an operator creates a task for a WS-connected agent, this component
// immediately pushes it rather than waiting for the agent to ask.
type WebSocketTaskPusher struct {
	mu          sync.RWMutex
	connections map[string]*websocket.Conn // agentID -> WebSocket connection
}

// NewWebSocketTaskPusher creates a new task pusher.
func NewWebSocketTaskPusher() *WebSocketTaskPusher {
	return &WebSocketTaskPusher{
		connections: make(map[string]*websocket.Conn),
	}
}

// RegisterConnection stores a WebSocket connection for an agent.
// Called when a WebSocket agent connects.
func (p *WebSocketTaskPusher) RegisterConnection(agentID string, conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close any existing connection for this agent
	if existingConn, exists := p.connections[agentID]; exists {
		log.Printf("[WS PUSHER] Closing existing connection for agent %s", agentID)
		existingConn.Close()
	}

	p.connections[agentID] = conn
	log.Printf("[WS PUSHER] Registered WebSocket connection for agent %s", agentID)
}

// UnregisterConnection removes a WebSocket connection.
// Called when a WebSocket agent disconnects.
func (p *WebSocketTaskPusher) UnregisterConnection(agentID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.connections, agentID)
	log.Printf("[WS PUSHER] Unregistered WebSocket connection for agent %s", agentID)
}

// IsAgentConnected checks if an agent has an active WebSocket connection.
func (p *WebSocketTaskPusher) IsAgentConnected(agentID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.connections[agentID]
	return exists
}

// PushTask immediately sends a task to a WebSocket-connected agent.
// Returns error if agent is not connected or push fails.
func (p *WebSocketTaskPusher) PushTask(agentID string, task *taskmanager.Task) error {
	p.mu.RLock()
	conn, exists := p.connections[agentID]
	p.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s is not WebSocket-connected", agentID)
	}

	// Prepare the task response (same format as HTTP check-in)
	response := models.ServerTaskResponse{
		TaskAvailable: true,
		TaskID:        task.ID,
		Command:       task.Command,
		Data:          task.Arguments,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Send via WebSocket
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		// Connection might be broken, unregister it
		p.UnregisterConnection(agentID)
		return fmt.Errorf("failed to send task via WebSocket: %w", err)
	}

	log.Printf("[WS PUSHER] Successfully pushed task %s to agent %s via WebSocket",
		task.ID, agentID)

	return nil
}

// GetConnectedAgents returns a list of all WebSocket-connected agent IDs.
func (p *WebSocketTaskPusher) GetConnectedAgents() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := make([]string, 0, len(p.connections))
	for agentID := range p.connections {
		agents = append(agents, agentID)
	}

	return agents
}
