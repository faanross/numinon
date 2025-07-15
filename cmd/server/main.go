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

var serverPorts = []string{"8888"}

func main() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	stopChan := make(chan struct{})

	// We move setting up our main router + routes back here from factory.go
	// We only ever create one router, it is recycled for each listener
	r := chi.NewRouter()
	router.SetupRoutes(r)

	// we need to create our new config

	var activeListeners []listener.Listener

	for _, serverPort := range serverPorts {
		newConfig := listener.NewListenerConfig(listener.TypeHTTP1TLS, serverAddr, serverPort, r)

		l, err := listener.NewListener(*newConfig)
		if err != nil {
			log.Fatal(err)
		}

		activeListeners = append(activeListeners, l)

		go func(l listener.Listener) {
			select {
			case <-stopChan:
				return
			default:
				err = l.Start()
				if err != nil {
					log.Fatal(err)
				}
			}
		}(l)

	}

	time.Sleep(1 * time.Second)
	fmt.Println("Server started, all listeners now running")

	<-sigChan

	log.Println("|🚦 SIG|-> Shutdown signal received.")

	StopAllListener(activeListeners, stopChan)

	time.Sleep(1 * time.Second)

	fmt.Println("All listeners stopped, now shutting down server...")

	log.Printf("Shutting down server at %s", serverAddr)

}

func StopAllListener(listeners []listener.Listener, stopChan chan struct{}) {
	log.Printf("|🛑 STP|-> Initiating shutdown for %d listener(s)...", len(listeners))

	// Signal the listener goroutines to exit by closing the stopChan.
	close(stopChan)

	for _, l := range listeners {
		err := l.Stop()
		log.Printf("|🛑 STP|-> Stopping listener [Type: %s, Addr: %s]", l.Type(), l.Addr())
		if err != nil {
			log.Fatal(err)
		}

	}
	log.Println("|🛑 STP|-> All active listeners have been instructed to stop.")
}
