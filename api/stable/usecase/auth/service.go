package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"webdesa/api/domain/auth"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/jwt"
	"webdesa/api/pkg/password"

	"github.com/google/uuid"
)

// Service implements authentication business logic.
// It accepts interfaces (Repository, UserRepository, EmailSender) and
// returns concrete structs. This follows the "accept interfaces, return
// structs" Go idiom.
type Service struct {
	repo            Repository
	userRepo        UserRepository
	email           EmailSender
	clock           clock.Clock
	jwtSecret       string
	resetTokenTTL   time.Duration
	rateLimitCount  int
	rateLimitWindow time.Duration
	resetURLBase    string
}

// NewService creates a new authentication service.
// Dependencies are injected via constructor following Clean Architecture principles.
//
// emailSender may be nil; in that case, password reset token generation still
// happens but no email is sent (legacy behavior). Callers should pass a real
// implementation in production. resetURLBase is the public URL prefix used
// to construct the reset link sent in the email (e.g., "https://desa.go.id").
func NewService(
	repo Repository,
	userRepo UserRepository,
	emailSender EmailSender,
	clk clock.Clock,
	jwtSecret string,
	resetTokenTTL time.Duration,
	resetURLBase string,
) *Service {
	return &Service{
		repo:            repo,
		userRepo:        userRepo,
		email:           emailSender,
		clock:           clk,
		jwtSecret:       jwtSecret,
		resetTokenTTL:   resetTokenTTL,
		rateLimitCount:  5, // 5 requests per hour per email
		rateLimitWindow: time.Hour,
		resetURLBase:    resetURLBase,
	}
}

// Login authenticates a user with email and password.
// Returns concrete TokenPair struct containing access and refresh tokens.
//
// Validates: Requirements 1.1, 1.2, 1.4
func (s *Service) Login(ctx context.Context, email, plainPassword string) (*auth.TokenPair, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password using bcrypt
	if !password.Verify(plainPassword, user.HashedPassword) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Parse user ID as UUID
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Generate access token (24 hours)
	accessToken, err := jwt.GenerateAccessToken(userID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (7 days)
	refreshToken, err := jwt.GenerateRefreshToken(userID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Calculate expiration time for access token (24 hours from now)
	expiresAt := s.clock.Now().Add(24 * time.Hour)

	return &auth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// RefreshToken validates a refresh token and generates a new access token.
// Returns concrete AccessToken struct.
//
// Validates: Requirements 2.1, 2.2, 2.3
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*auth.AccessToken, error) {
	// Validate refresh token
	claims, err := jwt.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Generate new access token (24 hours)
	accessToken, err := jwt.GenerateAccessToken(claims.UserID, claims.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Calculate expiration time for access token (24 hours from now)
	expiresAt := s.clock.Now().Add(24 * time.Hour)

	return &auth.AccessToken{
		Token:     accessToken,
		ExpiresAt: expiresAt,
	}, nil
}

// RequestPasswordReset initiates a password reset flow.
// Enforces rate limiting (5 requests per hour per email).
// Ensures only 1 active reset token per user at any time.
//
// Validates: Requirements 3.1, 3.2, 3.3, 3.4
func (s *Service) RequestPasswordReset(ctx context.Context, email, ipAddress, userAgent string) error {
	// Check rate limit: 5 requests per hour per email
	// Use UTC to ensure consistent timezone handling with database
	since := s.clock.Now().UTC().Add(-s.rateLimitWindow)
	count, err := s.repo.CountResetRequestsByEmail(ctx, email, since)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if count >= s.rateLimitCount {
		return fmt.Errorf("too many password reset requests, please try again later")
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not (security best practice)
		return nil
	}

	// Invalidate all existing reset tokens for this user
	// This ensures only 1 active reset token per user at any time
	if err := s.repo.InvalidateUserResetTokens(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to invalidate old tokens: %w", err)
	}

	// Generate secure random token
	token, err := generateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Calculate expiration time
	expiresAt := s.clock.Now().Add(s.resetTokenTTL)

	// Create reset token with metadata
	metadata := auth.ResetMetadata{
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := s.repo.CreateResetToken(ctx, user.ID, token, expiresAt, metadata); err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// Send reset email. If no email sender is wired, log and continue
	// (legacy fallback for setups without SMTP). When the sender IS wired
	// but the SMTP call fails, surface the error to the operator so they
	// know the user won't receive the link (was silently swallowed before
	// — see bugs.md#1).
	if s.email == nil {
		return nil
	}

	resetLink := fmt.Sprintf("%s/auth/password-reset/confirm?token=%s", s.resetURLBase, token)
	if err := s.email.SendPasswordResetEmail(ctx, user.Email, user.Name, resetLink); err != nil {
		// Token already persisted — leave it for retry; surface error.
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// CheckResetToken validates a reset token and returns its validity status.
// Returns concrete ResetTokenInfo struct.
//
// Validates: Requirements 3.5
func (s *Service) CheckResetToken(ctx context.Context, token string) (*auth.ResetTokenInfo, error) {
	// Find reset token
	resetToken, err := s.repo.FindResetToken(ctx, token)
	if err != nil {
		// Token not found or error
		return &auth.ResetTokenInfo{
			Valid:     false,
			ExpiresAt: nil,
		}, nil
	}

	// Check if token is already used
	if resetToken.Used {
		return &auth.ResetTokenInfo{
			Valid:     false,
			ExpiresAt: nil,
		}, nil
	}

	// Check if token is expired
	now := s.clock.Now()
	if now.After(resetToken.ExpiresAt) {
		return &auth.ResetTokenInfo{
			Valid:     false,
			ExpiresAt: &resetToken.ExpiresAt,
		}, nil
	}

	// Token is valid
	return &auth.ResetTokenInfo{
		Valid:     true,
		ExpiresAt: &resetToken.ExpiresAt,
	}, nil
}

// ResetPassword completes the password reset flow.
// Validates the reset token and updates the user's password.
//
// Validates: Requirements 3.6, 3.7
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Find reset token
	resetToken, err := s.repo.FindResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is already used
	if resetToken.Used {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is expired
	// Use UTC to ensure consistent timezone handling with database
	now := s.clock.Now().UTC()
	expiresAt := resetToken.ExpiresAt.UTC()

	if now.After(expiresAt) {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Hash the new password using bcrypt with cost factor 12
	hashedPassword, err := password.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user's password
	if err := s.repo.UpdatePassword(ctx, resetToken.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate the reset token (mark as used)
	if err := s.repo.InvalidateUserResetTokens(ctx, resetToken.UserID); err != nil {
		return fmt.Errorf("failed to invalidate reset token: %w", err)
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
