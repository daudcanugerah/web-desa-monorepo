package banner

import (
	"webdesa/api/interface/http/middleware"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts banner + banner-category + banner-upload endpoints.
// Pass ScopePublic to mount only /public/banner/* (no auth required)
// or ScopeProtected to mount only /banners/* (requires JWT + RBAC).
//
// Task 7.4: removed `media *gallery.FeaturePublicMediaHandler` parameter and
// the /public/banner/media/{id}/content|thumbnail routes — every public
// banner media stream now flows through the unified
// /api/v1/media/{id}/...?jwt= handler.
func RegisterRoutes(
	r chi.Router,
	h *BannerHandler,
	cat *BannerCategoryHandler,
	up *BannerUploadHandler,
	mw *middleware.MiddlewareDeps,
	scope middleware.RouteScope,
) {
	switch scope {
	case middleware.ScopePublic:
		r.Get("/public/banner", h.ListBannersPublic)
		r.Get("/public/banner/{id}", h.GetBannerPublic)
		r.Get("/banners/active", h.GetActiveBanners)
		r.Get("/public/banner/categories", cat.ListBannerCategories)
	default:
		// Read
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("banners", "read"))
			r.Get("/banners", h.ListBanners)
			r.Get("/banners/{id}", h.GetBanner)
			r.Get("/banners/categories", cat.ListBannerCategories)
		})
		// Write
		r.Group(func(r chi.Router) {
			r.Use(mw.RBAC("banners", "write"))
			r.Post("/banners", h.CreateBanner)
			r.Put("/banners/{id}", h.UpdateBanner)
			r.Patch("/banners/{id}/status", h.UpdateBannerStatus)
			r.Delete("/banners/{id}", h.DeleteBanner)
			r.Post("/banners/upload-media", up.UploadMedia)
			r.Post("/banners/categories", cat.CreateBannerCategory)
			r.Delete("/banners/categories/{id}", cat.DeleteBannerCategory)
		})
	}
}