package fasilitas

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts fasilitas + fasilitas-category endpoints.
func RegisterRoutes(
	r chi.Router,
	h *FasilitasHandler,
	cat *FasilitasCategoryHandler,
	upload *FasilitasUploadHandler,
	mw *middleware.MiddlewareDeps,
	scope middleware.RouteScope,
) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/fasilitas/list", h.ListFasilitasPublic)
		r.Get("/public/fasilitas/{id}", h.GetFasilitasPublic)
		r.Get("/public/fasilitas/categories", cat.ListFasilitasCategories)
	default:
		// Read
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("fasilitas", "read"))
			r.Get("/fasilitas", h.ListFasilitas)
			r.Get("/fasilitas/{id}", h.GetFasilitas)
		})
		// Write
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("fasilitas", "write"))
			r.Post("/fasilitas", h.CreateFasilitas)
			r.Put("/fasilitas/{id}", h.UpdateFasilitas)
			r.Delete("/fasilitas/{id}", h.DeleteFasilitas)
			r.Post("/fasilitas/upload-media", upload.UploadMedia)
			r.Delete("/fasilitas/{id}/images/{imageIndex}", h.RemoveImage)
			r.Post("/fasilitas/categories", cat.CreateFasilitasCategory)
			r.Delete("/fasilitas/categories/{id}", cat.DeleteFasilitasCategory)
		})
	}
}