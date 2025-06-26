package listener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type http1TLSListener struct {
	commonListener
	server *http.Server // As we saw before we need to access this to shut it down
}

func NewHTTP1TLSListener(id string, config ListenerConfig) (Listener, error) {

	fullAddr := fmt.Sprintf("%s:%s", config.IP, config.Port)

	srv := &http.Server{
		Addr:    fullAddr,
		Handler: config.Handler,
	}

	l := &http1TLSListener{
		commonListener: commonListener{ // Initialize the embedded struct
			id:     id,
			addr:   config.IP,
			config: config,
		},
		server: srv,
	}

	log.Printf("Starting %s listener on %s\n", l.config.Type, l.addr)

	return l, nil
}

func (l *http1TLSListener) Start() error {

	err := l.server.ListenAndServeTLS(l.config.TLSCertPath, l.config.TLSKeyPath)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("|❗ERR|-> Listener %s (%s) failed: %v", l.id, l.config.Type, err)
		return fmt.Errorf("listener %s failed: %w", l.id, err)
	}

	log.Printf("Started %s listener on %s\n", l.config.Type, l.addr)
	return nil
}

func (l *http1TLSListener) Stop() error {
	log.Printf("|🛑 STP|-> Shutting down listener %s (%s) on %s...", l.id, l.config.Type, l.addr)

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	err := l.server.Shutdown(ctx)
	if err != nil {
		log.Printf("|❗ERR|-> Error shutting down listener %s (%s): %v", l.id, l.config.Type, err)
		return fmt.Errorf("error shutting down listener %s: %w", l.id, err)
	}
	log.Printf("Shutting down %s listener on %s\n", l.config.Type, l.addr)
	return nil
}

// Compile-time check for listener implementation
var _ Listener = (*http1TLSListener)(nil)
