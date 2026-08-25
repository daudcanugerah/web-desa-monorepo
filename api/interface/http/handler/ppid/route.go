package ppid

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts PPID document + request + download + category endpoints.
func RegisterRoutes(
	r chi.Router,
	h *PPIDHandler,
	cat *PPIDCategoryHandler,
	upload *PPIDUploadHandler,
	mw *middleware.MiddlewareDeps,
	scope middleware.RouteScope,
) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/ppid/list", h.ListPPIDPublic)
		r.Get("/public/ppid/{id}", h.GetPPIDPublic)
		r.Get("/public/ppid/categories", cat.ListPPIDCategories)
		r.Post("/public/ppid/{id}/requests", h.CreatePPIDRequest)
		// Task 7.4: removed /ppid/document/{documentId}/download?token=...
		// The legacy token mechanism is replaced by the unified signed-URL
		// flow: ApproveRequest embeds `/api/v1/media/{id}/content?jwt=...`
		// in the approval email. Any in-flight legacy links from before
		// the migration now 404 (their tokens are signed with the old
		// schema and won't verify against the unified handler).
	default:
		// Read
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("ppid", "read"))
			r.Get("/ppid", h.ListPPID)
			r.Get("/ppid/{id}", h.GetPPID)
			r.Get("/ppid/requests", h.ListPPIDRequests)
		})
		// Write
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("ppid", "write"))
			r.Post("/ppid", h.CreatePPID)
			r.Put("/ppid/{id}", h.UpdatePPID)
			r.Delete("/ppid/{id}", h.DeletePPID)
			r.Post("/ppid/upload-media", upload.UploadDocument)
			r.Post("/ppid/upload-thumbnail", upload.UploadThumbnail)
			r.Post("/ppid/upload", upload.Upload)
			r.Post("/ppid/requests/{id}/approve", h.ApproveRequest)
			r.Post("/ppid/requests/{id}/revoke", h.RevokeRequest)
			r.Post("/ppid/categories", cat.CreatePPIDCategory)
			r.Delete("/ppid/categories/{id}", cat.DeletePPIDCategory)
		})
	}
}