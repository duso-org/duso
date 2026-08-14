# format_time()

Format a Unix timestamp to a human-readable string.

`format_time(timestamp [, format])`

## Parameters

- `timestamp` (number) - Unix timestamp
- `format` (optional, string) - Go time layout, or a named format. Defaults to `"2006-01-02 15:04:05"`

## Returns

Formatted timestamp string

## Format strings

Layouts are Go-style — you spell out the reference time
`Mon Jan 2 15:04:05 MST 2006` in the shape you want. This is the preferred
style.

These named shorthands are also accepted:

| Name | Layout | Example |
|---|---|---|
| `iso` | `2006-01-02T15:04:05Z` | `2024-01-23T15:30:45Z` |
| `date` | `2006-01-02` | `2024-01-23` |
| `time` | `15:04:05` | `15:30:45` |
| `long_date` | `January 2, 2006` | `January 23, 2024` |
| `long_date_dow` | `Mon January 2, 2006` | `Tue January 23, 2024` |
| `short_date` | `Jan 2, 2006` | `Jan 23, 2024` |
| `short_date_dow` | `Mon Jan 2, 2006` | `Tue Jan 23, 2024` |

For compatibility with habits from other languages, the placeholders `YYYY`,
`YY`, `MM`, `DD`, `HH`, `mm`, and `ss` are translated to their Go equivalents,
so `"YYYY-MM-DD"` is identical to `"2006-01-02"`. Prefer Go layouts in new
code — that placeholder list is the whole set, and anything outside it is
passed through to Go untouched.

## Examples

Default format:

```duso
ts = now()
formatted = format_time(ts)
print(formatted)                // "2024-01-23 15:30:45"
```

Custom format:

```duso
ts = now()
// Use Go time format layout
formatted = format_time(ts, "2006-01-02 15:04:05")
print(formatted)                // "2024-01-23 15:30:45"
```

Simple date:

```duso
ts = now()
date_only = format_time(ts, "2006-01-02")
print(date_only)                // "2024-01-23"
```

## See Also

- [now() - Get current timestamp](/docs/reference/now.md)
- [parse_time() - Parse timestamp](/docs/reference/parse_time.md)
