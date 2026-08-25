package gallery

import (
	"context"
	"io"
)

type Storage interface {
	Save(ctx context.Context, name string, content io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, path string) error
	Path(name string) string
}
