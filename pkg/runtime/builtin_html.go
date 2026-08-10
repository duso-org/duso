package runtime

import (
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// htmlEscaper replaces all five HTML special characters in a single pass.
//
// A Replacer never rescans what it has already written, so an ampersand becomes
// &amp; without the following entities being escaped a second time. That is the
// property a naive sequence of Replace calls gets wrong.
//
// The double quote becomes &quot; rather than Go's &#34; (see html.EscapeString).
// Both are correct and render identically; &quot; matches what PHP's
// htmlspecialchars and Python's html.escape produce, which is what people expect
// to see when they view source.
var htmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&#39;",
)

// builtinEscapeHTML escapes text for safe interpolation into HTML.
//
// All five special characters are escaped, including both quote styles, so one
// function is safe in element bodies and in quoted attribute values alike:
//
//	<p>{{escape_html(comment)}}</p>
//	<a title="{{escape_html(name)}}">
//
// nil becomes an empty string and non-strings are converted as tostring() would,
// so templates do not need guards around optional fields.
//
// Example:
//
//	escape_html("<script>alert('x')</script>")
//	// &lt;script&gt;alert(&#39;x&#39;)&lt;/script&gt;
func builtinEscapeHTML(evaluator *Evaluator, args map[string]any) (any, error) {
	arg, ok := args["0"]
	if !ok {
		if arg, ok = args["value"]; !ok {
			return "", nil
		}
	}
	return htmlEscaper.Replace(htmlStringify(arg)), nil
}

// htmlStringify renders a value the way tostring() does, except that nil becomes
// an empty string rather than the text "nil" -- interpolating "nil" into a page
// is never what an absent optional field means.
func htmlStringify(arg any) string {
	if arg == nil {
		return ""
	}
	if s, ok := arg.(string); ok {
		return s
	}
	val := InterfaceToValue(arg)
	if val.IsString() {
		return val.AsString()
	}
	if val.Type == script.VAL_NIL {
		return ""
	}
	return script.ValueToDusoString(val)
}
