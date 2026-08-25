package config

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	AuthRequestsPerMinute      int `mapstructure:"auth_requests_per_minute"`
	PublicRequestsPerMinute    int `mapstructure:"public_requests_per_minute"`
	ProtectedRequestsPerMinute int `mapstructure:"protected_requests_per_minute"`
}

// GetAuthRequestsPerMinute returns auth rate limit with default
func (r *RateLimitConfig) GetAuthRequestsPerMinute() int {
	if r.AuthRequestsPerMinute == 0 {
		return 5
	}
	return r.AuthRequestsPerMinute
}

// GetPublicRequestsPerMinute returns public rate limit with default
func (r *RateLimitConfig) GetPublicRequestsPerMinute() int {
	if r.PublicRequestsPerMinute == 0 {
		return 100
	}
	return r.PublicRequestsPerMinute
}

// GetProtectedRequestsPerMinute returns protected rate limit with default
func (r *RateLimitConfig) GetProtectedRequestsPerMinute() int {
	if r.ProtectedRequestsPerMinute == 0 {
		return 30
	}
	return r.ProtectedRequestsPerMinute
}
