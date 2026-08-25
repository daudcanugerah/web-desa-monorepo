package user

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webdesa/api/domain/user"
	galleryUsecase "webdesa/api/usecase/gallery"
	"webdesa/api/pkg/clock"
)

type stubRepo struct {
	mu   sync.Mutex
	rows map[string]*user.User
}

func newStubRepo() *stubRepo {
	return &stubRepo{rows: map[string]*user.User{}}
}

func (r *stubRepo) Create(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[u.ID] = u
	return nil
}

func (r *stubRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.rows[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (r *stubRepo) Update(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[u.ID] = u
	return nil
}

func (r *stubRepo) UpdateProfileImage(ctx context.Context, id string, mediaID *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.rows[id]
	if !ok {
		return errors.New("user not found")
	}
	u.ProfileImageMediaID = mediaID
	return nil
}

func (r *stubRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rows, id)
	return nil
}

func (r *stubRepo) List(ctx context.Context, offset, limit int) ([]*user.User, int, error) {
	return nil, 0, nil
}

func (r *stubRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, errors.New("not found")
}

func (r *stubRepo) FindByEmailExcluding(ctx context.Context, email, excludeID string) (*user.User, error) {
	return nil, errors.New("not found")
}

func (r *stubRepo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	return nil
}

func (r *stubRepo) UpdateLastLogin(ctx context.Context, id string, t time.Time) error {
	return nil
}

type stubFileStore struct {
	mu      sync.Mutex
	deleted []string
	known   map[string]bool
}

func newStubFileStore() *stubFileStore {
	return &stubFileStore{known: map[string]bool{}}
}

func (f *stubFileStore) SaveImage(ctx context.Context, feature string, in galleryUsecase.FileInput) (galleryUsecase.SavedFile, error) {
	id := uuid.NewString()
	f.mu.Lock()
	defer f.mu.Unlock()
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
	return &galleryUsecase.MediaBinary{ContentType: "image/jpeg"}, nil
}

func TestUpdateProfileImageMediaID(t *testing.T) {
	repo := newStubRepo()
	fs := newStubFileStore()
	svc := NewService(repo, fs, clock.NewFixedClock(time.Now()))
	ctx := context.Background()

	u := &user.User{
		ID:                  uuid.NewString(),
		Name:                "Budi",
		Email:               "budi@example.com",
		ProfileImageMediaID: strptr("00000000-0000-0000-0000-00000000000a"),
	}
	require.NoError(t, repo.Create(ctx, u))
	fs.known["00000000-0000-0000-0000-00000000000a"] = true

	// unknown media -> error, nothing changed
	_, err := svc.UpdateProfileImageMediaID(ctx, u.ID, "00000000-0000-0000-0000-000000000099")
	require.Error(t, err)
	cur, _ := repo.FindByID(ctx, u.ID)
	assert.Equal(t, "00000000-0000-0000-0000-00000000000a", *cur.ProfileImageMediaID)

	// valid media -> attached, old deleted
	fs.known["00000000-0000-0000-0000-00000000000b"] = true
	got, err := svc.UpdateProfileImageMediaID(ctx, u.ID, "00000000-0000-0000-0000-00000000000b")
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-00000000000b", *got.ProfileImageMediaID)
	assert.Contains(t, fs.deleted, "00000000-0000-0000-0000-00000000000a")
}

func strptr(s string) *string { return &s }
