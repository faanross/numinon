package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/command/upload"
	"numinon_shadow/internal/models"
	"strings"
)

// orchestrateUpload is the orchestrator for the UPLOAD command.
func (a *Agent) orchestrateUpload(task models.ServerTaskResponse) models.AgentTaskResult {

	// Create an instance of the command-specific args struct
	var args models.UploadArgs

	// ServerTaskResponse.data contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal UploadArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))
		log.Printf("|❗ERR UPLOAD ORCHESTRATOR| %s", errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureUnmarshallError,
			Error:  errMsg,
		}
	}

	log.Printf("|✅ UPLOAD ORCHESTRATOR| Task ID: %s. Orchestrating Upload of %s to %s",
		task.TaskID, args.TargetFilename, args.TargetDirectory)

	// Call the "doer" function
	uploadResult, err := upload.DoUpload(args) // create os-specific Download struct ("decided" when compiled)

	// REMEMBER uploadResult is models.UploadResult
	// WE NEED TO WRAP IT IN models.AgentTaskResult.Output

	// Prepare the final TaskResult
	finalResult := models.AgentTaskResult{
		TaskID: task.TaskID, // Hash of the raw file content
		Output: nil,         // Will be process info content on success
		Error:  "",          // Will be error message on failure
	}

	if err != nil {
		finalResult.Error = err.Error()

		log.Printf("|❗ERR UPLOAD ORCHESTRATOR| Upload execution failed for Task ID %s: %s.",
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

		outputJSON, err := json.Marshal(uploadResult)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to marshal UploadResult for Task ID %s: %v", task.TaskID, err)
			log.Printf("|❗ERR UPLOAD ORCHESTRATOR| %s", errMsg)
			return models.AgentTaskResult{
				TaskID: task.TaskID,
				Status: models.StatusFailureUnmarshallError,
				Error:  errMsg,
			}
		}

		finalResult.Output = outputJSON

		finalResult.Status = models.StatusSuccess
		log.Printf("|✅ UPLOAD ORCHESTRATOR| Execution successful for Task ID %s.",
			task.TaskID)
	}
	return finalResult
}
