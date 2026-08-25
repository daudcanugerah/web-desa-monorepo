package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	domaingallery "webdesa/api/domain/gallery"
	galleryUsecase "webdesa/api/usecase/gallery"
)

type GalleryRepository struct {
	db *sqlx.DB
}

type galleryFolderRow struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	IsPublic     bool      `db:"is_public"`
	IsSystem     bool      `db:"is_system"`
	FeatureSlug  *string   `db:"feature_slug"`
	CoverMediaID *string   `db:"cover_media_id"`
	CoverManual  bool      `db:"cover_manual"`
	MediaCount   int       `db:"media_count"`
	CreatedBy    string    `db:"created_by"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

const galleryFolderColumns = `
	f.id,
	f.name,
	f.description,
	f.is_public,
	f.is_system,
	f.feature_slug,
	f.cover_media_id,
	f.cover_manual,
	f.created_by,
	f.created_at,
	f.updated_at`

func NewGalleryRepository(db *sqlx.DB) galleryUsecase.Repository {
	return &GalleryRepository{db: db}
}

func (r *GalleryRepository) CreateFolder(ctx context.Context, f *domaingallery.Folder) error {
	if f.ID == "" {
		f.ID = uuid.NewString()
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gallery_folders (id, name, description, is_public, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`, f.ID, f.Name, f.Description, f.IsPublic, f.CreatedBy)
	if err != nil {
		return wrapGalleryFolderError("create", err)
	}
	return nil
}

func (r *GalleryRepository) UpdateFolder(ctx context.Context, f *domaingallery.Folder) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE gallery_folders
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
	`, f.Name, f.Description, f.ID)
	if err != nil {
		return wrapGalleryFolderError("update", err)
	}
	if err := requireRows(result, galleryUsecase.ErrFolderNotFound, f.ID); err != nil {
		return err
	}
	return nil
}

func (r *GalleryRepository) UpdateFolderVisibility(ctx context.Context, id string, isPublic bool) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE gallery_folders
		SET is_public = $1, updated_at = NOW()
		WHERE id = $2
	`, isPublic, id)
	if err != nil {
		return fmt.Errorf("gallery folder visibility update: %w", err)
	}
	if err := requireRows(result, galleryUsecase.ErrFolderNotFound, id); err != nil {
		return err
	}
	return nil
}

func (r *GalleryRepository) DeleteFolder(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM gallery_folders WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("gallery folder delete: %w", err)
	}
	if err := requireRows(result, galleryUsecase.ErrFolderNotFound, id); err != nil {
		return err
	}
	return nil
}

func (r *GalleryRepository) GetFolderByID(ctx context.Context, id string) (*domaingallery.Folder, error) {
	var row galleryFolderRow
	err := r.db.GetContext(ctx, &row, fmt.Sprintf(`
		SELECT %s, COUNT(m.id) AS media_count
		FROM gallery_folders f
		LEFT JOIN gallery_media m ON m.folder_id = f.id
		WHERE f.id = $1
		GROUP BY f.id, f.name, f.description, f.is_public, f.cover_media_id, f.cover_manual, f.created_by, f.created_at, f.updated_at
	`, galleryFolderColumns), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, folderNotFound(id)
		}
		return nil, fmt.Errorf("gallery folder find: %w", err)
	}
	return row.folder(), nil
}

func (r *GalleryRepository) GetPublicFolderByID(ctx context.Context, id string) (*domaingallery.Folder, error) {
	var row galleryFolderRow
	err := r.db.GetContext(ctx, &row, fmt.Sprintf(`
		SELECT %s, COUNT(m.id) AS media_count
		FROM gallery_folders f
		JOIN gallery_media m ON m.folder_id = f.id AND m.is_public = TRUE
		WHERE f.id = $1 AND f.is_public = TRUE
		GROUP BY f.id, f.name, f.description, f.is_public, f.cover_media_id, f.cover_manual, f.created_by, f.created_at, f.updated_at
	`, galleryFolderColumns), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, folderNotFound(id)
		}
		return nil, fmt.Errorf("gallery public folder find: %w", err)
	}
	return row.folder(), nil
}

func (r *GalleryRepository) GetFolderByName(ctx context.Context, name string) (*domaingallery.Folder, error) {
	var row galleryFolderRow
	err := r.db.GetContext(ctx, &row, fmt.Sprintf(`
		SELECT %s, COUNT(m.id) AS media_count
		FROM gallery_folders f
		LEFT JOIN gallery_media m ON m.folder_id = f.id
		WHERE LOWER(f.name) = LOWER($1)
		GROUP BY f.id, f.name, f.description, f.is_public, f.is_system, f.feature_slug, f.cover_media_id, f.cover_manual, f.created_by, f.created_at, f.updated_at
	`, galleryFolderColumns), strings.TrimSpace(name))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, folderNotFound(name)
		}
		return nil, fmt.Errorf("gallery folder find by name: %w", err)
	}
	return row.folder(), nil
}

func (r *GalleryRepository) GetFolderByFeatureSlug(ctx context.Context, slug string) (*domaingallery.Folder, error) {
	if slug == "" {
		return nil, folderNotFound(slug)
	}
	var row galleryFolderRow
	err := r.db.GetContext(ctx, &row, fmt.Sprintf(`
		SELECT %s
		FROM gallery_folders f
		WHERE f.feature_slug = $1
	`, galleryFolderColumns), slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, folderNotFound(slug)
		}
		return nil, fmt.Errorf("gallery folder find by feature slug: %w", err)
	}
	return row.folder(), nil
}

func (r *GalleryRepository) CreateSystemFolder(ctx context.Context, f *domaingallery.Folder) error {
	if f.ID == "" {
		f.ID = uuid.NewString()
	}
	slug := ""
	if f.FeatureSlug != nil {
		slug = *f.FeatureSlug
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gallery_folders (
			id, name, description, is_public, is_system, feature_slug, created_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, TRUE, $5, $6, NOW(), NOW())
		ON CONFLICT (feature_slug) WHERE feature_slug IS NOT NULL DO UPDATE
		SET name = EXCLUDED.name,
		    description = EXCLUDED.description,
		    updated_at = NOW()
	`, f.ID, f.Name, f.Description, f.IsPublic, slug, f.CreatedBy)
	if err != nil {
		return wrapGalleryFolderError("create system", err)
	}
	return nil
}

func (r *GalleryRepository) FindSystemFolderOwner(ctx context.Context) (string, error) {
	var id string
	if err := r.db.GetContext(ctx, &id, `
		SELECT id FROM users ORDER BY created_at ASC, id ASC LIMIT 1
	`); err != nil {
		return "", fmt.Errorf("gallery system folder owner: %w", err)
	}
	return id, nil
}

func (r *GalleryRepository) ListFolders(ctx context.Context, q galleryUsecase.FolderListInput) ([]domaingallery.Folder, int, error) {
	where, args := folderListWhere(q, false)
	return r.listFolders(ctx, where, args, false, pageOffset(q.Page, q.Limit), q.Limit)
}

func (r *GalleryRepository) ListFoldersWithPublicMedia(ctx context.Context, q galleryUsecase.FolderListInput) ([]domaingallery.Folder, int, error) {
	where, args := folderListWhere(q, true)
	return r.listFolders(ctx, where, args, true, pageOffset(q.Page, q.Limit), q.Limit)
}

func (r *GalleryRepository) listFolders(ctx context.Context, where string, args []interface{}, publicMediaOnly bool, offset, limit int) ([]domaingallery.Folder, int, error) {
	var total int
	countWhere := where
	if publicMediaOnly {
		countWhere += ` AND EXISTS (
			SELECT 1
			FROM gallery_media public_media
			WHERE public_media.folder_id = f.id AND public_media.is_public = TRUE
		)`
	}
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM gallery_folders f `+countWhere, args...); err != nil {
		return nil, 0, fmt.Errorf("gallery folder count: %w", err)
	}

	join := "LEFT JOIN gallery_media m ON m.folder_id = f.id"
	if publicMediaOnly {
		join = "JOIN gallery_media m ON m.folder_id = f.id AND m.is_public = TRUE"
	}

	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, limit, offset)
	query := fmt.Sprintf(`
		SELECT %s, COUNT(m.id) AS media_count
		FROM gallery_folders f
		%s
		%s
		GROUP BY f.id, f.name, f.description, f.is_public, f.cover_media_id, f.cover_manual, f.created_by, f.created_at, f.updated_at
		ORDER BY f.created_at DESC, f.id DESC
		LIMIT $%d OFFSET $%d
	`, galleryFolderColumns, join, where, len(args)+1, len(args)+2)

	var rows []galleryFolderRow
	if err := r.db.SelectContext(ctx, &rows, query, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("gallery folder list: %w", err)
	}
	folders := make([]domaingallery.Folder, 0, len(rows))
	for _, row := range rows {
		folders = append(folders, *row.folder())
	}
	return folders, total, nil
}

func (r *GalleryRepository) RecomputeFolderCover(ctx context.Context, folderID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gallery folder cover transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var lockedFolderID string
	if err := tx.GetContext(ctx, &lockedFolderID, `
		SELECT id
		FROM gallery_folders
		WHERE id = $1
		FOR UPDATE
	`, folderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return folderNotFound(folderID)
		}
		return fmt.Errorf("gallery folder cover lock: %w", err)
	}

	// A manually chosen cover is preserved until the cover media itself is
	// deleted (SetFolderCover with NULL) or a new manual cover is set.
	var coverManual bool
	if err := tx.GetContext(ctx, &coverManual, `
		SELECT cover_manual FROM gallery_folders WHERE id = $1
	`, folderID); err == nil && coverManual {
		return tx.Commit()
	}

	var candidate sql.NullString
	err = tx.GetContext(ctx, &candidate, `
		SELECT m.id
		FROM gallery_media m
		JOIN gallery_folders f ON f.id = m.folder_id
		WHERE m.folder_id = $1
		  AND m.thumbnail_url IS NOT NULL
		  AND (f.is_public = FALSE OR m.is_public = TRUE)
		ORDER BY CASE WHEN m.media_type = 'image' THEN 0 ELSE 1 END, m.created_at DESC, m.id DESC
		LIMIT 1
	`, folderID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("gallery folder cover candidate: %w", err)
	}

	var coverMediaID interface{}
	if candidate.Valid {
		coverMediaID = candidate.String
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE gallery_folders
		SET cover_media_id = $1, updated_at = NOW()
		WHERE id = $2
	`, coverMediaID, lockedFolderID); err != nil {
		return fmt.Errorf("gallery folder cover update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gallery folder cover commit: %w", err)
	}
	return nil
}

// SetFolderCover pins the folder cover to the given media and marks it
// manual so auto-recompute leaves it alone. Passing a NULL mediaID clears
// the manual pin and lets RecomputeFolderCover resume auto behaviour.
func (r *GalleryRepository) SetFolderCover(ctx context.Context, folderID string, mediaID *string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE gallery_folders
		SET cover_media_id = $1, cover_manual = $2, updated_at = NOW()
		WHERE id = $3
	`, mediaID, mediaID != nil, folderID)
	if err != nil {
		return fmt.Errorf("gallery folder set cover: %w", err)
	}
	if err := requireRows(result, galleryUsecase.ErrFolderNotFound, folderID); err != nil {
		return err
	}
	return nil
}

func (r galleryFolderRow) folder() *domaingallery.Folder {
	return &domaingallery.Folder{
		ID:           r.ID,
		Name:         r.Name,
		Description:  r.Description,
		IsPublic:     r.IsPublic,
		IsSystem:     r.IsSystem,
		FeatureSlug:  r.FeatureSlug,
		CoverMediaID: r.CoverMediaID,
		CoverManual:  r.CoverManual,
		MediaCount:   r.MediaCount,
		CreatedBy:    r.CreatedBy,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func pageOffset(page, limit int) int {
	if page <= 1 || limit <= 0 {
		return 0
	}
	return (page - 1) * limit
}

func folderListWhere(q galleryUsecase.FolderListInput, publicOnly bool) (string, []interface{}) {
	clauses := []string{"WHERE 1=1"}
	args := make([]interface{}, 0, 3)
	arg := 1
	if publicOnly {
		clauses = append(clauses, "f.is_public = TRUE")
	}
	if q.IsPublic != nil {
		clauses = append(clauses, fmt.Sprintf("f.is_public = $%d", arg))
		args = append(args, *q.IsPublic)
		arg++
	}
	if strings.TrimSpace(q.Query) != "" {
		clauses = append(clauses, fmt.Sprintf("(f.name ILIKE $%d OR f.description ILIKE $%d)", arg, arg))
		args = append(args, "%"+strings.TrimSpace(q.Query)+"%")
	}
	return strings.Join(clauses, " AND "), args
}

func wrapGalleryFolderError(operation string, err error) error {
	if isUniqueViolation(err) {
		return fmt.Errorf("%w: %v", galleryUsecase.ErrDuplicateFolderName, err)
	}
	return fmt.Errorf("gallery folder %s: %w", operation, err)
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505" && (pqErr.Constraint == "" || pqErr.Constraint == "uq_gallery_folders_name")
}

func requireRows(result sql.Result, sentinel error, id string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gallery rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: %w: %s", sentinel, sql.ErrNoRows, id)
	}
	return nil
}

func folderNotFound(id string) error {
	return fmt.Errorf("%w: %w: %s", galleryUsecase.ErrFolderNotFound, sql.ErrNoRows, id)
}
