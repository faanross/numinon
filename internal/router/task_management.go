package router

import (
	"log"
	"numinon_shadow/internal/listener"
	"numinon_shadow/internal/orchestration"
	"numinon_shadow/internal/taskmanager"
)

// Global instances for task and orchestration management
var TaskManager taskmanager.TaskManager
var Orchestrators *orchestration.Registry

// InitializeTaskManagement sets up the task management and orchestration system
func InitializeTaskManagement(listenerMgr *listener.Manager) {
	// Create task manager
	TaskManager = taskmanager.NewMemoryTaskStore()

	// Create orchestrator registry
	Orchestrators = orchestration.NewRegistry()

	// Register command-specific orchestrators
	downloadOrch := orchestration.NewDownloadOrchestrator("./downloads")
	uploadOrch := orchestration.NewUploadOrchestrator()
	runCmdOrch := orchestration.NewRunCmdOrchestrator()
	shellcodeOrch := orchestration.NewShellcodeOrchestrator()
	enumerateOrch := orchestration.NewEnumerationOrchestrator()
	morphOrch := orchestration.NewMorphOrchestrator()
	hopOrch := orchestration.NewHopOrchestrator(listenerMgr)

	Orchestrators.Register("upload", uploadOrch)
	Orchestrators.Register("download", downloadOrch)
	Orchestrators.Register("run_cmd", runCmdOrch)
	Orchestrators.Register("shellcode", shellcodeOrch)
	Orchestrators.Register("enumerate", enumerateOrch)
	Orchestrators.Register("morph", morphOrch)
	Orchestrators.Register("hop", hopOrch)

	log.Println("|📋 TASK MGR| Task management and orchestration system initialized")
}
