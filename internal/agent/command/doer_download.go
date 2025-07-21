package command

import (
	"fmt"
	"numinon_shadow/internal/models"
)

func DoDownload(args models.DownloadArgs, taskid string) models.AgentTaskResult {
	fmt.Println("The DOWNLOAD command has been executed.")

	fmt.Printf("The download path called by the command is: %s\n", args.SourceFilePath)

	output := fmt.Sprintln("Let's just assume for now it succeeded, will implement later.")

	fmt.Println(output)

	return models.AgentTaskResult{
		TaskID: taskid,
		Status: "success",
		Output: []byte(output),
	}
}
