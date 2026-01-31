# parse_time()

Parse a time string to a Unix timestamp.

## Signature

```duso
parse_time(string [, format])
```

## Parameters

- `string` (string) - Time string to parse
- `format` (optional, string) - Go time format layout. Defaults to RFC3339

## Returns

Unix timestamp as a number

## Examples

Parse RFC3339:

```duso
ts = parse_time("2024-01-23T15:30:45Z")
print(ts)                       // 1705961445
```

Custom format:

```duso
ts = parse_time("2024-01-23", "2006-01-02")
print(ts)                       // timestamp
```

Parse and format:

```duso
input_time = "2024-01-23 14:30:00"
ts = parse_time(input_time, "2006-01-02 15:04:05")
formatted = format_time(ts)
print(formatted)
```

## See Also

- [now() - Get current timestamp](/docs/reference/now.md)
- [format_time() - Format timestamp](/docs/reference/format_time.md)
