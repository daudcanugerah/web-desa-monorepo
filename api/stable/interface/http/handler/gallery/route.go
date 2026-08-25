package gallery

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	h *GalleryHandler,
	mw *middleware.MiddlewareDeps,
	scope middleware.RouteScope,
) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/gallery/folders", h.ListFoldersPublic)
		r.Get("/public/gallery/folders/{id}", h.GetFolderPublic)
		r.Get("/public/gallery/media/{id}", h.GetMediaPublic)
		r.Get("/public/gallery/media/{id}/content", h.ServeMediaContentPublic)
		r.Get("/public/gallery/media/{id}/thumbnail", h.ServeMediaThumbnailPublic)
	default:
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("gallery", "read"))
			r.Get("/gallery/folders", h.ListFoldersAdmin)
			r.Get("/gallery/folders/{id}", h.GetFolderAdmin)
			r.Get("/gallery/media", h.ListMediaAdmin)
			r.Get("/gallery/media/{id}", h.GetMediaAdmin)
			r.Get("/gallery/media/{id}/content", h.ServeMediaContentAdmin)
			r.Get("/gallery/media/{id}/thumbnail", h.ServeMediaThumbnailAdmin)
		})
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("gallery", "write"))
			r.Post("/gallery/folders", h.CreateFolder)
			r.Put("/gallery/folders/{id}", h.UpdateFolder)
			r.Patch("/gallery/folders/{id}/visibility", h.UpdateFolderVisibility)
			r.Delete("/gallery/folders/{id}", h.DeleteFolder)
			r.Post("/gallery/folders/{id}/media", h.UploadMedia)
			r.Patch("/gallery/media/{id}/visibility", h.UpdateMediaVisibility)
			r.Post("/gallery/folders/{id}/media/visibility", h.BulkUpdateMediaVisibility)
			r.Delete("/gallery/media/{id}", h.DeleteMedia)
			r.Post("/gallery/media/{id}/regenerate-thumbnail", h.RegenerateMediaThumbnail)
		})
	}
}
