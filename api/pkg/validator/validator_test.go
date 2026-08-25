package validator

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "invalid email - no @",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "invalid email - no domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "invalid email - no local part",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "invalid email - spaces",
			email:   "user @example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid strong password",
			password: "Password123",
			wantErr:  false,
		},
		{
			name:     "valid password with special chars",
			password: "Pass@word123!",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1",
			wantErr:  true,
		},
		{
			name:     "no uppercase",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "no lowercase",
			password: "PASSWORD123",
			wantErr:  true,
		},
		{
			name:     "no digit",
			password: "PasswordOnly",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "exactly 8 characters valid",
			password: "Pass1234",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateImageFile(t *testing.T) {
	tests := []struct {
		name       string
		fileHeader *multipart.FileHeader
		wantErr    bool
	}{
		{
			name:       "valid JPEG image",
			fileHeader: createFileHeader("test.jpg", "image/jpeg", 1024*1024), // 1MB
			wantErr:    false,
		},
		{
			name:       "valid PNG image",
			fileHeader: createFileHeader("test.png", "image/png", 5*1024*1024), // 5MB
			wantErr:    false,
		},
		{
			name:       "valid WebP image",
			fileHeader: createFileHeader("test.webp", "image/webp", 2*1024*1024), // 2MB
			wantErr:    false,
		},
		{
			name:       "image too large",
			fileHeader: createFileHeader("large.jpg", "image/jpeg", 11*1024*1024), // 11MB
			wantErr:    true,
		},
		{
			name:       "invalid type - PDF",
			fileHeader: createFileHeader("doc.pdf", "application/pdf", 1024*1024),
			wantErr:    true,
		},
		{
			name:       "invalid type - text",
			fileHeader: createFileHeader("file.txt", "text/plain", 1024),
			wantErr:    true,
		},
		{
			name:       "nil file header",
			fileHeader: nil,
			wantErr:    true,
		},
		{
			name:       "empty content type",
			fileHeader: createFileHeader("test.jpg", "", 1024*1024),
			wantErr:    true,
		},
		{
			name:       "content type with charset",
			fileHeader: createFileHeader("test.jpg", "image/jpeg; charset=utf-8", 1024*1024),
			wantErr:    false,
		},
		{
			name:       "exactly 10MB",
			fileHeader: createFileHeader("test.jpg", "image/jpeg", 10*1024*1024),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageFile(tt.fileHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDocumentFile(t *testing.T) {
	tests := []struct {
		name       string
		fileHeader *multipart.FileHeader
		wantErr    bool
	}{
		{
			name:       "valid PDF",
			fileHeader: createFileHeader("doc.pdf", "application/pdf", 5*1024*1024), // 5MB
			wantErr:    false,
		},
		{
			name:       "valid DOC",
			fileHeader: createFileHeader("doc.doc", "application/msword", 10*1024*1024), // 10MB
			wantErr:    false,
		},
		{
			name:       "valid DOCX",
			fileHeader: createFileHeader("doc.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", 15*1024*1024), // 15MB
			wantErr:    false,
		},
		{
			name:       "valid XLS",
			fileHeader: createFileHeader("sheet.xls", "application/vnd.ms-excel", 20*1024*1024), // 20MB
			wantErr:    false,
		},
		{
			name:       "valid XLSX",
			fileHeader: createFileHeader("sheet.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", 25*1024*1024), // 25MB
			wantErr:    false,
		},
		{
			name:       "document too large",
			fileHeader: createFileHeader("large.pdf", "application/pdf", 51*1024*1024), // 51MB
			wantErr:    true,
		},
		{
			name:       "invalid type - image",
			fileHeader: createFileHeader("image.jpg", "image/jpeg", 1024*1024),
			wantErr:    true,
		},
		{
			name:       "invalid type - text",
			fileHeader: createFileHeader("file.txt", "text/plain", 1024),
			wantErr:    true,
		},
		{
			name:       "nil file header",
			fileHeader: nil,
			wantErr:    true,
		},
		{
			name:       "empty content type",
			fileHeader: createFileHeader("doc.pdf", "", 1024*1024),
			wantErr:    true,
		},
		{
			name:       "exactly 50MB",
			fileHeader: createFileHeader("doc.pdf", "application/pdf", 50*1024*1024),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDocumentFile(tt.fileHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDocumentFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFileSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		maxSize int64
		wantErr bool
	}{
		{
			name:    "size within limit",
			size:    1024,
			maxSize: 2048,
			wantErr: false,
		},
		{
			name:    "size equals limit",
			size:    2048,
			maxSize: 2048,
			wantErr: false,
		},
		{
			name:    "size exceeds limit",
			size:    3000,
			maxSize: 2048,
			wantErr: true,
		},
		{
			name:    "zero size",
			size:    0,
			maxSize: 1024,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileSize(tt.size, tt.maxSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFileType(t *testing.T) {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	tests := []struct {
		name         string
		contentType  string
		allowedTypes map[string]bool
		wantErr      bool
	}{
		{
			name:         "valid type - jpeg",
			contentType:  "image/jpeg",
			allowedTypes: allowedTypes,
			wantErr:      false,
		},
		{
			name:         "valid type - png",
			contentType:  "image/png",
			allowedTypes: allowedTypes,
			wantErr:      false,
		},
		{
			name:         "invalid type",
			contentType:  "image/gif",
			allowedTypes: allowedTypes,
			wantErr:      true,
		},
		{
			name:         "empty content type",
			contentType:  "",
			allowedTypes: allowedTypes,
			wantErr:      true,
		},
		{
			name:         "content type with charset",
			contentType:  "image/jpeg; charset=utf-8",
			allowedTypes: allowedTypes,
			wantErr:      false,
		},
		{
			name:         "content type with spaces",
			contentType:  "  image/png  ",
			allowedTypes: allowedTypes,
			wantErr:      false,
		},
		{
			name:         "uppercase content type",
			contentType:  "IMAGE/JPEG",
			allowedTypes: allowedTypes,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileType(tt.contentType, tt.allowedTypes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to create a multipart.FileHeader for testing
func createFileHeader(filename, contentType string, size int64) *multipart.FileHeader {
	// Create a buffer with dummy data
	buf := bytes.NewBuffer(make([]byte, size))

	// Create multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}

	part, _ := writer.CreatePart(h)
	part.Write(buf.Bytes())
	writer.Close()

	// Parse the multipart form
	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(size + 1024)

	if len(form.File["file"]) > 0 {
		return form.File["file"][0]
	}

	// Fallback: create a simple FileHeader
	return &multipart.FileHeader{
		Filename: filename,
		Header:   h,
		Size:     size,
	}
}
