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
		return nil, 0, 0, fmt.Errorf("decode image: %w", err)
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
