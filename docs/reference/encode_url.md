# encode_url()

Percent-encode a value so it can be safely placed in a URL. Matches JavaScript's `encodeURIComponent()`.

`encode_url(value)`

## Parameters

- `value` - text to encode. `nil` becomes an empty string, and non-strings are converted the way `tostring()` would, so optional fields need no guard.

## Returns

String with every character outside the unreserved set replaced by its `%XX` UTF-8 byte escapes.

These pass through untouched:

```
A-Z  a-z  0-9  -  _  .  ~  !  *  '  (  )
```

Everything else is escaped, including `& = ? # / : ; , @ $ +` and space. Space
becomes `%20`, not `+`, so the result is legal in a path segment as well as a
query string.

## Examples

Building a URL by hand:

```duso
url = "https://api.example.com/users/" + encode_url(name) + "?q=" + encode_url(term)
```

Each piece is encoded on its own, which is the point - a value containing `&` or
`/` cannot break out into the surrounding URL:

```duso
encode_url("a&b=c")   // "a%26b%3Dc"
encode_url("a/b")     // "a%2Fb"
encode_url("a b")     // "a%20b"
```

Non-ASCII is encoded as UTF-8 bytes:

```duso
encode_url("café")    // "caf%C3%A9"
encode_url("日本")     // "%E6%97%A5%E6%9C%AC"
```

Optional fields and non-strings need no conversion first:

```duso
encode_url(nil)       // ""
encode_url(42)        // "42"
```

## Notes

This encodes one *component* - a single path segment, query value, or fragment -
not a whole URL. Passing a complete URL escapes its `:` and `/` too, which is
almost never what you want:

```duso
encode_url("https://example.com/a")  // "https%3A%2F%2Fexample.com%2Fa"
```

That result is correct when the URL is itself a parameter, as in an OAuth
`redirect_uri`, and wrong everywhere else.

For a whole object of query parameters, [`format_form()`](/docs/reference/format_form.md)
is shorter and handles repeated keys:

```duso
"?" + format_form({q = "a b", page = 2})   // "?page=2&q=a+b"
```

Note that `format_form()` writes space as `+`, the form-encoding convention. Both
spellings decode to a space in a query string; only `%20` is legal in a path.

Output matches `encodeURIComponent()` byte for byte, so a value encoded here
compares equal to the same value encoded in a browser - which matters for OAuth
signatures and cache keys.

## See also

- [`decode_url()`](/docs/reference/decode_url.md) - reverse the encoding
- [`format_form()`](/docs/reference/format_form.md) - encode an object of parameters as a query string
- [`escape_html()`](/docs/reference/escape_html.md) - escape text for HTML, a different context with different rules
- [`fetch()`](/docs/reference/fetch.md) - make the request
