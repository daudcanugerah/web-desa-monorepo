package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin_ValidCredentials(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user
	email := "test@example.com"
	password := "password123"
	ts.CreateTestUser(t, "Test User", email, password)

	// Attempt login
	loginReq := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/login", loginReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	// LoginResponse: access_token, refresh_token, expires_at
	assert.NotEmpty(t, data["access_token"], "access_token must not be empty")
	assert.NotEmpty(t, data["refresh_token"], "refresh_token must not be empty")
	assert.NotEmpty(t, data["expires_at"], "expires_at must not be empty")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user
	email := "test@example.com"
	password := "password123"
	ts.CreateTestUser(t, "Test User", email, password)

	// Attempt login with wrong password
	loginReq := map[string]interface{}{
		"email":    email,
		"password": "wrongpassword",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/login", loginReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "Invalid credentials", body["error"])
}

func TestLogin_NonExistentUser(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Attempt login with non-existent user
	loginReq := map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/login", loginReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "Invalid credentials", body["error"])
}

func TestLogin_MissingFields(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	tests := []struct {
		name     string
		request  map[string]interface{}
		expected string
	}{
		{
			name:     "missing email",
			request:  map[string]interface{}{"password": "password123"},
			expected: "Email is required",
		},
		{
			name:     "missing password",
			request:  map[string]interface{}{"email": "test@example.com"},
			expected: "Password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ts.MakeRequest("POST", "/api/v1/auth/login", tt.request, "")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			body, err := parseJSONResponse(resp.Body)
			require.NoError(t, err)

			assert.False(t, body["success"].(bool))
			assert.Equal(t, tt.expected, body["error"])
		})
	}
}

func TestRefreshToken_ValidToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user and login
	email := "test@example.com"
	password := "password123"
	ts.CreateTestUser(t, "Test User", email, password)

	tokenPair, err := ts.AuthService.Login(context.Background(), email, password)
	require.NoError(t, err)

	// Refresh token
	refreshReq := map[string]interface{}{
		"refresh_token": tokenPair.RefreshToken,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/refresh", refreshReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	// RefreshTokenResponse: access_token, expires_at
	assert.NotEmpty(t, data["access_token"], "access_token must not be empty")
	assert.NotEmpty(t, data["expires_at"], "expires_at must not be empty")
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Attempt refresh with invalid token
	refreshReq := map[string]interface{}{
		"refresh_token": "invalid-token",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/refresh", refreshReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Equal(t, "Invalid or expired refresh token", body["error"])
}

func TestPasswordReset_RequestFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user
	email := "test@example.com"
	password := "password123"
	ts.CreateTestUser(t, "Test User", email, password)

	// Request password reset
	resetReq := map[string]interface{}{
		"email": email,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/request", resetReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Contains(t, data["message"], "password reset link has been sent")
}

func TestPasswordReset_NonExistentEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Request password reset for non-existent email
	resetReq := map[string]interface{}{
		"email": "nonexistent@example.com",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/request", resetReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response - should return success for security (don't reveal if email exists)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
}

func TestPasswordReset_RateLimit(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user
	email := "test@example.com"
	password := "password123"
	ts.CreateTestUser(t, "Test User", email, password)

	resetReq := map[string]interface{}{
		"email": email,
	}

	// Make 5 requests (should succeed)
	for i := 0; i < 5; i++ {
		resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/request", resetReq, "")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// 6th request should be rate limited
	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/request", resetReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Contains(t, body["error"], "too many password reset requests")
}

func TestPasswordReset_CheckToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user
	email := "test@example.com"
	password := "password123"
	userID := ts.CreateTestUser(t, "Test User", email, password)

	// Create reset token directly in database
	token := "test-reset-token-123"
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	_, err := ts.DB.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, token, expiresAt, "127.0.0.1", "test-agent")
	require.NoError(t, err)

	// Check valid token
	resp, err := ts.MakeRequest("GET", "/api/v1/auth/password-reset/check?token="+token, nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	// PasswordResetCheckResponse: valid, expires_at
	assert.True(t, data["valid"].(bool), "valid must be true for a valid token")
	assert.NotEmpty(t, data["expires_at"], "expires_at must not be empty for a valid token")
}

func TestPasswordReset_CheckInvalidToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Check invalid token
	resp, err := ts.MakeRequest("GET", "/api/v1/auth/password-reset/check?token=invalid-token", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	// PasswordResetCheckResponse: valid=false, expires_at absent/nil
	assert.False(t, data["valid"].(bool), "valid must be false for an invalid token")
	assert.Nil(t, data["expires_at"], "expires_at must be absent for an invalid token")
}

func TestPasswordReset_ConfirmFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user
	email := "test@example.com"
	password := "password123"
	userID := ts.CreateTestUser(t, "Test User", email, password)

	// Create reset token directly in database
	token := "test-reset-token-456"
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	_, err := ts.DB.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, token, expiresAt, "127.0.0.1", "test-agent")
	require.NoError(t, err)

	// Confirm password reset
	newPassword := "newpassword456"
	confirmReq := map[string]interface{}{
		"token":        token,
		"new_password": newPassword,
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/confirm", confirmReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.True(t, body["success"].(bool))
	data := getDataField(body)
	assert.Contains(t, data["message"], "Password has been reset successfully")

	// Verify can login with new password
	loginReq := map[string]interface{}{
		"email":    email,
		"password": newPassword,
	}

	loginResp, err := ts.MakeRequest("POST", "/api/v1/auth/login", loginReq, "")
	require.NoError(t, err)
	defer loginResp.Body.Close()

	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Verify cannot login with old password
	oldLoginReq := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	oldLoginResp, err := ts.MakeRequest("POST", "/api/v1/auth/login", oldLoginReq, "")
	require.NoError(t, err)
	defer oldLoginResp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, oldLoginResp.StatusCode)
}

func TestPasswordReset_ConfirmExpiredToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user
	email := "test@example.com"
	password := "password123"
	userID := ts.CreateTestUser(t, "Test User", email, password)

	// Create expired reset token
	token := "expired-token-789"
	expiresAt := time.Now().UTC().Add(-1 * time.Hour) // Expired 1 hour ago (UTC)
	_, err := ts.DB.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, token, expiresAt, "127.0.0.1", "test-agent")
	require.NoError(t, err)

	// Attempt to confirm with expired token
	confirmReq := map[string]interface{}{
		"token":        token,
		"new_password": "newpassword789",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/confirm", confirmReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.Contains(t, body["error"], "Invalid or expired reset token")
}

// Bad Path Tests - Error Scenarios

func TestLogin_EmptyEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	loginReq := map[string]interface{}{
		"email":    "",
		"password": "password123",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/login", loginReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
	assert.NotEmpty(t, body["error"])
}

func TestLogin_EmptyPassword(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	loginReq := map[string]interface{}{
		"email":    "test@example.com",
		"password": "",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/login", loginReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
}

func TestLogin_InvalidEmailFormat(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	loginReq := map[string]interface{}{
		"email":    "not-an-email",
		"password": "password123",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/login", loginReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
}

func TestRefreshToken_MissingToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	refreshReq := map[string]interface{}{
		"refresh_token": "",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/refresh", refreshReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
}

func TestRefreshToken_MalformedToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	refreshReq := map[string]interface{}{
		"refresh_token": "not.a.valid.jwt.token",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/refresh", refreshReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
}

func TestPasswordReset_InvalidEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resetReq := map[string]interface{}{
		"email": "not-an-email",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/request", resetReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
}

func TestPasswordReset_EmptyEmail(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resetReq := map[string]interface{}{
		"email": "",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/request", resetReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
}

func TestPasswordReset_CheckMissingToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	resp, err := ts.MakeRequest("GET", "/api/v1/auth/password-reset/check?token=", nil, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPasswordReset_ConfirmMissingToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	confirmReq := map[string]interface{}{
		"token":        "",
		"new_password": "newpassword",
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/confirm", confirmReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
}

func TestPasswordReset_ConfirmWeakPassword(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test user
	email := "test@example.com"
	password := "password123"
	userID := ts.CreateTestUser(t, "Test User", email, password)

	// Create reset token
	token := "test-token-weak"
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	_, err := ts.DB.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, token, expiresAt, "127.0.0.1", "test-agent")
	require.NoError(t, err)

	// Attempt to confirm with weak password
	confirmReq := map[string]interface{}{
		"token":        token,
		"new_password": "123", // Too weak
	}

	resp, err := ts.MakeRequest("POST", "/api/v1/auth/password-reset/confirm", confirmReq, "")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)

	assert.False(t, body["success"].(bool))
}
