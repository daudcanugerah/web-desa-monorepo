package fasilitas

import (
	"context"
	"io"
)

// FileHandler defines the interface for file operations in the fasilitas usecase.
// The implementation will be in interface/file/local_handler.go, which depends on this interface.
// This maintains the Dependency Rule: interface/file → usecase → domain
type FileHandler interface {
	// SaveImage saves an image file and returns the relative file path
	// Validates MIME type (JPEG, PNG, WebP) and size (max 10MB)
	SaveImage(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (string, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, filePath string) error
}
