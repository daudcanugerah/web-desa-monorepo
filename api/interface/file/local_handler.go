package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalHandler implements file storage using the local filesystem.
// It implements the FileHandler interfaces from multiple usecase packages:
// - usecase/banner.FileHandler
// - usecase/berita.FileHandler
// - usecase/ppid.FileHandler
// - usecase/struktur.FileHandler
//
// This follows Clean Architecture: interface/file → usecase → domain
type LocalHandler struct {
	uploadDir string
}

// NewLocalHandler creates a new local file storage handler.
// uploadDir is the directory where files will be stored (e.g., "/uploads" or "./uploads")
func NewLocalHandler(uploadDir string) (*LocalHandler, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &LocalHandler{
		uploadDir: uploadDir,
	}, nil
}

// NewLocalHandlerWithSubdir creates a new local file storage handler with a subdirectory.
// uploadDir is the base directory (e.g., "./uploads")
// subdir is the subdirectory for this handler (e.g., "ppid", "public")
func NewLocalHandlerWithSubdir(uploadDir, subdir string) (*LocalHandler, error) {
	// Construct full path with subdirectory
	fullPath := filepath.Join(uploadDir, subdir)

	// Ensure subdirectory exists
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload subdirectory: %w", err)
	}

	return &LocalHandler{
		uploadDir: fullPath,
	}, nil
}

// SaveImage saves an image file and returns the relative file path.
// Validates MIME type (JPEG, PNG, WebP) and size (max 10MB).
//
// Validates: Requirements 19.1, 19.3, 19.5, 19.6, 19.7
func (h *LocalHandler) SaveImage(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (string, error) {
	// Validate file size (max 10MB)
	const maxImageSize = 10 * 1024 * 1024 // 10MB
	if size > maxImageSize {
		return "", fmt.Errorf("image file size exceeds maximum allowed size of 10MB")
	}

	// Validate MIME type
	validImageTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}

	if !validImageTypes[strings.ToLower(contentType)] {
		return "", fmt.Errorf("invalid image type: %s (allowed: JPEG, PNG, WebP)", contentType)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		// Infer extension from content type
		ext = inferExtensionFromContentType(contentType)
	}
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Save file
	return h.saveFile(uniqueFilename, content)
}

// SaveDocument saves a document file and returns the relative file path.
// Validates MIME type (PDF, DOC, DOCX, XLS, XLSX) and size (max 50MB).
//
// Validates: Requirements 19.2, 19.4, 19.5, 19.6, 19.7
func (h *LocalHandler) SaveDocument(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (string, error) {
	// Validate file size (max 50MB)
	const maxDocumentSize = 50 * 1024 * 1024 // 50MB
	if size > maxDocumentSize {
		return "", fmt.Errorf("document file size exceeds maximum allowed size of 50MB")
	}

	// Validate MIME type
	validDocumentTypes := map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	}

	if !validDocumentTypes[strings.ToLower(contentType)] {
		return "", fmt.Errorf("invalid document type: %s (allowed: PDF, DOC, DOCX, XLS, XLSX)", contentType)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		// Infer extension from content type
		ext = inferExtensionFromContentType(contentType)
	}
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Save file
	return h.saveFile(uniqueFilename, content)
}

func (h *LocalHandler) Save(ctx context.Context, name string, content io.Reader, size int64, contentType string) (string, error) {
	if size <= 0 {
		return "", fmt.Errorf("size must be positive")
	}
	safe, err := safeStorageName(name)
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(h.uploadDir, safe)
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	limited := io.LimitReader(content, size+1)
	written, err := io.Copy(file, limited)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file content: %w", err)
	}
	if written != size {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("content size %d does not match declared size %d", written, size)
	}
	return safe, nil
}

func (h *LocalHandler) Path(name string) string {
	return h.GetFilePath(filepath.Base(name))
}

func safeStorageName(name string) (string, error) {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return "", fmt.Errorf("invalid filename")
	}
	return name, nil
}

// Delete removes a file from storage.
//
// Validates: Requirements 19.7
func (h *LocalHandler) Delete(ctx context.Context, path string) error {
	// Construct full file path
	fullPath := filepath.Join(h.uploadDir, path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// File doesn't exist, consider it already deleted
		return nil
	}

	// Remove file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// saveFile is a helper function that saves content to a file and returns the relative path.
func (h *LocalHandler) saveFile(filename string, content io.Reader) (string, error) {
	// Construct full file path
	fullPath := filepath.Join(h.uploadDir, filename)

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content to file
	if _, err := io.Copy(file, content); err != nil {
		// Clean up file on error
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	// Return relative path (just the filename)
	return filename, nil
}

// GetFilePath returns the full file path for a given relative path.
func (h *LocalHandler) GetFilePath(relativePath string) string {
	return filepath.Join(h.uploadDir, relativePath)
}

// inferExtensionFromContentType returns the file extension for a given content type.
func inferExtensionFromContentType(contentType string) string {
	extensions := map[string]string{
		"image/jpeg":         ".jpg",
		"image/jpg":          ".jpg",
		"image/png":          ".png",
		"image/webp":         ".webp",
		"application/pdf":    ".pdf",
		"application/msword": ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.ms-excel": ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
	}

	if ext, ok := extensions[strings.ToLower(contentType)]; ok {
		return ext
	}

	return ""
}
