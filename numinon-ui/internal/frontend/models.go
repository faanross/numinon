// Package frontend contains Data Transfer Objects (DTOs) used for communication
// between the Go backend and the frontend. These structs are simplified and
// contain only types that are safe for JSON serialization and Wails's
// TypeScript generation.
package frontend

// ConnectionStatusDTO is the data transfer object for ConnectionStatus.
type ConnectionStatusDTO struct {
	Connected bool   `json:"connected"`
	ServerURL string `json:"serverUrl"`
	LastPing  string `json:"lastPing"` // Changed from time.Time
	Latency   int    `json:"latency"`
	Error     string `json:"error,omitempty"`
}

// ServerMessageDTO is the data transfer object for ServerMessage.
type ServerMessageDTO struct {
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"` // Changed from time.Time
	Payload   interface{} `json:"payload"`
}

// AgentDTO is the data transfer object for Agent.
type AgentDTO struct {
	ID        string `json:"id"`
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Status    string `json:"status"`
	LastSeen  string `json:"lastSeen"` // Changed from time.Time
	IPAddress string `json:"ipAddress"`
}

// CommandRequestDTO is the data transfer object for CommandRequest.
// Although it contains no complex types, we create a DTO for consistency.
type CommandRequestDTO struct {
	AgentID   string `json:"agentId"`
	Command   string `json:"command"`
	Arguments string `json:"arguments"`
}

// CommandResponseDTO is the data transfer object for CommandResponse.
// It is already safe, but we create a DTO to maintain the pattern.
type CommandResponseDTO struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}
