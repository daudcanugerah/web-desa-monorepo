package config

// GalleryConfig holds gallery-specific configuration for thumbnail
// generation, upload size limits, bulk upload caps, and ffmpeg/ffprobe
// binary paths.
type GalleryConfig struct {
	ThumbnailMaxWidth          int    `mapstructure:"thumbnail_max_width"`
	ThumbnailMaxHeight         int    `mapstructure:"thumbnail_max_height"`
	ThumbnailQuality           int    `mapstructure:"thumbnail_quality"`
	ImageMaxSizeMB             int    `mapstructure:"image_max_size_mb"`
	VideoMaxSizeMB             int    `mapstructure:"video_max_size_mb"`
	BulkUploadMaxFiles         int    `mapstructure:"bulk_upload_max_files"`
	BulkUploadMaxTotalMB       int    `mapstructure:"bulk_upload_max_total_mb"`
	FfmpegPath                 string `mapstructure:"ffmpeg_path"`
	FfprobePath                string `mapstructure:"ffprobe_path"`
	VideoThumbnailFrameSeconds int    `mapstructure:"video_thumbnail_frame_seconds"`
}

// GetThumbnailMaxWidth returns the max thumbnail width with default (400)
func (g *GalleryConfig) GetThumbnailMaxWidth() int {
	if g.ThumbnailMaxWidth <= 0 {
		return 400
	}
	return g.ThumbnailMaxWidth
}

// GetThumbnailMaxHeight returns the max thumbnail height with default (400)
func (g *GalleryConfig) GetThumbnailMaxHeight() int {
	if g.ThumbnailMaxHeight <= 0 {
		return 400
	}
	return g.ThumbnailMaxHeight
}

// GetThumbnailQuality returns the WebP thumbnail quality with default (80)
// and clamps to the [1, 100] range.
func (g *GalleryConfig) GetThumbnailQuality() int {
	if g.ThumbnailQuality <= 0 {
		return 80
	}
	if g.ThumbnailQuality > 100 {
		return 100
	}
	return g.ThumbnailQuality
}

// GetImageMaxSizeMB returns the max image upload size in MB with default (10)
func (g *GalleryConfig) GetImageMaxSizeMB() int {
	if g.ImageMaxSizeMB <= 0 {
		return 10
	}
	return g.ImageMaxSizeMB
}

// GetVideoMaxSizeMB returns the max video upload size in MB with default (100)
func (g *GalleryConfig) GetVideoMaxSizeMB() int {
	if g.VideoMaxSizeMB <= 0 {
		return 100
	}
	return g.VideoMaxSizeMB
}

// GetBulkUploadMaxFiles returns the max number of files per bulk upload with default (20)
func (g *GalleryConfig) GetBulkUploadMaxFiles() int {
	if g.BulkUploadMaxFiles <= 0 {
		return 20
	}
	return g.BulkUploadMaxFiles
}

// GetBulkUploadMaxTotalMB returns the max total bulk upload size in MB with default (250)
func (g *GalleryConfig) GetBulkUploadMaxTotalMB() int {
	if g.BulkUploadMaxTotalMB <= 0 {
		return 250
	}
	return g.BulkUploadMaxTotalMB
}

// GetFfmpegPath returns the ffmpeg binary path with default ("ffmpeg")
func (g *GalleryConfig) GetFfmpegPath() string {
	if g.FfmpegPath == "" {
		return "ffmpeg"
	}
	return g.FfmpegPath
}

// GetFfprobePath returns the ffprobe binary path with default ("ffprobe")
func (g *GalleryConfig) GetFfprobePath() string {
	if g.FfprobePath == "" {
		return "ffprobe"
	}
	return g.FfprobePath
}

// GetVideoThumbnailFrameSeconds returns the frame offset used when extracting
// a video thumbnail, with default (1s).
func (g *GalleryConfig) GetVideoThumbnailFrameSeconds() int {
	if g.VideoThumbnailFrameSeconds <= 0 {
		return 1
	}
	return g.VideoThumbnailFrameSeconds
}
