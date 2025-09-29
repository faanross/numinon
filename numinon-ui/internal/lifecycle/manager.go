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
	shutdownOnce sync.Once
}

// NewManager creates a new lifecycle manager
func NewManager() *Manager {
	return &Manager{}
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
// Return true to prevent close
func (m *Manager) BeforeClose(ctx context.Context) bool {
	// Check if we should minimize to tray instead of closing
	if m.config.General.MinimizeToTray && !m.trayManager.IsQuitting() {
		runtime.WindowHide(ctx)

		// Show notification on first minimize
		if m.config.General.ShowNotifications {
			runtime.EventsEmit(ctx, "notification", map[string]string{
				"title":   "Running in background",
				"message": "Numinon has been minimized to the background. Use Alt+Tab to restore.",
			})
		}
		return true // Prevent close
	}

	// Save window state before closing
	m.windowManager.SaveCurrentState()

	return false // Allow close
}

// Shutdown performs graceful shutdown
func (m *Manager) Shutdown() {
	m.shutdownOnce.Do(func() {
		runtime.LogInfo(m.ctx, "Shutting down application...")

		// Mark as quitting so BeforeClose doesn't prevent it
		if m.trayManager != nil {
			m.trayManager.SetQuitting(true)
		}

		// Emit shutdown event for frontend
		runtime.EventsEmit(m.ctx, "app:shutdown", true)

		// Save current configuration
		if m.config != nil {
			if err := m.config.Save(); err != nil {
				runtime.LogError(m.ctx, "Failed to save config on shutdown: "+err.Error())
			}
		}

		// Clean up resources
		// (Add any cleanup code here)

		// Quit the application
		runtime.Quit(m.ctx)
	})
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

	// Apply updates to config
	// (This is simplified - in production you'd want proper validation)

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
