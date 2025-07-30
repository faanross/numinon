//go:build darwin

package enumerate

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// macEnumerate implements the CommandEnumerate interface for Darwin.
type macEnumerate struct{}

// New is the constructor for our Darwin-specific Enumerate command
func New() CommandEnumerate {
	return &macEnumerate{}
}

func (me *macEnumerate) DoEnumerate(args models.EnumerateArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ ENUMERATE DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
