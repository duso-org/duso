# timer()

Get the current time with sub-second precision for benchmarking and performance measurement.

## Signature

```duso
timer()
```

## Parameters

None

## Returns

Current time as a number (seconds since epoch) with fractional/decimal precision for measuring elapsed time.

## Examples

Simple elapsed time measurement:

```duso
start = timer()
// do some work
elapsed = timer() - start
print("Took " + elapsed + " seconds")
```

Benchmark a function:

```duso
function do_work()
  // some expensive operation
end

iterations = 1000
start = timer()
for i = 0, iterations do
  do_work()
end
elapsed = timer() - start

avg = elapsed / iterations
print("Average per iteration: " + avg + " seconds")
```

Measure with millisecond precision:

```duso
start = timer()
// do something
elapsed_ms = (timer() - start) * 1000
print("Took " + elapsed_ms + " ms")
```

## Notes

`timer()` is optimized for measuring short intervals and provides decimal/fractional seconds. For getting the current timestamp for logging or data purposes, use `now()` or `timestamp()` instead.

## See Also

- [now() - Get current local time](/docs/reference/now.md)
- [timestamp() - Get current time in any timezone](/docs/reference/timestamp.md)
