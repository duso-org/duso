package runtime

import (
	"strings"
	"testing"
)

func TestStaticCacheControlByExt(t *testing.T) {
	s := &HTTPServerValue{StaticCacheControl: "public, max-age=3600"}
	err := parseStaticCacheByExt(s, map[string]any{
		"html":           "no-cache",
		"jpg, .PNG,webp": "public, max-age=86400",
		"txt":            "",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	cases := map[string]string{
		"public/index.html":       "no-cache",
		"/EMBED/site/INDEX.HTML":  "no-cache",              // case-insensitive, like getContentType
		"public/a.png":            "public, max-age=86400", // ".PNG" key normalized
		"public/b.webp":           "public, max-age=86400",
		"public/app.js":           "public, max-age=3600", // no rule: default
		"public/LICENSE":          "public, max-age=3600", // no extension: default
		"public/notes.txt":        "",                     // "" sends no header
		"public/archive.tar.html": "no-cache",             // last dot wins
	}
	for path, want := range cases {
		if got := s.staticCacheControlFor(path); got != want {
			t.Errorf("staticCacheControlFor(%q) = %q, want %q", path, got, want)
		}
	}
	if got := s.staticCacheControlForExt("html"); got != "no-cache" {
		t.Errorf("directory listing (html) = %q, want no-cache", got)
	}
}

func TestStaticCacheControlStarReplacesDefault(t *testing.T) {
	s := &HTTPServerValue{StaticCacheControl: "public, max-age=3600"}
	if err := parseStaticCacheByExt(s, map[string]any{"*": "no-store", "css": "max-age=60"}); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := s.staticCacheControlFor("a.js"); got != "no-store" {
		t.Errorf("unlisted extension = %q, want the * value", got)
	}
	if got := s.staticCacheControlFor("a.css"); got != "max-age=60" {
		t.Errorf("listed extension = %q, want its own value", got)
	}
}

func TestStaticCacheControlParseErrors(t *testing.T) {
	bad := map[string]map[string]any{
		"listed more than once": {"html": "a", "htm,html": "b"},
		"must be a string":      {"html": 3600.0},
		"on its own":            {"js,*": "a"},
	}
	for want, cfg := range bad {
		err := parseStaticCacheByExt(&HTTPServerValue{}, cfg)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("config %v: got error %v, want one containing %q", cfg, err, want)
		}
	}
}

func TestEtagMatches(t *testing.T) {
	etag := staticETag([]byte("hello"))
	if etag != staticETag([]byte("hello")) || etag == staticETag([]byte("hello!")) {
		t.Fatalf("ETag must follow the contents: %s", etag)
	}

	cases := map[string]bool{
		etag:                     true,
		"W/" + etag:              true, // weak comparison for If-None-Match
		`"other", ` + etag:       true, // list
		"*":                      true,
		`"other"`:                false,
		"":                       false,
		etag[:len(etag)-2] + `"`: false, // a near miss is still a miss
	}
	for header, want := range cases {
		if got := etagMatches(header, etag); got != want {
			t.Errorf("etagMatches(%q) = %v, want %v", header, got, want)
		}
	}
}
