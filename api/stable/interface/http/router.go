package http

import (
	"net/http"
	"time"

	"webdesa/api/interface/http/handler/auth"
	"webdesa/api/interface/http/handler/banner"
	"webdesa/api/interface/http/handler/berita"
	"webdesa/api/interface/http/handler/desa"
	"webdesa/api/interface/http/handler/fasilitas"
	"webdesa/api/interface/http/handler/file"
	"webdesa/api/interface/http/handler/gallery"
	"webdesa/api/interface/http/handler/health"
	"webdesa/api/interface/http/handler/infographic"
	"webdesa/api/interface/http/handler/ppid"
	"webdesa/api/interface/http/handler/profile"
	"webdesa/api/interface/http/handler/role"
	"webdesa/api/interface/http/handler/struktur"
	"webdesa/api/interface/http/handler/umkm"
	"webdesa/api/interface/http/handler/user"
	"webdesa/api/interface/http/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RouterConfig holds all dependencies needed to configure the HTTP router.
// All handlers are injected via this config following dependency injection principles.
type RouterConfig struct {
	// Handlers (one per feature package)
	AuthHandler                *auth.AuthHandler
	UserHandler                *user.UserHandler
	RoleHandler                *role.RoleHandler
	BannerHandler              *banner.BannerHandler
	BannerCategoryHandler      *banner.BannerCategoryHandler
	BeritaHandler              *berita.BeritaHandler
	BeritaCategoryHandler      *berita.BeritaCategoryHandler
	BeritaUploadHandler        *berita.BeritaUploadHandler
	UMKMHandler                *umkm.UMKMHandler
	UMKMCategoryHandler        *umkm.UMKMCategoryHandler
	FasilitasHandler           *fasilitas.FasilitasHandler
	FasilitasCategoryHandler   *fasilitas.FasilitasCategoryHandler
	GalleryHandler             *gallery.GalleryHandler
	PPIDHandler                *ppid.PPIDHandler
	PPIDCategoryHandler        *ppid.PPIDCategoryHandler
	StrukturHandler            *struktur.StrukturHandler
	DesaHandler                *desa.DesaHandler
	ProfileHandler             *profile.ProfileHandler
	InfographicHandler         *infographic.InfographicHandler
	InfographicCategoryHandler *infographic.InfographicCategoryHandler
	HealthHandler              *health.HealthHandler
	PublicFileHandler          *file.FileHandler

	// Config
	UploadPublicDirectory string

	// Middleware dependencies
	Enforcer       *casbin.Enforcer
	JWTSecret      string
	AllowedOrigins []string
	Logger         middleware.Logger

	// DisableRateLimit disables rate limiting (useful for tests)
	DisableRateLimit bool

	// EnableSwagger mounts /swagger/* for interactive API docs
	EnableSwagger bool
}

// NewRouter creates and configures the HTTP router with all endpoints and middleware.
//
// The route table is split per-feature in interface/http/handler/<feature>/route.go.
// Each feature package exports one RegisterRoutes function called from the
// appropriate scope (public or protected) below.
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RecoveryMiddleware(cfg.Logger))
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware(cfg.Logger))
	r.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Top-level routes
	health.RegisterRoutes(r, cfg.HealthHandler)

	fileServer := http.FileServer(http.Dir(cfg.UploadPublicDirectory))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	if cfg.EnableSwagger {
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
		r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./docs/swagger.json")
		})
	}

	file.RegisterRoutes(r, cfg.PublicFileHandler)

	// /api/v1 scope tree
	r.Route("/api/v1", func(r chi.Router) {
		mw := &middleware.MiddlewareDeps{
			Enforcer:  cfg.Enforcer,
			JWTSecret: cfg.JWTSecret,
		}

		// ---- PUBLIC SCOPE ----
		if !cfg.DisableRateLimit {
			r.Use(middleware.PublicRateLimitMiddleware())
		}
		banner.RegisterRoutes(r, cfg.BannerHandler, cfg.BannerCategoryHandler, mw, middleware.ScopePublic)
		berita.RegisterRoutes(r, cfg.BeritaHandler, cfg.BeritaCategoryHandler, cfg.BeritaUploadHandler, mw, middleware.ScopePublic)
		umkm.RegisterRoutes(r, cfg.UMKMHandler, cfg.UMKMCategoryHandler, mw, middleware.ScopePublic)
		fasilitas.RegisterRoutes(r, cfg.FasilitasHandler, cfg.FasilitasCategoryHandler, mw, middleware.ScopePublic)
		gallery.RegisterRoutes(r, cfg.GalleryHandler, mw, middleware.ScopePublic)
		ppid.RegisterRoutes(r, cfg.PPIDHandler, cfg.PPIDCategoryHandler, mw, middleware.ScopePublic)
		struktur.RegisterRoutes(r, cfg.StrukturHandler, mw, middleware.ScopePublic)
		profile.RegisterRoutes(r, cfg.ProfileHandler, mw, middleware.ScopePublic)
		infographic.RegisterRoutes(r, cfg.InfographicHandler, cfg.InfographicCategoryHandler, mw, middleware.ScopePublic)
		desa.RegisterRoutes(r, cfg.DesaHandler, mw, middleware.ScopePublic)

		// ---- AUTH SCOPE ----
		r.Group(func(r chi.Router) {
			if !cfg.DisableRateLimit {
				r.Use(middleware.AuthRateLimitMiddleware())
			}
			auth.RegisterRoutes(r, cfg.AuthHandler)
		})

		// ---- PROTECTED SCOPE ----
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth())
			if !cfg.DisableRateLimit {
				r.Use(mw.ProtectedRateLimit())
			}
			user.RegisterRoutes(r, cfg.UserHandler, mw)
			role.RegisterRoutes(r, cfg.RoleHandler, mw)
			role.RegisterUserRoleRoutes(r, cfg.RoleHandler, mw)
			banner.RegisterRoutes(r, cfg.BannerHandler, cfg.BannerCategoryHandler, mw, middleware.ScopeProtected)
			berita.RegisterRoutes(r, cfg.BeritaHandler, cfg.BeritaCategoryHandler, cfg.BeritaUploadHandler, mw, middleware.ScopeProtected)
			umkm.RegisterRoutes(r, cfg.UMKMHandler, cfg.UMKMCategoryHandler, mw, middleware.ScopeProtected)
			fasilitas.RegisterRoutes(r, cfg.FasilitasHandler, cfg.FasilitasCategoryHandler, mw, middleware.ScopeProtected)
			gallery.RegisterRoutes(r, cfg.GalleryHandler, mw, middleware.ScopeProtected)
			ppid.RegisterRoutes(r, cfg.PPIDHandler, cfg.PPIDCategoryHandler, mw, middleware.ScopeProtected)
			struktur.RegisterRoutes(r, cfg.StrukturHandler, mw, middleware.ScopeProtected)
			desa.RegisterRoutes(r, cfg.DesaHandler, mw, middleware.ScopeProtected)
			profile.RegisterRoutes(r, cfg.ProfileHandler, mw, middleware.ScopeProtected)
			infographic.RegisterRoutes(r, cfg.InfographicHandler, cfg.InfographicCategoryHandler, mw, middleware.ScopeProtected)
		})
	})

	return r
}
