package config

import "time"

type AgentConfig struct {
	Protocol          AgentProtocol
	ServerIP          string
	ServerPort        string
	CheckInEndpoint   string
	ResultsEndpoint   string
	WebsocketEndpoint string
	Delay             time.Duration
	Jitter            float64
	UUID              string
}

func NewAgentConfig(protocol AgentProtocol, ip string, port string, endpoint string,
	results string, wsendpoint string, delay time.Duration, jitter float64, uuid string) *AgentConfig {
	return &AgentConfig{
		Protocol:          protocol,
		ServerIP:          ip,
		ServerPort:        port,
		CheckInEndpoint:   endpoint,
		ResultsEndpoint:   results,
		WebsocketEndpoint: wsendpoint,
		Delay:             delay,
		Jitter:            jitter,
		UUID:              uuid,
	}
}

type AgentProtocol string

const (
	HTTP1Clear      AgentProtocol = "H1C"
	HTTP1TLS        AgentProtocol = "H1TLS"
	HTTP2TLS        AgentProtocol = "H2TLS"
	HTTP3           AgentProtocol = "H3"
	WebsocketClear  AgentProtocol = "WS"
	WebsocketSecure AgentProtocol = "WSS"
)
