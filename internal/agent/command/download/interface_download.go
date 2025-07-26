package download

import "numinon_shadow/internal/models"

type CommandDownload interface {
	DoDownload(args models.DownloadArgs) (models.AgentTaskResult, error)
}
