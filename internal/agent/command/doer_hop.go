package command

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoHop(args models.HopArgs) (models.AgentTaskResult, error) {
	fmt.Println("|✅ HOP DOER| The HOP command has been executed.")

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	result := models.AgentTaskResult{
		Status: "success",
		Output: []byte(output),
	}

	return result, nil
}
