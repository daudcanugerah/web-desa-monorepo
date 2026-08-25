package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// MessageResponse represents a generic API message response
type MessageResponse struct {
	Message string `json:"message"`
}

// CategoryInfo represents a category with id and name for embedded responses.
type CategoryInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// HealthResponse represents a health-check response payload
type HealthResponse struct {
	Status string `json:"status"`
}

// FileResponse represents metadata for a served file
type FileResponse struct {
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Success writes a successful JSON response
func Success(w http.ResponseWriter, status int, data interface{}) {
	JSON(w, status, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// Error writes an error JSON response and logs the error
func Error(w http.ResponseWriter, status int, message string) {
	// Log error with status code
	if status >= 500 {
		log.Printf("[ERROR] HTTP %d: %s", status, message)
	} else if status >= 400 {
		log.Printf("[WARN] HTTP %d: %s", status, message)
	}

	JSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
	})
}

// ErrorWithDetails writes an error JSON response with additional error details logged
func ErrorWithDetails(w http.ResponseWriter, status int, message string, err error) {
	// Log error with details
	if status >= 500 {
		log.Printf("[ERROR] HTTP %d: %s - Details: %v", status, message, err)
	} else if status >= 400 {
		log.Printf("[WARN] HTTP %d: %s - Details: %v", status, message, err)
	}

	JSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
	})
}
