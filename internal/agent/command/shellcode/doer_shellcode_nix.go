//go:build linux

package shellcode

import (
	"fmt"
	"github.com/faanross/numinon/internal/models"
)

// nixShellcode implements the CommandShellcode interface for Linux.
type nixShellcode struct{}

// New is the constructor for our Linux-specific Shellcode command
func New() CommandShellcode {
	return &nixShellcode{}
}

func (ns *nixShellcode) DoShellcode(args models.ShellcodeArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ SHELLCODE DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
