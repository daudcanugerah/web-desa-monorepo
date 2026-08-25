package cmd

import (
	"context"
	"database/sql"
	"fmt"

	"braces.dev/errtrace"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  "Run database migrations using Goose",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	Long:  "Apply all pending migrations to the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrationsUp()
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Long:  "Rollback the last applied migration from the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrationsDown()
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  "Show the status of all migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return showMigrationStatus()
	},
}

func runMigrationsUp() error {
	if systemConfig == nil {
		return errtrace.Wrap(fmt.Errorf("system config not initialized"))
	}

	ctx := context.Background()

	// Open database connection
	db, err := sql.Open("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to ping database: %w", err))
	}

	mainOtel.Log.Info(ctx, "Connected to database successfully")

	// Set migrations directory
	migrationsDir := "db/migrations"

	// Run migrations
	mainOtel.Log.Infof(ctx, "Running migrations from %s", migrationsDir)
	if err := goose.Up(db, migrationsDir); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to run migrations: %w", err))
	}

	mainOtel.Log.Info(ctx, "Migrations completed successfully")
	return nil
}

func runMigrationsDown() error {
	if systemConfig == nil {
		return errtrace.Wrap(fmt.Errorf("system config not initialized"))
	}

	ctx := context.Background()

	// Open database connection
	db, err := sql.Open("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to ping database: %w", err))
	}

	mainOtel.Log.Info(ctx, "Connected to database successfully")

	// Set migrations directory
	migrationsDir := "db/migrations"

	// Rollback last migration
	mainOtel.Log.Infof(ctx, "Rolling back last migration from %s", migrationsDir)
	if err := goose.Down(db, migrationsDir); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to rollback migration: %w", err))
	}

	mainOtel.Log.Info(ctx, "Migration rollback completed successfully")
	return nil
}

func showMigrationStatus() error {
	if systemConfig == nil {
		return errtrace.Wrap(fmt.Errorf("system config not initialized"))
	}

	ctx := context.Background()

	// Open database connection
	db, err := sql.Open("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to ping database: %w", err))
	}

	mainOtel.Log.Info(ctx, "Connected to database successfully")

	// Set migrations directory
	migrationsDir := "db/migrations"

	// Show migration status
	mainOtel.Log.Infof(ctx, "Checking migration status from %s", migrationsDir)
	if err := goose.Status(db, migrationsDir); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get migration status: %w", err))
	}

	return nil
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	rootCmd.AddCommand(migrateCmd)
}
