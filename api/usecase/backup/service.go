package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"braces.dev/errtrace"

	"webdesa/api/domain/backup"
)

// Service provides backup and restore operations for the database and
// the on-disk upload tree.
//
// CreateBackup produces a plain pg_dump custom-format file (db-only,
// back-compat). CreateFullBackup produces a tar.gz archive that bundles
// the pg_dump output plus every path in includePaths (default
// ./uploads/), so a single restore restores both the database and the
// media files.
type Service struct {
	repo         Repository
	backupDir    string
	includePaths []string
	dbDSN        string
	dbHost       string
	dbPort       string
	dbName       string
	dbUser       string
	dbPassword   string
}

// NewService creates a new backup service instance. includePaths is
// optional — pass nil/empty to fall back to the default ("./uploads/").
func NewService(repo Repository, backupDir string, includePaths []string, dbHost, dbPort, dbName, dbUser, dbPassword string) *Service {
	if len(includePaths) == 0 {
		includePaths = []string{"./uploads/"}
	}
	return &Service{
		repo:         repo,
		backupDir:    backupDir,
		includePaths: includePaths,
		dbHost:       dbHost,
		dbPort:       dbPort,
		dbName:       dbName,
		dbUser:       dbUser,
		dbPassword:   dbPassword,
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
		Kind:      "db",
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

// RestoreFromBackup restores the database (and, for full backups, the
// uploads tree) from an existing backup file. The on-disk format is
// auto-detected from the filename: .tar.gz is a full backup (Task 5.2);
// anything else is treated as a db-only pg_dump custom-format file.
// Validates: Requirements 15.1, 15.2
func (s *Service) RestoreFromBackup(ctx context.Context, backupID string) error {
	// Get backup metadata
	b, err := s.repo.FindByID(ctx, backupID)
	if err != nil {
		return errtrace.Wrap(err)
	}

	// Build file path
	archivePath := filepath.Join(s.backupDir, b.Filename)

	// Verify file exists
	if _, err := os.Stat(archivePath); err != nil {
		if os.IsNotExist(err) {
			return errtrace.Wrap(fmt.Errorf("backup file not found: %s", b.Filename))
		}
		return errtrace.Wrap(fmt.Errorf("failed to access backup file: %w", err))
	}

	if strings.HasSuffix(b.Filename, ".tar.gz") {
		return s.restoreFull(ctx, archivePath)
	}
	return s.executeRestore(ctx, archivePath)
}

// restoreFull untars the full backup, runs pg_restore on the embedded
// pg_dump file, and re-extracts the uploads tree back to disk.
func (s *Service) restoreFull(ctx context.Context, archivePath string) error {
	dumpPath, err := os.CreateTemp("", "desa-restore-*.pgdump")
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create temp dump: %w", err))
	}
	defer os.Remove(dumpPath.Name())
	if err := extractFromTarGz(archivePath, "db.pgdump", dumpPath); err != nil {
		dumpPath.Close()
		return errtrace.Wrap(fmt.Errorf("failed to extract pg_dump from archive: %w", err))
	}
	dumpPath.Close()

	if err := s.executeRestore(ctx, dumpPath.Name()); err != nil {
		return errtrace.Wrap(err)
	}

	if err := extractUploadsFromTarGz(archivePath); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to restore uploads tree: %w", err))
	}
	return nil
}

// CreateFullBackup produces a tar.gz archive that bundles a pg_dump
// custom-format output plus every includePath. The restore side
// (RestoreFromBackup) auto-detects the .tar.gz suffix and unpacks both
// halves. The metadata row is stored with kind="full".
// Validates: Task 5.2
func (s *Service) CreateFullBackup(ctx context.Context) (*backup.Backup, error) {
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to create backup directory: %w", err))
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("full_%s.tar.gz", timestamp)
	archivePath := filepath.Join(s.backupDir, filename)

	dumpPath, err := os.CreateTemp("", "desa-full-dump-*.pgdump")
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to create temp dump: %w", err))
	}
	defer os.Remove(dumpPath.Name())

	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", s.dbName,
		"-F", "c",
		"-f", dumpPath.Name(),
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("pg_dump failed: %w, output: %s", err, string(out)))
	}

	archive, err := os.Create(archivePath)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to create archive: %w", err))
	}
	gz := gzip.NewWriter(archive)
	tw := tar.NewWriter(gz)

	if err := addFileToTar(tw, "db.pgdump", dumpPath.Name()); err != nil {
		tw.Close()
		gz.Close()
		archive.Close()
		os.Remove(archivePath)
		return nil, errtrace.Wrap(err)
	}
	dumpPath.Close()

	for _, root := range s.includePaths {
		if err := addDirToTar(tw, root); err != nil {
			tw.Close()
			gz.Close()
			archive.Close()
			os.Remove(archivePath)
			return nil, errtrace.Wrap(fmt.Errorf("failed to add %q to archive: %w", root, err))
		}
	}

	if err := tw.Close(); err != nil {
		gz.Close()
		archive.Close()
		os.Remove(archivePath)
		return nil, errtrace.Wrap(err)
	}
	if err := gz.Close(); err != nil {
		archive.Close()
		os.Remove(archivePath)
		return nil, errtrace.Wrap(err)
	}
	if err := archive.Close(); err != nil {
		os.Remove(archivePath)
		return nil, errtrace.Wrap(err)
	}

	info, err := os.Stat(archivePath)
	if err != nil {
		os.Remove(archivePath)
		return nil, errtrace.Wrap(err)
	}

	b := &backup.Backup{
		Filename:  filename,
		Size:      info.Size(),
		Kind:      "full",
		CreatedAt: time.Now(),
	}
	if err := b.Validate(); err != nil {
		os.Remove(archivePath)
		return nil, errtrace.Wrap(err)
	}
	if err := s.repo.Create(ctx, b); err != nil {
		os.Remove(archivePath)
		return nil, errtrace.Wrap(err)
	}
	return b, nil
}

// addFileToTar streams a single on-disk file into the archive under
// the supplied name (no path prefix).
func addFileToTar(tw *tar.Writer, name, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	hdr := &tar.Header{Name: name, Mode: 0644, Size: info.Size(), ModTime: info.ModTime()}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

// addDirToTar walks root (relative to the working directory) and adds
// every regular file under it, preserving the relative path inside the
// archive. Missing roots are silently skipped so a fresh install that
// hasn't seen any uploads yet doesn't fail the backup.
func addDirToTar(tw *tar.Writer, root string) error {
	root = strings.TrimRight(root, "/")
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return addFileToTar(tw, filepath.Base(root), root)
	}
	return filepath.Walk(root, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fi.IsDir() {
			return nil
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(".", path)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		hdr := &tar.Header{Name: rel, Mode: 0644, Size: fi.Size(), ModTime: fi.ModTime()}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = io.Copy(tw, f)
		return err
	})
}

// extractFromTarGz pulls a single named entry from the archive and
// writes it to w.
func extractFromTarGz(archivePath, name string, w io.Writer) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("entry %q not found in archive", name)
		}
		if err != nil {
			return err
		}
		if hdr.Name != name {
			continue
		}
		_, err = io.Copy(w, tr)
		return err
	}
}

// extractUploadsFromTarGz walks the archive and re-emits every entry
// whose path doesn't start with "db." to its on-disk location relative
// to the working directory. The on-disk tree is preserved exactly so
// media files end up where the running app expects them.
func extractUploadsFromTarGz(archivePath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if strings.HasPrefix(hdr.Name, "db.") {
			continue
		}
		if !strings.HasSuffix(hdr.Name, "/") {
			if err := os.MkdirAll(filepath.Dir(hdr.Name), 0755); err != nil {
				return err
			}
			out, err := os.Create(hdr.Name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
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
		Kind:      "db",
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
