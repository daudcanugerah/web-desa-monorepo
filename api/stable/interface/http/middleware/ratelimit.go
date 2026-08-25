package middleware

import (
	"webdesa/api/pkg/response"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// tokenBucket implements the token bucket algorithm
type tokenBucket struct {
	tokens         float64
	capacity       float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
	mu             sync.Mutex
}

// newTokenBucket creates a new token bucket
func newTokenBucket(requestsPerMinute int) *tokenBucket {
	capacity := float64(requestsPerMinute)
	return &tokenBucket{
		tokens:         capacity,
		capacity:       capacity,
		refillRate:     capacity / 60.0, // convert per minute to per second
		lastRefillTime: time.Now(),
	}
}

// allow checks if a request can proceed and consumes a token if allowed
func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()

	// Refill tokens based on elapsed time
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefillTime = now

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// rateLimiter manages rate limiting for multiple keys
type rateLimiter struct {
	buckets map[string]*tokenBucket
	config  RateLimitConfig
	mu      sync.RWMutex
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(config RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*tokenBucket),
		config:  config,
	}

	// Start cleanup goroutine to remove old buckets
	go rl.cleanup()

	return rl
}

// allow checks if a request from the given key is allowed
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		bucket, exists = rl.buckets[key]
		if !exists {
			bucket = newTokenBucket(rl.config.RequestsPerMinute)
			rl.buckets[key] = bucket
		}
		rl.mu.Unlock()
	}

	return bucket.allow()
}

// cleanup removes inactive buckets periodically
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			bucket.mu.Lock()
			// Remove buckets that haven't been used in 10 minutes
			if now.Sub(bucket.lastRefillTime) > 10*time.Minute {
				delete(rl.buckets, key)
			}
			bucket.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) func(http.Handler) http.Handler {
	limiter := newRateLimiter(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as the key for rate limiting
			key := r.RemoteAddr

			// For authenticated requests, use user ID if available
			if userID, ok := GetUserIDFromContext(r.Context()); ok {
				key = userID
			}

			if !limiter.allow(key) {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
				w.Header().Set("Retry-After", "60")
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthRateLimitMiddleware creates rate limiting for authentication endpoints (5 req/min)
func AuthRateLimitMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(RateLimitConfig{
		RequestsPerMinute: 5,
		BurstSize:         5,
	})
}

// PublicRateLimitMiddleware creates rate limiting for public endpoints (100 req/min)
func PublicRateLimitMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(RateLimitConfig{
		RequestsPerMinute: 100,
		BurstSize:         100,
	})
}

// ProtectedRateLimitMiddleware creates rate limiting for protected endpoints (30 req/min)
func ProtectedRateLimitMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(RateLimitConfig{
		RequestsPerMinute: 30,
		BurstSize:         30,
	})
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
