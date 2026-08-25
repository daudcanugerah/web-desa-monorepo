package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FileUploadConfig holds file upload configuration
type FileUploadConfig struct {
	MaxImageSizeMB         int    `mapstructure:"max_image_size_mb"`
	MaxDocumentSizeMB      int    `mapstructure:"max_document_size_mb"`
	UploadPublicDirectory  string `mapstructure:"upload_public_directory"`
	UploadPPIDDirectory    string `mapstructure:"upload_ppid_directory"`
	UploadPrivateDirectory string `mapstructure:"upload_private_directory"`
}

// GetMaxImageSizeMB returns max image size with default (10MB)
func (f *FileUploadConfig) GetMaxImageSizeMB() int {
	if f.MaxImageSizeMB == 0 {
		return 10
	}
	return f.MaxImageSizeMB
}

// GetMaxDocumentSizeMB returns max document size with default (50MB)
func (f *FileUploadConfig) GetMaxDocumentSizeMB() int {
	if f.MaxDocumentSizeMB == 0 {
		return 50
	}
	return f.MaxDocumentSizeMB
}

// GetUploadDirectory returns upload directory with default
func (f *FileUploadConfig) GetPublicUploadDirectory() string {
	if f.UploadPublicDirectory == "" {
		return "./uploads/public/"
	}
	return f.UploadPublicDirectory
}

// GetPPIDUploadDirectory returns the PPID upload directory (separate from
// public). Defaults to ./uploads/ppid/ if not configured.
func (f *FileUploadConfig) GetPPIDUploadDirectory() string {
	if f.UploadPPIDDirectory == "" {
		return filepath.Join(f.GetPublicUploadDirectory(), "..", "ppid")
	}
	return f.UploadPPIDDirectory
}

// GetPrivateUploadDirectory returns the private upload directory.
func (f *FileUploadConfig) GetPrivateUploadDirectory() string {
	if f.UploadPrivateDirectory == "" {
		return "./uploads/private/"
	}
	return f.UploadPrivateDirectory
}

func (f *FileUploadConfig) ValidateDirectoryIsolation() error {
	publicDir, err := filepath.Abs(f.GetPublicUploadDirectory())
	if err != nil {
		return fmt.Errorf("resolve public upload directory: %w", err)
	}
	for name, directory := range map[string]string{
		"private": f.GetPrivateUploadDirectory(),
		"ppid":    f.GetPPIDUploadDirectory(),
	} {
		resolved, err := filepath.Abs(directory)
		if err != nil {
			return fmt.Errorf("resolve %s upload directory: %w", name, err)
		}
		relative, err := filepath.Rel(publicDir, resolved)
		if err != nil {
			return fmt.Errorf("compare %s upload directory: %w", name, err)
		}
		if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
			return fmt.Errorf("%s upload directory must be outside public upload directory", name)
		}
	}
	return nil
}
