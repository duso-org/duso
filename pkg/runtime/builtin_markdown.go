package runtime

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

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
		"tasklists":     false,
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

	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
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
	md := goldmark.New(
		goldmark.WithRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.Prioritized(NewANSIRenderer(theme), 1000),
				),
			),
		),
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
		),
	)

	// Parse and render
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdownText), &buf); err != nil {
		return nil, fmt.Errorf("markdown_ansi() error: %w", err)
	}

	return buf.String(), nil
}
