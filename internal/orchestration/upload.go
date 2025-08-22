package orchestration

import (
	"encoding/json"
	"fmt"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/taskmanager"
)

// UploadOrchestrator handles the upload command lifecycle
type UploadOrchestrator struct {
	// in future add option to display result, save result, or both
}

// NewUploadOrchestrator creates a new runcmd orchestrator
func NewUploadOrchestrator() *UploadOrchestrator {
	return &UploadOrchestrator{}
}

// PrepareTask sets up upload-specific metadata
func (d *UploadOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// Validate arguments first

	if err := d.ValidateArgs(task.Arguments); err != nil {
		return fmt.Errorf("invalid upload arguments: %w", err)
	}

	return nil
}

// ProcessResult handles the upload result processing
// NOTE for now, since we just need to print a confirmation back on server post-processing
// This is just a stub
func (d *UploadOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {

	// For now not much we need to do here so this is just a stub
	// So we can satisfy the interface

	return nil
}

// ValidateArgs checks if download arguments are valid
func (d *UploadOrchestrator) ValidateArgs(args json.RawMessage) error {
	var uploadArgs models.UploadArgs

	// For now just going to keep things extremely simple - just going to ensure all values exists
	// TODO add more sophisticated validation to ensure args are valid

	if err := json.Unmarshal(args, &uploadArgs); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	if uploadArgs.TargetDirectory == "" {
		return fmt.Errorf("TargetDirectory is required")
	}

	if uploadArgs.TargetFilename == "" {
		return fmt.Errorf("TargetFilename is required")
	}

	if uploadArgs.FileContentBase64 == "" {
		return fmt.Errorf("FileContentBase64 is required")
	}

	if uploadArgs.ExpectedSha256 == "" {
		return fmt.Errorf("ExpectedSha256 is required")
	}

	return nil
}
