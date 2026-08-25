package desa

import (
	"net/http"
	"strings"

	"webdesa/api/domain/desa"
	"webdesa/api/pkg/response"
	desaUsecase "webdesa/api/usecase/desa"

	"github.com/ggicci/httpin"

	"webdesa/api/pkg/handlerutil")

// DesaHandler handles HTTP requests for village profile (desa) management.
// The profile is stored as JSON in the settings table under key "desa_profile".
type DesaHandler struct {
	desaService *desaUsecase.Service
}

// NewDesaHandler creates a new desa handler with service dependency injected.
func NewDesaHandler(desaService *desaUsecase.Service) *DesaHandler {
	return &DesaHandler{desaService: desaService}
}

// DesaResponse represents the village profile response.
type DesaResponse struct {
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	Address       *string `json:"address,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Email         *string `json:"email,omitempty"`
	Website       *string `json:"website,omitempty"`
	VisionMission *string `json:"vision_mission,omitempty"`
	UpdatedAt     string  `json:"updated_at"`
}

func toDesaResponse(d *desa.Desa) DesaResponse {
	return DesaResponse{
		Name:          d.Name,
		Description:   d.Description,
		Address:       d.Address,
		Phone:         d.Phone,
		Email:         d.Email,
		Website:       d.Website,
		VisionMission: d.VisionMission,
		UpdatedAt:     d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GetDesa godoc
// @Summary      Get village profile (admin)
// @Description  RBAC: desa:read.
// @Tags         desa
// @Produce      json
// @Success      200  {object} desa.DesaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /desa [get]
func (h *DesaHandler) GetDesa(w http.ResponseWriter, r *http.Request) {
	d, err := h.desaService.Get(r.Context())
	if err != nil {
		response.Error(w, http.StatusNotFound, "Village profile not found")
		return
	}
	response.Success(w, http.StatusOK, toDesaResponse(d))
}

// UpdateDesa godoc
// @Summary      Update village profile (admin)
// @Description  Stored in settings table. Restart may be required for PPID email config. RBAC: desa:write.
// @Tags         desa
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "{name (required), description, address, phone, email, website, vision_mission}"
// @Success      200  {object} desa.DesaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /desa [put]
func (h *DesaHandler) UpdateDesa(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			Name          string  `json:"name" validate:"required,min=1"`
			Description   *string `json:"description"`
			Address       *string `json:"address"`
			Phone         *string `json:"phone"`
			Email         *string `json:"email" validate:"omitempty,email"`
			Website       *string `json:"website"`
			VisionMission *string `json:"vision_mission"`
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

	input := desaUsecase.UpdateDesaInput{
		Name:          req.Payload.Name,
		Description:   req.Payload.Description,
		Address:       req.Payload.Address,
		Phone:         req.Payload.Phone,
		Email:         req.Payload.Email,
		Website:       req.Payload.Website,
		VisionMission: req.Payload.VisionMission,
	}

	d, err := h.desaService.Update(r.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "invalid village profile") ||
			strings.Contains(err.Error(), "website must be") ||
			strings.Contains(err.Error(), "email must be") {
			response.ErrorWithDetails(w, http.StatusBadRequest, "Invalid village profile data", err)
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to update village profile", err)
		return
	}

	response.Success(w, http.StatusOK, toDesaResponse(d))
}

// GetDesaPublic godoc
// @Summary      Get village profile (public)
// @Description  Public read of village profile.
// @Tags         desa
// @Produce      json
// @Success      200  {object} desa.DesaResponse
// @Failure		400	{object} response.ErrorResponse
// @Failure		401	{object} response.ErrorResponse
// @Failure		403	{object} response.ErrorResponse
// @Failure		404	{object} response.ErrorResponse
// @Failure		409	{object} response.ErrorResponse
// @Failure		429	{object} response.ErrorResponse
// @Router       /public/desa [get]
func (h *DesaHandler) GetDesaPublic(w http.ResponseWriter, r *http.Request) {
	d, err := h.desaService.Get(r.Context())
	if err != nil {
		if strings.Contains(err.Error(), "desa not found") {
			response.Error(w, http.StatusNotFound, "Desa not found")
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to get desa", err)
		return
	}

	resp := DesaResponse{
		Name:          d.Name,
		Description:   d.Description,
		Address:       d.Address,
		Phone:         d.Phone,
		Email:         d.Email,
		Website:       d.Website,
		VisionMission: d.VisionMission,
		UpdatedAt:     d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}
