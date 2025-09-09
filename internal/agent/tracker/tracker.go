package tracker

// ConnectionState represents the current state of an agent connection
type ConnectionState string

const (
	StateConnected     ConnectionState = "connected"
	StateDisconnected  ConnectionState = "disconnected"
	StateTransitioning ConnectionState = "transitioning" // During HOP
)

// ConnectionType represents how the agent is connected
type ConnectionType string

const (
	TypeHTTP      ConnectionType = "http"
	TypeWebSocket ConnectionType = "websocket"
)
