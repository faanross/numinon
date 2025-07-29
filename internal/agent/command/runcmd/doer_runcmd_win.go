//go:build windows

package runcmd

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// windowsRunCmd implements the CommandRunCmd interface for Windows.
type windowsRunCmd struct{}

// New is the constructor for our Windows-specific Hop command
func New() CommandRunCmd {
	return &windowsRunCmd{}
}

func (wr *windowsRunCmd) DoRunCmd(args models.RunCmdArgs) (models.AgentTaskResult, error) {
	fmt.Println("|✅ RUN_CMD DOER| The RUN_CMD command has been executed.")

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	result := models.AgentTaskResult{
		Status: "success",
		Output: []byte(output),
	}

	return result, nil
}
