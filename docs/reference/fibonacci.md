# fibonacci(n)

Compute the nth Fibonacci number efficiently using native Go code. This builtin exists primarily to illustrate Duso's design philosophy: the interpreter is designed to drive a robust runtime, not spin in tight-loop benchmarks. In real applications, Duso makes it easy to build performance-critical functionality in Go and expose it as builtins for excellent performance.

`fibonacci(n)`

## Parameters

- `n` (number) - The index of the Fibonacci number to compute. Must be a non-negative integer.

## Returns

The nth Fibonacci number as a number (int64).

## Examples

Basic usage:

```duso
result = fibonacci(30)
print(result)  // 832040
```

Larger computation:

```duso
result = fibonacci(1000)
print(result)
```

Single call (real-world performance):

```duso
start = timer()
result = fibonacci(10000)
elapsed = (timer() - start) * 1000
print("fibonacci(10000) computed in " + elapsed + "ms")
```

## See Also

- [timer() - Measure elapsed time](/docs/reference/timer.md)
- [Duso Performance Report](/docs/performance-report.md)
