package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"webdesa/api/config"
	"webdesa/api/interface/file"
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
	"webdesa/api/usecase/profilecategory"
	"webdesa/api/usecase/role"
	"webdesa/api/usecase/struktur"
	"webdesa/api/usecase/umkm"
	"webdesa/api/usecase/umkmcategory"
	"webdesa/api/usecase/user"
)

// globalTS is the single shared test server for the entire test suite.
var globalTS *TestServer

// TestMain spins up one Postgres container for the whole suite, runs all tests,
// then tears everything down. This avoids the ~5s container startup cost per test.
//
// Set INTEGRATION_TEST_DB_URL to skip testcontainers and use an existing Postgres
// (useful in environments without Docker, e.g. CI without testcontainers support).
func TestMain(m *testing.M) {
	ctx := context.Background()

	var connStr string
	var pgContainer *postgres.PostgresContainer

	if existing := os.Getenv("INTEGRATION_TEST_DB_URL"); existing != "" {
		// Use externally-provided database
		connStr = existing
		fmt.Fprintf(os.Stderr, "Using INTEGRATION_TEST_DB_URL (skipping testcontainers)\n")
	} else {
		var err error
		pgContainer, err = postgres.Run(ctx,
			"postgres:16-alpine",
			postgres.WithDatabase("testdb"),
			postgres.WithUsername("testuser"),
			postgres.WithPassword("testpass"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second)),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start PostgreSQL container: %v\n", err)
			os.Exit(1)
		}

		connStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get connection string: %v\n", err)
			os.Exit(1)
		}
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Always wipe tables on startup so each test run starts clean
	fmt.Fprintf(os.Stderr, "[ZDEBUG] TestMain: TRUNCATE starting\n")
	for _, table := range []string{
		"gallery_media", "gallery_folders",
		"ppid_requests", "ppid", "ppid_categories",
		"password_reset_tokens", "user_roles", "users",
		"berita", "berita_categories", "umkm", "umkm_categories",
		"fasilitas", "fasilitas_categories", "struktur_organisasi",
		"profile", "profile_categories", "infographic", "infographic_categories", "banners", "banner_categories",
		"backups", "casbin_rule", "settings", "infographic_access_log",
	} {
		if _, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE"); err != nil {
			fmt.Fprintf(os.Stderr, "TRUNCATE %s failed: %v\n", table, err)
		}
	}

	if err := goose.Up(db.DB, "../db/migrations"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	globalTS = buildTestServer(db)
	if pgContainer != nil {
		globalTS.Container = pgContainer
	}

	code := m.Run()

	// Teardown: close server, DB, container
	globalTS.Server.Close()
	globalTS.DB.Close()
	if pgContainer != nil {
		if err := pgContainer.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to terminate container: %v\n", err)
		}
	}

	os.Exit(code)
}

// TestServer wraps the test HTTP server and database
type TestServer struct {
	Server                     *httptest.Server
	DB                         *sqlx.DB
	Container                  *postgres.PostgresContainer
	AuthService                *auth.Service
	UserService                *user.Service
	RoleService                *role.Service
	BannerCategoryService      *bannercategory.Service
	BeritaCategoryService      *beritacategory.Service
	UMKMCategoryService        *umkmcategory.Service
	PPIDCategoryService        *ppidcategory.Service
	FasilitasCategoryService   *fasilitascategory.Service
	GalleryService             *galleryusecase.Service
	InfographicCategoryService *infographiccategory.Service
	Enforcer                   *casbin.Enforcer
}

// SetupTestServer returns the shared global test server after cleaning all tables
// and reloading Casbin policy. This is cheap — no container startup.
func SetupTestServer(t *testing.T) *TestServer {
	t.Helper()
	globalTS.CleanDatabase(t)
	return globalTS
}

// Cleanup is a no-op for the shared server — teardown happens in TestMain.
// It exists so existing `defer ts.Cleanup(t)` calls compile without change.
func (ts *TestServer) Cleanup(_ *testing.T) {}

// buildTestServer initializes the HTTP server with all dependencies.
func buildTestServer(db *sqlx.DB) *TestServer {
	casbinEnforcer, err := rbac.NewCasbinEnforcer("../rbac/rbac_model.conf", db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Casbin enforcer: %v\n", err)
		os.Exit(1)
	}

	casbinEnforcerImpl := casbinEnforcer.(*rbac.CasbinEnforcer)
	rawEnforcer := casbinEnforcerImpl.GetEnforcer()

	// Seed admin wildcard policy so tests that assign the admin role pass RBAC checks.
	if _, err := rawEnforcer.AddPolicy("admin", "*", "*"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to seed admin policy: %v\n", err)
		os.Exit(1)
	}

	// Seed admin user (needed for system folder created_by) so the feature
	// services can resolve their system folders by feature slug.
	const systemFolderOwnerID = "00000000-0000-0000-0000-000000000001"
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO users (id, name, email, hashed_password, created_at, updated_at)
		VALUES ($1, 'System Owner', 'system@desa.local', '$2a$12$system', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, systemFolderOwnerID); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to seed system owner: %v\n", err)
		os.Exit(1)
	}

	clk := clock.RealClock{}
	fileHandler, err := filehandler.NewLocalHandler("./test_uploads")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file handler: %v\n", err)
		os.Exit(1)
	}
	_ = fileHandler // retained for legacy /uploads/* static route scaffolding
	galleryOriginalStorage, err := filehandler.NewLocalHandlerWithSubdir("./test_uploads/private", "gallery/originals")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create gallery original storage: %v\n", err)
		os.Exit(1)
	}
	galleryThumbnailStorage, err := filehandler.NewLocalHandlerWithSubdir("./test_uploads/private", "gallery/thumbnails")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create gallery thumbnail storage: %v\n", err)
		os.Exit(1)
	}

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
	profileCategoryRepo := repopg.NewProfileCategoryRepository(db)
	infographicRepo := repopg.NewInfographicRepository(db)
	infographicCategoryRepo := repopg.NewInfographicCategoryRepository(db)

	jwtSecret := "test-secret-key"
	jwtExpiration := 24 * time.Hour

	authService := auth.NewService(authRepo, userRepo, nil, clk, jwtSecret, jwtExpiration, "http://localhost:8080")
	roleService := role.NewService(casbinEnforcer, roleRepo)
	bannerCategoryService := bannercategory.NewService(bannerCategoryRepo)
	beritaCategoryService := beritacategory.NewService(beritaCategoryRepo, clk)
	umkmCategoryService := umkmcategory.NewService(umkmCategoryRepo, clk)
	ppidCategoryService := ppidcategory.NewService(ppidCategoryRepo, clk)
	fasilitasCategoryService := fasilitascategory.NewService(fasilitasCategoryRepo, clk)
	infographicCategoryService := infographiccategory.NewService(infographicCategoryRepo, clk)

	galleryConfig := config.GalleryConfig{}
	galleryService := galleryusecase.NewService(
		galleryRepo,
		galleryOriginalStorage,
		galleryThumbnailStorage,
		galleryusecase.NewImageProcessor(),
		galleryusecase.NewVideoProcessor(galleryusecase.VideoProcessorConfig{
			FfmpegPath:                 galleryConfig.GetFfmpegPath(),
			FfprobePath:                galleryConfig.GetFfprobePath(),
			ThumbnailMaxWidth:          galleryConfig.GetThumbnailMaxWidth(),
			ThumbnailMaxHeight:         galleryConfig.GetThumbnailMaxHeight(),
			ThumbnailQuality:           galleryConfig.GetThumbnailQuality(),
			VideoThumbnailFrameSeconds: galleryConfig.GetVideoThumbnailFrameSeconds(),
		}),
		galleryConfig,
		clk,
		galleryusecase.NewMediaDeletionHub(),
		// Task 7.1: signed URLs disabled in this CLI/integration path
		nil,
	)
	// gallery deletion listeners (Task 5.1)
	galleryService.Hub().AddListener(umkmRepo)
	galleryService.Hub().AddListener(fasilitasRepo)
	// fileStore exposes gallery.FileStore so every feature (banner, berita,
	// struktur, umkm, fasilitas, user, ppid) routes its uploads into the
	// matching system folder instead of the legacy public directory.
	fileStore := galleryusecase.NewFileStoreService(galleryService)

	// Seed the system folders that back every feature's uploads. The
	// fileStore resolves feature slugs to these rows; without them every
	// upload returns "folder not found: feature <slug>".
	if _, err := galleryService.EnsureSystemFolders(context.Background(), galleryusecase.SystemFolderSpecs(), systemFolderOwnerID); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to seed gallery system folders: %v\n", err)
		os.Exit(1)
	}

	userService := user.NewService(userRepo, fileStore, clk)
	bannerService := banner.NewService(bannerRepo, fileStore, bannerCategoryService, clk)

	beritaService := berita.NewService(beritaRepo, fileStore, clk, "./uploads", beritaCategoryService)
	umkmService := umkm.NewService(umkmRepo, fileStore, clk, umkmCategoryService, &galleryConfig)
	fasilitasService := fasilitas.NewService(fasilitasRepo, fileStore, clk, fasilitasCategoryService, &galleryConfig)
	strukturService := struktur.NewService(strukturRepo, fileStore, clk)
	desaService := desa.NewService(desaRepo, clk)
	profileCategoryService := profilecategory.NewService(profileCategoryRepo, clk)
	profileService := profile.NewService(profileRepo, clk, profileCategoryService)
	infographicAccessLogRepo := repopg.NewInfographicAccessLogRepository(db)
	infographicService := infographic.NewService(infographicRepo, clk, config.MetabaseConfig{
		SecretKey: "test-secret-key",
		URL:       "http://localhost:3000",
	}, infographicCategoryService, infographicAccessLogRepo)

	authHandler := authhandler.NewAuthHandler(authService)
	userHandler := userhandler.NewUserHandler(userService, roleService, nil, false)
	roleHandler := rolehandler.NewRoleHandler(roleService)
	bannerHandler := bannerhandler.NewBannerHandler(bannerService, nil, false)
	bannerCategoryHandler := bannerhandler.NewBannerCategoryHandler(bannerCategoryService)
	beritaHandler := beritahandler.NewBeritaHandler(beritaService, nil, false)
	beritaCategoryHandler := beritacategoryhandler.NewBeritaCategoryHandler(beritaCategoryService)

	logger := &testLogger{}
	beritaUploadHandler := beritahandler.NewBeritaUploadHandler(fileStore, logger)
	umkmHandler := umkmhandler.NewUMKMHandler(umkmService, nil, false)
	umkmCategoryHandler := umkmcategoryhandler.NewUMKMCategoryHandler(umkmCategoryService)
	fasilitasHandler := fasilitashandler.NewFasilitasHandler(fasilitasService, nil, false)
	fasilitasCategoryHandler := fasilitashandler.NewFasilitasCategoryHandler(fasilitasCategoryService)
	galleryHandler := galleryhandler.NewGalleryHandler(galleryService, galleryConfig, nil, false)

	ppidFileHandler, err := file.NewLocalHandlerWithSubdir("./test_uploads", "ppid")
	_ = ppidFileHandler // legacy PPID handler no longer used; gallery stores PPID docs
	_ = err

	// Recreate PPID service with separate file handler. The approval workflow
	// requires an email service; tests use a no-op stub so approving requests
	// never tries to reach a real SMTP server.
	ppidEmailConfig := ppid.EmailConfig{
		VillageName:  "Test Village",
		SupportEmail: "support@test.com",
		WebsiteURL:   "http://localhost:8080",
		DomainAddr:   "http://localhost:8080",
	}
	ppidService := ppid.NewServiceWithEmail(ppidRepo, fileStore, &noopEmailService{}, clk, jwtSecret, ppidEmailConfig, ppidCategoryService)
	ppidHandler := ppidhandler.NewPPIDHandler(ppidService, "./test_uploads/ppid", logger, nil, false)
	ppidCategoryHandler := ppidcategoryhandler.NewPPIDCategoryHandler(ppidCategoryService)
	strukturHandler := strukturhandler.NewStrukturHandler(strukturService, nil, false)
	desaHandler := desahandler.NewDesaHandler(desaService, nil, false)
	desaUploadHandler := desahandler.NewDesaUploadHandler(fileStore, logger)
	profileHandler := profilehandler.NewProfileHandler(profileService)
	profileCategoryHandler := profilehandler.NewProfileCategoryHandler(profileCategoryService)
	infographicHandler := infographichandler.NewInfographicHandler(infographicService)
	infographicCategoryHandler := infographiccategoryhandler.NewInfographicCategoryHandler(infographicCategoryService)
	healthHandler := healthhandler.NewHealthHandler()

	fileHandler2 := filehandlerpkg.NewFileHandler("./test_uploads", logger)

	router := httpserver.NewRouter(httpserver.RouterConfig{
		AuthHandler:                authHandler,
		UserHandler:                userHandler,
		RoleHandler:                roleHandler,
		BannerHandler:              bannerHandler,
		BannerCategoryHandler:      bannerCategoryHandler,
		BeritaHandler:              beritaHandler,
		BeritaCategoryHandler:      beritaCategoryHandler,
		BeritaUploadHandler:        beritaUploadHandler,
		UMKMHandler:                umkmHandler,
		UMKMCategoryHandler:        umkmCategoryHandler,
		FasilitasHandler:           fasilitasHandler,
		FasilitasCategoryHandler:   fasilitasCategoryHandler,
		GalleryHandler:             galleryHandler,
		PPIDHandler:                ppidHandler,
		PPIDCategoryHandler:        ppidCategoryHandler,
		StrukturHandler:            strukturHandler,
		DesaHandler:                desaHandler,
		DesaUploadHandler:          desaUploadHandler,
		ProfileHandler:             profileHandler,
		ProfileCategoryHandler:     profileCategoryHandler,
		InfographicHandler:         infographicHandler,
		InfographicCategoryHandler: infographicCategoryHandler,
		HealthHandler:              healthHandler,
		PublicFileHandler:          fileHandler2,
		UploadPublicDirectory:      "./test_uploads",
		Enforcer:                   rawEnforcer,
		JWTSecret:                  jwtSecret,
		AllowedOrigins:             []string{"*"},
		Logger:                     logger,
		DisableRateLimit:           true,
	})

	server := httptest.NewServer(router)

	return &TestServer{
		Server:                     server,
		DB:                         db,
		AuthService:                authService,
		UserService:                userService,
		RoleService:                roleService,
		BannerCategoryService:      bannerCategoryService,
		BeritaCategoryService:      beritaCategoryService,
		UMKMCategoryService:        umkmCategoryService,
		PPIDCategoryService:        ppidCategoryService,
		FasilitasCategoryService:   fasilitasCategoryService,
		GalleryService:             galleryService,
		InfographicCategoryService: infographicCategoryService,
		Enforcer:                   rawEnforcer,
	}
}

// testLogger is a no-op logger for testing
type testLogger struct{}

func (l *testLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (l *testLogger) Error(_ context.Context, _ string, _ ...any) {}

// CreateTestUser creates a test user and returns their ID
func (ts *TestServer) CreateTestUser(t *testing.T, name, email, password string) string {
	t.Helper()
	input := user.CreateUserInput{
		Name:     name,
		Email:    email,
		Password: password,
	}
	u, err := ts.UserService.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return u.ID
}

// AssignAdminRole assigns the admin role to a user via the role service (Casbin in-memory).
func (ts *TestServer) AssignAdminRole(t *testing.T, email string) {
	t.Helper()
	var userID string
	if err := ts.DB.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID); err != nil {
		t.Fatalf("Failed to find user %s: %v", email, err)
	}
	if err := ts.RoleService.AssignRoleToUser(context.Background(), userID, "admin"); err != nil {
		t.Fatalf("Failed to assign admin role to %s: %v", email, err)
	}
}

// GetAuthToken logs in a user and returns the access token
func (ts *TestServer) GetAuthToken(t *testing.T, email, password string) string {
	t.Helper()
	tokenPair, err := ts.AuthService.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Failed to login test user: %v", err)
	}
	return tokenPair.AccessToken
}

// AdminToken returns the admin's access token, logging in if needed.
func (ts *TestServer) AdminToken(t *testing.T) string {
	t.Helper()
	return ts.GetAuthToken(t, "admin@test.com", "password123")
}

// CreateTestBannerWithCategory creates a banner via the API and returns the banner ID.
// Image is required by the API; we don't include one (the banner won't pass image
// validation) so this helper directly creates a banner via the repository instead.
func (ts *TestServer) CreateTestBannerWithCategory(t *testing.T, title, description, categoryID string) string {
	t.Helper()
	// The banner create endpoint requires multipart with an image file. Easier to
	// insert directly via the repo for test setup.
	imageURL := "/uploads/test-banner.jpg"
	q := `INSERT INTO banners (id, title, description, link, image_url, status, category, metadata, created_at, updated_at)
	       VALUES (gen_random_uuid(), $1, $2, '', $3, 'inactive', $4, '{}', NOW(), NOW())`
	_, err := ts.DB.ExecContext(context.Background(), q, title, description, imageURL, nullIfEmpty(categoryID))
	if err != nil {
		t.Fatalf("create test banner: %v", err)
	}
	// Look up the banner ID we just created
	var id string
	err = ts.DB.QueryRowContext(context.Background(),
		`SELECT id FROM banners WHERE title = $1 ORDER BY created_at DESC LIMIT 1`,
		title,
	).Scan(&id)
	if err != nil {
		t.Fatalf("lookup test banner: %v", err)
	}
	return id
}

// noopEmailService is a no-op PPID EmailService used by integration tests so
// ApproveRequest never attempts to reach an SMTP server.
type noopEmailService struct{}

func (noopEmailService) SendApprovalEmail(_ context.Context, _ ppid.SendApprovalEmailInput) error {
	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// CreateTestBeritaCategory creates a berita category and returns its UUID.
// Used as a helper to set up valid category IDs for berita creation in tests.
func (ts *TestServer) CreateTestBeritaCategory(t *testing.T, name string) string {
	t.Helper()
	c, err := ts.BeritaCategoryService.Create(context.Background(), name)
	if err != nil {
		t.Fatalf("Failed to create test berita category %q: %v", name, err)
	}
	return c.ID
}

// CreateTestUMKMCategory creates a UMKM category and returns its UUID.
func (ts *TestServer) CreateTestUMKMCategory(t *testing.T, name string) string {
	t.Helper()
	c, err := ts.UMKMCategoryService.Create(context.Background(), name)
	if err != nil {
		t.Fatalf("Failed to create test umkm category %q: %v", name, err)
	}
	return c.ID
}

// CreateTestPPIDCategory creates a PPID category and returns its UUID.
func (ts *TestServer) CreateTestPPIDCategory(t *testing.T, name string) string {
	t.Helper()
	c, err := ts.PPIDCategoryService.Create(context.Background(), name)
	if err != nil {
		t.Fatalf("Failed to create test ppid category %q: %v", name, err)
	}
	return c.ID
}

// CreateTestFasilitasCategory creates a Fasilitas category and returns its UUID.
func (ts *TestServer) CreateTestFasilitasCategory(t *testing.T, name string) string {
	t.Helper()
	c, err := ts.FasilitasCategoryService.Create(context.Background(), name)
	if err != nil {
		t.Fatalf("Failed to create test fasilitas category %q: %v", name, err)
	}
	return c.ID
}

// CreateTestInfographicCategory creates an Infographic category and returns its UUID.
func (ts *TestServer) CreateTestInfographicCategory(t *testing.T, name string) string {
	t.Helper()
	c, err := ts.InfographicCategoryService.Create(context.Background(), name)
	if err != nil {
		t.Fatalf("Failed to create test infographic category %q: %v", name, err)
	}
	return c.ID
}

// MakeRequest makes an HTTP request to the test server
func (ts *TestServer) MakeRequest(method, path string, body any, token string) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, ts.Server.URL+path, toJSONReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, ts.Server.URL+path, nil)
		if err != nil {
			return nil, err
		}
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return http.DefaultClient.Do(req)
}

// CleanDatabase truncates all tables and reloads Casbin policy for test isolation.
// The admin wildcard policy is re-seeded after reload.
func (ts *TestServer) CleanDatabase(t *testing.T) {
	t.Helper()

	tables := []string{
		"gallery_media",
		"gallery_folders",
		"password_reset_tokens",
		"user_roles",
		"casbin_rule",
		"users",
		"banners",
		"berita",
		"berita_categories",
		"umkm",
		"umkm_categories",
		"fasilitas",
		"fasilitas_categories",
		"ppid_requests",
		"ppid",
		"ppid_categories",
		"infographic",
		"infographic_categories",
		"struktur_organisasi",
		"settings",
		"infographic_access_log",
	}

	for _, table := range tables {
		if _, err := ts.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
	fmt.Fprintf(os.Stderr, "[ZDEBUG] CleanDatabase: truncated %d tables\n", len(tables))

	// Re-seed the system owner user and the gallery system folders so
	// fileStore can resolve feature slugs after every CleanDatabase.
	if _, err := ts.DB.ExecContext(context.Background(), `
		INSERT INTO users (id, name, email, hashed_password, created_at, updated_at)
		VALUES ('00000000-0000-0000-0000-000000000001', 'System Owner', 'system@desa.local', '$2a$12$system', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		t.Fatalf("Failed to re-seed system owner: %v", err)
	}
	if _, err := ts.GalleryService.EnsureSystemFolders(context.Background(), galleryusecase.SystemFolderSpecs(), "00000000-0000-0000-0000-000000000001"); err != nil {
		t.Fatalf("Failed to re-seed gallery system folders: %v", err)
	}

	// Reload Casbin policy from DB (now empty) to clear in-memory state.
	if err := ts.Enforcer.LoadPolicy(); err != nil {
		t.Fatalf("Failed to reload Casbin policy: %v", err)
	}

	// Re-seed the admin wildcard policy after reload.
	if _, err := ts.Enforcer.AddPolicy("admin", "*", "*"); err != nil {
		t.Fatalf("Failed to re-seed admin policy: %v", err)
	}
}
