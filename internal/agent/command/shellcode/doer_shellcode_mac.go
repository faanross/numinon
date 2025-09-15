//go:build darwin

package shellcode

import (
	"fmt"
	"github.com/faanross/numinon/internal/models"
)

// macShellcode implements the CommandShellcode interface for Darwin.
type macShellcode struct{}

// New is the constructor for our Darwin-specific Shellcode command
func New() CommandShellcode {
	return &macShellcode{}
}

func (mc *macShellcode) DoShellcode(args models.ShellcodeArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ SHELLCODE DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
