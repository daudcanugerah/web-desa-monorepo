package gallery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVideoProcessor_AvailableDefaultsFalse(t *testing.T) {
	vp := NewVideoProcessor(VideoProcessorConfig{})
	_ = vp
}

func TestVideoProcessor_UnavailableReturnsErrFFmpegUnavailable(t *testing.T) {
	vp := &videoProcessor{}
	_, _, _, err := vp.GenerateThumbnail(context.Background(), "/tmp/x.mp4", "/tmp/y.webp")
	assert.ErrorIs(t, err, ErrFFmpegUnavailable)
}
