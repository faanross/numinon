package enumerate

import "numinon_shadow/internal/models"

type CommandEnumerate interface {
	DoEnumerate(args models.EnumerateArgs) (models.AgentTaskResult, error)
}

