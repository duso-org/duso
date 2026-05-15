package runtime

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// slugify converts heading text to a valid HTML id
func slugify(text string) string {
	// Convert to lowercase
	s := strings.ToLower(text)
	// Remove HTML tags
	s = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, "")
	// Replace spaces and underscores with hyphens
	s = regexp.MustCompile(`[\s_]+`).ReplaceAllString(s, "-")
	// Remove non-alphanumeric characters except hyphens
	s = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(s, "")
	// Remove duplicate hyphens
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	// Trim hyphens from edges
	s = strings.Trim(s, "-")
	return s
}

// HeadingIDExtension is a goldmark extension for adding heading IDs
type HeadingIDExtension struct{}

// Extend adds the heading ID transformer to the parser
func (e *HeadingIDExtension) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&HeadingIDTransformer{}, 100),
	))
}

// HeadingIDTransformer adds id attributes to heading nodes
type HeadingIDTransformer struct{}

// Transform walks the AST and adds IDs to headings
func (t *HeadingIDTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		// Extract text content from heading nodes
		var text strings.Builder
		for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
			if textNode, ok := child.(*ast.Text); ok {
				text.Write(textNode.Segment.Value(source))
			}
		}

		// Generate and set ID attribute
		id := slugify(text.String())
		if id != "" {
			heading.SetAttributeString("id", []byte(id))
		}

		return ast.WalkContinue, nil
	})
}

// builtinMarkdownHTML renders markdown to HTML
func builtinMarkdownHTML(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get markdown text (required)
	markdownText, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_html() requires markdown text as first argument")
	}

	// Get options (optional)
	opts := map[string]bool{
		"tables":        true,
		"strikethrough": true,
		"footnotes":     false,
		"tasklists":     true,
		"smartquotes":   true,
	}

	if optArg, ok := args["1"].(map[string]any); ok {
		// Override defaults with provided options
		for key, val := range optArg {
			if boolVal, ok := val.(bool); ok {
				opts[key] = boolVal
			}
		}
	}

	// Build goldmark with requested extensions
	extensions := []goldmark.Extender{}

	if opts["tables"] {
		extensions = append(extensions, extension.Table)
	}
	if opts["strikethrough"] {
		extensions = append(extensions, extension.Strikethrough)
	}
	if opts["footnotes"] {
		extensions = append(extensions, extension.Footnote)
	}
	if opts["tasklists"] {
		extensions = append(extensions, extension.TaskList)
	}
	if opts["smartquotes"] {
		extensions = append(extensions, extension.Typographer)
	}

	// Build extensions list with heading ID transformer
	allExtensions := append(extensions, &HeadingIDExtension{})

	md := goldmark.New(
		goldmark.WithExtensions(allExtensions...),
	)

	// Parse and render
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdownText), &buf); err != nil {
		return nil, fmt.Errorf("markdown_html() error: %w", err)
	}

	return buf.String(), nil
}

// builtinMarkdownANSI renders markdown to ANSI terminal output
func builtinMarkdownANSI(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get markdown text (required)
	markdownText, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_ansi() requires markdown text as first argument")
	}

	// Get options (optional theme)
	theme := DefaultANSITheme()

	if optArg, ok := args["1"].(map[string]any); ok {
		// Apply custom theme values from opts
		// Theme keys: h1, h2, h3, h4, h5, h6, code_start, code_inline, blockquote, list_item, bold, italic, link, hr, reset
		if val, ok := optArg["h1"].(string); ok {
			theme.H1 = val
		}
		if val, ok := optArg["h2"].(string); ok {
			theme.H2 = val
		}
		if val, ok := optArg["h3"].(string); ok {
			theme.H3 = val
		}
		if val, ok := optArg["h4"].(string); ok {
			theme.H4 = val
		}
		if val, ok := optArg["h5"].(string); ok {
			theme.H5 = val
		}
		if val, ok := optArg["h6"].(string); ok {
			theme.H6 = val
		}
		if val, ok := optArg["code_start"].(string); ok {
			theme.CodeBlock = val
		}
		if val, ok := optArg["code_inline"].(string); ok {
			theme.CodeInline = val
		}
		if val, ok := optArg["blockquote"].(string); ok {
			theme.Blockquote = val
		}
		if val, ok := optArg["list_item"].(string); ok {
			theme.ListItem = val
		}
		if val, ok := optArg["bold"].(string); ok {
			theme.Bold = val
		}
		if val, ok := optArg["italic"].(string); ok {
			theme.Italic = val
		}
		if val, ok := optArg["link"].(string); ok {
			theme.Link = val
		}
		if val, ok := optArg["hr"].(string); ok {
			theme.HR = val
		}
		if val, ok := optArg["reset"].(string); ok {
			theme.Reset = val
		}
	}

	// Build goldmark with ANSI renderer
	// Note: Table extension disabled for ANSI output (tables not well-suited for terminal rendering)
	md := goldmark.New(
		goldmark.WithRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.Prioritized(NewANSIRenderer(theme), 1000),
				),
			),
		),
		goldmark.WithExtensions(
			extension.Strikethrough,
		),
	)

	// Parse and render
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdownText), &buf); err != nil {
		return nil, fmt.Errorf("markdown_ansi() error: %w", err)
	}

	result := buf.String()
	return result, nil
}

// builtinMarkdownText renders markdown to plain text without ANSI codes
func builtinMarkdownText(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get markdown text (required)
	markdownText, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_text() requires markdown text as first argument")
	}

	// Create an empty theme with no ANSI codes
	emptyTheme := &ANSITheme{
		H1:        "",
		H2:        "",
		H3:        "",
		H4:        "",
		H5:        "",
		H6:        "",
		CodeBlock: "",
		CodeInline: "",
		Blockquote: "",
		ListItem:  "",
		Bold:      "",
		Italic:    "",
		Link:      "",
		HR:        "",
		Reset:     "",
	}

	// Build goldmark with ANSI renderer using empty theme
	// Note: Table extension disabled for text output (tables not well-suited for terminal rendering)
	md := goldmark.New(
		goldmark.WithRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.Prioritized(NewANSIRenderer(emptyTheme), 1000),
				),
			),
		),
		goldmark.WithExtensions(
			extension.Strikethrough,
		),
	)

	// Parse and render
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdownText), &buf); err != nil {
		return nil, fmt.Errorf("markdown_text() error: %w", err)
	}

	result := buf.String()
	// Normalize spacing: bullets and numbers followed by extra spaces
	// TODO: debug why this regex may be interfering with bold text
	// re := regexp.MustCompile(`(•|\d+\.)\s+`)
	// result = re.ReplaceAllString(result, "$1 ")

	return result, nil
}
