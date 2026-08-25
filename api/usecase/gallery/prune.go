package gallery

import "context"

// PruneDanglingMediaIDs returns the subset of mediaIDs whose gallery row
// still exists. Any ID whose FileStore.Open returns ErrMediaNotFound is
// dropped (the deleted media may still appear in legacy `images_media_ids`
// JSONB arrays; this filter is applied on read so feature responses never
// surface dangling UUIDs).
//
// Errors other than ErrMediaNotFound (e.g. storage outage) fail closed:
// the ID is kept and the error wrapped, so callers don't silently lose
// references.
func PruneDanglingMediaIDs(ctx context.Context, fs FileStore, mediaIDs []string) ([]string, error) {
	if len(mediaIDs) == 0 {
		return mediaIDs, nil
	}
	out := make([]string, 0, len(mediaIDs))
	for _, id := range mediaIDs {
		if id == "" {
			continue
		}
		_, err := fs.Open(ctx, id)
		if err == nil {
			out = append(out, id)
			continue
		}
		if err == ErrMediaNotFound {
			continue
		}
		return nil, err
	}
	return out, nil
}
