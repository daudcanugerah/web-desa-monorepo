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
		// Task 7.4: removed /public/gallery/folders (list/get public folder
		// info); removed /public/gallery/media/{id}/content|thumbnail and
		// /public/gallery/media/{id} — every public media stream now goes
		// through the unified /api/v1/media/{id}/...?jwt= handler.
	default:
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("gallery", "read"))
			r.Get("/gallery/folders", h.ListFoldersAdmin)
			r.Get("/gallery/folders/{id}", h.GetFolderAdmin)
			r.Get("/gallery/media", h.ListMediaAdmin)
			r.Get("/gallery/media/{id}", h.GetMediaAdmin)
			// Task 7.4: removed /gallery/media/{id}/content|thumbnail — the
			// admin URL now flows through /api/v1/media/{id}/...?jwt=
			// (scope=admin). The JWT replaces the bearer middleware gate.
		})
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("gallery", "write"))
			r.Post("/gallery/folders", h.CreateFolder)
			r.Put("/gallery/folders/{id}", h.UpdateFolder)
			r.Patch("/gallery/folders/{id}/visibility", h.UpdateFolderVisibility)
			r.Patch("/gallery/folders/{id}/cover", h.SetFolderCover)
			r.Delete("/gallery/folders/{id}", h.DeleteFolder)
			r.Post("/gallery/folders/{id}/media", h.UploadMedia)
			r.Patch("/gallery/media/{id}/visibility", h.UpdateMediaVisibility)
			r.Post("/gallery/folders/{id}/media/visibility", h.BulkUpdateMediaVisibility)
			r.Delete("/gallery/media/{id}", h.DeleteMedia)
			r.Post("/gallery/media/{id}/regenerate-thumbnail", h.RegenerateMediaThumbnail)
		})
	}
}

// RegisterSignedMediaRoutes mounts the unified /api/v1/media/{id}/...
// routes introduced in Task 7.1. Caller passes a *SignedMediaHandler;
// routes stay reachable regardless of RouteScope so that anonymous
// public traffic and authenticated admin traffic both reach the same
// endpoint. The signed URL in the query string carries the auth — no
// per-route RBAC middleware needed.
func RegisterSignedMediaRoutes(r chi.Router, smh *SignedMediaHandler) {
	r.Get("/media/{id}/content", smh.ServeContent)
	r.Get("/media/{id}/thumbnail", smh.ServeThumbnail)
	// Task 7.3: clients POST here with their current ?jwt= to get a
	// fresh token before it expires. The response carries the new
	// content + thumbnail URLs.
	r.Post("/media/refresh", smh.RefreshMedia)
}
