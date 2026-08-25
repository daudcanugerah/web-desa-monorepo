package config

import "time"

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

// GetExpiration returns JWT expiration with default
func (j *JWTConfig) GetExpiration() time.Duration {
	if j.Expiration == 0 {
		return 24 * time.Hour
	}
	return j.Expiration
}
