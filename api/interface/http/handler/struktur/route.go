package struktur

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts struktur organisasi + struktur-upload endpoints.
func RegisterRoutes(r chi.Router, h *StrukturHandler, up *StrukturUploadHandler, mw *middleware.MiddlewareDeps, scope middleware.RouteScope) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/struktur/list", h.ListStrukturPublic)
		r.Get("/public/struktur/{id}", h.GetStrukturPublic)
	default:
		// Read
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("struktur", "read"))
			r.Get("/struktur", h.ListStruktur)
			r.Get("/struktur/{id}", h.GetStruktur)
		})
		// Write
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("struktur", "write"))
			r.Post("/struktur", h.CreateStruktur)
			r.Put("/struktur/{id}", h.UpdateStruktur)
			r.Delete("/struktur/{id}", h.DeleteStruktur)
			r.Post("/struktur/upload-media", up.UploadMedia)
		})
	}
}