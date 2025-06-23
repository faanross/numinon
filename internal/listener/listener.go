package listener

import (
	"fmt"
	"net/http"
)

type Listener interface {
	Start() error
	Stop() error
	Addr() string
	ID() string
	Type() ListenerType
}

type ListenerType string

const (
	TypeHTTP1Clear      ListenerType = "H1C"
	TypeHTTP1TLS        ListenerType = "H1TLS"
	TypeHTTP2TLS        ListenerType = "H2TLS"
	TypeHTTP3           ListenerType = "H3"
	TypeWebsocketClear  ListenerType = "WS"
	TypeWebsocketSecure ListenerType = "WSS"
)

type ListenerConfig struct {
	Type        ListenerType
	IP          string
	Port        string
	Handler     http.Handler
	TLSCertPath string
	TLSKeyPath  string
}

func NewListener(config ListenerConfig) (Listener, error) {
	err := ValidateListener(config)

	if err != nil {
		return nil, err
	}

	// --- Switch based on Listener Type ---
	switch config.Type {

	case TypeHTTP1Clear:

		return nil, fmt.Errorf("protocol not yet implemented")
		//return newHttp1ClearListener(listenerID, config)

	case TypeHTTP1TLS:
		return nil, fmt.Errorf("protocol not yet implemented")
		//return newHttp1TLSListener(listenerID, config)

	case TypeHTTP2TLS:
		return nil, fmt.Errorf("protocol not yet implemented")
		//return newHttp2TLSListener(listenerID, config)

	case TypeHTTP3:
		return nil, fmt.Errorf("protocol not yet implemented")
		//return newHttp3Listener(listenerID, config)

	case TypeWebsocketClear:
		return nil, fmt.Errorf("protocol not yet implemented")
		//return newWebsocketClearListener(listenerID, config)

	case TypeWebsocketSecure:
		return nil, fmt.Errorf("protocol not yet implemented")
		//return newWebsocketSecureListener(listenerID, config)

	default:
		return nil, fmt.Errorf("unknown listener type specified: %s", config.Type)
	}

}

func ValidateListener(config ListenerConfig) error {
	if config.Type == "" || config.IP == "" || config.Port == "" {
		return fmt.Errorf("listener config requires a Type, IP, and Port to be set")
	}

	if config.Handler == nil {
		return fmt.Errorf("listener config requires a Handler")
	}

	isTLSRequired := config.Type == TypeHTTP1TLS || config.Type == TypeHTTP2TLS ||
		config.Type == TypeHTTP3 || config.Type == TypeWebsocketSecure

	if isTLSRequired {
		if config.TLSCertPath == "" || config.TLSKeyPath == "" {
			return fmt.Errorf("listener config requires a TLS cert and key path")
		}
	}

	return nil
}
