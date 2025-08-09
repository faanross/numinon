package models

import "encoding/json"

// ServerTaskResponse represents the structure expected from the /checkin endpoint response.
type ServerTaskResponse struct {
	TaskAvailable bool            `json:"task_available"`
	TaskID        string          `json:"task_id,omitempty"`
	Command       string          `json:"command,omitempty"`
	Data          json.RawMessage `json:"data,omitempty"` // for example if command has arguments, a file (upload) etc
}

// AgentTaskResult represents the structure sent back to the /results endpoint.
type AgentTaskResult struct {
	TaskID     string          `json:"task_id"`
	Status     string          `json:"status"`
	Output     json.RawMessage `json:"output,omitempty"`
	Error      string          `json:"error,omitempty"`
	FileSha256 string          `json:"file_sha256,omitempty"`
}

// AgentCheckIn is used to allow for payload jitter (padding) if enabled
type AgentCheckIn struct {
	Padding string
}
