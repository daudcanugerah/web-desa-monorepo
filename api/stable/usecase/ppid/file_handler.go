package ppid

import (
	"context"
	"io"
)

// FileHandler defines the file storage interface for PPID document management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/file/local_handler.go, which depends on this interface.
// This maintains the Dependency Rule: interface/file → usecase → domain
type FileHandler interface {
	// SaveDocument saves a document file and returns the relative file path
	// Validates MIME type (PDF, DOC, DOCX, XLS, XLSX) and size (max 50MB)
	SaveDocument(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (string, error)

	// SaveImage saves an image file and returns the relative file path
	// Validates MIME type (JPEG, PNG, WebP) and size (max 10MB)
	SaveImage(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (string, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, path string) error

	// GetFilePath returns the full file path for a given relative path
	GetFilePath(relativePath string) string
}
