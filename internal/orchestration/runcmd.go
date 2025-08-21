package orchestration

import (
	"encoding/json"
	"fmt"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/taskmanager"
)

// RunCmdOrchestrator handles the runcmd command lifecycle
type RunCmdOrchestrator struct {
	// in future add option to display result, save result, or both
}

// NewRunCmdOrchestrator creates a new runcmd orchestrator
func NewRunCmdOrchestrator() *RunCmdOrchestrator {
	return &RunCmdOrchestrator{}
}

// PrepareTask sets up runcmd-specific metadata
func (d *RunCmdOrchestrator) PrepareTask(task *taskmanager.Task) error {
	// Validate arguments first
	if err := d.ValidateArgs(task.Arguments); err != nil {
		return fmt.Errorf("invalid runcmd arguments: %w", err)
	}

	return nil
}

// ProcessResult handles the runcmd
func (d *RunCmdOrchestrator) ProcessResult(task *taskmanager.Task, result json.RawMessage) error {

	// For now not much we need to do here so this is just a stub
	// So we can satisfy the interface

	return nil
}

// ValidateArgs checks if download arguments are valid
func (d *RunCmdOrchestrator) ValidateArgs(args json.RawMessage) error {
	var runCmdArgs models.RunCmdArgs

	if err := json.Unmarshal(args, &runCmdArgs); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	// if shell is selection is not valid, default to powershell
	// TODO, set value equal to default shell selection specified in server config file
	// TODO, when implementing DARWIN and/or NIX, this has to be expanded upon
	if runCmdArgs.Shell != "cmd" && runCmdArgs.Shell != "powershell" && runCmdArgs.Shell != "ps" {
		runCmdArgs.Shell = "powershell"
	}

	// right now very simply just check if its empty
	// TODO might consider some type of allowlist reference to ensure it's a legit shell command before sending off
	if runCmdArgs.CommandLine == "" {
		return fmt.Errorf("command line is required")
	}

	return nil
}
