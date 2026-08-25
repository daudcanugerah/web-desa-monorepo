package infographic

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ggicci/httpin"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"webdesa/api/pkg/response"
	"webdesa/api/usecase/infographic"

	"webdesa/api/pkg/handlerutil"
)

// InfographicHandler handles HTTP requests for infographic management operations.
// It accepts the infographic service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
type InfographicHandler struct {
	infographicService *infographic.Service
}

// NewInfographicHandler creates a new infographic handler with service dependency injected.
func NewInfographicHandler(infographicService *infographic.Service) *InfographicHandler {
	return &InfographicHandler{
		infographicService: infographicService,
	}
}

// buildCategoryInfo assembles a response.CategoryInfo from raw id and name pointers.
// Returns nil when id is nil (no category linked).
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

// InfographicResponse represents the response for infographic data
type InfographicResponse struct {
	ID              string                 `json:"id"`
	ComponentID     int64                  `json:"component_id"`
	ComponentType   string                 `json:"component_type"`
	SectionName     string                 `json:"section_name"`
	SectionEndpoint string                 `json:"section_endpoint"`
	CategoryID      *string                `json:"category_id,omitempty"`
	Category        *response.CategoryInfo `json:"category,omitempty"`
	State           bool                   `json:"state"`
	CreatedAt       string                 `json:"created_at"`      // ISO 8601 format
	UpdatedAt       string                 `json:"updated_at"`      // ISO 8601 format
	Token           string                 `json:"token,omitempty"` // JWT token for Metabase (only in GET)
}

// InfographicPaginatedResponse is a paginated list of infographics
type InfographicPaginatedResponse struct {
	Infographics []InfographicResponse  `json:"infographic"`
	Pagination   map[string]interface{} `json:"pagination"`
}

// InfographicSectionsResponse is the list of distinct infographic section names
type InfographicSectionsResponse struct {
	Sections []string `json:"section_names"`
}

// InfographicTokenResponse carries a freshly generated Metabase preview JWT
type InfographicTokenResponse struct {
	Token string `json:"token"`
}

// CreateInfographic godoc
// @Summary      Create infographic (admin)
// @Description  Required: component_id, component_type (question|dashboard), section_name, section_endpoint. Optional: category (UUID), state. RBAC: infographic:write.
// @Tags         infographic
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "{component_id, component_type, section_name, section_endpoint, category?, state}"
// @Success      201  {object}  infographic.InfographicResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic [post]
func (h *InfographicHandler) CreateInfographic(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			ComponentID     int64   `json:"component_id" validate:"required,min=1"`
			ComponentType   string  `json:"component_type" validate:"required,min=1"`
			SectionName     string  `json:"section_name" validate:"required,min=1"`
			SectionEndpoint string  `json:"section_endpoint" validate:"required,min=1"`
			Category        *string `json:"category"`
			State           bool    `json:"state"`
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

	serviceInput := infographic.CreateInfographicInput{
		ComponentID:     req.Payload.ComponentID,
		ComponentType:   req.Payload.ComponentType,
		SectionName:     req.Payload.SectionName,
		SectionEndpoint: req.Payload.SectionEndpoint,
		Category:        req.Payload.Category,
		State:           req.Payload.State,
	}

	i, err := h.infographicService.Create(r.Context(), serviceInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := InfographicResponse{
		ID:              i.ID,
		ComponentID:     i.ComponentID,
		ComponentType:   string(i.ComponentType),
		SectionName:     i.SectionName,
		SectionEndpoint: i.SectionEndpoint,
		CategoryID:      i.Category,
		Category:        buildCategoryInfo(i.Category, i.CategoryName),
		State:           i.State,
		CreatedAt:       i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       i.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusCreated, resp)
}

// GetInfographic godoc
// @Summary      Get infographic by ID (admin)
// @Description  Generates Metabase JWT for embedding. RBAC: infographic:read.
// @Tags         infographic
// @Produce      json
// @Param        id         path     string true   "Infographic UUID"
// @Success      200  {object}  infographic.InfographicResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic/{id} [get]
func (h *InfographicHandler) GetInfographic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Infographic ID is required")
		return
	}

	i, err := h.infographicService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Infographic not found")
		return
	}

	// Generate Metabase JWT token via service. Admin path uses 10m TTL.
	token, err := h.infographicService.GenerateMetabaseToken(i.ComponentID, string(i.ComponentType))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}
	_ = h.infographicService.LogAccess(r.Context(), id, i.ComponentID, string(i.ComponentType), "admin_detail", extractRequestMeta(r), 10*time.Minute)

	resp := InfographicResponse{
		ID:              i.ID,
		ComponentID:     i.ComponentID,
		ComponentType:   string(i.ComponentType),
		SectionName:     i.SectionName,
		SectionEndpoint: i.SectionEndpoint,
		CategoryID:      i.Category,
		Category:        buildCategoryInfo(i.Category, i.CategoryName),
		State:           i.State,
		CreatedAt:       i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       i.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Token:           token,
	}

	response.Success(w, http.StatusOK, resp)
}

// ListInfographic godoc
// @Summary      List infographics (admin)
// @Description  Filter by section_name, state, q (search section_name + component_id), and category (UUID). RBAC: infographic:read.
// @Tags         infographic
// @Produce      json
//
//	@Param	section_name	query	string	false	"Section"
//	@Param	state	query	boolean	false	"State"
//	@Param	q	query	string	false	"Search section_name or component_id"
//	@Param	category	query	string	false	"Category UUID"
//	@Param	page	query	int	false	"Page"
//	@Param	limit	query	int	false	"Page size"
//
// @Success      200  {object}  infographic.InfographicPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic [get]
func (h *InfographicHandler) ListInfographic(w http.ResponseWriter, r *http.Request) {
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
	categoryStr := r.URL.Query().Get("category")

	var state *bool
	if stateStr != "" {
		s := stateStr == "true"
		state = &s
	}

	serviceInput := infographic.ListInfographicsInput{
		SectionName: nil,
		State:       state,
		Query:       nil,
		Category:    nil,
		Page:        page,
		Limit:       limit,
	}

	if sectionName != "" {
		serviceInput.SectionName = &sectionName
	}

	if qStr != "" {
		serviceInput.Query = &qStr
	}

	if categoryStr != "" {
		serviceInput.Category = &categoryStr
	}

	infographics, paginationResult, err := h.infographicService.List(r.Context(), serviceInput)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list infographics")
		return
	}

	var infographicResponses = make([]InfographicResponse, 0)
	for _, i := range infographics {
		infographicResponses = append(infographicResponses, InfographicResponse{
			ID:              i.ID,
			ComponentID:     i.ComponentID,
			ComponentType:   string(i.ComponentType),
			SectionName:     i.SectionName,
			SectionEndpoint: i.SectionEndpoint,
			CategoryID:      i.Category,
			Category:        buildCategoryInfo(i.Category, i.CategoryName),
			State:           i.State,
			CreatedAt:       i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       i.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	resp := map[string]interface{}{
		"infographic": infographicResponses,
		"pagination":  paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// UpdateInfographic godoc
// @Summary      Update infographic (admin)
// @Description  RBAC: infographic:write.
// @Tags         infographic
// @Accept       json
// @Produce      json
// @Param        id         path     string true   "Infographic UUID"
// @Success      200  {object}  infographic.InfographicResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic/{id} [put]
func (h *InfographicHandler) UpdateInfographic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Infographic ID is required")
		return
	}

	var req struct {
		Payload struct {
			ComponentID     int64   `json:"component_id" validate:"required,min=1"`
			ComponentType   string  `json:"component_type" validate:"required,min=1"`
			SectionName     string  `json:"section_name" validate:"required,min=1"`
			SectionEndpoint string  `json:"section_endpoint" validate:"required,min=1"`
			Category        *string `json:"category"`
			State           bool    `json:"state"`
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

	serviceInput := infographic.UpdateInfographicInput{
		ComponentID:     req.Payload.ComponentID,
		ComponentType:   req.Payload.ComponentType,
		SectionName:     req.Payload.SectionName,
		SectionEndpoint: req.Payload.SectionEndpoint,
		Category:        req.Payload.Category,
		State:           req.Payload.State,
	}

	i, err := h.infographicService.Update(r.Context(), id, serviceInput)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := InfographicResponse{
		ID:              i.ID,
		ComponentID:     i.ComponentID,
		ComponentType:   string(i.ComponentType),
		SectionName:     i.SectionName,
		SectionEndpoint: i.SectionEndpoint,
		CategoryID:      i.Category,
		Category:        buildCategoryInfo(i.Category, i.CategoryName),
		State:           i.State,
		CreatedAt:       i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       i.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// DeleteInfographic godoc
// @Summary      Delete infographic (admin)
// @Description  RBAC: infographic:write.
// @Tags         infographic
// @Produce      json
// @Param        id         path     string true   "Infographic UUID"
// @Success      200  {object}  response.MessageResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic/{id} [delete]
func (h *InfographicHandler) DeleteInfographic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Infographic ID is required")
		return
	}

	if err := h.infographicService.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Infographic deleted successfully"})
}

// GetInfographicSectionNames godoc
// @Summary      List distinct infographic section names (admin)
// @Description  RBAC: infographic:read.
// @Tags         infographic
// @Produce      json
// @Success      200  {object}  infographic.InfographicSectionsResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic/sections/names [get]
func (h *InfographicHandler) GetInfographicSectionNames(w http.ResponseWriter, r *http.Request) {
	names, err := h.infographicService.GetSectionNames(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get section names")
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"section_names": names,
	})
}

// GeneratePreviewToken godoc
// @Summary      Generate Metabase JWT for preview (admin)
// @Description  Generates HS256 JWT for a (component_id, component_type) pair without persisting. RBAC: infographic:write.
// @Tags         infographic
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "{component_id, component_type}"
// @Success      200  {object}  infographic.InfographicTokenResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic/preview/token [post]
func (h *InfographicHandler) GeneratePreviewToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			ComponentID   int64  `json:"component_id" validate:"required,min=1"`
			ComponentType string `json:"component_type" validate:"required,min=1"`
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

	// Validate component type
	if req.Payload.ComponentType != "question" && req.Payload.ComponentType != "dashboard" {
		response.Error(w, http.StatusBadRequest, "Validation failed: component_type must be 'question' or 'dashboard'")
		return
	}

	// Generate token via service
	token, err := h.infographicService.GenerateMetabaseToken(req.Payload.ComponentID, req.Payload.ComponentType)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	resp := map[string]interface{}{
		"component_id":   req.Payload.ComponentID,
		"component_type": req.Payload.ComponentType,
		"token":          token,
	}

	response.Success(w, http.StatusOK, resp)
}

// ListInfographicPublic godoc
// @Summary      List infographics (public)
// @Description  Public read, no token field.
// @Tags         infographic
// @Produce      json
//
//	@Param	page	query	int	false	"Page"
//	@Param	limit	query	int	false	"Page size"
//
// @Success      200  {object}  infographic.InfographicPaginatedResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/infographic/list [get]
func (h *InfographicHandler) ListInfographicPublic(w http.ResponseWriter, r *http.Request) {
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

	serviceInput := infographic.ListInfographicsInput{
		Page:  input.Page,
		Limit: input.Limit,
	}

	infographics, paginationResult, err := h.infographicService.List(r.Context(), serviceInput)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list infographics")
		return
	}

	var infographicResponses = make([]InfographicResponse, 0)
	for _, inf := range infographics {
		infographicResponses = append(infographicResponses, InfographicResponse{
			ID:              inf.ID,
			ComponentID:     inf.ComponentID,
			ComponentType:   string(inf.ComponentType),
			SectionName:     inf.SectionName,
			SectionEndpoint: inf.SectionEndpoint,
			CategoryID:      inf.Category,
			Category:        buildCategoryInfo(inf.Category, inf.CategoryName),
			State:           inf.State,
			CreatedAt:       inf.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       inf.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	resp := map[string]interface{}{
		"infographic": infographicResponses,
		"pagination":  paginationResult,
	}

	response.Success(w, http.StatusOK, resp)
}

// GetInfographicPublic godoc
// @Summary      Get infographic (public)
// @Description  Public read with Metabase JWT.
// @Tags         infographic
// @Produce      json
// @Param        id         path     string true   "Infographic UUID"
// @Success      200  {object}  infographic.InfographicResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/infographic/{id} [get]
func (h *InfographicHandler) GetInfographicPublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Infographic ID is required")
		return
	}

	meta := extractRequestMeta(r)
	inf, token, _, err := h.infographicService.GetPublicAccess(r.Context(), id, meta)
	if err != nil {
		if strings.Contains(err.Error(), "rate limit exceeded") {
			response.Error(w, http.StatusTooManyRequests, err.Error())
			return
		}
		response.Error(w, http.StatusNotFound, "Infographic not found")
		return
	}

	resp := InfographicResponse{
		ID:              inf.ID,
		ComponentID:     inf.ComponentID,
		ComponentType:   string(inf.ComponentType),
		SectionName:     inf.SectionName,
		SectionEndpoint: inf.SectionEndpoint,
		CategoryID:      inf.Category,
		Category:        buildCategoryInfo(inf.Category, inf.CategoryName),
		State:           inf.State,
		CreatedAt:       inf.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       inf.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Token:           token,
	}

	response.Success(w, http.StatusOK, resp)
}

// extractRequestMeta pulls IP, User-Agent, and Referer from the HTTP request
// for audit logging and per-component rate limiting. IP is taken from
// X-Forwarded-For when present (trusted proxy) or RemoteAddr otherwise.
func extractRequestMeta(r *http.Request) infographic.RequestMeta {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx > 0 {
			ip = ip[:idx]
		}
	}
	return infographic.RequestMeta{
		IPAddress: ip,
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
	}
}

// AccessLogResponse is the JSON shape for an audit log entry.
type AccessLogResponse struct {
	ID             string `json:"id"`
	InfographicID  string `json:"infographic_id"`
	ComponentID    int64  `json:"component_id"`
	ComponentType  string `json:"component_type"`
	Endpoint       string `json:"endpoint"`
	IPAddress      string `json:"ip_address"`
	UserAgent      string `json:"user_agent"`
	Referer        string `json:"referer"`
	TokenIssuedAt  string `json:"token_issued_at"`
	TokenExpiresAt string `json:"token_expires_at"`
	CreatedAt      string `json:"created_at"`
}

type AccessLogListResponse struct {
	Logs       []AccessLogResponse    `json:"logs"`
	Pagination map[string]interface{} `json:"pagination"`
}

// ListAccessLogs godoc
// @Summary      List infographic access logs (admin)
// @Description  Audit log of every Metabase token generation. Filter by infographic_id or ip. RBAC: infographic:read.
// @Tags         infographic
// @Produce      json
// @Param        infographic_id   query     string  false  "Filter by infographic UUID"
// @Param        ip                query     string  false  "Filter by IP address"
// @Param        page              query     int     false  "Page default(1)"
// @Param        limit             query     int     false  "Page size default(20)"
// @Success      200  {object}  infographic.AccessLogListResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /infographic/access-logs [get]
func (h *InfographicHandler) ListAccessLogs(w http.ResponseWriter, r *http.Request) {
	var input struct {
		InfographicID *string `in:"query=infographic_id"`
		IP            *string `in:"query=ip"`
		Page          int     `in:"query=page;default=1" validate:"min=1"`
		Limit         int     `in:"query=limit;default=20" validate:"min=1"`
	}
	if err := httpin.DecodeTo(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}
	if err := handlerutil.ValidateStruct(input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	var infID *uuid.UUID
	if input.InfographicID != nil && *input.InfographicID != "" {
		parsed, err := uuid.Parse(*input.InfographicID)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid infographic_id")
			return
		}
		infID = &parsed
	}
	ip := ""
	if input.IP != nil {
		ip = *input.IP
	}
	if infID == nil && ip == "" {
		response.Error(w, http.StatusBadRequest, "either infographic_id or ip is required")
		return
	}

	logs, total, err := h.infographicService.ListAccessLogs(r.Context(), infID, ip, input.Page, input.Limit)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to list access logs", err)
		return
	}

	out := make([]AccessLogResponse, 0, len(logs))
	for _, l := range logs {
		ip := ""
		if l.IPAddress != "" {
			ip = l.IPAddress
		}
		out = append(out, AccessLogResponse{
			ID:             l.ID.String(),
			InfographicID:  l.InfographicID.String(),
			ComponentID:    l.ComponentID,
			ComponentType:  l.ComponentType,
			Endpoint:       l.Endpoint,
			IPAddress:      ip,
			UserAgent:      l.UserAgent,
			Referer:        l.Referer,
			TokenIssuedAt:  l.TokenIssuedAt.Format(time.RFC3339),
			TokenExpiresAt: l.TokenExpiresAt.Format(time.RFC3339),
			CreatedAt:      l.CreatedAt.Format(time.RFC3339),
		})
	}

	resp := map[string]interface{}{
		"logs":       out,
		"pagination": map[string]interface{}{"page": input.Page, "limit": input.Limit, "total": total},
	}
	response.Success(w, http.StatusOK, resp)
}
