package script

import (
	"encoding/binary"
	"fmt"
	"math"
	"regexp"
	"sort"
)

// Binary codec for Duso values. One encoding shared by the datastore WAL, the
// snapshot file, and eventually the replication wire.
// Format spec: docs/ideas/wal-and-codec-plan.md
//
// Tags are allocated in reserved ranges — scalars 0x01-0x0F, containers
// 0x10-0x1F, runtime types 0x20-0x2F, references 0x30-0x3F — so the category of
// an unrecognized tag stays inferable as the format grows.
const (
	tagNil    byte = 0x00
	tagFalse  byte = 0x01
	tagTrue   byte = 0x02
	tagNumber byte = 0x03
	tagInt    byte = 0x04
	tagString byte = 0x05

	tagArray  byte = 0x10
	tagObject byte = 0x11

	tagBinary byte = 0x20
	tagError  byte = 0x21
	tagRegex  byte = 0x22
	tagCode   byte = 0x23

	tagBinaryRef byte = 0x30 // reserved for a future blob store; never emitted
)

// maxExactInt is the largest magnitude a float64 holds exactly. Integral values
// within it encode as a varint instead of 8 fixed bytes, which is the single
// biggest size lever in the format — ids, counts and indexes dominate real data.
const maxExactInt = 1 << 53

// EncodeValue appends the binary encoding of v to buf and returns the extended
// buffer. v is an any-tree of the kind a datastore holds.
//
// Functions cannot be encoded and are elided rather than raising an error,
// matching the deep_copy() builtin exactly: a function becomes nil, and a
// function stored as an object value has its key dropped entirely.
func EncodeValue(buf []byte, v any) ([]byte, error) {
	return encodeAny(buf, v)
}

// DecodeValue decodes one value from buf, returning it with the remaining bytes.
func DecodeValue(buf []byte) (any, []byte, error) {
	return decodeValue(buf)
}

func encodeAny(buf []byte, v any) ([]byte, error) {
	switch t := v.(type) {
	case nil:
		return append(buf, tagNil), nil
	case bool:
		return appendBool(buf, t), nil
	case float64:
		return appendNumber(buf, t), nil
	case string:
		return appendTaggedString(buf, t), nil
	case []any:
		buf = append(buf, tagArray)
		buf = binary.AppendUvarint(buf, uint64(len(t)))
		var err error
		for _, elem := range t {
			if isFunctionValue(elem) {
				// nil rather than dropped: removing it would shift every index
				buf = append(buf, tagNil)
				continue
			}
			if buf, err = encodeAny(buf, elem); err != nil {
				return nil, err
			}
		}
		return buf, nil
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k, elem := range t {
			if isFunctionValue(elem) {
				continue // key dropped entirely, matching deep_copy()
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)

		buf = append(buf, tagObject)
		buf = binary.AppendUvarint(buf, uint64(len(keys)))
		var err error
		for _, k := range keys {
			buf = appendRawString(buf, k)
			if buf, err = encodeAny(buf, t[k]); err != nil {
				return nil, err
			}
		}
		return buf, nil
	case *ValueRef:
		return encodeVal(buf, t.Val)
	case Value:
		return encodeVal(buf, t)
	case *[]Value:
		// Datastore writes normalize these to []any; handled defensively so the
		// codec never depends on that normalization having happened.
		arr := *t
		buf = append(buf, tagArray)
		buf = binary.AppendUvarint(buf, uint64(len(arr)))
		var err error
		for _, elem := range arr {
			if buf, err = encodeVal(buf, elem); err != nil {
				return nil, err
			}
		}
		return buf, nil
	case int:
		return appendNumber(buf, float64(t)), nil
	case int64:
		return appendNumber(buf, float64(t)), nil
	default:
		return nil, fmt.Errorf("cannot encode value of Go type %T", v)
	}
}

func encodeVal(buf []byte, v Value) ([]byte, error) {
	switch v.Type {
	case VAL_NIL:
		return append(buf, tagNil), nil
	case VAL_BOOL:
		return appendBool(buf, v.AsBool()), nil
	case VAL_NUMBER:
		return appendNumber(buf, v.Num), nil
	case VAL_STRING:
		return appendTaggedString(buf, v.AsString()), nil
	case VAL_ARRAY:
		arr := v.AsArray()
		buf = append(buf, tagArray)
		buf = binary.AppendUvarint(buf, uint64(len(arr)))
		var err error
		for _, elem := range arr {
			if elem.Type == VAL_FUNCTION {
				buf = append(buf, tagNil)
				continue
			}
			if buf, err = encodeVal(buf, elem); err != nil {
				return nil, err
			}
		}
		return buf, nil
	case VAL_OBJECT:
		return encodeValObject(buf, v.AsObject())
	case VAL_BINARY:
		bin := v.AsBinary()
		if bin == nil {
			return nil, fmt.Errorf("cannot encode malformed binary value")
		}
		var data []byte
		if bin.Data != nil {
			data = *bin.Data
		}
		buf = append(buf, tagBinary)
		buf = binary.AppendUvarint(buf, uint64(len(data)))
		buf = append(buf, data...)
		return encodeValObject(buf, bin.Metadata)
	case VAL_ERROR:
		ev := v.AsErrorVal()
		if ev == nil {
			return nil, fmt.Errorf("cannot encode malformed error value")
		}
		buf = append(buf, tagError)
		buf, err := encodeVal(buf, ev.Message)
		if err != nil {
			return nil, err
		}
		return appendRawString(buf, ev.Stack), nil
	case VAL_REGEX:
		rv := v.AsRegex()
		if rv == nil {
			return nil, fmt.Errorf("cannot encode malformed regex value")
		}
		buf = append(buf, tagRegex)
		return appendRawString(buf, rv.Pattern), nil
	case VAL_CODE:
		// Only the source and metadata are stored; the AST is rebuilt by
		// re-parsing on decode.
		cv := v.AsCode()
		if cv == nil {
			return nil, fmt.Errorf("cannot encode malformed code value")
		}
		buf = append(buf, tagCode)
		buf = appendRawString(buf, cv.Source)
		return encodeValObject(buf, cv.Metadata)
	case VAL_FUNCTION:
		// Elided, never an error — see EncodeValue. Reached when a function sits
		// somewhere the caller could not skip it, e.g. a bare top-level value.
		return append(buf, tagNil), nil
	default:
		return nil, fmt.Errorf("cannot encode value of type %s", v.Type)
	}
}

func encodeValObject(buf []byte, obj map[string]Value) ([]byte, error) {
	keys := make([]string, 0, len(obj))
	for k, elem := range obj {
		if elem.Type == VAL_FUNCTION {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf = append(buf, tagObject)
	buf = binary.AppendUvarint(buf, uint64(len(keys)))
	var err error
	for _, k := range keys {
		buf = appendRawString(buf, k)
		if buf, err = encodeVal(buf, obj[k]); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func decodeValue(buf []byte) (any, []byte, error) {
	if len(buf) == 0 {
		return nil, nil, fmt.Errorf("truncated value: no tag byte")
	}
	tag := buf[0]
	buf = buf[1:]

	switch tag {
	case tagNil:
		return nil, buf, nil
	case tagFalse:
		return false, buf, nil
	case tagTrue:
		return true, buf, nil
	case tagNumber:
		if len(buf) < 8 {
			return nil, nil, fmt.Errorf("truncated number: want 8 bytes, have %d", len(buf))
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(buf)), buf[8:], nil
	case tagInt:
		u, adv := binary.Uvarint(buf)
		if adv <= 0 {
			return nil, nil, fmt.Errorf("malformed int varint")
		}
		return float64(unzigzag(u)), buf[adv:], nil
	case tagString:
		return readRawString(buf)
	case tagArray:
		n, rest, err := readCount(buf)
		if err != nil {
			return nil, nil, err
		}
		arr := make([]any, n)
		for i := range arr {
			if arr[i], rest, err = decodeValue(rest); err != nil {
				return nil, nil, err
			}
		}
		return arr, rest, nil
	case tagObject:
		n, rest, err := readCount(buf)
		if err != nil {
			return nil, nil, err
		}
		obj := make(map[string]any, n)
		for i := uint64(0); i < n; i++ {
			var k string
			if k, rest, err = readRawString(rest); err != nil {
				return nil, nil, err
			}
			var v any
			if v, rest, err = decodeValue(rest); err != nil {
				return nil, nil, err
			}
			obj[k] = v
		}
		return obj, rest, nil
	case tagBinary:
		n, rest, err := readCount(buf)
		if err != nil {
			return nil, nil, err
		}
		// Copy rather than alias: the caller's buffer may be reused.
		data := make([]byte, n)
		copy(data, rest[:n])
		rest = rest[n:]

		meta, rest, err := readMetadata(rest)
		if err != nil {
			return nil, nil, err
		}
		return &ValueRef{Val: Value{Type: VAL_BINARY, Data: &BinaryValue{Data: &data, Metadata: meta}}}, rest, nil
	case tagError:
		msg, rest, err := decodeValue(buf)
		if err != nil {
			return nil, nil, err
		}
		stack, rest, err := readRawString(rest)
		if err != nil {
			return nil, nil, err
		}
		return &ValueRef{Val: NewErrorValue(InterfaceToValue(msg), stack)}, rest, nil
	case tagRegex:
		pattern, rest, err := readRawString(buf)
		if err != nil {
			return nil, nil, err
		}
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("stored regex %q no longer compiles: %v", pattern, err)
		}
		return &ValueRef{Val: NewRegex(pattern, compiled)}, rest, nil
	case tagCode:
		src, rest, err := readRawString(buf)
		if err != nil {
			return nil, nil, err
		}
		meta, rest, err := readMetadata(rest)
		if err != nil {
			return nil, nil, err
		}
		// Rebuild the AST from source. A failure here means the stored source no
		// longer parses under this build — surface it rather than degrading.
		tokens, err := NewLexer(src).Tokenize()
		if err != nil {
			return nil, nil, fmt.Errorf("stored code no longer tokenizes: %v", err)
		}
		prog, err := NewParser(tokens).Parse()
		if err != nil {
			return nil, nil, fmt.Errorf("stored code no longer parses: %v", err)
		}
		return &ValueRef{Val: NewCode(src, prog, meta)}, rest, nil
	case tagBinaryRef:
		return nil, nil, fmt.Errorf("binary reference values are reserved and not supported by this build")
	default:
		return nil, nil, fmt.Errorf("unknown value tag 0x%02X", tag)
	}
}

// readMetadata decodes an object into the map[string]Value shape that binary and
// code values carry.
func readMetadata(buf []byte) (map[string]Value, []byte, error) {
	v, rest, err := decodeValue(buf)
	if err != nil {
		return nil, nil, err
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("expected object for metadata, got %T", v)
	}
	meta := make(map[string]Value, len(obj))
	for k, item := range obj {
		meta[k] = InterfaceToValue(item)
	}
	return meta, rest, nil
}

// readCount reads a length prefix and rejects any count that cannot fit in the
// remaining input, so corrupt data can't drive a huge allocation.
func readCount(buf []byte) (uint64, []byte, error) {
	n, adv := binary.Uvarint(buf)
	if adv <= 0 {
		return 0, nil, fmt.Errorf("malformed length prefix")
	}
	buf = buf[adv:]
	if n > uint64(len(buf)) {
		return 0, nil, fmt.Errorf("truncated value: length %d exceeds %d remaining bytes", n, len(buf))
	}
	return n, buf, nil
}

func readRawString(buf []byte) (string, []byte, error) {
	n, buf, err := readCount(buf)
	if err != nil {
		return "", nil, err
	}
	return string(buf[:n]), buf[n:], nil
}

func appendBool(buf []byte, b bool) []byte {
	if b {
		return append(buf, tagTrue)
	}
	return append(buf, tagFalse)
}

func appendTaggedString(buf []byte, s string) []byte {
	buf = append(buf, tagString)
	return appendRawString(buf, s)
}

func appendRawString(buf []byte, s string) []byte {
	buf = binary.AppendUvarint(buf, uint64(len(s)))
	return append(buf, s...)
}

// appendNumber picks the compact varint form for integral values and falls back
// to a fixed f64 otherwise. Negative zero takes the f64 path so its sign bit
// survives the round trip.
func appendNumber(buf []byte, f float64) []byte {
	if f == math.Trunc(f) && !math.IsInf(f, 0) && math.Abs(f) <= maxExactInt &&
		!(f == 0 && math.Signbit(f)) {
		buf = append(buf, tagInt)
		return binary.AppendUvarint(buf, zigzag(int64(f)))
	}
	buf = append(buf, tagNumber)
	return binary.LittleEndian.AppendUint64(buf, math.Float64bits(f))
}

func isFunctionValue(v any) bool {
	switch t := v.(type) {
	case *ValueRef:
		return t.Val.Type == VAL_FUNCTION
	case Value:
		return t.Type == VAL_FUNCTION
	}
	return false
}

func zigzag(i int64) uint64 {
	return uint64(i<<1) ^ uint64(i>>63)
}

func unzigzag(u uint64) int64 {
	return int64(u>>1) ^ -int64(u&1)
}
