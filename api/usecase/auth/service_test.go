package auth

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"webdesa/api/domain/auth"
	"webdesa/api/domain/user"
	"webdesa/api/pkg/clock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubEmailSender records calls and can be configured to fail.
type stubEmailSender struct {
	calls    int32
	failWith error
	gotEmail string
	gotName  string
	gotLink  string
}

func (s *stubEmailSender) SendPasswordResetEmail(_ context.Context, email, name, link string) error {
	atomic.AddInt32(&s.calls, 1)
	s.gotEmail = email
	s.gotName = name
	s.gotLink = link
	return s.failWith
}

type stubUserRepo struct {
	user *user.User
}

type authUser struct {
	id, email, name string
}

func newUser(id, email, name string) *user.User {
	return &user.User{ID: id, Email: email, Name: name}
}

func (r *stubUserRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	if r.user != nil && r.user.ID == id {
		return r.user, nil
	}
	return nil, errors.New("not found")
}

func (r *stubUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if r.user != nil && r.user.Email == email {
		return r.user, nil
	}
	return nil, errors.New("not found")
}

type stubResetRepo struct {
	count   int
	token   string
	expires time.Time
}

func (r *stubResetRepo) CreateResetToken(_ context.Context, _ string, tok string, exp time.Time, _ auth.ResetMetadata) error {
	r.token = tok
	r.expires = exp
	return nil
}
func (r *stubResetRepo) FindResetToken(_ context.Context, _ string) (*auth.ResetToken, error) {
	return nil, errors.New("not found")
}
func (r *stubResetRepo) InvalidateUserResetTokens(_ context.Context, _ string) error { return nil }
func (r *stubResetRepo) UpdatePassword(_ context.Context, _ string, _ string) error   { return nil }
func (r *stubResetRepo) CountResetRequestsByEmail(_ context.Context, _ string, _ time.Time) (int, error) {
	return r.count, nil
}

func newService(t *testing.T, repo *stubResetRepo, userRepo *stubUserRepo, email EmailSender, baseURL string) *Service {
	t.Helper()
	clk := clock.NewFixedClock(time.Now())
	s := NewService(repo, userRepo, email, clk, "test-secret", time.Hour, baseURL)
	return s
}

func TestRequestPasswordReset_EmailSentOnSuccess(t *testing.T) {
	resetRepo := &stubResetRepo{count: 0}
	userRepo := &stubUserRepo{user: newUser(uuid.NewString(), "alice@desa.id", "Alice")}
	email := &stubEmailSender{}

	s := newService(t, resetRepo, userRepo, email, "https://desa.id")

	err := s.RequestPasswordReset(context.Background(), "alice@desa.id", "127.0.0.1", "test-agent")
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&email.calls))
	assert.Equal(t, "alice@desa.id", email.gotEmail)
	assert.Equal(t, "Alice", email.gotName)
	assert.Contains(t, email.gotLink, "https://desa.id/auth/password-reset/confirm?token=")
	assert.Equal(t, 64, len(resetRepo.token)) // hex-encoded 32 bytes
}

func TestRequestPasswordReset_EmailFailureSurfacesError(t *testing.T) {
	// Regression test for bugs.md#1 — previously the handler swallowed
	// the email error and returned 200. Now it surfaces to the operator.
	resetRepo := &stubResetRepo{count: 0}
	userRepo := &stubUserRepo{user: newUser(uuid.NewString(), "bob@desa.id", "Bob")}
	email := &stubEmailSender{failWith: errors.New("smtp connection refused")}

	s := newService(t, resetRepo, userRepo, email, "https://desa.id")

	err := s.RequestPasswordReset(context.Background(), "bob@desa.id", "127.0.0.1", "test-agent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send password reset email")
	assert.Contains(t, err.Error(), "smtp connection refused")
}

func TestRequestPasswordReset_NoEmailSender_LegacyBehavior(t *testing.T) {
	// When no email sender is wired, the service silently succeeds
	// (legacy fallback for setups without SMTP). Token still persisted.
	resetRepo := &stubResetRepo{count: 0}
	userRepo := &stubUserRepo{user: newUser(uuid.NewString(), "carol@desa.id", "Carol")}

	s := newService(t, resetRepo, userRepo, nil, "https://desa.id")

	err := s.RequestPasswordReset(context.Background(), "carol@desa.id", "127.0.0.1", "test-agent")
	require.NoError(t, err)
	assert.NotEmpty(t, resetRepo.token)
}

func TestRequestPasswordReset_RateLimit(t *testing.T) {
	resetRepo := &stubResetRepo{count: 5}
	userRepo := &stubUserRepo{user: newUser(uuid.NewString(), "dave@desa.id", "Dave")}
	s := newService(t, resetRepo, userRepo, nil, "https://desa.id")

	err := s.RequestPasswordReset(context.Background(), "dave@desa.id", "127.0.0.1", "test-agent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many")
}