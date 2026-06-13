package cli

import (
	"bytes"
	"compress/gzip"
	"io"
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
// Text assets are gzip-compressed at build time (see cmd/duso/embed); this is
// the single choke point through which every embedded byte read passes, so
// inflating here makes compression transparent to all callers. Filenames are
// unchanged, so ReadDir/Glob/Stat are unaffected.
func EmbeddedFileRead(path string) ([]byte, error) {
	data, err := embeddedFS.ReadFile(normalizeEmbeddedPath(path))
	if err != nil {
		return nil, err
	}
	return maybeGunzip(data)
}

// maybeGunzip inflates data that carries the gzip magic header (0x1f 0x8b);
// anything else is returned unchanged. This lets the embed tree hold a mix of
// compressed (text) and raw (already-compressed, e.g. PNG) files without the
// reader needing to know which is which.
//
// Limitation: detection is by content, not intent. It cannot distinguish a file
// we gzipped for storage from a file that is *legitimately* gzip and should be
// returned raw. This matters for load_binary() on embedded assets: a binary that
// happens to begin with 0x1f 0x8b (an actual .gz, or a coincidental byte match)
// will be auto-inflated rather than handed back verbatim. Today the only embedded
// binaries are PNGs (magic 0x89 0x50), so there is no collision; if a real gzip
// asset is ever embedded and must be read raw, exclude it here or route it
// through a path that bypasses this sniff.
func maybeGunzip(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return data, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
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
