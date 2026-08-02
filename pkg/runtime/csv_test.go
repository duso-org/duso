package runtime

import (
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// rows builds the *[]Value that format_csv() receives for an array of rows.
func rows(v ...script.Value) map[string]any {
	arr := v
	return map[string]any{"0": &arr}
}

func formatCSV(t *testing.T, args map[string]any) string {
	t.Helper()
	out, err := builtinFormatCSV(nil, args)
	if err != nil {
		t.Fatalf("format_csv failed: %v", err)
	}
	s, ok := out.(string)
	if !ok {
		t.Fatalf("format_csv returned %T, want string", out)
	}
	return s
}

// parseCSV runs parse_csv through the map calling convention and flattens the
// result to plain strings for assertions.
func parseCSV(t *testing.T, args map[string]any) [][]string {
	t.Helper()
	out, err := builtinParseCSV(nil, args)
	if err != nil {
		t.Fatalf("parse_csv failed: %v", err)
	}
	v, ok := out.(script.Value)
	if !ok {
		t.Fatalf("parse_csv returned %T, want script.Value", out)
	}
	return csvRows(t, v)
}

func csvRows(t *testing.T, v script.Value) [][]string {
	t.Helper()
	if !v.IsArray() {
		t.Fatalf("parse_csv returned %s, want an array", v.Type)
	}
	records := v.AsArray()
	out := make([][]string, 0, len(records))
	for _, record := range records {
		if !record.IsArray() {
			t.Fatalf("record is %s, want an array", record.Type)
		}
		fields := record.AsArray()
		row := make([]string, 0, len(fields))
		for _, field := range fields {
			if !field.IsString() {
				t.Fatalf("field is %s, want a string", field.Type)
			}
			row = append(row, field.AsString())
		}
		out = append(out, row)
	}
	return out
}

func sameRows(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

func TestFormatCSVNilIsAnEmptyField(t *testing.T) {
	// Regression: nil rendered as the literal text "nil", which read back from
	// parse_csv as a four-character string.
	got := formatCSV(t, rows(arr(str("a"), script.NewNil(), str("b"))))
	if want := "a,,b\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatCSVEmptyRow(t *testing.T) {
	// Regression: an empty array has a nil backing slice, so the old nil check
	// mistook it for a non-row and wrote the Duso rendering "[]".
	got := formatCSV(t, rows(arr(str("a")), arr(), arr(str("b"))))
	if want := "a\n\nb\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCSVMultiByteDelimiter(t *testing.T) {
	// Regression: rune(delimiter[0]) took the first byte, so any non-ASCII
	// delimiter became a garbage rune that never matched.
	const arrow = "→"

	got := formatCSV(t, map[string]any{
		"0": func() *[]script.Value { a := []script.Value{arr(str("a"), str("b"))}; return &a }(),
		"1": arrow,
	})
	if want := "a→b\n"; got != want {
		t.Fatalf("format: got %q, want %q", got, want)
	}

	parsed := parseCSV(t, map[string]any{"0": "a→b→c", "1": arrow})
	if want := [][]string{{"a", "b", "c"}}; !sameRows(parsed, want) {
		t.Errorf("got %v, want %v", parsed, want)
	}
}

func TestCSVTabDelimiter(t *testing.T) {
	// TSV is the delimiter the primer documents, and quoting has to key off the
	// active delimiter: a comma is an ordinary character here, a tab is not.
	got := formatCSV(t, map[string]any{
		"0": func() *[]script.Value {
			a := []script.Value{
				arr(str("a,b"), str("c")),
				arr(str("has\ttab"), script.NewNil()),
			}
			return &a
		}(),
		"1": "\t",
	})
	if want := "a,b\tc\n\"has\ttab\"\t\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	parsed := parseCSV(t, map[string]any{"0": "a\tb\tc\nd\te\tf", "1": "\t"})
	if want := [][]string{{"a", "b", "c"}, {"d", "e", "f"}}; !sameRows(parsed, want) {
		t.Errorf("got %v, want %v", parsed, want)
	}
}

func TestCSVRoundTripsEmptyValues(t *testing.T) {
	encoded := formatCSV(t, rows(arr(str("x"), script.NewNil())))

	parsed := parseCSV(t, map[string]any{"0": encoded})
	if want := [][]string{{"x", ""}}; !sameRows(parsed, want) {
		t.Errorf("got %v, want %v", parsed, want)
	}
}

func TestParseCSVCallingConventionsAgree(t *testing.T) {
	// A fast builtin must be semantically identical to its map-based twin —
	// named args and indirect calls still take the map path.
	cases := []struct{ text, delim string }{
		{"a,b,c\nd,e,f", ""},
		{"a\tb\nc\td", "\t"},
		{"a→b→c", "→"},
		{"", ""},
		{`"quoted,field",x`, ""},
		{"trailing,\n,leading", ""},
	}

	for _, tc := range cases {
		mapArgs := map[string]any{"0": tc.text}
		fastArgs := []script.Value{script.NewString(tc.text)}
		if tc.delim != "" {
			mapArgs["1"] = tc.delim
			fastArgs = append(fastArgs, script.NewString(tc.delim))
		}

		viaMap := parseCSV(t, mapArgs)

		out, err := fastParseCSV(nil, fastArgs)
		if err != nil {
			t.Fatalf("fast parse_csv failed on %q: %v", tc.text, err)
		}
		viaFast := csvRows(t, out)

		if !sameRows(viaMap, viaFast) {
			t.Errorf("conventions disagree on %q: map=%v fast=%v", tc.text, viaMap, viaFast)
		}
	}
}

func TestParseCSVRejectsNonString(t *testing.T) {
	if _, err := fastParseCSV(nil, []script.Value{num(42)}); err == nil {
		t.Error("fast parse_csv should reject a non-string argument")
	}
	if _, err := builtinParseCSV(nil, map[string]any{"0": 42.0}); err == nil {
		t.Error("parse_csv should reject a non-string argument")
	}
}

func TestFormatCSVScalarRow(t *testing.T) {
	// A bare (non-array) row is still accepted as a single-field record.
	got := formatCSV(t, rows(str("solo"), arr(str("a"), str("b"))))
	if want := "solo\na,b\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
