# Duso Benchmark Results

## Speed (milliseconds)

| Test | Duso | Python | Node.js | Ruby |
|------|------|--------|---------|------|
| fib(30) | 1,310 | 125 | 20 | 85 |
| loop (1000x1000) | 233 | 103 | 6 | 37 |
| sort (10k) | 2 | 1 | 4 | 2 |

## Memory Baseline (MB)

| Language | Memory |
|----------|--------|
| Duso | 6.7 |
| Python | 8.7 |
| Ruby | 12.3 |
| Node.js | 30.3 |

## Peak Memory During Tests (MB)

| Test | Duso | Python | Node.js | Ruby |
|------|------|--------|---------|------|
| fib(30) | 14.2 | 8.8 | 33.1 | 12.3 |
| loop | 12.5 | 11.4 | 35.1 | 12.3 |
| sort | 12.2 | 12.5 | 34.1 | - |
