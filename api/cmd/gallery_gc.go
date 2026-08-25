package cmd

import (
	"context"
	"fmt"
	"time"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	galleryusecase "webdesa/api/usecase/gallery"
	repopg "webdesa/api/interface/postgres"
	"webdesa/api/pkg/clock"
	"webdesa/api/interface/file"
)

var galleryGCCmd = &cobra.Command{
	Use:   "gallery-gc",
	Short: "Remove orphaned gallery media (system folders, unreferenced by any feature)",
	Long:  "Finds system-folder media that no feature references (no *_media_id column and not present in any images_media_ids JSONB) and older than the cutoff, then deletes the media rows and their files. Use --dry-run to preview.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGalleryGC(cmd, args)
	},
}

var (
	galleryGCDryRun       bool
	galleryGCOlderThan    int
	galleryGCLimit        int
)

func init() {
	galleryGCCmd.Flags().BoolVar(&galleryGCDryRun, "dry-run", false, "list candidates without deleting")
	galleryGCCmd.Flags().IntVar(&galleryGCOlderThan, "older-than-days", 7, "only consider media older than N days")
	galleryGCCmd.Flags().IntVar(&galleryGCLimit, "limit", 50, "maximum number of media to process per run")
	rootCmd.AddCommand(galleryGCCmd)
}

func runGalleryGC(cmd *cobra.Command, _ []string) error {
	if systemConfig == nil {
		return errtrace.Wrap(fmt.Errorf("system config not initialized"))
	}
	ctx := context.Background()

	db, err := sqlx.Connect("postgres", systemConfig.Postgres.DSN)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()

	galleryRepo := repopg.NewGalleryRepository(db)
	galleryOriginals, err := file.NewLocalHandlerWithSubdir(systemConfig.FileUpload.GetPrivateUploadDirectory(), "gallery/originals")
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create gallery originals handler: %w", err))
	}
	galleryThumbs, err := file.NewLocalHandlerWithSubdir(systemConfig.FileUpload.GetPrivateUploadDirectory(), "gallery/thumbnails")
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create gallery thumbs handler: %w", err))
	}
	galleryService := galleryusecase.NewService(
		galleryRepo,
		galleryOriginals,
		galleryThumbs,
		galleryusecase.NewImageProcessor(),
		galleryusecase.NewVideoProcessor(galleryusecase.VideoProcessorConfig{
			FfmpegPath:                 systemConfig.Gallery.GetFfmpegPath(),
			FfprobePath:                systemConfig.Gallery.GetFfprobePath(),
			ThumbnailMaxWidth:          systemConfig.Gallery.GetThumbnailMaxWidth(),
			ThumbnailMaxHeight:         systemConfig.Gallery.GetThumbnailMaxHeight(),
			ThumbnailQuality:           systemConfig.Gallery.GetThumbnailQuality(),
			VideoThumbnailFrameSeconds: systemConfig.Gallery.GetVideoThumbnailFrameSeconds(),
		}),
		systemConfig.Gallery,
		clock.RealClock{},
		galleryusecase.NewMediaDeletionHub(),
	// Task 7.1: signed URLs disabled in this CLI/integration path
	nil,
	)

	if galleryGCOlderThan < 0 {
		return errtrace.Wrap(fmt.Errorf("--older-than-days must be >= 0"))
	}
	cutoff := time.Now().Add(-time.Duration(galleryGCOlderThan) * 24 * time.Hour)

	orphans, err := galleryService.FindOrphanMedia(ctx, cutoff, galleryGCLimit)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to find orphan media: %w", err))
	}

	mainOtel.Log.Infof(ctx, "Found %d orphan media candidate(s) older than %d day(s)", len(orphans), galleryGCOlderThan)
	for _, m := range orphans {
		mainOtel.Log.Infof(ctx, "  - %s (%s, created %s)", m.ID, m.OriginalFilename, m.CreatedAt.Format(time.RFC3339))
	}

	if galleryGCDryRun {
		mainOtel.Log.Info(ctx, "Dry run — nothing deleted")
		return nil
	}

	deleted := 0
	for _, m := range orphans {
		if err := galleryService.DeleteMediaForFeature(ctx, m.ID); err != nil {
			mainOtel.Log.Infof(ctx, "  !! failed to delete %s: %v", m.ID, err)
			continue
		}
		deleted++
	}
	mainOtel.Log.Infof(ctx, "Deleted %d media row(s) and their files", deleted)
	return nil
}
