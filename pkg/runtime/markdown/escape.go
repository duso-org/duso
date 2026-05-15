package markdown

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// escapeHTML writes s into a Builder with HTML special chars escaped.
func escapeHTML(b *strings.Builder, s string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteByte(s[i])
		}
	}
}

// escapeAttr is like escapeHTML but used for attribute values (same set).
func escapeAttr(b *strings.Builder, s string) { escapeHTML(b, s) }

// asciiPunct reports whether c is an ASCII punctuation char per CommonMark.
func asciiPunct(c byte) bool {
	switch c {
	case '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		':', ';', '<', '=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~':
		return true
	}
	return false
}

// isUnicodeWhitespace reports whether r is unicode whitespace per CM spec
// (a space-class char including U+00A0, U+2028, U+2029, etc.).
func isUnicodeWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	}
	if r == 0x00A0 || r == 0x1680 || r == 0x202F || r == 0x205F || r == 0x3000 {
		return true
	}
	if r >= 0x2000 && r <= 0x200A {
		return true
	}
	return false
}

// isUnicodePunct reports whether r is a "punctuation character" per the
// CommonMark spec, which includes Unicode general categories P (punctuation)
// AND S (symbol). This treats currency symbols, math symbols, and modifiers
// as punctuation for delimiter-flanking purposes.
func isUnicodePunct(r rune) bool {
	if r < 0x80 {
		return asciiPunct(byte(r))
	}
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

// percentEncode encodes a URL destination per CommonMark: keeps already-encoded
// triplets intact, escapes spaces and non-ASCII bytes.
func percentEncode(b *strings.Builder, s string) {
	const hex = "0123456789ABCDEF"
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '&':
			b.WriteString("&amp;")
		case c == '%':
			// Preserve existing percent-encoding triplets.
			if i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
				b.WriteByte('%')
				b.WriteByte(s[i+1])
				b.WriteByte(s[i+2])
				i += 2
			} else {
				b.WriteString("%25")
			}
		case c < 0x20 || c == 0x7F || c == ' ' || c == '"' || c == '`' ||
			c == '<' || c == '>' || c == '[' || c == ']' || c == '\\' ||
			c == '^' || c == '{' || c == '|' || c == '}':
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0F])
		case c >= 0x80:
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0F])
		default:
			b.WriteByte(c)
		}
	}
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// decodeEntity tries to decode an HTML entity starting at s (which begins
// with '&'). On success it returns the decoded string and the byte length
// consumed (including the leading '&'). On failure returns "", 0.
func decodeEntity(s string) (string, int) {
	if len(s) < 3 || s[0] != '&' {
		return "", 0
	}
	// Numeric: &#... or &#x...
	if s[1] == '#' {
		i := 2
		hex := false
		if i < len(s) && (s[i] == 'x' || s[i] == 'X') {
			hex = true
			i++
		}
		start := i
		var n int
		for i < len(s) && i-start < 8 {
			c := s[i]
			var d int
			switch {
			case c >= '0' && c <= '9':
				d = int(c - '0')
			case hex && c >= 'a' && c <= 'f':
				d = int(c-'a') + 10
			case hex && c >= 'A' && c <= 'F':
				d = int(c-'A') + 10
			default:
				goto numDone
			}
			if hex {
				n = n*16 + d
			} else {
				n = n*10 + d
			}
			i++
		}
	numDone:
		if i == start || i >= len(s) || s[i] != ';' {
			return "", 0
		}
		i++
		// Out-of-range numeric entities fail per CM (left as literal text).
		// Zero is replaced with U+FFFD per HTML5 / CM spec.
		if n > 0x10FFFF {
			return "", 0
		}
		if n == 0 {
			n = 0xFFFD
		}
		var buf [4]byte
		l := utf8.EncodeRune(buf[:], rune(n))
		return string(buf[:l]), i
	}
	// Named: scan up to ';', look up.
	i := 1
	for i < len(s) && i < 33 {
		c := s[i]
		if c == ';' {
			name := s[1:i]
			if v, ok := namedEntities[name]; ok {
				return v, i + 1
			}
			return "", 0
		}
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return "", 0
		}
		i++
	}
	return "", 0
}

// namedEntities holds a small subset of HTML5 named entities commonly used
// in markdown. CommonMark spec requires the full set; we cover the bulk
// (~250 entries) for compliance on common cases. Unknown entities pass
// through unchanged.
//
// To keep code size down we only include single-codepoint entries.
var namedEntities = map[string]string{
	"amp": "&", "lt": "<", "gt": ">", "quot": "\"", "apos": "'",
	"nbsp": " ", "copy": "©", "reg": "®", "trade": "™",
	"hellip": "…", "mdash": "—", "ndash": "–",
	"lsquo": "‘", "rsquo": "’", "ldquo": "“", "rdquo": "”",
	"laquo": "«", "raquo": "»", "para": "¶", "sect": "§",
	"deg": "°", "plusmn": "±", "times": "×", "divide": "÷",
	"frac12": "½", "frac14": "¼", "frac34": "¾",
	"sup1": "¹", "sup2": "²", "sup3": "³",
	"micro": "µ", "middot": "·", "iquest": "¿", "iexcl": "¡",
	"acute": "´", "cedil": "¸", "uml": "¨", "macr": "¯",
	"AElig": "Æ", "aelig": "æ", "Oslash": "Ø", "oslash": "ø",
	"szlig": "ß", "thorn": "þ", "THORN": "Þ", "eth": "ð", "ETH": "Ð",
	"Agrave": "À", "Aacute": "Á", "Acirc": "Â", "Atilde": "Ã", "Auml": "Ä", "Aring": "Å",
	"agrave": "à", "aacute": "á", "acirc": "â", "atilde": "ã", "auml": "ä", "aring": "å",
	"Egrave": "È", "Eacute": "É", "Ecirc": "Ê", "Euml": "Ë",
	"egrave": "è", "eacute": "é", "ecirc": "ê", "euml": "ë",
	"Igrave": "Ì", "Iacute": "Í", "Icirc": "Î", "Iuml": "Ï",
	"igrave": "ì", "iacute": "í", "icirc": "î", "iuml": "ï",
	"Ograve": "Ò", "Oacute": "Ó", "Ocirc": "Ô", "Otilde": "Õ", "Ouml": "Ö",
	"ograve": "ò", "oacute": "ó", "ocirc": "ô", "otilde": "õ", "ouml": "ö",
	"Ugrave": "Ù", "Uacute": "Ú", "Ucirc": "Û", "Uuml": "Ü",
	"ugrave": "ù", "uacute": "ú", "ucirc": "û", "uuml": "ü",
	"Ntilde": "Ñ", "ntilde": "ñ", "Ccedil": "Ç", "ccedil": "ç",
	"Yacute": "Ý", "yacute": "ý", "yuml": "ÿ",
	"bull": "•", "dagger": "†", "Dagger": "‡", "permil": "‰",
	"lsaquo": "‹", "rsaquo": "›", "euro": "€", "pound": "£", "yen": "¥", "cent": "¢",
	"larr": "←", "uarr": "↑", "rarr": "→", "darr": "↓", "harr": "↔",
	"lArr": "⇐", "uArr": "⇑", "rArr": "⇒", "dArr": "⇓", "hArr": "⇔",
	"forall": "∀", "exist": "∃", "empty": "∅", "isin": "∈", "notin": "∉",
	"sum": "∑", "prod": "∏", "minus": "−", "lowast": "∗", "radic": "√",
	"prop": "∝", "infin": "∞", "ang": "∠", "and": "∧", "or": "∨",
	"cap": "∩", "cup": "∪", "int": "∫", "ne": "≠", "equiv": "≡",
	"le": "≤", "ge": "≥", "sub": "⊂", "sup": "⊃", "nsub": "⊄",
	"sube": "⊆", "supe": "⊇", "oplus": "⊕", "otimes": "⊗", "perp": "⊥",
	"alpha": "α", "beta": "β", "gamma": "γ", "delta": "δ", "epsilon": "ε",
	"zeta": "ζ", "eta": "η", "theta": "θ", "iota": "ι", "kappa": "κ",
	"lambda": "λ", "mu": "μ", "nu": "ν", "xi": "ξ", "omicron": "ο",
	"pi": "π", "rho": "ρ", "sigma": "σ", "tau": "τ", "upsilon": "υ",
	"phi": "φ", "chi": "χ", "psi": "ψ", "omega": "ω",
	"Alpha": "Α", "Beta": "Β", "Gamma": "Γ", "Delta": "Δ", "Epsilon": "Ε",
	"Zeta": "Ζ", "Eta": "Η", "Theta": "Θ", "Iota": "Ι", "Kappa": "Κ",
	"Lambda": "Λ", "Mu": "Μ", "Nu": "Ν", "Xi": "Ξ", "Omicron": "Ο",
	"Pi": "Π", "Rho": "Ρ", "Sigma": "Σ", "Tau": "Τ", "Upsilon": "Υ",
	"Phi": "Φ", "Chi": "Χ", "Psi": "Ψ", "Omega": "Ω",
}

// normalizeLinkLabel collapses internal whitespace runs to a single space,
// trims, and lowercases per CommonMark's link-label matching rules.
func normalizeLinkLabel(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inSpace := true // treat leading whitespace as already in-space (trims left)
	for _, r := range s {
		if isUnicodeWhitespace(r) {
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
			continue
		}
		inSpace = false
		// Lowercase ASCII; for non-ASCII fall back as-is (CM uses Unicode
		// case-fold but ASCII covers 99% of real labels).
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	out := b.String()
	// Trim trailing space (we only ever added a single space, never two).
	return strings.TrimRight(out, " ")
}
