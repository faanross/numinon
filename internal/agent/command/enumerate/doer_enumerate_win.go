//go:build windows

package enumerate

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// windowsEnumerate implements the CommandEnumerate interface for Windows.
type windowsEnumerate struct{}

// New is the constructor for our Windows-specific Enumerate command
func New() CommandEnumerate {
	return &windowsEnumerate{}
}

func (we *windowsEnumerate) DoEnumerate(args models.EnumerateArgs) (models.AgentTaskResult, error) {
	fmt.Println("|✅ ENUMERATE DOER| The ENUMERATE command has been executed.")

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	result := models.AgentTaskResult{
		Status: "success",
		Output: []byte(output),
	}

	return result, nil
}
