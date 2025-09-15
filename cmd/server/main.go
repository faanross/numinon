package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"numinon_shadow/internal/listener"
	"numinon_shadow/internal/router"
	"numinon_shadow/internal/tracker"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	log.Println("|👽 SERVER|-> Starting Numinon Server...")

	// (1) SETUP SIGNALS AND CHANNELS
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create router
	r := chi.NewRouter()

	// Create agent tracker
	agentTracker := tracker.NewTracker()

	// Create listener manager with tracker
	listenerMgr := listener.NewManager(r, agentTracker)

	// Setup routes with the manager
	router.SetupRoutesWithManagerAndTracker(r, listenerMgr, agentTracker)

	fmt.Println("Server started, all listeners now running")

	<-sigChan

	log.Println("|🚦 SIG|-> Shutdown signal received.")
	listenerMgr.StopAll()

	fmt.Println("All listeners stopped, shutting down server...")
}
