package server

import (
	"net/http"
	"numinon_shadow/internal/agentstatemanager"
	"numinon_shadow/internal/clientapi"
	"numinon_shadow/internal/clientmanager"
	"numinon_shadow/internal/listener"
	"numinon_shadow/internal/listenermanager"
	"numinon_shadow/internal/orchestration"
	"numinon_shadow/internal/taskbroker"
	"numinon_shadow/internal/taskmanager"
	"numinon_shadow/internal/tracker"
)

// Dependencies holds all the core server components that various parts of the system need.
// This is our "dependency injection container" - instead of global variables,
// we pass this struct to components that need these services.
//
// Think of this as the "wiring diagram" of our server - it shows all the major
// components and makes their relationships explicit.
type Dependencies struct {
	// Core Infrastructure
	TaskStore     taskmanager.TaskManager // The actual task storage system
	Orchestrators *orchestration.Registry // Command-specific logic
	AgentTracker  *tracker.Tracker        // Tracks agent connections
	ListenerMgr   *listener.Manager       // Manages network listeners

	// Operator Layer
	ClientSessions clientapi.ClientSessionManager // Manages operator WebSocket connections
	TaskBroker     clientapi.TaskBroker           // Bridges operators to tasks (TaskStore)

	// These wrap existing systems for operator consumption
	ListenerAPI   clientapi.ListenerManager   // Operator-friendly listener control
	AgentStateAPI clientapi.AgentStateManager // Operator-friendly agent info
}

func NewDependencies(router http.Handler, existingTracker *tracker.Tracker) *Dependencies {
	// Create core infrastructure
	taskStore := taskmanager.NewMemoryTaskStore()
	orchestrators := orchestration.NewRegistry()

	// If no tracker provided, create one
	agentTracker := existingTracker
	if agentTracker == nil {
		agentTracker = tracker.NewTracker()
	}

	// Create listener manager with router and tracker
	listenerMgr := listener.NewManager(router, agentTracker)

	// Register all orchestrators
	downloadOrch := orchestration.NewDownloadOrchestrator("./downloads")
	uploadOrch := orchestration.NewUploadOrchestrator()
	// ... register all orchestrators ...

	// Create operator-layer wrappers
	listenerAPI := listenermanager.NewOperatorListenerManager(listenerMgr)
	agentStateAPI := agentstatemanager.NewOperatorAgentManager(agentTracker)

	// Create client session manager (needs to be created before task broker)
	// We'll need to modify clientmanager.New to return the interface type
	clientSessions := clientmanager.New(listenerAPI, nil, agentStateAPI) // nil for taskBroker temporarily

	// Create task broker with all its dependencies
	taskBroker := taskbroker.NewBroker(taskStore, orchestrators, clientSessions)

	// Now update client sessions with the task broker
	// (We'll need to add a SetTaskBroker method to clientmanager)

	return &Dependencies{
		TaskStore:      taskStore,
		Orchestrators:  orchestrators,
		AgentTracker:   agentTracker,
		ListenerMgr:    listenerMgr,
		ClientSessions: clientSessions,
		TaskBroker:     taskBroker,
		ListenerAPI:    listenerAPI,
		AgentStateAPI:  agentStateAPI,
	}
}
