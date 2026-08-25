package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDevConfig(t *testing.T) {
	viper.Reset()
	viper.SetConfigFile("dev.toml")
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	require.NoError(t, err, "should read dev.toml")

	var cfg Config
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err, "should unmarshal config")

	// Verify app config
	assert.Equal(t, "desa-api", cfg.App.Name)
	assert.Equal(t, ":8080", cfg.App.HTTPAddr)

	// Verify CORS config
	assert.NotEmpty(t, cfg.CORS.AllowedOrigins)
	assert.Contains(t, cfg.CORS.AllowedOrigins, "http://localhost:3000")

	// Verify rate limit config
	assert.Equal(t, 5, cfg.RateLimit.GetAuthRequestsPerMinute())
	assert.Equal(t, 100, cfg.RateLimit.GetPublicRequestsPerMinute())
	assert.Equal(t, 30, cfg.RateLimit.GetProtectedRequestsPerMinute())

	// Verify file upload config
	assert.Equal(t, 10, cfg.FileUpload.GetMaxImageSizeMB())
	assert.Equal(t, 50, cfg.FileUpload.GetMaxDocumentSizeMB())
	assert.Equal(t, "./uploads/public", cfg.FileUpload.GetPublicUploadDirectory())
	assert.Equal(t, "./uploads/private", cfg.FileUpload.GetPrivateUploadDirectory())
	assert.Equal(t, 20, cfg.Gallery.GetBulkUploadMaxFiles())
}

func TestLoadStagingConfig(t *testing.T) {
	viper.Reset()
	viper.SetConfigFile("staging.toml")
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	require.NoError(t, err, "should read staging.toml")

	var cfg Config
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err, "should unmarshal config")

	// Verify staging-specific settings
	assert.Contains(t, cfg.CORS.AllowedOrigins, "https://staging.desa.example.com")
	assert.Equal(t, "/var/backups/desa-api", cfg.Backup.GetDirectory())
	assert.Equal(t, "/var/uploads/desa-api/public", cfg.FileUpload.GetPublicUploadDirectory())
	assert.Equal(t, "/var/uploads/desa-api/private", cfg.FileUpload.GetPrivateUploadDirectory())
}

func TestLoadProdConfig(t *testing.T) {
	viper.Reset()
	viper.SetConfigFile("prod.toml")
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	require.NoError(t, err, "should read prod.toml")

	var cfg Config
	err = viper.Unmarshal(&cfg)
	require.NoError(t, err, "should unmarshal config")

	// Verify production-specific settings
	assert.Contains(t, cfg.CORS.AllowedOrigins, "https://desa.example.com")
	assert.Equal(t, "/var/backups/desa-api", cfg.Backup.GetDirectory())
	assert.False(t, cfg.Otel.EnableTrace, "trace should be disabled in production")
}

func TestConfigDefaults(t *testing.T) {
	// Test CORS defaults
	corsConfig := CORSConfig{}
	assert.Equal(t, []string{"http://localhost:3000"}, corsConfig.GetAllowedOrigins())

	// Test rate limit defaults
	rateLimitConfig := RateLimitConfig{}
	assert.Equal(t, 5, rateLimitConfig.GetAuthRequestsPerMinute())
	assert.Equal(t, 100, rateLimitConfig.GetPublicRequestsPerMinute())
	assert.Equal(t, 30, rateLimitConfig.GetProtectedRequestsPerMinute())

	// Test file upload defaults
	fileUploadConfig := FileUploadConfig{}
	assert.Equal(t, 10, fileUploadConfig.GetMaxImageSizeMB())
	assert.Equal(t, 50, fileUploadConfig.GetMaxDocumentSizeMB())
	assert.Equal(t, "./uploads/public/", fileUploadConfig.GetPublicUploadDirectory())
	assert.Equal(t, "./uploads/private/", fileUploadConfig.GetPrivateUploadDirectory())

	galleryConfig := GalleryConfig{}
	assert.Equal(t, 400, galleryConfig.GetThumbnailMaxWidth())
	assert.Equal(t, 400, galleryConfig.GetThumbnailMaxHeight())
	assert.Equal(t, 80, galleryConfig.GetThumbnailQuality())
	assert.Equal(t, 10, galleryConfig.GetImageMaxSizeMB())
	assert.Equal(t, 100, galleryConfig.GetVideoMaxSizeMB())
	assert.Equal(t, 20, galleryConfig.GetBulkUploadMaxFiles())
	assert.Equal(t, 250, galleryConfig.GetBulkUploadMaxTotalMB())
}

func TestFileUploadDirectoryIsolation(t *testing.T) {
	valid := FileUploadConfig{
		UploadPublicDirectory:  "./uploads/public",
		UploadPrivateDirectory: "./uploads/private",
		UploadPPIDDirectory:    "./uploads/ppid",
	}
	require.NoError(t, valid.ValidateDirectoryIsolation())

	privateInsidePublic := valid
	privateInsidePublic.UploadPrivateDirectory = "./uploads/public/private"
	require.Error(t, privateInsidePublic.ValidateDirectoryIsolation())

	ppidInsidePublic := valid
	ppidInsidePublic.UploadPPIDDirectory = "./uploads/public/ppid"
	require.Error(t, ppidInsidePublic.ValidateDirectoryIsolation())
}
