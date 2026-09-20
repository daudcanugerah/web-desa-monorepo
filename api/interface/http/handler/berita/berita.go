package berita

import (
	"net/http"
	"strings"
	"time"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	domainberita "webdesa/api/domain/berita"
	"webdesa/api/pkg/response"
	"webdesa/api/usecase/berita"
	galleryuc "webdesa/api/usecase/gallery"

	"webdesa/api/pkg/handlerutil"
)

// BeritaHandler handles HTTP requests for berita (news) management operations.
// It accepts the berita service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
//
// Task 7.1.6: signedURL + signedURLsEnabled mirror the banner handler.
// When enabled, media objects carry ?jwt= URLs pointing at the unified
// /api/v1/media/{id}/... route; when disabled, legacy FeatureURLFor /
// URLFor paths are emitted.
type BeritaHandler struct {
	beritaService     *berita.Service
	signedURL         *galleryuc.SignedURLService
	signedURLsEnabled bool
}

// NewBeritaHandler creates a new berita handler with service dependency injected.
func NewBeritaHandler(beritaService *berita.Service, signedURL *galleryuc.SignedURLService, signedURLsEnabled bool) *BeritaHandler {
	return &BeritaHandler{
		beritaService:     beritaService,
		signedURL:         signedURL,
		signedURLsEnabled: signedURLsEnabled,
	}
}

// ptrStr returns a pointer to the given string.
func ptrStr(s string) *string {
	return &s
}

// buildCategoryInfo builds a response.CategoryInfo pointer from id and name pointers.
// Returns nil when id is nil (e.g. uncategorised record).
func buildCategoryInfo(id *string, name *string) *response.CategoryInfo {
	if id == nil {
		return nil
	}
	n := ""
	if name != nil {
		n = *name
	}
	return &response.CategoryInfo{ID: *id, Name: n}
}

// resolveImageURL returns the admin binary URL for the gallery media, or
// falls back to the legacy string URL when the row predates the gallery
// port. Returns nil when neither is set.
// buildBeritaMedia emits the shared media shape (Task 4.3): media_id plus
// content/thumbnail URLs for gallery rows. When signed URLs are enabled
// (Task 7.1.6), the URLs embed ?jwt=... so the unified
// /api/v1/media/{id}/...?jwt= route serves the binary.
func (h *BeritaHandler) buildBeritaMedia(imageMediaID *string, public bool) *response.MediaInfo {
	if imageMediaID == nil || *imageMediaID == "" {
		return nil
	}
	id := *imageMediaID
	scope := galleryuc.ScopePublic
	sub := "anonymous"
	if !public {
		scope = galleryuc.ScopeAdmin
		sub = "user:admin" // tightened in 7.3 to real user id
	}

	if h.signedURLsEnabled && h.signedURL != nil {
		url := galleryuc.SignedURLPath("content", id) + h.signedURL.SignedURLQuery(scope, id, sub, 0)
		thumb := galleryuc.SignedURLPath("thumbnail", id) + h.signedURL.SignedURLQuery(scope, id, sub, 0)
		return &response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb}
	}

	var url, thumb string
	if public {
		url = galleryuc.FeatureURLFor(galleryuc.FeatureBerita, "content", id)
		thumb = galleryuc.FeatureURLFor(galleryuc.FeatureBerita, "thumbnail", id)
	} else {
		url = galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", id)
		thumb = galleryuc.URLFor(galleryuc.URLScopeAdmin, "thumbnail", id)
	}
	return &response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb}
}

func resolveImageURL(imageURL *string, imageMediaID *string) *string {
	return resolveImageURLScoped(imageURL, imageMediaID, false)
}

// resolveImageURLScoped returns a gallery binary URL for the media, or the
// legacy string URL when the row predates the gallery port. Public scopes
// use the feature-scoped public stream route so anonymous visitors can
// render system-folder media (banner covers, berita images).
func resolveImageURLScoped(imageURL *string, imageMediaID *string, public bool) *string {
	if imageMediaID != nil && *imageMediaID != "" {
		var u string
		if public {
			u = galleryuc.FeatureURLFor(galleryuc.FeatureBerita, "content", *imageMediaID)
		} else {
			u = galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", *imageMediaID)
		}
		return &u
	}
	if imageURL != nil && *imageURL != "" {
		return imageURL
	}
	return nil
}

// BeritaResponse represents the response for berita data
type BeritaResponse struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	CategoryID *string                `json:"category_id,omitempty"`
	Category   *response.CategoryInfo `json:"category,omitempty"`
	Media      *response.MediaInfo    `json:"media,omitempty"`
	Status     string                 `json:"status"`
	CreatedAt  string                 `json:"created_at"` // ISO 8601 format
	UpdatedAt  string                 `json:"updated_at"` // ISO 8601 format
}

// BeritaListResponse represents the response for berita list (without content)
type BeritaListResponse struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	CategoryID *string                `json:"category_id,omitempty"`
	Category   *response.CategoryInfo `json:"category,omitempty"`
	Media      *response.MediaInfo    `json:"media,omitempty"`
	Status     string                 `json:"status"`
	CreatedAt  string                 `json:"created_at"` // ISO 8601 format
	UpdatedAt  string                 `json:"updated_at"` // ISO 8601 format
}

// BeritaPaginatedResponse represents a paginated list of berita
type BeritaPaginatedResponse struct {
	Berita     []BeritaListResponse   `json:"berita"`
	Pagination map[string]interface{} `json:"pagination"`
}

// CreateBerita godoc
// @Summary      Create news article (admin)
// @Description  Multipart: title (required), content (Quill delta JSON, required), category (UUID required), status (optional, defaults to active), optional image or image_media_id (mutually exclusive). RBAC: berita:write.
// @Tags         berita
// @Accept       mpfd
// @Produce      json
// @Param        title      formData string true   "Title"
// @Param        content    formData string true   "Quill delta JSON"
// @Param        category   formData string true   "Category UUID"
// @Param        status     formData string false  "Publication status (active|inactive), default active" Enums(active, inactive)
// @Param        image	formData	file	false	"Cover image"
// @Param        image_media_id  formData  string  false  "Gallery media UUID (mutually exclusive with image)"
// @Success      201  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita [post]
func (h *BeritaHandler) CreateBerita(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string       `in:"form=title" validate:"required,min=1"`
		Content      string       `in:"form=content" validate:"required,min=1"`
		Category     string       `in:"form=category" validate:"required,min=1"`
		Status       *string      `in:"form=status" validate:"omitempty,oneof=active inactive"`
		Image        *httpin.File `in:"form=image"`
		ImageMediaID *string      `in:"form=image_media_id"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Validate required fields
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	if input.Image != nil && input.ImageMediaID != nil && *input.ImageMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either image or image_media_id, not both")
		return
	}

	// Get image file (optional)
	var beritaInput berita.CreateBeritaInput
	if input.Image != nil {
		imageFile, err := input.Image.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read image file")
			return
		}
		contentType := input.Image.MIMEHeader().Get("Content-Type")
		if contentType == "" || contentType == "application/octet-stream" {
			contentType = handlerutil.InferContentType(input.Image.Filename())
		}
		beritaInput = berita.CreateBeritaInput{
			Title:       input.Title,
			Content:     input.Content,
			Category:    input.Category,
			Status:      input.Status,
			ImageFile:   imageFile,
			ImageName:   input.Image.Filename(),
			ImageSize:   input.Image.Size(),
			ContentType: contentType,
		}
	} else {
		// No image file provided; media ref (if any) is passed through
		beritaInput = berita.CreateBeritaInput{
			Title:        input.Title,
			Content:      input.Content,
			Category:     input.Category,
			Status:       input.Status,
			ImageFile:    nil,
			ImageMediaID: input.ImageMediaID,
		}
	}

	// Call berita service
	b, err := h.beritaService.Create(r.Context(), beritaInput)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create berita", err)
		return
	}

	// Build response
	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    b.Content,
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		Media:      h.buildBeritaMedia(b.ImageMediaID, false),
		Status:     b.Status,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListBerita godoc
// @Summary      List news (admin)
// @Description  Paginated list. Query: q, category (UUID), status (active|inactive), since, until (YYYY-MM-DD), sort (created_at|title), order (asc|desc). RBAC: berita:read.
// @Tags         berita
// @Produce      json
//
//	@Param	q	query	string	false	"Search query"
//	@Param	category	query	string	false	"Category UUID"
//	@Param	status	query	string	false	"Filter by status"	Enums(active, inactive)
//	@Param	since	query	string	false	"Start date YYYY-MM-DD"
//	@Param	until	query	string	false	"End date YYYY-MM-DD"
//	@Param	sort	query	string	false	"Sort field: created_at or title (default created_at)"
//	@Param	order	query	string	false	"Sort direction: asc or desc (default desc - newest first)"
//	@Param	page	query	int	false	"Page default(1)"
//	@Param	limit	query	int	false	"Page size default(10)"
//
// @Success      200  {object} berita.BeritaPaginatedResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita [get]
func (h *BeritaHandler) ListBerita(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int     `in:"query=page;default=1" validate:"min=1"`
		Limit    int     `in:"query=limit;default=10" validate:"min=1"`
		Q        *string `in:"query=q"`
		Category *string `in:"query=category"`
		Status   *string `in:"query=status" validate:"omitempty,oneof=active inactive"`
		Since    *string `in:"query=since"`
		Until    *string `in:"query=until"`
		Sort     *string `in:"query=sort"`
		Order    *string `in:"query=order"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Validate pagination
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// Parse date range filters (optional)
	var since, until *time.Time
	if input.Since != nil {
		if t, err := time.Parse("2006-01-02", *input.Since); err == nil {
			since = &t
		}
	}

	if input.Until != nil {
		if t, err := time.Parse("2006-01-02", *input.Until); err == nil {
			until = &t
		}
	}

	// Call berita service
	serviceInput := berita.ListBeritaInput{
		Query:    input.Q,
		Category: input.Category,
		Status:   input.Status,
		Since:    since,
		Until:    until,
		Sort:     "created_at", // default
		Order:    "desc",       // default
		Page:     input.Page,
		Limit:    input.Limit,
	}
	if input.Sort != nil && *input.Sort != "" {
		serviceInput.Sort = *input.Sort
	}
	if input.Order != nil && *input.Order != "" {
		serviceInput.Order = *input.Order
	}

	beritaList, paginationResult, err := h.beritaService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list berita", err)
		return
	}

	// Build response
	beritaResponses := make([]BeritaListResponse, len(beritaList))
	for i, b := range beritaList {
		beritaResponses[i] = BeritaListResponse{
			ID:         b.ID,
			Title:      b.Title,
			CategoryID: ptrStr(b.Category),
			Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
			Media:      h.buildBeritaMedia(b.ImageMediaID, false),
			Status:     b.Status,
			CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"berita":     beritaResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetBerita godoc
// @Summary      Get news by ID (admin)
// @Description  RBAC: berita:read.
// @Tags         berita
// @Produce      json
// @Param        id         path     string true   "Berita UUID"
// @Success      200  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/{id} [get]
func (h *BeritaHandler) GetBerita(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid berita ID format")
		return
	}

	// Call berita service
	b, err := h.beritaService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Berita not found")
		return
	}

	// Build response
	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    b.Content,
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		Media:      h.buildBeritaMedia(b.ImageMediaID, false),
		Status:     b.Status,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateBerita godoc
// @Summary      Update news (admin)
// @Description  Partial update — missing fields fall back to existing values. Optional image or image_media_id (mutually exclusive). Status is NOT changed here — use PATCH /berita/{id}/status. RBAC: berita:write.
// @Tags         berita
// @Accept       mpfd
// @Produce      json
// @Param        id         path     string true   "Berita UUID"
// @Param        image      formData file    false  "New cover image"
// @Param        image_media_id  formData  string  false  "Gallery media UUID (mutually exclusive with image)"
// @Success      200  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/{id} [put]
func (h *BeritaHandler) UpdateBerita(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}

	// Check if berita exists first (before validating input)
	existing, err := h.beritaService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Berita not found")
		return
	}

	var input struct {
		Title        string       `in:"form=title"`
		Content      string       `in:"form=content"`
		Category     string       `in:"form=category"`
		Image        *httpin.File `in:"form=image"`
		ImageMediaID *string      `in:"form=image_media_id"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	if input.Image != nil && input.ImageMediaID != nil && *input.ImageMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either image or image_media_id, not both")
		return
	}

	// Get form fields - use existing values as defaults
	title := input.Title
	if title == "" {
		title = existing.Title
	}

	content := input.Content
	if content == "" {
		content = existing.Content
	}

	category := input.Category
	if category == "" {
		category = existing.Category
	}

	// Get image file (optional for update)
	var beritaInput berita.UpdateBeritaInput
	if input.Image != nil {
		imageFile, err := input.Image.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read image file")
			return
		}
		contentType := input.Image.MIMEHeader().Get("Content-Type")
		if contentType == "" || contentType == "application/octet-stream" {
			contentType = handlerutil.InferContentType(input.Image.Filename())
		}
		beritaInput = berita.UpdateBeritaInput{
			Title:       title,
			Content:     content,
			Category:    category,
			ImageFile:   imageFile,
			ImageName:   input.Image.Filename(),
			ImageSize:   input.Image.Size(),
			ContentType: contentType,
		}
	} else {
		// No new image file provided; media ref (if any) is passed through
		beritaInput = berita.UpdateBeritaInput{
			Title:        title,
			Content:      content,
			Category:     category,
			ImageFile:    nil,
			ImageMediaID: input.ImageMediaID,
		}
	}

	// Call berita service
	b, err := h.beritaService.Update(r.Context(), id, beritaInput)
	if err != nil {
		if strings.Contains(err.Error(), "berita not found") {
			response.Error(w, http.StatusNotFound, "Berita not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update berita", err)
		return
	}

	// Build response
	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    b.Content,
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		Media:      h.buildBeritaMedia(b.ImageMediaID, false),
		Status:     b.Status,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateBeritaStatus godoc
// @Summary      Update news status (admin)
// @Description  Toggle a news article between active and inactive. Inactive articles are hidden from the public endpoints. RBAC: berita:write.
// @Tags         berita
// @Accept       json
// @Produce      json
// @Param        id       path  string  true  "Berita UUID"
// @Param        request  body  map[string]interface{}  true  "New status (active|inactive)"
// @Success      200  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/{id}/status [patch]
func (h *BeritaHandler) UpdateBeritaStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}

	var req struct {
		Payload struct {
			Status string `json:"status" validate:"required,oneof=active inactive"`
		} `in:"body=json"`
	}
	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call berita service
	b, err := h.beritaService.UpdateStatus(r.Context(), id, req.Payload.Status)
	if err != nil {
		if strings.Contains(err.Error(), "berita not found") {
			response.Error(w, http.StatusNotFound, "Berita not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update berita status", err)
		return
	}

	// Build response
	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    b.Content,
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		Media:      h.buildBeritaMedia(b.ImageMediaID, false),
		Status:     b.Status,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeleteBerita godoc
// @Summary      Delete news (admin)
// @Description  Deletes article and its cover image. RBAC: berita:write.
// @Tags         berita
// @Produce      json
// @Param        id         path     string true   "Berita UUID"
// @Success      200  {object} response.MessageResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /berita/{id} [delete]
func (h *BeritaHandler) DeleteBerita(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}

	// Call berita service
	err := h.beritaService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "berita not found") {
			response.Error(w, http.StatusNotFound, "Berita not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete berita", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Berita deleted successfully",
	})
}

// ListBeritaPublic godoc
// @Summary      List news (public)
// @Description  Public paginated list without content body. Only active articles are returned.
// @Tags         berita
// @Produce      json
//
//	@Param	page	query	int	false	"Page default(1)"
//	@Param	limit	query	int	false	"Page size default(10)"
//	@Param	sort	query	string	false	"Sort field: created_at or title (default created_at)"
//	@Param	order	query	string	false	"Sort direction: asc or desc (default desc - newest first)"
//
// @Success      200  {object} berita.BeritaPaginatedResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Router       /public/berita/list [get]
func (h *BeritaHandler) ListBeritaPublic(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page  int     `in:"query=page;default=1"`
		Limit int     `in:"query=limit;default=10"`
		Sort  *string `in:"query=sort"`
		Order *string `in:"query=order"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if input.Page <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid page parameter: must be a positive integer")
		return
	}

	if input.Limit <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid limit parameter: must be a positive integer")
		return
	}

	serviceInput := berita.ListBeritaInput{
		Page:   input.Page,
		Limit:  input.Limit,
		Status: ptrStr(domainberita.StatusActive),
		Sort:   "created_at",
		Order:  "desc",
	}
	if input.Sort != nil && *input.Sort != "" {
		serviceInput.Sort = *input.Sort
	}
	if input.Order != nil && *input.Order != "" {
		serviceInput.Order = *input.Order
	}

	beritaList, paginationResult, err := h.beritaService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list berita", err)
		return
	}

	beritaResponses := make([]BeritaListResponse, len(beritaList))
	for i, b := range beritaList {
		beritaResponses[i] = BeritaListResponse{
			ID:         b.ID,
			Title:      b.Title,
			CategoryID: ptrStr(b.Category),
			Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
			Media:      h.buildBeritaMedia(b.ImageMediaID, true),
			Status:     b.Status,
			CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"berita":     beritaResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetBeritaPublic godoc
// @Summary      Get news (public)
// @Description  Public full detail. Inactive articles return 404.
// @Tags         berita
// @Produce      json
// @Param        id         path     string true   "Berita UUID"
// @Success      200  {object} berita.BeritaResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Failure      409  {object} response.ErrorResponse
// @Failure      429  {object} response.ErrorResponse
// @Router       /public/berita/{id} [get]
func (h *BeritaHandler) GetBeritaPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Berita ID is required")
		return
	}

	b, err := h.beritaService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "berita not found") {
			response.Error(w, http.StatusNotFound, "Berita not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get berita", err)
		return
	}

	// Inactive articles are unpublished; hide them from the public.
	if b.Status != domainberita.StatusActive {
		response.Error(w, http.StatusNotFound, "Berita not found")
		return
	}

	resp := BeritaResponse{
		ID:         b.ID,
		Title:      b.Title,
		Content:    rewriteEmbeddedMediaURLs(b.Content),
		CategoryID: ptrStr(b.Category),
		Category:   buildCategoryInfo(ptrStr(b.Category), b.CategoryName),
		Media:      h.buildBeritaMedia(b.ImageMediaID, true),
		Status:     b.Status,
		CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// rewriteEmbeddedMediaURLs rewrites admin gallery binary paths embedded in
// Quill delta content to the feature-scoped public route so anonymous
// visitors can render berita images.
func rewriteEmbeddedMediaURLs(content string) string {
	return strings.ReplaceAll(content, "/api/v1/gallery/media/", "/api/v1/public/berita/media/")
}
