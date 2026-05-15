package markdown

import (
	"strings"
)

// renderANSI emits ANSI-styled (or plain, if Theme is empty) terminal output.
func renderANSI(doc *Block, refs refMap, theme *Theme) string {
	r := &ansiRenderer{theme: theme, refs: refs, opts: DefaultOptions()}
	r.opts.Smartquotes = false
	r.renderChildren(doc)
	return r.b.String()
}

type ansiRenderer struct {
	b         strings.Builder
	theme     *Theme
	refs      refMap
	opts      Options
	listDepth int
}

func (r *ansiRenderer) renderChildren(b *Block) {
	for i, c := range b.Children {
		r.renderBlock(c)
		// Add blank line between top-level blocks (except after last and
		// inside list items).
		_ = i
	}
}

func (r *ansiRenderer) renderBlock(b *Block) {
	switch b.Kind {
	case BlockParagraph:
		r.renderInlines(parseInlines(b.Text, r.refs, r.opts))
		r.b.WriteByte('\n')
		if r.listDepth == 0 {
			r.b.WriteByte('\n')
		}
	case BlockHeading:
		r.b.WriteString(r.themeForHeading(b.Level))
		r.renderInlines(parseInlines(b.Text, r.refs, r.opts))
		r.b.WriteString(r.theme.Reset)
		r.b.WriteString("\n\n")
	case BlockBlockquote:
		// Render children and prefix each rendered line with the blockquote style.
		// Simplest approach: render to a temp buffer, then prefix.
		sub := &ansiRenderer{theme: r.theme, refs: r.refs, opts: r.opts, listDepth: r.listDepth}
		sub.renderChildren(b)
		out := sub.b.String()
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		for _, ln := range lines {
			r.b.WriteString(r.theme.Blockquote)
			r.b.WriteString("│ ")
			r.b.WriteString(ln)
			r.b.WriteString(r.theme.Reset)
			r.b.WriteByte('\n')
		}
		if r.listDepth == 0 {
			r.b.WriteByte('\n')
		}
	case BlockList:
		r.renderList(b)
	case BlockCodeBlock:
		r.b.WriteString(r.theme.CodeBlock)
		lines := strings.Split(strings.TrimRight(b.Text, "\n"), "\n")
		for _, ln := range lines {
			r.b.WriteString("  ")
			r.b.WriteString(ln)
			r.b.WriteByte('\n')
		}
		r.b.WriteString(r.theme.Reset)
		if r.listDepth == 0 {
			r.b.WriteByte('\n')
		}
	case BlockHTMLBlock:
		// Skip raw HTML in ANSI output.
	case BlockThematicBreak:
		r.b.WriteString(r.theme.HR)
		r.b.WriteString("---")
		r.b.WriteString(r.theme.Reset)
		r.b.WriteString("\n\n")
	case BlockTable:
		r.renderTable(b)
	}
}

func (r *ansiRenderer) renderList(b *Block) {
	r.listDepth++
	defer func() { r.listDepth-- }()
	num := b.Start
	if num == 0 {
		num = 1
	}
	for _, item := range b.Children {
		if item.Kind != BlockListItem {
			continue
		}
		indent := strings.Repeat("  ", r.listDepth-1)
		r.b.WriteString(indent)
		if b.Ordered {
			r.b.WriteString(itoa(num))
			r.b.WriteString(". ")
			num++
		} else {
			r.b.WriteString("• ")
		}
		// Task checkbox if applicable.
		if item.TaskState == 0 {
			r.b.WriteString("[ ] ")
		} else if item.TaskState == 1 {
			r.b.WriteString("[x] ")
		}
		r.renderListItemContent(item)
	}
	if r.listDepth == 1 {
		r.b.WriteByte('\n')
	}
}

func (r *ansiRenderer) renderListItemContent(item *Block) {
	for i, c := range item.Children {
		if c.Kind == BlockParagraph {
			r.renderInlines(parseInlines(c.Text, r.refs, r.opts))
			r.b.WriteByte('\n')
			continue
		}
		// Non-paragraph: render normally; ensure leading newline if first child.
		_ = i
		r.renderBlock(c)
	}
}

func (r *ansiRenderer) renderTable(b *Block) {
	// Render each row as: cell | cell | cell, with header underlined.
	if len(b.Children) == 0 {
		return
	}
	for ri, row := range b.Children {
		if ri == 0 && r.theme.H3 != "" {
			r.b.WriteString(r.theme.H3) // header style approximation
		}
		for ci, cell := range row.Children {
			if ci > 0 {
				r.b.WriteString(" │ ")
			}
			r.renderInlines(parseInlines(cell.Text, r.refs, r.opts))
		}
		if ri == 0 && r.theme.H3 != "" {
			r.b.WriteString(r.theme.Reset)
		}
		r.b.WriteByte('\n')
		if ri == 0 {
			// Separator below header.
			r.b.WriteString(r.theme.HR)
			r.b.WriteString(strings.Repeat("─", 40))
			r.b.WriteString(r.theme.Reset)
			r.b.WriteByte('\n')
		}
	}
	r.b.WriteByte('\n')
}

func (r *ansiRenderer) themeForHeading(level int) string {
	switch level {
	case 1:
		return r.theme.H1
	case 2:
		return r.theme.H2
	case 3:
		return r.theme.H3
	case 4:
		return r.theme.H4
	case 5:
		return r.theme.H5
	case 6:
		return r.theme.H6
	}
	return ""
}

func (r *ansiRenderer) renderInlines(nodes []*Inline) {
	for _, n := range nodes {
		r.renderInline(n)
	}
}

func (r *ansiRenderer) renderInline(n *Inline) {
	switch n.Kind {
	case InlineText:
		r.b.WriteString(n.Text)
	case InlineSoftBreak:
		r.b.WriteByte('\n')
	case InlineHardBreak:
		r.b.WriteByte('\n')
	case InlineEmphasis:
		r.b.WriteString(r.theme.Italic)
		r.renderInlines(n.Children)
		r.b.WriteString(r.theme.Reset)
	case InlineStrong:
		r.b.WriteString(r.theme.Bold)
		r.renderInlines(n.Children)
		r.b.WriteString(r.theme.Reset)
	case InlineStrike:
		r.b.WriteString(r.theme.Strike)
		r.renderInlines(n.Children)
		r.b.WriteString(r.theme.Reset)
	case InlineCode:
		r.b.WriteString(r.theme.CodeInline)
		r.b.WriteString(n.Text)
		r.b.WriteString(r.theme.Reset)
	case InlineLink:
		r.b.WriteString(r.theme.Link)
		r.b.WriteByte('[')
		r.renderInlines(n.Children)
		r.b.WriteString(r.theme.Link)
		r.b.WriteString("](")
		r.b.WriteString(n.URL)
		r.b.WriteByte(')')
		r.b.WriteString(r.theme.Reset)
	case InlineImage:
		// Images skipped in ANSI output (alt text rendered if present).
		var alt strings.Builder
		collectAltText(&alt, n.Children)
		if alt.Len() > 0 {
			r.b.WriteString(alt.String())
		}
	case InlineAutolink:
		r.b.WriteString(r.theme.Link)
		r.b.WriteString(n.Text)
		r.b.WriteString(r.theme.Reset)
	case InlineRawHTML:
		// Drop raw HTML for terminal output.
	}
}
