package main

import (
	"fmt"
	"log"
	"numinon_shadow/internal/factory"
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
	stopChan := make(chan struct{})

	newFactory := factory.NewListenerFactory()

	var listeners []*factory.Listener

	for _, port := range serverPorts {
		l := newFactory.NewListener(serverAddr, port)
		listeners = append(listeners, l)

		go func(l *factory.Listener) {
			log.Printf("Starting listener %s listening on port %s", l.ID, l.Port)
			err := l.Start()
			select {
			case <-stopChan:
				return
			default:
				if err != nil {
					log.Printf("Error starting listener: %s", err)
				}
			}
		}(l)

	}

	time.Sleep(1 * time.Second)
	fmt.Println("Server started, all listeners now running")

	<-sigChan

	StopListeners(listeners, stopChan)

	time.Sleep(1 * time.Second)

	fmt.Println("All listeners stopped, now shutting down server...")

	log.Printf("Shutting down server at %s", serverAddr)

}

func StopListeners(listeners []*factory.Listener, stopChan chan struct{}) {

	close(stopChan)
	
	for _, l := range listeners {
		err := l.Stop()
		if err != nil {
			log.Printf("Error stopping listener: %s", err)
		}

	}
}
