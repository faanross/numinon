//go:build darwin

package download

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoDownload(args models.DownloadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|❗ DOWNLOAD DOER DARWIN| This feature has not yet been implemented for Darwin OS.")

	result := models.AgentTaskResult{
		Status: "FAILURE",
	}
	return result, nil
}
