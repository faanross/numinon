package runcmd

import "numinon_shadow/internal/models"

type CommandRunCmd interface {
	DoRunCmd(args models.RunCmdArgs) (models.AgentTaskResult, error)
}
