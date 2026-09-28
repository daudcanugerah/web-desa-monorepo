package banner

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	domainbanner "webdesa/api/domain/banner"
	"webdesa/api/pkg/response"
	"webdesa/api/usecase/banner"
	galleryusecase "webdesa/api/usecase/gallery"

	"webdesa/api/pkg/handlerutil"
)

// BannerHandler handles HTTP requests for banner management operations.
// It accepts the banner service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain.
//
// The signedURL fields are optional (Task 7.1.5). When signedURLsEnabled
// is true and signedURL is non-nil, response mappers emit
// /api/v1/media/{id}/...?jwt= URLs (CDN-friendly, time-bounded) instead of
// the legacy per-scope paths. When false, the handler behaves exactly as
// before — same URLs, same RBAC, no client change required.
type BannerHandler struct {
	bannerService     *banner.Service
	signedURL         *galleryusecase.SignedURLService
	signedURLsEnabled bool
}

// NewBannerHandler creates a new banner handler with service dependency injected.
func NewBannerHandler(bannerService *banner.Service, signedURL *galleryusecase.SignedURLService, signedURLsEnabled bool) *BannerHandler {
	return &BannerHandler{
		bannerService:     bannerService,
		signedURL:         signedURL,
		signedURLsEnabled: signedURLsEnabled,
	}
}

// resolveBannerImageURL returns the banner image URL: the gallery binary
// URL when the row carries a media id.
func resolveBannerImageURL(b *domainbanner.Banner, public bool) string {
	if b.ImageMediaID != nil && *b.ImageMediaID != "" {
		if public {
			return galleryusecase.FeatureURLFor(galleryusecase.FeatureBanner, "content", *b.ImageMediaID)
		}
		return galleryusecase.URLFor(galleryusecase.URLScopeAdmin, "content", *b.ImageMediaID)
	}
	return ""
}

// buildBannerMedia emits the shared media shape (Task 4.3): media_id plus
// content/thumbnail URLs for gallery rows, or just the legacy URL.
//
// When the handler has a signed-URL service and signed URLs are enabled
// (Task 7.1.5), the URLs embed ?jwt=... so the unified
// /api/v1/media/{id}/...?jwt= route serves the binary — CDN-friendly,
// time-bounded, scope-bound. Otherwise the legacy FeatureURLFor / URLFor
// paths are emitted.
func (h *BannerHandler) buildBannerMedia(b *domainbanner.Banner, public bool) *response.MediaInfo {
	if b.ImageMediaID == nil || *b.ImageMediaID == "" {
		return nil
	}
	id := *b.ImageMediaID
	scope := galleryusecase.ScopePublic
	sub := "anonymous"
	if !public {
		scope = galleryusecase.ScopeAdmin
		sub = "user:admin" // admin sub gets tightened in 7.3 to use real user id
	}

	if h.signedURLsEnabled && h.signedURL != nil {
		url := galleryusecase.SignedURLPath("content", id)
		thumb := galleryusecase.SignedURLPath("thumbnail", id)
		url += h.signedURL.SignedURLQuery(scope, id, sub, 0)
		thumb += h.signedURL.SignedURLQuery(scope, id, sub, 0)
		return &response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb}
	}

	var url, thumb string
	if public {
		url = galleryusecase.FeatureURLFor(galleryusecase.FeatureBanner, "content", id)
		thumb = galleryusecase.FeatureURLFor(galleryusecase.FeatureBanner, "thumbnail", id)
	} else {
		url = galleryusecase.URLFor(galleryusecase.URLScopeAdmin, "content", id)
		thumb = galleryusecase.URLFor(galleryusecase.URLScopeAdmin, "thumbnail", id)
	}
	return &response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nilIfEmpty returns nil when the input string is empty, otherwise returns
// a pointer to the string. Used for optional form fields that map to
// *string in the usecase input.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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

// BannerResponse represents the response for banner data
type BannerResponse struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Link        string                 `json:"link"`
	Media       *response.MediaInfo    `json:"media,omitempty"`
	Status      string                 `json:"status"`
	CategoryID  *string                `json:"category_id,omitempty"`
	Category    *response.CategoryInfo `json:"category,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   string                 `json:"created_at"` // ISO 8601 format
	UpdatedAt   string                 `json:"updated_at"` // ISO 8601 format
}

// BannerListResponse represents a paginated list of banners
type BannerListResponse struct {
	Banners    []BannerResponse       `json:"banners"`
	Pagination map[string]interface{} `json:"pagination"`
}

// CreateBanner godoc
// @Summary      Create banner (admin)
// @Description  Create a banner. Multipart form: title (required), description, link, image or image_media_id (required, mutually exclusive), category (UUID, optional), metadata (JSON string), status (active|inactive, optional; defaults to inactive). RBAC: banners:write.
// @Tags         banner
// @Accept       mpfd
// @Produce      json
// @Param        title       formData  string  true  "Title"
// @Param        description formData  string  false "Description"
// @Param        link        formData  string  false "Click URL"
// @Param        image       formData  file    true  "Banner image"
// @Param        image_media_id formData string false "Gallery media UUID (mutually exclusive with image)"
// @Param        category    formData  string  false "Category UUID"
// @Param        metadata    formData  string  false "Metadata as JSON string"
// @Param        status      formData  string  false "Status: active or inactive (default inactive)" Enums(active, inactive)
// @Success      201         {object} banner.BannerResponse
// @Failure      400         {object} response.ErrorResponse
// @Failure      401         {object} response.ErrorResponse
// @Failure      403         {object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /banners [post]
//
// Validates: Requirements 7.3, 16.1, 16.2
func (h *BannerHandler) CreateBanner(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string       `in:"form=title" validate:"required,min=1"`
		Description  string       `in:"form=description"`
		Link         string       `in:"form=link"`
		Image        *httpin.File `in:"form=image"`
		ImageMediaID *string      `in:"form=image_media_id"`
		Category     string       `in:"form=category"`
		MetadataStr  string       `in:"form=metadata"`
		Status       *string      `in:"form=status" validate:"omitempty,oneof=active inactive"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Validate required fields
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if input.Image == nil && (input.ImageMediaID == nil || *input.ImageMediaID == "") {
		response.Error(w, http.StatusBadRequest, "Image file or image_media_id is required")
		return
	}
	if input.Image != nil && input.ImageMediaID != nil && *input.ImageMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either image or image_media_id, not both")
		return
	}

	// Parse metadata JSON if provided
	var metadata map[string]interface{}
	if input.MetadataStr != "" {
		if err := json.Unmarshal([]byte(input.MetadataStr), &metadata); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid metadata JSON format")
			return
		}
	}

	// Open the file for reading (media ref path skips this)
	var imageFile io.ReadCloser
	var imageName string
	var imageSize int64
	var contentType string
	if input.Image != nil {
		// Infer content type from filename if not provided
		contentType = input.Image.MIMEHeader().Get("Content-Type")
		if contentType == "" || contentType == "application/octet-stream" {
			contentType = handlerutil.InferContentType(input.Image.Filename())
		}
		var openErr error
		imageFile, openErr = input.Image.Open()
		if openErr != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read image file")
			return
		}
		defer imageFile.Close()
		imageName = input.Image.Filename()
		imageSize = input.Image.Size()
	}

	// Call banner service
	bannerInput := banner.CreateBannerInput{
		Title:        input.Title,
		Description:  input.Description,
		Link:         input.Link,
		ImageFile:    imageFile,
		ImageName:    imageName,
		ImageSize:    imageSize,
		ContentType:  contentType,
		ImageMediaID: input.ImageMediaID,
		Category:     nilIfEmpty(input.Category),
		Metadata:     metadata,
		Status:       input.Status,
	}

	b, err := h.bannerService.Create(r.Context(), bannerInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid image") || strings.Contains(err.Error(), "file size exceeds") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to create banner", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create banner", err)
		return
	}

	// Build response
	resp := BannerResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		Link:        b.Link,
		Media:       h.buildBannerMedia(b, false),
		Status:      b.Status,
		CategoryID:  b.Category,
		Category:    buildCategoryInfo(b.Category, b.CategoryName),
		Metadata:    b.Metadata,
		CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListBanners godoc
// @Summary      List banners (admin)
// @Description  Paginated list of banners, optionally filtered by status, q (title/description), or category. RBAC: banners:read.
// @Tags         banner
// @Produce      json
// @Param        page     query     int    false  "Page number"   default(1)
// @Param        limit    query     int    false  "Page size"     default(10)
// @Param        status   query     string false  "Filter by status"  Enums(active, inactive)
// @Param        q        query     string false  "Search across title and description"
// @Param        category query     string false  "Filter by category UUID"
// @Success      200      {object} banner.BannerListResponse
// @Failure      401      {object} response.ErrorResponse
// @Failure      403      {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /banners [get]
//
// Validates: Requirements 7.1, 17.2, 16.1, 16.2
func (h *BannerHandler) ListBanners(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int     `in:"query=page;default=1"`
		Limit    int     `in:"query=limit;default=10"`
		Status   *string `in:"query=status"`
		Q        *string `in:"query=q"`
		Category *string `in:"query=category"`
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

	// Call banner service
	serviceInput := banner.ListBannersInput{
		Status:   input.Status,
		Query:    input.Q,
		Category: input.Category,
		Page:     input.Page,
		Limit:    input.Limit,
	}

	banners, paginationResult, err := h.bannerService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list banners", err)
		return
	}

	// Build response
	bannerResponses := make([]BannerResponse, len(banners))
	for i, b := range banners {
		bannerResponses[i] = BannerResponse{
			ID:          b.ID,
			Title:       b.Title,
			Description: b.Description,
			Link:        b.Link,
			Media:       h.buildBannerMedia(b, false),
			Status:      b.Status,
			CategoryID:  b.Category,
			Category:    buildCategoryInfo(b.Category, b.CategoryName),
			Metadata:    b.Metadata,
			CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"banners":    bannerResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetActiveBanners godoc
// @Summary      List active banners
// @Description  Public endpoint returning up to 100 active banners.
// @Tags         banner
// @Produce      json
// @Success      200  {object} []banner.BannerResponse
// @Router       /banners/active [get]
//
// Validates: Requirements 7.7, 16.1, 16.2
func (h *BannerHandler) GetActiveBanners(w http.ResponseWriter, r *http.Request) {
	// Call banner service with active status filter
	activeStatus := "active"
	input := banner.ListBannersInput{
		Status: &activeStatus,
		Page:   1,
		Limit:  100, // Get all active banners (max 20 per requirement)
	}

	banners, _, err := h.bannerService.List(r.Context(), input)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list active banners", err)
		return
	}

	// Build response
	bannerResponses := make([]BannerResponse, len(banners))
	for i, b := range banners {
		bannerResponses[i] = BannerResponse{
			ID:          b.ID,
			Title:       b.Title,
			Description: b.Description,
			Link:        b.Link,
			Media:       h.buildBannerMedia(b, true),
			Status:      b.Status,
			CategoryID:  b.Category,
			Category:    buildCategoryInfo(b.Category, b.CategoryName),
			Metadata:    b.Metadata,
			CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response.Success(w, http.StatusOK, bannerResponses)
}

// GetBanner godoc
// @Summary      Get banner by ID (admin)
// @Description  Retrieve a single banner by UUID. RBAC: banners:read.
// @Tags         banner
// @Produce      json
// @Param        id  path  string  true  "Banner UUID"
// @Success      200  {object} banner.BannerResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /banners/{id} [get]
//
// Validates: Requirements 7.2, 16.1, 16.2
func (h *BannerHandler) GetBanner(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Banner ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid banner ID format")
		return
	}

	// Call banner service
	b, err := h.bannerService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Banner not found")
		return
	}

	// Build response
	resp := BannerResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		Link:        b.Link,
		Media:       h.buildBannerMedia(b, false),
		Status:      b.Status,
		CategoryID:  b.Category,
		Category:    buildCategoryInfo(b.Category, b.CategoryName),
		Metadata:    b.Metadata,
		CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateBanner godoc
// @Summary      Update banner (admin)
// @Description  Update banner fields. Image or image_media_id is optional (mutually exclusive). Category is optional (UUID). Status is NOT changed here — use PATCH /status. RBAC: banners:write.
// @Tags         banner
// @Accept       mpfd
// @Produce      json
// @Param        id    path  string  true  "Banner UUID"
// @Param        title       formData  string  false "Title"
// @Param        description formData  string  false "Description"
// @Param        link        formData  string  false "Click URL"
// @Param        image       formData  file    false "New banner image"
// @Param        image_media_id formData string false "Gallery media UUID (mutually exclusive with image)"
// @Param        category    formData  string  false "Category UUID"
// @Param        metadata    formData  string  false "Metadata as JSON string"
// @Success      200         {object} banner.BannerResponse
// @Failure      400         {object} response.ErrorResponse
// @Failure      401         {object} response.ErrorResponse
// @Failure      403         {object} response.ErrorResponse
// @Failure      404         {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /banners/{id} [put]
//
// Validates: Requirements 7.4, 16.1, 16.2
func (h *BannerHandler) UpdateBanner(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Banner ID is required")
		return
	}

	var input struct {
		Title        string       `in:"form=title" validate:"required,min=1"`
		Description  string       `in:"form=description"`
		Link         string       `in:"form=link"`
		Image        *httpin.File `in:"form=image"`
		ImageMediaID *string      `in:"form=image_media_id"`
		Category     string       `in:"form=category"`
		MetadataStr  string       `in:"form=metadata"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Validate required fields
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if input.Image != nil && input.ImageMediaID != nil && *input.ImageMediaID != "" {
		response.Error(w, http.StatusBadRequest, "Provide either image or image_media_id, not both")
		return
	}

	// Parse metadata JSON if provided
	var metadata map[string]interface{}
	if input.MetadataStr != "" {
		if err := json.Unmarshal([]byte(input.MetadataStr), &metadata); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid metadata JSON format")
			return
		}
	}

	// Get image file (optional for update)
	var bannerInput banner.UpdateBannerInput
	if input.Image != nil {
		contentType := input.Image.MIMEHeader().Get("Content-Type")
		if contentType == "" || contentType == "application/octet-stream" {
			contentType = handlerutil.InferContentType(input.Image.Filename())
		}
		imageFile, err := input.Image.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read image file")
			return
		}
		defer imageFile.Close()
		bannerInput = banner.UpdateBannerInput{
			Title:       input.Title,
			Description: input.Description,
			Link:        input.Link,
			ImageFile:   imageFile,
			ImageName:   input.Image.Filename(),
			ImageSize:   input.Image.Size(),
			ContentType: contentType,
			Category:    nilIfEmpty(input.Category),
			Metadata:    metadata,
		}
	} else {
		// No new image file provided; media ref (if any) is passed through
		bannerInput = banner.UpdateBannerInput{
			Title:        input.Title,
			Description:  input.Description,
			Link:         input.Link,
			ImageFile:    nil,
			ImageMediaID: input.ImageMediaID,
			Category:     nilIfEmpty(input.Category),
			Metadata:     metadata,
		}
	}

	// Call banner service
	b, err := h.bannerService.Update(r.Context(), id, bannerInput)
	if err != nil {
		if strings.Contains(err.Error(), "banner not found") {
			response.Error(w, http.StatusNotFound, "Banner not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update banner", err)
		return
	}

	// Build response
	resp := BannerResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		Link:        b.Link,
		Media:       h.buildBannerMedia(b, false),
		Status:      b.Status,
		CategoryID:  b.Category,
		Category:    buildCategoryInfo(b.Category, b.CategoryName),
		Metadata:    b.Metadata,
		CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateBannerStatus godoc
// @Summary      Update banner status (admin)
// @Description  Toggle a banner between active and inactive. Enforces a maximum of 20 active banners. RBAC: banners:write.
// @Tags         banner
// @Accept       json
// @Produce      json
// @Param        id       path  string  true  "Banner UUID"
// @Param        request    body      map[string]interface{}  true  "New status (active|inactive)"
// @Success      200      {object} banner.BannerResponse
// @Failure      400      {object} response.ErrorResponse "Max 20 active banners reached"
// @Failure      401      {object} response.ErrorResponse
// @Failure      403      {object} response.ErrorResponse
// @Failure      404      {object} response.ErrorResponse
// @ Security     BearerAuth
// @Router       /banners/{id}/status [patch]
//
// Validates: Requirements 7.5, 16.1, 16.2
func (h *BannerHandler) UpdateBannerStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Banner ID is required")
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

	// Validate input
	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call banner service
	b, err := h.bannerService.UpdateStatus(r.Context(), id, req.Payload.Status)
	if err != nil {
		if strings.Contains(err.Error(), "banner not found") {
			response.Error(w, http.StatusNotFound, "Banner not found")
			return
		}
		if err.Error() == "cannot activate banner: maximum of 20 active banners reached" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update banner status", err)
		return
	}

	// Build response
	resp := BannerResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		Link:        b.Link,
		Media:       h.buildBannerMedia(b, false),
		Status:      b.Status,
		CategoryID:  b.Category,
		Category:    buildCategoryInfo(b.Category, b.CategoryName),
		Metadata:    b.Metadata,
		CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeleteBanner godoc
// @Summary      Delete banner (admin)
// @Description  Delete a banner by UUID and remove its image file. RBAC: banners:write.
// @Tags         banner
// @Produce      json
// @Param        id  path  string  true  "Banner UUID"
// @Success      204  {object} response.MessageResponse
// @Failure      400  {object} response.ErrorResponse
// @Failure      401  {object} response.ErrorResponse
// @Failure      403  {object} response.ErrorResponse
// @Failure      404  {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /banners/{id} [delete]
//
// Validates: Requirements 7.6, 16.1, 16.2
func (h *BannerHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Banner ID is required")
		return
	}

	// Call banner service
	err := h.bannerService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "banner not found") {
			response.Error(w, http.StatusNotFound, "Banner not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete banner", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Banner deleted successfully",
	})
}

// ListBannersPublic godoc
// @Summary      List banners (public)
// @Description  Public endpoint to list all banners, optionally filtered by q (title/description) or category.
// @Tags         banner
// @Produce      json
// @Param        page     query  int    false  "Page number"  default(1)
// @Param        limit    query  int    false  "Page size"    default(10)
// @Param        q        query  string false  "Search across title and description"
// @Param        category query  string false  "Filter by category UUID"
// @Success      200      {object} banner.BannerListResponse
// @Router       /public/banner [get]
func (h *BannerHandler) ListBannersPublic(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int     `in:"query=page;default=1"`
		Limit    int     `in:"query=limit;default=10"`
		Q        *string `in:"query=q"`
		Category *string `in:"query=category"`
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

	serviceInput := banner.ListBannersInput{
		Query:    input.Q,
		Category: input.Category,
		Page:     input.Page,
		Limit:    input.Limit,
	}

	banners, paginationResult, err := h.bannerService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list banners", err)
		return
	}

	bannerResponses := make([]BannerResponse, len(banners))
	for i, b := range banners {
		bannerResponses[i] = BannerResponse{
			ID:          b.ID,
			Title:       b.Title,
			Description: b.Description,
			Link:        b.Link,
			Media:       h.buildBannerMedia(b, true),
			Status:      b.Status,
			CategoryID:  b.Category,
			Category:    buildCategoryInfo(b.Category, b.CategoryName),
			Metadata:    b.Metadata,
			CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"banners":    bannerResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetBannerPublic godoc
// @Summary      Get banner (public)
// @Description  Public endpoint to retrieve a single banner.
// @Tags         banner
// @Produce      json
// @Param        id  path  string  true  "Banner UUID"
// @Success      200  {object} banner.BannerResponse
// @Failure      404  {object} response.ErrorResponse
// @Router       /public/banner/{id} [get]
func (h *BannerHandler) GetBannerPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Banner ID is required")
		return
	}

	b, err := h.bannerService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "banner not found") {
			response.Error(w, http.StatusNotFound, "Banner not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get banner", err)
		return
	}

	resp := BannerResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		Link:        b.Link,
		Media:       h.buildBannerMedia(b, true),
		Status:      b.Status,
		CategoryID:  b.Category,
		Category:    buildCategoryInfo(b.Category, b.CategoryName),
		Metadata:    b.Metadata,
		CreatedAt:   b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}
