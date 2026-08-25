// Package bannercategory defines the BannerCategory domain entity.
//
// BannerCategory is a managed vocabulary for grouping banners
// (e.g., "Promo", "Pengumuman", "Event"). Same pattern as berita_categories,
// umkm_categories, ppid_categories, fasilitas_categories, and
// infographic_categories: a separate table with FK ON DELETE RESTRICT and
// a "Lainnya" placeholder for legacy rows.
package bannercategory

import (
	"fmt"
	"strings"
	"time"
)

// Category is a banner category entity (managed vocabulary row).
type Category struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Validate checks domain invariants for the Category entity.
func (c *Category) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Name) > 100 {
		return fmt.Errorf("name too long (max 100 chars)")
	}
	if c.CreatedAt.IsZero() || c.UpdatedAt.IsZero() {
		return fmt.Errorf("timestamps must be set")
	}
	return nil
}