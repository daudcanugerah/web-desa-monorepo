package user

import (
	"net/http"

	"webdesa/api/interface/http/middleware"
	galleryUsecase "webdesa/api/usecase/gallery"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
)

// UserUploadHandler handles user avatar uploads. Each upload is written to
// the system/user gallery folder and returns the admin binary URL plus the
// media_id the caller can attach via profile_image_media_id on
// PUT /users/{id} or PUT /users/me.
type UserUploadHandler struct {
	fileStore galleryUsecase.FileStore
	logger    middleware.Logger
}

// NewUserUploadHandler creates a new user upload handler.
func NewUserUploadHandler(fileStore galleryUsecase.FileStore, logger middleware.Logger) *UserUploadHandler {
	return &UserUploadHandler{
		fileStore: fileStore,
		logger:    logger,
	}
}

// UserUploadMediaResponse represents the upload response
type UserUploadMediaResponse struct {
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`
	Filename string `json:"filename"`
}

// UploadMedia godoc
// @Summary      Upload user avatar media (admin)
// @Description  Multipart: file (image only). Validates against the user feature allowlist (max size, MIME type). Returns admin content URL plus the media_id to attach via PUT /users/{id}. RBAC: users:write.
// @Tags         user-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Media file"
// @Success      200  {object} user.UserUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/upload-media [post]
func (h *UserUploadHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	h.uploadMedia(w, r, nil)
}

// UploadMediaSelf godoc
// @Summary      Upload own avatar media (authenticated)
// @Description  Multipart: file (image only). Returns admin content URL plus the media_id to attach via PUT /users/me. Requires JWT; no RBAC.
// @Tags         user-upload
// @Accept       mpfd
// @Produce      json
// @Param        file       formData file   true   "Media file"
// @Success      200  {object} user.UserUploadMediaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/upload-media [post]
func (h *UserUploadHandler) UploadMediaSelf(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	h.uploadMedia(w, r, &userID)
}

func (h *UserUploadHandler) uploadMedia(w http.ResponseWriter, r *http.Request, _ *string) {
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

	saved, err := h.fileStore.SaveImage(ctx, galleryUsecase.FeatureUser, galleryUsecase.FileInput{
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

	h.logger.Info(ctx, "user media uploaded successfully", "media_id", saved.MediaID, "filename", fileHeader.Filename, "size", fileHeader.Size)

	response.Success(w, http.StatusOK, UserUploadMediaResponse{
		URL:      url,
		MediaID:  saved.MediaID,
		Filename: fileHeader.Filename,
	})
}
