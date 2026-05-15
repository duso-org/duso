// Package markdown is a CommonMark-compliant markdown parser/renderer
// purpose-built as a small, dependency-free replacement for goldmark.
//
// Pipeline:
//
//	source bytes
//	  └─ ScanLinkRefs:       collects [label]: dest "title" definitions
//	  └─ ParseBlocks:        line-by-line block tree
//	      └─ on render:      inlines parsed lazily per block
//	  └─ RenderHTML / RenderANSI
//
// There is no AST visitor; renderers walk Block / Inline structs directly.
package markdown

// Options controls HTML rendering. Defaults match the previous goldmark setup.
type Options struct {
	Tables        bool
	Strikethrough bool
	Tasklists     bool
	Smartquotes   bool
	HeadingIDs    bool
}

// DefaultOptions returns the defaults used by markdown_html() in duso.
func DefaultOptions() Options {
	return Options{
		Tables:        true,
		Strikethrough: true,
		Tasklists:     true,
		Smartquotes:   true,
		HeadingIDs:    true,
	}
}

// Theme holds ANSI escape sequences for terminal rendering. An empty Theme
// yields plain text (no escape codes), which is how ToText is implemented.
type Theme struct {
	H1, H2, H3, H4, H5, H6 string
	CodeBlock              string
	CodeInline             string
	Blockquote             string
	ListItem               string
	Bold                   string
	Italic                 string
	Strike                 string
	Link                   string
	HR                     string
	Reset                  string
}

// DefaultTheme matches duso's default_ansi_theme().
func DefaultTheme() *Theme {
	return &Theme{
		H1:         "\033[33m\033[1m\033[4m",
		H2:         "\033[37m\033[1m\033[4m",
		H3:         "\033[37m\033[4m",
		H4:         "\033[37m\033[4m",
		H5:         "\033[37m\033[4m",
		H6:         "\033[37m\033[4m",
		CodeBlock:  "\033[32m",
		CodeInline: "\033[32m",
		Blockquote: "\033[90m",
		ListItem:   "",
		Bold:       "\033[1m",
		Italic:     "\033[3m",
		Strike:     "\033[9m",
		Link:       "\033[34m\033[1m\033[4m",
		HR:         "\033[90m",
		Reset:      "\033[0m",
	}
}

// EmptyTheme returns a theme with no escape codes — used for plain text output.
func EmptyTheme() *Theme {
	return &Theme{}
}

// ToHTML renders markdown source to HTML.
func ToHTML(src string, opts Options) string {
	doc, refs := parse(src, opts)
	return renderHTML(doc, refs, opts)
}

// ToANSI renders markdown source to ANSI-escaped terminal output.
func ToANSI(src string, theme *Theme) string {
	if theme == nil {
		theme = DefaultTheme()
	}
	opts := DefaultOptions()
	opts.Smartquotes = false // smartquotes are HTML-only
	doc, refs := parse(src, opts)
	return renderANSI(doc, refs, theme)
}

// ToText renders markdown source to plain text (no ANSI codes).
func ToText(src string) string {
	return ToANSI(src, EmptyTheme())
}

// parse runs block parsing; link reference definitions are extracted from
// paragraph-start positions during finalization (so they respect block
// boundaries — defs inside code blocks and other non-paragraph contexts are
// ignored, as the spec requires).
func parse(src string, opts Options) (*Block, refMap) {
	refs := refMap{}
	doc := parseBlocksWithRefs(src, opts, refs)
	return doc, refs
}
