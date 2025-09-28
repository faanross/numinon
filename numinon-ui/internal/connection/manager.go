package connection

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"numinon-ui/internal/models"
)

// Manager handles our mock C2 server connection
type Manager struct {
	ctx         context.Context
	mu          sync.RWMutex
	status      models.ConnectionStatus
	isConnected bool
	pingTicker  *time.Ticker
	pingStop    chan bool
	mockAgents  []models.Agent // For demonstration
}

// NewManager creates a new connection manager
func NewManager() *Manager {
	return &Manager{
		status: models.ConnectionStatus{
			Connected: false,
			ServerURL: "ws://localhost:8080/client",
		},
		pingStop:   make(chan bool),
		mockAgents: generateMockAgents(),
	}
}

// Startup is called when the app starts
func (m *Manager) Startup(ctx context.Context) {
	m.ctx = ctx
	// Emit initial status with consistent naming
	runtime.EventsEmit(m.ctx, "connection_status", m.status)
}

// Connect establishes a connection to the C2 server
func (m *Manager) Connect(serverURL string) models.ConnectionStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simulate connection attempt
	m.status.ServerURL = serverURL

	// TODO: Real WebSocket connection here
	// For now, simulate with delay
	time.Sleep(500 * time.Millisecond)

	// Simulate 90% success rate for demonstration
	if rand.Float32() > 0.1 {
		m.isConnected = true
		m.status.Connected = true
		m.status.LastPing = time.Now()
		m.status.Error = ""
		m.status.Latency = rand.Intn(50) + 10 // 10-60ms

		// Start ping routine
		m.startPingRoutine()

		// CRITICAL: Emit the agents immediately after connection
		runtime.EventsEmit(m.ctx, "agent_update", m.mockAgents)

		// Start sending random events
		go m.simulateServerEvents()

	} else {
		m.isConnected = false
		m.status.Connected = false
		m.status.Error = "Connection refused: unable to reach server"
	}

	// Always emit status update with consistent naming
	runtime.EventsEmit(m.ctx, "connection_status", m.status)

	return m.status
}

// Disconnect closes the connection
func (m *Manager) Disconnect() models.ConnectionStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isConnected {
		// Stop ping routine
		if m.pingTicker != nil {
			m.pingTicker.Stop()
			select {
			case m.pingStop <- true:
			default:
				// Channel might be closed, ignore
			}
		}

		m.isConnected = false
		m.status.Connected = false
		m.status.Error = ""

		// Clear agents on disconnect
		runtime.EventsEmit(m.ctx, "agent_update", []models.Agent{})
	}

	runtime.EventsEmit(m.ctx, "connection_status", m.status)
	return m.status
}

// GetStatus returns the current connection status
func (m *Manager) GetStatus() models.ConnectionStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// startPingRoutine sends periodic pings to keep connection alive
func (m *Manager) startPingRoutine() {
	if m.pingTicker != nil {
		m.pingTicker.Stop()
	}

	m.pingTicker = time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-m.pingTicker.C:
				m.sendPing()
			case <-m.pingStop:
				return
			}
		}
	}()
}

// sendPing simulates sending a ping and updates latency
func (m *Manager) sendPing() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isConnected {
		return
	}

	// Simulate ping
	m.status.LastPing = time.Now()
	m.status.Latency = rand.Intn(50) + 10 // 10-60ms

	// Emit the full status update (frontend expects this)
	runtime.EventsEmit(m.ctx, "connection_status", m.status)
}

// simulateServerEvents mimics receiving events from the C2 server
func (m *Manager) simulateServerEvents() {
	eventTypes := []string{"agent:connected", "agent:disconnected", "task:completed"}
	ticker := time.NewTicker(time.Duration(rand.Intn(10)+5) * time.Second)
	defer ticker.Stop()

	for m.isConnected {
		select {
		case <-ticker.C:
			if !m.isConnected {
				return
			}

			// Create random event
			eventType := eventTypes[rand.Intn(len(eventTypes))]
			message := models.ServerMessage{
				Type:      eventType,
				Timestamp: time.Now(),
			}

			// Add appropriate payload based on event type
			switch eventType {
			case "agent:connected":
				// Pick a random agent and mark it online
				if len(m.mockAgents) > 0 {
					agentIdx := rand.Intn(len(m.mockAgents))
					m.mockAgents[agentIdx].Status = "online"
					m.mockAgents[agentIdx].LastSeen = time.Now()
					message.Payload = m.mockAgents[agentIdx]

					// Emit updated agent list
					runtime.EventsEmit(m.ctx, "agent_update", m.mockAgents)
				}

			case "agent:disconnected":
				// Pick a random agent and mark it offline
				if len(m.mockAgents) > 0 {
					agentIdx := rand.Intn(len(m.mockAgents))
					m.mockAgents[agentIdx].Status = "offline"
					message.Payload = m.mockAgents[agentIdx]

					// Emit updated agent list
					runtime.EventsEmit(m.ctx, "agent_update", m.mockAgents)
				}

			case "task:completed":
				message.Payload = map[string]interface{}{
					"taskId":  fmt.Sprintf("task_%d", rand.Intn(1000)),
					"agentId": m.mockAgents[rand.Intn(len(m.mockAgents))].ID,
					"success": rand.Float32() > 0.3,
				}
			}

			// Emit the server event
			runtime.EventsEmit(m.ctx, "server_message", message)

		case <-m.pingStop:
			return
		}
	}
}

// GetAgents returns the list of agents
func (m *Manager) GetAgents() []models.Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Update last seen times for online agents
	for i := range m.mockAgents {
		if m.mockAgents[i].Status == "online" {
			m.mockAgents[i].LastSeen = time.Now().Add(-time.Duration(rand.Intn(300)) * time.Second)
		}
	}

	return m.mockAgents
}

// SendCommand simulates sending a command to an agent
func (m *Manager) SendCommand(req models.CommandRequest) models.CommandResponse {
	if !m.isConnected {
		return models.CommandResponse{
			Success: false,
			Error:   "Not connected to server",
		}
	}

	// Simulate command execution
	time.Sleep(time.Duration(rand.Intn(2000)+500) * time.Millisecond)

	// Simulate 80% success rate
	if rand.Float32() > 0.2 {
		// Simulate different outputs based on command
		var output string
		switch req.Command {
		case "whoami":
			output = "nt authority\\system"
		case "hostname":
			output = "DESKTOP-" + fmt.Sprintf("%06d", rand.Intn(1000000))
		case "pwd":
			output = "C:\\Windows\\System32"
		default:
			output = fmt.Sprintf("Command '%s' executed successfully\n[Output would appear here]", req.Command)
		}

		return models.CommandResponse{
			Success: true,
			Output:  output,
		}
	}

	return models.CommandResponse{
		Success: false,
		Error:   "Command execution failed: agent timeout",
	}
}

// Helper function to generate mock agents
func generateMockAgents() []models.Agent {
	return []models.Agent{
		{
			ID:        "agent_001",
			Hostname:  "DESKTOP-WIN10",
			OS:        "Windows 10",
			Status:    "online",
			IPAddress: "192.168.1.100",
			LastSeen:  time.Now(),
		},
		{
			ID:        "agent_002",
			Hostname:  "macbook-pro",
			OS:        "macOS 13.0",
			Status:    "online",
			IPAddress: "192.168.1.101",
			LastSeen:  time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        "agent_003",
			Hostname:  "ubuntu-server",
			OS:        "Ubuntu 22.04",
			Status:    "offline",
			IPAddress: "192.168.1.102",
			LastSeen:  time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "agent_004",
			Hostname:  "win-server-2019",
			OS:        "Windows Server 2019",
			Status:    "online",
			IPAddress: "192.168.1.103",
			LastSeen:  time.Now().Add(-10 * time.Minute),
		},
	}
}
