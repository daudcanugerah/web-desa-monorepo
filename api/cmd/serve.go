package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/spf13/cobra"
	_ "webdesa/api/docs" // generated swagger docs

	"webdesa/api/interface/email"
	filehandler "webdesa/api/interface/file"
	httpserver "webdesa/api/interface/http"
	authhandler "webdesa/api/interface/http/handler/auth"
	bannerhandler "webdesa/api/interface/http/handler/banner"
	beritacategoryhandler "webdesa/api/interface/http/handler/berita"
	beritahandler "webdesa/api/interface/http/handler/berita"
	desahandler "webdesa/api/interface/http/handler/desa"
	fasilitashandler "webdesa/api/interface/http/handler/fasilitas"
	filehandlerpkg "webdesa/api/interface/http/handler/file"
	galleryhandler "webdesa/api/interface/http/handler/gallery"
	healthhandler "webdesa/api/interface/http/handler/health"
	infographiccategoryhandler "webdesa/api/interface/http/handler/infographic"
	infographichandler "webdesa/api/interface/http/handler/infographic"
	ppidcategoryhandler "webdesa/api/interface/http/handler/ppid"
	ppidhandler "webdesa/api/interface/http/handler/ppid"
	profilehandler "webdesa/api/interface/http/handler/profile"
	rolehandler "webdesa/api/interface/http/handler/role"
	strukturhandler "webdesa/api/interface/http/handler/struktur"
	umkmcategoryhandler "webdesa/api/interface/http/handler/umkm"
	umkmhandler "webdesa/api/interface/http/handler/umkm"
	userhandler "webdesa/api/interface/http/handler/user"
	authpg "webdesa/api/interface/postgres"
	repopg "webdesa/api/interface/postgres"
	"webdesa/api/interface/rbac"
	"webdesa/api/pkg/clock"
	"webdesa/api/usecase/auth"
	"webdesa/api/usecase/banner"
	"webdesa/api/usecase/bannercategory"
	"webdesa/api/usecase/berita"
	"webdesa/api/usecase/beritacategory"
	"webdesa/api/usecase/desa"
	"webdesa/api/usecase/fasilitas"
	"webdesa/api/usecase/fasilitascategory"
	galleryusecase "webdesa/api/usecase/gallery"
	"webdesa/api/usecase/infographic"
	"webdesa/api/usecase/infographiccategory"
	"webdesa/api/usecase/ppid"
	"webdesa/api/usecase/ppidcategory"
	"webdesa/api/usecase/profile"
	"webdesa/api/usecase/role"
	"webdesa/api/usecase/struktur"
	"webdesa/api/usecase/umkm"
	"webdesa/api/usecase/umkmcategory"
	"webdesa/api/usecase/user"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server",
	Long:  "Start the HTTP server and serve API requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer(cmd.Context())
	},
}

func runServer(ctx context.Context) error {
	mainOtel.Log.Info(ctx, "Starting Desa API server")
	if err := systemConfig.FileUpload.ValidateDirectoryIsolation(); err != nil {
		return errtrace.Wrap(fmt.Errorf("invalid file upload configuration: %w", err))
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(systemConfig.Postgres.GetMaxOpenConns())
	db.SetMaxIdleConns(systemConfig.Postgres.GetMaxIdleConns())
	db.SetConnMaxLifetime(systemConfig.Postgres.GetConnMaxLifetime())

	mainOtel.Log.Info(ctx, "Connected to database successfully")

	// Run migrations
	mainOtel.Log.Info(ctx, "Running database migrations")
	if err := goose.Up(db.DB, "db/migrations"); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to run migrations: %w", err))
	}
	mainOtel.Log.Info(ctx, "Migrations completed successfully")

	// Initialize Casbin enforcer (implements usecase/role.Enforcer interface)
	// Uses PostgreSQL adapter for persistent policy storage
	casbinEnforcer, err := rbac.NewCasbinEnforcer("rbac/rbac_model.conf", db)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create Casbin enforcer: %w", err))
	}
	mainOtel.Log.Info(ctx, "Casbin enforcer initialized")

	// Get the underlying Casbin enforcer for middleware
	// The CasbinEnforcer wrapper implements GetEnforcer() method
	casbinEnforcerImpl := casbinEnforcer.(*rbac.CasbinEnforcer)

	// Initialize shared utilities
	clk := clock.RealClock{}

	galleryOriginalStorage, err := filehandler.NewLocalHandlerWithSubdir(systemConfig.FileUpload.GetPrivateUploadDirectory(), "gallery/originals")
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create gallery original storage: %w", err))
	}
	galleryThumbnailStorage, err := filehandler.NewLocalHandlerWithSubdir(systemConfig.FileUpload.GetPrivateUploadDirectory(), "gallery/thumbnails")
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create gallery thumbnail storage: %w", err))
	}

	// Initialize repositories (implement usecase repository interfaces)
	// Dependencies: interface/postgres → usecase → domain
	userRepo := authpg.NewUserRepository(db)
	authRepo := authpg.NewAuthRepository(db)
	roleRepo := repopg.NewRoleRepository(db)
	bannerRepo := repopg.NewBannerRepository(db)
	bannerCategoryRepo := repopg.NewBannerCategoryRepository(db)
	beritaRepo := repopg.NewBeritaRepository(db)
	beritaCategoryRepo := repopg.NewBeritaCategoryRepository(db)
	umkmRepo := repopg.NewUMKMRepository(db)
	umkmCategoryRepo := repopg.NewUMKMCategoryRepository(db)
	fasilitasRepo := repopg.NewFasilitasRepository(db)
	fasilitasCategoryRepo := repopg.NewFasilitasCategoryRepository(db)
	galleryRepo := repopg.NewGalleryRepository(db)
	ppidRepo := repopg.NewPPIDRepository(db)
	ppidCategoryRepo := repopg.NewPPIDCategoryRepository(db)
	strukturRepo := repopg.NewStrukturRepository(db)
	desaRepo := repopg.NewDesaRepository(db)
	profileRepo := repopg.NewProfileRepository(db)
	infographicRepo := repopg.NewInfographicRepository(db)
	infographicCategoryRepo := repopg.NewInfographicCategoryRepository(db)
	// backupRepo := repopg.NewBackupRepository(db) // Will be used when backup service is enabled

	// Village profile defaults (will be replaced with DB values below once
	// desaService is constructed). Used by email senders.
	emailVillageName := "Village Administration"
	emailSupportEmail := "support@village.go.id"
	emailWebsiteURL := systemConfig.App.DomainAddr

	// Auth email service for password reset (fixes bugs.md#1). Village
	// profile is loaded below; default values are used if not yet loaded.
	authEmailSvc := email.NewAuthEmailService(mainOtel.Log, &email.EmailConfig{
		VillageName:  emailVillageName,
		SupportEmail: emailSupportEmail,
		WebsiteURL:   emailWebsiteURL,
	}, &systemConfig.SMTP)

	// Task 7.1: build the SignedURLService. When SignedURLsEnabled is
	// false the signer stays nil and the unified media route returns
	// 503 — the legacy per-scope routes keep serving traffic.
	var signedURLService *galleryusecase.SignedURLService
	if systemConfig.Gallery.IsSignedURLsEnabled() {
		signedURLService = galleryusecase.NewSignedURLService(systemConfig.JWT.Secret, nil)
	}

	// Gallery service must be built first because the feature service
	// galleryService is built before feature services so the deletion
	// hub can be wired to feature repos below (Task 5.1 on-delete sweep).
	galleryService := galleryusecase.NewService(
		galleryRepo,
		galleryOriginalStorage,
		galleryThumbnailStorage,
		galleryusecase.NewImageProcessor(),
		galleryusecase.NewVideoProcessor(galleryusecase.VideoProcessorConfig{
			FfmpegPath:                 systemConfig.Gallery.GetFfmpegPath(),
			FfprobePath:                systemConfig.Gallery.GetFfprobePath(),
			ThumbnailMaxWidth:          systemConfig.Gallery.GetThumbnailMaxWidth(),
			ThumbnailMaxHeight:         systemConfig.Gallery.GetThumbnailMaxHeight(),
			ThumbnailQuality:           systemConfig.Gallery.GetThumbnailQuality(),
			VideoThumbnailFrameSeconds: systemConfig.Gallery.GetVideoThumbnailFrameSeconds(),
		}),
		systemConfig.Gallery,
		clk,
		galleryusecase.NewMediaDeletionHub(),
		// Task 7.1: SignedURLService backs the unified /api/v1/media/{id}/...?jwt= route.
		// The signer stays nil when SignedURLsEnabled is false — the
		// handler returns 503 so misconfig is loud.
		signedURLService,
	)
	// Register feature repos as gallery deletion listeners so the gallery
	// service can sweep dangling media ids out of `*_media_ids` JSONB
	// columns on media delete (Task 5.1).
	galleryService.Hub().AddListener(umkmRepo)
	galleryService.Hub().AddListener(fasilitasRepo)
	// fileStore exposes gallery.FileStore so every feature (banner, berita,
	// struktur, umkm, fasilitas, user, ppid) routes its uploads into the
	// matching system folder instead of the legacy public directory.
	fileStore := galleryusecase.NewFileStoreService(galleryService)

	// Initialize use case services (accept repository interfaces, return concrete service structs)
	// Dependencies: services → repository interfaces → domain
	authService := auth.NewService(
		authRepo,
		userRepo,
		authEmailSvc,
		clk,
		systemConfig.JWT.Secret,
		systemConfig.JWT.GetExpiration(),
		systemConfig.App.DomainAddr,
	)
	userService := user.NewService(userRepo, fileStore, clk)
	roleService := role.NewService(casbinEnforcer, roleRepo)

	// Category services must be constructed before content services
	// because those depend on them for FK validation.
	bannerCategoryService := bannercategory.NewService(bannerCategoryRepo)
	beritaCategoryService := beritacategory.NewService(beritaCategoryRepo, clk)
	bannerService := banner.NewService(bannerRepo, fileStore, bannerCategoryService, clk)
	umkmCategoryService := umkmcategory.NewService(umkmCategoryRepo, clk)
	ppidCategoryService := ppidcategory.NewService(ppidCategoryRepo, clk)
	fasilitasCategoryService := fasilitascategory.NewService(fasilitasCategoryRepo, clk)
	infographicCategoryService := infographiccategory.NewService(infographicCategoryRepo, clk)

	beritaService := berita.NewService(
		beritaRepo,
		fileStore,
		clk,
		systemConfig.FileUpload.GetPublicUploadDirectory(),
		beritaCategoryService,
	)
	umkmService := umkm.NewService(umkmRepo, fileStore, clk, umkmCategoryService, &systemConfig.Gallery)
	fasilitasService := fasilitas.NewService(fasilitasRepo, fileStore, clk, fasilitasCategoryService, &systemConfig.Gallery)
	_ = galleryService // already used to build fileStore; bound to keep alive in cmd scope

	// Village profile from settings table (used by email senders).
	// Defaults are used until the profile is loaded.
	villageName := "Village Administration"
	supportEmail := "support@village.go.id"
	websiteURL := systemConfig.App.DomainAddr

	desaService := desa.NewService(desaRepo, clk)
	if d, err := desaService.Get(context.Background()); err == nil && d != nil {
		if d.Name != "" {
			villageName = d.Name
		}
		if d.Email != nil && *d.Email != "" {
			supportEmail = *d.Email
		}
		if d.Website != nil && *d.Website != "" {
			websiteURL = *d.Website
		}
	}

	// PPID service with email support using village profile
	ppidEmailConfig := ppid.EmailConfig{
		VillageName:  villageName,
		SupportEmail: supportEmail,
		WebsiteURL:   websiteURL,
		DomainAddr:   systemConfig.App.DomainAddr,
	}
	ppidEmailSvc := email.NewPPIDEmailService(mainOtel.Log, &email.EmailConfig{
		VillageName:  ppidEmailConfig.VillageName,
		SupportEmail: ppidEmailConfig.SupportEmail,
		WebsiteURL:   ppidEmailConfig.WebsiteURL,
	}, &systemConfig.SMTP)
	ppidService := ppid.NewServiceWithSignedURL(ppidRepo, fileStore, ppidEmailSvc, clk, systemConfig.JWT.Secret, ppidEmailConfig, ppidCategoryService, signedURLService)

	profileService := profile.NewService(profileRepo, clk)
	infographicAccessLogRepo := repopg.NewInfographicAccessLogRepository(db)
	infographicService := infographic.NewService(infographicRepo, clk, systemConfig.Metabase, infographicCategoryService, infographicAccessLogRepo)
	strukturService := struktur.NewService(strukturRepo, fileStore, clk)

	// Parse database connection details for backup service
	// Backup service will be enabled when backup routes are added to router
	// dbHost, dbPort, dbName, dbUser, dbPassword := parseDSN(systemConfig.Postgres.DSN)
	// backupService := backup.NewService(
	// 	backupRepo,
	// 	systemConfig.Backup.GetDirectory(),
	// 	dbHost,
	// 	dbPort,
	// 	dbName,
	// 	dbUser,
	// 	dbPassword,
	// )

	// Initialize HTTP handlers (accept service structs from usecase layer)
	// Dependencies: handlers → services → repositories
	authHandler := authhandler.NewAuthHandler(authService)
	userHandler := userhandler.NewUserHandler(userService, roleService, signedURLService, systemConfig.Gallery.IsSignedURLsEnabled())
	userUploadHandler := userhandler.NewUserUploadHandler(fileStore, mainOtel.Log)
	roleHandler := rolehandler.NewRoleHandler(roleService)
	bannerHandler := bannerhandler.NewBannerHandler(bannerService, signedURLService, systemConfig.Gallery.IsSignedURLsEnabled())
	bannerCategoryHandler := bannerhandler.NewBannerCategoryHandler(bannerCategoryService)
	beritaHandler := beritahandler.NewBeritaHandler(beritaService, signedURLService, systemConfig.Gallery.IsSignedURLsEnabled())
	beritaCategoryHandler := beritacategoryhandler.NewBeritaCategoryHandler(beritaCategoryService)
	bannerUploadHandler := bannerhandler.NewBannerUploadHandler(fileStore, mainOtel.Log)
	beritaUploadHandler := beritahandler.NewBeritaUploadHandler(fileStore, mainOtel.Log)
	umkmHandler := umkmhandler.NewUMKMHandler(umkmService, signedURLService, systemConfig.Gallery.IsSignedURLsEnabled())
	umkmUploadHandler := umkmhandler.NewUMKMUploadHandler(fileStore, mainOtel.Log)
	umkmCategoryHandler := umkmcategoryhandler.NewUMKMCategoryHandler(umkmCategoryService)
	fasilitasHandler := fasilitashandler.NewFasilitasHandler(fasilitasService, signedURLService, systemConfig.Gallery.IsSignedURLsEnabled())
	fasilitasUploadHandler := fasilitashandler.NewFasilitasUploadHandler(fileStore, mainOtel.Log)
	fasilitasCategoryHandler := fasilitashandler.NewFasilitasCategoryHandler(fasilitasCategoryService)
	galleryHandler := galleryhandler.NewGalleryHandler(galleryService, systemConfig.Gallery, signedURLService, systemConfig.Gallery.IsSignedURLsEnabled())
	// Task 7.1: the unified /api/v1/media/{id}/...?jwt= handler. nil
	// when SignedURLsEnabled is false so misconfig is loud (503).
	signedMediaHandler := galleryhandler.NewSignedMediaHandler(galleryService, signedURLService, mainOtel.Log)
	ppidHandler := ppidhandler.NewPPIDHandler(ppidService, systemConfig.FileUpload.GetPPIDUploadDirectory(), mainOtel.Log, signedURLService, systemConfig.Gallery.IsSignedURLsEnabled())
	ppidCategoryHandler := ppidcategoryhandler.NewPPIDCategoryHandler(ppidCategoryService)
	ppidUploadHandler := ppidhandler.NewPPIDUploadHandler(fileStore, mainOtel.Log)
	strukturHandler := strukturhandler.NewStrukturHandler(strukturService, signedURLService, systemConfig.Gallery.IsSignedURLsEnabled())
	strukturUploadHandler := strukturhandler.NewStrukturUploadHandler(fileStore, mainOtel.Log)
	desaHandler := desahandler.NewDesaHandler(desaService)
	profileHandler := profilehandler.NewProfileHandler(profileService)
	infographicHandler := infographichandler.NewInfographicHandler(infographicService)
	infographicCategoryHandler := infographiccategoryhandler.NewInfographicCategoryHandler(infographicCategoryService)
	healthHandler := healthhandler.NewHealthHandler()
	fileHandler2 := filehandlerpkg.NewFileHandler(systemConfig.FileUpload.GetPublicUploadDirectory(), mainOtel.Log)
	// backupHandler := handler.NewBackupHandler(backupService) // Will be added to router when backup routes are enabled

	// Setup router with all dependencies
	router := httpserver.NewRouter(httpserver.RouterConfig{
		AuthHandler:                authHandler,
		UserHandler:                userHandler,
		UserUploadHandler:          userUploadHandler,
		RoleHandler:                roleHandler,
		BannerHandler:              bannerHandler,
		BannerCategoryHandler:      bannerCategoryHandler,
		BannerUploadHandler:        bannerUploadHandler,
		BeritaHandler:              beritaHandler,
		BeritaCategoryHandler:      beritaCategoryHandler,
		BeritaUploadHandler:        beritaUploadHandler,
		UMKMHandler:                umkmHandler,
		UMKMUploadHandler:          umkmUploadHandler,
		UMKMCategoryHandler:        umkmCategoryHandler,
		FasilitasHandler:           fasilitasHandler,
		FasilitasUploadHandler:     fasilitasUploadHandler,
		FasilitasCategoryHandler:   fasilitasCategoryHandler,
		GalleryHandler:             galleryHandler,
		SignedMediaHandler:         signedMediaHandler,
		PPIDHandler:                ppidHandler,
		PPIDCategoryHandler:        ppidCategoryHandler,
		PPIDUploadHandler:          ppidUploadHandler,
		StrukturHandler:            strukturHandler,
		StrukturUploadHandler:      strukturUploadHandler,
		DesaHandler:                desaHandler,
		ProfileHandler:             profileHandler,
		InfographicHandler:         infographicHandler,
		InfographicCategoryHandler: infographicCategoryHandler,
		HealthHandler:              healthHandler,
		PublicFileHandler:          fileHandler2,
		UploadPublicDirectory:      systemConfig.FileUpload.GetPublicUploadDirectory(),
		Enforcer:                   casbinEnforcerImpl.GetEnforcer(),
		JWTSecret:                  systemConfig.JWT.Secret,
		AllowedOrigins:             systemConfig.CORS.GetAllowedOrigins(),
		EnableSwagger:              true,
		Logger:                     mainOtel.Log,
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         systemConfig.App.HTTPAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		mainOtel.Log.Infof(ctx, "HTTP server listening on %s", systemConfig.App.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return errtrace.Wrap(fmt.Errorf("server error: %w", err))

	case sig := <-shutdown:
		mainOtel.Log.Infof(ctx, "Received signal %v, starting graceful shutdown", sig)

		// Give outstanding requests 30 seconds to complete
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return errtrace.Wrap(fmt.Errorf("graceful shutdown failed: %w", err))
		}

		mainOtel.Log.Info(ctx, "Server stopped gracefully")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

// parseDSN parses a PostgreSQL DSN string and extracts connection details.
// DSN format: "postgres://user:password@host:port/dbname?params"
// Returns: host, port, dbname, user, password
func parseDSN(dsn string) (host, port, dbname, user, password string) {
	// Default values
	host = "localhost"
	port = "5432"
	dbname = "postgres"
	user = "postgres"
	password = ""

	// Simple DSN parsing for postgres://user:password@host:port/dbname format
	// This is a basic implementation - for production, consider using url.Parse
	if len(dsn) == 0 {
		return
	}

	// Remove postgres:// prefix if present
	dsn = dsn[len("postgres://"):]

	// Split user:password@host:port/dbname
	atIndex := -1
	for i, c := range dsn {
		if c == '@' {
			atIndex = i
			break
		}
	}

	if atIndex > 0 {
		// Extract user:password
		userPass := dsn[:atIndex]
		colonIndex := -1
		for i, c := range userPass {
			if c == ':' {
				colonIndex = i
				break
			}
		}
		if colonIndex > 0 {
			user = userPass[:colonIndex]
			password = userPass[colonIndex+1:]
		} else {
			user = userPass
		}

		// Extract host:port/dbname
		remaining := dsn[atIndex+1:]
		slashIndex := -1
		for i, c := range remaining {
			if c == '/' {
				slashIndex = i
				break
			}
		}

		if slashIndex > 0 {
			hostPort := remaining[:slashIndex]
			colonIndex := -1
			for i, c := range hostPort {
				if c == ':' {
					colonIndex = i
					break
				}
			}
			if colonIndex > 0 {
				host = hostPort[:colonIndex]
				port = hostPort[colonIndex+1:]
			} else {
				host = hostPort
			}

			// Extract dbname (remove query params if present)
			dbnameWithParams := remaining[slashIndex+1:]
			questionIndex := -1
			for i, c := range dbnameWithParams {
				if c == '?' {
					questionIndex = i
					break
				}
			}
			if questionIndex > 0 {
				dbname = dbnameWithParams[:questionIndex]
			} else {
				dbname = dbnameWithParams
			}
		}
	}

	return
}
