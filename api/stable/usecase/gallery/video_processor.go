package gallery

import "context"

type VideoProcessor interface {
	GenerateThumbnail(ctx context.Context, srcPath, dstPath string) (width, height int, duration float64, err error)
	Available() bool
}
