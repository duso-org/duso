# escape_html()

Escape text so it can be safely interpolated into HTML. Converts the five characters that would otherwise be read as markup into their entities.

`escape_html(value)`

## Parameters

- `value` - text to escape. `nil` becomes an empty string, and non-strings are converted the way `tostring()` would, so templates do not need guards around optional fields.

## Returns

String with these replacements applied in a single pass:

| Character | Becomes |
|-----------|---------|
| `&` | `&amp;` |
| `<` | `&lt;` |
| `>` | `&gt;` |
| `"` | `&quot;` |
| `'` | `&#39;` |

Both quote styles are escaped, so one function is safe in element bodies and in
quoted attribute values alike.

Escaping happens in one pass, so the ampersand written for `<` is not escaped a
second time - `escape_html("<")` is `&lt;`, never `&amp;lt;`.

## Examples

Escaping user content in a page:

```duso
comment = "<script>alert('xss')</script>"
print(escape_html(comment))
// &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;
```

Inside a template, in both contexts:

```duso
page = template("""
  <a title="{{escape_html(user.name)}}">
    <p>{{escape_html(user.bio)}}</p>
  </a>
""")

print(page({user = current_user}))
```

Optional fields need no guard, because `nil` yields an empty string:

```duso
escape_html(nil)          // ""
escape_html(user.nickname) // "" when the field is absent
```

Non-strings are converted first:

```duso
escape_html(42)    // "42"
escape_html(true)  // "true"
```

## Notes

Escaping is deliberately not idempotent. Text that already contains an entity is
escaped again, because `escape_html` cannot know whether `&amp;` in its input was
meant as markup or as the literal five characters a user typed:

```duso
escape_html("&amp;")   // "&amp;amp;"
```

Escape once, at the point where text is placed into HTML - not earlier, and not
twice.

This function covers text placed in element bodies and quoted attribute values.
It is not sufficient for text placed inside a `<script>` block, inside a CSS
rule, or in an unquoted attribute, all of which need context-specific escaping
that duso does not provide. Do not interpolate untrusted content into those
positions.

## See also

- [`template()`](/docs/reference/template.md) - build reusable templates with `{{expression}}` interpolation
- [`markdown_html()`](/docs/reference/markdown_html.md) - render markdown to HTML, which escapes as part of rendering
- [`http_server()`](/docs/reference/http_server.md) - serve the resulting HTML
