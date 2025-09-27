package main

import (
	"context"
	"numinon-ui/internal/connection"
	"numinon-ui/internal/models"
)

// App struct
type App struct {
	ctx     context.Context
	connMgr *connection.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		connMgr: connection.NewManager(),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.connMgr.Startup(ctx)
}

// --- Connection Methods (Frontend Accessible) ---

// Connect to the C2 server
func (a *App) Connect(serverURL string) models.ConnectionStatus {
	return a.connMgr.Connect(serverURL)
}

// Disconnect from the C2 server
func (a *App) Disconnect() models.ConnectionStatus {
	return a.connMgr.Disconnect()
}

// GetConnectionStatus returns current connection status
func (a *App) GetConnectionStatus() models.ConnectionStatus {
	return a.connMgr.GetStatus()
}

// GetAgents returns list of agents
func (a *App) GetAgents() []models.Agent {
	return a.connMgr.GetAgents()
}

// SendCommand sends a command to an agent
func (a *App) SendCommand(request models.CommandRequest) models.CommandResponse {
	return a.connMgr.SendCommand(request)
}
