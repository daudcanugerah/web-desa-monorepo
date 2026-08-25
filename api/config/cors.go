package config

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// GetAllowedOrigins returns allowed origins with default
func (c *CORSConfig) GetAllowedOrigins() []string {
	if len(c.AllowedOrigins) == 0 {
		// Dev-friendly defaults: the Vite dev servers for webdesa/admin +
		// webdesa/web both default to 5173. `webdesa/web` and `webdesa/admin`
		// may also run on 3000 if a developer prefers that. Add both so a fresh
		// `go run . serve` works without editing config.
		// For production, set [cors].allowed_origins in config.toml.
		return []string{"http://localhost:3000", "http://localhost:5173"}
	}
	return c.AllowedOrigins
}
