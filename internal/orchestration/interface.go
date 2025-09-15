// Package orchestration defines command orchestration on the SERVER SIDE
package orchestration

import (
	"encoding/json"
	"github.com/faanross/numinon/internal/taskmanager"
)

// CommandOrchestrator handles the full lifecycle of a command
type CommandOrchestrator interface {
	// PrepareTask sets up server-side metadata and validates arguments before dispatch
	PrepareTask(task *taskmanager.Task) error

	// ProcessResult handles command-specific result processing after execution
	ProcessResult(task *taskmanager.Task, result json.RawMessage) error

	// ValidateArgs checks if the command arguments are valid
	ValidateArgs(args json.RawMessage) error
}
