package orchestration

import (
	"encoding/json"
	"fmt"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/taskmanager"
)

// ShellcodeOrchestrator handles the shellcode command lifecycle
type ShellcodeOrchestrator struct {
	// in future add option to display result, save result, or both
}

// NewShellcodeOrchestrator creates a new shellcode orchestrator
func NewShellcodeOrchestrator() *ShellcodeOrchestrator {
	return &ShellcodeOrchestrator{}
}

// PrepareTask sets up shellcode-specific metadata
func (d *ShellcodeOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// Validate arguments first
	if err := d.ValidateArgs(task.Arguments); err != nil {
		return fmt.Errorf("invalid shellcode arguments: %w", err)
	}

	// RunCmd has no additional preparation work to do, so just validation is good

	return nil
}

// ValidateArgs checks if shellcode arguments are valid
func (d *ShellcodeOrchestrator) ValidateArgs(args json.RawMessage) error {
	var shellcodeArgs models.ShellcodeArgs

	if err := json.Unmarshal(args, &shellcodeArgs); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	if shellcodeArgs.ShellcodeBase64 == "" {
		return fmt.Errorf("no shellcode loaded, required")
	}

	if shellcodeArgs.TargetPID != 0 {
		return fmt.Errorf("at present, only process auto-injection is supported - please select 0 as PID")
	}

	return nil
}

// ProcessResult handles the shellcode results processing
func (d *ShellcodeOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {

	// For now not much we need to do here so this is just a stub
	// So we can satisfy the interface

	return nil
}
