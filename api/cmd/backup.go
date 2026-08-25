package cmd

import (
	"fmt"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	repopg "webdesa/api/interface/postgres"
	"webdesa/api/usecase/backup"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage database backups",
	Long:  "Create, list, and restore database backups",
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new database backup",
	Long:  "Create a new database backup using pg_dump. Pass --with-files to bundle the uploads/ tree (Task 5.2).",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBackupCreate()
	},
}

var backupWithFiles bool

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all database backups",
	Long:  "List all available database backups with metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBackupList()
	},
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore <backup-id>",
	Short: "Restore database from backup",
	Long:  "Restore database from an existing backup using pg_restore",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]
		return runBackupRestore(backupID)
	},
}

func runBackupCreate() error {
	if systemConfig == nil {
		return errtrace.Wrap(fmt.Errorf("system config not initialized"))
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	mainOtel.Log.Info(ctx, "Connected to database successfully")

	// Initialize backup repository
	backupRepo := repopg.NewBackupRepository(db)

	// Parse DSN to extract connection details
	dbHost, dbPort, dbName, dbUser, dbPassword := parseDSN(systemConfig.Postgres.DSN)

	// Initialize backup service
	backupService := backup.NewService(
		backupRepo,
		systemConfig.Backup.GetDirectory(),
		systemConfig.Backup.GetIncludePaths(),
		dbHost,
		dbPort,
		dbName,
		dbUser,
		dbPassword,
	)

	// Create backup
	mainOtel.Log.Info(ctx, "Creating database backup...")
	if backupWithFiles {
		mainOtel.Log.Info(ctx, "--with-files enabled: bundling uploads/ tree")
		b, err := backupService.CreateFullBackup(ctx)
		if err != nil {
			return errtrace.Wrap(fmt.Errorf("failed to create full backup: %w", err))
		}
		mainOtel.Log.Infof(ctx, "Full backup created: %s (ID: %s, Size: %d bytes)", b.Filename, b.ID, b.Size)
		return nil
	}
	b, err := backupService.CreateBackup(ctx)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create backup: %w", err))
	}

	mainOtel.Log.Infof(ctx, "Backup created successfully: %s (ID: %s, Size: %d bytes)", b.Filename, b.ID, b.Size)
	return nil
}

func runBackupList() error {
	if systemConfig == nil {
		return errtrace.Wrap(fmt.Errorf("system config not initialized"))
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	mainOtel.Log.Info(ctx, "Connected to database successfully")

	// Initialize backup repository
	backupRepo := repopg.NewBackupRepository(db)

	// Parse DSN to extract connection details
	dbHost, dbPort, dbName, dbUser, dbPassword := parseDSN(systemConfig.Postgres.DSN)

	// Initialize backup service
	backupService := backup.NewService(
		backupRepo,
		systemConfig.Backup.GetDirectory(),
		systemConfig.Backup.GetIncludePaths(),
		dbHost,
		dbPort,
		dbName,
		dbUser,
		dbPassword,
	)

	// List backups (get all, no pagination for CLI)
	mainOtel.Log.Info(ctx, "Fetching backup list...")
	backups, total, err := backupService.ListBackups(ctx, 0, 100)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to list backups: %w", err))
	}

	if total == 0 {
		mainOtel.Log.Info(ctx, "No backups found")
		return nil
	}

	mainOtel.Log.Infof(ctx, "Found %d backup(s):", total)
	fmt.Println("\nID\t\t\t\t\tFilename\t\t\tSize\t\tCreated At")
	fmt.Println("------------------------------------------------------------------------------------")
	for _, b := range backups {
		fmt.Printf("%s\t%s\t%d bytes\t%s\n", b.ID, b.Filename, b.Size, b.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func runBackupRestore(backupID string) error {
	if systemConfig == nil {
		return errtrace.Wrap(fmt.Errorf("system config not initialized"))
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	mainOtel.Log.Info(ctx, "Connected to database successfully")

	// Initialize backup repository
	backupRepo := repopg.NewBackupRepository(db)

	// Parse DSN to extract connection details
	dbHost, dbPort, dbName, dbUser, dbPassword := parseDSN(systemConfig.Postgres.DSN)

	// Initialize backup service
	backupService := backup.NewService(
		backupRepo,
		systemConfig.Backup.GetDirectory(),
		systemConfig.Backup.GetIncludePaths(),
		dbHost,
		dbPort,
		dbName,
		dbUser,
		dbPassword,
	)

	// Restore backup
	mainOtel.Log.Infof(ctx, "Restoring database from backup ID: %s", backupID)
	mainOtel.Log.Info(ctx, "WARNING: This will drop and recreate all database objects")

	if err := backupService.RestoreFromBackup(ctx, backupID); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to restore backup: %w", err))
	}

	mainOtel.Log.Info(ctx, "Database restored successfully")
	return nil
}

func init() {
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCreateCmd.Flags().BoolVar(&backupWithFiles, "with-files", false, "Bundle uploads/ tree into the backup archive (Task 5.2)")
	rootCmd.AddCommand(backupCmd)
}
