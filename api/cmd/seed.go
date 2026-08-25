package cmd

import (
	"context"
	"fmt"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"webdesa/api/interface/file"
	authpg "webdesa/api/interface/postgres"
	repopg "webdesa/api/interface/postgres"
	"webdesa/api/interface/rbac"
	"webdesa/api/pkg/clock"
	bannercategoryUsecase "webdesa/api/usecase/bannercategory"
	beritacategoryUsecase "webdesa/api/usecase/beritacategory"
	fasilitascategoryUsecase "webdesa/api/usecase/fasilitascategory"
	galleryusecase "webdesa/api/usecase/gallery"
	infographiccategoryUsecase "webdesa/api/usecase/infographiccategory"
	ppidcategoryUsecase "webdesa/api/usecase/ppidcategory"
	roleUsecase "webdesa/api/usecase/role"
	umkmcategoryUsecase "webdesa/api/usecase/umkmcategory"
	userUsecase "webdesa/api/usecase/user"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed initial data",
	Long:  "Seed the database with initial data (admin user, operator user, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSeed(cmd.Context())
	},
}

func runSeed(ctx context.Context) error {
	if systemConfig == nil {
		return errtrace.Wrap(fmt.Errorf("system config not initialized"))
	}

	mainOtel.Log.Info(ctx, "Starting database seeding")

	// Connect to database
	db, err := sqlx.Connect("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	mainOtel.Log.Info(ctx, "Connected to database successfully")

	// Initialize repositories
	userRepo := authpg.NewUserRepository(db)
	roleRepo := repopg.NewRoleRepository(db)
	bannerCategoryRepo := repopg.NewBannerCategoryRepository(db)
	beritaCategoryRepo := repopg.NewBeritaCategoryRepository(db)
	umkmCategoryRepo := repopg.NewUMKMCategoryRepository(db)
	fasilitasCategoryRepo := repopg.NewFasilitasCategoryRepository(db)
	ppidCategoryRepo := repopg.NewPPIDCategoryRepository(db)
	infographicCategoryRepo := repopg.NewInfographicCategoryRepository(db)

	// Initialize clock
	clk := clock.RealClock{}

	// Initialize file handler
	fileHandler, err := file.NewLocalHandler(systemConfig.FileUpload.GetPublicUploadDirectory())
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create file handler: %w", err))
	}
	_ = fileHandler // retained for legacy /uploads/* static route scaffolding

	// Initialize gallery service so we can seed the system folders that
	// back every feature's file uploads.
	galleryRepo := repopg.NewGalleryRepository(db)
	galleryOriginals, err := file.NewLocalHandlerWithSubdir(systemConfig.FileUpload.GetPrivateUploadDirectory(), "gallery/originals")
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create gallery originals handler: %w", err))
	}
	galleryThumbs, err := file.NewLocalHandlerWithSubdir(systemConfig.FileUpload.GetPrivateUploadDirectory(), "gallery/thumbnails")
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create gallery thumbs handler: %w", err))
	}
	galleryService := galleryusecase.NewService(
		galleryRepo,
		galleryOriginals,
		galleryThumbs,
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
	// Task 7.1: signed URLs disabled in this CLI/integration path
	nil,
	)
	fileStore := galleryusecase.NewFileStoreService(galleryService)

	// Initialize user service
	userService := userUsecase.NewService(userRepo, fileStore, clk)

	// Initialize Casbin enforcer for role assignment
	casbinEnforcer, err := rbac.NewCasbinEnforcer("rbac/rbac_model.conf", db)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create Casbin enforcer: %w", err))
	}

	// Initialize role service
	roleService := roleUsecase.NewService(casbinEnforcer, roleRepo)

	// Initialize category services for seeding defaults
	bannerCategoryService := bannercategoryUsecase.NewService(bannerCategoryRepo)
	beritaCategoryService := beritacategoryUsecase.NewService(beritaCategoryRepo, clk)
	umkmCategoryService := umkmcategoryUsecase.NewService(umkmCategoryRepo, clk)
	fasilitasCategoryService := fasilitascategoryUsecase.NewService(fasilitasCategoryRepo, clk)
	ppidCategoryService := ppidcategoryUsecase.NewService(ppidCategoryRepo, clk)
	infographicCategoryService := infographiccategoryUsecase.NewService(infographicCategoryRepo, clk)

	// Seed admin user
	adminUserID, err := seedAdminUser(ctx, userService, userRepo, roleService)
	if err != nil {
		return errtrace.Wrap(err)
	}

	// Seed the gallery system folders (system/banner, system/berita, ...)
	// after the admin user exists — gallery_folders.created_by references
	// users, so the owner must be a real row. Seeding up-front makes a
	// fresh `make db-reset` immediately usable instead of waiting for the
	// first feature upload to lazily create each folder.
	if _, err := galleryService.EnsureSystemFolders(ctx, galleryusecase.SystemFolderSpecs(), adminUserID); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to seed gallery system folders: %w", err))
	}

	// Seed operator user
	if err := seedOperatorUser(ctx, userService, userRepo, roleService); err != nil {
		return errtrace.Wrap(err)
	}

	// Seed role permissions
	if err := seedRolePermissions(ctx, roleService); err != nil {
		return errtrace.Wrap(err)
	}

	// Seed default categories for banner, berita, UMKM, fasilitas, PPID and infographic
	if err := seedDefaultBannerCategories(ctx, bannerCategoryService); err != nil {
		return errtrace.Wrap(err)
	}
	if err := seedDefaultBeritaCategories(ctx, beritaCategoryService); err != nil {
		return errtrace.Wrap(err)
	}
	if err := seedDefaultUMKMCategories(ctx, umkmCategoryService); err != nil {
		return errtrace.Wrap(err)
	}
	if err := seedDefaultFasilitasCategories(ctx, fasilitasCategoryService); err != nil {
		return errtrace.Wrap(err)
	}
	if err := seedDefaultPPIDCategories(ctx, ppidCategoryService); err != nil {
		return errtrace.Wrap(err)
	}
	if err := seedDefaultInfographicCategories(ctx, infographicCategoryService); err != nil {
		return errtrace.Wrap(err)
	}

	mainOtel.Log.Info(ctx, "Database seeding completed successfully")
	return nil
}

func seedAdminUser(ctx context.Context, userService *userUsecase.Service, userRepo userUsecase.Repository, roleService *roleUsecase.Service) (string, error) {
	mainOtel.Log.Info(ctx, "Seeding admin user...")

	// Check if admin user already exists
	existingUser, err := userRepo.FindByEmail(ctx, "admin@desa.local")
	if err == nil && existingUser != nil {
		mainOtel.Log.Info(ctx, "Admin user already exists, skipping creation")
		// Still ensure role is assigned
		if err := roleService.AssignRoleToUser(ctx, existingUser.ID, "admin"); err != nil {
			mainOtel.Log.Infof(ctx, "Warning: failed to assign admin role: %v", err)
		}
		return existingUser.ID, nil
	}

	// Create admin user
	input := userUsecase.CreateUserInput{
		Name:     "Administrator",
		Email:    "admin@desa.local",
		Password: "admin123",
	}

	adminUser, err := userService.Create(ctx, input)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to create admin user: %w", err))
	}

	// Assign admin role via Casbin
	if err := roleService.AssignRoleToUser(ctx, adminUser.ID, "admin"); err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to assign admin role: %w", err))
	}

	mainOtel.Log.Infof(ctx, "✓ Admin user created successfully (ID: %s, Email: %s)", adminUser.ID, adminUser.Email)
	mainOtel.Log.Info(ctx, "  Default credentials - Email: admin@desa.local, Password: admin123")
	return adminUser.ID, nil
}

func seedOperatorUser(ctx context.Context, userService *userUsecase.Service, userRepo userUsecase.Repository, roleService *roleUsecase.Service) error {
	mainOtel.Log.Info(ctx, "Seeding operator user...")

	// Check if operator user already exists
	existingUser, err := userRepo.FindByEmail(ctx, "operator@desa.local")
	if err == nil && existingUser != nil {
		mainOtel.Log.Info(ctx, "Operator user already exists, skipping creation")
		// Still ensure role is assigned
		if err := roleService.AssignRoleToUser(ctx, existingUser.ID, "operator"); err != nil {
			mainOtel.Log.Infof(ctx, "Warning: failed to assign operator role: %v", err)
		}
		return nil
	}

	// Create operator user
	input := userUsecase.CreateUserInput{
		Name:     "Operator",
		Email:    "operator@desa.local",
		Password: "operator123",
	}

	operatorUser, err := userService.Create(ctx, input)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create operator user: %w", err))
	}

	// Assign operator role via Casbin
	if err := roleService.AssignRoleToUser(ctx, operatorUser.ID, "operator"); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to assign operator role: %w", err))
	}

	mainOtel.Log.Infof(ctx, "✓ Operator user created successfully (ID: %s, Email: %s)", operatorUser.ID, operatorUser.Email)
	mainOtel.Log.Info(ctx, "  Default credentials - Email: operator@desa.local, Password: operator123")
	return nil
}

func seedRolePermissions(ctx context.Context, roleService *roleUsecase.Service) error {
	mainOtel.Log.Info(ctx, "Seeding role permissions...")

	// admin gets full access via wildcard
	adminPerms := []roleUsecase.Permission{
		{Resource: "*", Action: "*"},
	}

	// operator gets write access to content resources only
	operatorPerms := []roleUsecase.Permission{
		{Resource: "berita", Action: "read"},
		{Resource: "berita", Action: "write"},
		{Resource: "umkm", Action: "read"},
		{Resource: "umkm", Action: "write"},
		{Resource: "fasilitas", Action: "read"},
		{Resource: "fasilitas", Action: "write"},
		{Resource: "ppid", Action: "read"},
		{Resource: "ppid", Action: "write"},
		{Resource: "struktur", Action: "read"},
		{Resource: "struktur", Action: "write"},
		{Resource: "banner", Action: "read"},
		{Resource: "banner", Action: "write"},
		{Resource: "desa", Action: "read"},
		{Resource: "desa", Action: "write"},
		{Resource: "gallery", Action: "read"},
		{Resource: "gallery", Action: "write"},
	}

	for _, perm := range adminPerms {
		if err := roleService.AddPermissionToRole(ctx, "admin", perm); err != nil {
			mainOtel.Log.Infof(ctx, "Warning: failed to add admin permission %s:%s: %v", perm.Resource, perm.Action, err)
		}
	}

	for _, perm := range operatorPerms {
		if err := roleService.AddPermissionToRole(ctx, "operator", perm); err != nil {
			mainOtel.Log.Infof(ctx, "Warning: failed to add operator permission %s:%s: %v", perm.Resource, perm.Action, err)
		}
	}

	mainOtel.Log.Info(ctx, "✓ Role permissions seeded successfully")
	return nil
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

// defaultBeritaCategories are the default berita categories seeded on first run.
// "Lainnya" is intentionally not in this list — it is inserted by migration 00027
// as the backfill target for empty legacy category values.
var defaultBeritaCategories = []string{
	"Berita Desa",
	"Pengumuman",
	"Kegiatan",
}

// defaultBannerCategories are the default banner categories seeded on first run.
// "Lainnya" is intentionally not in this list — it is inserted by migration 00039
// as the backfill target for banners without a category.
var defaultBannerCategories = []string{
	"Promo",
	"Pengumuman",
	"Event",
}

// defaultUMKMCategories are the default UMKM categories seeded on first run.
// "Umum" is intentionally not in this list — it is inserted by migration 00028
// as the backfill target for empty legacy category values.
var defaultUMKMCategories = []string{
	"Kuliner",
	"Kerajinan",
	"Pertanian",
	"Jasa",
	"Perdagangan",
}

// defaultPPIDCategories are the default PPID categories seeded on first run.
// "Tanpa Kategori" is intentionally not in this list — it is inserted by
// migration 00031 as the default NULL-categorization placeholder.
var defaultPPIDCategories = []string{
	"Anggaran",
	"Peraturan",
	"Laporan",
	"Profil",
	"Keuangan",
}

// defaultFasilitasCategories are the default Fasilitas categories seeded on
// first run. "Lainnya" is inserted by migration 00033 as the backfill fallback.
var defaultFasilitasCategories = []string{
	"Pendidikan",
	"Kesehatan",
	"Ibadah",
	"Olahraga",
	"Pemerintahan",
	"Pasar",
}

// defaultInfographicCategories are the default Infographic categories seeded
// on first run. "Lainnya" is inserted by migration 00035 as the backfill fallback.
var defaultInfographicCategories = []string{
	"Statistik",
	"Keuangan",
	"Kependudukan",
	"Pendidikan",
	"Kesehatan",
}

// seedDefaultBannerCategories inserts the default banner categories,
// skipping any that already exist (idempotent).
func seedDefaultBannerCategories(ctx context.Context, service *bannercategoryUsecase.Service) error {
	mainOtel.Log.Info(ctx, "Seeding default banner categories...")
	for _, name := range defaultBannerCategories {
		if _, err := service.Create(ctx, name); err != nil {
			mainOtel.Log.Infof(ctx, "  banner category %q already exists or failed: %v", name, err)
			continue
		}
		mainOtel.Log.Infof(ctx, "  ✓ banner category %q seeded", name)
	}
	return nil
}

// seedDefaultBeritaCategories inserts the default berita categories,
// skipping any that already exist (idempotent).
func seedDefaultBeritaCategories(ctx context.Context, service *beritacategoryUsecase.Service) error {
	mainOtel.Log.Info(ctx, "Seeding default berita categories...")
	for _, name := range defaultBeritaCategories {
		if _, err := service.Create(ctx, name); err != nil {
			// ON CONFLICT (unique name) is expected; log and continue.
			mainOtel.Log.Infof(ctx, "  berita category %q already exists or failed: %v", name, err)
			continue
		}
		mainOtel.Log.Infof(ctx, "  ✓ berita category %q seeded", name)
	}
	return nil
}

// seedDefaultUMKMCategories inserts the default UMKM categories,
// skipping any that already exist (idempotent).
func seedDefaultUMKMCategories(ctx context.Context, service *umkmcategoryUsecase.Service) error {
	mainOtel.Log.Info(ctx, "Seeding default UMKM categories...")
	for _, name := range defaultUMKMCategories {
		if _, err := service.Create(ctx, name); err != nil {
			mainOtel.Log.Infof(ctx, "  umkm category %q already exists or failed: %v", name, err)
			continue
		}
		mainOtel.Log.Infof(ctx, "  ✓ umkm category %q seeded", name)
	}
	return nil
}

// seedDefaultPPIDCategories inserts the default PPID categories,
// skipping any that already exist (idempotent).
func seedDefaultPPIDCategories(ctx context.Context, service *ppidcategoryUsecase.Service) error {
	mainOtel.Log.Info(ctx, "Seeding default PPID categories...")
	for _, name := range defaultPPIDCategories {
		if _, err := service.Create(ctx, name); err != nil {
			mainOtel.Log.Infof(ctx, "  ppid category %q already exists or failed: %v", name, err)
			continue
		}
		mainOtel.Log.Infof(ctx, "  ✓ ppid category %q seeded", name)
	}
	return nil
}

// seedDefaultFasilitasCategories inserts the default Fasilitas categories,
// skipping any that already exist (idempotent).
func seedDefaultFasilitasCategories(ctx context.Context, service *fasilitascategoryUsecase.Service) error {
	mainOtel.Log.Info(ctx, "Seeding default Fasilitas categories...")
	for _, name := range defaultFasilitasCategories {
		if _, err := service.Create(ctx, name); err != nil {
			mainOtel.Log.Infof(ctx, "  fasilitas category %q already exists or failed: %v", name, err)
			continue
		}
		mainOtel.Log.Infof(ctx, "  ✓ fasilitas category %q seeded", name)
	}
	return nil
}

// seedDefaultInfographicCategories inserts the default Infographic categories,
// skipping any that already exist (idempotent).
func seedDefaultInfographicCategories(ctx context.Context, service *infographiccategoryUsecase.Service) error {
	mainOtel.Log.Info(ctx, "Seeding default Infographic categories...")
	for _, name := range defaultInfographicCategories {
		if _, err := service.Create(ctx, name); err != nil {
			mainOtel.Log.Infof(ctx, "  infographic category %q already exists or failed: %v", name, err)
			continue
		}
		mainOtel.Log.Infof(ctx, "  ✓ infographic category %q seeded", name)
	}
	return nil
}
