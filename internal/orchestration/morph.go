package orchestration

import (
	"encoding/json"
	"fmt"
	"github.com/faanross/numinon/internal/models"
	"github.com/faanross/numinon/internal/taskmanager"
	"time"
)

// MorphOrchestrator handles the Morph command lifecycle
type MorphOrchestrator struct {
	// in future add option to display result, save result, or both
}

// NewMorphOrchestrator creates a new Morph orchestrator
func NewMorphOrchestrator() *MorphOrchestrator {
	return &MorphOrchestrator{}
}

// PrepareTask sets up Morph-specific metadata
func (d *MorphOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// Validate arguments first
	if err := d.ValidateArgs(task.Arguments); err != nil {
		return fmt.Errorf("invalid Morph arguments: %w", err)
	}

	// Morph has no additional preparation work to do, so just validation is good

	return nil
}

// ValidateArgs checks if Morph arguments are valid
func (d *MorphOrchestrator) ValidateArgs(args json.RawMessage) error {

	var morphArgs models.MorphArgs

	if err := json.Unmarshal(args, &morphArgs); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	newDelay := *morphArgs.NewDelay

	newDelayDuration, parseErr := time.ParseDuration(newDelay)

	if parseErr != nil {
		return fmt.Errorf("BaseSleep update failed: invalid duration format '%s'. Error: %v", morphArgs.NewDelay, parseErr)
	}

	if newDelayDuration <= 0 {
		return fmt.Errorf("BaseSleep update failed, must be a positive duration, current valueinvalid duration format '%s'. Error: %v", *morphArgs.NewDelay, parseErr)
	}

	newJitter := *morphArgs.NewJitter

	if newJitter < 0.0 || newJitter > 1.0 {

		return fmt.Errorf("jitter update failed: value %f is out of acceptable range [0.0 - 1.0]", newJitter)
	}

	return nil
}

// ProcessResult handles the Morph results processing
func (d *MorphOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {

	// For now not much we need to do here so this is just a stub
	// So we can satisfy the interface

	return nil
}
