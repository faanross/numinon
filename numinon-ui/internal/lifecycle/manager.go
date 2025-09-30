package lifecycle

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"numinon-ui/internal/config"
	"numinon-ui/internal/tray"
	"numinon-ui/internal/window"
	"sync"
)

// Manager coordinates application lifecycle
type Manager struct {
	ctx    context.Context
	config *config.AppConfig

	// Sub-managers
	trayManager   *tray.Manager
	windowManager *window.Manager

	// State
	mu           sync.RWMutex
	isReady      bool
	isQuitting   bool // Add this flag to track quitting state
	shutdownOnce sync.Once
}

// NewManager creates a new lifecycle manager
func NewManager() *Manager {
	return &Manager{
		isQuitting: false,
	}
}

// Startup is called when the application starts
func (m *Manager) Startup(ctx context.Context) error {
	m.ctx = ctx

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		runtime.LogWarning(ctx, "Failed to load config, using defaults: "+err.Error())
		cfg = config.DefaultConfig()
	}
	m.config = cfg

	// Initialize sub-managers
	m.trayManager = tray.NewManager(cfg)
	m.windowManager = window.NewManager(cfg)

	// Set up tray-like behavior (limited in Wails v2)
	if err := m.trayManager.Setup(ctx); err != nil {
		runtime.LogError(ctx, "Failed to setup tray manager: "+err.Error())
	}

	// Set up window management
	if err := m.windowManager.Setup(ctx); err != nil {
		runtime.LogError(ctx, "Failed to setup window manager: "+err.Error())
	}

	// Mark as ready
	m.mu.Lock()
	m.isReady = true
	m.mu.Unlock()

	// Emit ready event
	runtime.EventsEmit(ctx, "app:ready", true)

	// Log successful startup
	runtime.LogInfo(ctx, "Application started successfully")

	return nil
}

// BeforeClose is called when the window is about to close
// Return true to prevent close, false to allow it
func (m *Manager) BeforeClose(ctx context.Context) bool {
	m.mu.RLock()
	isQuitting := m.isQuitting
	m.mu.RUnlock()

	// If we're already quitting, don't prevent the close
	if isQuitting {
		runtime.LogInfo(ctx, "Window closing - app is quitting")
		return false // Allow close
	}

	// Check if we should minimize to tray instead of closing
	if m.config != nil && m.config.General.MinimizeToTray {
		runtime.LogInfo(ctx, "Minimizing to tray instead of closing")
		runtime.WindowHide(ctx)

		// Show notification on first minimize
		if m.config.General.ShowNotifications {
			runtime.EventsEmit(ctx, "notification", map[string]string{
				"title":   "Running in background",
				"message": "Numinon has been minimized to the background. Click the app icon to restore.",
				"type":    "info",
			})
		}
		return true // Prevent close
	}

	// If not minimizing to tray, proceed with shutdown
	runtime.LogInfo(ctx, "Closing window and shutting down")
	go m.Shutdown() // Run shutdown asynchronously to avoid blocking
	return false    // Allow close
}

// Shutdown performs graceful shutdown
func (m *Manager) Shutdown() {
	m.shutdownOnce.Do(func() {
		runtime.LogInfo(m.ctx, "Starting graceful shutdown...")

		// Mark as quitting first
		m.mu.Lock()
		m.isQuitting = true
		m.mu.Unlock()

		// Also mark tray manager as quitting
		if m.trayManager != nil {
			m.trayManager.SetQuitting(true)
		}

		// Save window state before closing
		if m.windowManager != nil {
			m.windowManager.SaveCurrentState()
		}

		// Emit shutdown event for frontend
		runtime.EventsEmit(m.ctx, "app:shutdown", true)

		// Save current configuration
		if m.config != nil {
			if err := m.config.Save(); err != nil {
				runtime.LogError(m.ctx, "Failed to save config on shutdown: "+err.Error())
			} else {
				runtime.LogInfo(m.ctx, "Configuration saved successfully")
			}
		}

		// Clean up resources
		runtime.LogInfo(m.ctx, "Cleanup complete, quitting application")

		// Quit the application
		runtime.Quit(m.ctx)
	})
}

// ForceShutdown provides a way to force shutdown (e.g., from menu or keyboard shortcut)
func (m *Manager) ForceShutdown() {
	runtime.LogInfo(m.ctx, "Force shutdown requested")
	m.Shutdown()
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *config.AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig updates and saves configuration
func (m *Manager) UpdateConfig(updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure config exists
	if m.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Apply updates to config based on the key
	if generalUpdates, ok := updates["general"].(map[string]interface{}); ok {
		if val, ok := generalUpdates["minimizeToTray"].(bool); ok {
			m.config.General.MinimizeToTray = val
		}
		if val, ok := generalUpdates["startOnBoot"].(bool); ok {
			m.config.General.StartOnBoot = val
		}
		if val, ok := generalUpdates["showNotifications"].(bool); ok {
			m.config.General.ShowNotifications = val
		}
	}

	if connectionUpdates, ok := updates["connection"].(map[string]interface{}); ok {
		if val, ok := connectionUpdates["serverUrl"].(string); ok {
			m.config.Connection.ServerURL = val
		}
		if val, ok := connectionUpdates["autoConnect"].(bool); ok {
			m.config.Connection.AutoConnect = val
		}
		if val, ok := connectionUpdates["reconnectDelay"].(float64); ok {
			m.config.Connection.ReconnectDelay = int(val)
		}
	}

	if themeUpdates, ok := updates["theme"].(map[string]interface{}); ok {
		if val, ok := themeUpdates["mode"].(string); ok {
			m.config.Theme.Mode = val
		}
		if val, ok := themeUpdates["accentColor"].(string); ok {
			m.config.Theme.AccentColor = val
		}
	}

	// Save configuration
	if err := m.config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Emit config change event
	runtime.EventsEmit(m.ctx, "config:changed", m.config)

	return nil
}

// IsReady returns whether the app is fully initialized
func (m *Manager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isReady
}

// IsQuitting returns whether the app is in the process of quitting
func (m *Manager) IsQuitting() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isQuitting
}
