package integration

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBeritaUpload_UploadMediaImage tests uploading an image to temporary directory
func TestBeritaUpload_UploadMediaImage(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create a test image file
	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG header
	imageData = append(imageData, make([]byte, 1000)...)

	// Create multipart request
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file with explicit Content-Type
	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {"form-data; name=\"file\"; filename=\"test-image.jpg\""},
		"Content-Type":        {"image/jpeg"},
	})
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(imageData))
	require.NoError(t, err)

	// Add media type
	err = writer.WriteField("type", "image")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// Make request
	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, respBody["success"].(bool))

	data := getDataField(respBody)
	require.NotNil(t, data)
	assert.NotEmpty(t, data["url"])
	assert.Contains(t, data["url"].(string), "tmp-")
}

// TestBeritaUpload_UploadMediaVideo tests uploading a video to temporary directory
func TestBeritaUpload_UploadMediaVideo(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create a test video file
	videoData := make([]byte, 1000)

	// Create multipart request
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file with explicit Content-Type
	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {"form-data; name=\"file\"; filename=\"test-video.mp4\""},
		"Content-Type":        {"video/mp4"},
	})
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(videoData))
	require.NoError(t, err)

	// Add media type
	err = writer.WriteField("type", "video")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// Make request
	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, respBody["success"].(bool))

	data := getDataField(respBody)
	require.NotNil(t, data)
	assert.NotEmpty(t, data["url"])
	assert.Contains(t, data["url"].(string), "tmp-")
}

// TestBeritaUpload_UploadMediaWithoutAuth tests that upload requires authentication
func TestBeritaUpload_UploadMediaWithoutAuth(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	imageData = append(imageData, make([]byte, 1000)...)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test-image.jpg")
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(imageData))
	require.NoError(t, err)

	err = writer.WriteField("type", "image")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestBeritaUpload_UploadMediaInvalidMimeType tests rejection of invalid MIME types
func TestBeritaUpload_UploadMediaInvalidMimeType(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create a test file with invalid MIME type
	fileData := []byte("invalid file content")

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test-file.txt")
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(fileData))
	require.NoError(t, err)

	err = writer.WriteField("type", "image")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	respBody, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, respBody["success"].(bool))
}

// TestBeritaUpload_UploadMediaInvalidMediaType tests rejection of invalid media type
func TestBeritaUpload_UploadMediaInvalidMediaType(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	imageData = append(imageData, make([]byte, 1000)...)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test-image.jpg")
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(imageData))
	require.NoError(t, err)

	err = writer.WriteField("type", "invalid")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	respBody, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, respBody["success"].(bool))
}

// TestBeritaUpload_UploadMediaMissingFile tests rejection when file is missing
func TestBeritaUpload_UploadMediaMissingFile(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	err := writer.WriteField("type", "image")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	respBody, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, respBody["success"].(bool))
}

// TestBeritaUpload_UploadMediaDefaultType tests that media type defaults to "image"
func TestBeritaUpload_UploadMediaDefaultType(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	imageData = append(imageData, make([]byte, 1000)...)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file with explicit Content-Type
	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {"form-data; name=\"file\"; filename=\"test-image.jpg\""},
		"Content-Type":        {"image/jpeg"},
	})
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(imageData))
	require.NoError(t, err)

	// Don't set type field - should default to "image"
	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, respBody["success"].(bool))
}

// TestBeritaUpload_UploadMediaFileSizeExceeded tests rejection of oversized files
func TestBeritaUpload_UploadMediaFileSizeExceeded(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Create a file larger than 10MB for image
	largeData := make([]byte, 11*1024*1024) // 11MB

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "large-image.jpg")
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(largeData))
	require.NoError(t, err)

	err = writer.WriteField("type", "image")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	respBody, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.False(t, respBody["success"].(bool))
}

// TestBeritaUpload_TmpDirectoryCreation tests that tmp directory is created if it doesn't exist
// TestBeritaUpload_TmpDirectoryCreation tests that uploads directory is created if it doesn't exist
func TestBeritaUpload_TmpDirectoryCreation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	imageData = append(imageData, make([]byte, 1000)...)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file with explicit Content-Type
	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {"form-data; name=\"file\"; filename=\"test-image.jpg\""},
		"Content-Type":        {"image/jpeg"},
	})
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(imageData))
	require.NoError(t, err)

	err = writer.WriteField("type", "image")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify uploads directory was created
	assert.DirExists(t, "./test_uploads")
}

// TestBeritaUpload_GetUploadedFilePreview tests retrieving uploaded file for preview via file access endpoint
// TestBeritaUpload_GetUploadedFilePreview tests retrieving uploaded file for preview via file access endpoint
// TestBeritaUpload_GetUploadedFilePreview tests retrieving uploaded file for preview via file access endpoint
func TestBeritaUpload_GetUploadedFilePreview(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	ts.CreateTestUser(t, "Admin User", "admin@test.com", "password123")
	token := ts.GetAuthToken(t, "admin@test.com", "password123")
	ts.AssignAdminRole(t, "admin@test.com")

	// Upload an image
	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	imageData = append(imageData, make([]byte, 1000)...)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {"form-data; name=\"file\"; filename=\"test-image.jpg\""},
		"Content-Type":        {"image/jpeg"},
	})
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(imageData))
	require.NoError(t, err)

	err = writer.WriteField("type", "image")
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.Server.URL+"/api/v1/berita/upload-media", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := parseJSONResponse(resp.Body)
	require.NoError(t, err)
	assert.True(t, respBody["success"].(bool))

	data := getDataField(respBody)
	require.NotNil(t, data)
	uploadURL := data["url"].(string)
	assert.Contains(t, uploadURL, "tmp-")

	// URL is now just the filename (e.g., "tmp-20260317-uuid.jpg")
	filename := uploadURL

	// Now retrieve the file using the file access endpoint
	fileReq, err := http.NewRequest("GET", ts.Server.URL+"/files/"+filename, nil)
	require.NoError(t, err)

	fileResp, err := http.DefaultClient.Do(fileReq)
	require.NoError(t, err)
	defer fileResp.Body.Close()

	assert.Equal(t, http.StatusOK, fileResp.StatusCode)
	assert.Equal(t, "image/jpeg", fileResp.Header.Get("Content-Type"))

	// Verify file content
	fileContent, err := io.ReadAll(fileResp.Body)
	require.NoError(t, err)
	assert.Equal(t, len(imageData), len(fileContent))
}

// TestBeritaUpload_GetUploadedFileWithInvalidPath tests that directory traversal is prevented
func TestBeritaUpload_GetUploadedFileWithInvalidPath(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	// Try to access file with directory traversal (contains / which is not allowed in filename)
	req, err := http.NewRequest("GET", ts.Server.URL+"/files/../../../etc/passwd", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Route doesn't match because filename contains /, so returns 404
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
