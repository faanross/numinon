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
	Orchestrators.Register("download", downloadOrch)

	log.Println("|📋 TASK MGR| Task management and orchestration system initialized")
	log.Println("|📋 ORCH| Registered orchestrator for 'download' command")
}
