package clientapi

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"numinon_shadow/internal/listener"
)

// TODO give server it's own YAML config system, move these values there, pass cfg as argument
const (
	// Dedicated port for the client API WebSocket listener
	clientAPIListenerPort = "8080"
	clientAPIListenerIP   = "0.0.0.0"
)

// StartClientAPIListener creates and starts a dedicated WebSocket listener for the client API.
func StartClientAPIListener(router *chi.Mux, stopChan chan struct{}) (listener.Listener, error) {
	log.Printf("|🚀 CLIENT_API_LST|-> Initializing dedicated WebSocket listener for Client API on %s:%s...", clientAPIListenerIP, clientAPIListenerPort)

	// We use TypeWebsocketClear FOR NOW
	// TODO: Change to WSS
	cfg := listener.ListenerConfig{
		Type:    listener.TypeWebsocketClear,
		IP:      clientAPIListenerIP,
		Port:    clientAPIListenerPort,
		Handler: router,
		// TLSCertPath: "certs/cert.pem", // For WSS eventually TODO
		// TLSKeyPath:  "certs/key.pem",   // For WSS eventually TODO
	}

	// Create the listener instance using the shared NewListener logic.
	l, err := listener.NewListener(cfg)
	if err != nil {
		log.Printf("|❗ERR CLIENT_API_LST|-> Error creating dedicated Client API listener for %s:%s (%s): %v", cfg.IP, cfg.Port, cfg.Type, err)
		return nil, fmt.Errorf("creating dedicated client api listener: %w", err)
	}

	// Start the listener in a goroutine
	go func(lst listener.Listener) {
		log.Printf("|👂 CLIENT_API_LST|-> Starting dedicated Client API listener %s (%s) on %s:%s", lst.ID(), lst.Type(), cfg.IP, cfg.Port)
		err := lst.Start() // This blocks until the listener is stopped or an error occurs.
		select {
		case <-stopChan:
			// Stop signal received, expected exit.
			log.Printf("|👂 CLIENT_API_LST|-> Goroutine for Client API listener %s received stop signal. Listener should be stopping.", lst.ID())
			return
		default:
			// If Start() returns an error and we weren't asked to stop, it's unexpected.
			if err != nil && !errors.Is(err, http.ErrServerClosed) { // http.ErrServerClosed is expected on graceful shutdown
				log.Printf("|❗ERR CLIENT_API_LST|-> Client API listener %s (%s) stopped unexpectedly: %v", lst.ID(), lst.Type(), err)
			} else if errors.Is(err, http.ErrServerClosed) {
				log.Printf("|👂 CLIENT_API_LST|-> Client API listener %s (%s) closed gracefully (ErrServerClosed).", lst.ID(), lst.Type())
			} else {
				log.Printf("|👂 CLIENT_API_LST|-> Client API listener %s (%s) stopped without error (but not via stopChan).", lst.ID(), lst.Type())
			}
		}
	}(l)

	log.Printf("|🚀 CLIENT_API_LST|-> Dedicated Client API listener %s (%s) startup process initiated.", l.ID(), l.Type())
	return l, nil
}
