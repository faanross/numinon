package main

import (
	"encoding/json"
	"github.com/faanross/numinon/internal/agent/agent"
	"github.com/faanross/numinon/internal/agent/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("|AGENT MAIN|-> Starting Numinon Instigator...")

	// STEP ONE IS CREATING OUR CONFIG
	agentConfig := getEmbeddedAgentConfig()

	prettyPrintConfig(&agentConfig)

	// We can now create our agent struct with Agent's constructor
	// REMEMBER -> It will in turn create the specific protocol's communicator based on config

	instigator, err := agent.NewAgent(agentConfig)
	if err != nil {
		// Use log.Fatalf to print error and exit if agent creation fails
		log.Fatalf("|❗ERR AGENT MAIN| Failed to create agent: %v", err)
	}

	log.Println("|AGENT MAIN|-> Instigator instance created.")

	// We'll now start our agent in its own goroutine
	agentDone := make(chan struct{}) // Channel to signal when agent loop finishes
	go func() {
		log.Println("|AGENT MAIN|-> Starting agent loop via goroutine...")
		defer close(agentDone) // Signal completion when this goroutine exits

		err := instigator.Start() // This blocks until the agent loop exits
		if err != nil {
			log.Printf("|❗ERR AGENT MAIN| Agent loop exited with error: %v", err)
		} else {
			log.Println("|AGENT MAIN|-> Agent loop exited gracefully.")
		}
		// If Start() returns nil, it means a graceful shutdown via stopChan occurred.
	}()
	log.Println("|AGENT MAIN|-> Agent started. Waiting for shutdown signal (CTRL+C)...")

	// Setting up signal for graceful shutdown (SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received OR the agent loop exits on its own.
	select {
	case receivedSignal := <-sigChan:
		log.Printf("|AGENT MAIN|-> Received signal: %s. Initiating graceful shutdown...", receivedSignal)
		// Initiate agent shutdown
		instigator.Stop()
		// Remember: Calling Stop() will cause Start() to exit
		// That in turn closes agentDone, meaning below we can pass <-agentDone
	case <-agentDone:
		log.Println("|AGENT MAIN|-> Agent loop exited independently. Proceeding with shutdown.")
	}

	<-agentDone // Wait until the agent goroutine has actually finished

	log.Println("|AGENT MAIN|-> Shutdown complete. Exiting.")
	time.Sleep(500 * time.Millisecond)
}

func prettyPrintConfig(agentConfig *config.AgentConfig) {
	// Marshal the config into a nicely indented JSON string for readability.
	configBytes, err := json.MarshalIndent(agentConfig, "", "  ")
	if err != nil {
		// If for some reason marshalling fails, fall back to the old, less readable format.
		log.Printf("|🔥 AGENT CONFIG|-> (JSON log failed) Using configuration: %+v", agentConfig)
	} else {
		// The leading "\n" adds a newline before the JSON block, making the log output cleaner.
		log.Printf("|🔥 AGENT CONFIG|-> Using configuration:\n%s", string(configBytes))
	}
}
