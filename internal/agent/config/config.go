package config

import "time"

type AgentConfig struct {
	ServerIP        string
	ServerPort      string
	CheckInEndpoint string
	Delay           time.Duration
	Jitter          float64
}
