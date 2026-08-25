package fasilitas

import (
	"net/http"

	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
	galleryUsecase "webdesa/api/usecase/gallery"
)

// FasilitasUploadHandler handles fasilitas image uploads. Each call stores
// one file in the system/fasilitas gallery folder and returns the admin
// content URL plus the media_id the caller attaches via images_media_ids on
// POST /fasilitas or PUT /fasilitas/{id}. Repeatable: call once per image.
type FasilitasUploadHandler struct {
	fileStore galleryUsecase.FileStore
	logger    middleware.Logger
}

// NewFasilitasUploadHandler creates a new fasilitas upload handler.
func NewFasilitasUploadHandler(fileStore galleryUsecase.FileStore, logger middleware.Logger) *FasilitasUploadHandler {
	return &FasilitasUploadHandler{
		fileStore: fileStore,
		logger:    logger,
	}
}

// FasilitasUploadMediaResponse represents the upload response
type FasilitasUploadMediaResponse struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Filename string `json:"filename"`
}

// UploadMedia godoc
// @Summary      Upload fasilitas image (admin)
// @Description  Multipart: file (image only). Validates against the fasilitas image allowlist (max size, MIME type). Returns admin content URL plus the media_id to attach via images_media_ids. RBAC: fasilitas:write.
// @Tags         fasilitas-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Image file"
// @Success      200  {object} fasilitas.FasilitasUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas/upload-media [post]
func (h *FasilitasUploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
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

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = handlerutil.InferContentType(fileHeader.Filename)
	}

	saved, err := h.fileStore.SaveImage(ctx, galleryUsecase.FeatureFasilitas, galleryUsecase.FileInput{
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

	h.logger.Info(ctx, "fasilitas media uploaded successfully", "media_id", saved.MediaID, "filename", fileHeader.Filename, "size", fileHeader.Size)

	response.Success(w, http.StatusOK, FasilitasUploadMediaResponse{
		URL:      url,
		MediaID:  saved.MediaID,
		Filename: fileHeader.Filename,
	})
}
