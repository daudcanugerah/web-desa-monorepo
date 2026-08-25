package integration

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
)

// tinyJPEG returns the bytes of a minimal 1x1 baseline JPEG, used by
// integration tests that need to upload a real image that passes the
// gallery service's image validation (which decodes the bytes via
// image.DecodeConfig / image.Decode and rejects non-image payloads
// with ErrInvalidMimeType).
//
// The bytes are produced at test-run time by jpegEncodeWhite1x1()
// below so we get a payload that is decodable by Go's stdlib image
// package (both DecodeConfig and Decode must succeed).
func tinyJPEG() []byte {
	return jpegEncodeWhite1x1()
}

// jpegEncodeWhite1x1 returns a freshly encoded 1x1 white-pixel JPEG
// that passes both image.DecodeConfig and image.Decode in Go's
// standard library. Used by tinyJPEG so the gallery service's image
// validation accepts the bytes as a real JPEG.
func jpegEncodeWhite1x1() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		panic("jpeg.Encode failed: " + err.Error())
	}
	return buf.Bytes()
}
