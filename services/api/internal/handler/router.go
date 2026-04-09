package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter wires the Handler to a chi router and returns it.
func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", h.Health)
	r.Route("/api", func(r chi.Router) {
		r.Get("/items", h.ListItems)
		r.Post("/items", h.CreateItem)
	})

	return r
}
