# decode_url()

Decode percent-encoded text from a URL. Matches JavaScript's `decodeURIComponent()`.

`decode_url(string)`

## Parameters

- `string` - percent-encoded text. Must be a string.

## Returns

String with every `%XX` escape replaced by the byte it names, decoded as UTF-8.
Hex digits may be upper or lower case.

## Examples

```duso
decode_url("a%20b")                 // "a b"
decode_url("a%26b%3Dc")             // "a&b=c"
decode_url("caf%C3%A9")             // "café"
decode_url("%E6%97%A5%E6%9C%AC")    // "日本"
```

Round-trips with `encode_url()`:

```duso
decode_url(encode_url("a/b?c#d"))   // "a/b?c#d"
```

## Notes

`+` stays a literal plus:

```duso
decode_url("a+b")   // "a+b"
```

Plus-for-space is a form-encoding convention, not a URL one.
[`parse_form()`](/docs/reference/parse_form.md) applies it where it belongs, when
reading a query string or form body. Reaching for `decode_url()` on a whole query
string is usually a sign that `parse_form()` is the function you want.

A malformed escape - a truncated `%2`, or a `%` followed by non-hex - throws
rather than passing through, because it means the string was mangled upstream:

```duso
try
  decode_url("%2")
catch (e)
  print(e)   // decode_url() failed to decode: ...
end
```

Values arriving through `request().query`, `request().params`, and `parse_form()`
are already decoded. Calling `decode_url()` on them decodes a second time, which
corrupts any value that legitimately contains a `%`.

## See also

- [`encode_url()`](/docs/reference/encode_url.md) - percent-encode a value for a URL
- [`parse_form()`](/docs/reference/parse_form.md) - parse a whole query string or form body into an object
- [`http_server()`](/docs/reference/http_server.md) - where decoded query and path parameters come from
