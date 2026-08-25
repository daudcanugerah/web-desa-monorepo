package config

// BackupConfig holds backup configuration
type BackupConfig struct {
	Directory string `mapstructure:"directory"`
	// IncludePaths is a list of additional on-disk roots to bundle into
	// the backup archive alongside the pg_dump output (Task 5.2). Each
	// path is preserved relative to the current working directory when
	// the backup is created, so the restore command can unpack files
	// back to the same layout. Defaults to ["./uploads/"] when empty.
	IncludePaths []string `mapstructure:"include_paths"`
}

// GetDirectory returns backup directory with default
func (b *BackupConfig) GetDirectory() string {
	if b.Directory == "" {
		return "./backups"
	}
	return b.Directory
}

// GetIncludePaths returns the configured include paths, falling back to
// the default ["./uploads/"] when the operator hasn't set anything.
func (b *BackupConfig) GetIncludePaths() []string {
	if len(b.IncludePaths) == 0 {
		return []string{"./uploads/"}
	}
	return b.IncludePaths
}
