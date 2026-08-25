package banner

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

	"webdesa/api/domain/banner"
	"webdesa/api/domain/bannercategory"
	"webdesa/api/pkg/clock"
	galleryUsecase "webdesa/api/usecase/gallery"
)

type stubRepo struct {
	mu    sync.Mutex
	rows  map[string]*banner.Banner
	order []string
}

func newStubRepo() *stubRepo {
	return &stubRepo{rows: map[string]*banner.Banner{}}
}

func (r *stubRepo) Create(ctx context.Context, b *banner.Banner) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[b.ID] = b
	r.order = append(r.order, b.ID)
	return nil
}

func (r *stubRepo) FindByID(ctx context.Context, id string) (*banner.Banner, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.rows[id]
	if !ok {
		return nil, errors.New("banner not found")
	}
	return b, nil
}

func (r *stubRepo) List(ctx context.Context, status *string, query *string, category *string, offset, limit int) ([]*banner.Banner, int, error) {
	return nil, 0, nil
}

func (r *stubRepo) Update(ctx context.Context, b *banner.Banner) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[b.ID] = b
	return nil
}

func (r *stubRepo) CountActiveBanners(ctx context.Context) (int, error) {
	return 0, nil
}

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
		return nil, errors.New("media not found")
	}
	return &galleryUsecase.MediaBinary{ContentType: "image/jpeg", Size: 10}, nil
}

type stubCategoryLookup struct {
	known map[string]bool
}

func (m *stubCategoryLookup) FindByID(ctx context.Context, id string) (*bannercategory.Category, error) {
	if m.known != nil && !m.known[id] {
		return nil, errors.New("category not found")
	}
	return &bannercategory.Category{ID: id, Name: "mock"}, nil
}

func TestCreateWithMediaRef(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := NewService(repo, fs, &stubCategoryLookup{}, clock.NewFixedClock(time.Now()))
	ctx := context.Background()

	fs.known["00000000-0000-0000-0000-000000000001"] = true
	ref := "00000000-0000-0000-0000-000000000001"
	b, err := svc.Create(ctx, CreateBannerInput{
		Title:        "Banner Baru",
		ImageMediaID: &ref,
	})
	require.NoError(t, err)
	require.NotNil(t, b.ImageMediaID)
	assert.Equal(t, ref, *b.ImageMediaID)
	assert.Empty(t, fs.saved)
	assert.Empty(t, fs.deleted)

	// unknown ref -> error, nothing persisted
	badRef := "00000000-0000-0000-0000-000000000099"
	_, err = svc.Create(ctx, CreateBannerInput{
		Title:        "Gagal",
		ImageMediaID: &badRef,
	})
	require.Error(t, err)
	assert.Len(t, repo.order, 1)

	// ref + file together -> error
	_, err = svc.Create(ctx, CreateBannerInput{
		Title:        "Konflik",
		ImageMediaID: &ref,
		ImageFile:    bytes.NewReader([]byte("x")),
	})
	require.Error(t, err)
}

func TestCreateCleanupOnCategoryFailure(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	cats := &stubCategoryLookup{known: map[string]bool{"cat-ok": true}}
	svc := NewService(repo, fs, cats, clock.NewFixedClock(time.Now()))
	ctx := context.Background()

	// inline file is created by the op, so it must be cleaned up when a
	// later step (category validation) fails
	badCat := "cat-nope"
	_, err := svc.Create(ctx, CreateBannerInput{
		Title:      "Gagal",
		ImageFile:  bytes.NewReader([]byte("x")),
		ImageName:  "a.jpg",
		ImageSize:  1,
		Category:   &badCat,
		ContentType: "image/jpeg",
	})
	require.Error(t, err)
	assert.Len(t, fs.saved, 1)
	assert.Len(t, fs.deleted, 1)
	assert.Equal(t, fs.saved, fs.deleted)
	assert.Empty(t, fs.known)
	assert.Len(t, repo.order, 0)
}

func TestUpdateWithMediaRef(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := NewService(repo, fs, &stubCategoryLookup{}, clock.NewFixedClock(time.Now()))
	ctx := context.Background()

	old := &banner.Banner{
		ID:           uuid.NewString(),
		Title:        "Lama",
		ImageMediaID: nil,
		Status:       banner.StatusInactive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, old))

	// attach a pre-uploaded ref -> old row updated, nothing saved/deleted
	fs.known["00000000-0000-0000-0000-00000000000a"] = true
	ref := "00000000-0000-0000-0000-00000000000a"
	b, err := svc.Update(ctx, old.ID, UpdateBannerInput{
		Title:        "Baru",
		ImageMediaID: &ref,
	})
	require.NoError(t, err)
	require.NotNil(t, b.ImageMediaID)
	assert.Equal(t, ref, *b.ImageMediaID)
	assert.Empty(t, fs.saved)
	assert.Empty(t, fs.deleted)

	// replace ref with another ref -> old media deleted after success
	fs.known["00000000-0000-0000-0000-00000000000b"] = true
	ref2 := "00000000-0000-0000-0000-00000000000b"
	b, err = svc.Update(ctx, old.ID, UpdateBannerInput{
		Title:        "Baru Lagi",
		ImageMediaID: &ref2,
	})
	require.NoError(t, err)
	require.NotNil(t, b.ImageMediaID)
	assert.Equal(t, ref2, *b.ImageMediaID)
	assert.Contains(t, fs.deleted, ref)
	assert.Empty(t, fs.saved)

	// unknown ref -> error, nothing changed
	badRef := "00000000-0000-0000-0000-000000000099"
	_, err = svc.Update(ctx, old.ID, UpdateBannerInput{
		Title:        "Gagal",
		ImageMediaID: &badRef,
	})
	require.Error(t, err)
	cur, _ := repo.FindByID(ctx, old.ID)
	require.NotNil(t, cur.ImageMediaID)
	assert.Equal(t, ref2, *cur.ImageMediaID)
}

func strptr(s string) *string { return &s }
