package struktur

import (
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	"webdesa/api/pkg/response"
	"webdesa/api/usecase/struktur"

	"webdesa/api/pkg/handlerutil")

// StrukturHandler handles HTTP requests for struktur (organizational structure) management operations.
// It accepts the struktur service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
type StrukturHandler struct {
	strukturService *struktur.Service
}

// NewStrukturHandler creates a new struktur handler with service dependency injected.
func NewStrukturHandler(strukturService *struktur.Service) *StrukturHandler {
	return &StrukturHandler{
		strukturService: strukturService,
	}
}

// StrukturResponse represents the response for struktur data
type StrukturResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Position        *string `json:"position,omitempty"`
	Email           *string `json:"email,omitempty"`
	Phone           *string `json:"phone,omitempty"`
	Description     *string `json:"description,omitempty"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
	CreatedAt       string  `json:"created_at"` // ISO 8601 format
	UpdatedAt       string  `json:"updated_at"` // ISO 8601 format
}

// StrukturPaginatedResponse represents a paginated list of struktur members.
type StrukturPaginatedResponse struct {
	Struktur   []StrukturResponse      `json:"struktur"`
	Pagination map[string]interface{} `json:"pagination"`
}

// CreateStruktur godoc
// @Summary      Create struktur member (admin)
// @Description  Multipart. Required: name. Optional: position, email, phone, description, profile_image. RBAC: struktur:write.
// @Tags         struktur
// @Accept       mpfd
// @Produce      json
// @Param        name       formData string true   "Member name"
// @Param        position   formData string  false  "Position"
// @Param        email      formData string  false  "Email"
// @Param        phone      formData string  false  "Phone"
// @Param        description formData string  false  "Description"
// @Param        profile_image formData file  false  "Profile image"
// @Success      200  {object} struktur.StrukturResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /struktur [post]
func (h *StrukturHandler) CreateStruktur(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string       `in:"form=name" validate:"required,min=1"`
		Position    *string      `in:"form=position"`
		Email       *string      `in:"form=email" validate:"omitempty,email"`
		Phone       *string      `in:"form=phone"`
		Description *string      `in:"form=description"`
		Image       *httpin.File `in:"form=image"`
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

	// Get profile image file (optional)
	var strukturInput struktur.CreateStrukturInput
	if input.Image != nil {
		imageFile, err := input.Image.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read image file")
			return
		}
		strukturInput = struktur.CreateStrukturInput{
			Name:        input.Name,
			Position:    input.Position,
			Email:       input.Email,
			Phone:       input.Phone,
			Description: input.Description,
			ImageFile:   imageFile,
			ImageName:   input.Image.Filename(),
			ImageSize:   input.Image.Size(),
			ContentType: input.Image.MIMEHeader().Get("Content-Type"),
		}
	} else {
		// No image provided
		strukturInput = struktur.CreateStrukturInput{
			Name:        input.Name,
			Position:    input.Position,
			Email:       input.Email,
			Phone:       input.Phone,
			Description: input.Description,
			ImageFile:   nil,
		}
	}

	// Call struktur service
	st, err := h.strukturService.Create(r.Context(), strukturInput)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to create struktur", err)
		return
	}

	// Build response
	resp := StrukturResponse{
		ID:              st.ID,
		Name:            st.Name,
		Position:        st.Position,
		Email:           st.Email,
		Phone:           st.Phone,
		Description:     st.Description,
		ProfileImageURL: st.ProfileImageURL,
		CreatedAt:       st.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       st.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// ListStruktur godoc
// @Summary      List struktur (admin)
// @Description  RBAC: struktur:read.
// @Tags         struktur
// @Produce      json
// @Param        q          query    string  false  "Search"
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object} struktur.StrukturPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /struktur [get]
func (h *StrukturHandler) ListStruktur(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page  int     `in:"query=page;default=1" validate:"min=1"`
		Limit int     `in:"query=limit;default=10" validate:"min=1"`
		Q     *string `in:"query=q"`
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

	// Call struktur service
	serviceInput := struktur.ListStrukturInput{
		Query: input.Q,
		Page:  input.Page,
		Limit: input.Limit,
	}

	strukturList, paginationResult, err := h.strukturService.List(r.Context(), serviceInput)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list struktur", err)
		return
	}

	// Build response
	strukturResponses := make([]StrukturResponse, len(strukturList))
	for i, st := range strukturList {
		strukturResponses[i] = StrukturResponse{
			ID:              st.ID,
			Name:            st.Name,
			Position:        st.Position,
			Email:           st.Email,
			Phone:           st.Phone,
			Description:     st.Description,
			ProfileImageURL: st.ProfileImageURL,
			CreatedAt:       st.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       st.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"struktur":   strukturResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetStruktur godoc
// @Summary      Get struktur by ID (admin)
// @Description  RBAC: struktur:read.
// @Tags         struktur
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Success      200      {object}  struktur.StrukturResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /struktur/{id} [get]
// GetStruktur handles GET /api/v1/struktur/:id
// Retrieves a struktur member by ID (public endpoint).
//
// Validates: Requirements 13.3, 16.1, 16.2
func (h *StrukturHandler) GetStruktur(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Struktur ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid struktur ID format")
		return
	}

	// Call struktur service
	st, err := h.strukturService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Struktur not found")
		return
	}

	// Build response
	resp := StrukturResponse{
		ID:              st.ID,
		Name:            st.Name,
		Position:        st.Position,
		Email:           st.Email,
		Phone:           st.Phone,
		Description:     st.Description,
		ProfileImageURL: st.ProfileImageURL,
		CreatedAt:       st.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       st.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateStruktur godoc
// @Summary      Update struktur (admin)
// @Description  RBAC: struktur:write.
// @Tags         struktur
// @Accept       mpfd
// @Produce      json
// @Param        id  path  string  true  "Resource ID"
// @Param        request  body      object  false  "JSON payload"
// @Success      200      {object}  struktur.StrukturResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /struktur/{id} [put]
// UpdateStruktur handles PUT /api/v1/struktur/:id
// Updates an existing struktur member with optional profile image update.
//
// Validates: Requirements 13.5, 16.1, 16.2
func (h *StrukturHandler) UpdateStruktur(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Struktur ID is required")
		return
	}

	var input struct {
		Name        string       `in:"form=name" validate:"required,min=1"`
		Position    *string      `in:"form=position"`
		Email       *string      `in:"form=email" validate:"omitempty,email"`
		Phone       *string      `in:"form=phone"`
		Description *string      `in:"form=description"`
		Image       *httpin.File `in:"form=image"`
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

	// Get profile image file (optional for update)
	var strukturInput struktur.UpdateStrukturInput
	if input.Image != nil {
		imageFile, err := input.Image.Open()
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Failed to read image file")
			return
		}
		strukturInput = struktur.UpdateStrukturInput{
			Name:        input.Name,
			Position:    input.Position,
			Email:       input.Email,
			Phone:       input.Phone,
			Description: input.Description,
			ImageFile:   imageFile,
			ImageName:   input.Image.Filename(),
			ImageSize:   input.Image.Size(),
			ContentType: input.Image.MIMEHeader().Get("Content-Type"),
		}
	} else {
		// No new image provided
		strukturInput = struktur.UpdateStrukturInput{
			Name:        input.Name,
			Position:    input.Position,
			Email:       input.Email,
			Phone:       input.Phone,
			Description: input.Description,
			ImageFile:   nil,
		}
	}

	// Call struktur service
	st, err := h.strukturService.Update(r.Context(), id, strukturInput)
	if err != nil {
		if strings.Contains(err.Error(), "struktur not found") {
			response.Error(w, http.StatusNotFound, "Struktur not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusBadRequest, "Failed to update struktur", err)
		return
	}

	// Build response
	resp := StrukturResponse{
		ID:              st.ID,
		Name:            st.Name,
		Position:        st.Position,
		Email:           st.Email,
		Phone:           st.Phone,
		Description:     st.Description,
		ProfileImageURL: st.ProfileImageURL,
		CreatedAt:       st.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       st.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeleteStruktur godoc
// @Summary      Delete struktur (admin)
// @Description  Deletes profile image. RBAC: struktur:write.
// @Tags         struktur
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
// @Router       /struktur/{id} [delete]
// DeleteStruktur handles DELETE /api/v1/struktur/:id
// Deletes a struktur member and associated profile image.
//
// Validates: Requirements 13.6, 16.1, 16.2
func (h *StrukturHandler) DeleteStruktur(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Struktur ID is required")
		return
	}

	// Call struktur service
	err := h.strukturService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "struktur not found") {
			response.Error(w, http.StatusNotFound, "Struktur not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to delete struktur", err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Struktur deleted successfully",
	})
}

// ListStrukturPublic godoc
// @Summary      List struktur (public)
// @Description  Public paginated list.
// @Tags         struktur
// @Produce      json
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object} struktur.StrukturPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/struktur/list [get]
func (h *StrukturHandler) ListStrukturPublic(w http.ResponseWriter, r *http.Request) {
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

	serviceInput := struktur.ListStrukturInput{
		Page:  input.Page,
		Limit: input.Limit,
	}

	strukturList, paginationResult, err := h.strukturService.List(r.Context(), serviceInput)
	if err != nil {
		if strings.Contains(err.Error(), "invalid pagination") || strings.Contains(err.Error(), "limit cannot exceed") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid pagination parameters", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list struktur", err)
		return
	}

	strukturResponses := make([]StrukturResponse, len(strukturList))
	for i, s := range strukturList {
		strukturResponses[i] = StrukturResponse{
			ID:              s.ID,
			Name:            s.Name,
			Position:        s.Position,
			Email:           s.Email,
			Phone:           s.Phone,
			Description:     s.Description,
			ProfileImageURL: s.ProfileImageURL,
			CreatedAt:       s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := map[string]interface{}{
		"struktur":   strukturResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetStrukturPublic godoc
// @Summary      Get struktur (public)
// @Description  Public full detail.
// @Tags         struktur
// @Produce      json
// @Param        id         path     string true   "Struktur UUID"
// @Success      200  {object} struktur.StrukturResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/struktur/{id} [get]
func (h *StrukturHandler) GetStrukturPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Struktur ID is required")
		return
	}
	if !handlerutil.IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, "Invalid struktur ID format")
		return
	}

	s, err := h.strukturService.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "struktur not found") {
			response.Error(w, http.StatusNotFound, "Struktur not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get struktur", err)
		return
	}

	resp := StrukturResponse{
		ID:              s.ID,
		Name:            s.Name,
		Position:        s.Position,
		Email:           s.Email,
		Phone:           s.Phone,
		Description:     s.Description,
		ProfileImageURL: s.ProfileImageURL,
		CreatedAt:       s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}
