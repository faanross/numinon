package config

import "time"

type AgentConfig struct {
	// core info
	Protocol   AgentProtocol `yaml:"protocol"`
	ServerIP   string        `yaml:"server_ip"`
	ServerPort string        `yaml:"server_port"`

	// endpoints
	CheckInEndpoint   string `yaml:"check_in_endpoint"`
	ResultsEndpoint   string `yaml:"results_endpoint"`
	WebsocketEndpoint string `yaml:"websocket_endpoint"`

	// agent properties
	Delay  time.Duration `yaml:"delay"`
	Jitter float64       `yaml:"jitter"`

	// New fields related to payload jitter
	CheckinMethod   string `yaml:"checkin_method"`
	EnablePadding   bool   `yaml:"enable_padding"`
	MinPaddingBytes int    `yaml:"min_padding_bytes"`
	MaxPaddingBytes int    `yaml:"max_padding_bytes"`

	// Connection Mode
	BeaconMode bool `yaml:"beacon_mode"`

	// Agent ID
	UUID string `yaml:"uuid"`

	//Self-signed TLS allowance
	SkipVerifyTLS bool `yaml:"skip_verify_tls"`
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
