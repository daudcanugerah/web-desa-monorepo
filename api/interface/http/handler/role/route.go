package role

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts /roles/* and /permissions endpoints.
// Caller must have already applied mw.Auth() and mw.ProtectedRateLimit().
func RegisterRoutes(r chi.Router, h *RoleHandler, mw *middleware.MiddlewareDeps) {
	// Read
	r.Group(func(r chi.Router) {
		r.Use(mw.RBAC("roles", "read"))
		r.Get("/roles", h.ListRoles)
		r.Get("/roles/{role}/permissions", h.GetRolePermissions)
		r.Get("/permissions", h.GetAvailablePermissions)
	})

	// Write
	r.Group(func(r chi.Router) {
		r.Use(mw.RBAC("roles", "write"))
		r.Post("/roles", h.CreateRole)
		r.Post("/roles/{role}/permissions", h.AddPermissionToRole)
		r.Delete("/roles/{role}/permissions/{permission}", h.RemovePermissionFromRole)
		r.Delete("/roles/{role}", h.DeleteRole)
	})
}

// RegisterUserRoleRoutes mounts /users/{id}/roles* endpoints.
// Implementation lives on RoleHandler because it's RBAC assignment logic.
func RegisterUserRoleRoutes(r chi.Router, h *RoleHandler, mw *middleware.MiddlewareDeps) {
	r.Group(func(r chi.Router) {
		r.Use(mw.RBAC("roles", "write"))
		r.Post("/users/{id}/roles", h.AssignRoleToUser)
		r.Delete("/users/{id}/roles/{role}", h.RemoveRoleFromUser)
	})
}