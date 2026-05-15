package markdown

import (
	"strings"
	"unicode/utf8"
)

// parseInlines tokenizes raw inline text (as it appears in a block) into a
// flat slice of *Inline. Refs is consulted for [label][...] / [label] forms.
func parseInlines(raw string, refs refMap, opts Options) []*Inline {
	p := &inlineParser{src: raw, refs: refs, opts: opts}
	p.parse()
	p.processEmphasis(nil)
	return p.collect(p.head)
}

// inode is the internal doubly-linked list node used during inline parsing.
type inode struct {
	kind  InlineKind
	text  string
	url   string
	title string

	// Sibling links inside the current scope.
	prev *inode
	next *inode

	// First child (used for InlineEmphasis/Strong/Strike/Link/Image).
	first *inode
	last  *inode

	// Delimiter metadata. When isDelim is true the node is a text node
	// holding a run of '*' or '_' (or '~' for strikethrough) that may
	// participate in emphasis matching.
	isDelim   bool
	delimCh   byte
	delimN    int  // remaining count
	origCount int
	canOpen   bool
	canClose  bool

	// Linkage into the delimiter stack (separate from sibling list).
	dprev *inode
	dnext *inode

	// Bracket-open marker for links/images (when delimCh == '[' or '!').
	bracketIsImage bool
	bracketActive  bool
}

type inlineParser struct {
	src  string
	pos  int
	refs refMap
	opts Options

	head *inode
	tail *inode

	// Delimiter stack (doubly linked list, oldest at dhead).
	dhead *inode
	dtail *inode
}

// --- list helpers -----------------------------------------------------------

func (p *inlineParser) append(n *inode) {
	n.prev = p.tail
	n.next = nil
	if p.tail != nil {
		p.tail.next = n
	} else {
		p.head = n
	}
	p.tail = n
}

func (p *inlineParser) pushDelim(n *inode) {
	n.dprev = p.dtail
	n.dnext = nil
	if p.dtail != nil {
		p.dtail.dnext = n
	} else {
		p.dhead = n
	}
	p.dtail = n
}

func (p *inlineParser) removeDelim(n *inode) {
	if n.dprev != nil {
		n.dprev.dnext = n.dnext
	} else {
		p.dhead = n.dnext
	}
	if n.dnext != nil {
		n.dnext.dprev = n.dprev
	} else {
		p.dtail = n.dprev
	}
	n.dprev, n.dnext = nil, nil
}

func (p *inlineParser) removeNode(n *inode) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		p.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		p.tail = n.prev
	}
	n.prev, n.next = nil, nil
}

// collect walks a sibling list and returns a flat []*Inline, with InlineText
// nodes merged where adjacent.
func (p *inlineParser) collect(start *inode) []*Inline {
	var out []*Inline
	for n := start; n != nil; n = n.next {
		if n.isDelim && n.delimN > 0 {
			// Unmatched delimiter — output as literal text.
			lit := strings.Repeat(string(n.delimCh), n.delimN)
			if len(out) > 0 && out[len(out)-1].Kind == InlineText {
				out[len(out)-1].Text += lit
			} else {
				out = append(out, &Inline{Kind: InlineText, Text: lit})
			}
			continue
		}
		if n.kind == InlineText {
			if len(out) > 0 && out[len(out)-1].Kind == InlineText {
				out[len(out)-1].Text += n.text
				continue
			}
		}
		o := &Inline{Kind: n.kind, Text: n.text, URL: n.url, Title: n.title}
		if n.first != nil {
			o.Children = p.collect(n.first)
		}
		out = append(out, o)
	}
	return out
}

// --- main scan --------------------------------------------------------------

func (p *inlineParser) parse() {
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch c {
		case '\\':
			p.handleBackslash()
		case '`':
			p.handleBacktick()
		case '*', '_':
			p.handleEmphasisDelim()
		case '~':
			if p.opts.Strikethrough {
				p.handleStrikeDelim()
			} else {
				p.handleText(1)
			}
		case '[':
			p.handleOpenBracket()
		case ']':
			p.handleCloseBracket()
		case '!':
			if p.pos+1 < len(p.src) && p.src[p.pos+1] == '[' {
				p.handleImageOpen()
			} else {
				p.handleText(1)
			}
		case '<':
			p.handleAngleBracket()
		case '&':
			p.handleEntity()
		case '\n':
			p.handleNewline()
		default:
			p.handleText(1)
		}
	}
}

func (p *inlineParser) handleText(n int) {
	end := p.pos + n
	if end > len(p.src) {
		end = len(p.src)
	}
	chunk := p.src[p.pos:end]
	p.pos = end
	p.appendText(chunk)
}

func (p *inlineParser) appendText(s string) {
	if s == "" {
		return
	}
	if p.tail != nil && p.tail.kind == InlineText && !p.tail.isDelim {
		p.tail.text += s
		return
	}
	p.append(&inode{kind: InlineText, text: s})
}

func (p *inlineParser) handleBackslash() {
	if p.pos+1 < len(p.src) {
		nxt := p.src[p.pos+1]
		if nxt == '\n' {
			// Hard line break.
			p.append(&inode{kind: InlineHardBreak})
			p.pos += 2
			return
		}
		if asciiPunct(nxt) {
			p.appendText(string(nxt))
			p.pos += 2
			return
		}
	}
	p.appendText("\\")
	p.pos++
}

func (p *inlineParser) handleNewline() {
	// Soft break by default. If the preceding text ends with 2+ spaces, it's
	// a hard break.
	hard := false
	if p.tail != nil && p.tail.kind == InlineText {
		t := p.tail.text
		spaces := 0
		for i := len(t) - 1; i >= 0 && t[i] == ' '; i-- {
			spaces++
		}
		if spaces >= 2 {
			hard = true
			p.tail.text = strings.TrimRight(t, " ")
			if p.tail.text == "" {
				p.removeNode(p.tail)
			}
		} else if spaces > 0 {
			// Trim trailing spaces before a soft break.
			p.tail.text = strings.TrimRight(t, " ")
			if p.tail.text == "" {
				p.removeNode(p.tail)
			}
		}
	}
	if hard {
		p.append(&inode{kind: InlineHardBreak})
	} else {
		p.append(&inode{kind: InlineSoftBreak})
	}
	p.pos++
	// Skip leading spaces on the next line.
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

// --- code spans -------------------------------------------------------------

func (p *inlineParser) handleBacktick() {
	// Count opening backtick run.
	start := p.pos
	n := 0
	for p.pos < len(p.src) && p.src[p.pos] == '`' {
		n++
		p.pos++
	}
	// Search for a matching run of exactly n backticks.
	bodyStart := p.pos
	for p.pos < len(p.src) {
		if p.src[p.pos] != '`' {
			p.pos++
			continue
		}
		m := 0
		runStart := p.pos
		for p.pos < len(p.src) && p.src[p.pos] == '`' {
			m++
			p.pos++
		}
		if m == n {
			body := p.src[bodyStart:runStart]
			// Normalize newlines to spaces FIRST, then strip a single leading
			// + trailing space (spec rule). This order matters when the body
			// begins/ends with a newline.
			body = strings.ReplaceAll(body, "\n", " ")
			if len(body) >= 2 && body[0] == ' ' && body[len(body)-1] == ' ' &&
				strings.TrimSpace(body) != "" {
				body = body[1 : len(body)-1]
			}
			p.append(&inode{kind: InlineCode, text: body})
			return
		}
	}
	// No match — backticks become literal text.
	p.pos = start
	p.appendText(p.src[p.pos : p.pos+n])
	p.pos += n
}

// --- emphasis ---------------------------------------------------------------

func (p *inlineParser) handleEmphasisDelim() {
	ch := p.src[p.pos]
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] == ch {
		p.pos++
	}
	n := p.pos - start

	prev := byte(' ')
	if start > 0 {
		prev = p.src[start-1]
	}
	next := byte(' ')
	if p.pos < len(p.src) {
		next = p.src[p.pos]
	}

	canOpen, canClose := classifyDelim(ch, prev, next)

	node := &inode{
		kind:      InlineText,
		text:      p.src[start:p.pos],
		isDelim:   true,
		delimCh:   ch,
		delimN:    n,
		origCount: n,
		canOpen:   canOpen,
		canClose:  canClose,
	}
	p.append(node)
	p.pushDelim(node)
}

func (p *inlineParser) handleStrikeDelim() {
	ch := byte('~')
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] == ch {
		p.pos++
	}
	n := p.pos - start
	if n < 1 || n > 2 {
		// Treat as literal text.
		p.appendText(p.src[start:p.pos])
		return
	}
	prev := byte(' ')
	if start > 0 {
		prev = p.src[start-1]
	}
	next := byte(' ')
	if p.pos < len(p.src) {
		next = p.src[p.pos]
	}
	canOpen, canClose := classifyDelim(ch, prev, next)
	node := &inode{
		kind:      InlineText,
		text:      p.src[start:p.pos],
		isDelim:   true,
		delimCh:   ch,
		delimN:    n,
		origCount: n,
		canOpen:   canOpen,
		canClose:  canClose,
	}
	p.append(node)
	p.pushDelim(node)
}

// classifyDelim implements CM's left-flanking / right-flanking rules.
func classifyDelim(ch, prev, next byte) (bool, bool) {
	prevR, _ := utf8.DecodeLastRuneInString(string(prev))
	nextR, _ := utf8.DecodeRuneInString(string(next))
	prevWS := isUnicodeWhitespace(prevR) || prev == 0
	nextWS := isUnicodeWhitespace(nextR) || next == 0
	prevPunct := isUnicodePunct(prevR)
	nextPunct := isUnicodePunct(nextR)

	leftFlanking := !nextWS && (!nextPunct || prevWS || prevPunct)
	rightFlanking := !prevWS && (!prevPunct || nextWS || nextPunct)

	switch ch {
	case '*', '~':
		return leftFlanking, rightFlanking
	case '_':
		// `_` cannot open inside a word, cannot close inside a word.
		canOpen := leftFlanking && (!rightFlanking || prevPunct)
		canClose := rightFlanking && (!leftFlanking || nextPunct)
		return canOpen, canClose
	}
	return false, false
}

// processEmphasis pairs delimiter runs into emphasis/strong/strike nodes,
// stopping at stack_bottom (or the start if nil).
func (p *inlineParser) processEmphasis(stackBottom *inode) {
	// openers_bottom by char.
	type key struct {
		ch  byte
		mod int
	}
	openersBottom := map[key]*inode{}
	startNode := p.dhead
	if stackBottom != nil {
		startNode = stackBottom.dnext
	}
	closer := startNode
	for closer != nil {
		if !closer.canClose || !closer.isDelim {
			closer = closer.dnext
			continue
		}
		// For strikethrough only count == 2 is valid.
		ch := closer.delimCh
		if ch == '~' && closer.origCount != 2 {
			closer = closer.dnext
			continue
		}
		// Find a matching opener walking backwards.
		k := key{ch, closer.origCount % 3}
		bottom := openersBottom[k]
		if bottom == nil {
			bottom = stackBottom
		}
		opener := closer.dprev
		var match *inode
		for opener != nil && opener != bottom {
			if opener.isDelim && opener.canOpen && opener.delimCh == ch {
				// Rule of three.
				oddMatch := false
				if (opener.canClose || closer.canOpen) &&
					(opener.origCount+closer.origCount)%3 == 0 &&
					!(opener.origCount%3 == 0 && closer.origCount%3 == 0) {
					oddMatch = true
				}
				if !oddMatch {
					match = opener
					break
				}
			}
			opener = opener.dprev
		}
		if match == nil {
			openersBottom[k] = closer.dprev
			if !closer.canOpen {
				next := closer.dnext
				p.removeDelim(closer)
				closer = next
				continue
			}
			closer = closer.dnext
			continue
		}
		// Pair! Determine emphasis vs strong.
		strong := false
		if ch != '~' && match.delimN >= 2 && closer.delimN >= 2 {
			strong = true
		}
		use := 1
		if strong {
			use = 2
		}
		if ch == '~' {
			use = 2
		}
		// Build node wrapping content between match and closer.
		emph := &inode{kind: InlineEmphasis}
		if strong {
			emph.kind = InlineStrong
		}
		if ch == '~' {
			emph.kind = InlineStrike
		}
		// Move siblings (match.next .. closer.prev) into emph.first..last.
		first := match.next
		last := closer.prev
		// Detach from sibling chain.
		if first != nil && last != nil {
			match.next = closer
			closer.prev = match
			first.prev = nil
			last.next = nil
			emph.first = first
			emph.last = last
		}
		// Remove delim entries strictly between match and closer.
		d := match.dnext
		for d != nil && d != closer {
			nxt := d.dnext
			p.removeDelim(d)
			d = nxt
		}
		// Trim used chars from match/closer text.
		match.delimN -= use
		closer.delimN -= use
		if match.delimN > 0 {
			match.text = strings.Repeat(string(match.delimCh), match.delimN)
		} else {
			match.text = ""
		}
		if closer.delimN > 0 {
			closer.text = strings.Repeat(string(closer.delimCh), closer.delimN)
		} else {
			closer.text = ""
		}
		// Insert emph right after match (before closer).
		emph.prev = match
		emph.next = closer
		match.next = emph
		closer.prev = emph
		// Remove drained delims from node + delim list.
		if match.delimN == 0 {
			p.removeDelim(match)
			p.removeNode(match)
		}
		if closer.delimN == 0 {
			next := closer.dnext
			p.removeDelim(closer)
			p.removeNode(closer)
			closer = next
			continue
		}
	}
	// Remove all unmatched delimiters in [startNode..end) from the delimiter
	// stack so they don't interfere with later matching. Re-read the live
	// list start — `startNode` captured at entry may be a detached node by
	// the time we reach here. The corresponding inline nodes stay in place
	// and render as literal text via collect().
	var d *inode
	if stackBottom != nil {
		d = stackBottom.dnext
	} else {
		d = p.dhead
	}
	for d != nil {
		nxt := d.dnext
		p.removeDelim(d)
		d = nxt
	}
}

// --- links & images ---------------------------------------------------------

func (p *inlineParser) handleOpenBracket() {
	node := &inode{kind: InlineText, text: "[", isDelim: true, delimCh: '[', delimN: 1, bracketActive: true}
	p.append(node)
	p.pushDelim(node)
	p.pos++
}

func (p *inlineParser) handleImageOpen() {
	node := &inode{kind: InlineText, text: "![", isDelim: true, delimCh: '[', delimN: 1, bracketIsImage: true, bracketActive: true}
	p.append(node)
	p.pushDelim(node)
	p.pos += 2
}

func (p *inlineParser) handleCloseBracket() {
	p.pos++
	// Walk back through delim stack to find the closest active bracket
	// opener. Inactive openers (left over from failed match attempts) are
	// skipped, allowing `[[...]](url)` to find the outer pair.
	var opener *inode
	for d := p.dtail; d != nil; d = d.dprev {
		if d.isDelim && d.delimCh == '[' && d.bracketActive {
			opener = d
			break
		}
	}
	if opener == nil {
		p.appendText("]")
		return
	}
	// Try inline form: ](url "title")  — both url and title are optional.
	if p.pos < len(p.src) && p.src[p.pos] == '(' {
		save := p.pos
		p.pos++
		p.skipInlineWS()
		dest := ""
		if p.pos < len(p.src) && p.src[p.pos] != ')' {
			d, np, ok := parseLinkDestination(p.src, p.pos)
			if !ok {
				p.pos = save
				goto noInline
			}
			dest = d
			p.pos = np
		}
		p.skipInlineWS()
		title := ""
		if p.pos < len(p.src) && (p.src[p.pos] == '"' || p.src[p.pos] == '\'' || p.src[p.pos] == '(') {
			t, np2, tok := parseLinkTitle(p.src, p.pos)
			if tok {
				title = t
				p.pos = np2
				p.skipInlineWS()
			}
		}
		if p.pos < len(p.src) && p.src[p.pos] == ')' {
			p.pos++
			p.formLink(opener, dest, title)
			return
		}
		p.pos = save
	}
noInline:
	// Try full reference form: ][label]
	if p.pos < len(p.src) && p.src[p.pos] == '[' {
		save := p.pos
		p.pos++
		labelEnd := p.findRefLabelEnd()
		if labelEnd > 0 {
			label := p.src[save+1 : labelEnd]
			normalized := normalizeLinkLabel(label)
			if strings.TrimSpace(normalized) == "" {
				// Collapsed form: empty label means use the text content.
				normalized = normalizeLinkLabel(p.textBetween(opener))
			}
			if ref, ok := p.refs[normalized]; ok {
				p.pos = labelEnd + 1
				p.formLink(opener, ref.URL, ref.Title)
				return
			}
		}
		// Full/collapsed form didn't match — fall through to shortcut.
		p.pos = save
	}
	// Shortcut form: label is the text content of the bracket pair.
	normalized := normalizeLinkLabel(p.textBetween(opener))
	if ref, ok := p.refs[normalized]; ok {
		p.formLink(opener, ref.URL, ref.Title)
		return
	}
	// No match — deactivate opener and emit literal ].
	opener.bracketActive = false
	p.appendText("]")
}

// formLink converts the opener bracket (and everything up to current pos) into
// a link or image inline node.
func (p *inlineParser) formLink(opener *inode, url, title string) {
	isImage := opener.bracketIsImage
	link := &inode{kind: InlineLink, url: url, title: title}
	if isImage {
		link.kind = InlineImage
	}
	// Run process_emphasis on the bracket contents before extracting them,
	// so emphasis inside the link text is resolved.
	p.processEmphasis(opener)

	// Detach siblings opener.next..tail into link.
	first := opener.next
	last := p.tail
	if first != nil {
		first.prev = nil
		last.next = nil
		link.first = first
		link.last = last
	}
	p.tail = opener
	opener.next = nil
	p.removeNode(opener)
	// Remove opener from delim list.
	p.removeDelim(opener)
	p.append(link)

	// Deactivate any remaining brackets in the surrounding scope so we don't
	// allow nested links (per spec).
	if !isImage {
		for d := p.dtail; d != nil; d = d.dprev {
			if d.isDelim && d.delimCh == '[' {
				d.bracketActive = false
			}
		}
	}
}

// textBetween returns the literal text content from opener.next to tail.
// Used for shortcut/collapsed reference labels.
func (p *inlineParser) textBetween(opener *inode) string {
	var b strings.Builder
	for n := opener.next; n != nil; n = n.next {
		if n.isDelim {
			// For delim runs, include their text chars too.
			b.WriteString(strings.Repeat(string(n.delimCh), n.delimN))
			continue
		}
		if n.kind == InlineText {
			b.WriteString(n.text)
			continue
		}
		if n.kind == InlineCode {
			b.WriteString("`")
			b.WriteString(n.text)
			b.WriteString("`")
		}
		// Other inlines: contribute nothing (or could recurse; not needed).
	}
	return b.String()
}

// findRefLabelEnd finds the closing ']' for a reference label starting at p.pos
// (which is just past the opening '['). Returns absolute index of the ']', or -1.
func (p *inlineParser) findRefLabelEnd() int {
	i := p.pos
	for i < len(p.src) && i-p.pos < 1000 {
		c := p.src[i]
		if c == '\\' && i+1 < len(p.src) {
			i += 2
			continue
		}
		if c == ']' {
			return i
		}
		if c == '[' {
			return -1
		}
		i++
	}
	return -1
}

func (p *inlineParser) skipInlineWS() {
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == ' ' || c == '\t' || c == '\n' {
			p.pos++
			continue
		}
		break
	}
}

// --- autolinks & raw HTML ---------------------------------------------------

func (p *inlineParser) handleAngleBracket() {
	// Try absolute autolink: <scheme:...>
	if uri, n := matchAutolink(p.src[p.pos:]); n > 0 {
		p.append(&inode{kind: InlineAutolink, url: uri, text: uri})
		p.pos += n
		return
	}
	// Try email autolink.
	if email, n := matchEmailAutolink(p.src[p.pos:]); n > 0 {
		p.append(&inode{kind: InlineAutolink, url: "mailto:" + email, text: email})
		p.pos += n
		return
	}
	// Try raw HTML tag.
	if raw, n := matchRawHTML(p.src[p.pos:]); n > 0 {
		p.append(&inode{kind: InlineRawHTML, text: raw})
		p.pos += n
		return
	}
	p.appendText("<")
	p.pos++
}

func matchAutolink(s string) (string, int) {
	if len(s) < 4 || s[0] != '<' {
		return "", 0
	}
	// Scheme per CM: 2-32 chars total, starts with an ASCII letter, then
	// ASCII letter/digit/+./- chars.
	i := 1
	if !isASCIILetter(s[i]) {
		return "", 0
	}
	i++
	for i < len(s) && i-1 < 32 {
		c := s[i]
		if isASCIILetter(c) || isDigit(c) || c == '+' || c == '.' || c == '-' {
			i++
			continue
		}
		break
	}
	// Scheme must be at least 2 chars total (letter + ≥1 more) before `:`.
	if i-1 < 2 || i >= len(s) || s[i] != ':' {
		return "", 0
	}
	i++
	urlStart := 1
	for i < len(s) {
		c := s[i]
		if c == '>' {
			return s[urlStart:i], i + 1
		}
		if c == ' ' || c == '\t' || c == '\n' || c == '<' {
			return "", 0
		}
		i++
	}
	return "", 0
}

func matchEmailAutolink(s string) (string, int) {
	if len(s) < 5 || s[0] != '<' {
		return "", 0
	}
	// Simple email pattern.
	i := 1
	hasAt := false
	for i < len(s) {
		c := s[i]
		if c == '>' {
			if !hasAt || i == 1 {
				return "", 0
			}
			return s[1:i], i + 1
		}
		if c == '@' {
			if hasAt {
				return "", 0
			}
			hasAt = true
		} else if !(isASCIILetter(c) || isDigit(c) || c == '.' || c == '_' || c == '-' || c == '+') {
			return "", 0
		}
		i++
	}
	return "", 0
}

func matchRawHTML(s string) (string, int) {
	// Match <tag ...>, </tag>, <!-- comment -->, <?...?>, <!DECL>, <![CDATA[...]]>.
	if len(s) < 3 || s[0] != '<' {
		return "", 0
	}
	if strings.HasPrefix(s, "<!--") {
		end := strings.Index(s[4:], "-->")
		if end < 0 {
			return "", 0
		}
		return s[:4+end+3], 4 + end + 3
	}
	if strings.HasPrefix(s, "<?") {
		end := strings.Index(s[2:], "?>")
		if end < 0 {
			return "", 0
		}
		return s[:2+end+2], 2 + end + 2
	}
	if strings.HasPrefix(s, "<![CDATA[") {
		end := strings.Index(s[9:], "]]>")
		if end < 0 {
			return "", 0
		}
		return s[:9+end+3], 9 + end + 3
	}
	if len(s) > 2 && s[1] == '!' && isASCIILetter(s[2]) {
		end := strings.IndexByte(s, '>')
		if end < 0 {
			return "", 0
		}
		return s[:end+1], end + 1
	}
	// Open or close tag.
	start := 1
	if s[1] == '/' {
		start = 2
	}
	tag, w := readTagName(s[start:])
	if tag == "" {
		return "", 0
	}
	i := start + w
	// Attributes (open tag only). Each attribute must be preceded by ≥1
	// whitespace (separating it from the tag name or the previous attribute).
	if s[1] != '/' {
		for i < len(s) && s[i] != '>' && s[i] != '/' {
			if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' {
				return "", 0
			}
			for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n') {
				i++
			}
			if i >= len(s) || s[i] == '>' || s[i] == '/' {
				break
			}
			c := s[i]
			if !(isASCIILetter(c) || c == '_' || c == ':') {
				return "", 0
			}
			i++
			for i < len(s) {
				c = s[i]
				if isASCIILetter(c) || isDigit(c) || c == '_' || c == ':' || c == '.' || c == '-' {
					i++
					continue
				}
				break
			}
			// Optional value.
			j := i
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n') {
				j++
			}
			if j < len(s) && s[j] == '=' {
				j++
				for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n') {
					j++
				}
				if j >= len(s) {
					return "", 0
				}
				q := s[j]
				if q == '"' || q == '\'' {
					end := strings.IndexByte(s[j+1:], q)
					if end < 0 {
						return "", 0
					}
					i = j + 1 + end + 1
				} else {
					for j < len(s) {
						c := s[j]
						if c == ' ' || c == '\t' || c == '\n' || c == '>' || c == '<' ||
							c == '"' || c == '\'' || c == '=' || c == '`' {
							break
						}
						j++
					}
					if j == i {
						return "", 0
					}
					i = j
				}
			}
		}
	} else {
		// Skip whitespace before '>'.
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n') {
			i++
		}
	}
	if i < len(s) && s[i] == '/' {
		i++
	}
	if i >= len(s) || s[i] != '>' {
		return "", 0
	}
	return s[:i+1], i + 1
}

func (p *inlineParser) handleEntity() {
	if dec, n := decodeEntity(p.src[p.pos:]); n > 0 {
		p.appendText(dec)
		p.pos += n
		return
	}
	p.appendText("&")
	p.pos++
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
