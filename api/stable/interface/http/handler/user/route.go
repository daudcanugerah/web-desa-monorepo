package user

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts /users/* endpoints.
// Caller must have already applied mw.Auth() and mw.ProtectedRateLimit().
func RegisterRoutes(r chi.Router, h *UserHandler, mw *middleware.MiddlewareDeps) {
	// Self-service (no RBAC, just JWT)
	r.Get("/users/me", h.GetCurrentUser)
	r.Put("/users/me", h.UpdateCurrentUserProfile)

	// Read
	r.Group(func(r chi.Router) {
		r.Use(mw.RBAC("users", "read"))
		r.Get("/users", h.ListUsers)
		r.Get("/users/{id}", h.GetUser)
	})

	// Write
	r.Group(func(r chi.Router) {
		r.Use(mw.RBAC("users", "write"))
		r.Post("/users", h.CreateUser)
		r.Put("/users/{id}", h.UpdateUser)
		r.Delete("/users/{id}", h.DeleteUser)
		r.Put("/users/{id}/password", h.UpdatePassword)
	})

	// Note: /users/{id}/roles* is registered in role/route.go (RegisterUserRoleRoutes)
	// since the implementation lives in RoleHandler.
}