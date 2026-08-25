package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	domaingallery "webdesa/api/domain/gallery"
	galleryUsecase "webdesa/api/usecase/gallery"
)

type galleryMediaRow struct {
	ID               string    `db:"id"`
	FolderID         string    `db:"folder_id"`
	MediaType        string    `db:"media_type"`
	FileURL          string    `db:"file_url"`
	ThumbnailURL     *string   `db:"thumbnail_url"`
	ThumbnailFailed  bool      `db:"thumbnail_failed"`
	OriginalFilename string    `db:"original_filename"`
	MimeType         string    `db:"mime_type"`
	FileSize         int64     `db:"file_size"`
	Width            *int      `db:"width"`
	Height           *int      `db:"height"`
	DurationSeconds  *float64  `db:"duration_seconds"`
	IsPublic         bool      `db:"is_public"`
	UploadedBy       string    `db:"uploaded_by"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

const galleryMediaColumns = `
	m.id,
	m.folder_id,
	m.media_type,
	m.file_url,
	m.thumbnail_url,
	m.thumbnail_failed,
	m.original_filename,
	m.mime_type,
	m.file_size,
	m.width,
	m.height,
	m.duration_seconds,
	m.is_public,
	m.uploaded_by,
	m.created_at,
	m.updated_at`

func (r *GalleryRepository) CreateMedia(ctx context.Context, m *domaingallery.Media) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gallery_media (
			id, folder_id, media_type, file_url, thumbnail_url, thumbnail_failed,
			original_filename, mime_type, file_size, width, height, duration_seconds,
			is_public, uploaded_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
	`, m.ID, m.FolderID, string(m.MediaType), m.FileURL, m.ThumbnailURL, m.ThumbnailFailed,
		m.OriginalFilename, m.MimeType, m.FileSize, m.Width, m.Height, m.DurationSeconds,
		m.IsPublic, m.UploadedBy)
	if err != nil {
		if isGalleryFolderForeignKeyViolation(err) {
			return folderNotFound(m.FolderID)
		}
		return fmt.Errorf("gallery media create: %w", err)
	}
	return nil
}

func (r *GalleryRepository) UpdateMedia(ctx context.Context, m *domaingallery.Media) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE gallery_media
		SET media_type = $1,
			file_url = $2,
			thumbnail_url = $3,
			thumbnail_failed = $4,
			original_filename = $5,
			mime_type = $6,
			file_size = $7,
			width = $8,
			height = $9,
			duration_seconds = $10,
			is_public = $11,
			updated_at = NOW()
		WHERE id = $12
	`, string(m.MediaType), m.FileURL, m.ThumbnailURL, m.ThumbnailFailed,
		m.OriginalFilename, m.MimeType, m.FileSize, m.Width, m.Height,
		m.DurationSeconds, m.IsPublic, m.ID)
	if err != nil {
		return fmt.Errorf("gallery media update: %w", err)
	}
	if err := requireRows(result, galleryUsecase.ErrMediaNotFound, m.ID); err != nil {
		return err
	}
	return nil
}

func (r *GalleryRepository) UpdateMediaVisibility(ctx context.Context, id string, isPublic bool) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE gallery_media
		SET is_public = $1, updated_at = NOW()
		WHERE id = $2
	`, isPublic, id)
	if err != nil {
		return fmt.Errorf("gallery media visibility update: %w", err)
	}
	if err := requireRows(result, galleryUsecase.ErrMediaNotFound, id); err != nil {
		return err
	}
	return nil
}

func (r *GalleryRepository) BulkUpdateMediaVisibility(ctx context.Context, folderID string, mediaIDs []string, isPublic bool) error {
	ids := uniqueStrings(mediaIDs)
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gallery media bulk visibility transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var folderExists string
	if err := tx.GetContext(ctx, &folderExists, `
		SELECT id
		FROM gallery_folders
		WHERE id = $1
		FOR SHARE
	`, folderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return folderNotFound(folderID)
		}
		return fmt.Errorf("gallery media bulk folder check: %w", err)
	}

	var matched []string
	if err := tx.SelectContext(ctx, &matched, `
		SELECT id
		FROM gallery_media
		WHERE folder_id = $1 AND id = ANY($2::uuid[])
		FOR UPDATE
	`, folderID, pq.Array(ids)); err != nil {
		return fmt.Errorf("gallery media bulk check: %w", err)
	}
	if len(matched) != len(ids) {
		return mediaNotFound("")
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE gallery_media
		SET is_public = $1, updated_at = NOW()
		WHERE folder_id = $2 AND id = ANY($3::uuid[])
	`, isPublic, folderID, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("gallery media bulk visibility update: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gallery media bulk rows affected: %w", err)
	}
	if int(rows) != len(ids) {
		return mediaNotFound("")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gallery media bulk visibility commit: %w", err)
	}
	return nil
}

func (r *GalleryRepository) DeleteMedia(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM gallery_media WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("gallery media delete: %w", err)
	}
	if err := requireRows(result, galleryUsecase.ErrMediaNotFound, id); err != nil {
		return err
	}
	return nil
}

func (r *GalleryRepository) GetMediaByID(ctx context.Context, id string) (*domaingallery.Media, error) {
	var row galleryMediaRow
	if err := r.db.GetContext(ctx, &row, `
		SELECT `+galleryMediaColumns+`
		FROM gallery_media m
		WHERE m.id = $1
	`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, mediaNotFound(id)
		}
		return nil, fmt.Errorf("gallery media find: %w", err)
	}
	return row.media(), nil
}

func (r *GalleryRepository) GetPublicMediaByID(ctx context.Context, id string) (*domaingallery.Media, error) {
	var row galleryMediaRow
	if err := r.db.GetContext(ctx, &row, `
		SELECT `+galleryMediaColumns+`
		FROM gallery_media m
		JOIN gallery_folders f ON f.id = m.folder_id
		WHERE m.id = $1 AND f.is_public = TRUE AND m.is_public = TRUE
	`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, mediaNotFound(id)
		}
		return nil, fmt.Errorf("gallery public media find: %w", err)
	}
	return row.media(), nil
}

func (r *GalleryRepository) ListMediaByFolder(ctx context.Context, folderID string, publicOnly bool, q galleryUsecase.MediaListInput) ([]domaingallery.Media, int, error) {
	clauses := []string{"WHERE m.folder_id = $1"}
	args := []interface{}{folderID}
	arg := 2
	if publicOnly {
		clauses = append(clauses, "f.is_public = TRUE", "m.is_public = TRUE")
	}
	appendMediaFilters(&clauses, &args, &arg, q, false)
	return r.listMedia(ctx, strings.Join(clauses, " AND "), args, q.Limit, pageOffset(q.Page, q.Limit))
}

func (r *GalleryRepository) ListAllMedia(ctx context.Context, q galleryUsecase.MediaListInput) ([]domaingallery.Media, int, error) {
	clauses := []string{"WHERE 1=1"}
	args := make([]interface{}, 0, 5)
	arg := 1
	appendMediaFilters(&clauses, &args, &arg, q, true)
	return r.listMedia(ctx, strings.Join(clauses, " AND "), args, q.Limit, pageOffset(q.Page, q.Limit))
}

func (r *GalleryRepository) listMedia(ctx context.Context, where string, args []interface{}, limit, offset int) ([]domaingallery.Media, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM gallery_media m
		JOIN gallery_folders f ON f.id = m.folder_id
		`+where, args...); err != nil {
		return nil, 0, fmt.Errorf("gallery media count: %w", err)
	}

	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, limit, offset)
	query := fmt.Sprintf(`
		SELECT %s
		FROM gallery_media m
		JOIN gallery_folders f ON f.id = m.folder_id
		%s
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $%d OFFSET $%d
	`, galleryMediaColumns, where, len(args)+1, len(args)+2)

	var rows []galleryMediaRow
	if err := r.db.SelectContext(ctx, &rows, query, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("gallery media list: %w", err)
	}
	media := make([]domaingallery.Media, 0, len(rows))
	for _, row := range rows {
		media = append(media, *row.media())
	}
	return media, total, nil
}

func (r galleryMediaRow) media() *domaingallery.Media {
	return &domaingallery.Media{
		ID:               r.ID,
		FolderID:         r.FolderID,
		MediaType:        domaingallery.MediaType(r.MediaType),
		FileURL:          r.FileURL,
		ThumbnailURL:     r.ThumbnailURL,
		ThumbnailFailed:  r.ThumbnailFailed,
		OriginalFilename: r.OriginalFilename,
		MimeType:         r.MimeType,
		FileSize:         r.FileSize,
		Width:            r.Width,
		Height:           r.Height,
		DurationSeconds:  r.DurationSeconds,
		IsPublic:         r.IsPublic,
		UploadedBy:       r.UploadedBy,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func appendMediaFilters(clauses *[]string, args *[]interface{}, arg *int, q galleryUsecase.MediaListInput, includeFolderID bool) {
	if strings.TrimSpace(q.Query) != "" {
		*clauses = append(*clauses, fmt.Sprintf("(m.original_filename ILIKE $%d OR m.mime_type ILIKE $%d OR m.id::text ILIKE $%d)", *arg, *arg, *arg))
		*args = append(*args, "%"+strings.TrimSpace(q.Query)+"%")
		(*arg)++
	}
	if q.IsPublic != nil {
		*clauses = append(*clauses, fmt.Sprintf("m.is_public = $%d", *arg))
		*args = append(*args, *q.IsPublic)
		(*arg)++
	}
	if includeFolderID && strings.TrimSpace(q.FolderID) != "" {
		*clauses = append(*clauses, fmt.Sprintf("m.folder_id::text = $%d", *arg))
		*args = append(*args, strings.TrimSpace(q.FolderID))
		(*arg)++
	}
	if strings.TrimSpace(q.MediaType) != "" {
		*clauses = append(*clauses, fmt.Sprintf("m.media_type::text = $%d", *arg))
		*args = append(*args, strings.TrimSpace(q.MediaType))
		(*arg)++
	}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func mediaNotFound(id string) error {
	return fmt.Errorf("%w: %w: %s", galleryUsecase.ErrMediaNotFound, sql.ErrNoRows, id)
}

func isGalleryFolderForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503" && strings.Contains(pqErr.Constraint, "gallery_media_folder_id")
}
