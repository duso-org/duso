# parse_form()

Parse a form-urlencoded string into an object.

`parse_form(str)`

## Parameters

- `str` (string) - A form-urlencoded string (`application/x-www-form-urlencoded`). A leading `?` is allowed

## Returns

An object mapping each key to its value, or to an array of values if the key appears more than once

## Examples

Parse a query string:

```duso
params = parse_form("code=xyz789&state=abc123")
print(params.code)              // "xyz789"
print(params.state)             // "abc123"
```

A leading `?` is accepted, so a URL's query portion can be passed as-is:

```duso
parse_form("?code=xyz&state=abc")
```

Percent-encoded values are decoded:

```duso
params = parse_form("redirect_uri=https%3A%2F%2Fa.b%2Fcb&q=a+b")
print(params.redirect_uri)      // "https://a.b/cb"
print(params.q)                 // "a b"
```

Repeated keys become an array, the same shape `request().query` returns:

```duso
params = parse_form("tag=js&tag=web")
print(params.tag)               // ["js", "web"]
```

Read a form-encoded API response — some OAuth token endpoints answer this way
instead of JSON:

```duso
response = fetch(token_url, {method = "POST", body = format_form(params)})
token = parse_form(response.body)
print(token.access_token)
```

## Features

- Decodes percent-encoding and `+` as space
- Repeated keys become arrays; single keys stay strings
- A key with no value (`a=`) yields an empty string
- Throws on malformed percent-encoding

Note that handler scripts do not need this for incoming requests: `request().query`,
`request().form`, and `request().cookies` are already parsed objects.

## See Also

- [format_form() - Encode objects as form-urlencoded](/docs/reference/format_form.md)
- [http_server() - Request object properties](/docs/reference/http_server.md)
- [parse_json() - Parse JSON strings](/docs/reference/parse_json.md)
