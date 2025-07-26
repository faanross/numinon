package upload

import "numinon_shadow/internal/models"

type CommandUpload interface {
	DoUpload(args models.UploadArgs) (models.AgentTaskResult, error)
}
