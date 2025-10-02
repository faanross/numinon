package main

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"numinon-ui/internal/connection"
	"numinon-ui/internal/frontend"
	"numinon-ui/internal/lifecycle"
	"numinon-ui/internal/models"
	"numinon-ui/internal/websocket"
)

// App struct - this is the main application structure
type App struct {
	ctx               context.Context
	connectionManager *connection.Manager
	lifecycleManager  *lifecycle.Manager
	wsManager         *websocket.Manager
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

	// Initialize WebSocket manager
	a.wsManager = websocket.NewManager(ctx, websocket.DefaultConfig())
	a.setupWebSocketHandlers()

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

// setupWebSocketHandlers registers WebSocket message handlers
func (a *App) setupWebSocketHandlers() {
	// Handle agent list updates
	a.wsManager.RegisterHandler(websocket.TypeAgentList, func(msg *websocket.Message) error {
		// Convert and emit to frontend
		runtime.EventsEmit(a.ctx, "agent_update", msg.Payload)
		return nil
	})

	// Handle task results
	a.wsManager.RegisterHandler(websocket.TypeTaskResult, func(msg *websocket.Message) error {
		runtime.EventsEmit(a.ctx, "task:complete", msg.Payload)
		return nil
	})
}

// --- NEW WebSocket Methods (exposed to frontend) ---

// ConnectWebSocket establishes WebSocket connection to C2 server
func (a *App) ConnectWebSocket(serverURL string) map[string]interface{} {
	err := a.wsManager.Connect(serverURL)

	result := map[string]interface{}{
		"success": err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
	} else {
		result["status"] = a.wsManager.GetStatus()
	}

	return result
}

// DisconnectWebSocket closes WebSocket connection
func (a *App) DisconnectWebSocket() map[string]interface{} {
	err := a.wsManager.Disconnect()

	return map[string]interface{}{
		"success": err == nil,
		"error":   err,
	}
}

// GetWebSocketStatus returns WebSocket connection status
func (a *App) GetWebSocketStatus() map[string]interface{} {
	return a.wsManager.GetStatus()
}

// SendWebSocketMessage sends a message through WebSocket
func (a *App) SendWebSocketMessage(msgType string, payload interface{}) map[string]interface{} {
	msg, err := a.wsManager.Send(websocket.MessageType(msgType), payload)

	result := map[string]interface{}{
		"success": err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
	} else {
		result["message"] = msg
	}

	return result
}

// ExecuteAgentTask sends a task execution request
func (a *App) ExecuteAgentTask(agentID string, taskType string, params map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{
		"agentId":    agentID,
		"taskType":   taskType,
		"parameters": params,
	}

	msg, err := a.wsManager.Send(websocket.TypeTaskExecute, payload)

	result := map[string]interface{}{
		"success": err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
	} else {
		result["taskId"] = msg.ID
		result["message"] = "Task queued for execution"
	}

	return result
}
