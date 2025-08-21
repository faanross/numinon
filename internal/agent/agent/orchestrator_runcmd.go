package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/command/runcmd"
	"numinon_shadow/internal/models"
	"strings"
)

// orchestrateRunCmd is the orchestrator for the RUN_CMD command.
func (a *Agent) orchestrateRunCmd(task models.ServerTaskResponse) models.AgentTaskResult {

	// Create an instance of the command-specific args struct
	var args models.RunCmdArgs

	// ServerTaskResponse.data contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal RunCmdArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))
		log.Printf("|❗ERR RUN_CMD ORCHESTRATOR| %s", errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureUnmarshallError,
			Error:  errMsg,
		}
	}

	log.Printf("|✅ RUN_CMD ORCHESTRATOR| Task ID: %s. Shell Command Executed: %s",
		task.TaskID, args.CommandLine)

	// Call the "doer" function
	commandRunCmd := runcmd.New()                     // create os-specific Download struct ("decided" when compiled)
	runCmdResult, err := commandRunCmd.DoRunCmd(args) // Call the interface method

	finalResult := models.AgentTaskResult{
		TaskID: task.TaskID,
		// Output will be set below after JSON encoding
	}

	// Add this right after creating finalResult:
	outputJSON, _ := json.Marshal(string(runCmdResult.CombinedOutput))
	finalResult.Output = outputJSON

	// Note: The 'execErr' from runner.Execute() is for catastrophic failures of the runner itself,
	// not for command execution errors (which are in cmdResult.SystemError or cmdResult.ExitCode).
	// Most of the time, execErr will be nil if the Execute method could attempt to run something.
	if err != nil { // This would be an unexpected error from the Execute method itself
		log.Printf("|❗CRIT RUN_CMD HANDLER| Unexpected error from runner.Execute for TaskID %s: %v", task.TaskID, err)
		finalResult.Status = models.StatusFailureExecutionError
		finalResult.Error = fmt.Sprintf("Internal runner error: %v. SystemMsg: %s. CommandMsg: %s", err, runCmdResult.SystemError, runCmdResult.CommandError)
		return finalResult
	}

	// Process based on cmdResult fields
	if runCmdResult.SystemError != "" {
		log.Printf("|❗ERR RUN_CMD HANDLER| System error for TaskID %s: %s", task.TaskID, runCmdResult.SystemError)
		finalResult.Error = runCmdResult.SystemError
		if strings.Contains(runCmdResult.SystemError, "timed out") {
			finalResult.Status = models.StatusFailureTimeout
		} else if strings.Contains(runCmdResult.SystemError, "validation:") {
			finalResult.Status = models.StatusFailureInvalidArgs
		} else {
			finalResult.Status = models.StatusFailureExecError
		}
	} else if runCmdResult.ExitCode != 0 {
		log.Printf("|INFO RUN_CMD HANDLER| Command for TaskID %s exited with code %d. Output may contain errors.", task.TaskID, runCmdResult.ExitCode)
		finalResult.Status = models.StatusSuccessExitNonZero
		finalResult.Error = fmt.Sprintf("Command exited with code %d. %s", runCmdResult.ExitCode, runCmdResult.CommandError)
	} else {
		log.Printf("|✅ RUN_CMD HANDLER| Command for TaskID %s executed successfully. Output length: %d", task.TaskID, len(runCmdResult.CombinedOutput))
		finalResult.Status = models.StatusSuccess
	}

	return finalResult

}
