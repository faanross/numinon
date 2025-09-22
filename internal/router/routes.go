package router

import (
	"github.com/go-chi/chi/v5"
)

//// SetupRoutesWithManagerAndTracker sets up routes with FULL support
//func SetupRoutesWithManagerAndTracker(r *chi.Mux, listenerMgr *listener.Manager, agentTracker *tracker.Tracker) {
//	// Set global tracker reference for handlers to use
//	AgentTracker = agentTracker
//
//	// Initialize with both manager and tracker
//	InitializeTaskManagementWithFullSupport(listenerMgr, agentTracker)
//
//	setupHandlers(r)
//}

//// SetupHandlers sets up the HTTP routes without reinitializing dependencies
//func SetupHandlers(r *chi.Mux) {
//	setupHandlers(r) // Call the existing private function
//}

// SetupHandlers contains the actual route definitions (extracted to avoid duplication)
func SetupHandlers(r *chi.Mux) {
	// HTTP-based endpoints
	r.Get("/", CheckinHandler)
	r.Post("/", CheckinHandler)

	r.Post("/results", ResultsHandler)

	// WS-based endpoint
	r.Get("/ws", WSHandler)

	// Debug endpoint for task statistics (remove in production)
	r.Get("/debug/tasks", TaskStatsHandler)
}
