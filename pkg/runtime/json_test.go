package runtime

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// num/str/arr/obj build script.Values the way the evaluator does, so tests
// exercise the same representation the builtins actually receive.
func num(n float64) script.Value                 { return script.NewNumber(n) }
func str(s string) script.Value                  { return script.NewString(s) }
func arr(v ...script.Value) script.Value         { return script.NewArray(v) }
func obj(m map[string]script.Value) script.Value { return script.NewObject(m) }

func TestEncodeNumbersInsideArrays(t *testing.T) {
	// Regression: numbers keep their payload in Value.Num, so any encoder that
	// reads Value.Data turns every number in an array into null.
	tests := []struct {
		name string
		in   script.Value
		want string
	}{
		{"flat", arr(num(1), num(2.5), num(-3)), `[1,2.5,-3]`},
		{"nested arrays", arr(arr(num(7), num(8)), arr(num(9))), `[[7,8],[9]]`},
		{
			"objects inside arrays",
			arr(obj(map[string]script.Value{"id": num(42), "name": str("x")})),
			`[{"id":42,"name":"x"}]`,
		},
		{
			"array inside object",
			obj(map[string]script.Value{"scores": arr(num(1), num(2))}),
			`{"scores":[1,2]}`,
		},
		{"mixed", arr(num(1), str("x"), script.NewBool(true), script.NewNil()), `[1,"x",true,null]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodeValueJSON(tt.in, "")
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEncodeKeepsNilValuedKeys(t *testing.T) {
	// Duso objects keep keys assigned nil (unlike Lua), so JSON must emit them.
	in := obj(map[string]script.Value{"a": num(1), "b": script.NewNil(), "c": str("z")})
	got, err := encodeValueJSON(in, "")
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if want := `{"a":1,"b":null,"c":"z"}`; string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestEncodeMatchesEncodingJSON(t *testing.T) {
	// The generic (map[string]any) path must stay byte-compatible with
	// encoding/json, which is what these builtins used to emit.
	cases := []any{
		nil,
		true,
		false,
		float64(0),
		float64(-1),
		float64(1) / 3,
		1e21,
		1e-7,
		math.MaxInt64 * 4.0,
		0.1 + 0.2,
		"",
		"plain",
		"quote\" backslash\\ newline\n tab\t",
		"html <b>&</b>",
		"unicode é ☃ 𝄞",
		"line\u2028para\u2029sep",
		"control\x00\x01\x1f",
		[]any{},
		map[string]any{},
		[]any{float64(1), "two", nil, true},
		map[string]any{"z": float64(1), "a": []any{float64(2)}, "m": nil},
		map[string]any{"nested": map[string]any{"deep": []any{map[string]any{"x": float64(9)}}}},
	}

	for _, c := range cases {
		want, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("encoding/json rejected %#v: %v", c, err)
		}
		got, err := encodeJSON(c, "")
		if err != nil {
			t.Fatalf("encodeJSON(%#v) failed: %v", c, err)
		}
		if string(got) != string(want) {
			t.Errorf("encodeJSON(%#v) = %s, encoding/json = %s", c, got, want)
		}
	}
}

func TestEncodeIndentMatchesEncodingJSON(t *testing.T) {
	cases := []any{
		map[string]any{"a": []any{float64(1), "x"}, "b": map[string]any{"c": float64(2)}},
		[]any{},
		map[string]any{},
		[]any{map[string]any{"k": []any{}}},
	}

	for _, c := range cases {
		want, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		got, err := encodeJSON(c, "  ")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("indent mismatch\n got: %s\nwant: %s", got, want)
		}
	}
}

func TestEncodeRejectsNaNAndInf(t *testing.T) {
	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := encodeValueJSON(num(f), ""); err == nil {
			t.Errorf("encoding %v should fail", f)
		}
	}
}

func TestEncodeCircularReference(t *testing.T) {
	// A self-referential array used to recurse until the goroutine stack blew.
	elems := []script.Value{num(1)}
	self := script.NewArray(elems)
	inner := self.Data.(*[]script.Value)
	*inner = append(*inner, self)

	if _, err := encodeValueJSON(self, ""); err == nil {
		t.Fatal("expected an error for a circular array")
	} else if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error should mention circularity, got: %v", err)
	}

	o := map[string]script.Value{"x": num(1)}
	ov := script.NewObject(o)
	o["self"] = ov
	if _, err := encodeValueJSON(ov, ""); err == nil {
		t.Fatal("expected an error for a circular object")
	}
}

func TestDecodeBasics(t *testing.T) {
	tests := []struct {
		src  string
		want string // re-encoded form
	}{
		{`null`, `null`},
		{`true`, `true`},
		{`false`, `false`},
		{`0`, `0`},
		{`-17`, `-17`},
		{`2.5`, `2.5`},
		{`1e3`, `1000`},
		{`-1.5e-3`, `-0.0015`},
		{`123456789012345678901234567890`, `1.2345678901234568e+29`},
		{`""`, `""`},
		{`"hi"`, `"hi"`},
		{`[]`, `[]`},
		{`{}`, `{}`},
		{`[1,2,3]`, `[1,2,3]`},
		{`{"a":1,"b":[2,{"c":null}]}`, `{"a":1,"b":[2,{"c":null}]}`},
		{"  \t\r\n [ 1 , 2 ] \n ", `[1,2]`},
		{`{"a":1,"a":2}`, `{"a":2}`}, // duplicate keys: last wins
		{`"Aé😀"`, `"Aé😀"`},
		{`"tab\tnl\nqt\"sl\\sol\/bs\bff\f"`, `"tab\tnl\nqt\"sl\\sol/bs\bff\f"`},
		{`-0`, `-0`},
		{`0.0`, `0`},
		{`0.5`, `0.5`},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			v, err := decodeJSON([]byte(tt.src))
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			got, err := encodeValueJSON(v, "")
			if err != nil {
				t.Fatalf("re-encode failed: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDecodeProducesRealValueTypes(t *testing.T) {
	v, err := decodeJSON([]byte(`{"n":5,"s":"x","b":true,"z":null,"a":[1]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !v.IsObject() {
		t.Fatalf("expected object, got %s", v.Type)
	}
	o := v.AsObject()
	if !o["n"].IsNumber() || o["n"].AsNumber() != 5 {
		t.Errorf("n: expected number 5, got %v", o["n"])
	}
	if !o["s"].IsString() || o["s"].AsString() != "x" {
		t.Errorf("s: expected string x, got %v", o["s"])
	}
	if !o["b"].IsBool() || !o["b"].AsBool() {
		t.Errorf("b: expected bool true, got %v", o["b"])
	}
	if !o["z"].IsNil() {
		t.Errorf("z: expected nil, got %v", o["z"])
	}
	if !o["a"].IsArray() || len(o["a"].AsArray()) != 1 {
		t.Errorf("a: expected 1-element array, got %v", o["a"])
	}
}

func TestDecodeRejectsInvalid(t *testing.T) {
	bad := []string{
		``, `   `, `{`, `[`, `}`, `]`, `[1,]`, `{"a":}`, `{"a" 1}`, `{a:1}`,
		`{"a":1,}`, `tru`, `nul`, `NaN`, `Infinity`, `01`, `-`, `1.`, `.5`,
		`1e`, `"unterminated`, `[1,2] extra`, `{"a":1} {"b":2}`, `'single'`,
		"\"raw\nnewline\"", `[1 2]`,
	}

	for _, src := range bad {
		t.Run(src, func(t *testing.T) {
			if _, err := decodeJSON([]byte(src)); err == nil {
				t.Errorf("decode(%q) should have failed", src)
			}
			// encoding/json should agree that this is invalid.
			var sink any
			if err := json.Unmarshal([]byte(src), &sink); err == nil {
				t.Errorf("encoding/json accepts %q — mismatch with our parser", src)
			}
		})
	}
}

func TestDecodeAgreesWithEncodingJSON(t *testing.T) {
	docs := []string{
		`{"users":[{"id":1,"name":"a","tags":["x","y"],"meta":{"ok":true,"score":1.5}}]}`,
		`[[[[[1]]]]]`,
		`{"big":1e100,"small":1e-100,"neg":-0.0}`,
		`{"esc":"a\"b\\c\/d","uni":"\u00e9\u2028\ud83d\ude00","lone":"\ud800"}`,
		`[0,-0,1,-1,0.5,-0.5,1e0,1E0,1e+0,1e-0]`,
	}

	for _, src := range docs {
		t.Run(src, func(t *testing.T) {
			var viaStdlib any
			if err := json.Unmarshal([]byte(src), &viaStdlib); err != nil {
				t.Fatalf("encoding/json rejected the fixture: %v", err)
			}
			want, err := json.Marshal(viaStdlib)
			if err != nil {
				t.Fatal(err)
			}

			v, err := decodeJSON([]byte(src))
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			got, err := encodeValueJSON(v, "")
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}

func TestDecodeDeepNestingIsBounded(t *testing.T) {
	src := strings.Repeat("[", maxJSONDepth+10) + strings.Repeat("]", maxJSONDepth+10)
	if _, err := decodeJSON([]byte(src)); err == nil {
		t.Error("expected deeply nested input to be rejected")
	}
}

func TestRoundTrip(t *testing.T) {
	original := obj(map[string]script.Value{
		"id":     num(12345),
		"score":  num(1.5),
		"name":   str("user-1"),
		"active": script.NewBool(true),
		"absent": script.NewNil(),
		"tags":   arr(str("a"), str("b")),
		"nested": obj(map[string]script.Value{"list": arr(num(1), num(2), num(3))}),
	})

	encoded, err := encodeValueJSON(original, "")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := encodeValueJSON(decoded, "")
	if err != nil {
		t.Fatal(err)
	}

	if string(encoded) != string(reencoded) {
		t.Errorf("round trip changed the document:\n first: %s\nsecond: %s", encoded, reencoded)
	}
}

func benchDocument() script.Value {
	objs := make([]script.Value, 5000)
	for i := range objs {
		objs[i] = obj(map[string]script.Value{
			"id":     num(float64(i)),
			"name":   str("user-" + itoa(i)),
			"active": script.NewBool(i%3 == 0),
			"score":  num(float64(i) * 1.5),
			"tags":   arr(str("a"), str("b"), str("c")),
		})
	}
	return script.NewArray(objs)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

func BenchmarkEncode(b *testing.B) {
	doc := benchDocument()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := encodeValueJSON(doc, ""); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	src, err := encodeValueJSON(benchDocument(), "")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := decodeJSON(src); err != nil {
			b.Fatal(err)
		}
	}
}

// FuzzDecodeAgainstStdlib checks that the hand-written parser accepts exactly
// what encoding/json accepts, and produces the same document when it does.
func FuzzDecodeAgainstStdlib(f *testing.F) {
	seeds := []string{
		`{"a":1}`, `[1,2,3]`, `null`, `"x"`, `-0`, `1e-7`, `{"a":[{"b":null}]}`,
		`"\ud83d\ude00"`, `"\u00e9"`, `[]`, `{}`, `0.1`, `  1  `,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, src string) {
		var viaStdlib any
		stdErr := json.Unmarshal([]byte(src), &viaStdlib)

		got, err := decodeJSON([]byte(src))

		if (stdErr == nil) != (err == nil) {
			t.Fatalf("acceptance mismatch for %q: stdlib err=%v, ours err=%v", src, stdErr, err)
		}
		if stdErr != nil {
			return
		}

		want, err2 := json.Marshal(viaStdlib)
		if err2 != nil {
			return // NaN/Inf can't arise from JSON input, but be safe
		}
		mine, err2 := encodeValueJSON(got, "")
		if err2 != nil {
			t.Fatalf("re-encode of %q failed: %v", src, err2)
		}
		if string(mine) != string(want) {
			t.Fatalf("round trip mismatch for %q:\n ours: %s\nstdlib: %s", src, mine, want)
		}
	})
}
