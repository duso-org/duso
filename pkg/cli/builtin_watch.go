package cli

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// watcherState tracks file signatures per (scriptPath:watchPath)
var watcherState = make(map[string]uint64)

// builtinWatch watches a directory or file for changes.
//
// watch(path [, timeout = 30])
// - path (string) - Directory, file, or glob pattern to watch
// - timeout (number) - Max seconds to block waiting for changes (default: 30)
// - Returns true if files changed since last call, false if timeout reached
// - First call always returns false (no previous state to compare)
// - Blocks during timeout, checking every 1 second
// - Rejects /EMBED/ paths (read-only filesystem)
//
// Example:
//
//	if watch("./src", timeout = 30) then
//	  print("Files changed, rebuilding...")
//	  rebuild()
//	end
//
//	if watch("*.md") then  // Wildcard pattern
//	  print("Markdown files changed")
//	end
func builtinWatch(evaluator *script.Evaluator, args map[string]any) (any, error) {
	// Parse path argument (positional or named)
	path := ""
	if p, ok := args["0"].(string); ok {
		path = p
	} else if p, ok := args["path"].(string); ok {
		path = p
	} else {
		return nil, fmt.Errorf("watch() requires a path argument")
	}

	// Parse timeout argument (default 30 seconds)
	timeout := 30.0
	if t, ok := args["1"].(float64); ok {
		timeout = t
	} else if t, ok := args["timeout"].(float64); ok {
		timeout = t
	}

	// Reject /EMBED/ (read-only)
	if core.HasPathPrefix(path, "EMBED") {
		return nil, fmt.Errorf("watch() cannot watch /EMBED/: embedded filesystem is read-only")
	}

	// Get script directory from context
	scriptPath := ""
	gid := script.GetGoroutineID()
	if ctx, ok := script.GetRequestContext(gid); ok && ctx.Frame != nil {
		scriptPath = ctx.Frame.Filename
	}

	// Resolve path relative to script directory (default to current dir)
	resolvedPath := path
	if !filepath.IsAbs(path) && !strings.HasPrefix(path, "/") {
		if scriptPath != "" {
			scriptDir := core.Dir(scriptPath)
			resolvedPath = filepath.Join(scriptDir, path)
		} else {
			// No script context, use cwd
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("watch() could not determine working directory: %w", err)
			}
			resolvedPath = filepath.Join(cwd, path)
		}
	}

	// Verify root path exists and is accessible (but allow wildcards that might not exist yet)
	if !hasWildcard(path) {
		_, err := os.Stat(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("watch() cannot access %q: %w", path, err)
		}
	}

	// State key: scriptPath:watchPath
	stateKey := scriptPath + ":" + path

	// Detect if path contains wildcards
	hasWild := hasWildcard(path)

	// Get current signature
	var newSig uint64
	var err error

	if hasWild {
		// Wildcard: expand glob and hash matching files
		newSig, err = computeGlobSignature(resolvedPath)
	} else {
		// Check if it's a directory or file
		info, err := os.Stat(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("watch() error: %w", err)
		}

		if info.IsDir() {
			// Directory: hash all subdirectories
			newSig, err = computeDirectorySignature(resolvedPath)
		} else {
			// File: hash just this file's modtime
			newSig, err = computeFileSignature(resolvedPath)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("watch() error: %w", err)
	}

	// On first call, store signature and return false (no previous state to compare)
	oldSig, exists := watcherState[stateKey]
	if !exists {
		watcherState[stateKey] = newSig
		return false, nil
	}

	// Check if changed
	if newSig != oldSig {
		watcherState[stateKey] = newSig
		return true, nil
	}

	// No change yet - block and poll until timeout
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		if time.Now().After(deadline) {
			return false, nil // Timeout reached
		}

		time.Sleep(1 * time.Second)

		// Recompute signature
		if hasWild {
			newSig, err = computeGlobSignature(resolvedPath)
		} else {
			info, err := os.Stat(resolvedPath)
			if err != nil {
				continue // Skip stat errors during polling
			}

			if info.IsDir() {
				newSig, err = computeDirectorySignature(resolvedPath)
			} else {
				newSig, err = computeFileSignature(resolvedPath)
			}
		}

		if err != nil {
			continue // Skip errors during polling
		}

		if newSig != oldSig {
			watcherState[stateKey] = newSig
			return true, nil
		}
	}
}

// computeFileSignature hashes a single file's modtime
func computeFileSignature(path string) (uint64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	h := fnv.New64a()
	h.Write([]byte(path))
	h.Write([]byte(info.ModTime().String()))
	return h.Sum64(), nil
}

// computeDirectorySignature hashes all subdirectories' modtimes (not files)
func computeDirectorySignature(rootPath string) (uint64, error) {
	h := fnv.New64a()

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Only hash directories (not files)
		if info.IsDir() {
			h.Write([]byte(path))
			h.Write([]byte(info.ModTime().String()))
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return h.Sum64(), nil
}

// computeGlobSignature hashes files matching a glob pattern
func computeGlobSignature(pattern string) (uint64, error) {
	matches, err := ExpandGlob(pattern)
	if err != nil {
		return 0, err
	}

	h := fnv.New64a()
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue // Skip inaccessible files
		}

		h.Write([]byte(match))
		h.Write([]byte(info.ModTime().String()))
	}

	return h.Sum64(), nil
}
