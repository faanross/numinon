package clientapi

import "encoding/json"

type StatusType string

const (
	StatusSuccess StatusType = "SUCCESS"
	StatusPending StatusType = "PENDING"
	StatusError   StatusType = "ERROR"
	StatusUpdate  StatusType = "UPDATE"
)

// Server -> Client Response
type ServerResponse struct {
	RequestID string          `json:"request_id,omitempty"` // Correlates to ClientRequest.RequestID
	Status    StatusType      `json:"status"`
	Action    ActionType      `json:"action,omitempty"`    // Original action or event type
	DataType  DataType        `json:"data_type,omitempty"` // Hint for payload structure, aids parsing
	Payload   json.RawMessage `json:"payload,omitempty"`   // Raw JSON for response data or error details
	Error     string          `json:"error,omitempty"`     // Simple error message if Status is Error
}
