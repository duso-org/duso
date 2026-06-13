package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duso-org/duso/pkg/core"
)

// This utility copies files or directories recursively for go:generate.
// Used to stage stdlib, docs, contrib, and individual files for embedding in the duso binary.
//
// Cross-platform alternative to shell cp command.
//
// Usage: go run ./embed <source> <dest>
// - If source is a file, copies it to dest
// - If source is a directory, recursively copies it to dest
func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: embed <source> <destination>\n")
		os.Exit(1)
	}

	source := os.Args[1]
	dest := os.Args[2]

	// Check if source is a file or directory
	srcInfo, err := os.Stat(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if srcInfo.IsDir() {
		// Copy directory recursively
		if err := copyDir(source, dest); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Copy single file - ensure destination directory exists
		destDir := core.Dir(dest)
		if destDir != "." && destDir != "" {
			if err := os.MkdirAll(destDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		if err := copyFile(source, dest); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Remove destination if it exists
	if _, err := os.Stat(dst); err == nil {
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("failed to remove existing destination: %w", err)
		}
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		// Never stage macOS Finder junk into the embedded binary.
		if entry.Name() == ".DS_Store" {
			continue
		}
		srcPath := core.Join(src, entry.Name())
		dstPath := core.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// compressExts lists file extensions whose contents are gzip-compressed when
// staged for embedding. Filenames are left unchanged; at runtime
// cli.EmbeddedFileRead detects the gzip magic bytes and inflates transparently,
// so directory listing, glob, and stat all still see the original names.
var compressExts = map[string]bool{
	".du": true, ".md": true, ".txt": true, ".html": true,
	".js": true, ".css": true, ".json": true, ".csv": true,
}

// copyFile copies a single file, gzip-compressing its contents when the
// extension is in compressExts and compression actually shrinks the file
// (tiny files can grow once the gzip header is added — those are left raw,
// and the runtime's magic-byte sniff reads them unchanged).
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}

	out := data
	if compressExts[strings.ToLower(filepath.Ext(src))] {
		var buf bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		if _, werr := zw.Write(data); werr == nil && zw.Close() == nil && buf.Len() < len(data) {
			out = buf.Bytes()
		}
	}

	if err := os.WriteFile(dst, out, 0644); err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	return nil
}
