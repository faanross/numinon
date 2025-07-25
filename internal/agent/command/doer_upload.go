package command

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoUpload(args models.UploadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|✅ UPLOAD DOER| The UPLOAD command has been executed.")

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	result := models.AgentTaskResult{
		Status: "success",
		Output: []byte(output),
	}

	return result, nil
}
