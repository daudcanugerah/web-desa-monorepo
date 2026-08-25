package gallery

import (
	"context"
	"errors"

	domaingallery "webdesa/api/domain/gallery"
)

var ErrDuplicateFolderName = errors.New("duplicate folder name")

type Repository interface {
	CreateFolder(ctx context.Context, f *domaingallery.Folder) error
	UpdateFolder(ctx context.Context, f *domaingallery.Folder) error
	UpdateFolderVisibility(ctx context.Context, id string, isPublic bool) error
	DeleteFolder(ctx context.Context, id string) error
	GetFolderByID(ctx context.Context, id string) (*domaingallery.Folder, error)
	GetPublicFolderByID(ctx context.Context, id string) (*domaingallery.Folder, error)
	GetFolderByName(ctx context.Context, name string) (*domaingallery.Folder, error)
	ListFolders(ctx context.Context, q FolderListInput) ([]domaingallery.Folder, int, error)
	ListFoldersWithPublicMedia(ctx context.Context, q FolderListInput) ([]domaingallery.Folder, int, error)

	CreateMedia(ctx context.Context, m *domaingallery.Media) error
	UpdateMedia(ctx context.Context, m *domaingallery.Media) error
	UpdateMediaVisibility(ctx context.Context, id string, isPublic bool) error
	BulkUpdateMediaVisibility(ctx context.Context, folderID string, mediaIDs []string, isPublic bool) error
	DeleteMedia(ctx context.Context, id string) error
	GetMediaByID(ctx context.Context, id string) (*domaingallery.Media, error)
	GetPublicMediaByID(ctx context.Context, id string) (*domaingallery.Media, error)
	ListMediaByFolder(ctx context.Context, folderID string, publicOnly bool, q MediaListInput) ([]domaingallery.Media, int, error)
	ListAllMedia(ctx context.Context, q MediaListInput) ([]domaingallery.Media, int, error)

	RecomputeFolderCover(ctx context.Context, folderID string) error
}
