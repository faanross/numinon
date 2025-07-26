package morph

import "numinon_shadow/internal/models"

type CommandMorph interface {
	DoMorph(args models.MorphArgs) (models.AgentTaskResult, error)
}
