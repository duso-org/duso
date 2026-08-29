package runtime

import (
	"fmt"
	"net/url"
	"strings"
)

// urlEscapeFixups reconciles url.QueryEscape with JavaScript's
// encodeURIComponent, which is the behaviour every web developer already has in
// their head and the one duso spells encode_url().
//
// Go and JavaScript disagree in exactly six places:
//
//   - Space. QueryEscape writes "+", legal only in a query string. JavaScript
//     writes "%20", legal in a path segment too, so one function covers both.
//   - ! ' ( ) *. QueryEscape escapes them; JavaScript leaves them alone. Both
//     decode identically, but matching JavaScript byte for byte means a value
//     encoded here compares equal to the same value encoded in the browser --
//     which matters for OAuth signatures and cache keys.
//
// Rewriting the output is safe because every "%" in escaped text begins an
// escape triple, so a "%21" in the result can only be an escaped "!" -- an input
// of literal "%21" comes back as "%2521" and is left alone.
var urlEscapeFixups = strings.NewReplacer(
	"+", "%20",
	"%21", "!",
	"%27", "'",
	"%28", "(",
	"%29", ")",
	"%2A", "*",
)

// builtinEncodeURL percent-encodes a value for use in a URL, matching
// JavaScript's encodeURIComponent().
//
// It encodes a single component -- one path segment, one query value, one
// fragment -- so every reserved character is escaped, "&" and "=" and "/"
// included. That is what makes it safe to concatenate:
//
//	"https://api.example.com/users/" + encode_url(name) + "?q=" + encode_url(term)
//
// For a whole object of query parameters, format_form() is the shorter path.
//
// nil becomes an empty string and non-strings are converted as tostring() would,
// so optional fields need no guard.
func builtinEncodeURL(evaluator *Evaluator, args map[string]any) (any, error) {
	arg, ok := args["0"]
	if !ok {
		if arg, ok = args["value"]; !ok {
			return "", nil
		}
	}
	return urlEscapeFixups.Replace(url.QueryEscape(escapeStringify(arg))), nil
}

// builtinDecodeURL reverses encode_url(), matching JavaScript's
// decodeURIComponent().
//
// "+" is left as a literal plus rather than decoded to a space: that conversion
// belongs to form encoding, not to URLs generally, and parse_form() already does
// it where it applies. Decoding "a+b" here yields "a+b", as it does in a browser.
//
// A malformed escape is an error rather than a silent pass-through, because a
// truncated "%2" in a URL means the string was mangled upstream and the caller
// should hear about it.
func builtinDecodeURL(evaluator *Evaluator, args map[string]any) (any, error) {
	arg, ok := args["0"]
	if !ok {
		arg, ok = args["value"]
	}
	if !ok {
		return nil, fmt.Errorf("decode_url() requires a string argument")
	}

	input, ok := arg.(string)
	if !ok {
		return nil, fmt.Errorf("decode_url() requires a string argument")
	}

	decoded, err := url.PathUnescape(input)
	if err != nil {
		return nil, fmt.Errorf("decode_url() failed to decode: %v", err)
	}
	return decoded, nil
}
