package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"braces.dev/errtrace"

	"webdesa/api/domain/backup"
)

// Service provides backup and restore operations for the database.
// This service handles both backup metadata (via Repository) and actual
// backup file operations (pg_dump, pg_restore).
//
// Following Clean Architecture:
// - Accepts Repository interface (defined in this package)
// - Returns concrete Backup structs
// - Dependencies point inward: service → repository interface → domain
type Service struct {
	repo       Repository
	backupDir  string
	dbDSN      string
	dbHost     string
	dbPort     string
	dbName     string
	dbUser     string
	dbPassword string
}

// NewService creates a new backup service instance.
// Parameters:
// - repo: Repository interface for backup metadata persistence
// - backupDir: Directory path where backup files are stored
// - dbDSN: PostgreSQL connection string (for parsing connection details)
// - dbHost, dbPort, dbName, dbUser, dbPassword: Database connection details for pg_dump/pg_restore
func NewService(repo Repository, backupDir, dbHost, dbPort, dbName, dbUser, dbPassword string) *Service {
	return &Service{
		repo:       repo,
		backupDir:  backupDir,
		dbHost:     dbHost,
		dbPort:     dbPort,
		dbName:     dbName,
		dbUser:     dbUser,
		dbPassword: dbPassword,
	}
}

// CreateBackup creates a new database backup using pg_dump.
// Steps:
// 1. Generate unique filename with timestamp
// 2. Execute pg_dump to create backup file
// 3. Get file size
// 4. Create backup metadata record in database
// Returns the created Backup entity
// Validates: Requirements 14.1, 14.2, 14.3
func (s *Service) CreateBackup(ctx context.Context) (*backup.Backup, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to create backup directory: %w", err))
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.sql", timestamp)
	filepath := filepath.Join(s.backupDir, filename)

	// Execute pg_dump
	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", s.dbName,
		"-F", "c", // custom format (compressed)
		"-f", filepath,
	)

	// Set password via environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("pg_dump failed: %w, output: %s", err, string(output)))
	}

	// Get file size
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to stat backup file: %w", err))
	}

	// Create backup entity
	b := &backup.Backup{
		Filename:  filename,
		Size:      fileInfo.Size(),
		CreatedAt: time.Now(),
	}

	// Validate entity
	if err := b.Validate(); err != nil {
		// Clean up file if validation fails
		os.Remove(filepath)
		return nil, errtrace.Wrap(fmt.Errorf("backup validation failed: %w", err))
	}

	// Save metadata to database
	if err := s.repo.Create(ctx, b); err != nil {
		// Clean up file if database insert fails
		os.Remove(filepath)
		return nil, errtrace.Wrap(err)
	}

	return b, nil
}

// ListBackups retrieves paginated list of backup metadata records.
// Returns backup slice, total count, and error
// Validates: Requirements 14.4, 14.5
func (s *Service) ListBackups(ctx context.Context, offset, limit int) ([]*backup.Backup, int, error) {
	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, offset, limit)
}

// RestoreFromBackup restores the database from an existing backup file.
// Steps:
// 1. Retrieve backup metadata from database
// 2. Verify backup file exists
// 3. Execute pg_restore to restore database
// Validates: Requirements 15.1, 15.2
func (s *Service) RestoreFromBackup(ctx context.Context, backupID string) error {
	// Get backup metadata
	b, err := s.repo.FindByID(ctx, backupID)
	if err != nil {
		return errtrace.Wrap(err)
	}

	// Build file path
	filepath := filepath.Join(s.backupDir, b.Filename)

	// Verify file exists
	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			return errtrace.Wrap(fmt.Errorf("backup file not found: %s", b.Filename))
		}
		return errtrace.Wrap(fmt.Errorf("failed to access backup file: %w", err))
	}

	// Execute pg_restore
	return s.executeRestore(ctx, filepath)
}

// RestoreFromFile restores the database from an uploaded backup file.
// Steps:
// 1. Save uploaded file to backup directory
// 2. Execute pg_restore to restore database
// 3. Create backup metadata record
// Validates: Requirements 15.2, 15.3
func (s *Service) RestoreFromFile(ctx context.Context, file io.Reader, filename string) error {
	// Ensure backup directory exists
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create backup directory: %w", err))
	}

	// Generate unique filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	safeFilename := fmt.Sprintf("restore_%s_%s", timestamp, filepath.Base(filename))
	filepath := filepath.Join(s.backupDir, safeFilename)

	// Save uploaded file
	outFile, err := os.Create(filepath)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create backup file: %w", err))
	}
	defer outFile.Close()

	size, err := io.Copy(outFile, file)
	if err != nil {
		os.Remove(filepath)
		return errtrace.Wrap(fmt.Errorf("failed to save backup file: %w", err))
	}

	// Execute pg_restore
	if err := s.executeRestore(ctx, filepath); err != nil {
		return errtrace.Wrap(err)
	}

	// Create backup metadata record
	b := &backup.Backup{
		Filename:  safeFilename,
		Size:      size,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return errtrace.Wrap(err)
	}

	return nil
}

// GetBackupFile retrieves a backup file as a stream for download.
// Returns file reader, file size, and error
// Validates: Requirements 14.4
func (s *Service) GetBackupFile(ctx context.Context, backupID string) (io.ReadCloser, int64, error) {
	// Get backup metadata
	b, err := s.repo.FindByID(ctx, backupID)
	if err != nil {
		return nil, 0, errtrace.Wrap(err)
	}

	// Build file path
	filepath := filepath.Join(s.backupDir, b.Filename)

	// Open file
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, errtrace.Wrap(fmt.Errorf("backup file not found: %s", b.Filename))
		}
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to open backup file: %w", err))
	}

	return file, b.Size, nil
}

// executeRestore executes pg_restore command to restore database from backup file.
// This is a helper method used by both RestoreFromBackup and RestoreFromFile.
func (s *Service) executeRestore(ctx context.Context, filepath string) error {
	// Execute pg_restore
	// --clean: drop database objects before recreating
	// --if-exists: use IF EXISTS when dropping objects
	// -F c: custom format
	cmd := exec.CommandContext(ctx, "pg_restore",
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", s.dbName,
		"--clean",
		"--if-exists",
		"-F", "c",
		filepath,
	)

	// Set password via environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("pg_restore failed: %w, output: %s", err, string(output)))
	}

	return nil
}
