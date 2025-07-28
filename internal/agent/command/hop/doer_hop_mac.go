//go:build darwin

package hop

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoHop(args models.HopArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ HOP DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
