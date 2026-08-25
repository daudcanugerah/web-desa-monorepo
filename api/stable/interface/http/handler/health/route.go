package health

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts the /health endpoint.
func RegisterRoutes(r chi.Router, h *HealthHandler) {
	r.Get("/health", h.HealthCheck)
}