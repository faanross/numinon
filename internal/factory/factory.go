package factory

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"math/rand"
	"net/http"
	"numinon_shadow/internal/router"
	"time"
)

type ListenerFactory struct{}

func NewListenerFactory() *ListenerFactory {
	return &ListenerFactory{}
}

type Listener struct {
	ID     string
	IP     string
	Port   string
	Router *chi.Mux
	Server *http.Server
}

func (f *ListenerFactory) NewListener(ip string, port string) *Listener {

	r := chi.NewRouter()

	router.SetupRoutes(r)

	return &Listener{
		ID:     id,
		IP:     ip,
		Port:   port,
		Router: r,
	}
}

func (l *Listener) Stop() error {
	if l.Server == nil {
		return fmt.Errorf("listener not started")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := l.Server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown server gracefully")
	}

	fmt.Printf("Successfully shut down listener with ID: %s (port: %s)\n", l.ID, l.Port)
	return nil

}
