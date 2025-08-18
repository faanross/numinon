package taskmanager

import (
	"encoding/json"
	"errors"
)

// Common errors
var (
	ErrTaskNotFound = errors.New("task not found")
	ErrTaskExists   = errors.New("task already exists")
	ErrInvalidState = errors.New("invalid state transition")
)

// TaskManager defines the interface for task storage and retrieval
type TaskManager interface {
	// CreateTask stores a new task and returns it
	CreateTask(agentID, command string, args json.RawMessage) (*Task, error)

	// GetTask retrieves a task by ID
	GetTask(taskID string) (*Task, error)

	// UpdateTask updates an existing task
	UpdateTask(task *Task) error

	// MarkDispatched updates task status when sent to agent
	MarkDispatched(taskID string) error

	// StoreResult stores result and marks task completed
	StoreResult(taskID string, result json.RawMessage) error

	// MarkFailed marks a task as failed with error message
	MarkFailed(taskID string, errorMsg string) error

	// GetAgentTasks returns all tasks for a specific agent (useful for debugging)
	GetAgentTasks(agentID string) ([]*Task, error)

	// GetPendingTasks returns tasks that haven't been dispatched yet
	GetPendingTasks() ([]*Task, error)
}

// ResultProcessor defines the interface for command-specific result handling
type ResultProcessor interface {
	// ProcessResult handles command-specific result processing
	// Returns error if processing fails (e.g., hash mismatch, file save failure)
	ProcessResult(task *Task, result json.RawMessage) error
}

// ResultProcessorRegistry manages command-specific processors
type ResultProcessorRegistry struct {
	processors map[string]ResultProcessor
}

// NewResultProcessorRegistry creates a new processor registry
func NewResultProcessorRegistry() *ResultProcessorRegistry {
	return &ResultProcessorRegistry{
		processors: make(map[string]ResultProcessor),
	}
}

// Register adds a processor for a specific command type
func (r *ResultProcessorRegistry) Register(command string, processor ResultProcessor) {
	r.processors[command] = processor
}

// Process looks up and executes the appropriate processor for a task
func (r *ResultProcessorRegistry) Process(task *Task) error {
	processor, exists := r.processors[task.Command]
	if !exists {
		// No processor registered - that's OK for simple commands
		return nil
	}
	return processor.ProcessResult(task, task.Result)
}
