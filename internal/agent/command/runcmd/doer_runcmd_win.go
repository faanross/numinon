//go:build windows

package runcmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/faanross/numinon/internal/models"
	"log"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultCommandTimeout = 60 * time.Second
)

// windowsRunCmd implements the CommandRunCmd interface for Windows.
type windowsRunCmd struct{}

// New is the constructor for our Windows-specific Hop command
func New() CommandRunCmd {
	return &windowsRunCmd{}
}

// isPowerShellDefaultWindows is a helper for Windows default shell choice.
// if ps is intended to be default - hardcode to true, if not false
// TODO more elegant way to handle this
func isPowerShellDefaultWindows() bool {
	// On Windows, prefer PowerShell if no shell is specified.
	// A more robust check could verify if 'powershell.exe' is in PATH.
	return true
}

func (wr *windowsRunCmd) DoRunCmd(args models.RunCmdArgs) (models.RunCmdResult, error) {
	fmt.Println("|✅ RUN_CMD DOER| The RUN_CMD command has been executed.")
	fmt.Printf("|📋 RUN_CMD DETAILS| CommandLine='%s', Shell='%s'\n", args.CommandLine, args.Shell)

	result := models.RunCmdResult{
		ExitCode: -1, // Default to fail, we'll update accordingly if succeeded
	}

	// Basic Validation
	if strings.TrimSpace(args.CommandLine) == "" {
		result.SystemError = "validation: CommandLine cannot be empty"
		log.Printf("|❗ERR RUNCMD DOER| %s", result.SystemError)
		return result, fmt.Errorf(result.SystemError)
	}

	var cmdPath string
	var cmdArgs []string
	shellToUse := strings.ToLower(strings.TrimSpace(args.Shell))

	// construct specific arguments based on shell selection
	if shellToUse == "powershell" || shellToUse == "ps" || (shellToUse == "" && isPowerShellDefaultWindows()) {
		cmdPath = "powershell.exe"
		cmdArgs = []string{"-NoProfile", "-NonInteractive", "-NoLogo", "-Command", args.CommandLine}
		log.Printf("|⚙️ RUNCMD ACTION| Using PowerShell: %s %s", cmdPath, strings.Join(cmdArgs, " "))
	} else if shellToUse == "cmd" || shellToUse == "" {
		cmdPath = "cmd.exe"
		cmdArgs = []string{"/c", args.CommandLine}
		log.Printf("|⚙️ RUNCMD ACTION| Using CMD: %s %s", cmdPath, strings.Join(cmdArgs, " "))
	} else {
		result.SystemError = fmt.Sprintf("unsupported shell '%s' for Windows", args.Shell)
		log.Printf("|❗ERR RUNCMD DOER| %s", result.SystemError)
		return result, fmt.Errorf(result.SystemError)
	}

	// context to be able to cancel cmd if timeout exceeds
	ctx, cancel := context.WithTimeout(context.Background(), defaultCommandTimeout)
	defer cancel()

	// construct the command (NOTE: does not execute yet, need to call Run())
	cmd := exec.CommandContext(ctx, cmdPath, cmdArgs...)

	// Buffers to capture results
	var combinedOutputBuffer bytes.Buffer
	cmd.Stdout = &combinedOutputBuffer
	cmd.Stderr = &combinedOutputBuffer

	// Finally execute actuall shell command using Run() and capture Stdout and Stderr below
	log.Printf("|⚙️ RUNCMD ACTION| Executing: %s %v", cmd.Path, cmd.Args)
	execErr := cmd.Run() // Store the error from cmd.Run()

	result.CombinedOutput = combinedOutputBuffer.Bytes()

	// Cancel execution of timeout exceeded specified in defaultCommandTimeout
	if ctx.Err() == context.DeadlineExceeded {
		result.SystemError = "command timed out after " + defaultCommandTimeout.String()
		log.Printf("|⚠️ WARN RUNCMD DOER| %s for command: %s", result.SystemError, args.CommandLine)
		// No top-level error from Execute, details are in RunCommandResult
		return result, nil
	}

	if execErr != nil {
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.CommandError = fmt.Sprintf("command exited with code %d", result.ExitCode)
			log.Printf("|⚙️ RUNCMD ACTION| Command '%s' exited with code %d.", args.CommandLine, result.ExitCode)
		} else {
			result.SystemError = fmt.Sprintf("failed to start or run command: %v", execErr)
			log.Printf("|❗ERR RUNCMD DOER| Failed to execute command '%s': %v", args.CommandLine, execErr)
		}
		// No top-level error from Execute if command ran but failed; details in RunCommandResult
		return result, nil
	}

	// Command was successful (exit code 0)
	result.ExitCode = 0
	log.Printf("|👊 RUNCMD SUCCESS| Command '%s' executed successfully (exit code 0). Output length: %d", args.CommandLine, len(result.CombinedOutput))
	return result, nil // Success, no top-level error

}
