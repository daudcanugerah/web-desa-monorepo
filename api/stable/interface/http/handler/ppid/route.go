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
	mw *middleware.MiddlewareDeps,
	scope middleware.RouteScope,
) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/ppid/list", h.ListPPIDPublic)
		r.Get("/public/ppid/{id}", h.GetPPIDPublic)
		r.Get("/public/ppid/categories", cat.ListPPIDCategories)
		r.Post("/public/ppid/{id}/requests", h.CreatePPIDRequest)
		r.Get("/ppid/document/{documentId}/download", h.DownloadDocument)
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
			r.Post("/ppid/requests/{id}/approve", h.ApproveRequest)
			r.Post("/ppid/requests/{id}/revoke", h.RevokeRequest)
			r.Post("/ppid/categories", cat.CreatePPIDCategory)
			r.Delete("/ppid/categories/{id}", cat.DeletePPIDCategory)
		})
	}
}