package runtime

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// formFieldValue renders one form field. Objects and other non-scalar values
// have no agreed spelling in form encoding — Stripe writes "key.1", Rails writes
// "key[]" — so rather than silently pick one and produce a body the server reads
// wrong, say so and let the caller flatten it.
func formFieldValue(key string, v any) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64, bool:
		return script.InterfaceToValue(val).String(), nil
	case map[string]any:
		return "", fmt.Errorf("format_form() cannot encode nested object at key %q — flatten it first", key)
	case *script.ValueRef:
		return "", fmt.Errorf("format_form() cannot encode %s at key %q", val.Val.Type.String(), key)
	default:
		return "", fmt.Errorf("format_form() cannot encode value at key %q", key)
	}
}

// builtinFormatForm encodes an object as an application/x-www-form-urlencoded
// string, for query strings and form request bodies.
func builtinFormatForm(evaluator *Evaluator, args map[string]any) (any, error) {
	objectArg := GetArg(args, 0, "object")
	if objectArg == nil {
		return nil, fmt.Errorf("format_form() requires an object as first argument")
	}

	object, ok := objectArg.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("format_form() requires an object as first argument")
	}

	values := url.Values{}
	for key, v := range object {
		// nil means "not provided" — omit the key entirely rather than send it
		// empty, since "key=" and a missing key mean different things to most APIs.
		if v == nil {
			continue
		}

		// An array writes the key once per element, matching how parse_form()
		// and request().query read repeated keys back.
		if arr, ok := v.(*[]script.Value); ok {
			for _, item := range *arr {
				if item.IsNil() {
					continue
				}
				values.Add(key, item.String())
			}
			continue
		}

		field, err := formFieldValue(key, v)
		if err != nil {
			return nil, err
		}
		values.Add(key, field)
	}

	// Encode() sorts by key, so the same object always produces the same string.
	return values.Encode(), nil
}

// builtinParseForm decodes an application/x-www-form-urlencoded string into an
// object.
func builtinParseForm(evaluator *Evaluator, args map[string]any) (any, error) {
	stringArg := GetArg(args, 0, "str")
	if stringArg == nil {
		return nil, fmt.Errorf("parse_form() requires a string as first argument")
	}

	formString, ok := stringArg.(string)
	if !ok {
		return nil, fmt.Errorf("parse_form() requires a string as first argument")
	}

	// Accept a whole query string as it appears in a URL, so callers can hand
	// over a callback URL's tail without trimming it first.
	formString = strings.TrimPrefix(formString, "?")

	parsed, err := url.ParseQuery(formString)
	if err != nil {
		return nil, fmt.Errorf("parse_form() error: %v", err)
	}

	// multiValued gives a repeated key an array and a single key a string, the
	// same shape request().query and request().cookies produce.
	fields := make(map[string]script.Value, len(parsed))
	for k, vv := range parsed {
		fields[k] = multiValued(vv)
	}

	return &script.ValueRef{Val: script.NewObject(fields)}, nil
}
