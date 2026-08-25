package file

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts the public file-serving endpoint.
func RegisterRoutes(r chi.Router, h *FileHandler) {
	r.Get("/files/{filename}", h.GetFile)
}