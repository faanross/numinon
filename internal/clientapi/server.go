package clientapi

import "encoding/json"

type StatusType string

const (
	StatusSuccess StatusType = "SUCCESS"
	StatusPending StatusType = "PENDING"
	StatusError   StatusType = "ERROR"
	StatusUpdate  StatusType = "UPDATE"
)

// ServerResponse is sent back to client after a request BEFORE it's passed to agent
type ServerResponse struct {
	RequestID string          `json:"request_id,omitempty"` // Correlates to ClientRequest.RequestID
	Status    StatusType      `json:"status"`
	Action    ActionType      `json:"action,omitempty"`    // Original action or event type
	DataType  DataType        `json:"data_type,omitempty"` // Hint for payload structure, aids parsing
	Payload   json.RawMessage `json:"payload,omitempty"`   // Raw JSON for response data or error details
	Error     string          `json:"error,omitempty"`     // Simple error message if Status is Error
}

// ErrorDetailsPayload provides a structured way to send back error details in a ServerResponse payload.
type ErrorDetailsPayload struct {
	Detail string `json:"detail"`
}

// TaskResultEventPayload is sent from server to client when an agent returns results
type TaskResultEventPayload struct {
	AgentID     string `json:"agent_id"`
	TaskID      string `json:"task_id"`      // The ID of the task this result is for
	CommandType string `json:"command_type"` // e.g., "run_cmd", "upload"

	// ResultData will contain the actual command-specific result,
	// e.g., models.RunCmdResult, models.DownloadResult.
	// Using json.RawMessage allows the client to unmarshal it based on CommandType.
	ResultData json.RawMessage `json:"result_data"`

	// Success indicates if the command execution on the agent was successful.
	// This is different from the overall ServerResponse.Status, which just indicates whether server understood initial request and was able to send it to the agent
	CommandSuccess bool   `json:"command_success"`
	ErrorMsg       string `json:"error_msg,omitempty"` // Error message from the agent if command_success is false
}
