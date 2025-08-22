package router

import (
	"log"
	"numinon_shadow/internal/orchestration"
	"numinon_shadow/internal/taskmanager"
)

// Global instances for task and orchestration management
var TaskManager taskmanager.TaskManager
var Orchestrators *orchestration.Registry

// InitializeTaskManagement sets up the task management and orchestration system
func InitializeTaskManagement() {
	// Create task manager
	TaskManager = taskmanager.NewMemoryTaskStore()

	// Create orchestrator registry
	Orchestrators = orchestration.NewRegistry()

	// Register command-specific orchestrators
	downloadOrch := orchestration.NewDownloadOrchestrator("./downloads")
	Orchestrators.Register("upload", downloadOrch)
	Orchestrators.Register("download", downloadOrch)
	Orchestrators.Register("run_cmd", downloadOrch)

	log.Println("|📋 TASK MGR| Task management and orchestration system initialized")
}
