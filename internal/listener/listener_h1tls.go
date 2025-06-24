package listener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type http1TLSListener struct {
	id     string
	addr   string
	config ListenerConfig //Store original config for ref
	server *http.Server   // As we saw before we need to access this to shut it down
}

func NewHTTP1TLSListener(id string, config ListenerConfig) (Listener, error) {

	fullAddr := fmt.Sprintf("%s:%s", config.IP, config.Port)

	srv := &http.Server{
		Addr:    fullAddr,
		Handler: config.Handler,
	}

	l := &http1TLSListener{
		id:     id,
		addr:   fullAddr,
		config: config, // remember cert and key are inside of config
		server: srv,
	}

	log.Printf("Starting HTTP1TLS listener on %s\n", l.addr)

	return l, nil
}

func (l *http1TLSListener) Start() error {

	err := l.server.ListenAndServeTLS(l.config.TLSCertPath, l.config.TLSKeyPath)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("|❗ERR|-> Listener %s (%s) failed: %v", l.id, l.config.Type, err)
		return fmt.Errorf("listener %s failed: %w", l.id, err)
	}

	log.Printf("Started HTTP1TLS listener on %s\n", l.addr)
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
	log.Printf("Shutting down HTTP1TLS listener on %s\n", l.addr)
	return nil
}

func (l *http1TLSListener) Addr() string {
	return l.addr
}

func (l *http1TLSListener) Type() ListenerType {
	return l.config.Type
}

func (l *http1TLSListener) ID() string {
	return l.id
}

// Compile-time check for listener implementation
var _ Listener = (*http1TLSListener)(nil)
