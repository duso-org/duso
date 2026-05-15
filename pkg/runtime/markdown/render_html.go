package markdown

import (
	"regexp"
	"strings"
)

// renderHTML emits HTML for the document tree.
func renderHTML(doc *Block, refs refMap, opts Options) string {
	r := &htmlRenderer{refs: refs, opts: opts}
	r.renderChildren(doc)
	return r.b.String()
}

type htmlRenderer struct {
	b           strings.Builder
	refs        refMap
	opts        Options
	tightItemCx bool // true while rendering direct children of a tight list item
}

func (r *htmlRenderer) renderChildren(b *Block) {
	for _, c := range b.Children {
		r.renderBlock(c)
	}
}

func (r *htmlRenderer) renderBlock(b *Block) {
	switch b.Kind {
	case BlockParagraph:
		// In a tight list, paragraph children of list items render without <p>.
		if r.tightItemCx {
			r.renderInlines(parseInlines(b.Text, r.refs, r.opts))
			return
		}
		r.b.WriteString("<p>")
		r.renderInlines(parseInlines(b.Text, r.refs, r.opts))
		r.b.WriteString("</p>\n")
	case BlockHeading:
		tag := htmlHeadingTag(b.Level)
		r.b.WriteByte('<')
		r.b.WriteString(tag)
		if r.opts.HeadingIDs {
			id := slugify(stripInlineSyntax(b.Text))
			if id != "" {
				r.b.WriteString(` id="`)
				escapeAttr(&r.b, id)
				r.b.WriteString(`"`)
			}
		}
		r.b.WriteByte('>')
		r.renderInlines(parseInlines(b.Text, r.refs, r.opts))
		r.b.WriteString("</")
		r.b.WriteString(tag)
		r.b.WriteString(">\n")
	case BlockBlockquote:
		r.b.WriteString("<blockquote>\n")
		r.renderChildren(b)
		r.b.WriteString("</blockquote>\n")
	case BlockList:
		tag := "ul"
		if b.Ordered {
			tag = "ol"
		}
		r.b.WriteByte('<')
		r.b.WriteString(tag)
		if b.Ordered && b.Start != 1 {
			r.b.WriteString(` start="`)
			r.b.WriteString(itoa(b.Start))
			r.b.WriteByte('"')
		}
		r.b.WriteString(">\n")
		for _, item := range b.Children {
			if item.Kind != BlockListItem {
				continue
			}
			r.renderListItem(item, b.Tight)
		}
		r.b.WriteString("</")
		r.b.WriteString(tag)
		r.b.WriteString(">\n")
	case BlockCodeBlock:
		r.b.WriteString("<pre><code")
		if b.Lang != "" {
			r.b.WriteString(` class="language-`)
			escapeAttr(&r.b, b.Lang)
			r.b.WriteString(`"`)
		}
		r.b.WriteByte('>')
		escapeHTML(&r.b, b.Text)
		r.b.WriteString("</code></pre>\n")
	case BlockHTMLBlock:
		r.b.WriteString(b.Text)
	case BlockThematicBreak:
		r.b.WriteString("<hr />\n")
	case BlockTable:
		r.renderTable(b)
	}
}

func (r *htmlRenderer) renderListItem(item *Block, tightList bool) {
	r.b.WriteString("<li>")
	// Task list checkbox.
	if item.TaskState >= 0 {
		if item.TaskState == 1 {
			r.b.WriteString(`<input type="checkbox" checked="" disabled="" /> `)
		} else {
			r.b.WriteString(`<input type="checkbox" disabled="" /> `)
		}
	}
	// Tight item: paragraphs render inline (no <p>). Non-paragraph children
	// get a preceding newline so they sit on their own line. If the first
	// child is non-paragraph we still emit that newline (gives <li>\n<ul>...).
	if tightList {
		prev := r.tightItemCx
		r.tightItemCx = true
		for i, c := range item.Children {
			if c.Kind == BlockParagraph {
				r.renderBlock(c)
				if i < len(item.Children)-1 {
					r.b.WriteByte('\n')
				}
				continue
			}
			if i == 0 {
				r.b.WriteByte('\n')
			}
			r.tightItemCx = false
			r.renderBlock(c)
			r.tightItemCx = true
		}
		r.tightItemCx = prev
		r.b.WriteString("</li>\n")
		return
	}
	// Loose item: if empty, render as <li></li>; otherwise normal rendering.
	if len(item.Children) == 0 {
		r.b.WriteString("</li>\n")
		return
	}
	r.b.WriteByte('\n')
	r.renderChildren(item)
	r.b.WriteString("</li>\n")
}

func (r *htmlRenderer) renderTable(b *Block) {
	if len(b.Children) == 0 {
		return
	}
	r.b.WriteString("<table>\n")
	header := b.Children[0]
	r.b.WriteString("<thead>\n")
	r.renderTableRow(header, b.Aligns, true)
	r.b.WriteString("</thead>\n")
	if len(b.Children) > 1 {
		r.b.WriteString("<tbody>\n")
		for _, row := range b.Children[1:] {
			r.renderTableRow(row, b.Aligns, false)
		}
		r.b.WriteString("</tbody>\n")
	}
	r.b.WriteString("</table>\n")
}

func (r *htmlRenderer) renderTableRow(row *Block, aligns []Align, header bool) {
	r.b.WriteString("<tr>\n")
	for i, cell := range row.Children {
		tag := "td"
		if header {
			tag = "th"
		}
		r.b.WriteByte('<')
		r.b.WriteString(tag)
		if i < len(aligns) {
			switch aligns[i] {
			case AlignLeft:
				r.b.WriteString(` align="left"`)
			case AlignCenter:
				r.b.WriteString(` align="center"`)
			case AlignRight:
				r.b.WriteString(` align="right"`)
			}
		}
		r.b.WriteByte('>')
		r.renderInlines(parseInlines(cell.Text, r.refs, r.opts))
		r.b.WriteString("</")
		r.b.WriteString(tag)
		r.b.WriteString(">\n")
	}
	r.b.WriteString("</tr>\n")
}

func (r *htmlRenderer) renderInlines(nodes []*Inline) {
	for _, n := range nodes {
		r.renderInline(n)
	}
}

func (r *htmlRenderer) renderInline(n *Inline) {
	switch n.Kind {
	case InlineText:
		t := n.Text
		if r.opts.Smartquotes {
			t = applySmartquotes(t)
		}
		escapeHTML(&r.b, t)
	case InlineSoftBreak:
		r.b.WriteByte('\n')
	case InlineHardBreak:
		r.b.WriteString("<br />\n")
	case InlineEmphasis:
		r.b.WriteString("<em>")
		r.renderInlines(n.Children)
		r.b.WriteString("</em>")
	case InlineStrong:
		r.b.WriteString("<strong>")
		r.renderInlines(n.Children)
		r.b.WriteString("</strong>")
	case InlineStrike:
		r.b.WriteString("<del>")
		r.renderInlines(n.Children)
		r.b.WriteString("</del>")
	case InlineCode:
		r.b.WriteString("<code>")
		escapeHTML(&r.b, n.Text)
		r.b.WriteString("</code>")
	case InlineLink:
		r.b.WriteString(`<a href="`)
		percentEncode(&r.b, n.URL)
		r.b.WriteByte('"')
		if n.Title != "" {
			r.b.WriteString(` title="`)
			escapeAttr(&r.b, n.Title)
			r.b.WriteByte('"')
		}
		r.b.WriteByte('>')
		r.renderInlines(n.Children)
		r.b.WriteString("</a>")
	case InlineImage:
		r.b.WriteString(`<img src="`)
		percentEncode(&r.b, n.URL)
		r.b.WriteString(`" alt="`)
		var alt strings.Builder
		collectAltText(&alt, n.Children)
		escapeAttr(&r.b, alt.String())
		r.b.WriteByte('"')
		if n.Title != "" {
			r.b.WriteString(` title="`)
			escapeAttr(&r.b, n.Title)
			r.b.WriteByte('"')
		}
		r.b.WriteString(" />")
	case InlineAutolink:
		r.b.WriteString(`<a href="`)
		percentEncode(&r.b, n.URL)
		r.b.WriteString(`">`)
		escapeHTML(&r.b, n.Text)
		r.b.WriteString("</a>")
	case InlineRawHTML:
		r.b.WriteString(n.Text)
	}
}

func collectAltText(b *strings.Builder, nodes []*Inline) {
	for _, n := range nodes {
		switch n.Kind {
		case InlineText, InlineCode, InlineAutolink:
			b.WriteString(n.Text)
		case InlineSoftBreak, InlineHardBreak:
			b.WriteByte(' ')
		case InlineEmphasis, InlineStrong, InlineStrike, InlineLink, InlineImage:
			collectAltText(b, n.Children)
		}
	}
}

// htmlHeadingTag returns h1..h6 with clamping.
func htmlHeadingTag(level int) string {
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	return "h" + string(rune('0'+level))
}

// itoa converts a non-negative int to its decimal string.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// applySmartquotes performs a simple text-level transformation of straight
// quotes/apostrophes/dashes to their typographic equivalents.
func applySmartquotes(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 8)
	// We rely on byte-level context which works fine for ASCII; for non-ASCII
	// neighbors we err on the side of opening/closing based on alphabetic check.
	prev := byte(' ')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\'':
			if isAlnum(prev) {
				b.WriteString("’") // ’
			} else {
				b.WriteString("‘") // ‘
			}
		case '"':
			if isAlnum(prev) || prev == '.' || prev == ',' || prev == '!' || prev == '?' {
				b.WriteString("”") // ”
			} else {
				b.WriteString("“") // “
			}
		case '-':
			if i+1 < len(s) && s[i+1] == '-' {
				if i+2 < len(s) && s[i+2] == '-' {
					b.WriteString("—") // —
					i += 2
					prev = '-'
					continue
				}
				b.WriteString("–") // –
				i++
				prev = '-'
				continue
			}
			b.WriteByte(c)
		case '.':
			if i+2 < len(s) && s[i+1] == '.' && s[i+2] == '.' {
				b.WriteString("…") // …
				i += 2
				prev = '.'
				continue
			}
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
		prev = c
	}
	return b.String()
}

func isAlnum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// stripInlineSyntax removes inline markdown syntax characters from a heading
// text, used as input to slugify(). It's intentionally simple.
var stripSyntaxRE = regexp.MustCompile(`[*_` + "`" + `~]`)

func stripInlineSyntax(s string) string {
	return stripSyntaxRE.ReplaceAllString(s, "")
}

// slugify converts heading text to a valid HTML id. Matches the prior
// behavior in pkg/runtime/builtin_markdown.go.
var (
	slugTagRE    = regexp.MustCompile(`<[^>]+>`)
	slugSpaceRE  = regexp.MustCompile(`[\s_]+`)
	slugNonAlnum = regexp.MustCompile(`[^a-z0-9-]`)
	slugDashRE   = regexp.MustCompile(`-+`)
)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = slugTagRE.ReplaceAllString(s, "")
	s = slugSpaceRE.ReplaceAllString(s, "-")
	s = slugNonAlnum.ReplaceAllString(s, "")
	s = slugDashRE.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

