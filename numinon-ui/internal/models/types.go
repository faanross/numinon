package models

import "time"

// ConnectionStatus represents the current connection state
type ConnectionStatus struct {
	Connected bool      `json:"connected"`
	ServerURL string    `json:"serverUrl"`
	LastPing  time.Time `json:"lastPing"`
	Latency   int       `json:"latency"` // milliseconds
	Error     string    `json:"error,omitempty"`
}

// ServerMessage represents a message from our C2 server
type ServerMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// Agent represents a connected agent (simplified for now)
type Agent struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	OS        string    `json:"os"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"lastSeen"`
	IPAddress string    `json:"ipAddress"`
}

// CommandRequest represents a command to send to an agent
type CommandRequest struct {
	AgentID   string `json:"agentId"`
	Command   string `json:"command"`
	Arguments string `json:"arguments"`
}

// CommandResponse represents the result of a command
type CommandResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}
