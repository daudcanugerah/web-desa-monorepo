package ppid

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webdesa/api/domain/ppid"
	"webdesa/api/pkg/clock"
	galleryUsecase "webdesa/api/usecase/gallery"
)

// trackingFileStore records saves/deletes and only accepts media ids that
// were either uploaded through it or pre-seeded as "known".
type trackingFileStore struct {
	mu      sync.Mutex
	known   map[string]bool
	saved   []string
	deleted []string
}

func newTrackingFileStore() *trackingFileStore {
	return &trackingFileStore{known: map[string]bool{}}
}

func (f *trackingFileStore) SaveImage(ctx context.Context, feature string, in galleryUsecase.FileInput) (galleryUsecase.SavedFile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := uuid.NewString()
	f.saved = append(f.saved, id)
	f.known[id] = true
	return galleryUsecase.SavedFile{MediaID: id}, nil
}

func (f *trackingFileStore) SaveDocument(ctx context.Context, feature string, in galleryUsecase.FileInput) (galleryUsecase.SavedFile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := uuid.NewString()
	f.saved = append(f.saved, id)
	f.known[id] = true
	return galleryUsecase.SavedFile{MediaID: id}, nil
}

func (f *trackingFileStore) SaveImages(ctx context.Context, feature string, inputs []galleryUsecase.FileInput) ([]galleryUsecase.SavedFile, []error) {
	return nil, nil
}

func (f *trackingFileStore) ValidateFiles(ctx context.Context, feature string, inputs []galleryUsecase.FileInput) []error {
	return nil
}

func (f *trackingFileStore) Delete(ctx context.Context, mediaID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, mediaID)
	delete(f.known, mediaID)
	return nil
}

func (f *trackingFileStore) Open(ctx context.Context, mediaID string) (*galleryUsecase.MediaBinary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.known[mediaID] {
		return nil, errors.New("media not found")
	}
	return &galleryUsecase.MediaBinary{ContentType: "application/pdf", Size: 10}, nil
}

func newTestService(t *testing.T, repo Repository, fs galleryUsecase.FileStore) *Service {
	t.Helper()
	return NewService(repo, fs, clock.NewFixedClock(time.Now()), "test-secret", &mockCategoryLookup{})
}

func TestCreateWithMediaRefs(t *testing.T) {
	repo := newMockRepository()
	fs := newTrackingFileStore()
	svc := newTestService(t, repo, fs)
	ctx := context.Background()

	// pre-uploaded doc + thumb are referenced directly, nothing saved
	fs.known["00000000-0000-0000-0000-000000000001"] = true
	fs.known["00000000-0000-0000-0000-000000000002"] = true
	p, err := svc.Create(ctx, CreatePPIDInput{
		Title:            "Dokumen APBDes",
		DocumentMediaID:  strptr("00000000-0000-0000-0000-000000000001"),
		ThumbnailMediaID: strptr("00000000-0000-0000-0000-000000000002"),
	})
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", *p.DocumentMediaID)
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", *p.ThumbnailMediaID)
	assert.Empty(t, fs.saved)
	assert.Empty(t, fs.deleted)

	// document is required
	_, err = svc.Create(ctx, CreatePPIDInput{Title: "Tanpa Dokumen"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "document file or media id is required")

	// unknown media id -> error, nothing persisted
	_, err = svc.Create(ctx, CreatePPIDInput{
		Title:           "Ref Salah",
		DocumentMediaID: strptr("00000000-0000-0000-0000-000000000099"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid media id")
	assert.Len(t, repo.ppidByID, 1)

	// both file + media ref -> conflict
	_, err = svc.Create(ctx, CreatePPIDInput{
		Title:           "Konflik",
		DocumentMediaID: strptr("00000000-0000-0000-0000-000000000001"),
		DocumentFile:    bytes.NewReader([]byte("x")),
	})
	require.Error(t, err)
	assert.Len(t, repo.ppidByID, 1)

	// inline file path still works
	p, err = svc.Create(ctx, CreatePPIDInput{
		Title:        "Dokumen Inline",
		DocumentFile: bytes.NewReader([]byte("pdf-bytes")),
		DocumentName: "doc.pdf",
		DocumentSize: 9,
		ContentType:  "application/pdf",
	})
	require.NoError(t, err)
	require.NotNil(t, p.DocumentMediaID)
	assert.Len(t, fs.saved, 1)
}

func TestUpdateWithMediaRefs(t *testing.T) {
	repo := newMockRepository()
	fs := newTrackingFileStore()
	svc := newTestService(t, repo, fs)
	ctx := context.Background()

	old := &ppid.PPID{
		ID:               uuid.NewString(),
		Title:            "Lama",
		DocumentMediaID:  strptr("00000000-0000-0000-0000-00000000000a"),
		ThumbnailMediaID: strptr("00000000-0000-0000-0000-00000000000b"),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	require.NoError(t, repo.Create(ctx, old))
	fs.known["00000000-0000-0000-0000-00000000000a"] = true
	fs.known["00000000-0000-0000-0000-00000000000b"] = true

	// replace both via pre-uploaded refs -> old media deleted, no new saves
	fs.known["00000000-0000-0000-0000-00000000000c"] = true
	fs.known["00000000-0000-0000-0000-00000000000d"] = true
	p, err := svc.Update(ctx, old.ID, UpdatePPIDInput{
		Title:            "Baru",
		DocumentMediaID:  strptr("00000000-0000-0000-0000-00000000000c"),
		ThumbnailMediaID: strptr("00000000-0000-0000-0000-00000000000d"),
	})
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-00000000000c", *p.DocumentMediaID)
	assert.Equal(t, "00000000-0000-0000-0000-00000000000d", *p.ThumbnailMediaID)
	assert.Contains(t, fs.deleted, "00000000-0000-0000-0000-00000000000a")
	assert.Contains(t, fs.deleted, "00000000-0000-0000-0000-00000000000b")
	assert.Empty(t, fs.saved)

	// unknown ref -> error, nothing changed
	_, err = svc.Update(ctx, old.ID, UpdatePPIDInput{
		Title:           "Gagal",
		DocumentMediaID: strptr("00000000-0000-0000-0000-000000000099"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid media id")
	cur, err := repo.FindByID(ctx, old.ID)
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-00000000000c", *cur.DocumentMediaID)

	// both file + media ref -> conflict
	_, err = svc.Update(ctx, old.ID, UpdatePPIDInput{
		Title:           "Konflik",
		DocumentMediaID: strptr("00000000-0000-0000-0000-00000000000c"),
		DocumentFile:    bytes.NewReader([]byte("x")),
	})
	require.Error(t, err)

	// inline file path still works and replaces
	p, err = svc.Update(ctx, old.ID, UpdatePPIDInput{
		Title:        "File Baru",
		DocumentFile: bytes.NewReader([]byte("pdf-bytes")),
		DocumentName: "doc.pdf",
		DocumentSize: 9,
		ContentType:  "application/pdf",
	})
	require.NoError(t, err)
	require.NotNil(t, p.DocumentMediaID)
	assert.NotEqual(t, "00000000-0000-0000-0000-00000000000c", *p.DocumentMediaID)
	assert.Contains(t, fs.deleted, "00000000-0000-0000-0000-00000000000c")
}

func TestCreateMediaRefCleanupOnThumbnailFailure(t *testing.T) {
	repo := newMockRepository()
	fs := newTrackingFileStore()
	svc := newTestService(t, repo, fs)
	ctx := context.Background()

	// document uploaded inline, thumbnail ref invalid -> created doc cleaned up
	_, err := svc.Create(ctx, CreatePPIDInput{
		Title:            "Bersih",
		DocumentFile:     bytes.NewReader([]byte("pdf-bytes")),
		DocumentName:     "doc.pdf",
		DocumentSize:     9,
		ContentType:      "application/pdf",
		ThumbnailMediaID: strptr("00000000-0000-0000-0000-0000000000ff"),
	})
	require.Error(t, err)
	assert.Len(t, fs.saved, 1)
	assert.Len(t, fs.deleted, 1)
	assert.Empty(t, repo.ppidByID)
}

func strptr(s string) *string { return &s }
