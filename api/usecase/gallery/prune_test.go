package gallery

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubListener struct {
	mu     sync.Mutex
	gotIDs []string
}

func (l *stubListener) OnMediaDeleted(_ context.Context, mediaID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.gotIDs = append(l.gotIDs, mediaID)
}

type stubFileStorePrune struct {
	mu    sync.Mutex
	known map[string]bool
}

func (f *stubFileStorePrune) SaveImage(_ context.Context, _ string, _ FileInput) (SavedFile, error) {
	return SavedFile{}, errors.New("not implemented")
}
func (f *stubFileStorePrune) SaveDocument(_ context.Context, _ string, _ FileInput) (SavedFile, error) {
	return SavedFile{}, errors.New("not implemented")
}
func (f *stubFileStorePrune) SaveImages(_ context.Context, _ string, _ []FileInput) ([]SavedFile, []error) {
	return nil, nil
}
func (f *stubFileStorePrune) ValidateFiles(_ context.Context, _ string, _ []FileInput) []error {
	return nil
}
func (f *stubFileStorePrune) Delete(_ context.Context, _ string) error { return nil }
func (f *stubFileStorePrune) Open(_ context.Context, id string) (*MediaBinary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.known[id] {
		return nil, ErrMediaNotFound
	}
	return &MediaBinary{ContentType: "image/jpeg", Size: 1}, nil
}

func TestPruneDanglingMediaIDs(t *testing.T) {
	fs := &stubFileStorePrune{known: map[string]bool{
		"live-1": true,
		"live-2": true,
	}}
	in := []string{"live-1", "dead-1", "", "live-2", "dead-2"}
	got, err := PruneDanglingMediaIDs(context.Background(), fs, in)
	require.NoError(t, err)
	assert.Equal(t, []string{"live-1", "live-2"}, got)
}

func TestPruneDanglingMediaIDsEmpty(t *testing.T) {
	fs := &stubFileStorePrune{known: map[string]bool{}}
	got, err := PruneDanglingMediaIDs(context.Background(), fs, nil)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestMediaDeletionHubNotifies(t *testing.T) {
	hub := NewMediaDeletionHub()
	l1 := &stubListener{}
	l2 := &stubListener{}
	hub.AddListener(l1)
	hub.AddListener(l2)
	hub.Notify(context.Background(), "abc")
	assert.Equal(t, []string{"abc"}, l1.gotIDs)
	assert.Equal(t, []string{"abc"}, l2.gotIDs)
}

func TestMediaDeletionHubNilReceiver(t *testing.T) {
	var hub *MediaDeletionHub
	hub.AddListener(&stubListener{})
	hub.Notify(context.Background(), "x") // must not panic
}
