package listener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type WSListener struct {
	id     string
	addr   string
	config ListenerConfig
	server *http.Server
}

func NewWSListener(id string, config ListenerConfig) (Listener, error) {

	fullAddr := fmt.Sprintf("%s:%s", config.IP, config.Port)

	srv := &http.Server{
		Addr:    fullAddr,
		Handler: config.Handler,
	}

	l := &WSListener{
		id:     id,
		addr:   fullAddr,
		config: config,
		server: srv,
	}

	log.Printf("Starting WS listener on %s\n", l.addr)

	return l, nil
}

func (l *WSListener) Start() error {

	err := l.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("|❗ERR|-> Listener %s (%s) failed: %v", l.id, l.config.Type, err)
		return fmt.Errorf("listener %s failed: %w", l.id, err)
	}

	log.Printf("Started WS listener on %s\n", l.addr)
	return nil
}

func (l *WSListener) Stop() error {
	log.Printf("|🛑 STP|-> Shutting down listener %s (%s) on %s...", l.id, l.config.Type, l.addr)

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	err := l.server.Shutdown(ctx)
	if err != nil {
		log.Printf("|❗ERR|-> Error shutting down listener %s (%s): %v", l.id, l.config.Type, err)
		return fmt.Errorf("error shutting down listener %s: %w", l.id, err)
	}
	log.Printf("Shutting down WS listener on %s\n", l.addr)
	return nil
}

func (l *WSListener) Addr() string {
	return l.addr
}

func (l *WSListener) Type() ListenerType {
	return l.config.Type
}

func (l *WSListener) ID() string {
	return l.id
}

// Compile-time check for listener implementation
var _ Listener = (*WSListener)(nil)
