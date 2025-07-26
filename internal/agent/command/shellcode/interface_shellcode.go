package shellcode

import "numinon_shadow/internal/models"

type CommandShellcode interface {
	DoShellcode(args models.ShellcodeArgs) (models.AgentTaskResult, error)
}
