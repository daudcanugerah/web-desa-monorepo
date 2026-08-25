package umkm

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts UMKM + UMKM-category endpoints.
func RegisterRoutes(
	r chi.Router,
	h *UMKMHandler,
	cat *UMKMCategoryHandler,
	upload *UMKMUploadHandler,
	mw *middleware.MiddlewareDeps,
	scope middleware.RouteScope,
) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/umkm/list", h.ListUMKMPublic)
		r.Get("/public/umkm/{id}", h.GetUMKMPublic)
		r.Get("/public/umkm/categories", cat.ListUMKMCategories)
	default:
		// Read
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("umkm", "read"))
			r.Get("/umkm", h.ListUMKM)
			r.Get("/umkm/{id}", h.GetUMKM)
		})
		// Write
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("umkm", "write"))
			r.Post("/umkm", h.CreateUMKM)
			r.Put("/umkm/{id}", h.UpdateUMKM)
			r.Delete("/umkm/{id}", h.DeleteUMKM)
			r.Post("/umkm/upload-media", upload.UploadMedia)
			r.Post("/umkm/categories", cat.CreateUMKMCategory)
			r.Delete("/umkm/categories/{id}", cat.DeleteUMKMCategory)
		})
	}
}