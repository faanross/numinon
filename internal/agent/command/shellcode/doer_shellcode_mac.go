//go:build darwin

package shellcode

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoShellcode(args models.ShellcodeArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ SHELLCODE DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
