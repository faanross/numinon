//go:build darwin

package enumerate

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoEnumerate(args models.EnumerateArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ ENUMERATE DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
