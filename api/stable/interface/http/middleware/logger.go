package middleware

import (
	"context"
	"net/http"
	"time"
)

// Logger defines the minimal logging interface needed by the middleware.
// This interface is defined here (where it's used) rather than in the logger package,
// following Go best practices for interface design.
type Logger interface {
	Info(ctx context.Context, msg string, keyvalues ...any)
	Error(ctx context.Context, msg string, keyvalues ...any)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	written     int64
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request details
			duration := time.Since(start)

			// Get request ID from context if available
			logFields := []interface{}{
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"bytes", wrapped.written,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			}

			if requestID, ok := GetRequestIDFromContext(r.Context()); ok {
				logFields = append(logFields, "request_id", requestID)
			}

			logger.Info(r.Context(), "HTTP request", logFields...)
		})
	}
}
