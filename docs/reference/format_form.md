# format_form()

Encode an object as a form-urlencoded string, for query strings and form request bodies.

`format_form(object)`

## Parameters

- `object` (object) - An object whose values are strings, numbers, booleans, or arrays of those

## Returns

A form-urlencoded string (`application/x-www-form-urlencoded`), with keys sorted alphabetically

## Examples

Build a query string:

```duso
params = {client_id = "abc123", redirect_uri = "https://myapp.com/callback"}
print(format_form(params))
// client_id=abc123&redirect_uri=https%3A%2F%2Fmyapp.com%2Fcallback

url = "https://example.com/authorize?" + format_form(params)
```

Send a form-encoded POST body:

```duso
response = fetch("https://api.example.com/token", {
  method = "POST",
  headers = {"Content-Type" = "application/x-www-form-urlencoded"},
  body = format_form({
    grant_type = "authorization_code",
    code = auth_code,
    client_id = client_id
  })
})
```

Arrays write the key once per element:

```duso
format_form({tag = ["js", "web"], q = "search"})
// q=search&tag=js&tag=web
```

Numbers and booleans use their Duso spelling:

```duso
format_form({page = 2, ratio = 1.5, active = true})
// active=true&page=2&ratio=1.5
```

`nil` values are omitted entirely, so optional parameters can be left unset without
building the object conditionally:

```duso
format_form({a = "1", b = nil, c = "3"})
// a=1&c=3
```

## Features

- Percent-encodes reserved characters and UTF-8 (spaces become `+`)
- Keys are sorted, so the same object always produces the same string
- `nil` values are omitted; `""` produces `key=`
- Arrays become repeated keys, which `parse_form()` and `request().query` read back as arrays
- Nested objects throw an error — form encoding has no agreed spelling for them
  (Stripe uses `key.1`, Rails uses `key[]`), so flatten the object first

## See Also

- [parse_form() - Parse form-urlencoded strings](/docs/reference/parse_form.md)
- [fetch() - HTTP client](/docs/reference/fetch.md)
- [format_json() - Format objects to JSON](/docs/reference/format_json.md)
