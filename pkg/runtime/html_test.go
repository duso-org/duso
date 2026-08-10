package runtime

import (
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

func escapeHTML(t *testing.T, arg any) string {
	t.Helper()
	out, err := builtinEscapeHTML(nil, map[string]any{"0": arg})
	if err != nil {
		t.Fatalf("escape_html failed: %v", err)
	}
	s, ok := out.(string)
	if !ok {
		t.Fatalf("escape_html returned %T, want string", out)
	}
	return s
}

func TestEscapeHTMLAllFiveCharacters(t *testing.T) {
	cases := []struct{ in, want string }{
		{"&", "&amp;"},
		{"<", "&lt;"},
		{">", "&gt;"},
		{`"`, "&quot;"},
		{"'", "&#39;"},
		{`&<>"'`, `&amp;&lt;&gt;&quot;&#39;`},
	}
	for _, tc := range cases {
		if got := escapeHTML(t, tc.in); got != tc.want {
			t.Errorf("escape_html(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEscapeHTMLNoDoubleEscaping(t *testing.T) {
	// The ampersand written for '<' must not itself be escaped again. A naive
	// sequence of Replace calls with '&' handled last produces &amp;lt;.
	if got, want := escapeHTML(t, "<"), "&lt;"; got != want {
		t.Errorf("escape_html(%q) = %q, want %q", "<", got, want)
	}
	if got, want := escapeHTML(t, "a & b < c"), "a &amp; b &lt; c"; got != want {
		t.Errorf("escape_html = %q, want %q", got, want)
	}
	// Pre-existing entities are escaped as literal text, which is correct:
	// escaping is not idempotent by design.
	if got, want := escapeHTML(t, "&amp;"), "&amp;amp;"; got != want {
		t.Errorf("escape_html(%q) = %q, want %q", "&amp;", got, want)
	}
}

func TestEscapeHTMLBlocksInjection(t *testing.T) {
	got := escapeHTML(t, `<script>alert('x')</script>`)
	want := `&lt;script&gt;alert(&#39;x&#39;)&lt;/script&gt;`
	if got != want {
		t.Errorf("escape_html = %q, want %q", got, want)
	}
}

func TestEscapeHTMLAttributeContext(t *testing.T) {
	// Both quote styles must be escaped or a value can break out of an
	// attribute delimited by either.
	if got, want := escapeHTML(t, `" onload="evil()`), `&quot; onload=&quot;evil()`; got != want {
		t.Errorf("double-quote break-out not escaped: %q", got)
	}
	if got, want := escapeHTML(t, `' onload='evil()`), `&#39; onload=&#39;evil()`; got != want {
		t.Errorf("single-quote break-out not escaped: %q", got)
	}
}

func TestEscapeHTMLNilAndMissing(t *testing.T) {
	if got := escapeHTML(t, nil); got != "" {
		t.Errorf("escape_html(nil) = %q, want empty string", got)
	}
	out, err := builtinEscapeHTML(nil, map[string]any{})
	if err != nil {
		t.Fatalf("escape_html() with no argument failed: %v", err)
	}
	if out != "" {
		t.Errorf("escape_html() = %v, want empty string", out)
	}
	// A nil arriving as a script value must behave the same way.
	if got := escapeHTML(t, script.Value{Type: script.VAL_NIL}); got != "" {
		t.Errorf("escape_html(nil value) = %q, want empty string", got)
	}
}

func TestEscapeHTMLCoercesNonStrings(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{float64(42), "42"},
		{float64(3.5), "3.5"},
		{true, "true"},
		{false, "false"},
	}
	for _, tc := range cases {
		if got := escapeHTML(t, tc.in); got != tc.want {
			t.Errorf("escape_html(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEscapeHTMLLeavesSafeTextAlone(t *testing.T) {
	safe := "Hello, world! Café ☕ 100% ok (really) [yes] {fine}"
	if got := escapeHTML(t, safe); got != safe {
		t.Errorf("escape_html altered safe text:\n got %q\nwant %q", got, safe)
	}
}

func TestEscapeHTMLNamedArgument(t *testing.T) {
	out, err := builtinEscapeHTML(nil, map[string]any{"value": "<b>"})
	if err != nil {
		t.Fatalf("escape_html failed: %v", err)
	}
	if out != "&lt;b&gt;" {
		t.Errorf("escape_html(value = %q) = %v, want %q", "<b>", out, "&lt;b&gt;")
	}
}
