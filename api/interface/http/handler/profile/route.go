package profile

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts CMS profile section endpoints.
func RegisterRoutes(r chi.Router, h *ProfileHandler, mw *middleware.MiddlewareDeps, scope middleware.RouteScope) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/profile/list", h.ListProfilePublic)
		r.Get("/public/profile/{id}", h.GetProfilePublic)
	default:
		// Read
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("profile", "read"))
			r.Get("/profile", h.ListProfile)
			r.Get("/profile/{id}", h.GetProfile)
			r.Get("/profile/sections/names", h.GetProfileSectionNames)
		})
		// Write
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("profile", "write"))
			r.Post("/profile", h.CreateProfile)
			r.Put("/profile/{id}", h.UpdateProfile)
			r.Delete("/profile/{id}", h.DeleteProfile)
		})
	}
}