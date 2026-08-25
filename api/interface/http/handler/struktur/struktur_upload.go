package struktur

import (
	"net/http"

	"webdesa/api/interface/http/middleware"
	galleryUsecase "webdesa/api/usecase/gallery"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
)

// StrukturUploadHandler handles struktur profile image uploads. Each upload
// is written to the system/struktur gallery folder and returns the admin
// binary URL; public struktur responses translate that to a feature-scoped
// public URL the anonymous site can stream.
type StrukturUploadHandler struct {
	fileStore galleryUsecase.FileStore
	logger    middleware.Logger
}

// NewStrukturUploadHandler creates a new struktur upload handler.
func NewStrukturUploadHandler(fileStore galleryUsecase.FileStore, logger middleware.Logger) *StrukturUploadHandler {
	return &StrukturUploadHandler{
		fileStore: fileStore,
		logger:    logger,
	}
}

// StrukturUploadMediaResponse represents the upload response
type StrukturUploadMediaResponse struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Filename string `json:"filename"`
}

// UploadMedia godoc
// @Summary      Upload struktur profile image (admin)
// @Description  Multipart: file (image only). Validates against the struktur feature allowlist (max size, MIME type). Returns admin content URL plus the media_id the front-end can persist alongside the struktur member. RBAC: struktur:write.
// @Tags         struktur-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Media file"
// @Success      200  {object} struktur.StrukturUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /struktur/upload-media [post]
func (h *StrukturUploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.logger.Error(ctx, "failed to parse multipart form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error(ctx, "failed to get file from form", "error", err.Error())
		response.Error(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// The multipart header is the source of truth; fall back to filename
	// inference so the FileStore sees a valid MIME for empty/octet-stream
	// headers.
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = handlerutil.InferContentType(fileHeader.Filename)
	}

	saved, err := h.fileStore.SaveImage(ctx, galleryUsecase.FeatureStruktur, galleryUsecase.FileInput{
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

	h.logger.Info(ctx, "struktur media uploaded successfully", "media_id", saved.MediaID, "filename", fileHeader.Filename, "size", fileHeader.Size)

	response.Success(w, http.StatusOK, StrukturUploadMediaResponse{
		URL:      url,
		MediaID:  saved.MediaID,
		Filename: fileHeader.Filename,
	})
}
