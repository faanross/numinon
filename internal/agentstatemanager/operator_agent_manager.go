package agentstatemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"numinon_shadow/internal/clientapi"
	"numinon_shadow/internal/tracker"
)

// OperatorAgentManager implements clientapi.AgentStateManager.
// It provides operator-layer access to agent information.
// This is achieved by wrapping our existing tracker.Tracker type
//
// This is like the "agent status dashboard" for operators:
// - Shows which agents are connected
// - Provides detailed agent information
// - Tracks agent state changes (like hops)
type OperatorAgentManager struct {
	tracker *tracker.Tracker // The actual agent tracker
}

// NewOperatorAgentManager creates a new operator-friendly agent manager.
func NewOperatorAgentManager(tracker *tracker.Tracker) *OperatorAgentManager {
	return &OperatorAgentManager{
		tracker: tracker,
	}
}

// ListAgents returns a summary of all known agents.
func (m *OperatorAgentManager) ListAgents(ctx context.Context, operatorSessionID string) (clientapi.ServerResponse, error) {
	log.Printf("[AGENT API] Operator %s requesting agent list", operatorSessionID)

	// Get all agents from tracker
	agents := m.tracker.GetAllAgents()

	// Convert to operator-friendly format
	var agentInfos []clientapi.AgentInfo
	for _, agent := range agents {
		// Determine status string
		statusStr := string(agent.State)
		if agent.IsHopping {
			statusStr = "hopping"
		}

		info := clientapi.AgentInfo{
			AgentID:           agent.ID,
			FirstSeen:         agent.ConnectedAt.Format("2006-01-02 15:04:05"),
			LastSeen:          agent.LastSeenAt.Format("2006-01-02 15:04:05"),
			SourceIP:          agent.RemoteAddr,
			CurrentListenerID: agent.ListenerID,
			// Add a Status field to AgentInfo if needed
		}
		agentInfos = append(agentInfos, info)
	}

	agentList := clientapi.AgentListPayload{
		Agents: agentInfos,
	}

	payloadBytes, _ := json.Marshal(agentList)

	return clientapi.ServerResponse{
		Status:   clientapi.StatusSuccess,
		DataType: clientapi.DataTypeAgentList,
		Payload:  payloadBytes,
	}, nil
}

// GetAgentDetails returns detailed information for a single agent.
func (m *OperatorAgentManager) GetAgentDetails(ctx context.Context, agentID string, operatorSessionID string) (clientapi.ServerResponse, error) {
	log.Printf("[AGENT API] Operator %s requesting details for agent %s", operatorSessionID, agentID)

	// Get agent from tracker
	agentInfo, exists := m.tracker.GetAgentInfo(agentID)
	if !exists {
		return clientapi.ServerResponse{
			Status: clientapi.StatusError,
			Error:  fmt.Sprintf("Agent %s not found", agentID),
		}, fmt.Errorf("agent %s not found", agentID)
	}

	// Convert to operator-friendly format
	details := clientapi.AgentInfo{
		AgentID:           agentInfo.ID,
		FirstSeen:         agentInfo.ConnectedAt.Format("2006-01-02 15:04:05"),
		LastSeen:          agentInfo.LastSeenAt.Format("2006-01-02 15:04:05"),
		SourceIP:          agentInfo.RemoteAddr,
		CurrentListenerID: agentInfo.ListenerID,
		// OS, Arch, Hostname, Username would come from agent enumeration
	}

	payloadBytes, _ := json.Marshal(details)

	return clientapi.ServerResponse{
		Status:   clientapi.StatusSuccess,
		DataType: clientapi.DataTypeAgentDetails,
		Payload:  payloadBytes,
	}, nil
}

// InitiateHop signals an agent to HOP, allowing for stateful reporting.
func (m *OperatorAgentManager) InitiateHop(agentID, taskID, listenerID string, cancelFunc context.CancelFunc) {
	log.Printf("[AGENT API] Marking agent %s as initiating hop to listener %s", agentID, listenerID)

	if err := m.tracker.InitiateHop(agentID, listenerID); err != nil {
		log.Printf("[AGENT API] Failed to mark hop initiation: %v", err)
	}
}

// ClearHopState removes the pending hop state from an agent following success/failure.
func (m *OperatorAgentManager) ClearHopState(agentID string) {
	log.Printf("[AGENT API] Clearing hop state for agent %s", agentID)
	// TODO: Add method to tracker to clear hop state
}
