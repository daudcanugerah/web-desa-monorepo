package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// TestGetFile tests the file access endpoint
func TestGetFile(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Create test file in uploads directory
	uploadsDir := "./test_uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatalf("failed to create uploads directory: %v", err)
	}

	testFilename := "test-file-12345.txt"
	testContent := "This is a test file content"
	testFilePath := filepath.Join(uploadsDir, testFilename)

	// Write test file
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer os.Remove(testFilePath)

	tests := []struct {
		name           string
		filename       string
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "valid file access",
			filename:       testFilename,
			expectedStatus: http.StatusOK,
			expectedBody:   testContent,
			description:    "should return file content with 200 status",
		},
		{
			name:           "file not found",
			filename:       "nonexistent-file.txt",
			expectedStatus: http.StatusNotFound,
			description:    "should return 404 for non-existent file",
		},
		{
			name:           "invalid filename with directory traversal",
			filename:       "../../../etc/passwd",
			expectedStatus: http.StatusNotFound,
			description:    "should reject directory traversal attempts (404 from router)",
		},
		{
			name:           "invalid filename with forward slash",
			filename:       "subdir/file.txt",
			expectedStatus: http.StatusNotFound,
			description:    "should reject filenames with path separators (404 from router)",
		},
		{
			name:           "invalid filename with backslash",
			filename:       "subdir\\file.txt",
			expectedStatus: http.StatusBadRequest,
			description:    "should reject filenames with backslashes",
		},
		{
			name:           "empty filename",
			filename:       "",
			expectedStatus: http.StatusNotFound,
			description:    "should reject empty filename (404 from router)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/files/%s", ts.Server.URL, tt.filename)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tt.expectedStatus, resp.StatusCode, tt.description)
			}

			if tt.expectedStatus == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}

				if string(body) != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
				}

				// Verify Content-Type header
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("expected Content-Type text/plain, got %s", contentType)
				}
			}
		})
	}
}

// TestGetFileWithDifferentContentTypes tests Content-Type detection
func TestGetFileWithDifferentContentTypes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	uploadsDir := "./test_uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatalf("failed to create uploads directory: %v", err)
	}

	tests := []struct {
		filename    string
		content     string
		contentType string
		description string
	}{
		{
			filename:    "test-image-12345.jpg",
			content:     "fake jpg content",
			contentType: "image/jpeg",
			description: "should return image/jpeg for .jpg files",
		},
		{
			filename:    "test-image-12345.png",
			content:     "fake png content",
			contentType: "image/png",
			description: "should return image/png for .png files",
		},
		{
			filename:    "test-doc-12345.pdf",
			content:     "fake pdf content",
			contentType: "application/pdf",
			description: "should return application/pdf for .pdf files",
		},
		{
			filename:    "test-unknown-12345.xyz",
			content:     "unknown content",
			contentType: "application/octet-stream",
			description: "should return application/octet-stream for unknown types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			testFilePath := filepath.Join(uploadsDir, tt.filename)

			// Write test file
			if err := os.WriteFile(testFilePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}
			defer os.Remove(testFilePath)

			url := fmt.Sprintf("%s/files/%s", ts.Server.URL, tt.filename)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType != tt.contentType {
				t.Errorf("expected Content-Type %s, got %s: %s", tt.contentType, contentType, tt.description)
			}
		})
	}
}

// TestGetFileWithUUIDFilename tests UUID-format filenames
func TestGetFileWithUUIDFilename(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	uploadsDir := "./test_uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatalf("failed to create uploads directory: %v", err)
	}

	// UUID-like filename
	uuidFilename := "550e8400-e29b-41d4-a716-446655440000.jpg"
	testContent := "fake image content"
	testFilePath := filepath.Join(uploadsDir, uuidFilename)

	// Write test file
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer os.Remove(testFilePath)

	url := fmt.Sprintf("%s/files/%s", ts.Server.URL, uuidFilename)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != testContent {
		t.Errorf("expected body %q, got %q", testContent, string(body))
	}
}
