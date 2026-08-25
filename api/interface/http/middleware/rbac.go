package middleware

import (
	"webdesa/api/pkg/response"
	"context"
	"net/http"

	"github.com/casbin/casbin/v2"
)

// RBACMiddleware checks permissions using Casbin
func RBACMiddleware(enforcer *casbin.Enforcer, resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user roles from context (set by auth middleware)
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "user not authenticated")
				return
			}

			// Get user roles from Casbin
			roles, err := enforcer.GetRolesForUser(userID)
			if err != nil {
				response.Error(w, http.StatusInternalServerError, "failed to get user roles")
				return
			}

			// Check if any role has permission
			hasPermission := false
			for _, role := range roles {
				allowed, err := enforcer.Enforce(role, resource, action)
				if err != nil {
					response.Error(w, http.StatusInternalServerError, "failed to check permissions")
					return
				}
				if allowed {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				response.Error(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RBACMiddlewareWithContext creates RBAC middleware that extracts resource/action from context
func RBACMiddlewareWithContext(enforcer *casbin.Enforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resource := r.Context().Value("rbac_resource")
			action := r.Context().Value("rbac_action")

			if resource == nil || action == nil {
				response.Error(w, http.StatusInternalServerError, "RBAC configuration error")
				return
			}

			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "user not authenticated")
				return
			}

			roles, err := enforcer.GetRolesForUser(userID)
			if err != nil {
				response.Error(w, http.StatusInternalServerError, "failed to get user roles")
				return
			}

			hasPermission := false
			for _, role := range roles {
				allowed, err := enforcer.Enforce(role, resource.(string), action.(string))
				if err != nil {
					response.Error(w, http.StatusInternalServerError, "failed to check permissions")
					return
				}
				if allowed {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				response.Error(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithRBAC adds RBAC resource and action to context
func WithRBAC(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "rbac_resource", resource)
			ctx = context.WithValue(ctx, "rbac_action", action)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
