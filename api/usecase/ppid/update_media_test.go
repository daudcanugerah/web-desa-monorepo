package ppid

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webdesa/api/domain/ppid"
)

func newValidPPID(id, docID, thumbID string) *ppid.PPID {
	now := time.Now()
	return &ppid.PPID{
		ID:               id,
		Title:            "Laporan Tahunan",
		DocumentMediaID:  strptr(docID),
		ThumbnailMediaID: strptr(thumbID),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// TestUpdate_SameMediaRefsAreKept is a regression test: the admin edit form
// resends the current document/thumbnail media ids on every save. Those must
// NOT be deleted, otherwise the document and cover become inaccessible.
func TestUpdate_SameMediaRefsAreKept(t *testing.T) {
	repo := newMockRepository()
	fs := newTrackingFileStore()
	svc := newTestService(t, repo, fs)
	ctx := context.Background()

	fs.known["doc-1"] = true
	fs.known["thumb-1"] = true
	require.NoError(t, repo.Create(ctx, newValidPPID("ppid-1", "doc-1", "thumb-1")))

	updated, err := svc.Update(ctx, "ppid-1", UpdatePPIDInput{
		Title:            "Laporan Tahunan (Revisi)",
		DocumentMediaID:  strptr("doc-1"),
		ThumbnailMediaID: strptr("thumb-1"),
	})
	require.NoError(t, err)
	assert.Equal(t, "doc-1", *updated.DocumentMediaID)
	assert.Equal(t, "thumb-1", *updated.ThumbnailMediaID)
	assert.Empty(t, fs.deleted, "media resending the same id must not be deleted")
}

// TestUpdate_ReplacedMediaRefsAreDeleted verifies the old media is still
// cleaned up when an edit actually swaps the document/thumbnail.
func TestUpdate_ReplacedMediaRefsAreDeleted(t *testing.T) {
	repo := newMockRepository()
	fs := newTrackingFileStore()
	svc := newTestService(t, repo, fs)
	ctx := context.Background()

	for _, id := range []string{"doc-1", "thumb-1", "doc-2", "thumb-2"} {
		fs.known[id] = true
	}
	require.NoError(t, repo.Create(ctx, newValidPPID("ppid-1", "doc-1", "thumb-1")))

	_, err := svc.Update(ctx, "ppid-1", UpdatePPIDInput{
		Title:            "Laporan Tahunan (Revisi)",
		DocumentMediaID:  strptr("doc-2"),
		ThumbnailMediaID: strptr("thumb-2"),
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"doc-1", "thumb-1"}, fs.deleted,
		"replaced media must be cleaned up")
}
