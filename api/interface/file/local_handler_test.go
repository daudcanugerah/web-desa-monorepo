package file

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"webdesa/api/usecase/gallery"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ gallery.Storage = (*LocalHandler)(nil)

func TestLocalHandlerSatisfiesGalleryStorage(t *testing.T) {
	dir := t.TempDir()
	h, err := NewLocalHandlerWithSubdir(dir, "gallery/originals")
	require.NoError(t, err)

	t.Run("Save writes content and returns relative name", func(t *testing.T) {
		body := []byte("hello gallery")
		rel, err := h.Save(context.Background(), "abc.jpg", bytes.NewReader(body), int64(len(body)), "image/jpeg")
		require.NoError(t, err)
		assert.Equal(t, "abc.jpg", rel)

		got, err := os.ReadFile(h.Path(rel))
		require.NoError(t, err)
		assert.Equal(t, body, got)
	})

	t.Run("Save rejects non-positive size", func(t *testing.T) {
		_, err := h.Save(context.Background(), "x.jpg", bytes.NewReader([]byte("x")), 0, "image/jpeg")
		assert.Error(t, err)
	})

	t.Run("Save rejects traversal filename", func(t *testing.T) {
		_, err := h.Save(context.Background(), "../escape.jpg", bytes.NewReader([]byte("x")), 1, "image/jpeg")
		assert.Error(t, err)
	})

	t.Run("Save rejects empty filename", func(t *testing.T) {
		_, err := h.Save(context.Background(), "", bytes.NewReader([]byte("x")), 1, "image/jpeg")
		assert.Error(t, err)
	})

	t.Run("Save rejects dot filenames", func(t *testing.T) {
		_, err := h.Save(context.Background(), ".", bytes.NewReader([]byte("x")), 1, "image/jpeg")
		assert.Error(t, err)
		_, err = h.Save(context.Background(), "..", bytes.NewReader([]byte("x")), 1, "image/jpeg")
		assert.Error(t, err)
	})

	t.Run("Save rejects content exceeding declared size and cleans up", func(t *testing.T) {
		name := "oversize.bin"
		full := h.Path(name)
		_, err := h.Save(context.Background(), name, bytes.NewReader([]byte("12345")), 3, "application/octet-stream")
		assert.Error(t, err)
		_, statErr := os.Stat(full)
		assert.True(t, os.IsNotExist(statErr), "partial file must be cleaned up")
	})

	t.Run("Save rejects content shorter than declared size and cleans up", func(t *testing.T) {
		name := "short.bin"
		full := h.Path(name)
		_, err := h.Save(context.Background(), name, bytes.NewReader([]byte("hi")), 10, "application/octet-stream")
		assert.Error(t, err)
		_, statErr := os.Stat(full)
		assert.True(t, os.IsNotExist(statErr), "partial file must be cleaned up")
	})

	t.Run("Path returns safe absolute path", func(t *testing.T) {
		got := h.Path("../escape.jpg")
		assert.False(t, strings.Contains(got, ".."))
		assert.Equal(t, filepath.Join(h.uploadDir, "escape.jpg"), got)
	})

	t.Run("Path keeps nested upload dir joined to basename", func(t *testing.T) {
		got := h.Path("nested/keep.jpg")
		assert.Equal(t, filepath.Join(h.uploadDir, "keep.jpg"), got)
	})

	t.Run("Delete removes saved file", func(t *testing.T) {
		body := []byte("delete-me")
		rel, err := h.Save(context.Background(), "del.bin", bytes.NewReader(body), int64(len(body)), "application/octet-stream")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(h.Path(rel), body, 0644))
		require.NoError(t, h.Delete(context.Background(), rel))
		_, err = os.Stat(h.Path(rel))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("GetFilePath still works for legacy callers", func(t *testing.T) {
		assert.Equal(t, filepath.Join(h.uploadDir, "legacy.jpg"), h.GetFilePath("legacy.jpg"))
	})

	t.Run("Save signature uses io.Reader", func(t *testing.T) {
		var r io.Reader = bytes.NewReader([]byte("ok"))
		rel, err := h.Save(context.Background(), "sig.bin", r, 2, "application/octet-stream")
		require.NoError(t, err)
		assert.Equal(t, "sig.bin", rel)
	})
}
