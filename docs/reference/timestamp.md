# timestamp()

Get the current Unix timestamp in UTC or a specific timezone/offset.

## Signature

```duso
timestamp()
timestamp(timezone)
```

## Parameters

- `timezone` (optional): IANA timezone name (e.g., `"America/New_York"`, `"Europe/London"`) or fixed offset (e.g., `"+7"`, `"-5:30"`)

## Returns

Current Unix timestamp as a number (seconds since epoch). 

When called without arguments, returns the current UTC time. When called with a timezone, returns the local time in that timezone as a timestamp.

## Examples

Get current UTC time:

```duso
utc = timestamp()
print(utc)  // e.g., 1675200000
```

Get current time in New York (as a local timestamp):

```duso
ny = timestamp("America/New_York")
print(ny)   // 4 hours earlier than UTC time
```

Get current time with a fixed offset:

```duso
bangkok = timestamp("+7")      // UTC+7
mumbai = timestamp("+05:30")   // UTC+5:30
```

Server logging with timezone:

```duso
utc_ts = timestamp()
local_ts = timestamp("America/Los_Angeles")
formatted = format_time(local_ts, "2006-01-02 15:04:05")
print("Event at: " + formatted)
```

## Notes

Unix timestamps are absolute points in time. When you pass a timezone to `timestamp()`, it returns the local time value as a timestamp (offset-adjusted), which is useful for server applications that need to log or process times in specific timezones.

## See Also

- [now() - Get current local time](/docs/reference/now.md)
- [timer() - Get high-precision time for benchmarking](/docs/reference/timer.md)
- [format_time() - Format timestamp](/docs/reference/format_time.md)
