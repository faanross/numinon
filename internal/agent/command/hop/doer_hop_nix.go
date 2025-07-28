//go:build linux

package hop

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoHop(args models.HopArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ HOP DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
