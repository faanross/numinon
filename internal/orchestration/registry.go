package orchestration

import (
	"encoding/json"
	"fmt"
	"github.com/faanross/numinon/internal/taskmanager"
	"log"
	"sync"
)

// Registry manages command orchestrators
type Registry struct {
	mu            sync.RWMutex
	orchestrators map[string]CommandOrchestrator
	defaultOrch   CommandOrchestrator
}

// NewRegistry creates a new orchestrator registry
func NewRegistry() *Registry {
	return &Registry{
		orchestrators: make(map[string]CommandOrchestrator),
		defaultOrch:   &DefaultOrchestrator{}, // For commands without special needs
	}
}

// Register adds an orchestrator for a specific command type
func (r *Registry) Register(command string, orchestrator CommandOrchestrator) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orchestrators[command] = orchestrator
	log.Printf("|📋 ORCH| Registered orchestrator for command: %s", command)
}

// Get returns the orchestrator for a command, or the default if not found
func (r *Registry) Get(command string) CommandOrchestrator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if orch, exists := r.orchestrators[command]; exists {
		return orch
	}

	// Return default orchestrator for commands without special needs
	return r.defaultOrch
}

// PrepareTask prepares a task using the appropriate orchestrator
func (r *Registry) PrepareTask(task *taskmanager.Task) error {
	orch := r.Get(task.Command)
	return orch.PrepareTask(task)
}

// ProcessResult processes a result using the appropriate orchestrator
func (r *Registry) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {
	orch := r.Get(task.Command)
	return orch.ProcessResult(task, result)
}

// DefaultOrchestrator handles commands that don't need special processing
type DefaultOrchestrator struct{}

// PrepareTask does nothing for simple commands
func (d *DefaultOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// No special preparation needed
	return nil
}

// ProcessResult does nothing for simple commands
func (d *DefaultOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {
	// No special processing needed
	log.Printf("|📋 ORCH| No special processing for command: %s", task.Command)
	return nil
}

// ValidateArgs does basic validation
func (d *DefaultOrchestrator) ValidateArgs(args json.RawMessage) error {
	// Basic check that args is valid JSON
	if len(args) > 0 && !json.Valid(args) {
		return fmt.Errorf("invalid JSON arguments")
	}
	return nil
}
