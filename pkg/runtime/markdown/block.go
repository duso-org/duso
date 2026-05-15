package markdown

// BlockKind discriminates Block.
type BlockKind uint8

const (
	BlockDocument BlockKind = iota
	BlockParagraph
	BlockHeading
	BlockBlockquote
	BlockList
	BlockListItem
	BlockCodeBlock
	BlockHTMLBlock
	BlockThematicBreak
	BlockTable
	BlockTableRow
	BlockTableCell
)

// Align is a table column alignment.
type Align uint8

const (
	AlignNone Align = iota
	AlignLeft
	AlignCenter
	AlignRight
)

// Block is a single block-level node in the tree. Container blocks
// (Document, Blockquote, List, ListItem, Table*) use Children. Leaf
// blocks (Paragraph, Heading, CodeBlock, HTMLBlock) carry raw text in
// Text. Inline parsing happens lazily during render.
type Block struct {
	Kind     BlockKind
	Children []*Block

	// Text holds raw content for leaf blocks (one line per logical line,
	// joined by '\n'). For code blocks the literal contents; for paragraphs
	// and headings the un-parsed inline text.
	Text string

	// Heading level (1-6) or table header flag (use IsHeader).
	Level int

	// List fields.
	Ordered    bool
	Tight      bool
	Start      int  // ordered list start number
	BulletChar byte // - * + . )

	// List item fields.
	TaskState  int8 // -1 = not a task; 0 = unchecked; 1 = checked
	BlankAfter bool // a blank line was seen inside this list item; used for loose detection

	// Code block fields.
	Lang   string
	Fenced bool

	// HTML block kind 1..7 (per CommonMark spec).
	HTMLKind uint8

	// Table fields.
	Aligns   []Align
	IsHeader bool
}

// InlineKind discriminates Inline.
type InlineKind uint8

const (
	InlineText InlineKind = iota
	InlineSoftBreak
	InlineHardBreak
	InlineEmphasis
	InlineStrong
	InlineStrike
	InlineCode
	InlineLink
	InlineImage
	InlineAutolink
	InlineRawHTML
)

// Inline is a single inline node. Containers (Emphasis, Strong, Strike,
// Link, Image) use Children. Leaves use Text. Links/images set URL/Title.
type Inline struct {
	Kind     InlineKind
	Text     string
	URL      string
	Title    string
	Children []*Inline
}
