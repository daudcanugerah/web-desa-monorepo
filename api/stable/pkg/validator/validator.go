package validator

import (
	"fmt"
	"mime/multipart"
	"net/mail"
	"strings"
	"unicode"
)

const (
	// Image constraints
	MaxImageSize = 10 * 1024 * 1024 // 10MB

	// Document constraints
	MaxDocumentSize = 50 * 1024 * 1024 // 50MB

	// Password constraints
	MinPasswordLength = 8
)

var (
	// AllowedImageTypes defines valid MIME types for image uploads
	AllowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}

	// AllowedDocumentTypes defines valid MIME types for document uploads
	AllowedDocumentTypes = map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	}
)

// ValidateEmail checks if the provided email address is valid
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidatePasswordStrength checks if the password meets strength requirements
// Requirements:
// - At least 8 characters long
// - Contains at least one uppercase letter
// - Contains at least one lowercase letter
// - Contains at least one digit
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	}

	var (
		hasUpper bool
		hasLower bool
		hasDigit bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

// ValidateImageFile validates an uploaded image file
// Checks file size and MIME type according to requirements 19.1 and 19.3
func ValidateImageFile(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return fmt.Errorf("file is required")
	}

	// Check file size (max 10MB)
	if fileHeader.Size > MaxImageSize {
		return fmt.Errorf("image file size exceeds maximum allowed size of 10MB")
	}

	// Check MIME type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("content type is required")
	}

	// Normalize content type (remove charset if present)
	contentType = strings.Split(contentType, ";")[0]
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	if !AllowedImageTypes[contentType] {
		return fmt.Errorf("invalid image type: only JPEG, PNG, and WebP are allowed")
	}

	return nil
}

// ValidateDocumentFile validates an uploaded document file
// Checks file size and MIME type according to requirements 19.2 and 19.4
func ValidateDocumentFile(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return fmt.Errorf("file is required")
	}

	// Check file size (max 50MB)
	if fileHeader.Size > MaxDocumentSize {
		return fmt.Errorf("document file size exceeds maximum allowed size of 50MB")
	}

	// Check MIME type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("content type is required")
	}

	// Normalize content type (remove charset if present)
	contentType = strings.Split(contentType, ";")[0]
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	if !AllowedDocumentTypes[contentType] {
		return fmt.Errorf("invalid document type: only PDF, DOC, DOCX, XLS, and XLSX are allowed")
	}

	return nil
}

// ValidateFileSize checks if a file size is within the specified limit
func ValidateFileSize(size int64, maxSize int64) error {
	if size > maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d bytes", size, maxSize)
	}
	return nil
}

// ValidateFileType checks if a MIME type is in the allowed list
func ValidateFileType(contentType string, allowedTypes map[string]bool) error {
	if contentType == "" {
		return fmt.Errorf("content type is required")
	}

	// Normalize content type
	contentType = strings.Split(contentType, ";")[0]
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	if !allowedTypes[contentType] {
		return fmt.Errorf("invalid file type: %s", contentType)
	}

	return nil
}
