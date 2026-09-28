package desa

import (
	"net/http"
	"strings"

	"webdesa/api/domain/desa"
	"webdesa/api/pkg/response"
	desaUsecase "webdesa/api/usecase/desa"
	galleryuc "webdesa/api/usecase/gallery"

	"github.com/ggicci/httpin"

	"webdesa/api/pkg/handlerutil"
)

// DesaHandler handles HTTP requests for village profile (desa) management.
// The profile is stored as JSON in the settings table under key "desa_profile".
type DesaHandler struct {
	desaService       *desaUsecase.Service
	signedURL         *galleryuc.SignedURLService
	signedURLsEnabled bool
}

// NewDesaHandler creates a new desa handler with service dependency injected.
func NewDesaHandler(desaService *desaUsecase.Service, signedURL *galleryuc.SignedURLService, signedURLsEnabled bool) *DesaHandler {
	return &DesaHandler{
		desaService:       desaService,
		signedURL:         signedURL,
		signedURLsEnabled: signedURLsEnabled,
	}
}

// kepalaDesaMedia builds the shared media shape for the kepala desa photo.
// Public responses embed a scope=public signed URL so anonymous visitors can
// stream the image via /api/v1/media/{id}/...?jwt=.
func (h *DesaHandler) kepalaDesaMedia(mediaID *string, public bool) *response.MediaInfo {
	if mediaID == nil || *mediaID == "" {
		return nil
	}
	id := *mediaID
	scope := galleryuc.ScopePublic
	sub := "anonymous"
	if !public {
		scope = galleryuc.ScopeAdmin
		sub = "user:admin"
	}
	if h.signedURLsEnabled && h.signedURL != nil {
		url := galleryuc.SignedURLPath("content", id) + h.signedURL.SignedURLQuery(scope, id, sub, 0)
		thumb := galleryuc.SignedURLPath("thumbnail", id) + h.signedURL.SignedURLQuery(scope, id, sub, 0)
		return &response.MediaInfo{MediaID: id, URL: url, ThumbnailURL: &thumb}
	}
	if public {
		return &response.MediaInfo{MediaID: id, URL: galleryuc.FeatureURLFor(galleryuc.FeatureDesa, "content", id)}
	}
	return &response.MediaInfo{MediaID: id, URL: galleryuc.URLFor(galleryuc.URLScopeAdmin, "content", id)}
}

// DesaResponse represents the village profile response.
type DesaResponse struct {
	Name              string              `json:"name"`
	Description       *string             `json:"description,omitempty"`
	Address           *string             `json:"address,omitempty"`
	Phone             *string             `json:"phone,omitempty"`
	Email             *string             `json:"email,omitempty"`
	Website           *string             `json:"website,omitempty"`
	VisionMission     *string             `json:"vision_mission,omitempty"`
	KepalaDesa        *string             `json:"kepala_desa,omitempty"`
	KepalaDesaMessage *string             `json:"kepala_desa_message,omitempty"`
	KepalaDesaMediaID *string             `json:"kepala_desa_media_id,omitempty"`
	KepalaDesaMedia   *response.MediaInfo `json:"kepala_desa_media,omitempty"`
	Motto             *string             `json:"motto,omitempty"`
	Kecamatan         *string             `json:"kecamatan,omitempty"`
	Kabupaten         *string             `json:"kabupaten,omitempty"`
	Provinsi          *string             `json:"provinsi,omitempty"`
	JumlahPenduduk    *int                `json:"jumlah_penduduk,omitempty"`
	JumlahKK          *int                `json:"jumlah_kk,omitempty"`
	JumlahDusun       *int                `json:"jumlah_dusun,omitempty"`
	JumlahRT          *int                `json:"jumlah_rt,omitempty"`
	JumlahRW          *int                `json:"jumlah_rw,omitempty"`
	JumlahUMKM        *int                `json:"jumlah_umkm,omitempty"`
	SocialMedia       []desa.SocialLink   `json:"social_media,omitempty"`
	UpdatedAt         string              `json:"updated_at"`
}

// nilIfBlank returns nil for a nil or blank-after-trim string pointer, so
// optional profile fields are omitted instead of stored as empty strings.
func nilIfBlank(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return s
}

// sanitizeSocialLinks drops blank entries, trims whitespace, and lowercases
// the platform slug so the stored JSON stays tidy. Returns nil when nothing
// remains (so the field is omitted from the response).
func sanitizeSocialLinks(links []desa.SocialLink) []desa.SocialLink {
	out := make([]desa.SocialLink, 0, len(links))
	for _, link := range links {
		platform := strings.ToLower(strings.TrimSpace(link.Platform))
		url := strings.TrimSpace(link.URL)
		if platform == "" || url == "" {
			continue
		}
		out = append(out, desa.SocialLink{Platform: platform, URL: url})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (h *DesaHandler) toDesaResponse(d *desa.Desa, public bool) DesaResponse {
	return DesaResponse{
		Name:              d.Name,
		Description:       d.Description,
		Address:           d.Address,
		Phone:             d.Phone,
		Email:             d.Email,
		Website:           d.Website,
		VisionMission:     d.VisionMission,
		KepalaDesa:        d.KepalaDesa,
		KepalaDesaMessage: d.KepalaDesaMessage,
		KepalaDesaMediaID: d.KepalaDesaMediaID,
		KepalaDesaMedia:   h.kepalaDesaMedia(d.KepalaDesaMediaID, public),
		Motto:             d.Motto,
		Kecamatan:         d.Kecamatan,
		Kabupaten:         d.Kabupaten,
		Provinsi:          d.Provinsi,
		JumlahPenduduk:    d.JumlahPenduduk,
		JumlahKK:          d.JumlahKK,
		JumlahDusun:       d.JumlahDusun,
		JumlahRT:          d.JumlahRT,
		JumlahRW:          d.JumlahRW,
		JumlahUMKM:        d.JumlahUMKM,
		SocialMedia:       d.SocialMedia,
		UpdatedAt:         d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
	response.Success(w, http.StatusOK, h.toDesaResponse(d, false))
}

// UpdateDesa godoc
// @Summary      Update village profile (admin)
// @Description  Stored in settings table. Restart may be required for PPID email config. RBAC: desa:write.
// @Tags         desa
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "{name (required), description, address, phone, email, website, vision_mission, kepala_desa, kepala_desa_message, kepala_desa_media_id, motto, kecamatan, kabupaten, provinsi, jumlah_*, social_media: [{platform, url}]}"
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
			Name              string  `json:"name" validate:"required,min=1"`
			Description       *string `json:"description"`
			Address           *string `json:"address"`
			Phone             *string `json:"phone"`
			Email             *string `json:"email" validate:"omitempty,email"`
			Website           *string `json:"website"`
			VisionMission     *string `json:"vision_mission"`
			KepalaDesa        *string `json:"kepala_desa"`
			KepalaDesaMessage *string `json:"kepala_desa_message"`
			KepalaDesaMediaID *string `json:"kepala_desa_media_id"`
			Motto             *string `json:"motto"`
			Kecamatan         *string `json:"kecamatan"`
			Kabupaten         *string `json:"kabupaten"`
			Provinsi          *string `json:"provinsi"`
			JumlahPenduduk    *int    `json:"jumlah_penduduk" validate:"omitempty,gte=0"`
			JumlahKK          *int    `json:"jumlah_kk" validate:"omitempty,gte=0"`
			JumlahDusun       *int    `json:"jumlah_dusun" validate:"omitempty,gte=0"`
			JumlahRT          *int    `json:"jumlah_rt" validate:"omitempty,gte=0"`
			JumlahRW          *int    `json:"jumlah_rw" validate:"omitempty,gte=0"`
			JumlahUMKM        *int    `json:"jumlah_umkm" validate:"omitempty,gte=0"`
			SocialMedia       []desa.SocialLink `json:"social_media"`
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
		Name:              req.Payload.Name,
		Description:       nilIfBlank(req.Payload.Description),
		Address:           nilIfBlank(req.Payload.Address),
		Phone:             nilIfBlank(req.Payload.Phone),
		Email:             nilIfBlank(req.Payload.Email),
		Website:           nilIfBlank(req.Payload.Website),
		VisionMission:     nilIfBlank(req.Payload.VisionMission),
		KepalaDesa:        nilIfBlank(req.Payload.KepalaDesa),
		KepalaDesaMessage: nilIfBlank(req.Payload.KepalaDesaMessage),
		KepalaDesaMediaID: nilIfBlank(req.Payload.KepalaDesaMediaID),
		Motto:             nilIfBlank(req.Payload.Motto),
		Kecamatan:         nilIfBlank(req.Payload.Kecamatan),
		Kabupaten:         nilIfBlank(req.Payload.Kabupaten),
		Provinsi:          nilIfBlank(req.Payload.Provinsi),
		JumlahPenduduk:    req.Payload.JumlahPenduduk,
		JumlahKK:          req.Payload.JumlahKK,
		JumlahDusun:       req.Payload.JumlahDusun,
		JumlahRT:          req.Payload.JumlahRT,
		JumlahRW:          req.Payload.JumlahRW,
		JumlahUMKM:        req.Payload.JumlahUMKM,
		SocialMedia:       sanitizeSocialLinks(req.Payload.SocialMedia),
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

	response.Success(w, http.StatusOK, h.toDesaResponse(d, false))
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

	response.Success(w, http.StatusOK, h.toDesaResponse(d, true))
}
