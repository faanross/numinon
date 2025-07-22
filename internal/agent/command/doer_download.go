package command

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoDownload(args models.DownloadArgs) (models.AgentTaskResult, error) {
	fmt.Println("The DOWNLOAD command has been executed.")

	fmt.Printf("The download path called by the command is: %s\n", args.SourceFilePath)

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
