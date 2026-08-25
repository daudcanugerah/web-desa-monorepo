package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"
	expiration := time.Hour

	token, err := Generate(userID, email, secret, expiration)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidate_ValidToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"
	expiration := time.Hour

	token, err := Generate(userID, email, secret, expiration)
	require.NoError(t, err)

	claims, err := Validate(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestValidate_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"
	expiration := -time.Hour // Expired token

	token, err := Generate(userID, email, secret, expiration)
	require.NoError(t, err)

	_, err = Validate(token, secret)
	assert.Error(t, err)
}

func TestValidate_InvalidSecret(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"
	wrongSecret := "wrong-secret"
	expiration := time.Hour

	token, err := Generate(userID, email, secret, expiration)
	require.NoError(t, err)

	_, err = Validate(token, wrongSecret)
	assert.Error(t, err)
}

func TestValidate_MalformedToken(t *testing.T) {
	secret := "test-secret"
	malformedToken := "not.a.valid.token"

	_, err := Validate(malformedToken, secret)
	assert.Error(t, err)
}

func TestValidate_EmptyToken(t *testing.T) {
	secret := "test-secret"

	_, err := Validate("", secret)
	assert.Error(t, err)
}

func TestGenerateAccessToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"

	token, err := GenerateAccessToken(userID, email, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := ValidateToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)

	// Check expiration is approximately 24 hours
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	timeDiff := actualExpiry.Sub(expectedExpiry)
	assert.Less(t, timeDiff.Abs(), 5*time.Second, "Expiration should be approximately 24 hours from now")
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"

	token, err := GenerateRefreshToken(userID, email, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := ValidateToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)

	// Check expiration is approximately 7 days
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	timeDiff := actualExpiry.Sub(expectedExpiry)
	assert.Less(t, timeDiff.Abs(), 5*time.Second, "Expiration should be approximately 7 days from now")
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"

	// Test with access token
	accessToken, err := GenerateAccessToken(userID, email, secret)
	require.NoError(t, err)

	claims, err := ValidateToken(accessToken, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)

	// Test with refresh token
	refreshToken, err := GenerateRefreshToken(userID, email, secret)
	require.NoError(t, err)

	claims, err = ValidateToken(refreshToken, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}
