package taskbroker

import (
	"github.com/faanross/numinon/internal/tracker"
	"log"
)

// determineDeliveryMethod decides how to deliver a task based on agent connection.
// This is our "smart routing" logic.
func (b *Broker) determineDeliveryMethod(agentID string) string {
	if b.agentTracker == nil {
		return "queue" // Default to queuing if we can't determine
	}

	agentInfo, exists := b.agentTracker.GetAgentInfo(agentID)
	if !exists {
		log.Printf("[TASK BROKER] Agent %s not found in tracker, defaulting to queue", agentID)
		return "queue"
	}

	// Check connection type and state
	if agentInfo.ConnectionType == tracker.TypeWebSocket &&
		agentInfo.State == tracker.StateConnected {
		// Check if we actually have the WebSocket connection
		if b.wsPusher != nil && b.wsPusher.IsAgentConnected(agentID) {
			return "push"
		}
	}

	// For HTTP agents or disconnected WS agents, queue the task
	return "queue"
}
