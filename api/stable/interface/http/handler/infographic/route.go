package infographic

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts infographic + infographic-category endpoints.
func RegisterRoutes(
	r chi.Router,
	h *InfographicHandler,
	cat *InfographicCategoryHandler,
	mw *middleware.MiddlewareDeps,
	scope middleware.RouteScope,
) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/infographic/list", h.ListInfographicPublic)
		r.Get("/public/infographic/{id}", h.GetInfographicPublic)
		r.Get("/public/infographic/categories", cat.ListInfographicCategories)
	default:
		// Read
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("infographic", "read"))
			r.Get("/infographic", h.ListInfographic)
			r.Get("/infographic/{id}", h.GetInfographic)
			r.Get("/infographic/sections/names", h.GetInfographicSectionNames)
			r.Get("/infographic/access-logs", h.ListAccessLogs)
		})
		// Write
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("infographic", "write"))
			r.Post("/infographic", h.CreateInfographic)
			r.Post("/infographic/preview/token", h.GeneratePreviewToken)
			r.Put("/infographic/{id}", h.UpdateInfographic)
			r.Delete("/infographic/{id}", h.DeleteInfographic)
			r.Post("/infographic/categories", cat.CreateInfographicCategory)
			r.Delete("/infographic/categories/{id}", cat.DeleteInfographicCategory)
		})
	}
}