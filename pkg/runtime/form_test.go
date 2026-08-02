package runtime

import (
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// strArray builds the *[]script.Value that a Duso array arrives as.
func strArray(items ...string) *[]script.Value {
	arr := make([]script.Value, len(items))
	for i, s := range items {
		arr[i] = script.NewString(s)
	}
	return &arr
}

func formatForm(t *testing.T, object map[string]any) string {
	t.Helper()
	out, err := builtinFormatForm(nil, map[string]any{"0": object})
	if err != nil {
		t.Fatalf("format_form failed: %v", err)
	}
	s, ok := out.(string)
	if !ok {
		t.Fatalf("format_form returned %T, want string", out)
	}
	return s
}

// parseForm runs parse_form and unwraps the object for assertions.
func parseForm(t *testing.T, str string) map[string]script.Value {
	t.Helper()
	out, err := builtinParseForm(nil, map[string]any{"0": str})
	if err != nil {
		t.Fatalf("parse_form failed: %v", err)
	}
	ref, ok := out.(*script.ValueRef)
	if !ok {
		t.Fatalf("parse_form returned %T, want *script.ValueRef", out)
	}
	if !ref.Val.IsObject() {
		t.Fatalf("parse_form returned %s, want an object", ref.Val.Type)
	}
	return ref.Val.AsObject()
}

func TestFormatForm(t *testing.T) {
	tests := []struct {
		name   string
		object map[string]any
		want   string
	}{
		{"basic", map[string]any{"client_id": "abc"}, "client_id=abc"},
		{"keys sorted", map[string]any{"z": "1", "a": "2", "m": "3"}, "a=2&m=3&z=1"},
		{"escapes reserved chars", map[string]any{"u": "https://a.b/cb?x=1&y=2"}, "u=https%3A%2F%2Fa.b%2Fcb%3Fx%3D1%26y%3D2"},
		{"space becomes plus", map[string]any{"q": "hello world"}, "q=hello+world"},
		{"utf8", map[string]any{"name": "José"}, "name=Jos%C3%A9"},
		// Numbers use Duso's spelling: 2, not 2.000000 or 2.0.
		{"integral number", map[string]any{"page": float64(2)}, "page=2"},
		{"fractional number", map[string]any{"ratio": 1.5}, "ratio=1.5"},
		{"boolean", map[string]any{"active": true}, "active=true"},
		{"empty object", map[string]any{}, ""},
		{"empty string keeps key", map[string]any{"a": ""}, "a="},
		// nil means "not provided" — the key is dropped, not sent empty.
		{"nil omitted", map[string]any{"a": "1", "b": nil, "c": "3"}, "a=1&c=3"},
		{"array repeats key", map[string]any{"tag": strArray("js", "web")}, "tag=js&tag=web"},
		{"empty array emits nothing", map[string]any{"a": "1", "tag": strArray()}, "a=1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatForm(t, tc.object); got != tc.want {
				t.Errorf("format_form() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatFormRejectsNonScalars(t *testing.T) {
	nested := map[string]any{"meta": map[string]any{"a": "1"}}
	if _, err := builtinFormatForm(nil, map[string]any{"0": nested}); err == nil {
		t.Error("format_form() accepted a nested object, want an error naming the key")
	}

	fn := map[string]any{"cb": &script.ValueRef{Val: script.NewGoFunction(
		func(*script.Evaluator, map[string]any) (any, error) { return nil, nil },
	)}}
	if _, err := builtinFormatForm(nil, map[string]any{"0": fn}); err == nil {
		t.Error("format_form() accepted a function, want an error naming the key")
	}
}

func TestFormatFormRequiresObject(t *testing.T) {
	if _, err := builtinFormatForm(nil, map[string]any{}); err == nil {
		t.Error("format_form() accepted no argument, want an error")
	}
	if _, err := builtinFormatForm(nil, map[string]any{"0": "a=1"}); err == nil {
		t.Error("format_form() accepted a string, want an error")
	}
}

func TestParseForm(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want map[string]string
	}{
		{"basic", "code=xyz&state=abc", map[string]string{"code": "xyz", "state": "abc"}},
		{"leading question mark", "?code=xyz", map[string]string{"code": "xyz"}},
		{"percent decoded", "u=https%3A%2F%2Fa.b%2Fcb", map[string]string{"u": "https://a.b/cb"}},
		{"plus is space", "q=a+b", map[string]string{"q": "a b"}},
		{"empty value", "a=&b=2", map[string]string{"a": "", "b": "2"}},
		{"bare key", "a&b=2", map[string]string{"a": "", "b": "2"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseForm(t, tc.str)
			if len(got) != len(tc.want) {
				t.Fatalf("parse_form() returned %d keys, want %d", len(got), len(tc.want))
			}
			for k, want := range tc.want {
				v, ok := got[k]
				if !ok {
					t.Fatalf("parse_form() missing key %q", k)
				}
				if !v.IsString() || v.AsString() != want {
					t.Errorf("parse_form()[%q] = %v, want %q", k, v, want)
				}
			}
		})
	}
}

// A repeated key becomes an array, matching request().query and request().cookies.
func TestParseFormRepeatedKey(t *testing.T) {
	got := parseForm(t, "tag=js&tag=web")
	v, ok := got["tag"]
	if !ok {
		t.Fatal("parse_form() missing key \"tag\"")
	}
	if !v.IsArray() {
		t.Fatalf("parse_form()[\"tag\"] is %s, want an array", v.Type)
	}
	items := v.AsArray()
	if len(items) != 2 || items[0].AsString() != "js" || items[1].AsString() != "web" {
		t.Errorf("parse_form()[\"tag\"] = %v, want [js web]", items)
	}
}

func TestParseFormEmpty(t *testing.T) {
	if got := parseForm(t, ""); len(got) != 0 {
		t.Errorf("parse_form(\"\") returned %d keys, want 0", len(got))
	}
}

func TestParseFormMalformed(t *testing.T) {
	if _, err := builtinParseForm(nil, map[string]any{"0": "a=%zz"}); err == nil {
		t.Error("parse_form() accepted invalid percent-encoding, want an error")
	}
}

func TestParseFormRequiresString(t *testing.T) {
	if _, err := builtinParseForm(nil, map[string]any{}); err == nil {
		t.Error("parse_form() accepted no argument, want an error")
	}
	if _, err := builtinParseForm(nil, map[string]any{"0": float64(1)}); err == nil {
		t.Error("parse_form() accepted a number, want an error")
	}
}

// Anything format_form() writes, parse_form() must read back unchanged.
func TestFormRoundTrip(t *testing.T) {
	object := map[string]any{
		"q":     "hello world & more",
		"u":     "https://a.b/cb?x=1",
		"name":  "José",
		"tag":   strArray("js", "web"),
		"page":  float64(2),
		"empty": "",
	}

	got := parseForm(t, formatForm(t, object))

	for _, k := range []string{"q", "u", "name", "empty"} {
		if got[k].AsString() != object[k].(string) {
			t.Errorf("round trip %q = %q, want %q", k, got[k].AsString(), object[k])
		}
	}
	if got["page"].AsString() != "2" {
		t.Errorf("round trip \"page\" = %q, want \"2\"", got["page"].AsString())
	}
	tags := got["tag"].AsArray()
	if len(tags) != 2 || tags[0].AsString() != "js" || tags[1].AsString() != "web" {
		t.Errorf("round trip \"tag\" = %v, want [js web]", tags)
	}
}
