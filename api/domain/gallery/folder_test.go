package gallery

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFolderValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		folder  Folder
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid folder",
			folder: Folder{
				ID:        "f0000000-0000-0000-0000-000000000001",
				Name:      "Panen Raya 2026",
				CreatedBy: "u0000000-0000-0000-0000-000000000001",
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "missing id",
			folder: Folder{
				Name:      "Panen",
				CreatedBy: "u0000000-0000-0000-0000-000000000001",
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "blank name",
			folder: Folder{
				ID:        "f0000000-0000-0000-0000-000000000001",
				Name:      "   ",
				CreatedBy: "u0000000-0000-0000-0000-000000000001",
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "name too long",
			folder: Folder{
				ID:        "f0000000-0000-0000-0000-000000000001",
				Name:      strings.Repeat("a", folderNameMaxLen+1),
				CreatedBy: "u0000000-0000-0000-0000-000000000001",
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "description too long",
			folder: Folder{
				ID:          "f0000000-0000-0000-0000-000000000001",
				Name:        "Panen",
				Description: strings.Repeat("a", folderDescriptionMaxLen+1),
				CreatedBy:   "u0000000-0000-0000-0000-000000000001",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "missing created_by",
			folder: Folder{
				ID:        "f0000000-0000-0000-0000-000000000001",
				Name:      "Panen",
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "missing timestamps",
			folder: Folder{
				ID:        "f0000000-0000-0000-0000-000000000001",
				Name:      "Panen",
				CreatedBy: "u0000000-0000-0000-0000-000000000001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.folder.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
