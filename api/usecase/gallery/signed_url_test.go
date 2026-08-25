package gallery

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSignVerifyHappyPath(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopePublic, "media-1", "anonymous", 0)
	if err != nil { t.Fatalf("sign err: %v", err) }
	require.NoError(t, err)
	assert.NotEmpty(t, tok)

	claims, err := svc.Verify(tok, "media-1")
	require.NoError(t, err)
	assert.Equal(t, "media-1", claims.MediaID)
	assert.Equal(t, ScopePublic, claims.Scope)
	assert.Equal(t, "anonymous", claims.Sub)
	assert.Equal(t, now.Unix(), claims.IssuedAt.Unix())
	assert.Equal(t, now.Add(24*time.Hour).Unix(), claims.ExpiresAt.Unix())
}

func TestVerifyExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopePPID, "media-1", "ppid_request:r1", 1*time.Hour)
	require.NoError(t, err)

	// 1 hour + 1 second later — token must be rejected
	later := now.Add(1*time.Hour + 1*time.Second)
	svc.clock = func() time.Time { return later }

	_, err = svc.Verify(tok, "media-1")
	require.ErrorIs(t, err, ErrInvalidSignedToken)
}

func TestVerifyMediaIDMismatch(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopePublic, "media-1", "anonymous", 0)
	if err != nil { t.Fatalf("sign err: %v", err) }
	require.NoError(t, err)

	_, err = svc.Verify(tok, "media-2")
	require.ErrorIs(t, err, ErrInvalidSignedToken)
}

func TestVerifyTamperedSignature(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopePublic, "media-1", "anonymous", 0)
	if err != nil { t.Fatalf("sign err: %v", err) }
	require.NoError(t, err)

	// Flip the last char of the signature segment
	parts := strings.Split(tok, ".")
	require.Len(t, parts, 3)
	tampered := parts[0] + "." + parts[1] + "." + parts[2][:len(parts[2])-1] + "A"

	_, err = svc.Verify(tampered, "media-1")
	require.ErrorIs(t, err, ErrInvalidSignedToken)
}

func TestVerifyDenyList(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopePPID, "media-1", "ppid_request:r1", 1*time.Hour)
	require.NoError(t, err)

	// Verify works before revoke
	_, err = svc.Verify(tok, "media-1")
	require.NoError(t, err)

	// Revoke the sub and verify is rejected even before expiry
	svc.DenyList().Revoke("ppid_request:r1")
	_, err = svc.Verify(tok, "media-1")
	require.ErrorIs(t, err, ErrInvalidSignedToken)
}

func TestSignInvalidScope(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	_, err := svc.Sign(SignedURLScope("garbage"), "m", "s", 0)
	require.ErrorIs(t, err, ErrInvalidSignedToken)
}

func TestVerifyEmptyToken(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	_, err := svc.Verify("", "m")
	require.ErrorIs(t, err, ErrInvalidSignedToken)
}

func TestVerifyWrongSecret(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	signer := NewSignedURLService("secret-A", fixedClock(now))
	verifier := NewSignedURLService("secret-B", fixedClock(now))
	tok, err := signer.Sign(ScopePublic, "m", "anonymous", 0)
	require.NoError(t, err)

	_, err = verifier.Verify(tok, "m")
	require.ErrorIs(t, err, ErrInvalidSignedToken)
}

// Ensures the Sign path produces a token whose alg is HMAC-SHA256 and
// that the header round-trips through jwt.NewParser.
func TestTokenAlgAndHeader(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopeAdmin, "m", "user:abc", 0)
	require.NoError(t, err)

	parser := jwt.NewParser()
	parsed, _, err := parser.ParseUnverified(tok, &SignedURLClaims{})
	require.NoError(t, err)
	assert.Equal(t, "HS256", parsed.Method.Alg())
}

// Refresh issues a new token carrying the same scope + sub.
func TestRefreshHappyPath(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	old, err := svc.Sign(ScopeAdmin, "m1", "user:xyz", 0)
	require.NoError(t, err)

	fresh, claims, err := svc.Refresh(old, "m1", 1*time.Hour)
	require.NoError(t, err)
	assert.NotEqual(t, old, fresh, "Refresh must mint a distinct token (new jti)")
	assert.Equal(t, "m1", claims.MediaID)
	assert.Equal(t, ScopeAdmin, claims.Scope)
	assert.Equal(t, "user:xyz", claims.Sub)

	// The fresh token must verify.
	_, err = svc.Verify(fresh, "m1")
	require.NoError(t, err)
}

// Refresh with a mismatched media id fails Verify (Refresh requires
// the input token to currently be valid against the supplied media id).
func TestRefreshMediaIDMismatch(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopeAdmin, "m1", "user:x", 0)
	require.NoError(t, err)

	_, _, err = svc.Refresh(tok, "different-media", 0)
	assert.ErrorIs(t, err, ErrInvalidSignedToken)
}

// Refresh after expiry fails.
func TestRefreshAfterExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopeAdmin, "m1", "user:x", 1*time.Hour)
	require.NoError(t, err)

	svc.clock = func() time.Time { return now.Add(2 * time.Hour) }
	_, _, err = svc.Refresh(tok, "m1", 0)
	assert.ErrorIs(t, err, ErrInvalidSignedToken)
}

// Refresh after revoke fails — PPID deny list enforcement.
func TestRefreshAfterRevoke(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := NewSignedURLService("test-secret-please-change-this-is-long-enough", fixedClock(now))
	tok, err := svc.Sign(ScopePPID, "m1", "ppid_request:r1", 0)
	require.NoError(t, err)
	svc.DenyList().Revoke("ppid_request:r1")

	_, _, err = svc.Refresh(tok, "m1", 0)
	assert.ErrorIs(t, err, ErrInvalidSignedToken)
}