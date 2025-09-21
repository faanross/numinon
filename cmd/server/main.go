package main

import (
	"fmt"
	"github.com/faanross/numinon/internal/clientapi"
	"github.com/faanross/numinon/internal/router"
	"github.com/faanross/numinon/internal/server"
	"github.com/faanross/numinon/internal/tracker"
	"github.com/go-chi/chi/v5"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("|👽 SERVER|-> Starting Numinon Server...")

	// (1) SETUP SIGNALS AND CHANNELS
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	stopChan := make(chan struct{})

	// Create router
	r := chi.NewRouter()

	// Create agent tracker
	agentTracker := tracker.NewTracker()

	// Create full dependencies with operator support
	deps := server.NewDependencies(r, agentTracker)

	// Set global references for the client API
	clientapi.CSM = deps.ClientSessions
	clientapi.TM = deps.TaskBroker

	// Setup routes with full operator support
	router.InitializeWithOperatorSupport(deps)
	router.SetupRoutesWithManagerAndTracker(r, deps.ListenerMgr, agentTracker)

	// Add the client WebSocket handler route
	r.HandleFunc("/client", clientapi.ClientWSHandler)

	// Start the Client API listener (WebSocket on port 8080)
	clientAPIListener, err := clientapi.StartClientAPIListener(r, stopChan)
	if err != nil {
		log.Fatalf("|❗ERR| Failed to start Client API listener: %v", err)
	}

	log.Println("|✅ SERVER| Server started with full operator support")
	log.Println("|📡 SERVER| Client API available at ws://localhost:8080/client")
	log.Println("|📡 SERVER| Ready for agent listeners to be created via client")

	// Wait for shutdown signal
	<-sigChan

	log.Println("|🚦 SIG|-> Shutdown signal received.")

	// Stop Client API listener
	if err := clientAPIListener.Stop(); err != nil {
		log.Printf("|❗ERR| Error stopping Client API listener: %v", err)
	}

	// Stop all agent listeners
	deps.ListenerMgr.StopAll()

	// Signal all goroutines to stop
	close(stopChan)

	fmt.Println("All listeners stopped, shutting down server...")
}
