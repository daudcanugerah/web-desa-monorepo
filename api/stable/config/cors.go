package config

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// GetAllowedOrigins returns allowed origins with default
func (c *CORSConfig) GetAllowedOrigins() []string {
	if len(c.AllowedOrigins) == 0 {
		return []string{"http://localhost:3000"}
	}
	return c.AllowedOrigins
}
