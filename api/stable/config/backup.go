package config

// BackupConfig holds backup configuration
type BackupConfig struct {
	Directory string `mapstructure:"directory"`
}

// GetDirectory returns backup directory with default
func (b *BackupConfig) GetDirectory() string {
	if b.Directory == "" {
		return "./backups"
	}
	return b.Directory
}
