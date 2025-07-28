//go:build darwin

package runcmd

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoRunCmd(args models.RunCmdArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ RUNCMD DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
