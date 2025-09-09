package router

import (
	"log"
	"numinon_shadow/internal/listener"
	"numinon_shadow/internal/orchestration"
	"numinon_shadow/internal/taskmanager"
	"numinon_shadow/internal/tracker"
)

// Global instances for task and orchestration management
var TaskManager taskmanager.TaskManager
var Orchestrators *orchestration.Registry

// InitializeTaskManagementWithFullSupport sets up WITH both listener manager AND tracking
func InitializeTaskManagementWithFullSupport(listenerMgr *listener.Manager, agentTracker *tracker.Tracker) {
	// Create task manager
	TaskManager = taskmanager.NewMemoryTaskStore()

	// Create orchestrator registry
	Orchestrators = orchestration.NewRegistry()

	// Register orchestrators WITH BOTH listener manager and tracker
	registerAllOrchestrators(listenerMgr, agentTracker)

	log.Println("|📋 TASK MGR| Task management initialized (full mode - HOP with agent tracking)")
}

// registerAllOrchestrators registers all command orchestrators
// This is the common registration logic used by all initialization paths
func registerAllOrchestrators(listenerMgr *listener.Manager, agentTracker *tracker.Tracker) {
	// Create standard orchestrators (these don't need special infrastructure)
	downloadOrch := orchestration.NewDownloadOrchestrator("./downloads")
	uploadOrch := orchestration.NewUploadOrchestrator()
	runCmdOrch := orchestration.NewRunCmdOrchestrator()
	shellcodeOrch := orchestration.NewShellcodeOrchestrator()
	enumerateOrch := orchestration.NewEnumerationOrchestrator()
	morphOrch := orchestration.NewMorphOrchestrator()

	// Create HOP orchestrator with whatever support is available
	// The HOP orchestrator should handle nil gracefully
	hopOrch := orchestration.NewHopOrchestrator(listenerMgr, agentTracker)

	// Register all orchestrators
	Orchestrators.Register("upload", uploadOrch)
	Orchestrators.Register("download", downloadOrch)
	Orchestrators.Register("run_cmd", runCmdOrch)
	Orchestrators.Register("shellcode", shellcodeOrch)
	Orchestrators.Register("enumerate", enumerateOrch)
	Orchestrators.Register("morph", morphOrch)
	Orchestrators.Register("hop", hopOrch)
}
