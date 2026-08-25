package config

// MetabaseConfig holds Metabase configuration
type MetabaseConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	URL       string `mapstructure:"url"`
}
