package gallery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/HugoSmits86/nativewebp"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const MaxImagePixels = 50_000_000

type imageProcessor struct{}

func NewImageProcessor() ImageProcessor {
	return &imageProcessor{}
}

func (p *imageProcessor) GenerateThumbnail(ctx context.Context, src io.Reader, maxW, maxH, quality int) ([]byte, int, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, 0, err
	}
	if src == nil {
		return nil, 0, 0, errors.New("image source is nil")
	}
	if maxW <= 0 || maxH <= 0 {
		return nil, 0, 0, fmt.Errorf("invalid thumbnail bounds: %dx%d", maxW, maxH)
	}
	if quality <= 0 || quality > 100 {
		return nil, 0, 0, fmt.Errorf("invalid webp quality: %d", quality)
	}

	buf, err := io.ReadAll(src)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("read image: %w", err)
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("decode image config: %w", err)
	}

	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, 0, 0, fmt.Errorf("invalid image dimensions: %dx%d", cfg.Width, cfg.Height)
	}

	if int64(cfg.Width)*int64(cfg.Height) > MaxImagePixels {
		return nil, 0, 0, fmt.Errorf("image too large: %dx%d exceeds %d pixel preflight", cfg.Width, cfg.Height, MaxImagePixels)
	}

	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		// Animated WebP (and other multi-frame containers) are rejected by
		// the standard decoder. Extract the first embedded image frame and
		// decode that instead so the thumbnail reflects the animation's
		// first frame rather than failing outright.
		img, err = decodeFirstWebPFrame(buf)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("decode image: %w", err)
		}
	}

	dstW, dstH := fitWithin(cfg.Width, cfg.Height, maxW, maxH)

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	if dstW == cfg.Width && dstH == cfg.Height {
		draw.Draw(dst, dst.Bounds(), img, image.Point{}, draw.Src)
	} else {
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	}

	var out bytes.Buffer
	opts := &nativewebp.Options{
		UseExtendedFormat: false,
		CompressionLevel:  compressionLevelFor(quality),
	}
	if err := nativewebp.Encode(&out, dst, opts); err != nil {
		return nil, 0, 0, fmt.Errorf("encode webp: %w", err)
	}

	return out.Bytes(), dstW, dstH, nil
}

// decodeFirstWebPFrame extracts and decodes the first embedded image
// frame from a WebP container (handles animated WebP, which the
// standard library decoder rejects). Returns the frame as an image.Image.
func decodeFirstWebPFrame(buf []byte) (image.Image, error) {
	if len(buf) < 12 || string(buf[0:4]) != "RIFF" || string(buf[8:12]) != "WEBP" {
		return nil, errors.New("not a webp container")
	}
	pos := 12
	for pos+8 <= len(buf) {
		chunkID := string(buf[pos : pos+4])
		size := int(buf[pos+4]) | int(buf[pos+5])<<8 | int(buf[pos+6])<<16 | int(buf[pos+7])<<24
		if pos+8+size > len(buf) {
			return nil, errors.New("webp container truncated")
		}
		body := buf[pos+8 : pos+8+size]
		switch chunkID {
		case "ANMF":
			// Animated WebP frame: 16-byte frame header followed by a
			// nested VP8 / VP8L image chunk. Decode that chunk directly.
			if len(body) > 16 {
				if img, err := decodeWebPImageChunk(body[16:]); err == nil {
					return img, nil
				}
			}
		case "VP8 ", "VP8L":
			if img, err := decodeWebPImageChunk(body); err == nil {
				return img, nil
			}
		}
		pos += 8 + size
		if size%2 == 1 {
			pos++
		}
	}
	return nil, errors.New("no decodable webp frame found")
}

// decodeWebPImageChunk wraps a bare VP8 / VP8L chunk in a minimal RIFF
// WEBP container and decodes it with the standard library decoder.
func decodeWebPImageChunk(chunk []byte) (image.Image, error) {
	if len(chunk) < 8 {
		return nil, errors.New("truncated webp image chunk")
	}
	size := int(chunk[4]) | int(chunk[5])<<8 | int(chunk[6])<<16 | int(chunk[7])<<24
	if len(chunk) < 8+size {
		return nil, errors.New("truncated webp image chunk")
	}
	frame := append([]byte("RIFF\x00\x00\x00\x00WEBP"), chunk[0:8+size]...)
	total := len(frame) - 8
	copy(frame[4:8], []byte{byte(total), byte(total >> 8), byte(total >> 16), byte(total >> 24)})
	img, _, err := image.Decode(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func compressionLevelFor(quality int) nativewebp.CompressionLevel {
	switch {
	case quality < 30:
		return nativewebp.BestSpeed
	case quality >= 80:
		return nativewebp.BestCompression
	default:
		return nativewebp.DefaultCompression
	}
}

func fitWithin(srcW, srcH, maxW, maxH int) (int, int) {
	if srcW <= maxW && srcH <= maxH {
		return srcW, srcH
	}
	sw := float64(srcW)
	sh := float64(srcH)
	ratio := float64(maxW) / sw
	if sh*ratio > float64(maxH) {
		ratio = float64(maxH) / sh
	}
	return int(sw * ratio), int(sh * ratio)
}
