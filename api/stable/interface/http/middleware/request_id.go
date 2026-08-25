package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIDKey string

const RequestIDKey requestIDKey = "request_id"

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate unique request ID
			requestID := uuid.New().String()

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// Process request with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}
