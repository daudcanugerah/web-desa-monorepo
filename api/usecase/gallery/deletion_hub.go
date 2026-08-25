package gallery

import "context"

// MediaDeletionListener is the port feature repos implement so the
// gallery service can sweep dangling media references out of feature
// tables when a media row is removed (Task 5.1, "on media delete").
// Implementations should best-effort update every row that still lists
// the deleted id; errors are logged, not returned, because the gallery
// row has already been deleted and re-trying the sweep is the GC's job.
type MediaDeletionListener interface {
	OnMediaDeleted(ctx context.Context, mediaID string)
}

// MediaDeletionHub lets feature repos register themselves as listeners.
// The gallery service receives the hub at construction time; callers
// attach concrete feature repos via AddListener. Listeners are invoked
// in registration order after a successful DeleteMedia.
type MediaDeletionHub struct {
	listeners []MediaDeletionListener
}

// NewMediaDeletionHub returns an empty hub.
func NewMediaDeletionHub() *MediaDeletionHub {
	return &MediaDeletionHub{}
}

// AddListener registers a feature repo to be notified on media deletes.
func (h *MediaDeletionHub) AddListener(l MediaDeletionListener) {
	if h == nil || l == nil {
		return
	}
	h.listeners = append(h.listeners, l)
}

// Notify calls every registered listener with the deleted media id.
// Errors from individual listeners are swallowed; the hub's purpose is
// best-effort bookkeeping, not a transactional guarantee.
func (h *MediaDeletionHub) Notify(ctx context.Context, mediaID string) {
	if h == nil {
		return
	}
	for _, l := range h.listeners {
		l.OnMediaDeleted(ctx, mediaID)
	}
}
