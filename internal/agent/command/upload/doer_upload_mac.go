//go:build darwin

package upload

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoUpload(args models.UploadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ UPLOAD DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
