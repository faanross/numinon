package window

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"numinon-ui/internal/config"
	"time"
)

// Manager handles window state persistence
type Manager struct {
	ctx    context.Context
	config *config.AppConfig

	// Debounce saving to avoid too many writes
	saveTimer   *time.Timer
	savePending bool
}

// NewManager creates a new window manager
func NewManager(cfg *config.AppConfig) *Manager {
	return &Manager{
		config: cfg,
	}
}

// Setup initializes window management
func (m *Manager) Setup(ctx context.Context) error {
	m.ctx = ctx

	// Restore window state from config
	m.restoreWindowState()

	// Note: Window event handling in Wails v2 is done differently
	// We'll handle this in the main app with the window runtime options

	return nil
}

// restoreWindowState applies saved window configuration
func (m *Manager) restoreWindowState() {
	// Set window size
	runtime.WindowSetSize(m.ctx, m.config.Window.Width, m.config.Window.Height)

	// Set window position
	runtime.WindowSetPosition(m.ctx, m.config.Window.X, m.config.Window.Y)

	// Handle maximized state
	if m.config.Window.Maximized {
		runtime.WindowMaximise(m.ctx)
	}

	// Handle start hidden
	if m.config.Window.StartHidden {
		runtime.WindowHide(m.ctx)
	}
}

// SaveCurrentState saves the current window state
func (m *Manager) SaveCurrentState() {
	// Don't save if maximized (we handle that separately)
	isMaximized := runtime.WindowIsMaximised(m.ctx)
	if isMaximized {
		m.config.Window.Maximized = true
		m.debouncedSave()
		return
	}

	// Get current window position
	x, y := runtime.WindowGetPosition(m.ctx)
	m.config.Window.X = x
	m.config.Window.Y = y

	// Get current window size
	width, height := runtime.WindowGetSize(m.ctx)
	m.config.Window.Width = width
	m.config.Window.Height = height
	m.config.Window.Maximized = false

	// Save with debouncing
	m.debouncedSave()
}

// debouncedSave saves config after a delay to avoid excessive writes
func (m *Manager) debouncedSave() {
	// Cancel existing timer if any
	if m.saveTimer != nil {
		m.saveTimer.Stop()
	}

	// Set new timer
	m.saveTimer = time.AfterFunc(500*time.Millisecond, func() {
		if err := m.config.Save(); err != nil {
			// Log error but don't crash
			runtime.LogError(m.ctx, "Failed to save window state: "+err.Error())
		}
	})
}

// GetWindowState returns current window state for frontend
func (m *Manager) GetWindowState() config.WindowConfig {
	return m.config.Window
}

// SetAlwaysOnTop toggles always-on-top mode
func (m *Manager) SetAlwaysOnTop(enabled bool) {
	runtime.WindowSetAlwaysOnTop(m.ctx, enabled)
}
