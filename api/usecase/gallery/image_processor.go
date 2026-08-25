package gallery

import (
	"context"
	"io"
)

type ImageProcessor interface {
	GenerateThumbnail(ctx context.Context, src io.Reader, maxW, maxH, quality int) (webpBytes []byte, width, height int, err error)
}
