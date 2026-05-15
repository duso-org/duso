package markdown

import "strings"

// linkRef is a resolved [label]: dest "title" definition.
type linkRef struct {
	URL   string
	Title string
}

// refMap maps normalized labels to their link refs.
type refMap map[string]linkRef

// tryParseLinkRef attempts to parse a link reference definition starting at
// lines[i]. Returns the number of lines consumed (0 if not a ref), the label,
// and the parsed ref. Multi-line forms (label on one line, dest on next, etc.)
// are supported per spec.
func tryParseLinkRef(lines []string, i int) (int, string, linkRef) {
	if i >= len(lines) {
		return 0, "", linkRef{}
	}
	// Concatenate up to 3 lines into a single buffer for parsing, separated
	// by '\n'. Ref defs cannot span past a blank line.
	var buf strings.Builder
	endLine := i
	for endLine < len(lines) && endLine-i < 8 {
		ln := lines[endLine]
		if endLine > i && strings.TrimSpace(ln) == "" {
			break
		}
		if endLine > i {
			buf.WriteByte('\n')
		}
		buf.WriteString(ln)
		endLine++
	}
	s := buf.String()

	// Strip up to 3 leading spaces.
	p := 0
	for p < len(s) && p < 3 && s[p] == ' ' {
		p++
	}
	if p >= len(s) || s[p] != '[' {
		return 0, "", linkRef{}
	}
	// Find matching ']' that's not escaped.
	labelStart := p + 1
	p++
	depth := 1
	labelEnd := -1
	for p < len(s) {
		c := s[p]
		if c == '\\' && p+1 < len(s) {
			p += 2
			continue
		}
		if c == '[' {
			depth++
		} else if c == ']' {
			depth--
			if depth == 0 {
				labelEnd = p
				break
			}
		} else if c == '\n' {
			// labels may contain newlines but no blank line
		}
		p++
	}
	if labelEnd < 0 {
		return 0, "", linkRef{}
	}
	label := s[labelStart:labelEnd]
	if strings.TrimSpace(label) == "" {
		return 0, "", linkRef{}
	}
	p = labelEnd + 1
	if p >= len(s) || s[p] != ':' {
		return 0, "", linkRef{}
	}
	p++
	// Optional whitespace (incl. one newline).
	p = skipSpacesAndUpToOneNewline(s, p)
	if p >= len(s) {
		return 0, "", linkRef{}
	}
	// Destination.
	destStart := p
	dest, np, ok := parseLinkDestination(s, p)
	if !ok {
		return 0, "", linkRef{}
	}
	_ = destStart
	p = np
	// Optional whitespace + title. Title must be on same line OR following
	// line preceded only by whitespace. If title is invalid we still accept
	// the def without a title — but we must verify the trailing chars to end
	// of line are whitespace only.
	saved := p
	pSpaced := skipSpacesAndUpToOneNewline(s, p)
	title := ""
	hasTitle := false
	if pSpaced > p && pSpaced < len(s) {
		t, np, ok := parseLinkTitle(s, pSpaced)
		if ok {
			// Whatever comes after title must be blank to end of line.
			rest := s[np:]
			nl := strings.IndexByte(rest, '\n')
			tail := rest
			if nl >= 0 {
				tail = rest[:nl]
			}
			if strings.TrimSpace(tail) == "" {
				title = t
				p = np
				if nl >= 0 {
					p += nl + 1
				} else {
					p = len(s)
				}
				hasTitle = true
			}
		}
	}
	if !hasTitle {
		// Validate trailing chars on dest line are whitespace.
		rest := s[saved:]
		nl := strings.IndexByte(rest, '\n')
		tail := rest
		if nl >= 0 {
			tail = rest[:nl]
		}
		if strings.TrimSpace(tail) != "" {
			return 0, "", linkRef{}
		}
		if nl >= 0 {
			p = saved + nl + 1
		} else {
			p = len(s)
		}
	}

	// Count how many original lines we consumed by counting newlines in
	// s[0:p] and adding 1.
	consumedLines := 1
	for k := 0; k < p; k++ {
		if s[k] == '\n' {
			consumedLines++
		}
	}
	return consumedLines, label, linkRef{URL: dest, Title: title}
}

func skipSpacesAndUpToOneNewline(s string, p int) int {
	seenNL := false
	for p < len(s) {
		c := s[p]
		if c == ' ' || c == '\t' {
			p++
			continue
		}
		if c == '\n' && !seenNL {
			seenNL = true
			p++
			continue
		}
		break
	}
	return p
}

// parseLinkDestination parses a link destination starting at s[p]. Two forms:
//
//	<...>            angle-bracket form (no unescaped < > or newlines)
//	bare             bare form (balanced parens, no spaces or controls)
//
// Returns the unescaped destination string (with backslash escapes resolved
// and entity references decoded), the position after it, and ok.
func parseLinkDestination(s string, p int) (string, int, bool) {
	if p >= len(s) {
		return "", p, false
	}
	if s[p] == '<' {
		// Angle bracket form — empty <> is allowed.
		i := p + 1
		var b strings.Builder
		for i < len(s) {
			c := s[i]
			if c == '>' {
				return b.String(), i + 1, true
			}
			if c == '<' || c == '\n' {
				return "", p, false
			}
			if c == '\\' && i+1 < len(s) && asciiPunct(s[i+1]) {
				b.WriteByte(s[i+1])
				i += 2
				continue
			}
			if c == '&' {
				if dec, n := decodeEntity(s[i:]); n > 0 {
					b.WriteString(dec)
					i += n
					continue
				}
			}
			b.WriteByte(c)
			i++
		}
		return "", p, false
	}
	// Bare form.
	i := p
	var b strings.Builder
	depth := 0
	for i < len(s) {
		c := s[i]
		if c <= ' ' || c == 0x7F {
			break
		}
		if c == '\\' && i+1 < len(s) && asciiPunct(s[i+1]) {
			b.WriteByte(s[i+1])
			i += 2
			continue
		}
		if c == '&' {
			if dec, n := decodeEntity(s[i:]); n > 0 {
				b.WriteString(dec)
				i += n
				continue
			}
		}
		if c == '(' {
			depth++
		} else if c == ')' {
			if depth == 0 {
				break
			}
			depth--
		}
		b.WriteByte(c)
		i++
	}
	if depth != 0 {
		return "", p, false
	}
	if b.Len() == 0 {
		// Empty bare destination only legal at p[len(s)-1] would already
		// have returned. Distinguish from "no destination found".
		return "", p, false
	}
	return b.String(), i, true
}

// parseLinkTitle parses a link title starting at s[p]. Forms:
//
//	"..."   '...'   (...)
//
// Returns the unescaped title (with backslash escapes resolved and entity
// references decoded), position after, and ok.
func parseLinkTitle(s string, p int) (string, int, bool) {
	if p >= len(s) {
		return "", p, false
	}
	var closeCh byte
	switch s[p] {
	case '"':
		closeCh = '"'
	case '\'':
		closeCh = '\''
	case '(':
		closeCh = ')'
	default:
		return "", p, false
	}
	i := p + 1
	var b strings.Builder
	for i < len(s) {
		c := s[i]
		if c == closeCh {
			return b.String(), i + 1, true
		}
		if c == '\\' && i+1 < len(s) && asciiPunct(s[i+1]) {
			b.WriteByte(s[i+1])
			i += 2
			continue
		}
		if c == '&' {
			if dec, n := decodeEntity(s[i:]); n > 0 {
				b.WriteString(dec)
				i += n
				continue
			}
		}
		if c == '\n' && i+1 < len(s) && s[i+1] == '\n' {
			return "", p, false
		}
		b.WriteByte(c)
		i++
	}
	return "", p, false
}

// splitLines splits src on '\n', stripping CR before each newline. The final
// trailing newline (if any) does NOT produce an empty trailing element.
func splitLines(src string) []string {
	// Normalize \r\n and \r to \n quickly.
	if strings.ContainsAny(src, "\r") {
		src = strings.ReplaceAll(src, "\r\n", "\n")
		src = strings.ReplaceAll(src, "\r", "\n")
	}
	if src == "" {
		return nil
	}
	lines := strings.Split(src, "\n")
	// strings.Split("a\n", "\n") = ["a", ""] — drop the empty trailer.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
