package runtime

import (
	"fmt"

	"github.com/duso-org/duso/pkg/runtime/markdown"
)

// builtinMarkdownHTML renders markdown to HTML.
func builtinMarkdownHTML(evaluator *Evaluator, args map[string]any) (any, error) {
	src, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_html() requires markdown text as first argument")
	}

	opts := markdown.DefaultOptions()

	if optArg, ok := args["1"].(map[string]any); ok {
		if v, ok := optArg["tables"].(bool); ok {
			opts.Tables = v
		}
		if v, ok := optArg["strikethrough"].(bool); ok {
			opts.Strikethrough = v
		}
		if v, ok := optArg["tasklists"].(bool); ok {
			opts.Tasklists = v
		}
		if v, ok := optArg["smartquotes"].(bool); ok {
			opts.Smartquotes = v
		}
	}

	return markdown.ToHTML(src, opts), nil
}

// builtinMarkdownANSI renders markdown to ANSI terminal output.
func builtinMarkdownANSI(evaluator *Evaluator, args map[string]any) (any, error) {
	src, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_ansi() requires markdown text as first argument")
	}

	theme := markdown.DefaultTheme()

	if optArg, ok := args["1"].(map[string]any); ok {
		applyThemeMap(theme, optArg)
	}

	return markdown.ToANSI(src, theme), nil
}

// builtinMarkdownText renders markdown to plain text without ANSI codes.
func builtinMarkdownText(evaluator *Evaluator, args map[string]any) (any, error) {
	src, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_text() requires markdown text as first argument")
	}
	return markdown.ToText(src), nil
}

// applyThemeMap overrides theme fields from a duso option map.
func applyThemeMap(theme *markdown.Theme, m map[string]any) {
	set := func(field *string, key string) {
		if v, ok := m[key].(string); ok {
			*field = v
		}
	}
	set(&theme.H1, "h1")
	set(&theme.H2, "h2")
	set(&theme.H3, "h3")
	set(&theme.H4, "h4")
	set(&theme.H5, "h5")
	set(&theme.H6, "h6")
	set(&theme.CodeBlock, "code_start")
	set(&theme.CodeInline, "code_inline")
	set(&theme.Blockquote, "blockquote")
	set(&theme.ListItem, "list_item")
	set(&theme.Bold, "bold")
	set(&theme.Italic, "italic")
	set(&theme.Link, "link")
	set(&theme.HR, "hr")
	set(&theme.Reset, "reset")
}
