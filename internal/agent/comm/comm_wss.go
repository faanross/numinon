package comm

import (
	"crypto/tls"
	"fmt"
	"github.com/faanross/numinon/internal/agent/config"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"sync"
	"time"
)

// Compile-time check for interface implementation.
var _ Communicator = (*WsSecureCommunicator)(nil)

// WsSecureCommunicator implements Communicator for TLS WebSocket (wss://).
type WsSecureCommunicator struct {
	agentConfig config.AgentConfig
	conn        *websocket.Conn  // The active websocket connection
	connMtx     sync.Mutex       // Mutex to protect concurrent access to conn
	dialer      websocket.Dialer // Dialer for establishing connections
}

// NewWsSecureCommunicator creates a WSS communicator.
func NewWsSecureCommunicator(cfg config.AgentConfig) (*WsSecureCommunicator, error) {

	err := BasicValidateWSS(cfg)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerifyTLS,
	}

	// Configure the dialer (can set timeouts, etc.)
	dialer := websocket.Dialer{
		Proxy:            websocket.DefaultDialer.Proxy, // Use system proxy settings
		HandshakeTimeout: 15 * time.Second,              // Timeout for the WS handshake
		TLSClientConfig:  tlsConfig,
	}

	log.Printf("|COMM INIT|-> Initializing WSS Secure Communicator for wss://%s:%s%s", // <-- CHANGE scheme in log
		cfg.ServerIP, cfg.ServerPort, cfg.WebsocketEndpoint)

	return &WsSecureCommunicator{ // <-- RENAME struct instance
		agentConfig: cfg,
		dialer:      dialer,
	}, nil
}

func BasicValidateWSS(cfg config.AgentConfig) error {
	if cfg.Protocol != config.WebsocketSecure {
		return fmt.Errorf("mismatched config: NewWsSecureCommunicator called with protocol %s", cfg.Protocol)
	}
	if cfg.ServerIP == "" || cfg.ServerPort == "" || cfg.CheckInEndpoint == "" {
		// CheckInEndpoint is used as the path for the WSS connection (e.g., /ws)
		return fmt.Errorf("config requires ServerIP, ServerPort, and CheckInEndpoint (path) for WSS")
	}
	return nil
}

// Connect establishes the WebSocket connection.
func (c *WsSecureCommunicator) Connect() error {
	c.connMtx.Lock()
	// If already connected, let's disconnect first.
	if c.conn != nil {
		log.Println("|COMM WSS| Connect called but already connected. Closing existing connection.")
		c.conn.Close()
		c.conn = nil
	}
	c.connMtx.Unlock()

	// (1) CONSTRUCT URL: wss://host:port/path
	targetURL := url.URL{
		Scheme: "wss",
		Host:   fmt.Sprintf("%s:%s", c.agentConfig.ServerIP, c.agentConfig.ServerPort),
		Path:   c.agentConfig.WebsocketEndpoint,
	}
	fullURL := targetURL.String()

	log.Printf("|COMM WSS|-> Attempting to connect to %s...", fullURL)

	// (2) DIAL THE SERVER
	// First argument = fullURL
	// The second argument is http.Header, can be nil or used to pass custom headers (like AgentID)
	header := map[string][]string{
		"User-Agent": {"PunkinAgent/0.1-WSS-Dialer"},
		"Agent-ID":   {c.agentConfig.UUID},
	}

	// (3) USE DIALER TO CREATE CONNECTION
	conn, resp, err := c.dialer.Dial(fullURL, header)
	if err != nil {
		// Log response details if available, helps debug handshake issues
		errMsg := fmt.Sprintf("WSS Dial error to %s: %v", fullURL, err)
		if resp != nil {
			errMsg = fmt.Sprintf("%s (Handshake Response: %s)", errMsg, resp.Status)
			// Optionally read and log resp.Body here for more detail
		} else {
			errMsg = fmt.Sprintf("%s (No response received)", errMsg)
		}
		log.Printf("|❗ERR COMM WSS| %s", errMsg)
		return fmt.Errorf(errMsg) // Return the specific error
	}
	// Handshake successful!

	log.Printf("|COMM WSS|-> WebSocket TLS connection established successfully to %s", fullURL)

	// (4) SAVE OUR CONNECTION OBJECT
	// Goes from nil (we did not instantiate in constructor)
	// To actual object we receive post-successful handshake
	c.connMtx.Lock()
	c.conn = conn
	c.connMtx.Unlock()
	return nil
}

// Disconnect closes the WebSocket connection.
func (c *WsSecureCommunicator) Disconnect() error {
	log.Println("|COMM WSS|-> Disconnect() called.")
	c.connMtx.Lock()
	defer c.connMtx.Unlock()

	// Guard Clause: Check if connection exists
	if c.conn == nil {
		log.Println("|COMM WSS|-> Already disconnected.")
		return nil // Nothing to do
	}

	// Attempt to close the connection
	log.Println("|COMM WSS|-> Closing WebSocket connection...")
	err := c.conn.Close()
	c.conn = nil // Set to nil regardless of close e rror

	// Guard Clause: Check error during close
	if err != nil {
		// Ignore "close sent" type errors which are normal during shutdown race conditions
		if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			log.Printf("|❗ERR COMM WSS| Error closing WebSocket: %v", err)
			return err // Return actual error
		}
		log.Printf("|COMM WSS|-> WebSocket Secure closed with expected close error: %v", err)
	} else {
		log.Println("|COMM WSS|-> WebSocket Secure closed successfully.")
	}

	return nil
}

// CheckIn is a no-op for persistent WebSocket connections.
func (c *WsSecureCommunicator) CheckIn() ([]byte, error) {
	return []byte{}, nil
}

// SendResult sends TaskResult data over the WebSocket connection.
func (c *WsSecureCommunicator) SendResult(resultData []byte) error {
	log.Printf("|COMM WSS|-> Attempting to send result (%d bytes)...", len(resultData))
	c.connMtx.Lock()
	conn := c.conn // Get current connection
	c.connMtx.Unlock()

	// Guard Clause: Check if connected
	if conn == nil {
		log.Println("|❗ERR COMM WSS| Cannot send result: Not connected.")
		return fmt.Errorf("cannot send result: not connected")
	}

	// Send the data as a binary message (or text if known to be JSON)
	// Using WriteMessage for simplicity, could use WriteJSON if data is always TaskResult struct
	err := conn.WriteMessage(websocket.TextMessage, resultData) // Assuming resultData is JSON text

	// Guard Clause: Check for write error
	if err != nil {
		log.Printf("|❗ERR COMM WSS| Failed to write result message: %v", err)
		// Consider triggering a disconnect/reconnect cycle here?
		// For now, just return the error. The read loop might detect the broken pipe too.
		return fmt.Errorf("wss write message failed: %w", err)
	}

	log.Printf("|COMM WSS|-> Successfully sent result message.")
	return nil
}

// ReadTaskMessage reads a single message from the WebSocket connection.
func (c *WsSecureCommunicator) ReadTaskMessage() ([]byte, error) {
	c.connMtx.Lock()
	conn := c.conn
	c.connMtx.Unlock()

	// Guard Clause: Check if connected
	if conn == nil {
		log.Println("|COMM WSS Read| Cannot read message: Not connected.")
		return nil, fmt.Errorf("not connected") // Or use io.EOF?
	}

	// Read a message (blocks)
	messageType, messageBytes, err := conn.ReadMessage()

	// Guard Clause: Check for read error
	if err != nil {
		log.Printf("|❗ERR COMM WSS Read| Error reading message: %v", err)
		// If it's a closure error, signal disconnect more clearly?
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			log.Printf("|COMM WSS Read| Unexpected close error detected.")
		} else if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			log.Printf("|COMM WSS Read| Normal close error detected.")
		}
		// Trigger disconnect if connection is broken
		_ = c.Disconnect() // Attempt to clean up on error
		return nil, err    // Return the actual error
	}

	// Process based on type? For now, assume text/binary are task data.
	log.Printf("|COMM WSS Read| Received message type %d (%d bytes)", messageType, len(messageBytes))
	if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
		return messageBytes, nil
	} else if messageType == websocket.CloseMessage {
		log.Println("|COMM WSS Read| Received WebSocket Close frame.")
		_ = c.Disconnect()                                 // Ensure state is cleaned up
		return nil, fmt.Errorf("websocket closed by peer") // Return specific error
	} else {
		log.Printf("|COMM WSS Read| Received unhandled message type: %d", messageType)
		return nil, nil
	}
}

// Type returns the protocol type.
func (c *WsSecureCommunicator) Type() config.AgentProtocol {
	return config.WebsocketSecure
}
