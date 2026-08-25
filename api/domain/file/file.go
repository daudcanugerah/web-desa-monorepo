package file

import (
	"io"
)

// File represents a file to be uploaded and stored.
// This is a concrete domain entity used for file operations.
// Content is an io.Reader to support streaming large files efficiently.
//
// Note: File storage interfaces (Handler) are defined in the usecase layer
// where they are USED, following Go best practices. See usecase/banner or
// usecase/user for examples of how to define storage interfaces.
type File struct {
	// Name is the original filename including extension
	Name string

	// Content is the file data as a stream
	Content io.Reader

	// Size is the file size in bytes
	Size int64

	// ContentType is the MIME type of the file (e.g., "image/jpeg", "application/pdf")
	ContentType string
}
