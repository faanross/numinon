package router

import (
	"github.com/faanross/numinon/internal/listener"
	"github.com/faanross/numinon/internal/tracker"
	"github.com/go-chi/chi/v5"
)

// SetupRoutesWithManagerAndTracker sets up routes with FULL support
func SetupRoutesWithManagerAndTracker(r *chi.Mux, listenerMgr *listener.Manager, agentTracker *tracker.Tracker) {
	// Set global tracker reference for handlers to use
	AgentTracker = agentTracker

	// Initialize with both manager and tracker
	InitializeTaskManagementWithFullSupport(listenerMgr, agentTracker)

	setupHandlers(r)
}

// setupHandlers contains the actual route definitions (extracted to avoid duplication)
func setupHandlers(r *chi.Mux) {
	// HTTP-based endpoints
	r.Get("/", CheckinHandler)
	r.Post("/", CheckinHandler)

	r.Post("/results", ResultsHandler)

	// WS-based endpoint
	r.Get("/ws", WSHandler)

	// Debug endpoint for task statistics (remove in production)
	r.Get("/debug/tasks", TaskStatsHandler)
}
