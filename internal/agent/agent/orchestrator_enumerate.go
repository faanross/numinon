package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"numinon_shadow/internal/agent/command/enumerate"
	"numinon_shadow/internal/models"
)

// orchestrateEnumerate is the orchestrator for the enumerate command.
func (a *Agent) orchestrateEnumerate(task models.ServerTaskResponse) models.AgentTaskResult {

	// Create an instance of the command-specific args struct
	var args models.EnumerateArgs

	// ServerTaskResponse.data contains the command-specific args, so now we unmarshall the field into the struct
	if err := json.Unmarshal(task.Data, &args); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal EnumerateArgs for Task ID %s: %v. Raw Data: %s", task.TaskID, err, string(task.Data))
		log.Printf("|❗ERR ENUMERATE ORCHESTRATOR| %s", errMsg)
		return models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureUnmarshallError,
			Error:  errMsg,
		}
	}

	log.Printf("|✅ ENUMERATE ORCHESTRATOR| Task ID: %s. Orchestrating enumeration for process: '%s'",
		task.TaskID, args.ProcessName)

	// Call the "doer" function
	commandEnumerate := enumerate.New()                        // create os-specific Download struct ("decided" when compiled)
	enumerateResult, err := commandEnumerate.DoEnumerate(args) // Call the interface method

	finalResult := models.AgentTaskResult{
		TaskID: task.TaskID,
	}

	if err != nil {
		log.Printf("|❗ERR ENUMERATE ORCHESTRATOR | Execution error for TaskID %s from enumerator: %v. Message: %s",
			task.TaskID, err, enumerateResult.Message)
		finalResult.Status = models.StatusFailureExecutionError // Could be more specific, e.g., models.StatusFailureEnumerationError
		finalResult.Error = err.Error()
		// Include message from doer if it's useful error context
		if enumerateResult.Message != "" && finalResult.Error != enumerateResult.Message {
			finalResult.Output = []byte(fmt.Sprintf("Details: %s", enumerateResult.Message))
		} else if enumerateResult.Message != "" {
			finalResult.Output = []byte(enumerateResult.Message)
		}
	} else {
		log.Printf("|✅ ENUMERATE ORCHESTRATOR | Successfully enumerated processes for TaskID %s. Found: %d. Message: %s",
			task.TaskID, len(enumerateResult.Processes), enumerateResult.Message)

		jsonData, marshalErr := json.Marshal(enumerateResult.Processes)
		if marshalErr != nil {
			log.Printf("|❗ERR ENUMERATE ORCHESTRATOR | Failed to marshal process list for TaskID %s: %v", task.TaskID, marshalErr)
			finalResult.Status = models.StatusFailureExecutionError // Or a "failure_marshal_result"
			finalResult.Error = fmt.Sprintf("agent failed to marshal process list: %v", marshalErr)
		} else {
			finalResult.Status = models.StatusSuccess
			finalResult.Output = jsonData
			// include enumCmdResult.Message if it provides useful success context
			if enumerateResult.Message != "" {
				finalResult.Error = enumerateResult.Message
			}
		}
	}
	return finalResult

}
