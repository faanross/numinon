package listener

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"log"
	"net"
	"net/http"
)

// http3Listener implements the Listener interface for HTTP/3 over QUIC.
type http3Listener struct {
	commonListener
	server   *http3.Server // Uses the http3 server type from quic-go
	errChan  chan error    // During listener startup this channel indicates failure
	listener quic.EarlyListener
}

func NewHTTP3Listener(id string, config ListenerConfig) (Listener, error) {

	fullAddr := fmt.Sprintf("%s:%s", config.IP, config.Port)

	tlsConf, err := generateTLSConfig(config.TLSCertPath, config.TLSKeyPath)
	if err != nil {
		log.Printf("|❗ERR|-> Listener %s (%s) failed to generate TLS config: %v", id, config.Type, err)
		return nil, fmt.Errorf("h3 listener failed to generate tls config: %w", err)
	}

	// Wrap the handler with middleware that adds listener ID
	wrappedHandler := WithListenerID(id, config.Handler)

	// 2. Create the http3.Server and explicitly provide the TLSConfig.
	srv := &http3.Server{
		Addr:      fullAddr,
		Handler:   wrappedHandler,
		TLSConfig: tlsConf, // Assign the config here
	}

	l := &http3Listener{
		commonListener: commonListener{ // Initialize the embedded struct
			id:     id,
			addr:   config.IP,
			config: config,
		},
		server:  srv,
		errChan: make(chan error, 1),
	}

	log.Printf("Starting %s listener on %s\n", l.config.Type, l.addr)

	return l, nil
}

// START AND STOP
func (l *http3Listener) Start() error {
	// --- Phase 1: Synchronous and Fallible Setup ---

	// 1. Get a UDP socket. This can fail immediately.
	udpConn, err := net.ListenPacket("udp", l.server.Addr)
	if err != nil {
		log.Printf("|❗ERR|-> Listener %s (%s) failed to get UDP socket on %s: %v", l.id, l.config.Type, l.server.Addr, err)
		return fmt.Errorf("listener %s failed to listen on packet sock: %w", l.id, err)
	}

	// 2. Create the QUIC listener from the UDP socket. This can also fail.
	// We pass in the server's TLS config here.
	tlsConf := l.server.TLSConfig
	quicListener, err := quic.ListenEarly(udpConn, tlsConf, nil)
	if err != nil {
		log.Printf("|❗ERR|-> Listener %s (%s) failed to create QUIC listener: %v", l.id, l.config.Type, err)
		// Clean up the UDP connection if QUIC fails to start
		_ = udpConn.Close()
		return fmt.Errorf("listener %s failed to start QUIC listener: %w", l.id, err)
	}

	// If we've reached this point, setup is a 100% success.
	// The OS has given us the port and the QUIC layer is ready.
	udpListener := quicListener
	log.Printf("|✅ OK |-> Listener %s (%s) started successfully on %s.", l.id, l.config.Type, l.config.IP)

	// --- Phase 2: Asynchronous, Long-Running Serving ---
	go func() {
		defer close(l.errChan)

		// The Serve() method takes the listener we know is good and blocks forever.
		// Any error from here is a runtime error, not a startup error.
		err := l.server.ServeListener(udpListener)
		if err != nil {
			// We check for ErrServerClosed to avoid logging an error on a graceful shutdown.
			if !errors.Is(err, http.ErrServerClosed) {
				log.Printf("|❗ERR|-> Listener %s (%s) runtime error: %v", l.id, l.config.Type, err)
				l.errChan <- err
			}
		}
	}()

	// Startup was a synchronous success. Return nil with full confidence.
	return nil
}

// Stop closes the HTTP/3 server.
func (l *http3Listener) Stop() error {
	if l.server == nil {
		return fmt.Errorf("listener %s (%s): server instance is nil", l.id, l.config.Type)
	}
	log.Printf("|🛑 STP|-> Shutting down listener %s (%s) on UDP %s...", l.id, l.config.Type, l.addr)

	// Close the http3 server.
	// This should cause the ListenAndServeQUIC goroutine to unblock and return an error.
	err := l.server.Close()
	if err != nil {
		log.Printf("|❗ERR|-> Error closing listener %s (%s): %v", l.id, l.config.Type, err)
		// Potentially drain errChan here if needed, though closing should signal the goroutine end.
		return fmt.Errorf("error closing listener %s: %w", l.id, err)
	}

	// Wait for the Start goroutine to finish and check its error
	finalErr := <-l.errChan
	// Check if the error is an expected close error (might need specific error type from quic-go/http3 if available)
	// For now, assume nil or a generic close error is fine after calling Close().
	if finalErr != nil {
		// Log errors that occurred during runtime or weren't the expected close error
		log.Printf("|👂LST|-> Listener %s (%s) goroutine finished with error: %v", l.id, l.config.Type, finalErr)
	}

	log.Printf("|✅ STP|-> Listener %s (%s) shut down successfully.", l.id, l.config.Type)
	return nil // Return nil from Stop if Close() succeeded, even if goroutine had runtime error
}

// HELPER FUNCTIONS BELOW

// generateTLSConfig creates a tls.Config for HTTP/3.
func generateTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3"}, // This is critical for HTTP/3
	}, nil
}

// Compile-time check to ensure *http3Listener implements the Listener interface.
var _ Listener = (*http3Listener)(nil)
