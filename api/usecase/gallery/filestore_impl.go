package gallery

import (
	"context"
	"errors"
	"fmt"

	"webdesa/api/domain/gallery"
)

// FileStoreService is the concrete FileStore implementation backed by the
// in-package Service. Feature services depend on this narrow port and
// never touch the rest of the gallery service surface.
type FileStoreService struct {
	svc *Service
}

// NewFileStoreService returns a FileStoreService bound to svc.
func NewFileStoreService(svc *Service) *FileStoreService {
	return &FileStoreService{svc: svc}
}

var _ FileStore = (*FileStoreService)(nil)

func (f *FileStoreService) resolveSystemFolder(ctx context.Context, feature string) (*gallery.Folder, error) {
	slug, err := normalizeFeature(feature)
	if err != nil {
		return nil, err
	}
	folder, err := f.svc.GetSystemFolderByFeature(ctx, slug)
	if err == nil {
		return folder, nil
	}
	if !errors.Is(err, ErrFolderNotFound) {
		return nil, err
	}
	// On-demand bootstrap: system folders must exist before the first
	// upload. Seed data may not have run (fresh env), so upsert the
	// missing folder idempotently, owned by the oldest user.
	spec := systemFolderSpecFor(slug)
	if spec == nil {
		return nil, ErrUnknownFeature
	}
	ownerID, err := f.svc.BootstrapSystemFolderOwner(ctx)
	if err != nil {
		return nil, fmt.Errorf("system folder bootstrap: %w", err)
	}
	if _, err := f.svc.EnsureSystemFolders(ctx, []SystemFolderSpec{*spec}, ownerID); err != nil {
		return nil, fmt.Errorf("system folder bootstrap: %w", err)
	}
	return f.svc.GetSystemFolderByFeature(ctx, slug)
}

func (f *FileStoreService) validateFile(feature, originalName, contentType string, size int64) error {
	slug, err := normalizeFeature(feature)
	if err != nil {
		return err
	}
	if originalName == "" {
		return ErrMediaFilenameEmpty
	}
	if size <= 0 {
		return ErrFileTooLarge
	}
	mt := mediaTypeFromMIME(contentType)
	if mt == "" {
		return ErrInvalidMimeType
	}
	maxBytes := maxBytesFor(f.svc.cfg, mt)
	if maxBytes > 0 && size > maxBytes {
		return ErrFileTooLarge
	}
	_ = slug
	return nil
}

func (f *FileStoreService) SaveImage(ctx context.Context, feature string, in FileInput) (SavedFile, error) {
	if err := f.validateFile(feature, in.OriginalName, in.ContentType, in.Size); err != nil {
		return SavedFile{}, err
	}
	folder, err := f.resolveSystemFolder(ctx, feature)
	if err != nil {
		return SavedFile{}, err
	}
	media, err := f.svc.CreateMedia(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: in.OriginalName,
		MimeType:     in.ContentType,
		Size:         in.Size,
		Content:      in.Content,
		UploadedBy:   folder.CreatedBy,
	})
	if err != nil {
		return SavedFile{}, err
	}
	return SavedFile{MediaID: media.ID}, nil
}

func (f *FileStoreService) SaveDocument(ctx context.Context, feature string, in FileInput) (SavedFile, error) {
	if err := f.validateFile(feature, in.OriginalName, in.ContentType, in.Size); err != nil {
		return SavedFile{}, err
	}
	folder, err := f.resolveSystemFolder(ctx, feature)
	if err != nil {
		return SavedFile{}, err
	}
	media, err := f.svc.CreateDocument(ctx, UploadMediaInput{
		FolderID:     folder.ID,
		OriginalName: in.OriginalName,
		MimeType:     in.ContentType,
		Size:         in.Size,
		Content:      in.Content,
		UploadedBy:   folder.CreatedBy,
	})
	if err != nil {
		return SavedFile{}, err
	}
	return SavedFile{MediaID: media.ID}, nil
}

func (f *FileStoreService) SaveImages(ctx context.Context, feature string, inputs []FileInput) ([]SavedFile, []error) {
	saved := make([]SavedFile, len(inputs))
	errs := make([]error, len(inputs))
	for i, in := range inputs {
		out, err := f.SaveImage(ctx, feature, in)
		saved[i] = out
		errs[i] = err
	}
	hasErr := false
	for _, e := range errs {
		if e != nil {
			hasErr = true
			break
		}
	}
	if !hasErr {
		return saved, nil
	}
	return saved, errs
}

func (f *FileStoreService) ValidateFiles(ctx context.Context, feature string, inputs []FileInput) []error {
	out := make([]error, len(inputs))
	for i, in := range inputs {
		out[i] = f.validateFile(feature, in.OriginalName, in.ContentType, in.Size)
	}
	return out
}

func (f *FileStoreService) Delete(ctx context.Context, mediaID string) error {
	if mediaID == "" {
		return nil
	}
	return f.svc.DeleteMediaForFeature(ctx, mediaID)
}

func (f *FileStoreService) Open(ctx context.Context, mediaID string) (*MediaBinary, error) {
	if mediaID == "" {
		return nil, ErrMediaNotFound
	}
	return f.svc.GetMediaContent(ctx, mediaID)
}

// Ensure FileStoreService satisfies the FileStore interface (compile-time).
var _ FileStore = (*FileStoreService)(nil)

// keep errors imported for SaveImages error wrapping; harmless if unused.
var _ = errors.New
