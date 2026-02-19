package core

import (
	"os"
	"path"
	"strings"
)

// pathutil provides path operations that always use forward slashes internally.
// This allows consistent handling across platforms, especially for embed.FS which requires forward slashes.

// IsAbsolute reports whether the path is absolute.
// A path is absolute if it starts with "/".
func IsAbsolute(p string) bool {
	return strings.HasPrefix(p, "/")
}

// IsSpecial reports whether the path starts with a special prefix like /EMBED/ or /STORE/.
func IsSpecial(p string) bool {
	return strings.HasPrefix(p, "/EMBED/") || strings.HasPrefix(p, "/STORE/")
}

// IsAbsoluteOrSpecial reports whether the path is absolute or starts with a special prefix.
func IsAbsoluteOrSpecial(p string) bool {
	return IsAbsolute(p) || IsSpecial(p)
}

// Clean returns the shortest path equivalent to p, removing . and .. elements.
// All paths use forward slashes.
func Clean(p string) string {
	return path.Clean(p)
}

// Join joins any number of path elements into a single path using forward slashes.
func Join(elem ...string) string {
	return path.Join(elem...)
}

// Dir returns all but the last element of path.
func Dir(p string) string {
	return path.Dir(p)
}

// Base returns the last element of path.
func Base(p string) string {
	return path.Base(p)
}

// Rel returns a relative path that is lexicographically equivalent to targpath
// when joined to basepath.
// Returns an error if both paths cannot be made relative to each other.
func Rel(basepath, targpath string) (string, error) {
	// Handle special cases
	if IsAbsolute(targpath) && !IsAbsolute(basepath) {
		return targpath, nil
	}
	if !IsAbsolute(targpath) && IsAbsolute(basepath) {
		return "", os.ErrNotExist
	}

	// Normalize paths
	basepath = Clean(basepath)
	targpath = Clean(targpath)

	// If they're the same, return .
	if basepath == targpath {
		return ".", nil
	}

	// If target starts with base path, return the suffix
	if strings.HasPrefix(targpath, basepath+"/") {
		return targpath[len(basepath)+1:], nil
	}

	// If base path is / and target is absolute, return target without leading /
	if basepath == "/" && IsAbsolute(targpath) {
		return targpath[1:], nil
	}

	// For more complex relative paths, use string manipulation
	baseparts := strings.Split(basepath, "/")
	targparts := strings.Split(targpath, "/")

	// Find common prefix
	var common int
	for i := 0; i < len(baseparts) && i < len(targparts); i++ {
		if baseparts[i] == targparts[i] {
			common = i + 1
		} else {
			break
		}
	}

	// Count how many directories we need to go up
	upcount := len(baseparts) - common

	// Build result
	result := strings.Repeat("../", upcount)
	if common < len(targparts) {
		result += strings.Join(targparts[common:], "/")
	}

	// Remove trailing / if any
	result = strings.TrimSuffix(result, "/")
	if result == "" {
		result = "."
	}

	return result, nil
}

// Abs returns an absolute representation of path.
// If the path is not absolute it is joined with the current working directory.
func Abs(p string) (string, error) {
	if IsAbsolute(p) {
		return p, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Normalize working directory to forward slashes
	wd = strings.ReplaceAll(wd, "\\", "/")

	return Join(wd, p), nil
}

// SplitList splits a path list separated by colons (Unix-style).
// This is used for environment variables like DUSO_LIB.
// We always use colon separation internally, even on Windows.
func SplitList(s string) []string {
	// Normalize backslashes to forward slashes in the input
	s = strings.ReplaceAll(s, "\\", "/")
	// Split on colon
	return strings.Split(s, ":")
}

// HasPathPrefix checks if path starts with a special prefix.
// prefixCore should be "EMBED" or "STORE" (without slashes).
func HasPathPrefix(path, prefixCore string) bool {
	return strings.HasPrefix(path, "/"+prefixCore+"/")
}

// TrimPathPrefix removes a special prefix from path.
// prefixCore should be "EMBED" or "STORE" (without slashes).
func TrimPathPrefix(path, prefixCore string) string {
	if strings.HasPrefix(path, "/"+prefixCore+"/") {
		return strings.TrimPrefix(path, "/"+prefixCore+"/")
	}
	return path
}

// Match reports whether name matches the shell pattern.
// Pattern uses forward slashes and supports * and ? wildcards.
func Match(pattern, name string) (bool, error) {
	return path.Match(pattern, name)
}
