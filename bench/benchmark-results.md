# Duso 1.6 Benchmark Results

**Test System:** MacBook Pro 16" (2019) — 8-Core Intel Core i9 @ 2.3 GHz, 16 GB RAM

**Measured 2026-07-27** on duso 1.6 (Go 1.26).

**Language comparisons:** Python 3.13.7, Node v22.20.0, Ruby 2.6.10. Each benchmark is the median of 3 runs; tests self-time their workload. Peak memory is max RSS via `/usr/bin/time -l`.

**Datastore benchmarks:** Encrypted, WAL-backed objects at 10,000 and 100,000 scales.

---

## Interpreter Performance

Duso 1.6 delivers a **2–4× speedup** over 1.5. The interpreter is highly optimized for tight loops and common operations, making it a practical choice for production workloads on single-server deployments.

### Speed (milliseconds)

| Test | Duso | Python | Node.js | Ruby |
|------|------|--------|---------|------|
| fib(30) iterative x10k | 66 | 8 | 2 | 22 |
| fib(30) via builtin x10k | **0.9** | — | — | — |
| loop (1000×1000) | 76 | 97 | 4 | 34 |
| sort (10k items) | 1.6 | 1.0 | 5 | 3 |

**What's happening:** Loop performance is now competitive with CPython (within ~20% on this workload). Builtin functions reach native Go speed — see the fib builtin row. Node.js and Ruby run compiled/JIT code, so they excel on compute-heavy tasks, but Duso's tight-loop efficiency makes it suitable for typical web/API workloads.

### Built-in Optimization Strategy

| Test (single call) | Script | Builtin | Speedup |
|--------------------|--------|---------|---------|
| fib(100000) | 21ms | 0.047ms | ~450× |

Duso's philosophy: **fast enough matters more than blazingly fast**. When 1.6's script performance isn't sufficient for a specific operation, drop it into a Go builtin using `RegisterBuiltin` or the zero-copy fast path (`RegisterBuiltinFast`). The fibonacci builtin example shows the ceiling — native Go speed with transparent integration into duso code.

`sort()` is already a builtin, which is why the sort row runs at native speed. Pattern: Script is for glue and control; builtins are for hot loops and heavy lifting.

---

## Memory Footprint

Duso is optimized for small-VM deployments where every megabyte counts. Memory discipline improves significantly in 1.6.

### Baseline Memory (MB)

| Language | Memory |
|----------|--------|
| Duso | 6.5 |
| Python | 8.6 |
| Ruby | 11.4 |
| Node.js | 28.8 |

Duso's resident footprint is the **smallest in the group** — important on machines running many processes (typical small-cloud deployments).

### Peak Memory During Tests (MB)

| Test | Duso | Python | Node.js | Ruby |
|------|------|--------|---------|------|
| fib(30) | 12.3 | 8.7 | 32.8 | 11.4 |
| loop | 7.3 | 8.6 | 35.3 | 11.5 |
| sort | 13.3 | 9.5 | 35.4 | 12.4 |

**Memory improved 1.5 → 1.6:** Peak-during-loop dropped from 12.5 MB to 7.3 MB. The new allocator reduces GC pressure by eliminating allocation floods. Workloads that generate a lot of temporary data see proportional gains.

---

## Datastore Benchmarks

Encrypted, WAL-backed persistence. These benchmarks exercise the real-world cases where duso's datastore proves useful: single-server caching, session stores, and application state.

### Load Performance (encrypted)

| Size | Load Time | Throughput |
|------|-----------|-----------|
| 10,000 objects | 0.24s | 41,202 objects/sec |
| 100,000 objects | 2.4s | 42,064 objects/sec |

### Query & Update Operations (10,000 objects)

| Operation | Time | Throughput |
|-----------|------|-----------|
| GET (10,000 ops) | 22ms | 443,105 ops/sec |
| SET (10,000 updates) | 242ms | 41,239 ops/sec |
| SELECT predicate (6,667 results) | 30ms | — |
| COUNT predicate | 16ms | — |
| SELECT with limit (top 100) | 1ms | — |
| Atomic increment (10,000) | 134ms | 74,262 ops/sec |

### Query & Update Operations (100,000 objects)

| Operation | Time | Throughput |
|-----------|------|-----------|
| GET (100,000 ops) | 238ms | 418,599 ops/sec |
| SET (10,000 updates) | 236ms | 42,245 ops/sec |
| SELECT predicate (66,667 results) | 349ms | — |
| COUNT predicate | 220ms | — |
| SELECT with limit (top 100) | 3ms | — |
| Atomic increment (100,000) | 1,337ms | 74,791 ops/sec |

**What this means:** GET performance is consistent across sizes (~420k ops/sec), even with encryption and WAL-backed persistence. SET operations include the full update cycle (read-modify-write with WAL sync) and scale linearly. Full-table scans (SELECT/COUNT) are linear in data size, completing in milliseconds for 100k objects — suitable for typical queries and aggregations. Atomic operations maintain lock-free performance at 74k ops/sec, suitable for counters and distributed coordination patterns.

---

## Context & Trade-offs

**Why Duso's approach works:**

- **Baseline speed:** Good enough for business logic, API glue, data transformation, and orchestration. The interpreter doesn't need to be "fastest" — it needs to be predictable and lean.
- **Builtins as escape hatch:** Hot paths get a native Go fast path. This is more maintainable than a bytecode VM or JIT, and easier to profile/debug in production.
- **Memory:** Small VMs (512MB–2GB) can run many duso processes concurrently, each with its own isolated state. No shared memory coordination overhead.
- **Datastore:** Built-in, encrypted, WAL-backed, no external dependencies. Good for cache warming, distributed state machines, and temporary queues. Not a replacement for PostgreSQL, but eliminates the need for Redis in many single-server topologies.

**Not a goal:** Beating Node.js/CPython on pure compute. Duso targets the deployment profile (small VMs, minimal ops burden, all-in-one binary) rather than raw performance.
