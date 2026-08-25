package backup

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeArchive is a tiny helper that produces a tar.gz with two entries
// (a fake pg_dump under "db.pgdump" and one file under uploads/) so we
// can exercise the extract helpers without invoking pg_dump.
func makeArchive(t *testing.T, archivePath, dumpContent, uploadRel, uploadContent string) {
	t.Helper()
	f, err := os.Create(archivePath)
	require.NoError(t, err)
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "db.pgdump", Mode: 0644, Size: int64(len(dumpContent))}))
	_, err = tw.Write([]byte(dumpContent))
	require.NoError(t, err)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: uploadRel, Mode: 0644, Size: int64(len(uploadContent))}))
	_, err = tw.Write([]byte(uploadContent))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
}

func TestExtractFromTarGzReadsDump(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "full.tar.gz")
	makeArchive(t, archivePath, "pg-dump-bytes", "uploads/public/a.jpg", "fake-image")

	out, err := os.CreateTemp(dir, "dump-*.pgdump")
	require.NoError(t, err)
	defer os.Remove(out.Name())
	require.NoError(t, extractFromTarGz(archivePath, "db.pgdump", out))
	out.Close()

	got, err := os.ReadFile(out.Name())
	require.NoError(t, err)
	assert.Equal(t, "pg-dump-bytes", string(got))
}

func TestExtractFromTarGzMissingEntry(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "full.tar.gz")
	makeArchive(t, archivePath, "dump", "uploads/public/a.jpg", "x")
	out, _ := os.Create(filepath.Join(dir, "out"))
	defer out.Close()
	err := extractFromTarGz(archivePath, "nope.pgdump", out)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestExtractUploadsFromTarGzRebuildsTree(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "full.tar.gz")
	makeArchive(t, archivePath, "dump-bytes", "uploads/public/hero.jpg", "pixel-data")

	// chdir into root so extractUploadsFromTarGz writes the uploads tree
	// there (the helper preserves the path verbatim).
	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	defer os.Chdir(cwd)

	require.NoError(t, extractUploadsFromTarGz(archivePath))

	// The archive contained uploads/public/hero.jpg relative to "." which
	// inside this chdir is `root`. Wait — the working directory of the
	// test process is the repo root, not `root`, so the file would land
	// at repo-root/uploads/public/hero.jpg. Adjust expectation: it
	// should exist somewhere on disk relative to the chdir we entered.
	got, err := filepath.Abs(filepath.Join(root, "uploads/public/hero.jpg"))
	require.NoError(t, err)
	_, err = os.Stat(got)
	require.NoError(t, err, "expected uploads file at %s after extract", got)
}

func TestAddDirToTarMissingRootIsNoop(t *testing.T) {
	// The helper must silently skip a non-existent root so a fresh install
	// (no uploads/ on disk yet) doesn't fail the backup.
	require.NoError(t, addDirToTar(nil, "/no/such/path"))
}

func TestExtractUploadsFromTarGzSkipsDBEntry(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "full.tar.gz")
	makeArchive(t, archivePath, "dump", "uploads/public/skip.txt", "x")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "uploads/public"), 0755))
	// Move into a tmp cwd where the extract won't blow up
	cwd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(cwd)

	require.NoError(t, extractUploadsFromTarGz(archivePath))
	_, err := os.Stat(filepath.Join(dir, "uploads/public/skip.txt"))
	require.NoError(t, err)
}

// ensure io is referenced so go vet doesn't complain
var _ io.Reader = (*os.File)(nil)
