package health

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts the /health endpoint.
func RegisterRoutes(r chi.Router, h *HealthHandler) {
	r.Get("/health", h.HealthCheck)
	// Versioned alias so client apps that use the /api/v1 prefix can
	// hit the same liveness probe without special-casing the root.
	r.Get("/api/v1/health", h.HealthCheckV1)
}