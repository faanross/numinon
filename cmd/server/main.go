package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"numinon_shadow/internal/listener"
	"numinon_shadow/internal/router"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const serverAddr = "127.0.0.1"

var serverPorts = []string{"7777", "8888", "9999"}

func main() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// We move setting up our main router + routes back here from factory.go
	// We only ever create one router, it is recycled for each listener
	r := chi.NewRouter()
	router.SetupRoutes(r)

	// we need to create our new config

	newConfig := listener.NewListenerConfig(listener.TypeHTTP1Clear, serverAddr, "7777", r)

	newListener, err := listener.NewListener(*newConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = newListener.Start()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	fmt.Println("Server started, all listeners now running")

	<-sigChan

	time.Sleep(1 * time.Second)

	fmt.Println("All listeners stopped, now shutting down server...")

	log.Printf("Shutting down server at %s", serverAddr)

}
