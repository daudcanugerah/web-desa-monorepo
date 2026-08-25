package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORSMiddleware configures CORS
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	// Default to allow all if no origins specified
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})
}
