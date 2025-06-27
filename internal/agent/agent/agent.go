package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"numinon_shadow/internal/agent/comm"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"time"
)

// Agent represents an agent instance
type Agent struct {
	config       config.AgentConfig
	communicator comm.Communicator
	stopChan     chan struct{}
	rng          *rand.Rand
}

// NewAgent creates and initializes a new Agent instance.
func NewAgent(cfg config.AgentConfig) (*Agent, error) {
	log.Println("|AGENT INIT|-> Creating new agent instance...")

	communicator, err := comm.NewHttp1ClearCommunicator(cfg)
	if err != nil {
		return nil, err
	}

	log.Printf("|AGENT INIT|-> Agent configured for protocol: %s", cfg.Protocol)

	agent := &Agent{
		config:       cfg,
		communicator: communicator,
		stopChan:     make(chan struct{}),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	log.Println("|AGENT INIT|-> Agent instance created successfully.")
	return agent, nil
}

// calculateSleepWithJitter calculates the next sleep duration based on BaseSleep and Jitter.
func (a *Agent) calculateSleepWithJitter() time.Duration {
	delay := float64(a.config.Delay)
	jitter := a.config.Jitter

	// Calculate the jitter amount
	jitterAmount := delay * jitter

	// Generate a random factor between -1.0 and 1.0
	randomFactor := (a.rng.Float64() * 2) - 1

	// Apply the jitter to the delay
	finalSleep := delay + (jitterAmount * randomFactor)

	if finalSleep < 0 {
		return time.Duration(math.Abs(finalSleep))
	}

	return time.Duration(finalSleep)
}

// executeTask processes a received task, performs the action, and sends the result.
func (a *Agent) executeTask(task models.ServerTaskResponse) {
	log.Printf("|AGENT TASK|-> Executing Task ID: %s, Command: %s", task.TaskID, task.Command)

	var result models.AgentTaskResult
	result.TaskID = task.TaskID

	// --- Placeholder Command Execution ---
	switch task.Command {
	case "ping":
		log.Println("|AGENT TASK|-> Handling 'ping' command.")
		result.Status = "success"
		result.Output = []byte("pong")
	case "echo":
		log.Printf("|AGENT TASK|-> Handling 'echo' command with args: %v", task.Data)
		result.Status = "success"
		result.Output = []byte(fmt.Sprintf("Echoing args: %v", task.Data))
	default:
		log.Printf("|WARN AGENT TASK| Received unknown command: %s", task.Command)
		result.Status = "error"
		result.Error = fmt.Sprintf("Unknown command received: %s", task.Command)
	}
	// --- End Placeholder Logic ---

	// Now marshall the result before sending it back using SendResult
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to marshal result for Task ID %s: %v", task.TaskID, err)
		return // Cannot send result if marshalling fails
	}

	log.Printf("|AGENT TASK|-> Sending result for Task ID %s (%d bytes)...", task.TaskID, len(resultBytes))
	err = a.communicator.SendResult(resultBytes)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to send result for Task ID %s: %v", task.TaskID, err)
	} else {
		log.Printf("|AGENT TASK|-> Successfully sent result for Task ID %s.", task.TaskID)
	}
}
