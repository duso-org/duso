# now()

Get the current Unix timestamp.

## Signature

```duso
now()
```

## Parameters

None

## Returns

Current Unix timestamp as a number (seconds since epoch)

## Examples

Get current time:

```duso
current = now()
print(current)                  // 1674567890 (example)
```

Measure elapsed time:

```duso
start = now()
// do something
end = now()
elapsed = end - start
print("Took " + elapsed + " seconds")
```

Timestamp data:

```duso
event = {
  name = "login",
  timestamp = now()
}
```

## See Also

- [format_time() - Format timestamp](/docs/reference/format_time.md)
- [parse_time() - Parse timestamp](/docs/reference/parse_time.md)
