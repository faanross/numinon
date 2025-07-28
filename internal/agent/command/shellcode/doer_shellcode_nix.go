//go:build linux

package shellcode

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoShellcode(args models.ShellcodeArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ SHELLCODE DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
