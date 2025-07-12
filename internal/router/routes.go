package router

import (
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r *chi.Mux) {
	// HTTP-based endpoints
	r.Get("/", CheckinHandler)
	r.Post("/", CheckinHandler)

	r.Post("/results", ResultsHandler)

	// WS-based endpoint
	r.Get("/ws", WSHandler)
}
