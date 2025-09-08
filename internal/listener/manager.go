package listener

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Manager handles the lifecycle of multiple listeners
type Manager struct {
	mu        sync.RWMutex
	listeners map[string]Listener // key: listenerID
	router    http.Handler        // Shared router for all listeners
	stopChan  chan struct{}       // For graceful shutdown
}

// NewManager creates a new listener manager
func NewManager(router http.Handler) *Manager {
	return &Manager{
		listeners: make(map[string]Listener),
		router:    router,
		stopChan:  make(chan struct{}),
	}
}

// CreateListener creates and starts a new listener
func (m *Manager) CreateListener(lType ListenerType, ip string, port string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create configuration
	config := NewListenerConfig(lType, ip, port, m.router)

	// Create the listener
	l, err := NewListener(*config)
	if err != nil {
		return "", fmt.Errorf("failed to create listener: %w", err)
	}

	// Start it in a goroutine
	go func(listener Listener) {
		select {
		case <-m.stopChan:
			return
		default:
			if err := listener.Start(); err != nil {
				log.Printf("|❗ERR MGR| Listener %s failed: %v", listener.ID(), err)
			}
		}
	}(l)

	// Store it
	m.listeners[l.ID()] = l

	log.Printf("|📡 MGR| Created and started listener %s (%s) on %s:%s",
		l.ID(), lType, ip, port)

	return l.ID(), nil
}

// StopListener stops and removes a specific listener
func (m *Manager) StopListener(listenerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, exists := m.listeners[listenerID]
	if !exists {
		return fmt.Errorf("listener %s not found", listenerID)
	}

	// Stop the listener
	if err := l.Stop(); err != nil {
		return fmt.Errorf("failed to stop listener %s: %w", listenerID, err)
	}

	// Remove from map
	delete(m.listeners, listenerID)

	log.Printf("|🛑 MGR| Stopped and removed listener %s", listenerID)
	return nil
}

// GetListener retrieves a listener by ID
func (m *Manager) GetListener(listenerID string) (Listener, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	l, exists := m.listeners[listenerID]
	return l, exists
}

// StopAll stops all managed listeners
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	close(m.stopChan)

	for id, l := range m.listeners {
		if err := l.Stop(); err != nil {
			log.Printf("|❗ERR MGR| Failed to stop listener %s: %v", id, err)
		}
	}

	m.listeners = make(map[string]Listener)
	log.Println("|🛑 MGR| All listeners stopped")
}
