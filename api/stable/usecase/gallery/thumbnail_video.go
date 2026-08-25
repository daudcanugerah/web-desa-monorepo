package gallery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var ErrFFmpegUnavailable = errors.New("ffmpeg/ffprobe unavailable")

type VideoProcessorConfig struct {
	FfmpegPath                 string
	FfprobePath                string
	ThumbnailMaxWidth          int
	ThumbnailMaxHeight         int
	ThumbnailQuality           int
	VideoThumbnailFrameSeconds int
}

type videoProcessor struct {
	cfg          VideoProcessorConfig
	availFFmpeg  bool
	availFFprobe bool
}

func NewVideoProcessor(cfg VideoProcessorConfig) VideoProcessor {
	vp := &videoProcessor{cfg: cfg}
	if vp.cfg.FfmpegPath == "" {
		vp.cfg.FfmpegPath = "ffmpeg"
	}
	if vp.cfg.FfprobePath == "" {
		vp.cfg.FfprobePath = "ffprobe"
	}
	if path, err := exec.LookPath(vp.cfg.FfmpegPath); err == nil && path != "" {
		vp.availFFmpeg = true
		vp.cfg.FfmpegPath = path
	}
	if path, err := exec.LookPath(vp.cfg.FfprobePath); err == nil && path != "" {
		vp.availFFprobe = true
		vp.cfg.FfprobePath = path
	}
	if vp.cfg.ThumbnailMaxWidth <= 0 {
		vp.cfg.ThumbnailMaxWidth = 400
	}
	if vp.cfg.ThumbnailMaxHeight <= 0 {
		vp.cfg.ThumbnailMaxHeight = 400
	}
	if vp.cfg.ThumbnailQuality <= 0 {
		vp.cfg.ThumbnailQuality = 80
	}
	if vp.cfg.VideoThumbnailFrameSeconds <= 0 {
		vp.cfg.VideoThumbnailFrameSeconds = 1
	}
	return vp
}

func (p *videoProcessor) Available() bool {
	return p.availFFmpeg
}

func (p *videoProcessor) GenerateThumbnail(ctx context.Context, srcPath, dstPath string) (int, int, float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, 0, err
	}
	if !p.availFFmpeg {
		return 0, 0, 0, ErrFFmpegUnavailable
	}
	if srcPath == "" || dstPath == "" {
		return 0, 0, 0, errors.New("video thumbnail: src and dst paths are required")
	}

	duration := 0.0
	if p.availFFprobe {
		if d, err := probeDuration(ctx, p.cfg.FfprobePath, srcPath); err == nil {
			duration = d
		}
	}

	seekAt := float64(p.cfg.VideoThumbnailFrameSeconds)
	if duration > 0 && seekAt > duration/2 {
		seekAt = duration / 2
	}
	if seekAt < 0 {
		seekAt = 0
	}

	width, height := 0, 0
	if p.availFFprobe {
		if w, h, err := probeDimensions(ctx, p.cfg.FfprobePath, srcPath); err == nil {
			width, height = w, h
		}
	}

	scaleW := p.cfg.ThumbnailMaxWidth
	scaleH := p.cfg.ThumbnailMaxHeight
	if width > 0 && width < scaleW {
		scaleW = width
	}
	if height > 0 && height < scaleH {
		scaleH = height
	}

	scaleArg := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", scaleW, scaleH)

	args := []string{
		"-y",
		"-ss", strconv.FormatFloat(seekAt, 'f', 2, 64),
		"-i", srcPath,
		"-frames:v", "1",
		"-vf", scaleArg,
		"-c:v", "libwebp",
		"-lossless", "0",
		"-q:v", strconv.Itoa(p.cfg.ThumbnailQuality),
		dstPath,
	}

	cmd := exec.CommandContext(ctx, p.cfg.FfmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(dstPath)
		return 0, 0, 0, fmt.Errorf("ffmpeg thumbnail: %w (%s)", err, stderr.String())
	}
	return width, height, duration, nil
}

func probeDuration(ctx context.Context, ffprobePath, srcPath string) (float64, error) {
	cmd := exec.CommandContext(ctx, ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		srcPath,
	)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe duration: %w (%s)", err, errb.String())
	}
	raw := strings.TrimSpace(out.String())
	if raw == "" || raw == "N/A" {
		return 0, nil
	}
	d, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration parse: %w", err)
	}
	return d, nil
}

func probeDimensions(ctx context.Context, ffprobePath, srcPath string) (int, int, error) {
	cmd := exec.CommandContext(ctx, ffprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		srcPath,
	)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("ffprobe dimensions: %w (%s)", err, errb.String())
	}
	raw := strings.TrimSpace(out.String())
	if raw == "" {
		return 0, 0, fmt.Errorf("ffprobe dimensions: empty")
	}
	parts := strings.SplitN(raw, "x", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("ffprobe dimensions: unexpected %q", raw)
	}
	w, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe width parse: %w", err)
	}
	h, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe height parse: %w", err)
	}
	return w, h, nil
}
