package fasilitas

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

	"webdesa/api/domain/fasilitas"
	"webdesa/api/domain/fasilitascategory"
	"webdesa/api/pkg/clock"
	galleryUsecase "webdesa/api/usecase/gallery"
)

type stubRepo struct {
	mu    sync.Mutex
	rows  map[string]*fasilitas.Fasilitas
	order []string
}

func newStubRepo() *stubRepo {
	return &stubRepo{rows: map[string]*fasilitas.Fasilitas{}}
}

func (r *stubRepo) Create(ctx context.Context, f *fasilitas.Fasilitas) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[f.ID] = f
	r.order = append(r.order, f.ID)
	return nil
}

func (r *stubRepo) FindByID(ctx context.Context, id string) (*fasilitas.Fasilitas, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.rows[id]
	if !ok {
		return nil, errors.New("fasilitas not found")
	}
	return f, nil
}

func (r *stubRepo) List(ctx context.Context, bbox *BoundingBox, query *string, category *string, offset, limit int) ([]*fasilitas.Fasilitas, int, error) {
	return nil, 0, nil
}

func (r *stubRepo) Update(ctx context.Context, f *fasilitas.Fasilitas) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[f.ID] = f
	return nil
}

func (r *stubRepo) OnMediaDeleted(ctx context.Context, mediaID string) {}

func (r *stubRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rows, id)
	return nil
}

type stubFileStore struct {
	mu      sync.Mutex
	known   map[string]bool
	saved   []string
	deleted []string
}

func newStubFileStore() *stubFileStore {
	return &stubFileStore{known: map[string]bool{}}
}

func (f *stubFileStore) SaveImage(ctx context.Context, feature string, in galleryUsecase.FileInput) (galleryUsecase.SavedFile, error) {
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
		return nil, galleryUsecase.ErrMediaNotFound
	}
	return &galleryUsecase.MediaBinary{ContentType: "image/jpeg", Size: 10}, nil
}

type stubCategoryLookup struct{}

func (m *stubCategoryLookup) FindByID(ctx context.Context, id string) (*fasilitascategory.Category, error) {
	return &fasilitascategory.Category{ID: id, Name: "mock"}, nil
}

type stubBulkLimits struct {
	maxFiles   int
	maxTotalMB int
}

func (l *stubBulkLimits) GetBulkUploadMaxFiles() int {
	if l.maxFiles <= 0 {
		return 20
	}
	return l.maxFiles
}

func (l *stubBulkLimits) GetBulkUploadMaxTotalMB() int {
	if l.maxTotalMB <= 0 {
		return 250
	}
	return l.maxTotalMB
}

func newTestService(repo *stubRepo, fs *stubFileStore) *Service {
	return NewService(repo, fs, clock.NewFixedClock(time.Now()), &stubCategoryLookup{}, &stubBulkLimits{})
}

func TestCreateWithMediaRefs(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := newTestService(repo, fs)
	ctx := context.Background()

	fs.known["00000000-0000-0000-0000-000000000001"] = true
	fs.known["00000000-0000-0000-0000-000000000002"] = true
	f, err := svc.Create(ctx, CreateFasilitasInput{
		Name:           "Balai Desa",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"}, f.ImagesMediaIDs)
	assert.Empty(t, fs.saved)
	assert.Empty(t, fs.deleted)

	// unknown ref -> error, nothing persisted
	_, err = svc.Create(ctx, CreateFasilitasInput{
		Name:           "Gagal",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-000000000099"},
	})
	require.Error(t, err)
	assert.Len(t, repo.order, 1)

	// refs + inline files merge (refs first)
	fs.known["00000000-0000-0000-0000-000000000003"] = true
	f, err = svc.Create(ctx, CreateFasilitasInput{
		Name:           "Campuran",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-000000000003"},
		ImageFiles: []ImageInput{
			{Filename: "a.jpg", Content: bytes.NewReader([]byte("a")), Size: 1, ContentType: "image/jpeg"},
		},
	})
	require.NoError(t, err)
	require.Len(t, f.ImagesMediaIDs, 2)
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", f.ImagesMediaIDs[0])
	assert.Len(t, fs.saved, 1)
}

func TestUpdateWithMediaRefs(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := newTestService(repo, fs)
	ctx := context.Background()

	old := &fasilitas.Fasilitas{
		ID:             uuid.NewString(),
		Name:           "Lama",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-00000000000a"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, repo.Create(ctx, old))
	fs.known["00000000-0000-0000-0000-00000000000a"] = true

	// replace via refs -> old media deleted, no new saves
	fs.known["00000000-0000-0000-0000-00000000000b"] = true
	f, err := svc.Update(ctx, old.ID, UpdateFasilitasInput{
		Name:           "Baru",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-00000000000b"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"00000000-0000-0000-0000-00000000000b"}, f.ImagesMediaIDs)
	assert.Contains(t, fs.deleted, "00000000-0000-0000-0000-00000000000a")
	assert.Empty(t, fs.saved)

	// unknown ref -> error, nothing changed
	_, err = svc.Update(ctx, old.ID, UpdateFasilitasInput{
		Name:           "Gagal",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-000000000099"},
	})
	require.Error(t, err)
	cur, _ := repo.FindByID(ctx, old.ID)
	assert.Equal(t, []string{"00000000-0000-0000-0000-00000000000b"}, cur.ImagesMediaIDs)
}

func TestCreateBulkCaps(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := newTestService(repo, fs)
	ctx := context.Background()

	// 21st image -> ErrBulkTooManyFiles, nothing saved
	files := make([]ImageInput, 21)
	for i := range files {
		files[i] = ImageInput{Filename: "a.jpg", Content: bytes.NewReader([]byte("x")), Size: 1024, ContentType: "image/jpeg"}
	}
	_, err := svc.Create(ctx, CreateFasilitasInput{
		Name:       "Banjir",
		Latitude:   -6.2,
		Longitude:  106.8,
		ImageFiles: files,
	})
	require.ErrorIs(t, err, galleryUsecase.ErrBulkTooManyFiles)
	assert.Empty(t, fs.saved)
	assert.Empty(t, fs.deleted)
	assert.Len(t, repo.order, 0)

	// refs count toward the cap: 20 refs + 1 file -> too many
	fs.known["00000000-0000-0000-0000-000000000001"] = true
	refs := make([]string, 20)
	for i := range refs {
		refs[i] = "00000000-0000-0000-0000-000000000001"
	}
	_, err = svc.Create(ctx, CreateFasilitasInput{
		Name:           "Banjir",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: refs,
		ImageFiles:     files[:1],
	})
	require.ErrorIs(t, err, galleryUsecase.ErrBulkTooManyFiles)
	assert.Empty(t, fs.saved)

	// 20 files within the count cap but >250MB total -> ErrBulkTotalTooLarge
	big := make([]ImageInput, 20)
	for i := range big {
		big[i] = ImageInput{Filename: "b.jpg", Content: bytes.NewReader([]byte("x")), Size: 13 * 1024 * 1024, ContentType: "image/jpeg"}
	}
	_, err = svc.Create(ctx, CreateFasilitasInput{
		Name:       "Banjir",
		Latitude:   -6.2,
		Longitude:  106.8,
		ImageFiles: big,
	})
	require.ErrorIs(t, err, galleryUsecase.ErrBulkTotalTooLarge)
	assert.Empty(t, fs.saved)
}

func TestUpdateBulkCaps(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := newTestService(repo, fs)
	ctx := context.Background()

	old := &fasilitas.Fasilitas{
		ID:             uuid.NewString(),
		Name:           "Lama",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-00000000000a"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, repo.Create(ctx, old))

	files := make([]ImageInput, 21)
	for i := range files {
		files[i] = ImageInput{Filename: "a.jpg", Content: bytes.NewReader([]byte("x")), Size: 1024, ContentType: "image/jpeg"}
	}
	_, err := svc.Update(ctx, old.ID, UpdateFasilitasInput{
		Name:       "Banjir",
		Latitude:   -6.2,
		Longitude:  106.8,
		ImageFiles: files,
	})
	require.ErrorIs(t, err, galleryUsecase.ErrBulkTooManyFiles)
	assert.Empty(t, fs.saved)
	cur, _ := repo.FindByID(ctx, old.ID)
	assert.Equal(t, []string{"00000000-0000-0000-0000-00000000000a"}, cur.ImagesMediaIDs)
}

func TestGetByIDPrunesDanglingMediaIDs(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := newTestService(repo, fs)
	ctx := context.Background()

	fs.known["00000000-0000-0000-0000-00000000000a"] = true
	f := &fasilitas.Fasilitas{
		ID:             uuid.NewString(),
		Name:           "Mixed",
		Latitude:       -6.2,
		Longitude:      106.8,
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-00000000000a", "00000000-0000-0000-0000-00000000000b"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, repo.Create(ctx, f))

	got, err := svc.GetByID(ctx, f.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"00000000-0000-0000-0000-00000000000a"}, got.ImagesMediaIDs)
}
