package markdown

import (
	"strings"
)

// parseBlocksWithRefs runs the line-oriented block parser and returns the
// document tree, also collecting link reference definitions found at the
// start of paragraphs into refs.
func parseBlocksWithRefs(src string, opts Options, refs refMap) *Block {
	p := &blockParser{
		lines:             splitLines(src),
		opts:              opts,
		doc:               &Block{Kind: BlockDocument},
		pendingBlankBQIdx: -1,
	}
	p.stack = []*Block{p.doc}
	for p.line = 0; p.line < len(p.lines); p.line++ {
		p.processLine(p.lines[p.line])
	}
	p.closeAll()
	finalizeParagraphs(p.doc, refs)
	convertTables(p.doc, opts)
	markTightness(p.doc)
	return p.doc
}

// blockParser holds parser state.
type blockParser struct {
	lines []string
	opts  Options
	line  int

	doc   *Block
	stack []*Block

	openFence   *Block
	fenceChar   byte
	fenceLen    int
	fenceIndent int

	openHTML *Block
	htmlKind uint8

	// pendingCodeBlanks holds blank lines (stripped of 4-space indent) that
	// were seen after an open indented code block. They get flushed into the
	// code block if more indented code follows, or discarded if it doesn't.
	pendingCodeBlanks []string

	// pendingBlank records that the previous non-blank line was preceded by
	// one or more blank lines. The next non-blank line's container-continuation
	// outcome decides which list items the blank "belongs to" for loose detection.
	pendingBlank bool
	// pendingBlankBQIdx is the stack index of the innermost blockquote that
	// was open when the blank was seen (-1 if none). Used to suppress the
	// blank from affecting items outside that blockquote.
	pendingBlankBQIdx int
}

// processLine processes a single source line. Splits into:
//
//	continueContainers — walk open container stack, peel off their prefixes
//	processContent     — handle the remainder as block starts / leaf content
func (p *blockParser) processLine(rawLine string) {
	// Expand leading tabs to spaces so all subsequent container strips and
	// indent measurements work purely in column terms. Tabs in content
	// (after the first non-whitespace) are preserved.
	rawLine = expandLeadingTabs(rawLine)
	line, allContinued, consumedDepth := p.continueContainers(rawLine)
	if !isBlankLine(line) {
		p.applyPendingBlank(consumedDepth, allContinued)
	}

	// Fenced code block continuation. Containers are stripped first so that
	// list-item indent is removed from the line before the fence's own
	// indent stripping is applied.
	if p.openFence != nil {
		if !allContinued {
			// Container chain broke; close the fence and fall through.
			p.openFence = nil
		} else {
			stripped, _ := stripIndent(line, p.fenceIndent)
			trimmed := strings.TrimLeft(stripped, " ")
			leading := len(stripped) - len(trimmed)
			if leading <= 3 && len(trimmed) >= p.fenceLen && allByte(trimmed[:p.fenceLen], p.fenceChar) {
				rest := strings.TrimRight(trimmed[p.fenceLen:], " \t")
				if rest == "" || allByte(rest, p.fenceChar) {
					p.openFence = nil
					return
				}
			}
			p.openFence.Text += stripped + "\n"
			return
		}
	}
	// HTML block continuation. Types 6 and 7 close on a blank line (that
	// blank line is NOT part of the block). Types 1-5 close on an in-line
	// marker (</script>, -->, ?>, etc.) — that line IS part of the block.
	// We continue on rawLine so existing leading whitespace is preserved
	// for type-1..5 content; for types 6,7 inside list items we still want
	// to keep indent intact (CM treats these as raw HTML).
	if p.openHTML != nil {
		if p.htmlBlockEnds(rawLine, p.htmlKind) {
			if p.htmlKind != 6 && p.htmlKind != 7 {
				p.openHTML.Text += rawLine + "\n"
			}
			p.openHTML = nil
			return
		}
		p.openHTML.Text += rawLine + "\n"
		return
	}

	tip := p.currentTip()
	if !allContinued && tip != nil && tip.Kind == BlockParagraph {
		if !isBlankLine(line) && !p.startsNewBlock(line) {
			tip.Text += "\n" + strings.TrimLeft(line, " \t")
			return
		}
	}

	if len(p.stack) > consumedDepth {
		p.closeTo(consumedDepth)
	}
	p.processContent(line, allContinued)
}

// continueContainers walks open containers and peels off their continuation
// prefixes. Returns the consumed line, whether all containers continued, and
// how many stack entries are still valid. An open leaf at the tip is counted
// as "continued" here — its own continuation/closure is decided in
// processContent (where paragraph continuation, setext promotion, blank
// closure, etc. live).
func (p *blockParser) continueContainers(rawLine string) (string, bool, int) {
	line := rawLine
	consumed := 1
	for i := 1; i < len(p.stack); i++ {
		c := p.stack[i]
		switch c.Kind {
		case BlockParagraph, BlockHeading, BlockCodeBlock, BlockHTMLBlock, BlockThematicBreak, BlockTable:
			consumed++ // leaf counted; leaf logic decides closure later
			return line, true, consumed
		}
		rest, ok := p.tryContinueContainer(c, line)
		if !ok {
			return line, false, consumed
		}
		line = rest
		consumed++
	}
	return line, true, consumed
}

// processContent handles a line after container continuation has been done.
// It tries to open new blocks or continues an existing leaf.
func (p *blockParser) processContent(line string, allContinued bool) {
	if isBlankLine(line) {
		// If an indented code block is open, buffer this blank (with its
		// over-indent stripped) — it belongs to the code block iff more
		// indented code follows.
		if tip := p.currentTip(); tip != nil && tip.Kind == BlockCodeBlock && !tip.Fenced {
			stripped, _ := stripIndent(line, 4)
			p.pendingCodeBlanks = append(p.pendingCodeBlanks, stripped)
			return
		}
		if p.currentTip() != nil {
			p.closeTip()
		}
		p.markBlankLine()
		return
	}

	indent, content := splitIndent(line)
	if indent >= 4 {
		// 4-space indent can't interrupt an open paragraph — treat the line
		// as a paragraph continuation. Otherwise it's an indented code block.
		if t := p.currentTip(); t != nil && t.Kind == BlockParagraph {
			p.handleParagraphLine(line)
			return
		}
		p.handleIndentedCode(line)
		return
	}

	switch {
	case isThematicBreak(content):
		if t := p.currentTip(); t != nil && t.Kind == BlockParagraph {
			if isSetextUnderline(content) == 2 {
				t.Kind = BlockHeading
				t.Level = 2
				return
			}
		}
		p.closeTip()
		p.appendToTop(&Block{Kind: BlockThematicBreak})
	case isATXHeading(content):
		p.closeTip()
		level, text := parseATXHeading(content)
		p.appendToTop(&Block{Kind: BlockHeading, Level: level, Text: text})
	case isSetextUnderline(content) > 0 && p.currentTip() != nil && p.currentTip().Kind == BlockParagraph && allContinued:
		t := p.currentTip()
		t.Kind = BlockHeading
		t.Level = isSetextUnderline(content)
		return
	case isFenceLine(content):
		p.closeTip()
		ch, fenceLen, info := parseFenceOpen(content)
		b := &Block{Kind: BlockCodeBlock, Fenced: true, Lang: info}
		p.appendToTop(b)
		p.openFence = b
		p.fenceChar = ch
		p.fenceLen = fenceLen
		p.fenceIndent = indent
	case isBlockquoteStart(content):
		p.closeTip()
		rest := blockquoteRemainder(content)
		bq := &Block{Kind: BlockBlockquote}
		p.appendToTop(bq)
		// processContent the remainder relative to the new blockquote.
		p.processContent(rest, true)
	case isListItemStart(content) >= 0:
		after, ordered, start, bullet := readListMarker(content)
		isEmpty := strings.TrimSpace(after) == ""
		// Paragraph-interruption rules:
		// - empty items: only if inside matching open list
		// - ordered N. items with N != 1: only if inside an open ordered list
		if t := p.currentTip(); t != nil && t.Kind == BlockParagraph {
			if isEmpty && !p.hasOpenList(ordered, bullet) {
				p.handleParagraphLine(line)
				return
			}
			if ordered && start != 1 && !p.hasOpenOrderedList() {
				p.handleParagraphLine(line)
				return
			}
		}
		p.openListItem(line, indent)
	case p.isHTMLBlockStart(content):
		p.closeTip()
		b := &Block{Kind: BlockHTMLBlock, Text: line + "\n"}
		b.HTMLKind = p.htmlKind
		p.appendToTop(b)
		if !p.htmlBlockEnds(line, p.htmlKind) {
			p.openHTML = b
		}
	default:
		p.handleParagraphLine(line)
	}
}

// currentTip returns the current open leaf block, or nil.
func (p *blockParser) currentTip() *Block {
	if len(p.stack) == 0 {
		return nil
	}
	tip := p.stack[len(p.stack)-1]
	switch tip.Kind {
	case BlockParagraph, BlockHeading, BlockCodeBlock, BlockHTMLBlock, BlockThematicBreak:
		return tip
	}
	return nil
}

// closeTip pops the open leaf, if any.
func (p *blockParser) closeTip() {
	if tip := p.currentTip(); tip != nil {
		if tip.Kind == BlockCodeBlock && !tip.Fenced {
			p.pendingCodeBlanks = p.pendingCodeBlanks[:0]
		}
		p.stack = p.stack[:len(p.stack)-1]
	}
}

// closeTo closes containers/leaves until len(p.stack) == depth.
func (p *blockParser) closeTo(depth int) {
	for len(p.stack) > depth {
		p.stack = p.stack[:len(p.stack)-1]
	}
}

// closeAll closes everything.
func (p *blockParser) closeAll() {
	p.stack = p.stack[:1] // keep doc
}

// appendToTop appends a child to the top container on the stack and, if the
// child is a leaf or container, pushes it onto the stack.
func (p *blockParser) appendToTop(b *Block) {
	top := p.stack[len(p.stack)-1]
	for top.Kind == BlockParagraph || top.Kind == BlockHeading ||
		top.Kind == BlockCodeBlock || top.Kind == BlockHTMLBlock {
		if top.Kind == BlockCodeBlock && !top.Fenced {
			p.pendingCodeBlanks = p.pendingCodeBlanks[:0]
		}
		p.stack = p.stack[:len(p.stack)-1]
		top = p.stack[len(p.stack)-1]
	}
	// If we're adding a non-ListItem block while a List is the open top,
	// the list must close (it can't hold anything but items).
	for top.Kind == BlockList && b.Kind != BlockListItem {
		p.stack = p.stack[:len(p.stack)-1]
		top = p.stack[len(p.stack)-1]
	}
	top.Children = append(top.Children, b)
	switch b.Kind {
	case BlockBlockquote, BlockList, BlockListItem:
		p.stack = append(p.stack, b)
	case BlockParagraph, BlockHeading, BlockCodeBlock, BlockHTMLBlock:
		p.stack = append(p.stack, b)
	}
}

// handleParagraphLine continues the open paragraph or starts a new one.
func (p *blockParser) handleParagraphLine(line string) {
	tip := p.currentTip()
	stripped := strings.TrimLeft(line, " \t")
	if tip != nil && tip.Kind == BlockParagraph {
		tip.Text += "\n" + stripped
		return
	}
	p.appendToTop(&Block{Kind: BlockParagraph, Text: stripped})
}

// handleIndentedCode appends a line to an open indented code block, or starts
// one. Any pending blank lines (buffered because we didn't yet know whether
// the code block would continue) are flushed here.
func (p *blockParser) handleIndentedCode(line string) {
	body, _ := stripIndent(line, 4)
	tip := p.currentTip()
	if tip != nil && tip.Kind == BlockCodeBlock && !tip.Fenced {
		for _, b := range p.pendingCodeBlanks {
			tip.Text += b + "\n"
		}
		p.pendingCodeBlanks = p.pendingCodeBlanks[:0]
		tip.Text += body + "\n"
		return
	}
	p.closeTip()
	p.appendToTop(&Block{Kind: BlockCodeBlock, Text: body + "\n"})
}

// markBlankLine records that a blank line was seen. The actual loose-list
// marking is deferred to applyPendingBlank, which runs against the NEXT
// non-blank line's container-continuation result. This way we can tell
// whether the blank belonged inside the deepest open item, or whether it
// sat between items at some shallower level.
func (p *blockParser) markBlankLine() {
	p.pendingBlank = true
	p.pendingBlankBQIdx = -1
	for i := len(p.stack) - 1; i >= 0; i-- {
		if p.stack[i].Kind == BlockBlockquote {
			p.pendingBlankBQIdx = i
			break
		}
	}
}

// applyPendingBlank, called by processLine for non-blank lines, marks list
// items that the pending blank affects. Items closed by failed continuation
// get flagged (the blank sat between them and the new line); the deepest
// item still open below the closure point also gets flagged (the blank was
// at the boundary between blocks within that item).
func (p *blockParser) applyPendingBlank(consumedDepth int, allContinued bool) {
	if !p.pendingBlank {
		return
	}
	p.pendingBlank = false
	minDepth := 0
	if p.pendingBlankBQIdx >= 0 {
		minDepth = p.pendingBlankBQIdx + 1
	}
	if !allContinued {
		start := consumedDepth
		if start < minDepth {
			start = minDepth
		}
		for i := start; i < len(p.stack); i++ {
			c := p.stack[i]
			if c.Kind == BlockBlockquote {
				break
			}
			if c.Kind == BlockListItem {
				c.BlankAfter = true
			}
		}
	}
	walkLimit := len(p.stack)
	if !allContinued {
		walkLimit = consumedDepth
	}
	for i := walkLimit - 1; i >= minDepth; i-- {
		c := p.stack[i]
		switch c.Kind {
		case BlockListItem:
			c.BlankAfter = true
			return
		case BlockList:
			// Mark this list's last item (blank sat between items here) but
			// keep walking up — an enclosing list item may also be affected.
			if len(c.Children) > 0 {
				last := c.Children[len(c.Children)-1]
				if last.Kind == BlockListItem {
					last.BlankAfter = true
				}
			}
		case BlockBlockquote:
			return
		}
	}
}

// tryContinueContainer returns the line remainder after consuming the
// continuation prefix for c, and whether c continues.
func (p *blockParser) tryContinueContainer(c *Block, line string) (string, bool) {
	switch c.Kind {
	case BlockBlockquote:
		indent, content := splitIndent(line)
		if indent <= 3 && len(content) > 0 && content[0] == '>' {
			rest := content[1:]
			// Expand leading tabs from the column right after '>'.
			rest = expandLeadingTabsFromCol(rest, indent+1)
			if len(rest) > 0 && rest[0] == ' ' {
				rest = rest[1:]
			}
			return rest, true
		}
		return "", false
	case BlockListItem:
		// A list item continues if the line has at least c.Level leading
		// spaces of indent OR the line is blank. Exception: a blank line
		// at the start of an EMPTY item closes the item (CM rule).
		if isBlankLine(line) {
			if len(c.Children) == 0 {
				return "", false
			}
			return strings.TrimLeft(line, " \t"), true
		}
		if measureIndent(line) >= c.Level {
			stripped, _ := stripIndent(line, c.Level)
			return stripped, true
		}
		return "", false
	case BlockList:
		// Lists always continue; their items handle indent.
		return line, true
	}
	return line, true
}

// openListItem opens (or continues into) a list item, given the line and the
// leading indent in spaces.
func (p *blockParser) openListItem(line string, indent int) {
	content := line
	if indent > 0 {
		content, _ = stripIndent(line, indent)
	}
	after, isOrdered, startNum, bullet := readListMarker(content)
	if bullet == 0 {
		// Shouldn't happen since we checked isListItemStart >= 0.
		p.handleParagraphLine(line)
		return
	}
	// Expand any leading tabs in `after` from the column position right
	// after the marker so width calculations stay accurate.
	after = expandLeadingTabsFromCol(after, indent+markerWidth(startNum, isOrdered))
	// Determine content indent for this item.
	// Count leading spaces of `after`. If 1-4, that's the gap. If 5+, only 1
	// of those spaces counts as the gap; the rest is indented-code content.
	leadSpaces := 0
	for leadSpaces < len(after) && after[leadSpaces] == ' ' {
		leadSpaces++
	}
	contentEmpty := strings.TrimSpace(after) == ""
	if contentEmpty {
		// Empty item — gap is always 1 column regardless of trailing spaces.
		leadSpaces = 1
	} else if leadSpaces >= 5 {
		// 5+ spaces means the gap is 1; the rest is indented-code content.
		leadSpaces = 1
	} else if leadSpaces == 0 {
		leadSpaces = 1
	}
	contentIndent := indent + markerWidth(startNum, isOrdered) + leadSpaces

	// Find or create the containing list.
	top := p.stack[len(p.stack)-1]
	for top.Kind == BlockParagraph || top.Kind == BlockHeading ||
		top.Kind == BlockCodeBlock || top.Kind == BlockHTMLBlock {
		p.stack = p.stack[:len(p.stack)-1]
		top = p.stack[len(p.stack)-1]
	}
	// If top is a list whose marker doesn't match, that list must close
	// (a list change-of-marker starts a new list).
	for top.Kind == BlockList && (top.BulletChar != bullet || top.Ordered != isOrdered) {
		p.stack = p.stack[:len(p.stack)-1]
		top = p.stack[len(p.stack)-1]
	}
	var list *Block
	if top.Kind == BlockList {
		list = top
	} else {
		list = &Block{Kind: BlockList, Ordered: isOrdered, Start: startNum, BulletChar: bullet, Tight: true}
		top.Children = append(top.Children, list)
		p.stack = append(p.stack, list)
	}
	item := &Block{Kind: BlockListItem, Level: contentIndent, TaskState: -1}
	list.Children = append(list.Children, item)
	p.stack = append(p.stack, item)

	// Strip leading spaces (up to leadSpaces) from after to get the item's content.
	rest := after
	for i := 0; i < leadSpaces && len(rest) > 0 && rest[0] == ' '; i++ {
		rest = rest[1:]
	}

	// Task list detection: rest begins with `[ ]` or `[x]` (case insensitive)
	// followed by space.
	if p.opts.Tasklists && len(rest) >= 3 && rest[0] == '[' && rest[2] == ']' &&
		(rest[1] == ' ' || rest[1] == 'x' || rest[1] == 'X') &&
		(len(rest) == 3 || rest[3] == ' ' || rest[3] == '\t') {
		if rest[1] == ' ' {
			item.TaskState = 0
		} else {
			item.TaskState = 1
		}
		rest = rest[3:]
		if len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\t') {
			rest = rest[1:]
		}
	}

	if rest != "" && strings.TrimSpace(rest) != "" {
		p.processContent(rest, true)
	}
}

// readListMarker recognizes - * + as bullets and 1. 1) as ordered.
// Returns the content after the marker, whether the list is ordered, the
// start number for ordered lists, and the bullet/punct char used to match
// items to the same list.
func readListMarker(s string) (string, bool, int, byte) {
	if len(s) == 0 {
		return "", false, 0, 0
	}
	c := s[0]
	if c == '-' || c == '*' || c == '+' {
		if len(s) == 1 {
			return "", false, 0, c
		}
		if s[1] == ' ' || s[1] == '\t' {
			return s[1:], false, 0, c
		}
		return "", false, 0, 0
	}
	// Ordered: 1-9 digits, then '.' or ')', then space/tab/EOL.
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' && i < 9 {
		i++
	}
	if i == 0 || i >= len(s) {
		return "", false, 0, 0
	}
	if s[i] != '.' && s[i] != ')' {
		return "", false, 0, 0
	}
	if i+1 < len(s) && s[i+1] != ' ' && s[i+1] != '\t' {
		return "", false, 0, 0
	}
	n := 0
	for k := 0; k < i; k++ {
		n = n*10 + int(s[k]-'0')
	}
	rest := ""
	if i+1 < len(s) {
		rest = s[i+1:]
	}
	return rest, true, n, s[i]
}

func markerWidth(start int, ordered bool) int {
	if !ordered {
		return 1
	}
	w := 1 // for '.' or ')'
	if start == 0 {
		w++ // "0"
	}
	for start > 0 {
		w++
		start /= 10
	}
	return w
}

// finalizeParagraphs strips leading link reference definitions from each
// paragraph (in any block scope) and records them into refs. Paragraphs
// emptied by this stripping are removed. We also trim trailing whitespace
// from the last line of paragraphs and the text of setext-style headings.
func finalizeParagraphs(b *Block, refs refMap) {
	if b.Kind == BlockParagraph && b.Text != "" {
		b.Text = stripLeadingRefDefs(b.Text, refs)
		b.Text = trimTrailingLineWhitespace(b.Text)
		if strings.TrimSpace(b.Text) == "" {
			b.Kind = BlockDocument
		}
	}
	if b.Kind == BlockHeading && b.Text != "" {
		b.Text = strings.TrimRight(b.Text, " \t")
	}
	if len(b.Children) > 0 {
		out := b.Children[:0]
		for _, c := range b.Children {
			finalizeParagraphs(c, refs)
			if c.Kind == BlockDocument {
				continue
			}
			out = append(out, c)
		}
		b.Children = out
	}
}

// trimTrailingLineWhitespace strips trailing space/tab from the last line of
// a paragraph's text. Trailing whitespace on interior lines is preserved (it
// may be a hard line break, "  \n").
func trimTrailingLineWhitespace(text string) string {
	nl := strings.LastIndexByte(text, '\n')
	if nl < 0 {
		return strings.TrimRight(text, " \t")
	}
	return text[:nl+1] + strings.TrimRight(text[nl+1:], " \t")
}

// stripLeadingRefDefs consumes consecutive ref defs from the start of text,
// recording them in refs, and returns the remaining paragraph text.
func stripLeadingRefDefs(text string, refs refMap) string {
	for {
		lines := strings.Split(text, "\n")
		n, label, ref := tryParseLinkRef(lines, 0)
		if n == 0 {
			return text
		}
		key := normalizeLinkLabel(label)
		if key != "" {
			if _, exists := refs[key]; !exists {
				refs[key] = ref
			}
		}
		idx := 0
		for k := 0; k < n; k++ {
			j := strings.IndexByte(text[idx:], '\n')
			if j < 0 {
				return ""
			}
			idx += j + 1
		}
		text = text[idx:]
		if text == "" {
			return ""
		}
	}
}

// markTightness determines whether each list is tight or loose. A list is
// loose if any of its items contain a blank line, EXCEPT a blank line at the
// very end of the last item.
func markTightness(b *Block) {
	if b.Kind == BlockList {
		loose := false
		for i, item := range b.Children {
			if item.Kind != BlockListItem {
				continue
			}
			// Blank within the item between blocks → loose.
			if item.BlankAfter && len(item.Children) > 1 {
				loose = true
				break
			}
			// Blank between this item and the next → loose.
			if item.BlankAfter && i < len(b.Children)-1 {
				loose = true
				break
			}
		}
		b.Tight = !loose
	}
	for _, c := range b.Children {
		markTightness(c)
	}
}

// --- Line classification helpers --------------------------------------------

func isBlankLine(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return false
		}
	}
	return true
}

// startsNewBlock reports whether line would start a new block (used by lazy
// continuation rule). It checks only constructs that interrupt a paragraph.
// Some rules depend on parser state (e.g., an ordered marker N. interrupts
// only if N==1 OR we are already inside a matching ordered list).
func (p *blockParser) startsNewBlock(line string) bool {
	indent, content := splitIndent(line)
	if indent >= 4 {
		return false
	}
	if isThematicBreak(content) {
		return true
	}
	if isATXHeading(content) {
		return true
	}
	if isFenceLine(content) {
		return true
	}
	if isBlockquoteStart(content) {
		return true
	}
	if hk := htmlBlockStartKind(content); hk > 0 && hk < 7 {
		return true
	}
	if n := isListItemStart(content); n >= 0 {
		after, ordered, start, bullet := readListMarker(content)
		if strings.TrimSpace(after) == "" {
			if !p.hasOpenList(ordered, bullet) {
				return false
			}
		}
		if ordered && start != 1 && !p.hasOpenOrderedList() {
			return false
		}
		_ = n
		return true
	}
	return false
}

func (p *blockParser) hasOpenOrderedList() bool {
	for _, b := range p.stack {
		if b.Kind == BlockList && b.Ordered {
			return true
		}
	}
	return false
}

// hasOpenList reports whether there's an open list on the stack whose marker
// kind matches.
func (p *blockParser) hasOpenList(ordered bool, bullet byte) bool {
	for _, b := range p.stack {
		if b.Kind == BlockList && b.Ordered == ordered && b.BulletChar == bullet {
			return true
		}
	}
	return false
}

// expandLeadingTabs replaces leading tabs in line with the equivalent spaces
// (4-col tab stops from column 0). Stops on the first non-whitespace char;
// tabs in content are not touched.
func expandLeadingTabs(line string) string {
	return expandLeadingTabsFromCol(line, 0)
}

// expandLeadingTabsFromCol is like expandLeadingTabs but treats the line's
// first character as starting at startCol (used after a marker or container
// prefix has been consumed, so tab widths are computed from the right offset).
func expandLeadingTabsFromCol(line string, startCol int) string {
	if !strings.ContainsRune(line, '\t') {
		return line
	}
	var b strings.Builder
	col := startCol
	i := 0
	for i < len(line) {
		c := line[i]
		if c == ' ' {
			b.WriteByte(' ')
			col++
			i++
			continue
		}
		if c == '\t' {
			w := 4 - (col % 4)
			for k := 0; k < w; k++ {
				b.WriteByte(' ')
			}
			col += w
			i++
			continue
		}
		b.WriteString(line[i:])
		return b.String()
	}
	return b.String()
}

func splitIndent(line string) (int, string) {
	i := 0
	col := 0
	for i < len(line) {
		c := line[i]
		if c == ' ' {
			col++
			i++
		} else if c == '\t' {
			col += 4 - (col % 4)
			i++
		} else {
			break
		}
	}
	return col, line[i:]
}

// measureIndent counts leading indent in columns.
func measureIndent(line string) int {
	col, _ := splitIndent(line)
	return col
}

// stripIndent removes up to n columns of leading whitespace, returning the
// remainder and the actual columns removed.
func stripIndent(line string, n int) (string, int) {
	i := 0
	col := 0
	for i < len(line) && col < n {
		c := line[i]
		if c == ' ' {
			col++
			i++
		} else if c == '\t' {
			width := 4 - (col % 4)
			if col+width > n {
				// Partial tab consumption: replace remainder with spaces.
				spaces := col + width - n
				col = n
				return strings.Repeat(" ", spaces) + line[i+1:], col
			}
			col += width
			i++
		} else {
			break
		}
	}
	return line[i:], col
}

func isThematicBreak(s string) bool {
	count := 0
	var ch byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case ' ', '\t':
			continue
		case '-', '*', '_':
			if ch == 0 {
				ch = c
			} else if c != ch {
				return false
			}
			count++
		default:
			return false
		}
	}
	return count >= 3
}

// isSetextUnderline returns 1 (=) or 2 (-) if s is a valid setext underline,
// else 0.
func isSetextUnderline(s string) int {
	// All same char (= or -), optionally surrounded by spaces.
	var ch byte
	count := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' {
			if count > 0 {
				// trailing spaces only allowed
				for j := i; j < len(s); j++ {
					if s[j] != ' ' && s[j] != '\t' {
						return 0
					}
				}
				break
			}
			continue
		}
		if c != '=' && c != '-' {
			return 0
		}
		if ch == 0 {
			ch = c
		} else if ch != c {
			return 0
		}
		count++
	}
	if count == 0 {
		return 0
	}
	if ch == '=' {
		return 1
	}
	return 2
}

func isATXHeading(s string) bool {
	i := 0
	for i < len(s) && s[i] == '#' && i < 7 {
		i++
	}
	if i == 0 || i > 6 {
		return false
	}
	if i == len(s) {
		return true
	}
	return s[i] == ' ' || s[i] == '\t'
}

func parseATXHeading(s string) (int, string) {
	i := 0
	for i < len(s) && s[i] == '#' {
		i++
	}
	level := i
	text := strings.TrimLeft(s[i:], " \t")
	// Strip optional trailing # sequence preceded by space.
	text = strings.TrimRight(text, " \t")
	// Trailing #s only stripped if preceded by space (or text is empty).
	if text == "" {
		return level, ""
	}
	j := len(text)
	for j > 0 && text[j-1] == '#' {
		j--
	}
	if j < len(text) && (j == 0 || text[j-1] == ' ' || text[j-1] == '\t') {
		text = strings.TrimRight(text[:j], " \t")
	}
	return level, text
}

func isFenceLine(s string) bool {
	if len(s) < 3 {
		return false
	}
	c := s[0]
	if c != '`' && c != '~' {
		return false
	}
	n := 0
	for n < len(s) && s[n] == c {
		n++
	}
	if n < 3 {
		return false
	}
	// Backtick info string cannot contain backticks.
	if c == '`' {
		for k := n; k < len(s); k++ {
			if s[k] == '`' {
				return false
			}
		}
	}
	return true
}

func parseFenceOpen(s string) (byte, int, string) {
	ch := s[0]
	n := 0
	for n < len(s) && s[n] == ch {
		n++
	}
	info := strings.TrimSpace(s[n:])
	// Language is the first whitespace-delimited token; remainder is ignored.
	if space := strings.IndexAny(info, " \t"); space >= 0 {
		info = info[:space]
	}
	info = unescapeASCIIPunct(info)
	info = decodeEntities(info)
	return ch, n, info
}

// decodeEntities replaces all HTML entity references in s with their values.
// Stops on the first unrecognized & sequence (left as-is).
func decodeEntities(s string) string {
	if !strings.ContainsRune(s, '&') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '&' {
			if dec, n := decodeEntity(s[i:]); n > 0 {
				b.WriteString(dec)
				i += n
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func unescapeASCIIPunct(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && asciiPunct(s[i+1]) {
			b.WriteByte(s[i+1])
			i++
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isBlockquoteStart(s string) bool {
	return len(s) > 0 && s[0] == '>'
}

func blockquoteRemainder(s string) string {
	if len(s) == 0 {
		return s
	}
	s = s[1:]
	// Expand any leading tabs from column 1 (right after '>') so the optional
	// gap and any content-side indent line up correctly column-wise.
	s = expandLeadingTabsFromCol(s, 1)
	if len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	return s
}

// isListItemStart returns the marker width if s starts with a list marker,
// else -1. Note this returns >= 0 even for a blank-content list item.
func isListItemStart(s string) int {
	if len(s) == 0 {
		return -1
	}
	c := s[0]
	if c == '-' || c == '*' || c == '+' {
		if len(s) == 1 {
			return 1
		}
		if s[1] == ' ' || s[1] == '\t' {
			return 1
		}
		return -1
	}
	if c >= '0' && c <= '9' {
		i := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' && i < 9 {
			i++
		}
		if i >= len(s) {
			return -1
		}
		if s[i] != '.' && s[i] != ')' {
			return -1
		}
		if i+1 < len(s) && s[i+1] != ' ' && s[i+1] != '\t' {
			return -1
		}
		return i + 1
	}
	return -1
}

func allByte(s string, b byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != b {
			return false
		}
	}
	return true
}

// --- HTML block detection ---------------------------------------------------

// htmlBlockStartKind returns 1..7 if line starts an HTML block, else 0.
func htmlBlockStartKind(s string) uint8 {
	if len(s) == 0 || s[0] != '<' {
		return 0
	}
	rest := s[1:]
	// Type 2: <!--
	if strings.HasPrefix(rest, "!--") {
		return 2
	}
	// Type 3: <?
	if strings.HasPrefix(rest, "?") {
		return 3
	}
	// Type 4: <! followed by ASCII letter
	if strings.HasPrefix(rest, "!") && len(rest) > 1 {
		c := rest[1]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			return 4
		}
	}
	// Type 5: <![CDATA[
	if strings.HasPrefix(rest, "![CDATA[") {
		return 5
	}
	// Type 1: <script, <pre, <style, <textarea
	lower := strings.ToLower(rest)
	for _, tag := range []string{"script", "pre", "style", "textarea"} {
		if strings.HasPrefix(lower, tag) {
			after := lower[len(tag):]
			if after == "" || after[0] == ' ' || after[0] == '\t' || after[0] == '>' {
				return 1
			}
		}
	}
	// Type 6: <tagname or </tagname (specific block-level tags)
	startIdx := 0
	if strings.HasPrefix(rest, "/") {
		startIdx = 1
	}
	tag, tagEnd := readTagName(rest[startIdx:])
	if tag != "" {
		if isBlockTag(strings.ToLower(tag)) {
			// Followed by whitespace, end of line, > or />
			after := rest[startIdx+tagEnd:]
			if after == "" || after[0] == ' ' || after[0] == '\t' || after[0] == '>' ||
				strings.HasPrefix(after, "/>") {
				return 6
			}
		}
	}
	// Type 7: a complete open or close tag on its own line (only outside paragraph).
	// We don't try too hard; conservative heuristic.
	if isCompleteTagOnLine(s) {
		return 7
	}
	return 0
}

func (p *blockParser) isHTMLBlockStart(s string) bool {
	k := htmlBlockStartKind(s)
	if k == 0 {
		return false
	}
	// Type 7 does not interrupt paragraphs.
	if k == 7 {
		if t := p.currentTip(); t != nil && t.Kind == BlockParagraph {
			return false
		}
	}
	p.htmlKind = k
	return true
}

// htmlBlockEnds reports whether the HTML block of kind k closes on this line.
func (p *blockParser) htmlBlockEnds(line string, k uint8) bool {
	switch k {
	case 1:
		lower := strings.ToLower(line)
		return strings.Contains(lower, "</script>") || strings.Contains(lower, "</pre>") ||
			strings.Contains(lower, "</style>") || strings.Contains(lower, "</textarea>")
	case 2:
		return strings.Contains(line, "-->")
	case 3:
		return strings.Contains(line, "?>")
	case 4:
		return strings.Contains(line, ">")
	case 5:
		return strings.Contains(line, "]]>")
	case 6, 7:
		return isBlankLine(line)
	}
	return false
}

func readTagName(s string) (string, int) {
	if len(s) == 0 {
		return "", 0
	}
	c := s[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
		return "", 0
	}
	i := 1
	for i < len(s) {
		c = s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' {
			i++
			continue
		}
		break
	}
	return s[:i], i
}

func isBlockTag(name string) bool {
	switch name {
	case "address", "article", "aside", "base", "basefont", "blockquote", "body",
		"caption", "center", "col", "colgroup", "dd", "details", "dialog", "dir",
		"div", "dl", "dt", "fieldset", "figcaption", "figure", "footer", "form",
		"frame", "frameset", "h1", "h2", "h3", "h4", "h5", "h6", "head", "header",
		"hr", "html", "iframe", "legend", "li", "link", "main", "menu", "menuitem",
		"nav", "noframes", "ol", "optgroup", "option", "p", "param", "search",
		"section", "summary", "table", "tbody", "td", "tfoot", "th", "thead",
		"title", "tr", "track", "ul":
		return true
	}
	return false
}

// isCompleteTagOnLine checks for HTML block type 7: a complete open or close
// tag (per HTML syntax) optionally followed by whitespace to end of line.
// Reuses the inline matchRawHTML matcher so we don't false-positive on
// things like autolinks (e.g., <https://...>).
func isCompleteTagOnLine(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || s[0] != '<' {
		return false
	}
	raw, n := matchRawHTML(s)
	if raw == "" {
		return false
	}
	rest := strings.TrimSpace(s[n:])
	return rest == ""
}
