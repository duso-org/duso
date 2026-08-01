package script

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"regexp"
	"testing"
)

// roundTrip encodes v, decodes it back, and fails if any byte is left over.
func roundTrip(t *testing.T, v any) any {
	t.Helper()

	buf, err := EncodeValue(nil, v)
	if err != nil {
		t.Fatalf("EncodeValue(%#v) failed: %v", v, err)
	}
	got, rest, err := DecodeValue(buf)
	if err != nil {
		t.Fatalf("DecodeValue failed for %#v: %v", v, err)
	}
	if len(rest) != 0 {
		t.Fatalf("DecodeValue left %d trailing bytes for %#v", len(rest), v)
	}
	return got
}

func TestCodecRoundTripScalars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  any
	}{
		{"nil", nil},
		{"true", true},
		{"false", false},
		{"zero", 0.0},
		{"one", 1.0},
		{"negative", -42.0},
		{"fraction", 3.14159},
		{"empty string", ""},
		{"string", "hello"},
		{"unicode string", "héllo → 世界 🐈"},
		{"string with nulls", "a\x00b"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := roundTrip(t, tc.val); !reflect.DeepEqual(got, tc.val) {
				t.Errorf("round trip: got %#v, want %#v", got, tc.val)
			}
		})
	}
}

func TestCodecRoundTripContainers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  any
	}{
		{"empty array", []any{}},
		{"flat array", []any{1.0, "two", true, nil}},
		{"empty object", map[string]any{}},
		{"flat object", map[string]any{"a": 1.0, "b": "two", "c": false}},
		{"nested", map[string]any{
			"list": []any{1.0, []any{2.0, 3.0}, map[string]any{"deep": true}},
			"obj":  map[string]any{"inner": map[string]any{"deeper": nil}},
		}},
		{"array of objects", []any{
			map[string]any{"id": 1.0},
			map[string]any{"id": 2.0},
		}},
		{"unicode keys", map[string]any{"héllo": 1.0, "世界": 2.0, "🐈": 3.0}},
		{"empty key", map[string]any{"": "empty"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := roundTrip(t, tc.val); !reflect.DeepEqual(got, tc.val) {
				t.Errorf("round trip: got %#v, want %#v", got, tc.val)
			}
		})
	}
}

// Identical values must produce identical bytes regardless of map iteration
// order, which is what makes checksums and divergence comparison possible.
func TestCodecDeterministicObjectOrdering(t *testing.T) {
	t.Parallel()

	obj := map[string]any{
		"zebra": 1.0, "alpha": 2.0, "mike": 3.0, "": 4.0,
		"héllo": 5.0, "delta": 6.0, "🐈": 7.0, "bravo": 8.0,
	}

	first, err := EncodeValue(nil, obj)
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}
	// Re-encode many times; Go randomizes map iteration order per range.
	for i := 0; i < 100; i++ {
		again, err := EncodeValue(nil, obj)
		if err != nil {
			t.Fatalf("EncodeValue failed on pass %d: %v", i, err)
		}
		if !bytes.Equal(first, again) {
			t.Fatalf("encoding is not deterministic: pass %d differs", i)
		}
	}
}

func TestCodecNumericEdges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  float64
	}{
		{"positive zero", 0},
		{"max exact int", maxExactInt},
		{"min exact int", -maxExactInt},
		{"just past max exact int", maxExactInt + 2},
		{"just past min exact int", -maxExactInt - 2},
		{"max float", math.MaxFloat64},
		{"smallest nonzero", math.SmallestNonzeroFloat64},
		{"tiny fraction", 1e-300},
		{"huge", 1e300},
		{"integral but huge", 1e17},
		{"negative fraction", -0.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := roundTrip(t, tc.val)
			f, ok := got.(float64)
			if !ok {
				t.Fatalf("got %T, want float64", got)
			}
			if f != tc.val {
				t.Errorf("round trip: got %v, want %v", f, tc.val)
			}
		})
	}
}

// Negative zero must not collapse to positive zero via the INT path.
func TestCodecNegativeZero(t *testing.T) {
	t.Parallel()

	got := roundTrip(t, math.Copysign(0, -1))
	f, ok := got.(float64)
	if !ok {
		t.Fatalf("got %T, want float64", got)
	}
	if f != 0 || !math.Signbit(f) {
		t.Errorf("negative zero did not survive: got %v (signbit %v)", f, math.Signbit(f))
	}
}

func TestCodecNaNAndInf(t *testing.T) {
	t.Parallel()

	t.Run("NaN", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, math.NaN())
		f, ok := got.(float64)
		if !ok || !math.IsNaN(f) {
			t.Errorf("NaN did not survive: got %#v", got)
		}
	})

	for _, sign := range []int{1, -1} {
		inf := math.Inf(sign)
		t.Run("Inf", func(t *testing.T) {
			t.Parallel()
			got := roundTrip(t, inf)
			f, ok := got.(float64)
			if !ok || !math.IsInf(f, sign) {
				t.Errorf("Inf(%d) did not survive: got %#v", sign, got)
			}
		})
	}
}

// The varint path must be chosen for integral values in range and only there.
func TestCodecIntTagSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		val     float64
		wantTag byte
	}{
		{"small int", 42, tagInt},
		{"negative int", -42, tagInt},
		{"zero", 0, tagInt},
		{"max exact int", maxExactInt, tagInt},
		{"min exact int", -maxExactInt, tagInt},
		{"beyond exact range", maxExactInt * 4, tagNumber},
		{"fraction", 0.5, tagNumber},
		{"negative zero", math.Copysign(0, -1), tagNumber},
		{"NaN", math.NaN(), tagNumber},
		{"Inf", math.Inf(1), tagNumber},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf, err := EncodeValue(nil, tc.val)
			if err != nil {
				t.Fatalf("EncodeValue failed: %v", err)
			}
			if buf[0] != tc.wantTag {
				t.Errorf("tag for %v: got 0x%02X, want 0x%02X", tc.val, buf[0], tc.wantTag)
			}
		})
	}
}

// Small integers are the common case in real data, so the varint form must
// actually be smaller than the fixed f64 form.
func TestCodecIntIsCompact(t *testing.T) {
	t.Parallel()

	small, err := EncodeValue(nil, 42.0)
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}
	big, err := EncodeValue(nil, 0.5)
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}
	if len(small) != 2 {
		t.Errorf("small int encoding: got %d bytes, want 2", len(small))
	}
	if len(big) != 9 {
		t.Errorf("float encoding: got %d bytes, want 9", len(big))
	}
}

// Function elision must match the deep_copy() builtin exactly: nil at top level
// and inside arrays (position preserved), key dropped inside objects.
func TestCodecFunctionElision(t *testing.T) {
	t.Parallel()

	fn := &ValueRef{Val: NewGoFunction(func(*Evaluator, map[string]any) (any, error) {
		return nil, nil
	})}

	t.Run("top level becomes nil", func(t *testing.T) {
		t.Parallel()
		if got := roundTrip(t, fn); got != nil {
			t.Errorf("got %#v, want nil", got)
		}
	})

	t.Run("array element becomes nil, position preserved", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, []any{1.0, fn, 3.0})
		want := []any{1.0, nil, 3.0}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})

	t.Run("object key dropped entirely", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, map[string]any{"keep": 1.0, "drop": fn})
		want := map[string]any{"keep": 1.0}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})

	t.Run("nested object key dropped", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, map[string]any{"outer": map[string]any{"keep": 1.0, "drop": fn}})
		want := map[string]any{"outer": map[string]any{"keep": 1.0}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})
}

func TestCodecBinary(t *testing.T) {
	t.Parallel()

	payload := []byte{0x00, 0x01, 0xFF, 0xFE, 'h', 'i', 0x00}
	bin := NewBinary(payload)
	bin.AsBinary().Metadata = map[string]Value{
		"filename":     NewString("photo.png"),
		"content_type": NewString("image/png"),
		"size":         NewNumber(float64(len(payload))),
	}

	got := roundTrip(t, &ValueRef{Val: bin})
	ref, ok := got.(*ValueRef)
	if !ok {
		t.Fatalf("got %T, want *ValueRef", got)
	}
	if ref.Val.Type != VAL_BINARY {
		t.Fatalf("got type %s, want binary", ref.Val.Type)
	}

	decoded := ref.Val.AsBinary()
	if !bytes.Equal(*decoded.Data, payload) {
		t.Errorf("payload: got %v, want %v", *decoded.Data, payload)
	}
	if name := decoded.Metadata["filename"].AsString(); name != "photo.png" {
		t.Errorf("filename metadata: got %q, want %q", name, "photo.png")
	}
	if size := decoded.Metadata["size"].AsNumber(); size != float64(len(payload)) {
		t.Errorf("size metadata: got %v, want %v", size, len(payload))
	}
}

func TestCodecBinaryEmpty(t *testing.T) {
	t.Parallel()

	got := roundTrip(t, &ValueRef{Val: NewBinary(nil)})
	ref, ok := got.(*ValueRef)
	if !ok || ref.Val.Type != VAL_BINARY {
		t.Fatalf("got %#v, want a binary value", got)
	}
	if len(*ref.Val.AsBinary().Data) != 0 {
		t.Errorf("expected empty payload, got %d bytes", len(*ref.Val.AsBinary().Data))
	}
}

// The decoder must own its bytes: reusing the source buffer must not corrupt an
// already-decoded value.
func TestCodecBinaryDoesNotAliasInput(t *testing.T) {
	t.Parallel()

	payload := []byte("original")
	buf, err := EncodeValue(nil, &ValueRef{Val: NewBinary(payload)})
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}
	got, _, err := DecodeValue(buf)
	if err != nil {
		t.Fatalf("DecodeValue failed: %v", err)
	}
	for i := range buf {
		buf[i] = 0xFF // scribble over the source buffer
	}
	decoded := *got.(*ValueRef).Val.AsBinary().Data
	if !bytes.Equal(decoded, payload) {
		t.Errorf("decoded binary aliased the input buffer: got %q, want %q", decoded, payload)
	}
}

func TestCodecError(t *testing.T) {
	t.Parallel()

	original := NewErrorValue(NewString("something broke"), "file.du:12:3\n  at main()")
	got := roundTrip(t, &ValueRef{Val: original})

	ref, ok := got.(*ValueRef)
	if !ok || ref.Val.Type != VAL_ERROR {
		t.Fatalf("got %#v, want an error value", got)
	}
	ev := ref.Val.AsErrorVal()
	if msg := ev.Message.AsString(); msg != "something broke" {
		t.Errorf("message: got %q, want %q", msg, "something broke")
	}
	if ev.Stack != "file.du:12:3\n  at main()" {
		t.Errorf("stack: got %q", ev.Stack)
	}
}

// throw() can carry any value, not just a string, so the message must recurse
// through the codec.
func TestCodecErrorWithObjectMessage(t *testing.T) {
	t.Parallel()

	msg := NewObject(map[string]Value{
		"code":   NewNumber(404),
		"detail": NewString("not found"),
	})
	got := roundTrip(t, &ValueRef{Val: NewErrorValue(msg, "stack")})

	ev := got.(*ValueRef).Val.AsErrorVal()
	if ev.Message.Type != VAL_OBJECT {
		t.Fatalf("message type: got %s, want object", ev.Message.Type)
	}
	if code := ev.Message.AsObject()["code"].AsNumber(); code != 404 {
		t.Errorf("message.code: got %v, want 404", code)
	}
}

func TestCodecRegex(t *testing.T) {
	t.Parallel()

	pattern := `^[a-z]+\d{2,4}$`
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	got := roundTrip(t, &ValueRef{Val: NewRegex(pattern, compiled)})
	ref, ok := got.(*ValueRef)
	if !ok || ref.Val.Type != VAL_REGEX {
		t.Fatalf("got %#v, want a regex value", got)
	}

	rv := ref.Val.AsRegex()
	if rv.Pattern != pattern {
		t.Errorf("pattern: got %q, want %q", rv.Pattern, pattern)
	}
	if rv.Compiled == nil {
		t.Fatal("regex was not recompiled on decode")
	}
	if !rv.Compiled.MatchString("abc123") {
		t.Error("recompiled regex does not match what the original would")
	}
}

func TestCodecCode(t *testing.T) {
	t.Parallel()

	src := "x = 1\ny = x + 2\nreturn y"
	tokens, err := NewLexer(src).Tokenize()
	if err != nil {
		t.Fatalf("test setup tokenize: %v", err)
	}
	prog, err := NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("test setup parse: %v", err)
	}
	original := NewCode(src, prog, map[string]Value{"name": NewString("snippet")})

	got := roundTrip(t, &ValueRef{Val: original})
	ref, ok := got.(*ValueRef)
	if !ok || ref.Val.Type != VAL_CODE {
		t.Fatalf("got %#v, want a code value", got)
	}

	cv := ref.Val.AsCode()
	if cv.Source != src {
		t.Errorf("source: got %q, want %q", cv.Source, src)
	}
	if cv.Program == nil {
		t.Error("AST was not rebuilt on decode")
	}
	if name := cv.Metadata["name"].AsString(); name != "snippet" {
		t.Errorf("metadata name: got %q, want %q", name, "snippet")
	}
}

// A stored source that no longer parses must surface an error rather than
// silently decoding to something degraded.
func TestCodecCodeParseFailureSurfaces(t *testing.T) {
	t.Parallel()

	// Hand-build a code record carrying source this build cannot parse.
	buf := []byte{tagCode}
	buf = appendRawString(buf, "this is (((not duso")
	buf, err := encodeValObject(buf, map[string]Value{})
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	if _, _, err := DecodeValue(buf); err == nil {
		t.Fatal("expected an error decoding unparseable stored code, got nil")
	}
}

func TestCodecBinaryRefIsReserved(t *testing.T) {
	t.Parallel()

	if _, _, err := DecodeValue([]byte{tagBinaryRef}); err == nil {
		t.Fatal("expected BINARY_REF to be rejected, got nil error")
	}
}

func TestCodecRejectsUnknownTag(t *testing.T) {
	t.Parallel()

	for _, tag := range []byte{0x06, 0x12, 0x40, 0x80, 0xFF} {
		if _, _, err := DecodeValue([]byte{tag}); err == nil {
			t.Errorf("expected tag 0x%02X to be rejected, got nil error", tag)
		}
	}
}

// Corrupt or truncated input must produce an error, never a panic and never a
// huge speculative allocation.
func TestCodecRejectsTruncatedInput(t *testing.T) {
	t.Parallel()

	full, err := EncodeValue(nil, map[string]any{
		"list":   []any{1.0, 2.0, "three"},
		"nested": map[string]any{"a": true},
		"text":   "some string value",
	})
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}

	for cut := 0; cut < len(full); cut++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic decoding input truncated to %d bytes: %v", cut, r)
				}
			}()
			if _, _, err := DecodeValue(full[:cut]); err == nil {
				t.Errorf("expected error decoding input truncated to %d bytes", cut)
			}
		}()
	}
}

func TestCodecRejectsOversizedLength(t *testing.T) {
	t.Parallel()

	// A string claiming 2^40 bytes with nothing behind it.
	buf := binary.AppendUvarint([]byte{tagString}, 1<<40)
	if _, _, err := DecodeValue(buf); err == nil {
		t.Fatal("expected an error for a length prefix exceeding the input")
	}
}

func TestCodecDecodeLeavesRemainder(t *testing.T) {
	t.Parallel()

	buf, err := EncodeValue(nil, "first")
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}
	buf, err = EncodeValue(buf, "second")
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}

	first, rest, err := DecodeValue(buf)
	if err != nil {
		t.Fatalf("DecodeValue failed: %v", err)
	}
	if first != "first" {
		t.Errorf("first value: got %#v, want %q", first, "first")
	}
	second, rest, err := DecodeValue(rest)
	if err != nil {
		t.Fatalf("DecodeValue failed on second value: %v", err)
	}
	if second != "second" {
		t.Errorf("second value: got %#v, want %q", second, "second")
	}
	if len(rest) != 0 {
		t.Errorf("expected no trailing bytes, got %d", len(rest))
	}
}

// Values arriving as Value or *[]Value rather than an any-tree must encode to
// the same bytes, so the codec never depends on datastore normalization.
func TestCodecAcceptsValueForms(t *testing.T) {
	t.Parallel()

	viaAny, err := EncodeValue(nil, []any{1.0, "two", true})
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}
	arr := []Value{NewNumber(1), NewString("two"), NewBool(true)}
	viaPtr, err := EncodeValue(nil, &arr)
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}
	if !bytes.Equal(viaAny, viaPtr) {
		t.Errorf("*[]Value encoding differs from []any encoding:\n %v\n %v", viaAny, viaPtr)
	}

	viaValue, err := EncodeValue(nil, NewArray(arr))
	if err != nil {
		t.Fatalf("EncodeValue failed: %v", err)
	}
	if !bytes.Equal(viaAny, viaValue) {
		t.Errorf("Value encoding differs from []any encoding:\n %v\n %v", viaAny, viaValue)
	}
}

func TestCodecRejectsUnencodableGoType(t *testing.T) {
	t.Parallel()

	if _, err := EncodeValue(nil, make(chan int)); err == nil {
		t.Fatal("expected an error encoding an unsupported Go type, got nil")
	}
}

func TestCodecDeeplyNested(t *testing.T) {
	t.Parallel()

	var v any = "bottom"
	for i := 0; i < 200; i++ {
		v = map[string]any{"next": []any{v}}
	}
	if got := roundTrip(t, v); !reflect.DeepEqual(got, v) {
		t.Error("deeply nested value did not survive the round trip")
	}
}
