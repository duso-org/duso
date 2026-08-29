package runtime

import (
	"strings"
	"testing"
)

func encodeURL(t *testing.T, arg any) string {
	t.Helper()
	out, err := builtinEncodeURL(nil, map[string]any{"0": arg})
	if err != nil {
		t.Fatalf("encode_url failed: %v", err)
	}
	s, ok := out.(string)
	if !ok {
		t.Fatalf("encode_url returned %T, want string", out)
	}
	return s
}

func decodeURL(t *testing.T, arg any) string {
	t.Helper()
	out, err := builtinDecodeURL(nil, map[string]any{"0": arg})
	if err != nil {
		t.Fatalf("decode_url failed: %v", err)
	}
	s, ok := out.(string)
	if !ok {
		t.Fatalf("decode_url returned %T, want string", out)
	}
	return s
}

// The expected values here were produced by encodeURIComponent() in a browser.
// encode_url() exists to match it, so any divergence is a bug in encode_url().
func TestEncodeURLMatchesEncodeURIComponent(t *testing.T) {
	cases := []struct{ in, want string }{
		// Space is %20, not '+', so the result is legal in a path segment too.
		{"a b", "a%20b"},
		// Every reserved character is escaped -- that is what makes the result
		// safe to concatenate into a query string.
		{"a&b=c", "a%26b%3Dc"},
		{"a/b", "a%2Fb"},
		{"a?b#c", "a%3Fb%23c"},
		{"a+b", "a%2Bb"},
		{"user@example.com", "user%40example.com"},
		{"a:b;c,d$e", "a%3Ab%3Bc%2Cd%24e"},
		// Unreserved characters pass through untouched, including the five
		// JavaScript leaves alone that Go's QueryEscape would escape.
		{"AZaz09-_.~", "AZaz09-_.~"},
		{"!'()*", "!'()*"},
		// Non-ASCII is encoded as UTF-8 bytes.
		{"café", "caf%C3%A9"},
		{"日本", "%E6%97%A5%E6%9C%AC"},
		{"🙂", "%F0%9F%99%82"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := encodeURL(t, tc.in); got != tc.want {
			t.Errorf("encode_url(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// The fixup pass rewrites "%21" back to "!" in its own output. A literal percent
// in the input escapes to "%25" first, so an input of "%21" must survive intact.
func TestEncodeURLLiteralPercentEscapes(t *testing.T) {
	cases := []struct{ in, want string }{
		{"%21", "%2521"},
		{"%2A", "%252A"},
		{"100%", "100%25"},
		{"%", "%25"},
	}
	for _, tc := range cases {
		if got := encodeURL(t, tc.in); got != tc.want {
			t.Errorf("encode_url(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEncodeURLConvertsNonStrings(t *testing.T) {
	if got := encodeURL(t, nil); got != "" {
		t.Errorf("encode_url(nil) = %q, want \"\"", got)
	}
	if got := encodeURL(t, 42.0); got != "42" {
		t.Errorf("encode_url(42) = %q, want \"42\"", got)
	}
	if got := encodeURL(t, true); got != "true" {
		t.Errorf("encode_url(true) = %q, want \"true\"", got)
	}
}

func TestEncodeURLMissingArgument(t *testing.T) {
	out, err := builtinEncodeURL(nil, map[string]any{})
	if err != nil {
		t.Fatalf("encode_url() with no argument failed: %v", err)
	}
	if out != "" {
		t.Errorf("encode_url() with no argument = %q, want \"\"", out)
	}
}

func TestEncodeURLNamedArgument(t *testing.T) {
	out, err := builtinEncodeURL(nil, map[string]any{"value": "a b"})
	if err != nil {
		t.Fatalf("encode_url(value:) failed: %v", err)
	}
	if out != "a%20b" {
		t.Errorf("encode_url(value: \"a b\") = %q, want %q", out, "a%20b")
	}
}

func TestDecodeURLMatchesDecodeURIComponent(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a%20b", "a b"},
		{"a%26b%3Dc", "a&b=c"},
		{"caf%C3%A9", "café"},
		{"%E6%97%A5%E6%9C%AC", "日本"},
		{"unchanged", "unchanged"},
		// Lowercase hex decodes the same as uppercase.
		{"%c3%a9", "é"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := decodeURL(t, tc.in); got != tc.want {
			t.Errorf("decode_url(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// '+' is form encoding's space, not a URL's. decodeURIComponent leaves it alone
// and so does decode_url(); parse_form() is where '+' becomes a space.
func TestDecodeURLLeavesPlusAlone(t *testing.T) {
	if got := decodeURL(t, "a+b"); got != "a+b" {
		t.Errorf("decode_url(%q) = %q, want %q", "a+b", got, "a+b")
	}
}

func TestDecodeURLRoundTrips(t *testing.T) {
	for _, s := range []string{
		"a b", "a&b=c", "a/b?c#d", "user@example.com", "100%", "%21",
		"café", "日本", "🙂", "!'()*", "a+b", "",
	} {
		if got := decodeURL(t, encodeURL(t, s)); got != s {
			t.Errorf("decode_url(encode_url(%q)) = %q, want %q", s, got, s)
		}
	}
}

func TestDecodeURLRejectsMalformedEscape(t *testing.T) {
	for _, in := range []string{"%2", "%", "%zz", "a%2Gb"} {
		if _, err := builtinDecodeURL(nil, map[string]any{"0": in}); err == nil {
			t.Errorf("decode_url(%q) succeeded, want an error", in)
		}
	}
}

func TestDecodeURLRequiresString(t *testing.T) {
	if _, err := builtinDecodeURL(nil, map[string]any{}); err == nil {
		t.Error("decode_url() with no argument succeeded, want an error")
	}
	_, err := builtinDecodeURL(nil, map[string]any{"0": 42.0})
	if err == nil {
		t.Fatal("decode_url(42) succeeded, want an error")
	}
	if !strings.Contains(err.Error(), "requires a string") {
		t.Errorf("decode_url(42) error = %q, want it to mention a string argument", err)
	}
}
