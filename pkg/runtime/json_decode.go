package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/duso-org/duso/pkg/script"
)

// JSON decoding straight to script.Value. The previous path unmarshaled into
// any (reflection, one map[string]any tree), copied that tree to an identical
// one, then walked it a third time to build Values. This does it in one pass.

type jsonParser struct {
	src []byte
	pos int
	// keys interns object key strings. Documents are usually arrays of
	// uniform records, so the same handful of keys repeats once per element;
	// interning turns thousands of allocations into a handful.
	keys map[string]string
}

// decodeJSON parses a complete JSON document into a script.Value.
func decodeJSON(src []byte) (script.Value, error) {
	p := &jsonParser{src: src}

	p.skipSpace()
	v, err := p.parseValue(0)
	if err != nil {
		return script.NewNil(), err
	}

	p.skipSpace()
	if p.pos < len(p.src) {
		return script.NewNil(), p.errorf("unexpected trailing data")
	}
	return v, nil
}

func (p *jsonParser) parseValue(depth int) (script.Value, error) {
	if depth > maxJSONDepth {
		return script.NewNil(), p.errorf("exceeded maximum nesting depth of %d", maxJSONDepth)
	}
	if p.pos >= len(p.src) {
		return script.NewNil(), p.errorf("unexpected end of input")
	}

	switch c := p.src[p.pos]; c {
	case '{':
		return p.parseObject(depth)
	case '[':
		return p.parseArray(depth)
	case '"':
		s, err := p.parseString()
		if err != nil {
			return script.NewNil(), err
		}
		return script.NewString(s), nil
	case 't':
		if err := p.expectLiteral("true"); err != nil {
			return script.NewNil(), err
		}
		return script.NewBool(true), nil
	case 'f':
		if err := p.expectLiteral("false"); err != nil {
			return script.NewNil(), err
		}
		return script.NewBool(false), nil
	case 'n':
		if err := p.expectLiteral("null"); err != nil {
			return script.NewNil(), err
		}
		return script.NewNil(), nil
	default:
		if c == '-' || (c >= '0' && c <= '9') {
			return p.parseNumber()
		}
		return script.NewNil(), p.errorf("unexpected character %q", c)
	}
}

func (p *jsonParser) parseObject(depth int) (script.Value, error) {
	p.pos++ // consume '{'
	obj := make(map[string]script.Value)

	p.skipSpace()
	if p.pos < len(p.src) && p.src[p.pos] == '}' {
		p.pos++
		return script.NewObject(obj), nil
	}

	for {
		p.skipSpace()
		if p.pos >= len(p.src) || p.src[p.pos] != '"' {
			return script.NewNil(), p.errorf("expected object key")
		}
		key, err := p.parseKey()
		if err != nil {
			return script.NewNil(), err
		}

		p.skipSpace()
		if p.pos >= len(p.src) || p.src[p.pos] != ':' {
			return script.NewNil(), p.errorf("expected ':' after object key")
		}
		p.pos++

		p.skipSpace()
		val, err := p.parseValue(depth + 1)
		if err != nil {
			return script.NewNil(), err
		}
		obj[key] = val

		p.skipSpace()
		if p.pos >= len(p.src) {
			return script.NewNil(), p.errorf("unexpected end of input in object")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case '}':
			p.pos++
			return script.NewObject(obj), nil
		default:
			return script.NewNil(), p.errorf("expected ',' or '}' in object")
		}
	}
}

func (p *jsonParser) parseArray(depth int) (script.Value, error) {
	p.pos++ // consume '['
	arr := []script.Value{}

	p.skipSpace()
	if p.pos < len(p.src) && p.src[p.pos] == ']' {
		p.pos++
		return script.NewArray(arr), nil
	}

	for {
		p.skipSpace()
		val, err := p.parseValue(depth + 1)
		if err != nil {
			return script.NewNil(), err
		}
		arr = append(arr, val)

		p.skipSpace()
		if p.pos >= len(p.src) {
			return script.NewNil(), p.errorf("unexpected end of input in array")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return script.NewArray(arr), nil
		default:
			return script.NewNil(), p.errorf("expected ',' or ']' in array")
		}
	}
}

// parseKey reads an object key, reusing the string for keys already seen in
// this document. Keys with escapes are rare and fall through to parseString.
func (p *jsonParser) parseKey() (string, error) {
	// Scan for a plain ASCII key, which is nearly all of them.
	for i := p.pos + 1; i < len(p.src); i++ {
		c := p.src[i]
		if c == '"' {
			raw := p.src[p.pos+1 : i]
			p.pos = i + 1
			if s, ok := p.keys[string(raw)]; ok {
				return s, nil
			}
			s := string(raw)
			if p.keys == nil {
				p.keys = make(map[string]string, 8)
			}
			p.keys[s] = s
			return s, nil
		}
		if c == '\\' || c < 0x20 || c >= utf8.RuneSelf {
			break
		}
	}
	return p.parseString()
}

// parseString reads a quoted string. Strings without escapes are sliced
// directly out of the source, which is the common case.
func (p *jsonParser) parseString() (string, error) {
	p.pos++ // consume opening quote
	start := p.pos
	nonASCII := false

	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case c == '"':
			s := string(p.src[start:p.pos])
			p.pos++
			if nonASCII && !utf8.ValidString(s) {
				s = coerceValidUTF8(s)
			}
			return s, nil
		case c == '\\':
			return p.parseStringEscaped(start)
		case c < 0x20:
			return "", p.errorf("invalid control character in string")
		case c >= utf8.RuneSelf:
			nonASCII = true
		}
		p.pos++
	}
	return "", p.errorf("unterminated string")
}

// coerceValidUTF8 replaces each invalid byte with U+FFFD. strings.ToValidUTF8
// would collapse a run of bad bytes into one replacement char; encoding/json
// emits one per byte, and matching it keeps decoded strings identical.
func coerceValidUTF8(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			sb.WriteRune(utf8.RuneError)
			i++
			continue
		}
		sb.WriteString(s[i : i+size])
		i += size
	}
	return sb.String()
}

// parseStringEscaped handles the slow path once a backslash is seen. start is
// the offset of the string's first content byte.
func (p *jsonParser) parseStringEscaped(start int) (string, error) {
	var sb strings.Builder
	sb.Grow(len(p.src) - start)
	sb.Write(p.src[start:p.pos])

	for p.pos < len(p.src) {
		c := p.src[p.pos]

		if c == '"' {
			p.pos++
			s := sb.String()
			if !utf8.ValidString(s) {
				s = coerceValidUTF8(s)
			}
			return s, nil
		}
		if c < 0x20 {
			return "", p.errorf("invalid control character in string")
		}
		if c != '\\' {
			// Copy the whole run up to the next escape or quote in one write.
			chunk := p.pos
			for p.pos < len(p.src) && p.src[p.pos] != '\\' && p.src[p.pos] != '"' && p.src[p.pos] >= 0x20 {
				p.pos++
			}
			sb.Write(p.src[chunk:p.pos])
			continue
		}

		p.pos++ // consume backslash
		if p.pos >= len(p.src) {
			return "", p.errorf("unterminated escape sequence")
		}
		switch p.src[p.pos] {
		case '"':
			sb.WriteByte('"')
		case '\\':
			sb.WriteByte('\\')
		case '/':
			sb.WriteByte('/')
		case 'b':
			sb.WriteByte('\b')
		case 'f':
			sb.WriteByte('\f')
		case 'n':
			sb.WriteByte('\n')
		case 'r':
			sb.WriteByte('\r')
		case 't':
			sb.WriteByte('\t')
		case 'u':
			r, err := p.parseUnicodeEscape()
			if err != nil {
				return "", err
			}
			sb.WriteRune(r)
			continue
		default:
			return "", p.errorf("invalid escape character %q", p.src[p.pos])
		}
		p.pos++
	}
	return "", p.errorf("unterminated string")
}

// parseUnicodeEscape reads \uXXXX (p.pos is on the 'u'), joining surrogate
// pairs. It leaves p.pos just past the escape.
func (p *jsonParser) parseUnicodeEscape() (rune, error) {
	r, err := p.readHex4()
	if err != nil {
		return 0, err
	}

	if utf16.IsSurrogate(r) {
		// A trailing surrogate must follow, otherwise emit U+FFFD like
		// encoding/json does.
		if p.pos+6 <= len(p.src) && p.src[p.pos] == '\\' && p.src[p.pos+1] == 'u' {
			save := p.pos
			p.pos++ // consume backslash, leaving pos on 'u'
			r2, err := p.readHex4()
			if err != nil {
				return 0, err
			}
			if combined := utf16.DecodeRune(r, r2); combined != utf8.RuneError {
				return combined, nil
			}
			p.pos = save
		}
		return utf8.RuneError, nil
	}
	return r, nil
}

// readHex4 consumes "uXXXX" starting at p.pos (which is on the 'u').
func (p *jsonParser) readHex4() (rune, error) {
	if p.pos+5 > len(p.src) {
		return 0, p.errorf("incomplete \\u escape")
	}
	var r rune
	for i := 1; i <= 4; i++ {
		c := p.src[p.pos+i]
		var d rune
		switch {
		case c >= '0' && c <= '9':
			d = rune(c - '0')
		case c >= 'a' && c <= 'f':
			d = rune(c-'a') + 10
		case c >= 'A' && c <= 'F':
			d = rune(c-'A') + 10
		default:
			return 0, p.errorf("invalid hex digit %q in \\u escape", c)
		}
		r = r<<4 | d
	}
	p.pos += 5
	return r, nil
}

func (p *jsonParser) parseNumber() (script.Value, error) {
	start := p.pos

	if p.pos < len(p.src) && p.src[p.pos] == '-' {
		p.pos++
	}

	intStart := p.pos
	for p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
		p.pos++
	}
	if p.pos == intStart {
		return script.NewNil(), p.errorf("invalid number")
	}
	// JSON forbids leading zeros: 0 and 0.5 are fine, 01 is not.
	if p.src[intStart] == '0' && p.pos-intStart > 1 {
		return script.NewNil(), p.errorf("invalid number: leading zero")
	}

	isFloat := false
	if p.pos < len(p.src) && p.src[p.pos] == '.' {
		isFloat = true
		p.pos++
		fracStart := p.pos
		for p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
			p.pos++
		}
		if p.pos == fracStart {
			return script.NewNil(), p.errorf("invalid number: no digits after decimal point")
		}
	}
	if p.pos < len(p.src) && (p.src[p.pos] == 'e' || p.src[p.pos] == 'E') {
		isFloat = true
		p.pos++
		if p.pos < len(p.src) && (p.src[p.pos] == '+' || p.src[p.pos] == '-') {
			p.pos++
		}
		expStart := p.pos
		for p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
			p.pos++
		}
		if p.pos == expStart {
			return script.NewNil(), p.errorf("invalid number: no digits in exponent")
		}
	}

	text := p.src[start:p.pos]

	// Integers that fit comfortably in float64 skip strconv's parser.
	if !isFloat && len(text) <= 15 {
		var n int64
		i := 0
		neg := false
		if text[0] == '-' {
			neg = true
			i = 1
		}
		for ; i < len(text); i++ {
			n = n*10 + int64(text[i]-'0')
		}
		if neg {
			if n == 0 {
				// int64 has no negative zero; -0 must stay -0.
				return script.NewNumber(math.Copysign(0, -1)), nil
			}
			n = -n
		}
		return script.NewNumber(float64(n)), nil
	}

	f, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return script.NewNil(), p.errorf("invalid number %q", string(text))
	}
	return script.NewNumber(f), nil
}

func (p *jsonParser) expectLiteral(lit string) error {
	if p.pos+len(lit) > len(p.src) || string(p.src[p.pos:p.pos+len(lit)]) != lit {
		return p.errorf("invalid literal, expected %q", lit)
	}
	p.pos += len(lit)
	return nil
}

func (p *jsonParser) skipSpace() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *jsonParser) errorf(format string, args ...any) error {
	return fmt.Errorf("%s at offset %d", fmt.Sprintf(format, args...), p.pos)
}
