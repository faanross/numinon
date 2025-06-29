package router

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"numinon_shadow/internal/models"
	"time"
)

// CheckinHandler processes requests from clients checking in for tasks.
func CheckinHandler(w http.ResponseWriter, r *http.Request) {

	agentID := r.Header.Get("Agent-ID") // Read the custom header
	log.Printf("|✅ CHK_IN| Received check-in GET from Agent ID: %s via %s", agentID, r.RemoteAddr)

	var response models.ServerTaskResponse

	// Randomly decide if a task is available (50/50 chance).
	if seededRand.Intn(2) == 0 {
		// No task is available.
		response.TaskAvailable = false
	} else {
		// A task is available, so populate the details.
		response.TaskAvailable = true
		response.TaskID = generateTaskID()

		// Randomly select a command.
		commands := []string{"ping", "echo", "doesnotexist"}
		response.Command = commands[seededRand.Intn(len(commands))]

		// The 'Data' field is intentionally left empty as requested.
		response.Data = nil
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
	// This just needs to receive the result, for now, all we need to do is just print a test message
	agentID := r.Header.Get("Agent-ID") // Read the custom header
	log.Printf("|✅ CHK_IN| Received results POST from Agent ID: %s via %s", agentID, r.RemoteAddr)

	w.Write([]byte("The RESULTS endpoint was hit: " + r.URL.Path))
}

// seededRand is a random number generator seeded at application start.
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// generateTaskID creates a random task identifier.
func generateTaskID() string {
	return fmt.Sprintf("task_%06d", seededRand.Intn(1000000))
}
