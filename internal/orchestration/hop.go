package orchestration

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/listener"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/taskmanager"
	"numinon_shadow/internal/tracker"
	"strings"
	"time"
)

// HopOrchestrator handles the Hop command lifecycle
type HopOrchestrator struct {
	listenerManager *listener.Manager
	agentTracker    *tracker.Tracker
}

// NewHopOrchestrator creates a new Hop orchestrator with listener management capability
func NewHopOrchestrator(listenerManager *listener.Manager, agentTracker *tracker.Tracker) *HopOrchestrator {
	return &HopOrchestrator{
		listenerManager: listenerManager,
		agentTracker:    agentTracker,
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

	// Mark agent as initiating hop
	if h.agentTracker != nil {
		if err := h.agentTracker.InitiateHop(task.AgentID, listenerID); err != nil {
			log.Printf("|⚠️ HOP| Failed to mark hop initiation: %v", err)
			// Continue anyway - tracking is not critical
		}
	}

	// Store both listener IDs for cleanup
	if agentInfo, exists := h.agentTracker.GetAgentInfo(task.AgentID); exists {
		task.Metadata["old_listener_id"] = agentInfo.ListenerID
	}
	task.Metadata["new_listener_id"] = listenerID

	log.Printf("|🐇 HOP PREP| Created listener %s for agent %s hop to %s:%s",
		listenerID, task.AgentID, hopArgs.NewServerIP, hopArgs.NewServerPort)

	return nil
}

func (h *HopOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {
	var agentResult models.AgentTaskResult
	if err := json.Unmarshal(result, &agentResult); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}

	if agentResult.Status != models.StatusSuccessHopInitiated {
		// Hop failed - clean up new listener
		if newListenerID, ok := task.Metadata["new_listener_id"].(string); ok {
			// Try to stop immediately since it was never used
			if err := h.listenerManager.StopListener(newListenerID); err != nil {
				log.Printf("|⚠️ HOP| Failed to clean up unused listener %s: %v",
					newListenerID, err)
			}
		}
		return fmt.Errorf("hop failed: %s", agentResult.Error)
	}

	// Hop initiated successfully - now we wait for the new connection
	// and then clean up the old listener

	// Start a goroutine to clean up the old listener after a delay
	if oldListenerID, ok := task.Metadata["old_listener_id"].(string); ok {
		go h.scheduleOldListenerCleanup(task.AgentID, oldListenerID, 30*time.Second)
	}

	log.Printf("|✅ HOP RESULT| Agent %s hop initiated successfully", task.AgentID)
	return nil
}

// scheduleOldListenerCleanup attempts to stop the old listener after the agent has hopped
func (h *HopOrchestrator) scheduleOldListenerCleanup(agentID, oldListenerID string, delay time.Duration) {
	// Wait for agent to establish new connection
	time.Sleep(delay)

	// Check if hop completed
	if h.agentTracker != nil {
		if agentInfo, exists := h.agentTracker.GetAgentInfo(agentID); exists {
			if agentInfo.ListenerID != oldListenerID && !agentInfo.IsHopping {
				// Agent successfully moved to new listener
				log.Printf("|🧹 HOP CLEANUP| Agent %s confirmed on new listener, cleaning up old listener %s",
					agentID, oldListenerID)

				// Try to stop old listener with retries
				err := h.listenerManager.TryStopListener(oldListenerID, 3, 10*time.Second)
				if err != nil {
					log.Printf("|⚠️ HOP CLEANUP| Failed to stop old listener %s: %v",
						oldListenerID, err)
				} else {
					log.Printf("|✅ HOP CLEANUP| Successfully stopped old listener %s",
						oldListenerID)
				}
			} else {
				log.Printf("|⚠️ HOP CLEANUP| Agent %s hop not completed, keeping old listener %s",
					agentID, oldListenerID)
			}
		}
	}
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

// ValidateArgs checks if Hop arguments are valid
func (h *HopOrchestrator) ValidateArgs(args json.RawMessage) error {
	var hopArgs models.HopArgs

	if err := json.Unmarshal(args, &hopArgs); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	// Validate mandatory fields
	if hopArgs.NewProtocol == "" {
		return fmt.Errorf("NewProtocol is required")
	}

	if hopArgs.NewServerPort == "" {
		return fmt.Errorf("NewServerPort is required")
	}

	if hopArgs.NewServerIP == "" {
		return fmt.Errorf("NewServerIP is required") // Server must always specify the IP
	}

	// Validate protocol is a known type
	switch hopArgs.NewProtocol {
	case config.HTTP1Clear, config.HTTP1TLS, config.HTTP2TLS,
		config.HTTP3, config.WebsocketClear, config.WebsocketSecure:
		// Valid protocol
	default:
		return fmt.Errorf("invalid protocol specified: %s", hopArgs.NewProtocol)
	}

	// Validate optional parameters if provided
	if hopArgs.NewDelay != nil {
		if _, err := time.ParseDuration(*hopArgs.NewDelay); err != nil {
			return fmt.Errorf("invalid NewDelay format: %v", err)
		}
	}

	if hopArgs.NewJitter != nil {
		if *hopArgs.NewJitter < 0.0 || *hopArgs.NewJitter > 1.0 {
			return fmt.Errorf("NewJitter must be between 0.0 and 1.0, got %f", *hopArgs.NewJitter)
		}
	}

	if hopArgs.NewCheckinMethod != nil {
		method := strings.ToUpper(*hopArgs.NewCheckinMethod)
		if method != "GET" && method != "POST" {
			return fmt.Errorf("invalid NewCheckinMethod: %s (must be GET or POST)", *hopArgs.NewCheckinMethod)
		}
	}

	// Validate padding bounds if provided
	if hopArgs.NewMinPaddingBytes != nil && *hopArgs.NewMinPaddingBytes < 0 {
		return fmt.Errorf("NewMinPaddingBytes cannot be negative")
	}

	if hopArgs.NewMaxPaddingBytes != nil && *hopArgs.NewMaxPaddingBytes < 0 {
		return fmt.Errorf("NewMaxPaddingBytes cannot be negative")
	}

	// If both padding values provided, ensure max >= min
	if hopArgs.NewMinPaddingBytes != nil && hopArgs.NewMaxPaddingBytes != nil {
		if *hopArgs.NewMaxPaddingBytes < *hopArgs.NewMinPaddingBytes {
			return fmt.Errorf("NewMaxPaddingBytes (%d) must be >= NewMinPaddingBytes (%d)",
				*hopArgs.NewMaxPaddingBytes, *hopArgs.NewMinPaddingBytes)
		}
	}

	return nil
}
