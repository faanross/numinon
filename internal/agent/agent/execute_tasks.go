package agent

import (
	"encoding/json"
	"fmt"
	"github.com/faanross/numinon/internal/models"
	"log"
)

// executeTask processes a received task, performs the action, and sends the result.
func (a *Agent) executeTask(task models.ServerTaskResponse) {
	log.Printf("|AGENT TASK|-> Executing Task ID: %s, Command: %s", task.TaskID, task.Command)

	var result models.AgentTaskResult

	orchestrator, found := a.commandOrchestrators[task.Command]

	if found {
		result = orchestrator(a, task)
	} else {
		log.Printf("|WARN AGENT TASK| Received unknown command: '%s' (ID: %s)", task.Command, task.TaskID)
		result = models.AgentTaskResult{
			TaskID: task.TaskID,
			Status: models.StatusFailureUnknownCommand,
			Error:  fmt.Sprintf("Agent does not recognize command: '%s'", task.Command),
		}
	}

	// Now marshall the result before sending it back using SendResult
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to marshal result for Task ID %s: %v", task.TaskID, err)
		return // Cannot send result if marshalling fails
	}

	// If we get here our AgentTaskResult struct has been marshalled as resultBytes
	// Now pass it to SendResult()
	log.Printf("|AGENT TASK|-> Sending result for Task ID %s (%d bytes)...", task.TaskID, len(resultBytes))
	err = a.communicator.SendResult(resultBytes)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to send result for Task ID %s: %v", task.TaskID, err)
	}

	log.Printf("|AGENT TASK|-> Successfully sent result for Task ID %s.", task.TaskID)
}
