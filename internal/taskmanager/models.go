package taskmanager

import (
	"encoding/json"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusDispatched TaskStatus = "dispatched"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
	StatusTimeout    TaskStatus = "timeout"
)

// Task represents a command task issued to an agent
type Task struct {
	// Identity
	ID      string `json:"id"`       // Unique task identifier
	AgentID string `json:"agent_id"` // Target agent UUID
	Command string `json:"command"`  // Command type (upload, download, etc.)

	// Status
	Status TaskStatus `json:"status"`
	Error  string     `json:"error,omitempty"` // Error message if failed

	// Temporal
	CreatedAt    time.Time  `json:"created_at"`
	DispatchedAt *time.Time `json:"dispatched_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`

	// Payload & Results
	Arguments json.RawMessage `json:"arguments"`        // Original command arguments
	Result    json.RawMessage `json:"result,omitempty"` // Raw result from agent

	// Command-specific metadata (for reconciliation)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewTask creates a new task with initial state
func NewTask(id, agentID, command string, args json.RawMessage) *Task {
	return &Task{
		ID:        id,
		AgentID:   agentID,
		Command:   command,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		Arguments: args,
		Metadata:  make(map[string]interface{}),
	}
}

// MarkDispatched updates task status when sent to agent
func (t *Task) MarkDispatched() {
	now := time.Now()
	t.DispatchedAt = &now
	t.Status = StatusDispatched
}

// MarkCompleted updates task status when result received
func (t *Task) MarkCompleted(result json.RawMessage) {
	now := time.Now()
	t.CompletedAt = &now
	t.Status = StatusCompleted
	t.Result = result
}

// MarkFailed updates task status on failure
func (t *Task) MarkFailed(err string) {
	now := time.Now()
	t.CompletedAt = &now
	t.Status = StatusFailed
	t.Error = err
}

// IsTerminal returns true if task is in a final state
func (t *Task) IsTerminal() bool {
	return t.Status == StatusCompleted ||
		t.Status == StatusFailed ||
		t.Status == StatusTimeout
}
