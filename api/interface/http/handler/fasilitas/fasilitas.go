package fasilitas

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	domainfasilitas "webdesa/api/domain/fasilitas"
	"webdesa/api/pkg/response"
	"webdesa/api/usecase/fasilitas"
	galleryuc "webdesa/api/usecase/gallery"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	"webdesa/api/pkg/handlerutil"
)

// FasilitasHandler handles HTTP requests for Fasilitas (facility) management operations.
// It accepts the Fasilitas service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
type FasilitasHandler struct {
	fasilitasService  *fasilitas.Service
	signedURL         *galleryuc.SignedURLService
	signedURLsEnabled bool
}

// NewFasilitasHandler creates a new Fasilitas handler with service dependency injected.
func NewFasilitasHandler(fasilitasService *fasilitas.Service, signedURL *galleryuc.SignedURLService, signedURLsEnabled bool) *FasilitasHandler {
	return &FasilitasHandler{
		fasilitasService:  fasilitasService,
		signedURL:         signedURL,
		signedURLsEnabled: signedURLsEnabled,
	}
}

// FasilitasResponse represents the response for Fasilitas data
type FasilitasResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	CategoryID  *string                `json:"category_id,omitempty"`
	Category    *response.CategoryInfo `json:"category,omitempty"`
	Latitude    float64                `json:"latitude"`
	Longitude   float64                `json:"longitude"`
	Description *string                `json:"description,omitempty"`
	Media       []response.MediaInfo   `json:"media"`
	CreatedAt   string                 `json:"created_at"` // ISO 8601 format
	UpdatedAt   string                 `json:"updated_at"` // ISO 8601 format
}

// buildCategoryInfo builds a CategoryInfo from a nullable category ID pointer (Fasilitas.Category is *string).
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

// mediaFor builds the shared media shape (Task 4.3): one entry per gallery
// media id with content/thumbnail URLs.
func (h *FasilitasHandler) mediaFor(f *domainfasilitas.Fasilitas, public bool) []response.MediaInfo {
	out := make([]response.MediaInfo, 0, len(f.ImagesMediaIDs))
	scope := galleryuc.ScopePublic
	sub := "anonymous"
	if !public {
		scope = galleryuc.ScopeAdmin
		sub = "user:admin"
	}
	for _, id := range f.ImagesMediaIDs {
		var url, thumb string
		if h.signedURLsEnabled && h.signedURL != nil {
			url = galleryuc.SignedURLPath("content", id) + h.signedURL.SignedURLQuery(scope, id, sub, 0)
			thumb = galleryuc.SignedURLPath("thumbnail", id) + h.signedURL.SignedURLQuery(scope, id, sub, 0)
		} else if public {
			url = galleryuc.FeatureURLFor(galleryuc.FeatureFasilitas, "content", id)
			thumb = galleryuc.FeatureURLFor(galleryuc.FeatureFasilitas, "thumbnail", id)
		} else {
			url = galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", id)
			thumb = galleryuc.URLFor(galleryuc.URLScopeAdmin, "thumbnail", id)
		}
		out = append(out, response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb})
	}
	return out
}

// FasilitasPaginatedResponse represents a paginated list of fasilitas.
type FasilitasPaginatedResponse struct {
	Fasilitas  []FasilitasResponse    `json:"fasilitas"`
	Pagination map[string]interface{} `json:"pagination"`
}

// CreateFasilitas godoc
// @Summary      Create fasilitas (admin)
// @Description  Multipart or JSON. Required: name, latitude (-90..90), longitude (-180..180). Optional: category (UUID), description. Images must be pre-uploaded via POST /fasilitas/upload-media and referenced with images_media_ids. RBAC: fasilitas:write.
// @Tags         fasilitas
// @Accept       mpfd
// @Produce      json
// @Param        name             formData string   true   "Name"
// @Param        latitude         formData number   true   "Latitude (-90..90)"
// @Param        longitude        formData number   true   "Longitude (-180..180)"
// @Param        category         formData string   false  "Category UUID"
// @Param        description      formData string   false  "Description"
// @Param        images_media_ids formData []string false  "Pre-uploaded media ids (repeatable)"
// @Success      201  {object} fasilitas.FasilitasResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		413	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas [post]
func (h *FasilitasHandler) CreateFasilitas(w http.ResponseWriter, r *http.Request) {
	// Determine if request is form data or JSON
	contentType := r.Header.Get("Content-Type")
	var name string
	var description *string
	var categoryVal *string
	var latitude, longitude float64
	var imagesMediaIDs []string

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart/form-data
		var req struct {
			Name           string   `in:"form=name" validate:"required,min=1"`
			Category       *string  `in:"form=category"`
			Latitude       string   `in:"form=latitude" validate:"required"`
			Longitude      string   `in:"form=longitude" validate:"required"`
			Description    *string  `in:"form=description"`
			ImagesMediaIDs []string `in:"form=images_media_ids"`
		}

		if err := httpin.DecodeTo(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Parse latitude and longitude
		lat, err := strconv.ParseFloat(req.Latitude, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid latitude format")
			return
		}
		lon, err := strconv.ParseFloat(req.Longitude, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid longitude format")
			return
		}

		name = req.Name
		categoryVal = req.Category
		latitude = lat
		longitude = lon
		if req.Description != nil && *req.Description != "" {
			description = req.Description
		}
		imagesMediaIDs = req.ImagesMediaIDs
	} else {
		// Handle JSON body
		var req struct {
			Payload struct {
				Name           string   `json:"name" validate:"required,min=1"`
				Category       *string  `json:"category"`
				Latitude       float64  `json:"latitude" validate:"required"`
				Longitude      float64  `json:"longitude" validate:"required"`
				Description    *string  `json:"description"`
				ImagesMediaIDs []string `json:"images_media_ids"`
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

		name = req.Payload.Name
		categoryVal = req.Payload.Category
		latitude = req.Payload.Latitude
		longitude = req.Payload.Longitude
		if req.Payload.Description != nil && *req.Payload.Description != "" {
			description = req.Payload.Description
		}
		imagesMediaIDs = req.Payload.ImagesMediaIDs
	}

	serviceInput := fasilitas.CreateFasilitasInput{
		Name:           name,
		Category:       categoryVal,
		Latitude:       latitude,
		Longitude:      longitude,
		Description:    description,
		ImagesMediaIDs: imagesMediaIDs,
	}

	f, err := h.fasilitasService.Create(r.Context(), serviceInput)
	if err != nil {
		if respondIfBulkLimit(w, err) {
			return
		}
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to create fasilitas", err)
		return
	}

	// Build response
	resp := FasilitasResponse{
		ID:          f.ID,
		Name:        f.Name,
		CategoryID:  f.Category,
		Category:    buildCategoryInfo(f.Category, f.CategoryName),
		Latitude:    f.Latitude,
		Longitude:   f.Longitude,
		Description: f.Description,
		Media:       h.mediaFor(f, false),
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListFasilitas godoc
// @Summary      List fasilitas (admin)
// @Description  Optional bounding box filter (minLat,maxLat,minLon,maxLon), q (name search), or category (UUID). RBAC: fasilitas:read.
// @Tags         fasilitas
// @Produce      json
//
//	@Param	minLat	query	number	false	"Min latitude"
//	@Param	maxLat	query	number	false	"Max latitude"
//	@Param	minLon	query	number	false	"Min longitude"
//	@Param	maxLon	query	number	false	"Max longitude"
//	@Param	q		query	string	false	"Search by name"
//	@Param	category	query	string	false	"Filter by category UUID"
//	@Param	page	query	int	false	"Page"
//	@Param	limit	query	int	false	"Page size"
//
// @Success      200  {object} fasilitas.FasilitasPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas [get]
func (h *FasilitasHandler) ListFasilitas(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int     `in:"query=page;default=1" validate:"min=1"`
		Limit    int     `in:"query=limit;default=10" validate:"min=1"`
		MinLat   *string `in:"query=minLat"`
		MaxLat   *string `in:"query=maxLat"`
		MinLon   *string `in:"query=minLon"`
		MaxLon   *string `in:"query=maxLon"`
		Q        *string `in:"query=q"`
		Category *string `in:"query=category"`
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

	// Parse bounding box parameters (optional)
	var bbox *fasilitas.BoundingBox

	// If any bbox parameter is provided, all must be provided
	if input.MinLat != nil || input.MaxLat != nil || input.MinLon != nil || input.MaxLon != nil {
		if input.MinLat == nil || input.MaxLat == nil || input.MinLon == nil || input.MaxLon == nil {
			response.Error(w, http.StatusBadRequest, "All bounding box parameters (minLat, maxLat, minLon, maxLon) must be provided")
			return
		}

		minLat, err := strconv.ParseFloat(*input.MinLat, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid minLat parameter")
			return
		}

		maxLat, err := strconv.ParseFloat(*input.MaxLat, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid maxLat parameter")
			return
		}

		minLon, err := strconv.ParseFloat(*input.MinLon, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid minLon parameter")
			return
		}

		maxLon, err := strconv.ParseFloat(*input.MaxLon, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid maxLon parameter")
			return
		}

		bbox = &fasilitas.BoundingBox{
			MinLat: minLat,
			MaxLat: maxLat,
			MinLon: minLon,
			MaxLon: maxLon,
		}
	}

	// Call Fasilitas service
	serviceInput := fasilitas.ListFasilitasInput{
		BBox:     bbox,
		Query:    input.Q,
		Category: input.Category,
		Page:     input.Page,
		Limit:    input.Limit,
	}

	fasilitasList, paginationResult, err := h.fasilitasService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list fasilitas", err)
		return
	}

	// Build response
	fasilitasResponses := make([]FasilitasResponse, len(fasilitasList))
	for i, f := range fasilitasList {
		fasilitasResponses[i] = FasilitasResponse{
			ID:          f.ID,
			Name:        f.Name,
			CategoryID:  f.Category,
			Category:    buildCategoryInfo(f.Category, f.CategoryName),
			Latitude:    f.Latitude,
			Longitude:   f.Longitude,
			Description: f.Description,
			Media:       h.mediaFor(f, false),
			CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"fasilitas":  fasilitasResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetFasilitas godoc
// @Summary      Get fasilitas by ID (admin)
// @Description  RBAC: fasilitas:read.
// @Tags         fasilitas
// @Produce      json
// @Param        id         path     string true   "Fasilitas UUID"
// @Success      200  {object} fasilitas.FasilitasResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas/{id} [get]
func (h *FasilitasHandler) GetFasilitas(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Fasilitas ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid fasilitas ID format")
		return
	}

	// Call Fasilitas service
	f, err := h.fasilitasService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Fasilitas not found")
		return
	}

	// Build response
	resp := FasilitasResponse{
		ID:          f.ID,
		Name:        f.Name,
		CategoryID:  f.Category,
		Category:    buildCategoryInfo(f.Category, f.CategoryName),
		Latitude:    f.Latitude,
		Longitude:   f.Longitude,
		Description: f.Description,
		Media:       h.mediaFor(f, false),
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateFasilitas godoc
// @Summary      Update fasilitas (admin)
// @Description  RBAC: fasilitas:write.
// @Tags         fasilitas
// @Accept       mpfd
// @Produce      json
// @Param        id         path     string true   "Fasilitas UUID"
// @Success      200  {object} fasilitas.FasilitasResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		413	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas/{id} [put]
func (h *FasilitasHandler) UpdateFasilitas(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Fasilitas ID is required")
		return
	}

	contentType := r.Header.Get("Content-Type")
	var name string
	var description *string
	var categoryVal *string
	var latitude, longitude float64
	var imagesMediaIDs []string
	var imagesMediaIDsProvided bool

	if strings.Contains(contentType, "multipart/form-data") {
		var req struct {
			Name           string   `in:"form=name" validate:"required,min=1"`
			Category       *string  `in:"form=category"`
			Latitude       string   `in:"form=latitude" validate:"required"`
			Longitude      string   `in:"form=longitude" validate:"required"`
			Description    *string  `in:"form=description"`
			ImagesMediaIDs []string `in:"form=images_media_ids"`
		}
		if err := httpin.DecodeTo(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		lat, err := strconv.ParseFloat(req.Latitude, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid latitude format")
			return
		}
		lon, err := strconv.ParseFloat(req.Longitude, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid longitude format")
			return
		}
		name = req.Name
		categoryVal = req.Category
		latitude = lat
		longitude = lon
		if req.Description != nil && *req.Description != "" {
			description = req.Description
		}
		if req.ImagesMediaIDs != nil {
			imagesMediaIDs = req.ImagesMediaIDs
			imagesMediaIDsProvided = true
		}
	} else {
		var req struct {
			Payload struct {
				Name           string   `json:"name" validate:"required,min=1"`
				Category       *string  `json:"category"`
				Latitude       float64  `json:"latitude" validate:"required"`
				Longitude      float64  `json:"longitude" validate:"required"`
				Description    *string  `json:"description"`
				ImagesMediaIDs []string `json:"images_media_ids"`
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
		name = req.Payload.Name
		categoryVal = req.Payload.Category
		latitude = req.Payload.Latitude
		longitude = req.Payload.Longitude
		if req.Payload.Description != nil && *req.Payload.Description != "" {
			description = req.Payload.Description
		}
		if req.Payload.ImagesMediaIDs != nil {
			imagesMediaIDs = req.Payload.ImagesMediaIDs
			imagesMediaIDsProvided = true
		}
	}

	serviceInput := fasilitas.UpdateFasilitasInput{
		Name:        name,
		Category:    categoryVal,
		Latitude:    latitude,
		Longitude:   longitude,
		Description: description,
	}
	if imagesMediaIDsProvided {
		serviceInput.ImagesMediaIDs = imagesMediaIDs
	}

	f, err := h.fasilitasService.Update(r.Context(), id, serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "fasilitas not found") {
			response.Error(w, http.StatusNotFound, "Fasilitas not found")
			return
		}
		if respondIfBulkLimit(w, err) {
			return
		}
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to update fasilitas", err)
		return
	}

	// Build response
	resp := FasilitasResponse{
		ID:          f.ID,
		Name:        f.Name,
		CategoryID:  f.Category,
		Category:    buildCategoryInfo(f.Category, f.CategoryName),
		Latitude:    f.Latitude,
		Longitude:   f.Longitude,
		Description: f.Description,
		Media:       h.mediaFor(f, false),
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeleteFasilitas godoc
// @Summary      Delete fasilitas (admin)
// @Description  RBAC: fasilitas:write.
// @Tags         fasilitas
// @Produce      json
// @Param        id         path     string true   "Fasilitas UUID"
// @Success      200  {object} response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas/{id} [delete]
func (h *FasilitasHandler) DeleteFasilitas(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Fasilitas ID is required")
		return
	}

	// Call Fasilitas service
	err := h.fasilitasService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "fasilitas not found") {
			response.Error(w, http.StatusNotFound, "Fasilitas not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete fasilitas", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Fasilitas deleted successfully",
	})
}

// RemoveImage godoc
// @Summary      Remove one image from fasilitas (admin)
// @Description  RBAC: fasilitas:write.
// @Tags         fasilitas
// @Produce      json
// @Param        id         path     string true   "Fasilitas UUID"
// @Param        imageIndex path     int    true   "Zero-based index"
// @Success      200  {object} fasilitas.FasilitasResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /fasilitas/{id}/images/{imageIndex} [delete]
func (h *FasilitasHandler) RemoveImage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Fasilitas ID is required")
		return
	}

	imageIndexStr := chi.URLParam(r, "imageIndex")
	if imageIndexStr == "" {
		response.Error(w, http.StatusBadRequest, "Image index is required")
		return
	}

	imageIndex, err := strconv.Atoi(imageIndexStr)
	if err != nil || imageIndex < 0 {
		response.Error(w, http.StatusBadRequest, "Invalid image index: must be a non-negative integer")
		return
	}

	// Call Fasilitas service
	f, err := h.fasilitasService.RemoveImage(r.Context(), id, imageIndex)
	if err != nil {
		if strings.Contains(err.Error(), "fasilitas not found") {
			response.Error(w, http.StatusNotFound, "Fasilitas not found")
			return
		}
		if strings.Contains(err.Error(), "image index out of range") {
			response.Error(w, http.StatusBadRequest, "Image index out of range")
			return
		}
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to remove image", err)
		return
	}

	// Build response
	resp := FasilitasResponse{
		ID:          f.ID,
		Name:        f.Name,
		CategoryID:  f.Category,
		Category:    buildCategoryInfo(f.Category, f.CategoryName),
		Latitude:    f.Latitude,
		Longitude:   f.Longitude,
		Description: f.Description,
		Media:       h.mediaFor(f, false),
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// ListFasilitasPublic godoc
// @Summary      List fasilitas (public)
// @Description  Public paginated list with optional bbox, q (name search), or category (UUID).
// @Tags         fasilitas
// @Produce      json
//
//	@Param	minLat	query	number	false	"Min latitude"
//	@Param	maxLat	query	number	false	"Max latitude"
//	@Param	minLon	query	number	false	"Min longitude"
//	@Param	maxLon	query	number	false	"Max longitude"
//	@Param	q		query	string	false	"Search by name"
//	@Param	category	query	string	false	"Filter by category UUID"
//	@Param	page	query	int	false	"Page"
//	@Param	limit	query	int	false	"Page size"
//
// @Success      200  {object} fasilitas.FasilitasPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/fasilitas/list [get]
func (h *FasilitasHandler) ListFasilitasPublic(w http.ResponseWriter, r *http.Request) {
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

	serviceInput := fasilitas.ListFasilitasInput{
		Query:    input.Q,
		Category: input.Category,
		Page:     input.Page,
		Limit:    input.Limit,
	}

	fasilitasList, paginationResult, err := h.fasilitasService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list fasilitas", err)
		return
	}

	fasilitasResponses := make([]FasilitasResponse, len(fasilitasList))
	for i, f := range fasilitasList {
		fasilitasResponses[i] = FasilitasResponse{
			ID:          f.ID,
			Name:        f.Name,
			CategoryID:  f.Category,
			Category:    buildCategoryInfo(f.Category, f.CategoryName),
			Latitude:    f.Latitude,
			Longitude:   f.Longitude,
			Description: f.Description,
			Media:       h.mediaFor(f, true),
			CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"fasilitas":  fasilitasResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetFasilitasPublic godoc
// @Summary      Get fasilitas (public)
// @Description  Public full detail.
// @Tags         fasilitas
// @Produce      json
// @Param        id         path     string true   "Fasilitas UUID"
// @Success      200  {object} fasilitas.FasilitasResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/fasilitas/{id} [get]
func (h *FasilitasHandler) GetFasilitasPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Fasilitas ID is required")
		return
	}

	f, err := h.fasilitasService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "fasilitas not found") {
			response.Error(w, http.StatusNotFound, "Fasilitas not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get fasilitas", err)
		return
	}

	resp := FasilitasResponse{
		ID:          f.ID,
		Name:        f.Name,
		CategoryID:  f.Category,
		Category:    buildCategoryInfo(f.Category, f.CategoryName),
		Latitude:    f.Latitude,
		Longitude:   f.Longitude,
		Description: f.Description,
		Media:       h.mediaFor(f, true),
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// respondIfBulkLimit maps gallery bulk-cap violations (too many images per
// record, total size exceeded) to HTTP 413 Payload Too Large.
func respondIfBulkLimit(w http.ResponseWriter, err error) bool {
	if errors.Is(err, galleryuc.ErrBulkTooManyFiles) || errors.Is(err, galleryuc.ErrBulkTotalTooLarge) {
		response.Error(w, http.StatusRequestEntityTooLarge, err.Error())
		return true
	}
	return false
}
