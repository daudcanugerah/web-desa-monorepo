package profile

import (
	"net/http"
	"strconv"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"

	domainprofile "webdesa/api/domain/profile"
	"webdesa/api/pkg/response"
	"webdesa/api/usecase/profile"

	"webdesa/api/pkg/handlerutil"
)

// ProfileHandler handles HTTP requests for profile management operations.
// It accepts the profile service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
type ProfileHandler struct {
	profileService *profile.Service
}

// NewProfileHandler creates a new profile handler with service dependency injected.
func NewProfileHandler(profileService *profile.Service) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

// ProfileResponse represents the response for profile data
type ProfileResponse struct {
	ID              string                 `json:"id"`
	Content         string                 `json:"content"`
	SectionName     string                 `json:"section_name"`
	SectionEndpoint string                 `json:"section_endpoint"`
	State           bool                   `json:"state"`
	CategoryID      *string                `json:"category_id,omitempty"`
	Category        *response.CategoryInfo `json:"category,omitempty"`
	CreatedAt       string                 `json:"created_at"` // ISO 8601 format
	UpdatedAt       string                 `json:"updated_at"` // ISO 8601 format
}

// toProfileResponse builds the shared ProfileResponse shape (incl. category).
func toProfileResponse(p *domainprofile.Profile) ProfileResponse {
	resp := ProfileResponse{
		ID:              p.ID,
		Content:         p.Content,
		SectionName:     p.SectionName,
		SectionEndpoint: p.SectionEndpoint,
		State:           p.State,
		CategoryID:      p.Category,
		CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.Category != nil && *p.Category != "" {
		name := ""
		if p.CategoryName != nil {
			name = *p.CategoryName
		}
		resp.Category = &response.CategoryInfo{ID: *p.Category, Name: name}
	}
	return resp
}

// ProfilePaginatedResponse is a paginated list of profile sections
type ProfilePaginatedResponse struct {
	Profiles   []ProfileResponse      `json:"profile"`
	Pagination map[string]interface{} `json:"pagination"`
}

// ProfileSectionsResponse is the list of distinct profile section names
type ProfileSectionsResponse struct {
	Sections []string `json:"section_names"`
}

// CreateProfile godoc
// @Summary      Create profile section (admin)
// @Description  Required: content, section_name, section_endpoint (must start with /). Optional: state. RBAC: profile:write.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "{content, section_name, section_endpoint, state}"
// @Success      201  {object}  profile.ProfileResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /profile [post]
func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			Content         string  `json:"content" validate:"required,min=1"`
			SectionName     string  `json:"section_name" validate:"required,min=1"`
			SectionEndpoint string  `json:"section_endpoint" validate:"required,min=1"`
			State           bool    `json:"state"`
			Category        *string `json:"category"`
		} `in:"body=json"`
	}

	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	serviceInput := profile.CreateProfileInput{
		Content:         req.Payload.Content,
		SectionName:     req.Payload.SectionName,
		SectionEndpoint: req.Payload.SectionEndpoint,
		State:           req.Payload.State,
		Category:        req.Payload.Category,
	}

	p, err := h.profileService.Create(r.Context(), serviceInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := toProfileResponse(p)

	response.Success(w, http.StatusCreated, resp)
}

// GetProfile godoc
// @Summary      Get profile section by ID (admin)
// @Description  RBAC: profile:read.
// @Tags         profile
// @Produce      json
// @Param        id         path     string true   "Profile UUID"
// @Success      200  {object}  profile.ProfileResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /profile/{id} [get]
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	p, err := h.profileService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Profile not found")
		return
	}

	resp := toProfileResponse(p)

	response.Success(w, http.StatusOK, resp)
}

// ListProfile godoc
// @Summary      List profile sections (admin)
// @Description  Filter by section_name, state, and q (search section_name + content). RBAC: profile:read.
// @Tags         profile
// @Produce      json
// @Param        section_name query    string  false  "Section name"
// @Param        state      query    boolean  false  "Enabled state"
// @Param        q          query    string  false  "Search section_name or content"
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object}  profile.ProfilePaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /profile [get]
func (h *ProfileHandler) ListProfile(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	sectionName := r.URL.Query().Get("section_name")
	stateStr := r.URL.Query().Get("state")
	qStr := r.URL.Query().Get("q")

	var state *bool
	if stateStr != "" {
		s := stateStr == "true"
		state = &s
	}

	serviceInput := profile.ListProfilesInput{
		SectionName: nil,
		State:       state,
		Query:       nil,
		Page:        page,
		Limit:       limit,
	}

	if sectionName != "" {
		serviceInput.SectionName = &sectionName
	}

	if qStr != "" {
		serviceInput.Query = &qStr
	}

	profiles, paginationResult, err := h.profileService.List(r.Context(), serviceInput)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list profiles")
		return
	}

	var profileResponses = make([]ProfileResponse, 0)
	for _, p := range profiles {
		profileResponses = append(profileResponses, toProfileResponse(p))
	}

	resp := map[string]interface{}{
		"profile":    profileResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateProfile godoc
// @Summary      Update profile section (admin)
// @Description  RBAC: profile:write.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        id         path     string true   "Profile UUID"
// @Success      200  {object}  profile.ProfileResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /profile/{id} [put]
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	var req struct {
		Payload struct {
			Content         string  `json:"content" validate:"required,min=1"`
			SectionName     string  `json:"section_name" validate:"required,min=1"`
			SectionEndpoint string  `json:"section_endpoint" validate:"required,min=1"`
			State           bool    `json:"state"`
			Category        *string `json:"category"`
		} `in:"body=json"`
	}

	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	if err := handlerutil.ValidateStruct(req.Payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	serviceInput := profile.UpdateProfileInput{
		Content:         req.Payload.Content,
		SectionName:     req.Payload.SectionName,
		SectionEndpoint: req.Payload.SectionEndpoint,
		State:           req.Payload.State,
		Category:        req.Payload.Category,
	}

	p, err := h.profileService.Update(r.Context(), id, serviceInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := toProfileResponse(p)

	response.Success(w, http.StatusOK, resp)
}

// DeleteProfile godoc
// @Summary      Delete profile section (admin)
// @Description  RBAC: profile:write.
// @Tags         profile
// @Produce      json
// @Param        id         path     string true   "Profile UUID"
// @Success      200  {object}  response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /profile/{id} [delete]
func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	if err := h.profileService.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Profile deleted successfully"})
}

// GetProfileSectionNames godoc
// @Summary      List distinct section names (admin)
// @Description  RBAC: profile:read.
// @Tags         profile
// @Produce      json
// @Success      200  {object}  profile.ProfileSectionsResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /profile/sections/names [get]
func (h *ProfileHandler) GetProfileSectionNames(w http.ResponseWriter, r *http.Request) {
	names, err := h.profileService.GetSectionNames(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get section names")
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"section_names": names,
	})
}

// ListProfilePublic godoc
// @Summary      List profile sections (public)
// @Description  Public read.
// @Tags         profile
// @Produce      json
// @Param        page       query    int  false  "Page"
// @Param        limit      query    int  false  "Page size"
// @Success      200  {object}  profile.ProfilePaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/profile/list [get]
func (h *ProfileHandler) ListProfilePublic(w http.ResponseWriter, r *http.Request) {
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

	serviceInput := profile.ListProfilesInput{
		Page:  input.Page,
		Limit: input.Limit,
	}

	profiles, paginationResult, err := h.profileService.List(r.Context(), serviceInput)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list profiles")
		return
	}

	var profileResponses = make([]ProfileResponse, 0)
	for _, p := range profiles {
		profileResponses = append(profileResponses, toProfileResponse(p))
	}

	resp := map[string]interface{}{
		"profile":    profileResponses,
		"pagination": paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetProfilePublic godoc
// @Summary      Get profile section (public)
// @Description  Public read.
// @Tags         profile
// @Produce      json
// @Param        id         path     string true   "Profile UUID"
// @Success      200  {object}  profile.ProfileResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/profile/{id} [get]
func (h *ProfileHandler) GetProfilePublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	p, err := h.profileService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Profile not found")
		return
	}

	resp := toProfileResponse(p)

	response.Success(w, http.StatusOK, resp)
}
