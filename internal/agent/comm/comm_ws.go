package comm

import (
	"fmt"
	"github.com/faanross/numinon/internal/agent/config"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"sync"
	"time"
)

// Compile-time check for interface implementation.
var _ Communicator = (*WsClearCommunicator)(nil)

// WsClearCommunicator implements Communicator for cleartext WebSocket (ws://).
type WsClearCommunicator struct {
	agentConfig config.AgentConfig
	conn        *websocket.Conn  // The active websocket connection
	connMtx     sync.Mutex       // Mutex to protect concurrent access to conn
	dialer      websocket.Dialer // Dialer for establishing connections
}

// NewWsClearCommunicator creates a WS communicator.
func NewWsClearCommunicator(cfg config.AgentConfig) (*WsClearCommunicator, error) {

	err := BasicValidateWS(cfg)
	if err != nil {
		return nil, err
	}

	// Configure the dialer (can set timeouts, etc.)
	dialer := websocket.Dialer{
		Proxy:            websocket.DefaultDialer.Proxy, // Use system proxy settings
		HandshakeTimeout: 15 * time.Second,              // Timeout for the WS handshake
	}

	log.Printf("|COMM INIT|-> Initializing WS Clear Communicator for ws://%s:%s%s",
		cfg.ServerIP, cfg.ServerPort, cfg.CheckInEndpoint)

	return &WsClearCommunicator{
		agentConfig: cfg,
		dialer:      dialer,
		// conn is initially nil
	}, nil
}

func BasicValidateWS(cfg config.AgentConfig) error {
	if cfg.Protocol != config.WebsocketClear {
		return fmt.Errorf("mismatched config: NewWsClearCommunicator called with protocol %s", cfg.Protocol)
	}
	if cfg.ServerIP == "" || cfg.ServerPort == "" || cfg.CheckInEndpoint == "" {
		// CheckInEndpoint is used as the path for the WS connection (e.g., /ws)
		return fmt.Errorf("config requires ServerIP, ServerPort, and CheckInEndpoint (path) for WS")
	}
	return nil
}

// Connect establishes the WebSocket connection.
func (c *WsClearCommunicator) Connect() error {
	c.connMtx.Lock()
	// If already connected, let's disconnect first.
	if c.conn != nil {
		log.Println("|COMM WS| Connect called but already connected. Closing existing connection.")
		c.conn.Close()
		c.conn = nil
	}
	c.connMtx.Unlock()

	// (1) CONSTRUCT URL: ws://host:port/path
	targetURL := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%s", c.agentConfig.ServerIP, c.agentConfig.ServerPort),
		Path:   c.agentConfig.WebsocketEndpoint,
	}
	fullURL := targetURL.String()

	log.Printf("|COMM WS|-> Attempting to connect to %s...", fullURL)

	// (2) DIAL THE SERVER
	// First argument = fullURL
	// The second argument is http.Header, can be nil or used to pass custom headers (like AgentID)
	header := map[string][]string{
		"User-Agent": {"PunkinAgent/0.1-WS-Dialer"},
		"Agent-ID":   {c.agentConfig.UUID},
	}

	// (3) USE DIALER TO CREATE CONNECTION
	conn, resp, err := c.dialer.Dial(fullURL, header)
	if err != nil {
		// Log response details if available, helps debug handshake issues
		errMsg := fmt.Sprintf("WS Dial error to %s: %v", fullURL, err)
		if resp != nil {
			errMsg = fmt.Sprintf("%s (Handshake Response: %s)", errMsg, resp.Status)
			// Optionally read and log resp.Body here for more detail
		} else {
			errMsg = fmt.Sprintf("%s (No response received)", errMsg)
		}
		log.Printf("|❗ERR COMM WS| %s", errMsg)
		return fmt.Errorf(errMsg) // Return the specific error
	}
	// Handshake successful!

	log.Printf("|COMM WS|-> WebSocket connection established successfully to %s", fullURL)

	// (4) SAVE OUR CONNECTION OBJECT
	// Goes from nil (we did not instantiate in constructor)
	// To actual object we receive post-successful handshake
	c.connMtx.Lock()
	c.conn = conn
	c.connMtx.Unlock()
	return nil
}

// Disconnect closes the WebSocket connection.
func (c *WsClearCommunicator) Disconnect() error {
	log.Println("|COMM WS|-> Disconnect() called.")
	c.connMtx.Lock()
	defer c.connMtx.Unlock()

	// Guard Clause: Check if connection exists
	if c.conn == nil {
		log.Println("|COMM WS|-> Already disconnected.")
		return nil // Nothing to do
	}

	// Attempt to close the connection
	log.Println("|COMM WS|-> Closing WebSocket connection...")
	err := c.conn.Close()
	c.conn = nil // Set to nil regardless of close error

	// Guard Clause: Check error during close
	if err != nil {
		// Ignore "close sent" type errors which are normal during shutdown race conditions
		if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			log.Printf("|❗ERR COMM WS| Error closing WebSocket: %v", err)
			return err // Return actual error
		}
		log.Printf("|COMM WS|-> WebSocket closed with expected close error: %v", err)
	} else {
		log.Println("|COMM WS|-> WebSocket closed successfully.")
	}

	return nil
}

// CheckIn is a no-op for persistent WebSocket connections.
func (c *WsClearCommunicator) CheckIn() ([]byte, error) {
	return []byte{}, nil
}

// SendResult sends TaskResult data over the WebSocket connection.
func (c *WsClearCommunicator) SendResult(resultData []byte) error {
	log.Printf("|COMM WS|-> Attempting to send result (%d bytes)...", len(resultData))
	c.connMtx.Lock()
	conn := c.conn // Get current connection
	c.connMtx.Unlock()

	// Guard Clause: Check if connected
	if conn == nil {
		log.Println("|❗ERR COMM WS| Cannot send result: Not connected.")
		return fmt.Errorf("cannot send result: not connected")
	}

	// Send the data as a binary message (or text if known to be JSON)
	// Using WriteMessage for simplicity, could use WriteJSON if data is always TaskResult struct
	err := conn.WriteMessage(websocket.TextMessage, resultData) // Assuming resultData is JSON text

	// Guard Clause: Check for write error
	if err != nil {
		log.Printf("|❗ERR COMM WS| Failed to write result message: %v", err)
		// Consider triggering a disconnect/reconnect cycle here?
		// For now, just return the error. The read loop might detect the broken pipe too.
		return fmt.Errorf("ws write message failed: %w", err)
	}

	log.Printf("|COMM WS|-> Successfully sent result message.")
	return nil
}

// ReadTaskMessage reads a single message from the WebSocket connection.
func (c *WsClearCommunicator) ReadTaskMessage() ([]byte, error) {
	c.connMtx.Lock()
	conn := c.conn
	c.connMtx.Unlock()

	// Guard Clause: Check if connected
	if conn == nil {
		log.Println("|COMM WS Read| Cannot read message: Not connected.")
		return nil, fmt.Errorf("not connected") // Or use io.EOF?
	}

	// Read a message (blocks)
	messageType, messageBytes, err := conn.ReadMessage()

	// Guard Clause: Check for read error
	if err != nil {
		log.Printf("|❗ERR COMM WS Read| Error reading message: %v", err)
		// If it's a closure error, signal disconnect more clearly?
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			log.Printf("|COMM WS Read| Unexpected close error detected.")
		} else if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			log.Printf("|COMM WS Read| Normal close error detected.")
		}
		// Trigger disconnect if connection is broken
		_ = c.Disconnect() // Attempt to clean up on error
		return nil, err    // Return the actual error
	}

	// Process based on type? For now, assume text/binary are task data.
	log.Printf("|COMM WS Read| Received message type %d (%d bytes)", messageType, len(messageBytes))
	if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
		return messageBytes, nil
	} else if messageType == websocket.CloseMessage {
		log.Println("|COMM WS Read| Received WebSocket Close frame.")
		_ = c.Disconnect()                                 // Ensure state is cleaned up
		return nil, fmt.Errorf("websocket closed by peer") // Return specific error
	} else {
		log.Printf("|COMM WS Read| Received unhandled message type: %d", messageType)
		return nil, nil
	}
}

// Type returns the protocol type.
func (c *WsClearCommunicator) Type() config.AgentProtocol {
	return config.WebsocketClear
}
