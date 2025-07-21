package agent

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/command"
	"numinon_shadow/internal/models"
	"strings"
)

// orchestrateDownload is the orchestrator for the "download" command.
func (a *Agent) orchestrateDownload(task models.ServerTaskResponse) models.AgentTaskResult {

	// Create an instance of the command-specific args struct
	var args models.DownloadArgs

	// ServerTaskResponse.data contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal DownloadFileArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))
		log.Printf("|❗ERR DOWNLOAD_FILE HANDLER| %s", errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: "FAILED TO UNMARSHALL DATA FIELD", // We can later create a shared common error type system
			Error:  errMsg,
		}
	}

	log.Printf("|AGENT DOWNLOAD ORCHESTRATOR| Task ID: %s. Orchestrating download from agent path: '%s'",
		task.TaskID, args.SourceFilePath)

	// Call the "doer" function
	downloadResult := command.DoDownload(args, task.TaskID)

	//

	// Prepare the final TaskResult
	finalResult := models.AgentTaskResult{
		TaskID:     task.TaskID,
		FileSha256: downloadResult.FileSha256, // Hash of the raw file content
		Output:     nil,                       // Will be base64 content on success
		Error:      "",                        // Will be error message on failure
	}

	if err != nil {
		finalResult.Error = err.Error()
		finalResult.Output = []byte(downloadResult.Message) // Send back any message from doer
		log.Printf("|❗ERR DOWNLOAD_FILE HANDLER| Download execution failed for Task ID %s: %s. Detailed Message: %s",
			task.TaskID, finalResult.Error, downloadResult.Message)

		errorString := finalResult.Error
		switch {
		case strings.Contains(errorString, "validation:"):
			finalResult.Status = models.StatusFailureInvalidArgs
		case strings.Contains(errorString, "File not found"):
			finalResult.Status = models.StatusFailureFileNotFound // New status
		case strings.Contains(errorString, "Permission denied"):
			finalResult.Status = models.StatusFailurePermissionDenied
		default:
			finalResult.Status = models.StatusFailureReadError // General read error
		}
	} else {
		// Success from download.Execute()
		// Base64 encode the raw file bytes for transport
		encodedContent := base64.StdEncoding.EncodeToString(downloadOpResult.RawFileBytes)
		finalResult.Output = []byte(encodedContent)
		finalResult.Status = models.StatusSuccess
		log.Printf("|AGENT TASK DOWNLOAD_FILE HANDLER| Execution successful for Task ID %s. Sending %d base64 encoded bytes. Message: %s",
			task.TaskID, len(finalResult.Output), downloadOpResult.Message)
	}
	return finalResult
}
