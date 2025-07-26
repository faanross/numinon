package runcmd

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoRunCmd(args models.RunCmdArgs) (models.AgentTaskResult, error) {
	fmt.Println("|✅ RUN_CMD DOER| The RUN_CMD command has been executed.")

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	result := models.AgentTaskResult{
		Status: "success",
		Output: []byte(output),
	}

	return result, nil
}
