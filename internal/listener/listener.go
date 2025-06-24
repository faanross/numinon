package listener

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
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

var defaultShutdownTimeout = 10 * time.Second

type ListenerConfig struct {
	Type        ListenerType
	IP          string
	Port        string
	Handler     http.Handler
	TLSCertPath string
	TLSKeyPath  string
}

func NewListenerConfig(lType ListenerType, ip string, port string, lHandler http.Handler) *ListenerConfig {

	return &ListenerConfig{
		Type:        lType,
		IP:          ip,
		Port:        port,
		Handler:     lHandler,
		TLSCertPath: "cert.pem",
		TLSKeyPath:  "key.pem",
	}
}

func NewListener(config ListenerConfig) (Listener, error) {
	err := ValidateListener(config)

	if err != nil {
		return nil, err
	}

	id := NewID()

	// --- Switch based on Listener Type ---
	switch config.Type {

	case TypeHTTP1Clear:
		return NewHTTP1ClearListener(id, config)

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

func NewID() string {
	id := fmt.Sprintf("listener_%06d", rand.Intn(1000000))
	return id
}
