package router

import (
	"github.com/go-chi/chi/v5"
)

// SetupHandlers contains the actual route definitions
func SetupHandlers(r *chi.Mux) {
	// HTTP-based endpoints
	r.Get("/", CheckinHandler)
	r.Post("/", CheckinHandler)

	r.Post("/results", ResultsHandler)

	// WS-based endpoint
	r.Get("/ws", WSHandler)

	// Debug endpoint for task statistics
	r.Get("/debug/tasks", TaskStatsHandler)
}
