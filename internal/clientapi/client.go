package clientapi

import "encoding/json"

// Client -> Server Request
type ClientRequest struct {
	RequestID string          `json:"request_id"` // e.g., "req_123456"
	Action    ActionType      `json:"action"`     // e.g., ActionCreateListener
	Payload   json.RawMessage `json:"payload"`    // Raw JSON for action-specific parameters
}
