//go:build linux

package download

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoDownload(args models.DownloadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ DOWNLOAD DOER LINUX| This feature has not yet been implemented for Linux OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
