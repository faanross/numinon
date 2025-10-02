// internal/websocket/manager.go
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Connection Management
	TypeConnect    MessageType = "connect"
	TypeDisconnect MessageType = "disconnect"
	TypePing       MessageType = "ping"
	TypePong       MessageType = "pong"

	// Auth
	TypeAuthRequest  MessageType = "auth:request"
	TypeAuthResponse MessageType = "auth:response"

	// Agents
	TypeAgentList   MessageType = "agent:list"
	TypeAgentUpdate MessageType = "agent:update"

	// Tasks
	TypeTaskExecute MessageType = "task:execute"
	TypeTaskResult  MessageType = "task:result"
)

// Message represents a WebSocket message
type Message struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload,omitempty"`
}

// Manager handles WebSocket connections to the C2 server
type Manager struct {
	ctx context.Context
	mu  sync.RWMutex

	// Connection
	conn      *websocket.Conn
	url       string
	connected bool

	// Configuration
	config Config

	// Message handling
	handlers     map[MessageType]MessageHandler
	pendingCalls map[string]chan *Message // For request/response pattern
	messageQueue []*Message               // Queue for reconnection

	// Lifecycle
	pingTicker     *time.Ticker
	pongTimer      *time.Timer
	reconnectTimer *time.Timer
	done           chan struct{}
}

// Config holds WebSocket configuration
type Config struct {
	ReconnectDelay       time.Duration
	MaxReconnectAttempts int
	PingInterval         time.Duration
	PongTimeout          time.Duration
	MessageTimeout       time.Duration
}

// MessageHandler processes incoming messages
type MessageHandler func(msg *Message) error

// DefaultConfig returns default WebSocket configuration
func DefaultConfig() Config {
	return Config{
		ReconnectDelay:       5 * time.Second,
		MaxReconnectAttempts: 10,
		PingInterval:         30 * time.Second,
		PongTimeout:          10 * time.Second,
		MessageTimeout:       30 * time.Second,
	}
}

// NewManager creates a new WebSocket manager
func NewManager(ctx context.Context, config Config) *Manager {
	return &Manager{
		ctx:          ctx,
		config:       config,
		handlers:     make(map[MessageType]MessageHandler),
		pendingCalls: make(map[string]chan *Message),
		done:         make(chan struct{}),
	}
}

// Connect establishes WebSocket connection
func (m *Manager) Connect(serverURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return fmt.Errorf("already connected")
	}

	m.url = serverURL

	// Log connection attempt
	runtime.LogInfo(m.ctx, fmt.Sprintf("Connecting to WebSocket: %s", serverURL))
	runtime.EventsEmit(m.ctx, "ws:connecting", serverURL)

	// Establish connection
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(serverURL, nil)
	if err != nil {
		runtime.LogError(m.ctx, fmt.Sprintf("WebSocket connection failed: %v", err))
		runtime.EventsEmit(m.ctx, "ws:error", map[string]string{
			"error": err.Error(),
			"type":  "connection",
		})

		// Attempt reconnection
		go m.scheduleReconnect(1)
		return err
	}

	m.conn = conn
	m.connected = true

	// Start message handlers
	go m.readLoop()
	go m.pingLoop()

	// Emit connected event
	runtime.EventsEmit(m.ctx, "ws:connected", map[string]interface{}{
		"url":       serverURL,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	// Process queued messages
	go m.processQueue()

	runtime.LogInfo(m.ctx, "WebSocket connected successfully")
	return nil
}

// Disconnect closes the WebSocket connection
func (m *Manager) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil
	}

	runtime.LogInfo(m.ctx, "Disconnecting WebSocket")

	// Stop timers
	if m.pingTicker != nil {
		m.pingTicker.Stop()
	}
	if m.pongTimer != nil {
		m.pongTimer.Stop()
	}
	if m.reconnectTimer != nil {
		m.reconnectTimer.Stop()
	}

	// Send close message
	if m.conn != nil {
		m.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		m.conn.Close()
	}

	m.connected = false
	close(m.done)

	// Clear pending calls
	for id, ch := range m.pendingCalls {
		close(ch)
		delete(m.pendingCalls, id)
	}

	runtime.EventsEmit(m.ctx, "ws:disconnected", map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	})

	return nil
}

// Send sends a message and optionally waits for response
func (m *Manager) Send(msgType MessageType, payload interface{}) (*Message, error) {
	msg := &Message{
		ID:        uuid.New().String(),
		Type:      msgType,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	// Queue if not connected
	if !m.IsConnected() {
		m.queueMessage(msg)
		return nil, fmt.Errorf("not connected, message queued")
	}

	// Send message
	if err := m.sendMessage(msg); err != nil {
		m.queueMessage(msg)
		return nil, err
	}

	// For certain message types, wait for response
	if m.expectsResponse(msgType) {
		return m.waitForResponse(msg.ID)
	}

	return msg, nil
}

// sendMessage sends a message over the WebSocket connection
func (m *Manager) sendMessage(msg *Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected || m.conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := m.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		m.connected = false
		go m.scheduleReconnect(1)
		return fmt.Errorf("write error: %w", err)
	}

	runtime.LogDebug(m.ctx, fmt.Sprintf("Sent WS message: %s", msg.Type))
	return nil
}

// readLoop continuously reads messages from the WebSocket
func (m *Manager) readLoop() {
	defer func() {
		m.mu.Lock()
		m.connected = false
		m.mu.Unlock()
		m.scheduleReconnect(1)
	}()

	for {
		select {
		case <-m.done:
			return
		default:
			// Set read deadline
			m.conn.SetReadDeadline(time.Now().Add(m.config.PongTimeout))

			_, data, err := m.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure) {
					runtime.LogError(m.ctx, fmt.Sprintf("WebSocket read error: %v", err))
				}
				return
			}

			// Parse message
			var msg Message
			if err := json.Unmarshal(data, &msg); err != nil {
				runtime.LogError(m.ctx, fmt.Sprintf("Message parse error: %v", err))
				continue
			}

			// Handle message
			go m.handleMessage(&msg)
		}
	}
}

// handleMessage processes incoming messages
func (m *Manager) handleMessage(msg *Message) {
	runtime.LogDebug(m.ctx, fmt.Sprintf("Received WS message: %s", msg.Type))

	// Emit to frontend for real-time updates
	runtime.EventsEmit(m.ctx, string("ws:message:"+msg.Type), msg)

	// Check for pending response
	m.mu.RLock()
	if ch, ok := m.pendingCalls[msg.ID]; ok {
		m.mu.RUnlock()
		ch <- msg
		return
	}
	m.mu.RUnlock()

	// Handle via registered handler
	if handler, ok := m.handlers[msg.Type]; ok {
		if err := handler(msg); err != nil {
			runtime.LogError(m.ctx, fmt.Sprintf("Handler error for %s: %v", msg.Type, err))
		}
	}

	// Special message type handling
	switch msg.Type {
	case TypePong:
		m.handlePong()
	case TypeAgentList, TypeAgentUpdate:
		// Forward to frontend for state updates
		runtime.EventsEmit(m.ctx, "agents:update", msg.Payload)
	case TypeTaskResult:
		// Forward task results
		runtime.EventsEmit(m.ctx, "task:result", msg.Payload)
	}
}

// pingLoop sends periodic pings to keep connection alive
func (m *Manager) pingLoop() {
	m.pingTicker = time.NewTicker(m.config.PingInterval)
	defer m.pingTicker.Stop()

	for {
		select {
		case <-m.pingTicker.C:
			m.sendPing()
		case <-m.done:
			return
		}
	}
}

// sendPing sends a ping message
func (m *Manager) sendPing() {
	if err := m.sendMessage(&Message{
		ID:        uuid.New().String(),
		Type:      TypePing,
		Timestamp: time.Now(),
	}); err != nil {
		runtime.LogError(m.ctx, fmt.Sprintf("Ping failed: %v", err))
		return
	}

	// Start pong timer
	m.mu.Lock()
	if m.pongTimer != nil {
		m.pongTimer.Stop()
	}
	m.pongTimer = time.AfterFunc(m.config.PongTimeout, func() {
		runtime.LogWarning(m.ctx, "Pong timeout - reconnecting")
		m.mu.Lock()
		m.connected = false
		m.mu.Unlock()
		m.scheduleReconnect(1)
	})
	m.mu.Unlock()
}

// handlePong handles pong responses
func (m *Manager) handlePong() {
	m.mu.Lock()
	if m.pongTimer != nil {
		m.pongTimer.Stop()
		m.pongTimer = nil
	}
	m.mu.Unlock()

	// Calculate and emit latency
	// (In production, you'd track ping send time)
	runtime.EventsEmit(m.ctx, "ws:latency", map[string]interface{}{
		"latency": time.Since(time.Now()).Milliseconds(), // Simplified
	})
}

// scheduleReconnect attempts to reconnect after delay
func (m *Manager) scheduleReconnect(attempt int) {
	if attempt > m.config.MaxReconnectAttempts {
		runtime.LogError(m.ctx, "Max reconnection attempts reached")
		runtime.EventsEmit(m.ctx, "ws:reconnect:failed", map[string]interface{}{
			"attempts": attempt,
		})
		return
	}

	delay := time.Duration(attempt) * m.config.ReconnectDelay
	runtime.LogInfo(m.ctx, fmt.Sprintf("Reconnecting in %v (attempt %d/%d)",
		delay, attempt, m.config.MaxReconnectAttempts))

	runtime.EventsEmit(m.ctx, "ws:reconnecting", map[string]interface{}{
		"attempt": attempt,
		"delay":   delay.Seconds(),
	})

	m.reconnectTimer = time.AfterFunc(delay, func() {
		if err := m.Connect(m.url); err != nil {
			m.scheduleReconnect(attempt + 1)
		}
	})
}

// queueMessage adds a message to the queue for later sending
func (m *Manager) queueMessage(msg *Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messageQueue = append(m.messageQueue, msg)

	// Limit queue size
	const maxQueueSize = 100
	if len(m.messageQueue) > maxQueueSize {
		m.messageQueue = m.messageQueue[1:] // Remove oldest
	}
}

// processQueue sends queued messages after reconnection
func (m *Manager) processQueue() {
	m.mu.Lock()
	queue := m.messageQueue
	m.messageQueue = nil
	m.mu.Unlock()

	if len(queue) == 0 {
		return
	}

	runtime.LogInfo(m.ctx, fmt.Sprintf("Processing %d queued messages", len(queue)))

	for _, msg := range queue {
		if err := m.sendMessage(msg); err != nil {
			runtime.LogError(m.ctx, fmt.Sprintf("Failed to send queued message: %v", err))
			// Re-queue on failure
			m.queueMessage(msg)
			break
		}
		time.Sleep(100 * time.Millisecond) // Rate limiting
	}
}

// waitForResponse waits for a response to a specific message ID
func (m *Manager) waitForResponse(msgID string) (*Message, error) {
	ch := make(chan *Message, 1)

	m.mu.Lock()
	m.pendingCalls[msgID] = ch
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.pendingCalls, msgID)
		m.mu.Unlock()
	}()

	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(m.config.MessageTimeout):
		return nil, fmt.Errorf("response timeout")
	}
}

// expectsResponse returns true if the message type expects a response
func (m *Manager) expectsResponse(msgType MessageType) bool {
	switch msgType {
	case TypeAuthRequest, TypeTaskExecute:
		return true
	default:
		return false
	}
}

// RegisterHandler registers a message handler for a specific type
func (m *Manager) RegisterHandler(msgType MessageType, handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[msgType] = handler
}

// IsConnected returns the connection status
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// GetStatus returns detailed connection status
func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"connected":    m.connected,
		"url":          m.url,
		"queueSize":    len(m.messageQueue),
		"pendingCalls": len(m.pendingCalls),
	}
}
