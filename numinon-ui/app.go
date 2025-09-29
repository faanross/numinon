package main

import (
	"context"
	"numinon-ui/internal/connection"
	"numinon-ui/internal/frontend"
	"numinon-ui/internal/lifecycle"
	"numinon-ui/internal/models"
)

// App struct - this is the main application structure
type App struct {
	ctx               context.Context
	connectionManager *connection.Manager
	lifecycleManager  *lifecycle.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		connectionManager: connection.NewManager(),
		lifecycleManager:  lifecycle.NewManager(),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize lifecycle manager (handles config, tray, window state)
	if err := a.lifecycleManager.Startup(ctx); err != nil {
		// Log error but continue - app can work without some features
		println("Lifecycle startup error:", err.Error())
	}

	// Initialize connection manager
	a.connectionManager.Startup(ctx)
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	// Lifecycle manager handles shutdown
	a.lifecycleManager.Shutdown()
}

// beforeClose is called when the window is about to close
// Returns true to prevent the window from closing
func (a *App) beforeClose(ctx context.Context) bool {
	if a.lifecycleManager != nil {
		return a.lifecycleManager.BeforeClose(ctx)
	}
	return false
}

// --- Configuration Methods (exposed to frontend) ---

// GetPreferences returns the current app preferences
func (a *App) GetPreferences() map[string]interface{} {
	if a.lifecycleManager == nil {
		return map[string]interface{}{}
	}

	cfg := a.lifecycleManager.GetConfig()
	if cfg == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"theme":      cfg.Theme,
		"general":    cfg.General,
		"connection": cfg.Connection,
	}
}

// UpdatePreferences updates app preferences
func (a *App) UpdatePreferences(prefs map[string]interface{}) error {
	if a.lifecycleManager == nil {
		return nil
	}
	return a.lifecycleManager.UpdateConfig(prefs)
}

// --- Connection Methods ---

func (a *App) Connect(serverURL string) frontend.ConnectionStatusDTO {
	status := a.connectionManager.Connect(serverURL)
	return frontend.ToConnectionStatusDTO(status)
}

func (a *App) Disconnect() frontend.ConnectionStatusDTO {
	status := a.connectionManager.Disconnect()
	return frontend.ToConnectionStatusDTO(status)
}

func (a *App) GetConnectionStatus() frontend.ConnectionStatusDTO {
	status := a.connectionManager.GetStatus()
	return frontend.ToConnectionStatusDTO(status)
}

func (a *App) GetAgents() []frontend.AgentDTO {
	agents := a.connectionManager.GetAgents()
	return frontend.ToAgentDTOs(agents)
}

func (a *App) SendCommand(req frontend.CommandRequestDTO) frontend.CommandResponseDTO {
	// Convert DTO to internal model
	cmdReq := models.CommandRequest{
		AgentID:   req.AgentID,
		Command:   req.Command,
		Arguments: req.Arguments,
	}

	response := a.connectionManager.SendCommand(cmdReq)
	return frontend.ToCommandResponseDTO(response)
}

// GetServerMessages returns recent server messages
// This method ensures Wails generates the TypeScript type for ServerMessageDTO
func (a *App) GetServerMessages() []frontend.ServerMessageDTO {
	// In a real implementation, you might store and return actual messages
	// For now, return empty array - this method mainly exists to generate the TypeScript type
	return []frontend.ServerMessageDTO{}
}
