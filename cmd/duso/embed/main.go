package main

import (
	"fmt"
	"io"
	"os"

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

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
