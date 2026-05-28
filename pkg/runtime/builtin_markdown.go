package runtime

import (
	"fmt"

	"github.com/duso-org/duso/pkg/runtime/markdown"
	"github.com/duso-org/duso/pkg/script"
)

// builtinMarkdownHTML renders markdown to HTML.
func builtinMarkdownHTML(evaluator *Evaluator, args map[string]any) (any, error) {
	src, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_html() requires markdown text as first argument")
	}

	opts := markdown.DefaultOptions()

	// Support both named argument (options=...) and positional argument
	var optArg map[string]any
	if opt, ok := args["options"].(map[string]any); ok {
		optArg = opt
	} else if rawOpt, ok := args["options"]; ok {
		if val, ok := rawOpt.(script.Value); ok && val.IsObject() {
			scriptMap := val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		} else if val, ok := rawOpt.(*script.ValueRef); ok && val.Val.IsObject() {
			scriptMap := val.Val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		}
	} else if opt, ok := args["1"].(map[string]any); ok {
		optArg = opt
	} else if rawOpt, ok := args["1"]; ok {
		if val, ok := rawOpt.(script.Value); ok && val.IsObject() {
			scriptMap := val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		} else if val, ok := rawOpt.(*script.ValueRef); ok && val.Val.IsObject() {
			scriptMap := val.Val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		}
	}

	applyOptionsMap(&opts, optArg)

	return markdown.ToHTML(src, opts), nil
}

// builtinMarkdownANSI renders markdown to ANSI terminal output.
func builtinMarkdownANSI(evaluator *Evaluator, args map[string]any) (any, error) {
	src, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_ansi() requires markdown text as first argument")
	}

	opts := markdown.DefaultOptions()
	theme := markdown.DefaultTheme()

	// Support both named argument (options=...) and positional argument
	var optArg map[string]any
	if opt, ok := args["options"].(map[string]any); ok {
		optArg = opt
	} else if rawOpt, ok := args["options"]; ok {
		if val, ok := rawOpt.(script.Value); ok && val.IsObject() {
			scriptMap := val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		} else if val, ok := rawOpt.(*script.ValueRef); ok && val.Val.IsObject() {
			scriptMap := val.Val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		}
	} else if opt, ok := args["1"].(map[string]any); ok {
		optArg = opt
	} else if rawOpt, ok := args["1"]; ok {
		if val, ok := rawOpt.(script.Value); ok && val.IsObject() {
			scriptMap := val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		} else if val, ok := rawOpt.(*script.ValueRef); ok && val.Val.IsObject() {
			scriptMap := val.Val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		}
	}

	applyOptionsMap(&opts, optArg)
	applyThemeMap(theme, optArg)

	return markdown.ToANSI(src, opts, theme), nil
}

// builtinMarkdownText renders markdown to plain text without ANSI codes.
func builtinMarkdownText(evaluator *Evaluator, args map[string]any) (any, error) {
	src, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown_text() requires markdown text as first argument")
	}

	opts := markdown.DefaultOptions()

	// Support both named argument (options=...) and positional argument
	var optArg map[string]any
	if opt, ok := args["options"].(map[string]any); ok {
		optArg = opt
	} else if rawOpt, ok := args["options"]; ok {
		if val, ok := rawOpt.(script.Value); ok && val.IsObject() {
			scriptMap := val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		} else if val, ok := rawOpt.(*script.ValueRef); ok && val.Val.IsObject() {
			scriptMap := val.Val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		}
	} else if opt, ok := args["1"].(map[string]any); ok {
		optArg = opt
	} else if rawOpt, ok := args["1"]; ok {
		if val, ok := rawOpt.(script.Value); ok && val.IsObject() {
			scriptMap := val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		} else if val, ok := rawOpt.(*script.ValueRef); ok && val.Val.IsObject() {
			scriptMap := val.Val.AsObject()
			optArg = make(map[string]any)
			for k, v := range scriptMap {
				optArg[k] = v
			}
		}
	}

	if optArg != nil {
		if v, ok := optArg["tables"].(bool); ok {
			opts.Tables = v
		}
		if v, ok := optArg["strikethrough"].(bool); ok {
			opts.Strikethrough = v
		}
		if v, ok := optArg["highlight"].(bool); ok {
			opts.Highlight = v
		}
		if v, ok := optArg["tasklists"].(bool); ok {
			opts.Tasklists = v
		}
		if v, ok := optArg["code_language"].(bool); ok {
			opts.CodeLanguage = v
		}
	}

	return markdown.ToText(src, opts), nil
}

// applyOptionsMap applies option flags from a duso option map to the Options struct.
func applyOptionsMap(opts *markdown.Options, m map[string]any) {
	if m == nil {
		return
	}
	if v, ok := m["tables"].(bool); ok {
		opts.Tables = v
	}
	if v, ok := m["strikethrough"].(bool); ok {
		opts.Strikethrough = v
	}
	if v, ok := m["highlight"].(bool); ok {
		opts.Highlight = v
	}
	if v, ok := m["tasklists"].(bool); ok {
		opts.Tasklists = v
	}
	if v, ok := m["smartquotes"].(bool); ok {
		opts.Smartquotes = v
	}
	if v, ok := m["code_language"].(bool); ok {
		opts.CodeLanguage = v
	}
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
	set(&theme.Strike, "strike")
	set(&theme.Highlight, "highlight")
	set(&theme.Link, "link")
	set(&theme.HR, "hr")
	set(&theme.Reset, "reset")
}
