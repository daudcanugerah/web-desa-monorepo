package health

import (
	"net/http"

	"webdesa/api/pkg/response"
)

// HealthHandler handles HTTP requests for health check operations.
type HealthHandler struct{}

// NewHealthHandler creates a new health check handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns 200 OK if the service is alive. No authentication required.
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.HealthResponse
// @Router       /health [get]
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
