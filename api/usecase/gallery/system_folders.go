package gallery

// System feature slugs identify the system folder that backs uploads for a
// given feature. Each feature (banner, berita, struktur, umkm, fasilitas,
// user, ppid) has exactly one system folder, seeded on first run and never
// mutable via the API.
const (
	FeatureBanner    = "banner"
	FeatureBerita    = "berita"
	FeatureStruktur  = "struktur"
	FeatureUMKM      = "umkm"
	FeatureFasilitas = "fasilitas"
	FeatureUser      = "user"
	FeaturePPID      = "ppid"
	FeatureDesa      = "desa"
)

// SystemFolderSpec describes a system folder to seed.
type SystemFolderSpec struct {
	Slug        string
	Name        string
	Description string
}

// SystemFolderSpecs returns the canonical system folder set seeded on
// first run. Order is stable so admin listings remain deterministic.
func SystemFolderSpecs() []SystemFolderSpec {
	return []SystemFolderSpec{
		{Slug: FeatureBanner, Name: "system/banner", Description: "Backing folder for banner cover images."},
		{Slug: FeatureBerita, Name: "system/berita", Description: "Backing folder for berita cover images and embedded content."},
		{Slug: FeatureStruktur, Name: "system/struktur", Description: "Backing folder for struktur profile images."},
		{Slug: FeatureUMKM, Name: "system/umkm", Description: "Backing folder for UMKM gallery images."},
		{Slug: FeatureFasilitas, Name: "system/fasilitas", Description: "Backing folder for fasilitas map images."},
		{Slug: FeatureUser, Name: "system/user", Description: "Backing folder for user avatar images."},
		{Slug: FeaturePPID, Name: "system/ppid", Description: "Backing folder for PPID documents and thumbnails."},
		{Slug: FeatureDesa, Name: "system/desa", Description: "Backing folder for village profile images (kepala desa photo)."},
	}
}

// systemFolderSpecFor returns the single canonical spec for a slug, or nil.
func systemFolderSpecFor(slug string) *SystemFolderSpec {
	for _, spec := range SystemFolderSpecs() {
		if spec.Slug == slug {
			return &spec
		}
	}
	return nil
}
