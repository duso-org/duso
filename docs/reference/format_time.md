# format_time()

Format a Unix timestamp to a human-readable string.

## Signature

```duso
format_time(timestamp [, format])
```

## Parameters

- `timestamp` (number) - Unix timestamp
- `format` (optional, string) - Go time format string. Defaults to RFC3339

## Returns

Formatted timestamp string

## Examples

Default format (RFC3339):

```duso
ts = now()
formatted = format_time(ts)
print(formatted)                // "2024-01-23T15:30:45Z"
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

- [now() - Get current timestamp](./now.md)
- [parse_time() - Parse timestamp](./parse_time.md)
- [Date/Time functions](../language-spec.md#datetime-functions)
