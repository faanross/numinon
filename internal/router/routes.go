package router

import (
	"github.com/go-chi/chi/v5"
	"numinon_shadow/internal/listener"
)

// SetupRoutesWithManager sets up routes with listener management capability
func SetupRoutesWithManager(r *chi.Mux, listenerMgr *listener.Manager) {
	// Initialize with listener manager for HOP support
	InitializeTaskManagement(listenerMgr)

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
