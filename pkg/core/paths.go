package core

import "strings"

// HasPathPrefix checks if path starts with a prefix, handling both / and \ separators.
// prefixCore should be "EMBED" or "STORE" (without slashes).
func HasPathPrefix(path, prefixCore string) bool {
	return strings.HasPrefix(path, "/"+prefixCore+"/") ||
		strings.HasPrefix(path, "\\"+prefixCore+"\\")
}

// TrimPathPrefix removes a prefix from path, handling both / and \ separators.
// prefixCore should be "EMBED" or "STORE" (without slashes).
func TrimPathPrefix(path, prefixCore string) string {
	if strings.HasPrefix(path, "/"+prefixCore+"/") {
		return strings.TrimPrefix(path, "/"+prefixCore+"/")
	}
	if strings.HasPrefix(path, "\\"+prefixCore+"\\") {
		return strings.TrimPrefix(path, "\\"+prefixCore+"\\")
	}
	return path
}

// IsAbsoluteOrSpecial checks if a path is absolute or a special prefix path (like /EMBED/ or \EMBED\).
// Returns true if path starts with / or \ (or is an absolute OS path).
func IsAbsoluteOrSpecial(path string) bool {
	return strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\")
}
