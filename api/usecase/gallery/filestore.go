package gallery

import (
	"context"
	"errors"
	"io"
	"strings"
)

var (
	// ErrSystemFolderImmutable is returned when a caller attempts to mutate
	// a folder that backs a feature (system folders are seeded once and
	// never edited via the API).
	ErrSystemFolderImmutable = errors.New("system folder is immutable")
	// ErrUnknownFeature is returned by FileStore when the supplied feature
	// slug has no system folder row. Run `go run main.go seed` to ensure
	// all seven system folders exist before serving traffic.
	ErrUnknownFeature = errors.New("unknown feature slug")
)

// FileInput describes a single file handed to FileStore.SaveImages or
// FileStore.ValidateFiles.
type FileInput struct {
	OriginalName string
	Content      io.Reader
	Size         int64
	ContentType  string
}

// SavedFile is the canonical return from a successful upload. Only
// MediaID is required by feature services; the URL helper builds the
// public/admin HTTP path from MediaID.
type SavedFile struct {
	MediaID string
}

// FileStore is the narrow port every feature service uses to upload
// files into its dedicated system gallery folder. Implementations live
// in usecase/gallery.
type FileStore interface {
	// SaveImage validates the content type/size against the feature's
	// allowlist, finds or creates the feature's system folder, and stores
	// the file as a gallery media row. Returns the media UUID.
	SaveImage(ctx context.Context, feature string, in FileInput) (SavedFile, error)
	// SaveDocument is the document counterpart (PDF/DOC/etc.) for the
	// ppid feature. Other features must not use it.
	SaveDocument(ctx context.Context, feature string, in FileInput) (SavedFile, error)
	// SaveImages saves a batch of image files in feature order. Each
	// entry that fails is reported via the per-index error; successful
	// entries get a non-empty MediaID.
	SaveImages(ctx context.Context, feature string, inputs []FileInput) ([]SavedFile, []error)
	// ValidateFiles walks every input and applies the per-feature
	// content type/size allowlist without persisting anything.
	ValidateFiles(ctx context.Context, feature string, inputs []FileInput) []error
	// Delete removes the gallery media row plus its files. No-op if the
	// mediaID is empty or unknown. Errors are returned for the caller to
	// log; a missing row is not considered an error.
	Delete(ctx context.Context, mediaID string) error
	// Open returns the stored binary for a media id (file path, content
	// type, original filename, size). Used by features that must stream
	// bytes directly (e.g. ppid document download) instead of redirecting
	// to a gallery endpoint.
	Open(ctx context.Context, mediaID string) (*MediaBinary, error)
}

// normalizeFeature canonicalises the supplied feature slug.
func normalizeFeature(slug string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(slug))
	switch s {
	case FeatureBanner, FeatureBerita, FeatureStruktur, FeatureUMKM,
		FeatureFasilitas, FeatureUser, FeaturePPID:
		return s, nil
	}
	return "", ErrUnknownFeature
}
