package runtime

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// ANSITheme holds ANSI codes for various markdown elements
type ANSITheme struct {
	H1         string
	H2         string
	H3         string
	H4         string
	H5         string
	H6         string
	CodeBlock  string
	CodeInline string
	Blockquote string
	ListItem   string
	Bold       string
	Italic     string
	Link       string
	HR         string
	Reset      string
}

// DefaultANSITheme returns the default ANSI theme matching duso's default_ansi_theme()
// These codes correspond to: ansi.combine(fg="color", bold=true, underline=true, etc.)
func DefaultANSITheme() *ANSITheme {
	return &ANSITheme{
		H1:         "\033[33m\033[1m\033[4m",        // yellow, bold, underline
		H2:         "\033[37m\033[1m\033[4m",        // white, bold, underline
		H3:         "\033[37m\033[4m",               // white, underline
		H4:         "\033[37m\033[4m",               // white, underline
		H5:         "\033[37m\033[4m",               // white, underline
		H6:         "\033[37m\033[4m",               // white, underline
		CodeBlock:  "\033[32m",                      // green
		CodeInline: "\033[32m",                      // green
		Blockquote: "\033[90m",                      // gray
		ListItem:   "",                              // empty - bullet prefix handles styling
		Bold:       "\033[1m",                       // bold
		Italic:     "\033[3m",                       // italic
		Link:       "\033[34m\033[1m\033[4m",       // blue, bold, underline
		HR:         "\033[90m",                      // gray
		Reset:      "\033[0m",                       // reset
	}
}

// ANSIRenderer implements a goldmark renderer for ANSI terminal output
type ANSIRenderer struct {
	theme              *ANSITheme
	listDepth          int
	source             []byte
	paragraphStart     int
	paragraphEnd       int
	lastTextEndPos     int
	inParagraph        bool
	paragraphLineStart bool
	listIsOrdered      []bool // Stack: is list at each depth ordered?
	listItemNumber     []int   // Stack: current item number for ordered lists
}

// NewANSIRenderer creates a new ANSI renderer with the given theme
func NewANSIRenderer(theme *ANSITheme) renderer.NodeRenderer {
	if theme == nil {
		theme = DefaultANSITheme()
	}
	return &ANSIRenderer{
		theme:     theme,
		listDepth: 0,
	}
}

// writeWhitespaceOnly writes only spaces and tabs from the given bytes, skipping other characters
func (r *ANSIRenderer) writeWhitespaceOnly(w util.BufWriter, data []byte) {
	for _, b := range data {
		if b == ' ' || b == '\t' {
			w.WriteByte(b)
		}
	}
}

// findLineStart finds the position of the start of the line (after the previous newline)
func findLineStart(source []byte, pos int) int {
	for pos > 0 && source[pos-1] != '\n' {
		pos--
	}
	return pos
}

// outputLineIndentation writes the indentation from the start of the line to pos, skipping non-whitespace
func (r *ANSIRenderer) outputLineIndentation(w util.BufWriter, source []byte, pos int) {
	lineStart := findLineStart(source, pos)
	if lineStart < pos {
		r.writeWhitespaceOnly(w, source[lineStart:pos])
	}
}


// RegisterFuncs registers all node rendering functions (implements NodeRenderer interface)
func (r *ANSIRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindString, r.renderString)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
}

// renderDocument handles the document root node
func (r *ANSIRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Nothing to do for document itself, just continue to children
	return ast.WalkContinue, nil
}

// renderHeading handles heading nodes
func (r *ANSIRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		r.source = source
		// Reset text position tracking for this new heading
		r.lastTextEndPos = 0
		switch n.Level {
		case 1:
			w.WriteString(r.theme.H1)
		case 2:
			w.WriteString(r.theme.H2)
		case 3:
			w.WriteString(r.theme.H3)
		case 4:
			w.WriteString(r.theme.H4)
		case 5:
			w.WriteString(r.theme.H5)
		case 6:
			w.WriteString(r.theme.H6)
		}
	} else {
		// Reset tracking after heading
		r.lastTextEndPos = 0
		w.WriteString(r.theme.Reset)
		w.WriteString("\n\n")
	}
	return ast.WalkContinue, nil
}

// renderParagraph handles paragraph nodes
func (r *ANSIRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	n := node.(*ast.Paragraph)

	if entering {
		// Reset text position tracking for this new paragraph
		r.lastTextEndPos = 0
		r.inParagraph = true
		// Get first line and find where actual content starts
		firstLine := n.Lines().At(0)
		// Store the line start so we can output leading whitespace before first text
		r.paragraphStart = firstLine.Start

		// Output any leading whitespace (indentation) on the first line only if not in a list
		// (list items handle their own indentation)
		if r.listDepth == 0 {
			r.outputLineIndentation(w, source, firstLine.Start)
		}
	} else {
		// Reset tracking after paragraph
		r.lastTextEndPos = 0
		r.inParagraph = false
		w.WriteString("\n")
		// Add extra blank line after paragraph only if not inside a list
		if r.listDepth == 0 {
			w.WriteString("\n")
		}
	}
	return ast.WalkContinue, nil
}

// renderBlockquote handles blockquote nodes
func (r *ANSIRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	if entering {
		w.WriteString(r.theme.Blockquote)
	} else {
		w.WriteString(r.theme.Reset)
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// renderList handles list nodes
func (r *ANSIRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	n := node.(*ast.List)

	if entering {
		r.listDepth++
		// If this is a nested list (inside a list item), add a newline for separation
		if r.listDepth > 1 {
			w.WriteString("\n")
		}

		// Track if this list is ordered (marker is '.' or ')')
		isOrdered := n.Marker == '.' || n.Marker == ')'
		r.listIsOrdered = append(r.listIsOrdered, isOrdered)
		r.listItemNumber = append(r.listItemNumber, n.Start)
	} else {
		r.listDepth--
		// Pop the list state stacks
		if len(r.listIsOrdered) > 0 {
			r.listIsOrdered = r.listIsOrdered[:len(r.listIsOrdered)-1]
		}
		if len(r.listItemNumber) > 0 {
			r.listItemNumber = r.listItemNumber[:len(r.listItemNumber)-1]
		}
		// Add blank line after top-level list
		if r.listDepth == 0 {
			w.WriteString("\n")
		}
	}
	return ast.WalkContinue, nil
}

// renderListItem handles list item nodes
func (r *ANSIRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	if entering {
		indent := strings.Repeat("  ", r.listDepth-1)
		w.WriteString(indent)
		w.WriteString(r.theme.ListItem)

		// Get current list info (peek at the top of stack)
		depth := r.listDepth - 1
		if depth >= 0 && depth < len(r.listIsOrdered) {
			if r.listIsOrdered[depth] {
				// Ordered list - output number
				num := r.listItemNumber[depth]
				w.WriteString(fmt.Sprintf("%d. ", num))
				r.listItemNumber[depth]++ // Increment for next item
			} else {
				// Unordered list - output bullet
				w.WriteString("• ")
			}
		} else {
			// Fallback to bullet if state is corrupted
			w.WriteString("• ")
		}
	} else {
		// Only write reset if there's actual styling
		if r.theme.ListItem != "" {
			w.WriteString(r.theme.Reset)
		}
		// Paragraph children handle newlines, but ensure one exists between items
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// renderTextBlock is a no-op for text blocks
func (r *ANSIRenderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

// renderFencedCodeBlock handles fenced code blocks
func (r *ANSIRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	if entering {
		n := node.(*ast.FencedCodeBlock)
		w.WriteString(r.theme.CodeBlock)

		// Read all lines and add 2-space indent
		for i := 0; i < n.Lines().Len(); i++ {
			line := n.Lines().At(i)
			w.WriteString("  ")
			w.Write(line.Value(source))
		}

		w.WriteString(r.theme.Reset)
		w.WriteString("\n")
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

// renderCodeBlock handles code blocks
func (r *ANSIRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	if entering {
		n := node.(*ast.CodeBlock)
		w.WriteString(r.theme.CodeBlock)

		// Read all lines and add 2-space indent
		for i := 0; i < n.Lines().Len(); i++ {
			line := n.Lines().At(i)
			w.WriteString("  ")
			w.Write(line.Value(source))
		}

		w.WriteString(r.theme.Reset)
		w.WriteString("\n")
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

// renderThematicBreak handles horizontal rules
func (r *ANSIRenderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString(r.theme.HR)
		w.WriteString("---")
		w.WriteString(r.theme.Reset)
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// renderText handles text nodes
func (r *ANSIRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.Text)
		textStart := n.Segment.Start
		textStop := n.Segment.Stop

		// If this is the first text in a paragraph and there's leading whitespace, preserve it
		if r.inParagraph && r.lastTextEndPos == 0 && r.paragraphStart < textStart && r.listDepth == 0 {
			r.writeWhitespaceOnly(w, source[r.paragraphStart:textStart])
		} else if r.lastTextEndPos > 0 && r.lastTextEndPos < textStart {
			// Check if there's a newline between the last text and this text in the source
			// Only preserve line breaks outside of lists
			if r.listDepth == 0 {
				r.preserveLineBreaksAndIndentation(w, source, r.lastTextEndPos, textStart)
			} else {
				// In lists, check if there's a newline (which indicates transition between list items in same paragraph)
				// If so, don't output anything - the newline is just list formatting
				between := source[r.lastTextEndPos:textStart]
				hasNewline := false
				for _, b := range between {
					if b == '\n' {
						hasNewline = true
						break
					}
				}
				// Only output whitespace if there's no newline (which would be within a list item)
				if !hasNewline {
					for _, b := range between {
						if b == ' ' || b == '\t' {
							w.WriteByte(b)
						}
					}
				}
			}
		}

		w.Write(n.Segment.Value(source))
		r.lastTextEndPos = textStop
	}
	return ast.WalkContinue, nil
}

// preserveLineBreaksAndIndentation outputs newlines and indentation between text positions
func (r *ANSIRenderer) preserveLineBreaksAndIndentation(w util.BufWriter, source []byte, fromPos, toPos int) {
	between := source[fromPos:toPos]
	for i, b := range between {
		if b == '\n' {
			w.WriteString("\n")
			// Preserve indentation after the newline
			if i+1 < len(between) {
				r.writeWhitespaceOnly(w, between[i+1:])
			}
			break
		}
	}
}


// renderCodeSpan handles inline code
func (r *ANSIRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	if entering {
		w.WriteString(r.theme.CodeInline)
	} else {
		w.WriteString(r.theme.Reset)
	}
	return ast.WalkContinue, nil
}

// renderEmphasis handles bold and italic
func (r *ANSIRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	n := node.(*ast.Emphasis)

	if entering {
		if n.Level == 2 {
			w.WriteString(r.theme.Bold)
		} else {
			w.WriteString(r.theme.Italic)
		}
		return ast.WalkContinue, nil
	} else {
		w.WriteString(r.theme.Reset)
		return ast.WalkContinue, nil
	}
}

// renderLink handles links - keeps markdown syntax visible with full coloring
func (r *ANSIRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	r.source = source
	if entering {
		w.WriteString(r.theme.Link)
		w.WriteString("[")
	} else {
		// Reapply link color before closing bracket in case children reset it (e.g., code spans)
		w.WriteString(r.theme.Link)
		w.WriteString("]")
		w.WriteString("(")
		n := node.(*ast.Link)
		w.Write(n.Destination)
		w.WriteString(")")
		w.WriteString(r.theme.Reset)
	}
	return ast.WalkContinue, nil
}

// renderImage skips images as they're not meaningful in terminal
func (r *ANSIRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkSkipChildren, nil
}

// renderString handles string nodes - output them to consume them
func (r *ANSIRenderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	w.Write(n.Value)
	return ast.WalkContinue, nil
}

