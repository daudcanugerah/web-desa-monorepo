package middleware

import (
	"net/http"

	"github.com/casbin/casbin/v2"
)

// MiddlewareDeps bundles the dependencies needed to wire HTTP middleware
// in per-feature route registration functions.
//
// Pass an instance of this to each Register<Feature>Routes call so that
// feature files don't need to import individual middleware functions.
type MiddlewareDeps struct {
	Enforcer  *casbin.Enforcer
	JWTSecret string
}

// RBAC returns a middleware that enforces a Casbin (resource, action) policy.
func (m *MiddlewareDeps) RBAC(resource, action string) func(http.Handler) http.Handler {
	return RBACMiddleware(m.Enforcer, resource, action)
}

// Auth returns the JWT authentication middleware.
func (m *MiddlewareDeps) Auth() func(http.Handler) http.Handler {
	return AuthMiddleware(m.JWTSecret)
}

// ProtectedRateLimit returns the per-user 30 req/min bucket.
func (m *MiddlewareDeps) ProtectedRateLimit() func(http.Handler) http.Handler {
	return ProtectedRateLimitMiddleware()
}

// AuthRateLimit returns the per-IP 5 req/min bucket (used on /auth/*).
func (m *MiddlewareDeps) AuthRateLimit() func(http.Handler) http.Handler {
	return AuthRateLimitMiddleware()
}

// PublicRateLimit returns the per-IP 100 req/min bucket (used on /public/*).
func (m *MiddlewareDeps) PublicRateLimit() func(http.Handler) http.Handler {
	return PublicRateLimitMiddleware()
}

// RouteScope selects which set of routes a Register<Feature>Routes function
// mounts. Public routes are mounted on the public sub-router (no auth);
// Protected routes are mounted on the auth+RBAC sub-router.
type RouteScope string

const (
	// ScopePublic mounts only the /public/* paths (no JWT required).
	ScopePublic RouteScope = "public"

	// ScopeProtected mounts only the admin paths (JWT + RBAC required).
	// The caller must have already applied mw.Auth() and mw.ProtectedRateLimit().
	ScopeProtected RouteScope = "protected"
)