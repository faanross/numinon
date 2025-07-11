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

	// New fields related to payload jitter
	CheckinMethod   string
	EnablePadding   bool
	MinPaddingBytes int
	MaxPaddingBytes int

	// Connection Mode
	BeaconMode bool
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
