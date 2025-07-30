//go:build linux

package download

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// nixDownload implements the CommandDownload interface for Linux.
type nixDownload struct{}

// New is the constructor for our Linux-specific Download command
func New() CommandDownload {
	return &nixDownload{}
}

func (nd *nixDownload) DoDownload(args models.DownloadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ DOWNLOAD DOER DARWIN| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
