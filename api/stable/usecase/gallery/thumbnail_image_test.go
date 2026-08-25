package gallery

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCtx(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}

func makeTestImage(t *testing.T, format string, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}))
	case "png":
		require.NoError(t, png.Encode(&buf, img))
	}
	return buf.Bytes()
}

func TestImageProcessor_PreflightRejectsOversized(t *testing.T) {
	p := NewImageProcessor()
	large := makeTestImage(t, "jpeg", 8000, 7000)
	_, _, _, err := p.GenerateThumbnail(testCtx(t), bytes.NewReader(large), 400, 400, 80)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds")
}

func TestImageProcessor_Downscale(t *testing.T) {
	p := NewImageProcessor()
	src := makeTestImage(t, "jpeg", 1000, 500)
	webp, w, h, err := p.GenerateThumbnail(testCtx(t), bytes.NewReader(src), 400, 400, 80)
	require.NoError(t, err)
	require.NotEmpty(t, webp)
	assert.True(t, w <= 400)
	assert.True(t, h <= 400)
}

func TestImageProcessor_NoUpscale(t *testing.T) {
	p := NewImageProcessor()
	src := makeTestImage(t, "jpeg", 100, 80)
	_, w, h, err := p.GenerateThumbnail(testCtx(t), bytes.NewReader(src), 400, 400, 80)
	require.NoError(t, err)
	assert.Equal(t, 100, w)
	assert.Equal(t, 80, h)
}

func TestImageProcessor_PNG(t *testing.T) {
	p := NewImageProcessor()
	src := makeTestImage(t, "png", 500, 300)
	webp, w, h, err := p.GenerateThumbnail(testCtx(t), bytes.NewReader(src), 400, 400, 80)
	require.NoError(t, err)
	require.NotEmpty(t, webp)
	assert.True(t, w <= 400)
	assert.True(t, h <= 400)
}

func TestImageProcessor_BadQuality(t *testing.T) {
	p := NewImageProcessor()
	src := makeTestImage(t, "jpeg", 100, 100)
	_, _, _, err := p.GenerateThumbnail(testCtx(t), bytes.NewReader(src), 100, 100, 0)
	require.Error(t, err)
}

func TestImageProcessor_NilReader(t *testing.T) {
	p := NewImageProcessor()
	_, _, _, err := p.GenerateThumbnail(testCtx(t), nil, 100, 100, 80)
	require.Error(t, err)
}

func TestFitWithin(t *testing.T) {
	w, h := fitWithin(100, 100, 400, 400)
	assert.Equal(t, 100, w)
	assert.Equal(t, 100, h)

	w, h = fitWithin(800, 400, 400, 400)
	assert.Equal(t, 400, w)
	assert.Equal(t, 200, h)

	w, h = fitWithin(400, 800, 400, 400)
	assert.Equal(t, 200, w)
	assert.Equal(t, 400, h)

	w, h = fitWithin(0, 0, 400, 400)
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}
