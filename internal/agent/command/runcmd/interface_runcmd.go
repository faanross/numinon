package runcmd

import "github.com/faanross/numinon/internal/models"

type CommandRunCmd interface {
	DoRunCmd(args models.RunCmdArgs) (models.RunCmdResult, error)
}
