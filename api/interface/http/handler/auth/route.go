package auth

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts /auth/* endpoints.
// Caller must apply middleware.AuthRateLimit (or pass `DisableRateLimit=true`)
// before invoking this.
func RegisterRoutes(r chi.Router, h *AuthHandler) {
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.RefreshToken)
	r.Post("/auth/password-reset/request", h.RequestPasswordReset)
	r.Get("/auth/password-reset/check", h.CheckResetToken)
	r.Post("/auth/password-reset/confirm", h.ConfirmPasswordReset)
}

// Compile-time guard that mw is referenced (keeps import alive in some configs)
var _ = middleware.AuthRateLimitMiddleware