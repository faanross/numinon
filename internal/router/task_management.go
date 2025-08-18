package router

import (
	"log"
	"numinon_shadow/internal/taskmanager"
)

// Global task manager instance (this will later be initialized in main.go or router setup)
var TaskManager taskmanager.TaskManager
var ResultProcessors *taskmanager.ResultProcessorRegistry

// InitializeTaskManagement sets up the task management system
func InitializeTaskManagement() {
	// Create task manager
	TaskManager = taskmanager.NewMemoryTaskStore()

	// Create processor registry
	ResultProcessors = taskmanager.NewResultProcessorRegistry()

	// Register command-specific processors (TODO implement these next)
	// ResultProcessors.Register("download", &DownloadProcessor{})
	// ResultProcessors.Register("upload", &UploadProcessor{})

	log.Println("|📋 TASK MGR| Task management system initialized")
}
