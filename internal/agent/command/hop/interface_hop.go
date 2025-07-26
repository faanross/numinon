package hop

import "numinon_shadow/internal/models"

type CommandHop interface {
	DoHop(args models.HopArgs) (models.AgentTaskResult, error)
}
