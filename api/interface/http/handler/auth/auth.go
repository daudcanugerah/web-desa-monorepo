package auth

import (
	"net/http"
	"strings"

	"webdesa/api/pkg/response"
	"webdesa/api/usecase/auth"

	"github.com/ggicci/httpin"

	"webdesa/api/pkg/handlerutil")

// httpin directives for request binding
// See: https://github.com/ggicci/httpin

// AuthHandler handles HTTP requests for authentication operations.
// It accepts the auth service from the usecase layer as a dependency.
// Dependencies point inward: interface/http → usecase → domain
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new authentication handler with service dependency injected.
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse represents the response for successful login
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"` // ISO 8601 format
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the response for successful token refresh
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"` // ISO 8601 format
}

// PasswordResetRequestRequest represents the request body for password reset request
type PasswordResetRequestRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetCheckResponse represents the response for reset token check
type PasswordResetCheckResponse struct {
	Valid     bool    `json:"valid"`
	ExpiresAt *string `json:"expires_at,omitempty"` // ISO 8601 format, only if valid
}

// PasswordResetConfirmRequest represents the request body for password reset confirmation
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// MessageResponse represents a generic message response
type MessageResponse struct {
	Message string `json:"message"`
}

// Login godoc
// @Summary      User login
// @Description  Authenticate with email + password and receive JWT access (24h) + refresh (7d) tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "Login credentials"
// @Success      200      {object} auth.LoginResponse
// @Failure      400      {object} response.ErrorResponse "Invalid request body"
// @Failure      401      {object} response.ErrorResponse "Invalid credentials"
// @Failure      429      {object} response.ErrorResponse "Rate limit exceeded"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required,min=6"`
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

	// Call auth service
	tokenPair, err := h.authService.Login(r.Context(), req.Payload.Email, req.Payload.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Return tokens with expiration in ISO 8601 format
	resp := LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token for a new access token (24h).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "Refresh token"
// @Success      200      {object} auth.RefreshTokenResponse
// @Failure      400      {object} response.ErrorResponse "Invalid request body"
// @Failure      401      {object} response.ErrorResponse "Invalid or expired refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			RefreshToken string `json:"refresh_token" validate:"required"`
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

	// Call auth service
	accessToken, err := h.authService.RefreshToken(r.Context(), req.Payload.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// Return new access token with expiration in ISO 8601 format
	resp := RefreshTokenResponse{
		AccessToken: accessToken.Token,
		ExpiresAt:   accessToken.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, resp)
}

// RequestPasswordReset godoc
// @Summary      Request password reset
// @Description  Initiate the password-reset flow for the given email. Rate-limited to 5 requests per hour per email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "Email address"
// @Success      200      {object} auth.MessageResponse "Reset link sent (if email exists)"
// @Failure      400      {object} response.ErrorResponse "Invalid request body"
// @Failure      429      {object} response.ErrorResponse "Too many reset requests"
// @Router       /auth/password-reset/request [post]
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			Email string `json:"email" validate:"required,email"`
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

	// Extract IP address and user agent for security auditing
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	// Call auth service
	err := h.authService.RequestPasswordReset(r.Context(), req.Payload.Email, ipAddress, userAgent)
	if err != nil {
		// Check if rate limit exceeded
		if err.Error() == "too many password reset requests, please try again later" {
			response.Error(w, http.StatusTooManyRequests, err.Error())
			return
		}
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to process password reset request", err)
		return
	}

	// Return success (don't reveal if email exists or not for security)
	response.Success(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// CheckResetToken godoc
// @Summary      Validate password reset token
// @Description  Return whether the reset token is still valid and its expiry.
// @Tags         auth
// @Produce      json
// @Param        token  query     string  true  "Reset token"
// @Success      200    {object} auth.PasswordResetCheckResponse
// @Failure      400    {object} response.ErrorResponse "Token missing"
// @Router       /auth/password-reset/check [get]
func (h *AuthHandler) CheckResetToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `in:"query=token" validate:"required"`
	}

	if err := httpin.DecodeTo(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Token is required")
		return
	}

	// Validate input
	if err := handlerutil.ValidateStruct(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call auth service
	tokenInfo, err := h.authService.CheckResetToken(r.Context(), req.Token)
	if err != nil {
		response.ErrorWithDetails(w, http.StatusInternalServerError, "Failed to check reset token", err)
		return
	}

	// Build response
	resp := PasswordResetCheckResponse{
		Valid: tokenInfo.Valid,
	}

	// Include expiration time in ISO 8601 format if token is valid
	if tokenInfo.Valid && tokenInfo.ExpiresAt != nil {
		expiresAtStr := tokenInfo.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ExpiresAt = &expiresAtStr
	}

	response.Success(w, http.StatusOK, resp)
}

// ConfirmPasswordReset godoc
// @Summary      Confirm password reset
// @Description  Apply a new password using a valid reset token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request    body      map[string]interface{}  true  "Reset token + new password"
// @Success      200      {object} auth.MessageResponse "Password reset"
// @Failure      400      {object} response.ErrorResponse "Invalid or expired token"
// @Router       /auth/password-reset/confirm [post]
func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload struct {
			Token       string `json:"token" validate:"required"`
			NewPassword string `json:"new_password" validate:"required,min=6"`
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

	// Call auth service
	err := h.authService.ResetPassword(r.Context(), req.Payload.Token, req.Payload.NewPassword)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid or expired reset token")
		return
	}

	// Return success
	response.Success(w, http.StatusOK, map[string]string{
		"message": "Password has been reset successfully",
	})
}

// isValidEmail performs basic email format validation.
// Checks for presence of '@' and a '.' after the '@'.
func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at < 1 {
		return false
	}
	dot := strings.LastIndex(email[at:], ".")
	return dot > 1 && at+dot < len(email)-1
}
