package gallery

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	PublicContentPath   = "/api/v1/public/gallery/media/%s/content"
	PublicThumbnailPath = "/api/v1/public/gallery/media/%s/thumbnail"
	AdminContentPath    = "/api/v1/gallery/media/%s/content"
	AdminThumbnailPath  = "/api/v1/gallery/media/%s/thumbnail"

	// FeaturePublicContentPath / FeaturePublicThumbnailPath are the
	// feature-scoped public stream routes for system-folder media
	// (e.g. /api/v1/public/banner/media/{id}/content).
	FeaturePublicContentPath   = "/api/v1/public/%s/media/%s/content"
	FeaturePublicThumbnailPath = "/api/v1/public/%s/media/%s/thumbnail"

	// SignedContentPath / SignedThumbnailPath are the unified
	// /api/v1/media/{id}/...?jwt= routes introduced by Task 7.1. They
	// supersede the per-scope paths once SignedURLsEnabled is flipped on.
	SignedContentPath   = "/api/v1/media/%s/content"
	SignedThumbnailPath = "/api/v1/media/%s/thumbnail"
)

// SignedURLPath returns the unified media URL for content or thumbnail,
// bound to a specific mediaID. The returned path does not include the
// JWT — handlers attach it via SignedURLQuery.
func SignedURLPath(kind, mediaID string) string {
	switch strings.ToLower(kind) {
	case "content":
		return fmt.Sprintf(SignedContentPath, mediaID)
	case "thumbnail":
		return fmt.Sprintf(SignedThumbnailPath, mediaID)
	}
	return ""
}

// SignedURLQuery signs a media URL and returns the "?jwt=<token>" suffix
// (including the leading "?"). ttl <= 0 falls back to the service default
// for the scope. Returns "" when signer is nil (legacy mode).
func (s *SignedURLService) SignedURLQuery(scope SignedURLScope, mediaID, sub string, ttl time.Duration) string {
	if s == nil {
		return ""
	}
	tok, err := s.Sign(scope, mediaID, sub, ttl)
	if err != nil {
		return ""
	}
	return "?jwt=" + url.QueryEscape(tok)
}

// URLScope is the visibility scope for a generated binary URL.
type URLScope string

const (
	URLScopePublic URLScope = "public"
	URLScopeAdmin  URLScope = "admin"
)

// URLFor returns the binary HTTP path for a given media id and scope.
// The gallery binary endpoints are the only place that streams gallery
// media; admin binaries require gallery:read, public binaries require
// both folder.is_public and media.is_public.
func URLFor(scope URLScope, kind, mediaID string) string {
	switch strings.ToLower(kind) {
	case "content":
		switch scope {
		case URLScopePublic:
			return fmt.Sprintf(PublicContentPath, mediaID)
		default:
			return fmt.Sprintf(AdminContentPath, mediaID)
		}
	case "thumbnail":
		switch scope {
		case URLScopePublic:
			return fmt.Sprintf(PublicThumbnailPath, mediaID)
		default:
			return fmt.Sprintf(AdminThumbnailPath, mediaID)
		}
	}
	return ""
}

// FeatureURLFor returns the feature-scoped public binary path for a media
// id. These routes stream system-folder media for one feature only (banner
// covers, berita images) without exposing the generic gallery public
// endpoints.
func FeatureURLFor(feature, kind, mediaID string) string {
	switch strings.ToLower(kind) {
	case "content":
		return fmt.Sprintf(FeaturePublicContentPath, feature, mediaID)
	case "thumbnail":
		return fmt.Sprintf(FeaturePublicThumbnailPath, feature, mediaID)
	}
	return ""
}

// IsValidMediaID returns true when id is a non-empty UUID-shaped string.
func IsValidMediaID(id string) bool {
	if id == "" {
		return false
	}
	if len(id) < 32 {
		return false
	}
	for _, r := range id {
		if r == '-' {
			continue
		}
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

// SignedContentURL returns the full content URL for a media id, including
// the ?jwt= query string. ttl <= 0 falls back to the per-scope default.
// sub is opaque to the signer (e.g. "anonymous", "user:<id>", "ppid_request:<id>").
func (s *SignedURLService) SignedContentURL(mediaID, sub string, ttl time.Duration) string {
	path := SignedURLPath("content", mediaID)
	return path + s.SignedURLQuery(ScopePublic, mediaID, sub, ttl)
}

// SignedThumbnailURL is the thumbnail counterpart of SignedContentURL.
func (s *SignedURLService) SignedThumbnailURL(mediaID, sub string, ttl time.Duration) string {
	path := SignedURLPath("thumbnail", mediaID)
	return path + s.SignedURLQuery(ScopePublic, mediaID, sub, ttl)
}
