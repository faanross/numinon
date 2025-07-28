//go:build linux

package morph

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoMorph(args models.MorphArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ MORPH DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
