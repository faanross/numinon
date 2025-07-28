//go:build linux

package runcmd

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoRunCmd(args models.RunCmdArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ RUNCMD DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
