package desa

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the village profile endpoints.
func RegisterRoutes(r chi.Router, h *DesaHandler, up *DesaUploadHandler, mw *middleware.MiddlewareDeps, scope middleware.RouteScope) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/desa", h.GetDesaPublic)
	default:
		// Read + write (only one endpoint each)
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("desa", "read"))
			r.Get("/desa", h.GetDesa)
		})
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("desa", "write"))
			r.Put("/desa", h.UpdateDesa)
			r.Post("/desa/upload-media", up.UploadMedia)
		})
	}
}
