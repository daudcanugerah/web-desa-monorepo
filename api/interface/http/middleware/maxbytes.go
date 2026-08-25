package middleware

import (
	"net/http"
)

// DefaultMaxRequestBodyBytes is the default global cap on incoming request
// bodies. 60 MiB leaves headroom above the largest configured upload
// (50 MiB documents per [fileupload].max_document_size_mb) while still
// preventing unbounded body reads.
const DefaultMaxRequestBodyBytes int64 = 60 << 20

// MaxBytesMiddleware caps the size of every incoming request body. Requests
// exceeding the limit receive 413 Request Entity Too Large from the
// underlying http.MaxBytesHandler. A non-positive limit falls back to
// DefaultMaxRequestBodyBytes.
//
// Per-route upload handlers should still apply their own tighter limits via
// http.MaxBytesReader — this middleware is the floor, not the ceiling.
func MaxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxRequestBodyBytes
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
