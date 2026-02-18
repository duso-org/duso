package cli

import (
	"io/fs"
	"strings"
)

// vfs provides normalized access to the embedded filesystem.
// All functions automatically convert backslashes to forward slashes
// to ensure compatibility with embed.FS across platforms.

func normalizeEmbeddedPath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	return strings.TrimPrefix(path, "/")
}

// EmbeddedFileRead reads from embedded fs with path normalization.
func EmbeddedFileRead(path string) ([]byte, error) {
	return embeddedFS.ReadFile(normalizeEmbeddedPath(path))
}

// EmbeddedDirRead reads a directory from embedded fs with path normalization.
func EmbeddedDirRead(path string) ([]fs.DirEntry, error) {
	return embeddedFS.ReadDir(normalizeEmbeddedPath(path))
}

// EmbeddedStat gets file info from embedded fs with path normalization.
func EmbeddedStat(path string) (fs.FileInfo, error) {
	return fs.Stat(embeddedFS, normalizeEmbeddedPath(path))
}

// EmbeddedGlob matches files in embedded fs with path normalization.
func EmbeddedGlob(pattern string) ([]string, error) {
	return fs.Glob(embeddedFS, normalizeEmbeddedPath(pattern))
}
