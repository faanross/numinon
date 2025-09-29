package tray

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"numinon-ui/internal/config"
)

// Manager handles system tray functionality
// Note: Full system tray support in Wails v2 is limited
// This implementation uses window minimize/hide as a workaround
type Manager struct {
	ctx    context.Context
	config *config.AppConfig

	// State tracking
	isVisible  bool
	isQuitting bool
}

// NewManager creates a new tray manager
func NewManager(cfg *config.AppConfig) *Manager {
	return &Manager{
		config:    cfg,
		isVisible: true,
	}
}

// Setup initializes the tray-like behavior
func (m *Manager) Setup(ctx context.Context) error {
	m.ctx = ctx

	// Wails v2 doesn't have full system tray support yet
	// We'll use window events to simulate tray behavior

	return nil
}

// showWindow brings the window to front
func (m *Manager) ShowWindow() {
	runtime.WindowShow(m.ctx)
	m.isVisible = true
}

// hideWindow hides the window (simulates minimize to tray)
func (m *Manager) HideWindow() {
	if m.config.General.MinimizeToTray {
		runtime.WindowHide(m.ctx)
		m.isVisible = false

		// Show notification if enabled
		if m.config.General.ShowNotifications {
			m.showNotification("Numinon is still running",
				"The app has been minimized to the background")
		}
	} else {
		runtime.WindowMinimise(m.ctx)
	}
}

// toggleWindow shows or hides the main window
func (m *Manager) ToggleWindow() {
	if m.isVisible {
		m.HideWindow()
	} else {
		m.ShowWindow()
	}
}

// showNotification emits an event for the frontend to handle
func (m *Manager) showNotification(title, message string) {
	runtime.EventsEmit(m.ctx, "notification", map[string]string{
		"title":   title,
		"message": message,
	})
}

// IsQuitting returns whether the app is shutting down
func (m *Manager) IsQuitting() bool {
	return m.isQuitting
}

// SetQuitting marks the app as quitting
func (m *Manager) SetQuitting(quitting bool) {
	m.isQuitting = quitting
}
