package tracker

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ConnectionState represents the current state of an agent connection
type ConnectionState string

const (
	StateConnected     ConnectionState = "connected"
	StateDisconnected  ConnectionState = "disconnected"
	StateTransitioning ConnectionState = "transitioning" // During HOP
)

// ConnectionType represents how the agent is connected
type ConnectionType string

const (
	TypeHTTP      ConnectionType = "http"
	TypeWebSocket ConnectionType = "websocket"
)

// AgentInfo holds all information about a connected agent
type AgentInfo struct {
	ID             string          `json:"id"`
	ListenerID     string          `json:"listener_id"`
	ConnectionType ConnectionType  `json:"connection_type"`
	State          ConnectionState `json:"state"`
	ConnectedAt    time.Time       `json:"connected_at"`
	LastSeenAt     time.Time       `json:"last_seen_at"`
	Protocol       string          `json:"protocol"` // The specific protocol (H1C, WSS, etc.)
	RemoteAddr     string          `json:"remote_addr"`

	// HOP-specific fields
	IsHopping      bool       `json:"is_hopping"`
	NewListenerID  string     `json:"new_listener_id,omitempty"`
	HopInitiatedAt *time.Time `json:"hop_initiated_at,omitempty"`
}

// Tracker manages all agent connections and their states
type Tracker struct {
	mu             sync.RWMutex
	agents         map[string]*AgentInfo // key: agentID
	listenerAgents map[string][]string   // key: listenerID, value: []agentID

	// Configuration
	httpTimeout    time.Duration // How long before HTTP agent is considered disconnected
	hopGracePeriod time.Duration // How long to wait for hop completion
}

// NewTracker creates a new agent tracker
func NewTracker() *Tracker {
	t := &Tracker{
		agents:         make(map[string]*AgentInfo),
		listenerAgents: make(map[string][]string),
		// TODO these values should originate from the config system
		httpTimeout:    60 * time.Minute, // HTTP agents timeout after 60 min of no contact
		hopGracePeriod: 30 * time.Second, // 30 seconds to complete hop
	}

	// Start cleanup goroutine for stale connections
	go t.cleanupLoop()

	return t
}

// RegisterConnection records a new agent connection or updates an existing one
func (t *Tracker) RegisterConnection(agentID, listenerID, protocol, remoteAddr string, connType ConnectionType) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	// Check if agent exists
	if existing, exists := t.agents[agentID]; exists {
		// Agent reconnecting or checking in
		if existing.IsHopping && existing.NewListenerID == listenerID {
			// This is the expected new connection after hop!
			log.Printf("|🎯 TRACKER| Agent %s completed hop to listener %s", agentID, listenerID)

			// Remove from old listener
			t.removeAgentFromListener(agentID, existing.ListenerID)

			// Update agent info
			existing.ListenerID = listenerID
			existing.State = StateConnected
			existing.IsHopping = false
			existing.NewListenerID = ""
			existing.HopInitiatedAt = nil
			existing.Protocol = protocol
			existing.ConnectionType = connType
			existing.LastSeenAt = now

			// Add to new listener
			t.addAgentToListener(agentID, listenerID)

			return nil
		} else if existing.ListenerID == listenerID {
			// Same listener - just update last seen
			existing.LastSeenAt = now
			existing.State = StateConnected
			return nil
		} else {
			// Unexpected listener! This shouldn't happen unless hop wasn't tracked
			log.Printf("|⚠️ TRACKER| Agent %s connected to unexpected listener %s (expected %s)",
				agentID, listenerID, existing.ListenerID)
		}
	}

	// New agent registration
	agent := &AgentInfo{
		ID:             agentID,
		ListenerID:     listenerID,
		ConnectionType: connType,
		State:          StateConnected,
		ConnectedAt:    now,
		LastSeenAt:     now,
		Protocol:       protocol,
		RemoteAddr:     remoteAddr,
	}

	t.agents[agentID] = agent
	t.addAgentToListener(agentID, listenerID)

	log.Printf("|📝 TRACKER| Registered agent %s on listener %s (%s)",
		agentID, listenerID, protocol)

	return nil
}

// UpdateLastSeen updates the last seen timestamp for an agent
func (t *Tracker) UpdateLastSeen(agentID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if agent, exists := t.agents[agentID]; exists {
		agent.LastSeenAt = time.Now()
		agent.State = StateConnected
	}
}

// MarkDisconnected marks an agent as disconnected
func (t *Tracker) MarkDisconnected(agentID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if agent, exists := t.agents[agentID]; exists {
		agent.State = StateDisconnected
		log.Printf("|🔌 TRACKER| Agent %s marked as disconnected", agentID)
	}
}

// InitiateHop marks an agent as transitioning to a new listener
func (t *Tracker) InitiateHop(agentID, newListenerID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	agent, exists := t.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	now := time.Now()
	agent.IsHopping = true
	agent.NewListenerID = newListenerID
	agent.HopInitiatedAt = &now
	agent.State = StateTransitioning

	log.Printf("|🐇 TRACKER| Agent %s initiating hop from listener %s to %s",
		agentID, agent.ListenerID, newListenerID)

	return nil
}

// GetAgentInfo returns information about a specific agent
func (t *Tracker) GetAgentInfo(agentID string) (*AgentInfo, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	agent, exists := t.agents[agentID]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	agentCopy := *agent
	return &agentCopy, true
}

// GetListenerAgents returns all agents connected to a specific listener
func (t *Tracker) GetListenerAgents(listenerID string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	agents := t.listenerAgents[listenerID]

	// Return active agents only
	var activeAgents []string
	for _, agentID := range agents {
		if agent, exists := t.agents[agentID]; exists {
			if agent.State == StateConnected && !agent.IsHopping {
				activeAgents = append(activeAgents, agentID)
			}
		}
	}

	return activeAgents
}

// CanStopListener checks if a listener can be safely stopped
func (t *Tracker) CanStopListener(listenerID string) (bool, string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	agents := t.listenerAgents[listenerID]

	if len(agents) == 0 {
		return true, "no agents connected"
	}

	// Check if all agents are disconnected or have hopped away
	activeCount := 0
	for _, agentID := range agents {
		if agent, exists := t.agents[agentID]; exists {
			if agent.ListenerID == listenerID &&
				agent.State == StateConnected &&
				!agent.IsHopping {
				activeCount++
			}
		}
	}

	if activeCount == 0 {
		return true, "all agents have disconnected or hopped away"
	}

	return false, fmt.Sprintf("%d active agents still connected", activeCount)
}

// Internal helper functions

func (t *Tracker) addAgentToListener(agentID, listenerID string) {
	if t.listenerAgents[listenerID] == nil {
		t.listenerAgents[listenerID] = []string{}
	}

	// Check if already in list
	for _, id := range t.listenerAgents[listenerID] {
		if id == agentID {
			return
		}
	}

	t.listenerAgents[listenerID] = append(t.listenerAgents[listenerID], agentID)
}

func (t *Tracker) removeAgentFromListener(agentID, listenerID string) {
	agents := t.listenerAgents[listenerID]
	for i, id := range agents {
		if id == agentID {
			// Remove from slice
			t.listenerAgents[listenerID] = append(agents[:i], agents[i+1:]...)
			break
		}
	}

	// Clean up empty entries
	if len(t.listenerAgents[listenerID]) == 0 {
		delete(t.listenerAgents, listenerID)
	}
}

// cleanupLoop periodically checks for stale connections
func (t *Tracker) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t.cleanupStaleConnections()
	}
}

func (t *Tracker) cleanupStaleConnections() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	for agentID, agent := range t.agents {
		// Check HTTP agents for timeout
		if agent.ConnectionType == TypeHTTP &&
			agent.State == StateConnected &&
			now.Sub(agent.LastSeenAt) > t.httpTimeout {
			agent.State = StateDisconnected
			log.Printf("|⏰ TRACKER| HTTP agent %s timed out (last seen: %v ago)",
				agentID, now.Sub(agent.LastSeenAt))
		}

		// Check for stuck hop transitions
		if agent.IsHopping && agent.HopInitiatedAt != nil &&
			now.Sub(*agent.HopInitiatedAt) > t.hopGracePeriod {
			log.Printf("|⚠️ TRACKER| Agent %s hop timed out after %v",
				agentID, t.hopGracePeriod)
			agent.IsHopping = false
			agent.NewListenerID = ""
			agent.HopInitiatedAt = nil
			// Agent stays on original listener
		}
	}
}

// GetStats returns statistics about tracked agents
func (t *Tracker) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := map[string]interface{}{
		"total_agents":    len(t.agents),
		"total_listeners": len(t.listenerAgents),
		"connected":       0,
		"disconnected":    0,
		"transitioning":   0,
	}

	for _, agent := range t.agents {
		switch agent.State {
		case StateConnected:
			stats["connected"] = stats["connected"].(int) + 1
		case StateDisconnected:
			stats["disconnected"] = stats["disconnected"].(int) + 1
		case StateTransitioning:
			stats["transitioning"] = stats["transitioning"].(int) + 1
		}
	}

	return stats
}
