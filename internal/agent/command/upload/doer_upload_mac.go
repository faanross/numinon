//go:build darwin

package upload

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// macUpload implements the CommandUpload interface for Darwin.
type macUpload struct{}

// New is the constructor for our Darwin-specific Download command
func New() CommandUpload {
	return &macUpload{}
}

func (mu *macUpload) DoUpload(args models.UploadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ UPLOAD DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
