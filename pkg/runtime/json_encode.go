package runtime

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/duso-org/duso/pkg/script"
)

// JSON encoding. Values are written straight into a byte buffer rather than
// being converted to map[string]any and handed to encoding/json — that round
// trip cost two full copies of the data plus reflection, and it dropped
// numbers on the way: VAL_NUMBER keeps its payload in Value.Num, which the
// generic any conversion never read.
//
// Output stays byte-compatible with encoding/json: object keys are sorted,
// HTML delimiters and the U+2028/U+2029 line separators are escaped, and
// floats use the same format selection.

// maxJSONDepth bounds nesting so a self-referential array or object reports an
// error instead of overflowing the goroutine stack.
const maxJSONDepth = 512

// lineSeparator and paragraphSeparator are valid in JSON strings but terminate
// a JavaScript string literal, so they get escaped.
const (
	lineSeparator      = '\u2028'
	paragraphSeparator = '\u2029'
)

type jsonEncoder struct {
	buf    []byte
	indent string
}

// encodeJSON serializes a builtin argument to JSON. indent is "" for compact
// output, or the per-level indent string for pretty output.
func encodeJSON(v any, indent string) ([]byte, error) {
	e := jsonEncoder{buf: make([]byte, 0, 512), indent: indent}
	if err := e.encodeAny(v, 0); err != nil {
		return nil, err
	}
	return e.buf, nil
}

// encodeValueJSON is the script.Value entry point used by the fast builtin.
func encodeValueJSON(v script.Value, indent string) ([]byte, error) {
	e := jsonEncoder{buf: make([]byte, 0, 512), indent: indent}
	if err := e.encodeValue(v, 0); err != nil {
		return nil, err
	}
	return e.buf, nil
}

// encodeAny handles the Go-native shapes that reach builtins through the
// map[string]any argument path, delegating to encodeValue for script types.
func (e *jsonEncoder) encodeAny(v any, depth int) error {
	if depth > maxJSONDepth {
		return errJSONTooDeep()
	}

	switch val := v.(type) {
	case nil:
		e.buf = append(e.buf, "null"...)
	case script.Value:
		return e.encodeValue(val, depth)
	case *script.ValueRef:
		return e.encodeValue(val.Val, depth)
	case bool:
		e.encodeBool(val)
	case float64:
		return e.encodeFloat(val)
	case int:
		e.buf = strconv.AppendInt(e.buf, int64(val), 10)
	case int64:
		e.buf = strconv.AppendInt(e.buf, val, 10)
	case string:
		e.encodeString(val)
	case *[]script.Value:
		return e.encodeValueArray(*val, depth)
	case []script.Value:
		return e.encodeValueArray(val, depth)
	case map[string]script.Value:
		return e.encodeValueObject(val, depth)
	case []any:
		return e.encodeAnyArray(val, depth)
	case map[string]any:
		return e.encodeAnyObject(val, depth)
	case *script.ScriptFunction:
		e.encodeString("<function>")
	case error:
		e.encodeString(fmt.Sprintf("<error: %v>", val))
	default:
		e.encodeString(fmt.Sprintf("%v", val))
	}
	return nil
}

func (e *jsonEncoder) encodeValue(v script.Value, depth int) error {
	if depth > maxJSONDepth {
		return errJSONTooDeep()
	}

	switch v.Type {
	case script.VAL_NIL:
		e.buf = append(e.buf, "null"...)

	case script.VAL_NUMBER:
		return e.encodeFloat(v.Num)

	case script.VAL_STRING:
		e.encodeString(v.AsString())

	case script.VAL_BOOL:
		e.encodeBool(v.AsBool())

	case script.VAL_ARRAY:
		return e.encodeValueArray(v.AsArray(), depth)

	case script.VAL_OBJECT:
		return e.encodeValueObject(v.AsObject(), depth)

	case script.VAL_FUNCTION:
		e.encodeString("<function>")

	case script.VAL_ERROR:
		if errVal, ok := v.Data.(*script.ErrorValue); ok && errVal.Message.IsString() {
			e.encodeString(fmt.Sprintf("<error: %s>", errVal.Message.AsString()))
		} else {
			e.encodeString("<error>")
		}

	case script.VAL_BINARY:
		e.encodeString(v.String())

	case script.VAL_REGEX:
		if regex, ok := v.Data.(*script.RegexValue); ok {
			e.encodeString(fmt.Sprintf("~%s~", regex.Pattern))
		} else {
			e.encodeString("<regex>")
		}

	case script.VAL_CODE:
		if code, ok := v.Data.(*script.CodeValue); ok {
			e.encodeString(code.Source)
		} else {
			e.buf = append(e.buf, "null"...)
		}

	default:
		e.encodeString(v.String())
	}
	return nil
}

func (e *jsonEncoder) encodeValueArray(arr []script.Value, depth int) error {
	if len(arr) == 0 {
		e.buf = append(e.buf, "[]"...)
		return nil
	}
	e.buf = append(e.buf, '[')
	for i := range arr {
		if i > 0 {
			e.buf = append(e.buf, ',')
		}
		e.writeNewline(depth + 1)
		if err := e.encodeValue(arr[i], depth+1); err != nil {
			return err
		}
	}
	e.writeNewline(depth)
	e.buf = append(e.buf, ']')
	return nil
}

func (e *jsonEncoder) encodeValueObject(obj map[string]script.Value, depth int) error {
	if len(obj) == 0 {
		e.buf = append(e.buf, "{}"...)
		return nil
	}
	var scratch [16]string
	e.buf = append(e.buf, '{')
	for i, k := range sortedKeys(obj, scratch[:0]) {
		if i > 0 {
			e.buf = append(e.buf, ',')
		}
		e.writeNewline(depth + 1)
		e.encodeString(k)
		e.writeKeySeparator()
		if err := e.encodeValue(obj[k], depth+1); err != nil {
			return err
		}
	}
	e.writeNewline(depth)
	e.buf = append(e.buf, '}')
	return nil
}

func (e *jsonEncoder) encodeAnyArray(arr []any, depth int) error {
	if len(arr) == 0 {
		e.buf = append(e.buf, "[]"...)
		return nil
	}
	e.buf = append(e.buf, '[')
	for i, item := range arr {
		if i > 0 {
			e.buf = append(e.buf, ',')
		}
		e.writeNewline(depth + 1)
		if err := e.encodeAny(item, depth+1); err != nil {
			return err
		}
	}
	e.writeNewline(depth)
	e.buf = append(e.buf, ']')
	return nil
}

func (e *jsonEncoder) encodeAnyObject(obj map[string]any, depth int) error {
	if len(obj) == 0 {
		e.buf = append(e.buf, "{}"...)
		return nil
	}
	var scratch [16]string
	e.buf = append(e.buf, '{')
	for i, k := range sortedKeys(obj, scratch[:0]) {
		if i > 0 {
			e.buf = append(e.buf, ',')
		}
		e.writeNewline(depth + 1)
		e.encodeString(k)
		e.writeKeySeparator()
		if err := e.encodeAny(obj[k], depth+1); err != nil {
			return err
		}
	}
	e.writeNewline(depth)
	e.buf = append(e.buf, '}')
	return nil
}

// sortedKeys returns map keys in sorted order so output is deterministic
// (matching encoding/json's treatment of maps). scratch is caller-supplied
// stack space used when the object is small enough, which most are; slices.Sort
// is used over sort.Strings because the latter's interface conversion would
// force that scratch space onto the heap.
func sortedKeys[T any](m map[string]T, scratch []string) []string {
	keys := scratch
	if cap(keys) < len(m) {
		keys = make([]string, 0, len(m))
	}
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func errJSONTooDeep() error {
	return fmt.Errorf("value nested more than %d levels deep (circular reference?)", maxJSONDepth)
}

func (e *jsonEncoder) encodeBool(b bool) {
	if b {
		e.buf = append(e.buf, "true"...)
	} else {
		e.buf = append(e.buf, "false"...)
	}
}

// encodeFloat writes a number using encoding/json's format selection: 'f'
// notation except at very large or very small magnitudes, with the exponent
// trimmed from e-07 to e-7.
func (e *jsonEncoder) encodeFloat(f float64) error {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return fmt.Errorf("unsupported number value: %s", strconv.FormatFloat(f, 'g', -1, 64))
	}

	// Integral values are the common case and skip strconv's float path.
	// Negative zero is excluded: int64(-0.0) is 0, which would drop the sign.
	if f == math.Trunc(f) && math.Abs(f) < 1e15 && !(f == 0 && math.Signbit(f)) {
		e.buf = strconv.AppendInt(e.buf, int64(f), 10)
		return nil
	}

	format := byte('f')
	if abs := math.Abs(f); abs != 0 && (abs < 1e-6 || abs >= 1e21) {
		format = 'e'
	}
	e.buf = strconv.AppendFloat(e.buf, f, format, -1, 64)

	if format == 'e' {
		if n := len(e.buf); n >= 4 && e.buf[n-4] == 'e' && e.buf[n-3] == '-' && e.buf[n-2] == '0' {
			e.buf[n-2] = e.buf[n-1]
			e.buf = e.buf[:n-1]
		}
	}
	return nil
}

const hexDigits = "0123456789abcdef"

// encodeString writes a quoted, escaped JSON string, copying runs that need no
// escaping in bulk.
func (e *jsonEncoder) encodeString(s string) {
	e.buf = append(e.buf, '"')

	start := 0
	for i := 0; i < len(s); {
		if c := s[i]; c < utf8.RuneSelf {
			if jsonSafeChar[c] {
				i++
				continue
			}
			if start < i {
				e.buf = append(e.buf, s[start:i]...)
			}
			switch c {
			case '"':
				e.buf = append(e.buf, '\\', '"')
			case '\\':
				e.buf = append(e.buf, '\\', '\\')
			case '\n':
				e.buf = append(e.buf, '\\', 'n')
			case '\r':
				e.buf = append(e.buf, '\\', 'r')
			case '\t':
				e.buf = append(e.buf, '\\', 't')
			case '\b':
				e.buf = append(e.buf, '\\', 'b')
			case '\f':
				e.buf = append(e.buf, '\\', 'f')
			default:
				// Control characters, plus < > & which are escaped so output is
				// safe to embed in HTML (matching encoding/json's default).
				e.buf = append(e.buf, '\\', 'u', '0', '0', hexDigits[c>>4], hexDigits[c&0xF])
			}
			i++
			start = i
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			if start < i {
				e.buf = append(e.buf, s[start:i]...)
			}
			e.buf = append(e.buf, "\ufffd"...)
			i += size
			start = i
			continue
		}
		if r == lineSeparator || r == paragraphSeparator {
			if start < i {
				e.buf = append(e.buf, s[start:i]...)
			}
			e.buf = append(e.buf, '\\', 'u', '2', '0', '2', hexDigits[r&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}

	if start < len(s) {
		e.buf = append(e.buf, s[start:]...)
	}
	e.buf = append(e.buf, '"')
}

// jsonSafeChar reports which ASCII bytes can be copied into a JSON string
// verbatim.
var jsonSafeChar = func() [utf8.RuneSelf]bool {
	var t [utf8.RuneSelf]bool
	for c := 0x20; c < utf8.RuneSelf; c++ {
		t[c] = true
	}
	t['"'] = false
	t['\\'] = false
	t['<'] = false
	t['>'] = false
	t['&'] = false
	return t
}()

func (e *jsonEncoder) writeNewline(depth int) {
	if e.indent == "" {
		return
	}
	e.buf = append(e.buf, '\n')
	for i := 0; i < depth; i++ {
		e.buf = append(e.buf, e.indent...)
	}
}

func (e *jsonEncoder) writeKeySeparator() {
	e.buf = append(e.buf, ':')
	if e.indent != "" {
		e.buf = append(e.buf, ' ')
	}
}
