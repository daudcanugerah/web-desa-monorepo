package middleware

import (
	"webdesa/api/pkg/response"
	"fmt"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log panic with stack trace
					stack := debug.Stack()
					logger.Error(r.Context(), "panic recovered",
						"error", fmt.Sprintf("%v", err),
						"stack", string(stack),
						"method", r.Method,
						"path", r.URL.Path,
					)

					// Return 500 error
					response.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
