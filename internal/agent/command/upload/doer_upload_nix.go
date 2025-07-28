//go:build linux

package upload

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoUpload(args models.UploadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ UPLOAD DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
