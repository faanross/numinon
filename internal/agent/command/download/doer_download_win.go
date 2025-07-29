//go:build windows

package download

import (
	"fmt"
	"numinon_shadow/internal/models"
)

// windowsDownload implements the CommandDownload interface for Windows.
type windowsDownload struct{}

// New is the constructor for our Windows-specific Download command
func New() CommandDownload {
	return &windowsDownload{}
}

func (wd *windowsDownload) DoDownload(args models.DownloadArgs) (models.AgentTaskResult, error) {
	fmt.Println("|✅ DOWNLOAD DOER| The DOWNLOAD command has been executed.")

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	sha256hash := fmt.Sprintf("thisisnotarealhashjustaplaceholderchillbruh")

	result := models.AgentTaskResult{
		Status:     "success",
		Output:     []byte(output),
		FileSha256: sha256hash,
	}

	return result, nil
}
