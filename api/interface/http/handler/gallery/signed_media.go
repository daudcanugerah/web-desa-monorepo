package gallery

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/response"
	galleryuc "webdesa/api/usecase/gallery"
)

// SignedMediaHandler serves the unified `/api/v1/media/{id}/...`
// endpoint introduced in Task 7.1. The JWT in the query string is the
// auth — no Authorization header needed, so the route works seamlessly
// for `<img src>` and CDN caches. The same handler backs every scope
// (public/admin/ppid) and the verify step in SignedURLService rejects
// mismatches, expired tokens, and revoked subs.
type SignedMediaHandler struct {
	gallery   *galleryuc.Service
	signedURL *galleryuc.SignedURLService
	logger    middleware.Logger
}

// NewSignedMediaHandler wires the handler. Pass a nil signer when signed
// URLs are disabled — the routes return 503 Service Unavailable so
// operators notice the misconfig.
func NewSignedMediaHandler(gallery *galleryuc.Service, signedURL *galleryuc.SignedURLService, logger middleware.Logger) *SignedMediaHandler {
	return &SignedMediaHandler{gallery: gallery, signedURL: signedURL, logger: logger}
}

// ServeContent godoc
// @Summary      Stream media content (signed URL)
// @Description  Streams the original binary when the supplied JWT is valid and bound to the media id in the path. Scope is inferred from the token (public/admin/ppid).
// @Tags         media
// @Produce      octet-stream
// @Param        id   path     string  true   "Media UUID"
// @Param        jwt  query    string  true   "Signed media token"
// @Success      200  {string}  format=binary  "Binary stream"
// @Failure      401  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Router       /media/{id}/content [get]
func (h *SignedMediaHandler) ServeContent(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, "content")
}

// ServeThumbnail godoc
// @Summary      Stream media thumbnail (signed URL)
// @Description  Streams the thumbnail when the supplied JWT is valid and bound to the media id in the path.
// @Tags         media
// @Produce      octet-stream
// @Param        id   path     string  true   "Media UUID"
// @Param        jwt  query    string  true   "Signed media token"
// @Success      200  {string}  format=binary  "Binary stream"
// @Failure      401  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Router       /media/{id}/thumbnail [get]
func (h *SignedMediaHandler) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, "thumbnail")
}

// serve is the shared dispatch: validate the JWT, then look up the
// media. Content vs thumbnail only changes which gallery method we call.
func (h *SignedMediaHandler) serve(w http.ResponseWriter, r *http.Request, kind string) {
	if h.signedURL == nil {
		response.Error(w, http.StatusServiceUnavailable, "Signed URL service is not configured")
		return
	}
	id := chi.URLParam(r, "id")
	if !galleryuc.IsValidMediaID(id) {
		response.Error(w, http.StatusNotFound, "Media not found")
		return
	}
	token := r.URL.Query().Get("jwt")
	claims, err := h.signedURL.Verify(token, id)
	if err != nil {
		h.logger.Info(r.Context(), "signed media token rejected", "media_id", id, "error", err.Error())
		response.Error(w, http.StatusUnauthorized, "Invalid or expired link")
		return
	}

	var (
		bin    *galleryuc.MediaBinary
		binErr error
	)
	// Scope gates which binary a token may read:
	//   public → public-folder media only (folder.is_public && media.is_public)
	//   admin  → any media (system folders, private uploads, PPID docs)
	//   ppid   → the media bound to the approved PPID request (sub match
	//            already verified by SignedURLService; we route through the
	//            PPID system folder check below).
	switch claims.Scope {
	case galleryuc.ScopePublic:
		if kind == "content" {
			bin, binErr = h.gallery.GetSignedPublicMediaContent(r.Context(), id)
		} else {
			bin, binErr = h.gallery.GetSignedPublicMediaThumbnail(r.Context(), id)
		}
	case galleryuc.ScopePPID:
		// PPID document media lives in the system/ppid folder. Route the
		// read through the feature-scoped accessor so only media that
		// belongs to that folder can ever be streamed with a ppid token.
		if kind == "content" {
			bin, binErr = h.gallery.GetFeatureMediaForPublic(r.Context(), galleryuc.FeaturePPID, id)
		} else {
			bin, binErr = h.gallery.GetFeatureMediaThumbnailForPublic(r.Context(), galleryuc.FeaturePPID, id)
		}
	default: // admin + any unknown scope fall back to the admin accessor
		if kind == "content" {
			bin, binErr = h.gallery.GetMediaContent(r.Context(), id)
		} else {
			bin, binErr = h.gallery.GetMediaThumbnail(r.Context(), id)
		}
	}
	if binErr != nil {
		h.logger.Info(r.Context(), "signed media binary not found", "media_id", id, "kind", kind, "scope", claims.Scope)
		response.Error(w, http.StatusNotFound, "Media not found")
		return
	}

	streamSignedBinary(w, r, bin, claims.Scope == galleryuc.ScopePublic)
}

// streamSignedBinary sets headers (Cache-Control based on scope) and
// serves the file. The function lives here to keep signed-URL behaviour
// in one place; the legacy admin/public handlers reuse the shared
// streamBinary in gallery.go.
func streamSignedBinary(w http.ResponseWriter, r *http.Request, bin *galleryuc.MediaBinary, publicAccess bool) {
	if bin == nil || bin.FilePath == "" {
		response.Error(w, http.StatusNotFound, "Media not found")
		return
	}
	w.Header().Set("Content-Type", bin.ContentType)
	if bin.Size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", bin.Size))
	}
	if publicAccess {
		w.Header().Set("Cache-Control", "public, max-age=300")
	} else {
		w.Header().Set("Cache-Control", "private, max-age=300")
	}
	http.ServeFile(w, r, bin.FilePath)
}
// RefreshMedia godoc
// @Summary      Refresh a signed media URL
// @Description  Validates the supplied JWT and mints a fresh one carrying the same scope and sub. The old token is left intact — expiry is the only invalidation. Used by clients to renew expiring admin URLs transparently.
// @Tags         media
// @Produce      json
// @Param        jwt  query    string  true   "Current signed media token"
// @Param        id   query    string  false  "Media UUID — required to verify the token matches the media it was issued for"
// @Success      200  {object}  RefreshMediaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Router       /media/refresh [post]
func (h *SignedMediaHandler) RefreshMedia(w http.ResponseWriter, r *http.Request) {
	if h.signedURL == nil {
		response.Error(w, http.StatusServiceUnavailable, "Signed URL service is not configured")
		return
	}
	token := r.URL.Query().Get("jwt")
	if token == "" {
		response.Error(w, http.StatusBadRequest, "Missing ?jwt= query parameter")
		return
	}
	mediaID := r.URL.Query().Get("id")

	fresh, claims, err := h.signedURL.Refresh(token, mediaID, 0)
	if err != nil {
		h.logger.Info(r.Context(), "signed media refresh failed", "media_id", mediaID, "error", err.Error())
		response.Error(w, http.StatusUnauthorized, "Invalid or expired link")
		return
	}

	url := galleryuc.SignedURLPath("content", claims.MediaID) + "?jwt=" + fresh
	thumbURL := galleryuc.SignedURLPath("thumbnail", claims.MediaID) + "?jwt=" + fresh
	response.Success(w, http.StatusOK, RefreshMediaResponse{
		URL:           url,
		ThumbnailURL:  thumbURL,
		MediaID:       claims.MediaID,
		Scope:         string(claims.Scope),
		ExpiresAtUnix: claims.ExpiresAt.Unix(),
	})
}

// RefreshMediaResponse is the success body for /api/v1/media/refresh.
type RefreshMediaResponse struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	MediaID      string `json:"media_id"`
	Scope        string `json:"scope"`
	// ExpiresAtUnix is the new token's exp claim as a Unix timestamp.
	// Clients can use it to schedule the next refresh.
	ExpiresAtUnix int64 `json:"expires_at_unix"`
}
