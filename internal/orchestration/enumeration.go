package orchestration

import (
	"encoding/json"
	"fmt"
	"github.com/faanross/numinon/internal/taskmanager"
)

// EnumerationOrchestrator handles the Enumeration command lifecycle
type EnumerationOrchestrator struct {
	// in future add option to display result, save result, or both
}

// NewEnumerationOrchestrator creates a new Enumeration orchestrator
func NewEnumerationOrchestrator() *EnumerationOrchestrator {
	return &EnumerationOrchestrator{}
}

// PrepareTask sets up sEnumeration-specific metadata
func (d *EnumerationOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// Validate arguments first
	if err := d.ValidateArgs(task.Arguments); err != nil {
		return fmt.Errorf("invalid Enumeration arguments: %w", err)
	}

	// Enumeration has no additional preparation work to do, so just validation is good

	return nil
}

// ValidateArgs checks if shellcode arguments are valid
func (d *EnumerationOrchestrator) ValidateArgs(args json.RawMessage) error {

	// We only have 1 argument - ProcessName
	// But, it can be "" - meaning enumerate everything
	// So perhaps in future we can check whether if it is !"", if it is a valid processName
	// But for now, this just serves as a stub

	return nil
}

// ProcessResult handles the enumeration results processing
func (d *EnumerationOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {

	// For now not much we need to do here so this is just a stub
	// So we can satisfy the interface

	return nil
}
