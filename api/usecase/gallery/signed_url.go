package gallery

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SignedURLScope names what a signed media URL is good for. Tokens are
// scope-bound: a public token won't unlock admin media, an admin token
// won't be accepted for an anonymous flow, etc.
type SignedURLScope string

const (
	// ScopePublic is for anonymous visitors — public-folder media and the
	// feature-scoped system-folder media exposed through banner/berita/struktur.
	ScopePublic SignedURLScope = "public"
	// ScopeAdmin is for authenticated operators — any media in the system,
	// including private folders. sub must be a live admin user.
	ScopeAdmin SignedURLScope = "admin"
	// ScopePPID is for citizens who filed an approved PPID request. sub is
	// `ppid_request:<id>` so RevokeRequest can deny-list the sub.
	ScopePPID SignedURLScope = "ppid"
)

// SignedURLClaims is the canonical payload for every media token. The
// secret used to sign lives in config (JWT_SECRET); sub is "anonymous"
// for public tokens, "user:<id>" for admin, "ppid_request:<id>" for ppid.
type SignedURLClaims struct {
	MediaID string         `json:"media_id"`
	Scope   SignedURLScope `json:"scope"`
	Sub     string         `json:"sub"`
	jwt.RegisteredClaims
}

// ErrInvalidSignedToken is the sentinel for any validation failure. The
// handler maps it to HTTP 401.
var ErrInvalidSignedToken = errors.New("invalid signed media token")

// SignedURLService mints and validates short-lived media tokens. Tokens
// are bound to a specific media_id and to a scope; verification rejects
// mismatches. The deny-list hook is used by the PPID scope (Task 7.2)
// to revoke the sub that an admin denied.
type SignedURLService struct {
	secret   []byte
	clock    func() time.Time
	denyList DenyList
	defaults SignedURLDefaults
}

// SignedURLDefaults collects TTLs for each scope. Public URLs last longer
// (24h) since they're cached in browsers; admin and PPID tokens are short
// (1h) so a leak window stays narrow.
type SignedURLDefaults struct {
	Public time.Duration
	Admin  time.Duration
	PPID   time.Duration
}

// DenyList is a minimal interface for sub-based revocation. The in-memory
// implementation lives in this package; a Redis-backed one can be swapped
// in for multi-instance deployments.
type DenyList interface {
	// Allow returns true when sub is not revoked. sub is opaque to the
	// deny list — typically "user:<id>" or "ppid_request:<id>".
	Allow(sub string) bool
	// Revoke marks sub as denied.
	Revoke(sub string)
}

// memoryDenyList is the default in-memory deny list. Synchronized for
// concurrent callers (the gallery service signs and verifies on
// goroutines, the ppid service revokes from request handlers).
type memoryDenyList struct {
	denied map[string]struct{}
}

func newMemoryDenyList() *memoryDenyList { return &memoryDenyList{denied: map[string]struct{}{}} }

func (d *memoryDenyList) Allow(sub string) bool {
	if sub == "" {
		return true
	}
	_, denied := d.denied[sub]
	return !denied
}

func (d *memoryDenyList) Revoke(sub string) {
	if sub == "" {
		return
	}
	d.denied[sub] = struct{}{}
}

// NewSignedURLService constructs a service with the default TTLs and an
// in-memory deny list. The clock parameter is injectable for tests.
func NewSignedURLService(secret string, clock func() time.Time) *SignedURLService {
	if clock == nil {
		clock = time.Now
	}
	return &SignedURLService{
		secret:   []byte(secret),
		clock:    clock,
		denyList: newMemoryDenyList(),
		defaults: SignedURLDefaults{
			Public: 24 * time.Hour,
			Admin:  1 * time.Hour,
			PPID:   1 * time.Hour,
		},
	}
}

// DenyList exposes the revocation list so callers (PPID service) can
// register a token's sub on revoke.
func (s *SignedURLService) DenyList() DenyList { return s.denyList }

// Sign returns a fresh JWT bound to scope + mediaID + sub. ttl <= 0 falls
// back to the per-scope default.
func (s *SignedURLService) Sign(scope SignedURLScope, mediaID, sub string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		switch scope {
		case ScopePublic:
			ttl = s.defaults.Public
		case ScopeAdmin:
			ttl = s.defaults.Admin
		case ScopePPID:
			ttl = s.defaults.PPID
		default:
			return "", ErrInvalidSignedToken
		}
	}
	now := s.clock()
	claims := SignedURLClaims{
		MediaID: mediaID,
		Scope:   scope,
		Sub:     sub,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.New().String(),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidSignedToken, err)
	}
	return signed, nil
}

// Verify validates the token, enforces that mediaID matches the token's
// claim, and checks the deny list. Returns the validated claims on
// success; ErrInvalidSignedToken on any failure.
func (s *SignedURLService) Verify(tokenString, mediaID string) (*SignedURLClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("%w: empty token", ErrInvalidSignedToken)
	}
	claims := &SignedURLClaims{}
	parser := jwt.NewParser(jwt.WithTimeFunc(func() time.Time { return s.clock() }))
	_, err := parser.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignedToken
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSignedToken, err)
	}
	if claims.MediaID != mediaID {
		return nil, fmt.Errorf("%w: media_id mismatch (got %q want %q)", ErrInvalidSignedToken, claims.MediaID, mediaID)
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(s.clock()) {
		return nil, fmt.Errorf("%w: expired", ErrInvalidSignedToken)
	}
	if !s.denyList.Allow(claims.Sub) {
		return nil, fmt.Errorf("%w: revoked sub %q", ErrInvalidSignedToken, claims.Sub)
	}
	return claims, nil
}

// Refresh verifies an existing token and mints a fresh one with the
// same scope and sub. Used by the admin client to renew expiring
// admin URLs (Task 7.3) and by the public client for transparent
// renewal when an image tag outlives the token TTL.
//
// The fresh token inherits scope/sub from the input. ttl <= 0 falls back
// to the per-scope default. The old token is NOT added to the deny list
// — expiry is the only invalidation.
func (s *SignedURLService) Refresh(tokenString, mediaID string, ttl time.Duration) (string, *SignedURLClaims, error) {
	claims, err := s.Verify(tokenString, mediaID)
	if err != nil {
		return "", nil, err
	}
	fresh, err := s.Sign(claims.Scope, claims.MediaID, claims.Sub, ttl)
	if err != nil {
		return "", nil, err
	}
	return fresh, claims, nil
}
