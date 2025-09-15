package comm

import (
	"fmt"
	"github.com/faanross/numinon/internal/agent/config"
)

type Communicator interface {
	Connect() error
	Disconnect() error
	CheckIn() ([]byte, error)
	SendResult(resultData []byte) error
	Type() config.AgentProtocol
}

// WsCommunicator defines the behavior for WebSocket-based communicators,
// which have the unique ability to proactively read task messages from a
// persistent connection. It embeds the standard Communicator interface.
type WsCommunicator interface {
	Communicator // Embeds Connect, Disconnect, CheckIn, SendResult, Type
	ReadTaskMessage() ([]byte, error)
}

func NewCommunicator(cfg config.AgentConfig) (Communicator, error) {

	switch cfg.Protocol {

	case config.HTTP1Clear:
		return NewHttp1ClearCommunicator(cfg)

	case config.HTTP1TLS:
		return NewHttp1TLSCommunicator(cfg)

	case config.HTTP2TLS:
		return NewHttp2TLSCommunicator(cfg)

	case config.HTTP3:
		return NewHttp3Communicator(cfg)

	case config.WebsocketClear:
		return NewWsClearCommunicator(cfg)

	case config.WebsocketSecure:
		return NewWsSecureCommunicator(cfg)

	default:
		return nil, fmt.Errorf("unknown communicator type specified: %s", cfg.Protocol)
	}
}
