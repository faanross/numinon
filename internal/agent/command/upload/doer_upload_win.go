//go:build windows

package upload

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// windowsUpload implements the CommandUpload interface for Windows.
type windowsUpload struct{}

// New is the constructor for our Windows-specific Upload command
func New() CommandUpload {
	return &windowsUpload{}
}

func (wu *windowsUpload) DoUpload(args models.UploadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|✅ UPLOAD DOER| The UPLOAD command has been executed.")

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	result := models.AgentTaskResult{
		Status: "success",
		Output: []byte(output),
	}

	return result, nil
}
