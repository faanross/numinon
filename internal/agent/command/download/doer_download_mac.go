//go:build darwin

package download

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// macDownload implements the CommandDownload interface for Darwin.
type macDownload struct{}

// New is the constructor for our Darwin-specific Download command
func New() CommandDownload {
	return &macDownload{}
}

func (md *macDownload) DoDownload(args models.DownloadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ DOWNLOAD DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
