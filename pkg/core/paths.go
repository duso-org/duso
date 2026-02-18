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
