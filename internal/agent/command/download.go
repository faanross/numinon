package command

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func ExecuteDownload(task models.ServerTaskResponse) models.AgentTaskResult {
	fmt.Printf("The following command has been executed: %s\n", task.Command)
	return models.AgentTaskResult{
		TaskID: task.TaskID,
		Status: "success",
		Output: []byte(fmt.Sprintln("Called the Download command")),
	}
}
