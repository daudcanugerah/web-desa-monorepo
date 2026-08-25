package berita

import (
	"net/http"

	"webdesa/api/interface/http/middleware"
	galleryUsecase "webdesa/api/usecase/gallery"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
)

// BeritaUploadHandler handles media uploads for the Quill editor.
// It writes each upload to the gallery system folder backing the berita
// feature and returns the admin binary URL that the front-end can
// embed in the Quill delta. The admin URL is returned (rather than the
// public URL) because system folders default to is_public=false; the
// front-end already sends a Bearer token, so it can stream the binary
// via the admin endpoint.
type BeritaUploadHandler struct {
	fileStore galleryUsecase.FileStore
	logger    middleware.Logger
}

// NewBeritaUploadHandler creates a new berita upload handler.
func NewBeritaUploadHandler(fileStore galleryUsecase.FileStore, logger middleware.Logger) *BeritaUploadHandler {
	return &BeritaUploadHandler{
		fileStore: fileStore,
		logger:    logger,
	}
}

// UploadMediaResponse represents the upload response
type UploadMediaResponse struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Filename string `json:"filename"`
}

// UploadMedia godoc
// @Summary      Upload media for Quill editor (admin)
// @Description  Multipart: file (image only). Validates against the berita feature allowlist (max size, MIME type). Returns admin content URL plus the media_id the front-end can persist alongside the berita. RBAC: berita:write.
// @Tags         berita-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Media file"
// @Success      200  {object} berita.UploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/upload-media [post]
func (h *BeritaUploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form. 50 MB cap matches the previous Berita upload
	// behaviour and stays well above the gallery image limit; the
	// FileStore validates per-feature size/MIME on SaveImage.
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.logger.Error(ctx, "failed to parse multipart form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error(ctx, "failed to get file from form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Determine content type — the multipart header is the source of
	// truth; fall back to filename inference so the FileStore sees a
	// valid MIME for empty/octet-stream headers.
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = handlerutil.InferContentType(fileHeader.Filename)
	}

	saved, err := h.fileStore.SaveImage(ctx, galleryUsecase.FeatureBerita, galleryUsecase.FileInput{
		OriginalName: fileHeader.Filename,
		Content:      file,
		Size:         fileHeader.Size,
		ContentType:  contentType,
	})
	if err != nil {
		h.logger.Error(ctx, "failed to save upload", "error", err.Error())
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to upload media", err)
		return
	}

	url := galleryUsecase.URLFor(galleryUsecase.URLScopeAdmin, "content", saved.MediaID)

	h.logger.Info(ctx, "media uploaded successfully", "media_id", saved.MediaID, "filename", fileHeader.Filename, "size", fileHeader.Size)

	response.Success(w, http.StatusOK, UploadMediaResponse{
		URL:      url,
		MediaID:  saved.MediaID,
		Filename: fileHeader.Filename,
	})
}
