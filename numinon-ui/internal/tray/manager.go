package tray

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"numinon-ui/internal/config"
)

//go:embed icon.png
var icon []byte

// Manager handles system tray functionality
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

// Setup initializes the system tray
func (m *Manager) Setup(ctx context.Context) error {
	m.ctx = ctx

	// Create the tray menu
	menu := m.createMenu()

	// Set the tray with our icon and menu
	runtime.SetSystemTray(m.ctx, runtime.SystemTray{
		Icon:    icon, // We'll need to add an icon file
		Tooltip: "Numinon C2 Client",
		Menu:    menu,
	})

	// Set up tray click handler (single click to show/hide)
	runtime.OnSystemTrayClick(m.ctx, func() {
		m.toggleWindow()
	})

	// Set up tray right-click handler (show menu)
	runtime.OnSystemTrayRightClick(m.ctx, func() {
		runtime.ShowSystemTrayMenu(m.ctx)
	})

	return nil
}

// createMenu builds the tray menu structure
func (m *Manager) createMenu() *runtime.Menu {
	menu := runtime.NewMenu()

	// Show/Hide item (dynamically updates)
	if m.isVisible {
		menu.AddText("Hide Window", nil, func(cd *runtime.CallbackData) {
			m.hideWindow()
		})
	} else {
		menu.AddText("Show Window", nil, func(cd *runtime.CallbackData) {
			m.showWindow()
		})
	}

	menu.AddSeparator()

	// Connection status submenu
	connectionMenu := menu.AddSubmenu("Connection")
	connectionMenu.AddText("Connected to: "+m.config.Connection.ServerURL, nil, nil)
	connectionMenu.AddSeparator()
	connectionMenu.AddText("Disconnect", nil, func(cd *runtime.CallbackData) {
		runtime.EventsEmit(m.ctx, "tray:disconnect")
	})

	menu.AddSeparator()

	// Settings
	menu.AddText("Settings...", nil, func(cd *runtime.CallbackData) {
		m.showWindow()
		runtime.EventsEmit(m.ctx, "tray:show-settings")
	})

	menu.AddSeparator()

	// Quit
	menu.AddText("Quit Numinon", nil, func(cd *runtime.CallbackData) {
		m.quit()
	})

	return menu
}

// toggleWindow shows or hides the main window
func (m *Manager) toggleWindow() {
	if m.isVisible {
		m.hideWindow()
	} else {
		m.showWindow()
	}
}

// showWindow brings the window to front
func (m *Manager) showWindow() {
	runtime.Show(m.ctx)
	m.isVisible = true
	m.updateMenu() // Update menu text
}

// hideWindow hides the window to tray
func (m *Manager) hideWindow() {
	if m.config.General.MinimizeToTray {
		runtime.Hide(m.ctx)
		m.isVisible = false
		m.updateMenu() // Update menu text

		// Show notification on first minimize
		if m.config.General.ShowNotifications {
			m.showNotification("Numinon is still running",
				"The app has been minimized to the system tray")
		}
	} else {
		runtime.Minimise(m.ctx)
	}
}

// updateMenu refreshes the tray menu
func (m *Manager) updateMenu() {
	menu := m.createMenu()
	runtime.SetSystemTrayMenu(m.ctx, menu)
}

// showNotification displays a system notification
func (m *Manager) showNotification(title, message string) {
	// Note: Wails doesn't have built-in notifications yet,
	// but we can emit an event for the frontend to handle
	runtime.EventsEmit(m.ctx, "notification", map[string]string{
		"title":   title,
		"message": message,
	})
}

// quit handles application shutdown
func (m *Manager) quit() {
	m.isQuitting = true
	runtime.Quit(m.ctx)
}

// IsQuitting returns whether the app is shutting down
func (m *Manager) IsQuitting() bool {
	return m.isQuitting
}
