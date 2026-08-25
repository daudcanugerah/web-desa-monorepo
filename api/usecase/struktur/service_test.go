package struktur

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webdesa/api/domain/struktur"
	"webdesa/api/pkg/clock"
	galleryUsecase "webdesa/api/usecase/gallery"
)

type stubRepo struct {
	mu    sync.Mutex
	rows  map[string]*struktur.Struktur
	order []string
}

func newStubRepo() *stubRepo {
	return &stubRepo{rows: map[string]*struktur.Struktur{}}
}

func (r *stubRepo) Create(ctx context.Context, s *struktur.Struktur) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[s.ID] = s
	r.order = append(r.order, s.ID)
	return nil
}

func (r *stubRepo) FindByID(ctx context.Context, id string) (*struktur.Struktur, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.rows[id]
	if !ok {
		return nil, errors.New("struktur not found")
	}
	return s, nil
}

func (r *stubRepo) List(ctx context.Context, query *string, offset, limit int) ([]*struktur.Struktur, int, error) {
	return nil, 0, nil
}

func (r *stubRepo) Update(ctx context.Context, s *struktur.Struktur) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[s.ID] = s
	return nil
}

func (r *stubRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rows, id)
	return nil
}

type stubFileStore struct {
	mu           sync.Mutex
	saved        []string
	deleted      []string
	known        map[string]bool
	saveImageErr error
}

func newStubFileStore() *stubFileStore {
	return &stubFileStore{known: map[string]bool{}}
}

func (f *stubFileStore) SaveImage(ctx context.Context, feature string, in galleryUsecase.FileInput) (galleryUsecase.SavedFile, error) {
	if f.saveImageErr != nil {
		return galleryUsecase.SavedFile{}, f.saveImageErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	id := uuid.NewString()
	f.saved = append(f.saved, id)
	f.known[id] = true
	return galleryUsecase.SavedFile{MediaID: id}, nil
}

func (f *stubFileStore) SaveDocument(ctx context.Context, feature string, in galleryUsecase.FileInput) (galleryUsecase.SavedFile, error) {
	return galleryUsecase.SavedFile{}, errors.New("not implemented")
}

func (f *stubFileStore) SaveImages(ctx context.Context, feature string, inputs []galleryUsecase.FileInput) ([]galleryUsecase.SavedFile, []error) {
	return nil, nil
}

func (f *stubFileStore) ValidateFiles(ctx context.Context, feature string, inputs []galleryUsecase.FileInput) []error {
	return nil
}

func (f *stubFileStore) Delete(ctx context.Context, mediaID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, mediaID)
	delete(f.known, mediaID)
	return nil
}

func (f *stubFileStore) Open(ctx context.Context, mediaID string) (*galleryUsecase.MediaBinary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.known[mediaID] {
		return nil, errors.New("media not found")
	}
	return &galleryUsecase.MediaBinary{ContentType: "image/jpeg", Size: 10}, nil
}

func TestCreateWithMediaRef(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := NewService(repo, fs, clock.NewFixedClock(time.Now()))
	ctx := context.Background()

	// pre-uploaded media id is referenced directly, nothing saved
	fs.known["00000000-0000-0000-0000-000000000001"] = true
	st, err := svc.Create(ctx, CreateStrukturInput{
		Name:                "Budi",
		ProfileImageMediaID: strptr("00000000-0000-0000-0000-000000000001"),
	})
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", *st.ProfileImageMediaID)
	assert.Empty(t, fs.saved)

	// unknown media id -> error, nothing persisted
	_, err = svc.Create(ctx, CreateStrukturInput{
		Name:                "Susi",
		ProfileImageMediaID: strptr("00000000-0000-0000-0000-000000000002"),
	})
	require.Error(t, err)
	assert.Len(t, repo.order, 1)

	// both file + media ref -> conflict
	_, err = svc.Create(ctx, CreateStrukturInput{
		Name:                "Agus",
		ProfileImageMediaID: strptr("00000000-0000-0000-0000-000000000001"),
		ImageFile:           bytes.NewReader([]byte("x")),
	})
	require.Error(t, err)
	assert.Len(t, repo.order, 1)

	// inline file still works (legacy path)
	st, err = svc.Create(ctx, CreateStrukturInput{
		Name:      "Cici",
		ImageFile: bytes.NewReader([]byte("img")),
		ImageName: "p.jpg",
		ImageSize: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, st.ProfileImageMediaID)
	assert.NotEmpty(t, *st.ProfileImageMediaID)
}

func TestUpdateWithMediaRef(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := NewService(repo, fs, clock.NewFixedClock(time.Now()))
	ctx := context.Background()

	old := &struktur.Struktur{
		ID:                  uuid.NewString(),
		Name:                "Budi",
		ProfileImageMediaID: strptr("00000000-0000-0000-0000-00000000000a"),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	require.NoError(t, repo.Create(ctx, old))
	fs.known[*old.ProfileImageMediaID] = true

	// replace via pre-uploaded ref -> old media deleted, no new save
	fs.known["00000000-0000-0000-0000-00000000000b"] = true
	st, err := svc.Update(ctx, old.ID, UpdateStrukturInput{
		Name:                "Budi Baru",
		ProfileImageMediaID: strptr("00000000-0000-0000-0000-00000000000b"),
	})
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-00000000000b", *st.ProfileImageMediaID)
	assert.Contains(t, fs.deleted, "00000000-0000-0000-0000-00000000000a")
	assert.Empty(t, fs.saved)

	// unknown ref -> error, old media untouched
	_, err = svc.Update(ctx, old.ID, UpdateStrukturInput{
		Name:                "Budi Lagi",
		ProfileImageMediaID: strptr("00000000-0000-0000-0000-000000000099"),
	})
	require.Error(t, err)
	cur, _ := repo.FindByID(ctx, old.ID)
	assert.Equal(t, "00000000-0000-0000-0000-00000000000b", *cur.ProfileImageMediaID)

	// inline file still works (legacy path) and replaces
	st, err = svc.Update(ctx, old.ID, UpdateStrukturInput{
		Name:      "Budi File",
		ImageFile: io.NopCloser(bytes.NewReader([]byte("img"))),
		ImageName: "p.jpg",
		ImageSize: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, st.ProfileImageMediaID)
	assert.NotEqual(t, "00000000-0000-0000-0000-00000000000b", *st.ProfileImageMediaID)
}

func strptr(s string) *string { return &s }
