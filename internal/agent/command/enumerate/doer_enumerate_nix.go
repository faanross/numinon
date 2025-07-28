//go:build linux

package enumerate

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoEnumerate(args models.EnumerateArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ ENUMERATE DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
