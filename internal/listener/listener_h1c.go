package listener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type http1ClearListener struct {
	id     string
	addr   string
	config ListenerConfig //Store original config for ref
	server *http.Server   // As we saw before we need to access this to shut it down
}

// Now the constructor to implement it

func NewHTTP1ClearListener(id string, config ListenerConfig) (Listener, error) {

	// We could do a check here, but we already did it in the factory, so no real reason

	fullAddr := fmt.Sprintf("%s:%s", config.IP, config.Port)

	// Note we pass id, if we created it locally here, we need to do the exact same thing for every single protocol

	srv := &http.Server{
		Addr:    fullAddr,
		Handler: config.Handler,
	}

	l := &http1ClearListener{
		id:     id,
		addr:   fullAddr,
		config: config,
		server: srv,
	}

	log.Printf("Starting HTTP1Clear listener on %s\n", l.addr)

	return l, nil
}

func (l *http1ClearListener) Start() error {

	err := l.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("|❗ERR|-> Listener %s (%s) failed: %v", l.id, l.config.Type, err)
		return fmt.Errorf("listener %s failed: %w", l.id, err)
	}

	log.Printf("Started HTTP1Clear listener on %s\n", l.addr)
	return nil
}

func (l *http1ClearListener) Stop() error {
	log.Printf("|🛑 STP|-> Shutting down listener %s (%s) on %s...", l.id, l.config.Type, l.addr)

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	err := l.server.Shutdown(ctx)
	if err != nil {
		log.Printf("|❗ERR|-> Error shutting down listener %s (%s): %v", l.id, l.config.Type, err)
		return fmt.Errorf("error shutting down listener %s: %w", l.id, err)
	}
	log.Printf("Shutting down HTTP1Clear listener on %s\n", l.addr)
	return nil
}

func (l *http1ClearListener) Addr() string {
	return l.addr
}

func (l *http1ClearListener) Type() ListenerType {
	return l.config.Type
}

func (l *http1ClearListener) ID() string {
	return l.id
}

// Compile-time check for listener implementation
var _ Listener = (*http1ClearListener)(nil)
