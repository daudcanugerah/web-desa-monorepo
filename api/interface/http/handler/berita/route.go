package berita

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts berita + berita-category + berita-upload endpoints.
// Pass ScopePublic or ScopeProtected.
//
// Task 7.4: removed `media *gallery.FeaturePublicMediaHandler` parameter
// and the /public/berita/media/{id}/content|thumbnail routes.
func RegisterRoutes(
	r chi.Router,
	h *BeritaHandler,
	cat *BeritaCategoryHandler,
	up *BeritaUploadHandler,
	mw *middleware.MiddlewareDeps,
	scope middleware.RouteScope,
) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/berita/list", h.ListBeritaPublic)
		r.Get("/public/berita/{id}", h.GetBeritaPublic)
		r.Get("/public/berita/categories", cat.ListBeritaCategories)
	default:
		// Read
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("berita", "read"))
			r.Get("/berita", h.ListBerita)
			r.Get("/berita/{id}", h.GetBerita)
		})
		// Write + media upload + categories
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("berita", "write"))
			r.Post("/berita", h.CreateBerita)
			r.Put("/berita/{id}", h.UpdateBerita)
			r.Patch("/berita/{id}/status", h.UpdateBeritaStatus)
			r.Delete("/berita/{id}", h.DeleteBerita)
			r.Post("/berita/upload-media", up.UploadMedia)
			r.Post("/berita/categories", cat.CreateBeritaCategory)
			r.Delete("/berita/categories/{id}", cat.DeleteBeritaCategory)
		})
	}
}
