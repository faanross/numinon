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

const serverAddr = "0.0.0.0"

var serverPorts = []string{"8888"}

func main() {
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

	// Create initial listeners
	for _, port := range serverPorts {
		_, err := listenerMgr.CreateListener(
			listener.TypeWebsocketClear,
			serverAddr,
			port,
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Server started, all listeners now running")

	<-sigChan

	log.Println("|🚦 SIG|-> Shutdown signal received.")
	listenerMgr.StopAll()

	fmt.Println("All listeners stopped, shutting down server...")
}
