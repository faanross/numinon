package comm

import "numinon_shadow/internal/agent/config"

type Communicator interface {
	Connect() error
	Disconnect() error
	CheckIn() ([]byte, error)
	SendResult(resultData []byte) error
	Type() config.AgentProtocol
	IsConnected() bool
}
