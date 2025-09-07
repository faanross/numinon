package orchestration

import (
	"encoding/json"
	"fmt"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/taskmanager"
)

// HopOrchestrator handles the Hop command lifecycle
type HopOrchestrator struct {
	// in future add option to display result, save result, or both
}

// NewHopOrchestrator creates a new Hop orchestrator
func NewHopOrchestrator() *HopOrchestrator {
	return &HopOrchestrator{}
}

// PrepareTask sets up Hop-specific metadata
func (d *HopOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// Validate arguments first
	if err := d.ValidateArgs(task.Arguments); err != nil {
		return fmt.Errorf("invalid Hop arguments: %w", err)
	}

	// Morph has no additional preparation work to do, so just validation is good

	return nil
}

// ValidateArgs checks if Hop arguments are valid
func (d *HopOrchestrator) ValidateArgs(args json.RawMessage) error {

	var hopArgs models.HopArgs

	if err := json.Unmarshal(args, &hopArgs); err != nil {
	}

	return nil
}

// ProcessResult handles the Hop results processing
func (d *HopOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {

	// For now not much we need to do here so this is just a stub
	// So we can satisfy the interface

	return nil
}
