package shellcode

import "numinon_shadow/internal/models"

type CommandShellcode interface {
	DoShellcode(dllBytes []byte, targetPID uint32, shellcodeArgs []byte, exportName string) (models.ShellcodeResult, error)
}
