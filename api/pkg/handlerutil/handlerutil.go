// Package handlerutil contains shared HTTP handler utilities used across
// feature packages. It was extracted from the old `handler` package to
// avoid cross-package import cycles when feature handlers are split into
// separate Go packages (one per subdirectory).
package handlerutil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"webdesa/api/interface/http/middleware"
	"webdesa/api/pkg/response"
)

// validate is a singleton validator instance used by ValidateStruct.
var validate = validator.New()

// ValidateStruct validates a struct using the validator package and returns
// a user-friendly error message if validation fails.
func ValidateStruct(data interface{}) error {
	err := validate.Struct(data)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return err
	}

	firstErr := validationErrors[0]
	fieldName := firstErr.Field()
	tag := firstErr.Tag()

	switch tag {
	case "required":
		return fmt.Errorf("%s is required", fieldName)
	case "email":
		return fmt.Errorf("%s must be a valid email", fieldName)
	case "min":
		return fmt.Errorf("%s is too short", fieldName)
	case "max":
		return fmt.Errorf("%s is too long", fieldName)
	case "oneof":
		return fmt.Errorf("%s has an invalid value", fieldName)
	default:
		return fmt.Errorf("%s: %s", fieldName, tag)
	}
}

// InferContentType returns a MIME type based on the file extension.
// Used by FileHandler and similar handlers.
func InferContentType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".jpg"), strings.HasSuffix(filename, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(filename, ".png"):
		return "image/png"
	case strings.HasSuffix(filename, ".webp"):
		return "image/webp"
	case strings.HasSuffix(filename, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(filename, ".doc"):
		return "application/msword"
	case strings.HasSuffix(filename, ".docx"):
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case strings.HasSuffix(filename, ".xls"):
		return "application/vnd.ms-excel"
	case strings.HasSuffix(filename, ".xlsx"):
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

// IsValidUUID checks if a string is a valid UUID format.
func IsValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

// IsDuplicateKeyError reports whether err originates from a PostgreSQL
// unique-constraint violation (SQLSTATE 23505). The pq driver surfaces the
// code in the error message, so we match on the canonical substring.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "23505")
}

// RequireUserIDFromCtx extracts the authenticated user ID from the request
// context. On miss it writes a 401 to w and returns ok=false so the handler
// can early-return. Use inside handlers wrapped by middleware.Auth().
func RequireUserIDFromCtx(w http.ResponseWriter, r *http.Request) (string, bool) {
	uid, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || uid == "" {
		response.Error(w, http.StatusUnauthorized, "authentication required")
		return "", false
	}
	return uid, true
}

// ParseUUIDParam extracts and validates a chi URL parameter as a UUID. On
// miss or invalid format it writes a 400 to w and returns ok=false. The
// returned id is safe to forward to repos without further validation.
func ParseUUIDParam(w http.ResponseWriter, r *http.Request, paramName string) (string, bool) {
	id := chi.URLParam(r, paramName)
	if id == "" {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s is required", paramName))
		return "", false
	}
	if !IsValidUUID(id) {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s must be a valid UUID", paramName))
		return "", false
	}
	return id, true
}