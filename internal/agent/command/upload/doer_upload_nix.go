//go:build linux

package upload

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// nixUpload implements the CommandUpload interface for Linux.
type nixUpload struct{}

// New is the constructor for our Linux-specific Upload command
func New() CommandUpload {
	return &nixUpload{}
}

func (nd *nixUpload) DoUpload(args models.UploadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ UPLOAD DOER DARWIN| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
