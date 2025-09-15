//go:build darwin

package runcmd

import (
	"fmt"
	"github.com/faanross/numinon/internal/models"
)

// macRunCmd implements the CommandRunCmd interface for Darwin.
type macRunCmd struct{}

// New is the constructor for our Darwin-specific RunCmd command
func New() CommandRunCmd {
	return &macRunCmd{}
}

func (mr *macRunCmd) DoRunCmd(args models.RunCmdArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ RUNCMD DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
