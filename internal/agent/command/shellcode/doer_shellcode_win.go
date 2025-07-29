//go:build windows

package shellcode

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// windowsShellcode implements the CommandShellcode interface for Windows.
type windowsShellcode struct{}

// New is the constructor for our Windows-specific Shellcode command
func New() CommandShellcode {
	return &windowsShellcode{}
}

func (ws *windowsShellcode) DoShellcode(args models.ShellcodeArgs) (models.AgentTaskResult, error) {
	fmt.Println("|✅ SHELLCODE DOER| The SHELLCODE command has been executed.")

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	result := models.AgentTaskResult{
		Status: "success",
		Output: []byte(output),
	}

	return result, nil
}
