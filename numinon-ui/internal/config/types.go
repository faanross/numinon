package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig holds all application configuration
type AppConfig struct {
	Window     WindowConfig     `json:"window"`
	Connection ConnectionConfig `json:"connection"`
	Theme      ThemeConfig      `json:"theme"`
	General    GeneralConfig    `json:"general"`
}

// WindowConfig stores window preferences
type WindowConfig struct {
	Width       int  `json:"width"`
	Height      int  `json:"height"`
	X           int  `json:"x"`
	Y           int  `json:"y"`
	Maximized   bool `json:"maximized"`
	StartHidden bool `json:"startHidden"` // Start minimized to tray
}

// ConnectionConfig stores C2 connection preferences
type ConnectionConfig struct {
	ServerURL        string   `json:"serverUrl"`
	AutoConnect      bool     `json:"autoConnect"`
	ReconnectDelay   int      `json:"reconnectDelay"` // seconds
	RecentServers    []string `json:"recentServers"`
	MaxRecentServers int      `json:"maxRecentServers"`
}

// ThemeConfig stores UI theme preferences
type ThemeConfig struct {
	Mode        string `json:"mode"`        // "dark", "light", "auto"
	AccentColor string `json:"accentColor"` // hex color
}

// GeneralConfig stores general app preferences
type GeneralConfig struct {
	MinimizeToTray    bool   `json:"minimizeToTray"`
	StartOnBoot       bool   `json:"startOnBoot"`
	ShowNotifications bool   `json:"showNotifications"`
	Language          string `json:"language"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Window: WindowConfig{
			Width:       1200,
			Height:      800,
			X:           100,
			Y:           100,
			Maximized:   false,
			StartHidden: false,
		},
		Connection: ConnectionConfig{
			ServerURL:        "ws://localhost:8080/client",
			AutoConnect:      false,
			ReconnectDelay:   5,
			RecentServers:    []string{},
			MaxRecentServers: 5,
		},
		Theme: ThemeConfig{
			Mode:        "dark",
			AccentColor: "#667eea",
		},
		General: GeneralConfig{
			MinimizeToTray:    true,
			StartOnBoot:       false,
			ShowNotifications: true,
			Language:          "en",
		},
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	// Get user's config directory (OS-specific)
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	// Create our app's config directory
	appConfigDir := filepath.Join(configDir, "numinon-ui")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(appConfigDir, "config.json"), nil
}

// Load reads configuration from disk
func Load() (*AppConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// Save writes configuration to disk
func (c *AppConfig) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Marshal to JSON with pretty formatting
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddRecentServer adds a server to the recent servers list
func (c *AppConfig) AddRecentServer(serverURL string) {
	// Remove if already exists
	for i, url := range c.Connection.RecentServers {
		if url == serverURL {
			c.Connection.RecentServers = append(
				c.Connection.RecentServers[:i],
				c.Connection.RecentServers[i+1:]...,
			)
			break
		}
	}

	// Add to front
	c.Connection.RecentServers = append([]string{serverURL}, c.Connection.RecentServers...)

	// Trim to max size
	if len(c.Connection.RecentServers) > c.Connection.MaxRecentServers {
		c.Connection.RecentServers = c.Connection.RecentServers[:c.Connection.MaxRecentServers]
	}
}
