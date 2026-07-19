# Duso Benchmark Results

Measured 2026-07-19 on duso 1.6 (go 1.26), Intel i9-9880H Mac. Comparisons:
Python 3.13.7, Node v22.20.0, Ruby 2.6.10. Median of 3 runs, each script
self-timing its workload. Peak memory is max RSS via `/usr/bin/time -l`.

## Speed (milliseconds)

| Test | Duso | Python | Node.js | Ruby |
|------|------|--------|---------|------|
| fib(30) iterative x10k | 66 | 8 | 2 | 22 |
| fib(30) builtin x10k (`fib_builtin.du`) | **0.9** | — | — | — |
| loop (1000x1000) | 76 | 97 | 4 | 34 |
| sort (10k) | 1.6 | 1.0 | 5 | 3 |

Duso 1.6's interpreter is 2–4× faster than 1.5. Tight loops now beat CPython;
function-call-heavy code (the fib test is 10,000 calls) still trails it —
calls are the next frontier. Node is a JIT; different sport.

## The builtin escape hatch

| Test (single call) | Duso script | Duso builtin |
|--------------------|-------------|--------------|
| fib(100000) | 21ms | `fibonacci(100000)` — 0.047ms (~450×) |

The `fib_builtin.du` row in the speed table is the same workload as the fib
row — 10,000 fib(30) calls — via the `fibonacci()` builtin: 0.9ms, faster than
Node's JIT running it interpreted.

`fibonacci()` is in the standard build partly for humor, but it makes the
point: duso's answer to "this must run at native speed" is not a JIT — it's
"add it as a Go builtin to your own duso build." A ~450× speedup is one
`RegisterBuiltin` away, and 1.6's fast-path signature (`RegisterBuiltinFast`)
passes your Go function evaluated values with zero marshalling overhead.

The sort row above is the same philosophy already at work: `sort()` is a Go
builtin, which is why duso sorts 10k numbers at native speed while the
interpreted loop tests measure the interpreter itself.

## Memory Baseline (MB)

| Language | Memory |
|----------|--------|
| Duso | 6.5 |
| Python | 8.6 |
| Ruby | 11.4 |
| Node.js | 28.8 |

## Peak Memory During Tests (MB)

| Test | Duso | Python | Node.js | Ruby |
|------|------|--------|---------|------|
| fib(30) | 12.3 | 8.7 | 32.8 | 11.4 |
| loop | 7.3 | 8.6 | 35.3 | 11.5 |
| sort | 13.3 | 9.5 | 35.4 | 12.4 |

Duso remains the smallest resident footprint of the group at rest — the number
that matters most on a small VM running many processes. Peak-during-loop
dropped from 12.5 MB (1.5) to 7.3 MB: the interpreter no longer generates
allocation floods for the GC to chase.
