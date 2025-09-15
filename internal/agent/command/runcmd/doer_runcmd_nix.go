//go:build linux

package runcmd

import (
	"fmt"
	"github.com/faanross/numinon/internal/models"
)

// nixRunCmd implements the CommandRunCmd interface for Linux.
type nixRunCmd struct{}

// New is the constructor for our Linux-specific RunCmd command
func New() CommandRunCmd {
	return &nixRunCmd{}
}

func (nd *nixRunCmd) DoRunCmd(args models.RunCmdArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ RUNCMD DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
