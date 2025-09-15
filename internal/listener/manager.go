package listener

import (
	"fmt"
	"github.com/faanross/numinon/internal/tracker"
	"log"
	"net/http"
	"sync"
	"time"
)

// ListenerInfo displays important info for operators
type ListenerInfo struct {
	ID         string
	Type       ListenerType
	Address    string
	Agents     []string
	AgentCount int
}

// Manager handles the lifecycle of multiple listeners
type Manager struct {
	mu           sync.RWMutex
	listeners    map[string]Listener // key: listenerID
	router       http.Handler        // Shared router for all listeners
	stopChan     chan struct{}       // For graceful shutdown
	agentTracker *tracker.Tracker    // tracks + manages agent <-> listener life cycles
}

// NewManager creates a new listener manager
func NewManager(router http.Handler, agentTracker *tracker.Tracker) *Manager {
	return &Manager{
		listeners:    make(map[string]Listener),
		router:       router,
		stopChan:     make(chan struct{}),
		agentTracker: agentTracker,
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
func (m *Manager) StopListener(listenerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, exists := m.listeners[listenerID]
	if !exists {
		return fmt.Errorf("listener %s not found", listenerID)
	}

	// Check if safe to stop
	if m.agentTracker != nil {
		canStop, reason := m.agentTracker.CanStopListener(listenerID)
		if !canStop {
			return fmt.Errorf("cannot stop listener %s: %s", listenerID, reason)
		}
		log.Printf("|✅ MGR| Safe to stop listener %s: %s", listenerID, reason)
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

// TryStopListener attempts to stop a listener, with retries if agents are still connected
func (m *Manager) TryStopListener(listenerID string, maxRetries int, retryDelay time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		err := m.StopListener(listenerID)
		if err == nil {
			return nil // Success
		}

		if i < maxRetries-1 {
			log.Printf("|⏳ MGR| Cannot stop listener yet (%s), retry %d/%d in %v",
				err, i+1, maxRetries, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed to stop listener %s after %d retries", listenerID, maxRetries)
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

// GetAllListeners returns information about all active listeners
func (m *Manager) GetAllListeners() []ListenerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var listeners []ListenerInfo
	for id, l := range m.listeners {
		// Get connected agents for this listener
		agents := []string{}
		if m.agentTracker != nil {
			agents = m.agentTracker.GetListenerAgents(id)
		}

		info := ListenerInfo{
			ID:         id,
			Type:       l.Type(),
			Address:    l.Addr(),
			Agents:     agents,
			AgentCount: len(agents),
		}
		listeners = append(listeners, info)
	}

	return listeners
}
