package runtime

import (
	"strings"
	"testing"
)

func TestGetContentTypeDefaults(t *testing.T) {
	cases := map[string]string{
		"song.opus":         "audio/ogg",
		"song.M4A":          "audio/mp4",
		"font.woff2":        "font/woff2",
		"app.wasm":          "application/wasm",
		"photo.avif":        "image/avif",
		"page.html":         "text/html; charset=utf-8",  // existing entries unchanged
		"data.bin":          "application/octet-stream",  // unknown extension: opaque bytes
		"LICENSE":           "text/plain; charset=utf-8", // no extension: still text
		"/EMBED/site/a.css": "text/css; charset=utf-8",
	}
	for name, want := range cases {
		if got := getContentType(name); got != want {
			t.Errorf("getContentType(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestMimeTypesConfig(t *testing.T) {
	s := &HTTPServerValue{}
	values, _, _, err := parseExtKeyed("mime_types", map[string]any{
		"woff2":      "application/x-custom-font", // override a built-in
		"glb, .GLTF": "model/gltf-binary",         // add new ones
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	s.MimeTypes = values

	cases := map[string]string{
		"fonts/a.woff2": "application/x-custom-font",
		"models/a.glb":  "model/gltf-binary",
		"models/a.GLTF": "model/gltf-binary",
		"a.woff":        "font/woff",                 // not configured: built-in
		"a.unknownext":  "application/octet-stream",  // not configured, not built-in
		"README":        "text/plain; charset=utf-8", // no extension
	}
	for name, want := range cases {
		if got := s.contentTypeFor(name); got != want {
			t.Errorf("contentTypeFor(%q) = %q, want %q", name, got, want)
		}
	}

	if _, _, _, err := parseExtKeyed("mime_types", map[string]any{"glb": "a", "gltf,glb": "b"}); err == nil ||
		!strings.Contains(err.Error(), "mime_types: extension \"glb\" is listed more than once") {
		t.Errorf("duplicate extension: got %v", err)
	}
}
