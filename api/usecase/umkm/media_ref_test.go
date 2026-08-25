package umkm

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

	"webdesa/api/domain/umkm"
	"webdesa/api/domain/umkmcategory"
	"webdesa/api/pkg/clock"
	galleryUsecase "webdesa/api/usecase/gallery"
)

type stubRepo struct {
	mu    sync.Mutex
	rows  map[string]*umkm.UMKM
	order []string
}

func newStubRepo() *stubRepo {
	return &stubRepo{rows: map[string]*umkm.UMKM{}}
}

func (r *stubRepo) Create(ctx context.Context, u *umkm.UMKM) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[u.ID] = u
	r.order = append(r.order, u.ID)
	return nil
}

func (r *stubRepo) FindByID(ctx context.Context, id string) (*umkm.UMKM, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.rows[id]
	if !ok {
		return nil, errors.New("umkm not found")
	}
	return u, nil
}

func (r *stubRepo) List(ctx context.Context, query *string, category *string, offset, limit int) ([]*umkm.UMKM, int, error) {
	return nil, 0, nil
}

func (r *stubRepo) Update(ctx context.Context, u *umkm.UMKM) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[u.ID] = u
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

func (m *stubCategoryLookup) FindByID(ctx context.Context, id string) (*umkmcategory.Category, error) {
	return &umkmcategory.Category{ID: id, Name: "mock"}, nil
}

type stubBulkLimits struct {
	maxFiles int
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

	// pre-uploaded refs are attached directly, nothing saved
	fs.known["00000000-0000-0000-0000-000000000001"] = true
	fs.known["00000000-0000-0000-0000-000000000002"] = true
	u, err := svc.Create(ctx, CreateUMKMInput{
		Name:           "Warung Pak Budi",
		Category:       "cat-1",
		Description:    "desc",
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"}, u.ImagesMediaIDs)
	assert.Empty(t, fs.saved)
	assert.Empty(t, fs.deleted)

	// unknown ref -> error, nothing persisted
	_, err = svc.Create(ctx, CreateUMKMInput{
		Name:           "Gagal",
		Category:       "cat-1",
		Description:    "desc",
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-000000000099"},
	})
	require.Error(t, err)
	assert.Len(t, repo.order, 1)

	// refs + inline files merge (refs first)
	fs.known["00000000-0000-0000-0000-000000000003"] = true
	u, err = svc.Create(ctx, CreateUMKMInput{
		Name:           "Campuran",
		Category:       "cat-1",
		Description:    "desc",
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-000000000003"},
		ImageFiles: []ImageInput{
			{Filename: "a.jpg", Content: bytes.NewReader([]byte("a")), Size: 1, ContentType: "image/jpeg"},
		},
	})
	require.NoError(t, err)
	require.Len(t, u.ImagesMediaIDs, 2)
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", u.ImagesMediaIDs[0])
	assert.NotEqual(t, "", u.ImagesMediaIDs[1])
	assert.Len(t, fs.saved, 1)
}

func TestUpdateWithMediaRefs(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := newTestService(repo, fs)
	ctx := context.Background()

	old := &umkm.UMKM{
		ID:             uuid.NewString(),
		Name:           "Lama",
		Category:       "cat-1",
		CategoryName:   strptr("mock"),
		Description:    "d",
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-00000000000a"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, repo.Create(ctx, old))
	fs.known["00000000-0000-0000-0000-00000000000a"] = true

	// replace via refs -> old media deleted, no new saves
	fs.known["00000000-0000-0000-0000-00000000000b"] = true
	fs.known["00000000-0000-0000-0000-00000000000c"] = true
	u, err := svc.Update(ctx, old.ID, UpdateUMKMInput{
		Name:           "Baru",
		Category:       "cat-1",
		Description:    "d",
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-00000000000b", "00000000-0000-0000-0000-00000000000c"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"00000000-0000-0000-0000-00000000000b", "00000000-0000-0000-0000-00000000000c"}, u.ImagesMediaIDs)
	assert.Contains(t, fs.deleted, "00000000-0000-0000-0000-00000000000a")
	assert.Empty(t, fs.saved)

	// unknown ref -> error, nothing changed
	_, err = svc.Update(ctx, old.ID, UpdateUMKMInput{
		Name:           "Gagal",
		Category:       "cat-1",
		Description:    "d",
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-000000000099"},
	})
	require.Error(t, err)
	cur, _ := repo.FindByID(ctx, old.ID)
	assert.Equal(t, []string{"00000000-0000-0000-0000-00000000000b", "00000000-0000-0000-0000-00000000000c"}, cur.ImagesMediaIDs)
}

func strptr(s string) *string { return &s }

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
	_, err := svc.Create(ctx, CreateUMKMInput{Name: "Banjir", Category: "cat-1", Description: "d", ImageFiles: files})
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
	_, err = svc.Create(ctx, CreateUMKMInput{Name: "Banjir", Category: "cat-1", Description: "d", ImagesMediaIDs: refs, ImageFiles: files[:1]})
	require.ErrorIs(t, err, galleryUsecase.ErrBulkTooManyFiles)
	assert.Empty(t, fs.saved)

	// 20 files within the count cap but >250MB total -> ErrBulkTotalTooLarge
	big := make([]ImageInput, 20)
	for i := range big {
		big[i] = ImageInput{Filename: "b.jpg", Content: bytes.NewReader([]byte("x")), Size: 13 * 1024 * 1024, ContentType: "image/jpeg"}
	}
	_, err = svc.Create(ctx, CreateUMKMInput{Name: "Banjir", Category: "cat-1", Description: "d", ImageFiles: big})
	require.ErrorIs(t, err, galleryUsecase.ErrBulkTotalTooLarge)
	assert.Empty(t, fs.saved)
}

func TestUpdateBulkCaps(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := newTestService(repo, fs)
	ctx := context.Background()

	old := &umkm.UMKM{
		ID:             uuid.NewString(),
		Name:           "Lama",
		Category:       "cat-1",
		CategoryName:   strptr("mock"),
		Description:    "d",
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-00000000000a"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, repo.Create(ctx, old))

	files := make([]ImageInput, 21)
	for i := range files {
		files[i] = ImageInput{Filename: "a.jpg", Content: bytes.NewReader([]byte("x")), Size: 1024, ContentType: "image/jpeg"}
	}
	_, err := svc.Update(ctx, old.ID, UpdateUMKMInput{Name: "Banjir", Category: "cat-1", Description: "d", ImageFiles: files})
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
	u := &umkm.UMKM{
		ID:             uuid.NewString(),
		Name:           "Mixed",
		Category:       "cat-1",
		CategoryName:   strptr("mock"),
		Description:    "d",
		ImagesMediaIDs: []string{"00000000-0000-0000-0000-00000000000a", "00000000-0000-0000-0000-00000000000b"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, repo.Create(ctx, u))

	got, err := svc.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"00000000-0000-0000-0000-00000000000a"}, got.ImagesMediaIDs)
}

