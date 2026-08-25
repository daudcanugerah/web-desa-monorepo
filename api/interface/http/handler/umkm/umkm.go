package umkm

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	domainumkm "webdesa/api/domain/umkm"
	"webdesa/api/pkg/handlerutil"
	"webdesa/api/pkg/response"
	galleryuc "webdesa/api/usecase/gallery"
	"webdesa/api/usecase/umkm"
)

// UMKMHandler handles HTTP requests for UMKM (small business) management operations.
// It accepts the UMKM service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
//
// Task 7.1.6: signedURL fields mirror the banner handler.
type UMKMHandler struct {
	umkmService       *umkm.Service
	signedURL        *galleryuc.SignedURLService
	signedURLsEnabled bool
}

// NewUMKMHandler creates a new UMKM handler with service dependency injected.
func NewUMKMHandler(umkmService *umkm.Service, signedURL *galleryuc.SignedURLService, signedURLsEnabled bool) *UMKMHandler {
	return &UMKMHandler{
		umkmService:       umkmService,
		signedURL:        signedURL,
		signedURLsEnabled: signedURLsEnabled,
	}
}

// UMKMResponse represents the response for UMKM data
type UMKMResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	CategoryID  string                 `json:"category_id"`
	Category    *response.CategoryInfo `json:"category,omitempty"`
	Description string                 `json:"description"`
	Owner       *string                `json:"owner,omitempty"`
	Address     *string                `json:"address,omitempty"`
	Phone       *string                `json:"phone,omitempty"`
	Email       *string                `json:"email,omitempty"`
	Website     *string                `json:"website,omitempty"`
	Media       []response.MediaInfo   `json:"media"`
	CreatedAt   string                 `json:"created_at"` // ISO 8601 format
	UpdatedAt   string                 `json:"updated_at"` // ISO 8601 format
}

// buildCategoryInfoStr builds a CategoryInfo from a non-pointer category ID (UMKM.Category is a string).
func buildCategoryInfoStr(id string, name *string) *response.CategoryInfo {
	if id == "" {
		return nil
	}
	n := ""
	if name != nil {
		n = *name
	}
	return &response.CategoryInfo{ID: id, Name: n}
}

// mediaFor builds the shared media shape (Task 4.3): one entry per gallery
// media id with content/thumbnail URLs.
func (h *UMKMHandler) mediaFor(u *domainumkm.UMKM, public bool) []response.MediaInfo {
	out := make([]response.MediaInfo, 0, len(u.ImagesMediaIDs))
	scope := galleryuc.ScopePublic
	sub := "anonymous"
	if !public {
		scope = galleryuc.ScopeAdmin
		sub = "user:admin"
	}
	for _, id := range u.ImagesMediaIDs {
		var url, thumb string
		if h.signedURLsEnabled && h.signedURL != nil {
			url = galleryuc.SignedURLPath("content", id) + h.signedURL.SignedURLQuery(scope, id, sub, 0)
			thumb = galleryuc.SignedURLPath("thumbnail", id) + h.signedURL.SignedURLQuery(scope, id, sub, 0)
		} else if public {
			url = galleryuc.FeatureURLFor(galleryuc.FeatureUMKM, "content", id)
			thumb = galleryuc.FeatureURLFor(galleryuc.FeatureUMKM, "thumbnail", id)
		} else {
			url = galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", id)
			thumb = galleryuc.URLFor(galleryuc.URLScopeAdmin, "thumbnail", id)
		}
		out = append(out, response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb})
	}
	return out
}

// UMKMPaginatedResponse represents a paginated list of UMKM.
type UMKMPaginatedResponse struct {
	UMKM       []UMKMResponse         `json:"umkm"`
	Pagination map[string]interface{} `json:"pagination"`
}

// CreateUMKM godoc
// @Summary      Create UMKM (admin)
// @Description  Multipart or JSON. Required: name, description, category (UUID). Images must be pre-uploaded via POST /umkm/upload-media and referenced with images_media_ids. RBAC: umkm:write.
// @Tags         umkm
// @Accept       mpfd
// @Produce      json
// @Param        name             formData string   true   "Business name"
// @Param        description      formData string   true   "Description"
// @Param        category         formData string   true   "Category UUID"
// @Param        images_media_ids formData []string false  "Pre-uploaded media ids (repeatable)"
// @Success      201  {object} umkm.UMKMResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		413	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /umkm [post]
func (h *UMKMHandler) CreateUMKM(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	var name, category, description string
	var owner, address, phone, email, website *string
	var imagesMediaIDs []string

	if strings.Contains(contentType, "multipart/form-data") {
		var input struct {
			Name           string   `in:"form=name" validate:"required,min=1"`
			Category       string   `in:"form=category" validate:"required,min=1"`
			Description    string   `in:"form=description" validate:"required,min=1"`
			Owner          *string  `in:"form=owner"`
			Address        *string  `in:"form=address"`
			Phone          *string  `in:"form=phone"`
			Email          *string  `in:"form=email" validate:"omitempty,email"`
			Website        *string  `in:"form=website"`
			ImagesMediaIDs []string `in:"form=images_media_ids"`
		}
		if err := httpin.DecodeTo(r, &input); err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to parse form data")
			return
		}
		if err := handlerutil.ValidateStruct(input); err != nil {
			response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
			return
		}
		name = input.Name
		category = input.Category
		description = input.Description
		owner = input.Owner
		address = input.Address
		phone = input.Phone
		email = input.Email
		website = input.Website
		imagesMediaIDs = input.ImagesMediaIDs
	} else {
		var input struct {
			Payload struct {
				Name           string   `json:"name" validate:"required,min=1"`
				Category       string   `json:"category" validate:"required,min=1"`
				Description    string   `json:"description" validate:"required,min=1"`
				Owner          *string  `json:"owner"`
				Address        *string  `json:"address"`
				Phone          *string  `json:"phone"`
				Email          *string  `json:"email" validate:"omitempty,email"`
				Website        *string  `json:"website"`
				ImagesMediaIDs []string `json:"images_media_ids"`
			} `in:"body=json"`
		}
		if err := httpin.DecodeTo(r, &input); err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to parse JSON body")
			return
		}
		if err := handlerutil.ValidateStruct(input.Payload); err != nil {
			response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
			return
		}
		name = input.Payload.Name
		category = input.Payload.Category
		description = input.Payload.Description
		owner = input.Payload.Owner
		address = input.Payload.Address
		phone = input.Payload.Phone
		email = input.Payload.Email
		website = input.Payload.Website
		imagesMediaIDs = input.Payload.ImagesMediaIDs
	}

	serviceInput := umkm.CreateUMKMInput{
		Name:           name,
		Category:       category,
		Description:    description,
		Owner:          owner,
		Address:        address,
		Phone:          phone,
		Email:          email,
		Website:        website,
		ImagesMediaIDs: imagesMediaIDs,
	}

	u, err := h.umkmService.Create(r.Context(), serviceInput)
	if err != nil {
		if respondIfBulkLimit(w, err) {
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to create UMKM", err)
		return
	}

	resp := UMKMResponse{
		ID:          u.ID,
		Name:        u.Name,
		CategoryID:  u.Category,
		Category:    buildCategoryInfoStr(u.Category, u.CategoryName),
		Description: u.Description,
		Owner:       u.Owner,
		Address:     u.Address,
		Phone:       u.Phone,
		Email:       u.Email,
		Website:     u.Website,
		Media:       h.mediaFor(u, false),
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListUMKM godoc
// @Summary      List UMKM (admin)
// @Description  Query: q, category (UUID), page, limit. RBAC: umkm:read.
// @Tags         umkm
// @Produce      json
// @Param        q          query    string  false  "Search"
// @Param        category   query    string  false  "Category UUID"
// @Param        page       query    int  false  "Page default(1)"
// @Param        limit      query    int  false  "Page size default(10)"
// @Success      200  {object} umkm.UMKMPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /umkm [get]
func (h *UMKMHandler) ListUMKM(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int     `in:"query=page;default=1" validate:"min=1"`
		Limit    int     `in:"query=limit;default=10" validate:"min=1"`
		Q        *string `in:"query=q"`
		Category *string `in:"query=category"`
	}

	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	serviceInput := umkm.ListUMKMInput{
		Query:    input.Q,
		Category: input.Category,
		Page:     input.Page,
		Limit:    input.Limit,
	}

	umkmList, paginationResult, err := h.umkmService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list UMKM", err)
		return
	}

	umkmResponses := make([]UMKMResponse, len(umkmList))
	for i, u := range umkmList {
		umkmResponses[i] = UMKMResponse{
			ID:          u.ID,
			Name:        u.Name,
			CategoryID:  u.Category,
			Category:    buildCategoryInfoStr(u.Category, u.CategoryName),
			Description: u.Description,
			Owner:       u.Owner,
			Address:     u.Address,
			Phone:       u.Phone,
			Email:       u.Email,
			Website:     u.Website,
			Media:       h.mediaFor(u, false),
			CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"umkm":       umkmResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetUMKM godoc
// @Summary      Get UMKM by ID (admin)
// @Description  RBAC: umkm:read.
// @Tags         umkm
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Success      200      {object}  umkm.UMKMResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /umkm/{id} [get]
// GetUMKM handles GET /api/v1/umkm/:id
// Retrieves a UMKM by ID (public endpoint).
//
// Validates: Requirements 10.5, 11.2, 16.1, 16.2
func (h *UMKMHandler) GetUMKM(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "UMKM ID is required")
		return
	}

	u, err := h.umkmService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "UMKM not found")
		return
	}

	resp := UMKMResponse{
		ID:          u.ID,
		Name:        u.Name,
		CategoryID:  u.Category,
		Category:    buildCategoryInfoStr(u.Category, u.CategoryName),
		Description: u.Description,
		Owner:       u.Owner,
		Address:     u.Address,
		Phone:       u.Phone,
		Email:       u.Email,
		Website:     u.Website,
		Media:       h.mediaFor(u, false),
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateUMKM godoc
// @Summary      Update UMKM (admin)
// @Description  RBAC: umkm:write.
// @Tags         umkm
// @Accept       mpfd
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Param        request  body      object  false  "JSON payload"
// @Success      200      {object}  umkm.UMKMResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		413	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /umkm/{id} [put]
// UpdateUMKM handles PUT /api/v1/umkm/:id
// Updates an existing UMKM.
//
// Validates: Requirements 10.7, 11.4, 16.1, 16.2
func (h *UMKMHandler) UpdateUMKM(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "UMKM ID is required")
		return
	}

	contentType := r.Header.Get("Content-Type")
	var name, category, description string
	var owner, address, phone, email, website *string
	var imagesMediaIDs []string
	var imagesMediaIDsProvided bool

	if strings.Contains(contentType, "multipart/form-data") {
		var input struct {
			Name           string   `in:"form=name" validate:"required,min=1"`
			Category       string   `in:"form=category" validate:"required,min=1"`
			Description    string   `in:"form=description" validate:"required,min=1"`
			Owner          *string  `in:"form=owner"`
			Address        *string  `in:"form=address"`
			Phone          *string  `in:"form=phone"`
			Email          *string  `in:"form=email" validate:"omitempty,email"`
			Website        *string  `in:"form=website"`
			ImagesMediaIDs []string `in:"form=images_media_ids"`
		}
		if err := httpin.DecodeTo(r, &input); err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to parse form data")
			return
		}
		if err := handlerutil.ValidateStruct(input); err != nil {
			response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
			return
		}
		name = input.Name
		category = input.Category
		description = input.Description
		owner = input.Owner
		address = input.Address
		phone = input.Phone
		email = input.Email
		website = input.Website
		if input.ImagesMediaIDs != nil {
			imagesMediaIDs = input.ImagesMediaIDs
			imagesMediaIDsProvided = true
		}
	} else {
		var input struct {
			Payload struct {
				Name           string   `json:"name" validate:"required,min=1"`
				Category       string   `json:"category" validate:"required,min=1"`
				Description    string   `json:"description" validate:"required,min=1"`
				Owner          *string  `json:"owner"`
				Address        *string  `json:"address"`
				Phone          *string  `json:"phone"`
				Email          *string  `json:"email" validate:"omitempty,email"`
				Website        *string  `json:"website"`
				ImagesMediaIDs []string `json:"images_media_ids"`
			} `in:"body=json"`
		}
		if err := httpin.DecodeTo(r, &input); err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to parse JSON body")
			return
		}
		if err := handlerutil.ValidateStruct(input.Payload); err != nil {
			response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
			return
		}
		name = input.Payload.Name
		category = input.Payload.Category
		description = input.Payload.Description
		owner = input.Payload.Owner
		address = input.Payload.Address
		phone = input.Payload.Phone
		email = input.Payload.Email
		website = input.Payload.Website
		if input.Payload.ImagesMediaIDs != nil {
			imagesMediaIDs = input.Payload.ImagesMediaIDs
			imagesMediaIDsProvided = true
		}
	}

	serviceInput := umkm.UpdateUMKMInput{
		Name:        name,
		Category:    category,
		Description: description,
		Owner:       owner,
		Address:     address,
		Phone:       phone,
		Email:       email,
		Website:     website,
	}
	if imagesMediaIDsProvided {
		serviceInput.ImagesMediaIDs = imagesMediaIDs
	}

	u, err := h.umkmService.Update(r.Context(), id, serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "umkm not found") {
			response.Error(w, http.StatusNotFound, "UMKM not found")
			return
		}
		if respondIfBulkLimit(w, err) {
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update UMKM", err)
		return
	}

	resp := UMKMResponse{
		ID:          u.ID,
		Name:        u.Name,
		CategoryID:  u.Category,
		Category:    buildCategoryInfoStr(u.Category, u.CategoryName),
		Description: u.Description,
		Owner:       u.Owner,
		Address:     u.Address,
		Phone:       u.Phone,
		Email:       u.Email,
		Website:     u.Website,
		Media:       h.mediaFor(u, false),
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeleteUMKM godoc
// @Summary      Delete UMKM (admin)
// @Description  Deletes UMKM and all images. RBAC: umkm:write.
// @Tags         umkm
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Success      200      {object}  response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /umkm/{id} [delete]
// DeleteUMKM handles DELETE /api/v1/umkm/:id
// Deletes a UMKM.
//
// Validates: Requirements 10.8, 11.5, 16.1, 16.2
func (h *UMKMHandler) DeleteUMKM(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "UMKM ID is required")
		return
	}

	err := h.umkmService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "umkm not found") {
			response.Error(w, http.StatusNotFound, "UMKM not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete UMKM", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "UMKM deleted successfully",
	})
}

// ListUMKMPublic godoc
// @Summary      List UMKM (public)
// @Description  Public paginated list.
// @Tags         umkm
// @Produce      json
// @Param        q          query    string  false  "Search"
// @Param        category   query    string  false  "Category UUID"
// @Param        page       query    int  false  "Page default(1)"
// @Param        limit      query    int  false  "Page size default(10)"
// @Success      200  {object} umkm.UMKMPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/umkm/list [get]
func (h *UMKMHandler) ListUMKMPublic(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page  int `in:"query=page;default=1"`
		Limit int `in:"query=limit;default=10"`
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

	serviceInput := umkm.ListUMKMInput{
		Page:  input.Page,
		Limit: input.Limit,
	}

	umkmList, paginationResult, err := h.umkmService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list UMKM", err)
		return
	}

	umkmResponses := make([]UMKMResponse, len(umkmList))
	for i, u := range umkmList {
		umkmResponses[i] = UMKMResponse{
			ID:          u.ID,
			Name:        u.Name,
			CategoryID:  u.Category,
			Category:    buildCategoryInfoStr(u.Category, u.CategoryName),
			Description: u.Description,
			Owner:       u.Owner,
			Address:     u.Address,
			Phone:       u.Phone,
			Email:       u.Email,
			Website:     u.Website,
			Media:       h.mediaFor(u, true),
			CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"umkm":       umkmResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetUMKMPublic godoc
// @Summary      Get UMKM (public)
// @Description  Public full detail.
// @Tags         umkm
// @Produce      json
// @Param        id         path     string true   "UMKM UUID"
// @Success      200  {object} umkm.UMKMResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/umkm/{id} [get]
func (h *UMKMHandler) GetUMKMPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "UMKM ID is required")
		return
	}

	u, err := h.umkmService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "umkm not found") {
			response.Error(w, http.StatusNotFound, "UMKM not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get UMKM", err)
		return
	}

	resp := UMKMResponse{
		ID:          u.ID,
		Name:        u.Name,
		CategoryID:  u.Category,
		Category:    buildCategoryInfoStr(u.Category, u.CategoryName),
		Description: u.Description,
		Owner:       u.Owner,
		Address:     u.Address,
		Phone:       u.Phone,
		Email:       u.Email,
		Website:     u.Website,
		Media:       h.mediaFor(u, true),
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
