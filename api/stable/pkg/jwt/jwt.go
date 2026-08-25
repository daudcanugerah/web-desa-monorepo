package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// Generate generates a new JWT token
func Generate(userID uuid.UUID, email string, secret string, expiration time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateAccessToken generates a new access token with 24-hour expiration
func GenerateAccessToken(userID uuid.UUID, email string, secret string) (string, error) {
	return Generate(userID, email, secret, 24*time.Hour)
}

// GenerateRefreshToken generates a new refresh token with 7-day expiration
func GenerateRefreshToken(userID uuid.UUID, email string, secret string) (string, error) {
	return Generate(userID, email, secret, 7*24*time.Hour)
}

// Validate validates a JWT token and returns the claims
func Validate(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateToken is an alias for Validate for consistency with the spec
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	return Validate(tokenString, secret)
}
