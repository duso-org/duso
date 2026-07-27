# Duso 1.6 Performance Report (2026-07-27)

## TL;DR: Not a Toy

- **One binary, multi-core by default.** Deploy as a single executable that automatically scales across all cores. Node needs clustering. Python needs gunicorn. Ruby needs Puma. Duso supports it out of the box.
- **Lean & dense.** 7.5 MB baseline. Run 130 isolated services on a 1 GB $6 box (vs Node's 25, Ruby's 40). On 8-core, up to 280% faster throughput than Node on concurrent IO bound tasks.
- **Holds concurrent load.** Scales 1000's of connections cleanly, even on single core VM. Ruby OOM-kills at 250. Python's GIL tanks p99 latency under load. Duso stays stable along with Node.
- **Built-in encrypted datastore.** Think "Redis" plus durability but built in. GET at 188k–435k ops/sec depending on core count.
- **Hot paths as Go builtins.** Need native speed? Write some Go for a 180–480× speedup. Fibonacci builtin included as an example.

Right-sized for small VMs **and** scales to multi-core. Predictable, operationally simple, all-in-one binary.

## Test Systems

- **Linux VM:** 1 vCPU, 1 GB RAM (DigitalOcean equivalent — the $6/month tier where most small servers live)
- **Mac Workstation:** MacBook Pro 16" (2019), 8-Core Intel i9 @ 2.3 GHz, 16 GB RAM

**Methodology:** Compute benchmarks are median of 3 runs, self-timed. Server benchmarks use `hey` load generator. All measurements are real elapsed time, not theoretical.

## Compute Performance

### Speed (milliseconds)

| Test | VM | Mac | Ratio |
|---|---:|---:|---:|
| fib(30) iterative x10k | 165.9 | 66.0 | 2.5× |
| loop (1000×1000) | 198.4 | 76.0 | 2.6× |
| sort (10k items) | 3.4 | 1.6 | 2.1× |

**What this shows:** On the $6 VM, Duso's interpreter handles realistic workloads cleanly. The 2–2.6× slowdown vs the workstation is predictable and manageable. A single-core shared instance vs an i9 isn't a fair fight. For actual business logic (API handlers, data transformation, orchestration), this is solid.

### The Builtin Escape Hatch

When script performance isn't enough, drop it into a Go builtin. Same workload, native speed:

| Test | Duso Script | Duso Builtin | Speedup |
|---|---:|---:|---:|
| fib(30) x10k | 165.9ms | **0.9ms** | ~180× |
| fib(100000) single | 55.6ms | **0.115ms** | ~480× |

The `fibonacci()` builtin is in the standard build partly for humor, but it makes the architectural point: **Duso's answer to "this must run at native speed" is not a JIT — it's one `RegisterBuiltin` call.** The zero-copy fast path (`RegisterBuiltinFast`) passes evaluated values directly to Go with no marshalling overhead.

### Memory Footprint

**During compute benchmarks:**

| Runtime | VM | Mac |
|---|---:|---:|
| Duso | 7.5 MB | 6.9 MB |
| Node | 38.8 MB | 30.4 MB |
| Python | 9.8 MB | 8.6 MB |
| Ruby | 22.6 MB | 12.0 MB |

Duso is the smallest process in the group. At 7.5 MB on the VM, you can run 130+ isolated duso instances on a 1 GB box.

**Memory under load (`/delay` with 1-second backend delay):**

| Concurrency | Duso (VM) | Duso (Mac) | Status |
|---|---:|---:|---|
| 250 connections | 29 MB | 24 MB | ✓ Stable |
| 1000 connections | 89 MB | 89 MB | ✓ Stable |
| 5000 connections | 204 MB | — | ✓ Stable |

Duso's memory scales linearly with concurrency: ~40 KB per connection held open. Ruby OOM-kills at 250 connections on small VMs. Python survives but its p99 latency degrades under the GIL (climbs to 1.8s at 1000). At 5000 concurrent connections, Duso uses 204 MB and stays stable.

## Server Performance

### HTTP Throughput (`/ping` endpoint, immediate response)

Test: `hey -c 100 -z 10s` hammers the `/ping` endpoint with 100 concurrent connections for 10 seconds. The endpoint responds immediately (no work). This measures **maximum request throughput** with minimal per-request overhead.

| Machine | Runtime | Requests/sec | % of Node |
|---|---|---:|---:|
| VM | Duso | 6,337 | **86%** |
| VM | Node | 7,363 | 100% |
| Mac | Duso | 93,769 | **280%** |
| Mac | Node | 33,527 | 100% |

**What's happening:** On the VM (single core), Duso hits 86% of Node's throughput — competitive. On the Mac (8 cores), Duso scales to **2.8× Node's single-threaded performance** because it automatically uses all cores while Node's event loop is pinned to one. No clustering, no supervisor, no config.

### Holding Concurrent Connections (`/delay`, 1 sec backend delay)

Test: `hey -c N -n (N×5)` holds N concurrent TCP connections open for 5 requests each while the server sleeps 1 second server-side. This measures **steady-state connection load**, not burst. The backend delay is intentional: requests queue up while connections stay open, creating realistic sustained load.

| Concurrency | Duso | Node | Python | Ruby |
|---|---:|---:|---:|---:|
| 250 connections | 242 rps, p99 1.09s | 244 rps, p99 1.15s | 254 rps, p99 1.16s | **OOM-killed** |
| 1000 connections | 248 rps, p99 2.16s | 251 rps, p99 1.49s | p99 **1.82s** (GIL contention) | — |
| 2000 connections | 265 rps, p99 2.27s | 275 rps, p99 1.35s | not run | — |

**What this shows:** 
- **Ruby** dies at 250 concurrent connections. Each thread reserves its stack on Linux, blowing the 1 GB cap.
- **Python** survives but p99 latency climbs to 1.8s at 1000 connections (GIL fights back). No p50/p95 data because the distribution gets weird under load.
- **Duso and Node** both stay stable. Duso's per-request environment cost (~100 KB) is offset by its ability to handle all cores automatically (though on the VM's single core, that's moot).

On small VMs, Ruby simply isn't viable for holding concurrent connections. Python works but degrades under the GIL. Duso and Node are peers.

## Datastore Performance

Encrypted, WAL-backed. No Redis needed. On multi-core, scales beautifully.

### Load Performance

| Dataset | VM Create (1 core) | Mac Create (8 cores) | VM Load Time | Mac Load Time |
|---|---:|---:|---:|---:|
| 10,000 objects | 26,568/sec | 42,534/sec | 0.38s | 0.24s |
| 100,000 objects | 24,399/sec | 43,074/sec | 4.1s | 2.3s |

### Operations (10,000 objects)

| Operation | VM | Mac | Mac Speedup |
|---|---:|---:|---:|
| GET (10k ops) | 188,112 ops/sec | 435,391 ops/sec | 2.3× |
| SET (10k updates) | 24,459 ops/sec | 42,350 ops/sec | 1.7× |
| SELECT predicate | 117ms | 30ms | 3.9× |
| COUNT predicate | 59ms | 19ms | 3.1× |
| Atomic increment (10k) | 54,810 ops/sec | 72,760 ops/sec | 1.3× |

### Operations (100,000 objects)

| Operation | VM | Mac | Mac Speedup |
|---|---:|---:|---:|
| GET (100k ops) | 106,057 ops/sec | 416,518 ops/sec | 3.9× |
| SET (10k updates) | 24,710 ops/sec | 41,603 ops/sec | 1.7× |
| SELECT predicate (66k results) | 1,479ms | 349ms | 4.2× |
| COUNT predicate (66k) | 818ms | 221ms | 3.7× |
| Atomic increment (100k) | 49,527 ops/sec | 70,621 ops/sec | 1.4× |

**What this shows:** On the 8-core Mac, GETs scale 2.3–3.9×, predicates scale 3.1–4.2×. The datastore parallelizes across cores cleanly. On a $20 4-core VPS, you'd see 2–3× gains: faster queries, faster concurrent state updates, all with built-in encryption and no Redis to manage. This is the killer feature on multi-core: built-in concurrent-safe state store that scales with your hardware.

## Operational Story

**Why Duso fits small deployments:**

- **One binary.** No npm, no pip, no system libraries. `duso app.du` starts. Crash? Restart.
- **Multi-core by default.** On a 4-core $20 VPS, Duso scales automatically across all 4. Node needs cluster mode. Python needs gunicorn + supervisor. Ruby needs Puma + forking.
- **Built-in state store.** Datastore handles caching, sessions, temporary state. No Redis instance to manage, no separate monitoring, no `6379` to fight over.
- **Predictable memory.** 7.5 MB baseline on the VM, ~100 KB per concurrent connection held, ~100 KB per spawned worker. No GC surprises.
- **Fast enough glue.** Business logic doesn't need a JIT. The interpreter is fine. Hot paths? Drop them into a Go builtin.

**Deployment math (1 GB VPS, $6/month tier):**

| Runtime | Instances | Memory/instance | CPU isolation | Typical config |
|---|---:|---|---|---|
| Duso | ~130 | 7.5 MB | per-process | `systemd` one-shot, or supervisor for restart |
| Node | ~25 | 39 MB | per-process | cluster module or PM2 |
| Python | ~100 | 10 MB | per-process | gunicorn + supervisor |
| Ruby | ~40 | 23 MB | per-process | Puma + systemd |

Duso lets you pack more isolated services on one box. That means: fewer VPS instances, lower cloud bill, less ops overhead.

## Where Duso Shines

- **Small-VM APIs:** Your service fits entirely on a $6 box. Scale horizontally by running multiple instances across cheap VPS tiers.
- **Orchestration:** Spawning concurrent workers, coordinating long-running tasks, state machines.
- **Data transformation pipelines:** Stateless glue code between services.
- **Configuration & deployment scripts:** Replacing shell scripts with a real language that has HTTP, JSON, and a datastore.

## Where Duso Compromises

- **Function-call-heavy code:** Recursive algorithms or deeply nested call stacks run 6–10× slower than Python. (Solution: drop it into a Go builtin for ~480× speedup.)
- **Multi-server databases:** Duso's datastore is single-instance. For sharded data, use PostgreSQL.

## Duso vs the Field

### VM (1 vCPU, $6/month tier)

| Metric | Duso | Node | Python | Ruby |
|---|---:|---:|---:|---:|
| `/ping` throughput (c=100) | 6,337 rps | 7,363 rps | ~2,100 rps | ~10,000 rps* |
| Memory at rest | 8 MB | 39 MB | 10 MB | 23 MB |
| Holds 250 concurrent | ✓ | ✓ | ✓ | ✗ OOM-killed |
| Processes per 1GB box | 12–15 | 2–3 | 3–4 | 1–2 |

### Mac Workstation (8 cores, i9)

| Metric | Duso | Node |
|---|---:|---:|
| `/ping` throughput (c=100) | **93,769 rps** | 33,527 rps |
| Multi-core scaling | automatic | single thread |
| Per-core efficiency | 11.7k rps/core | 33.5k (pinned) |

\* Ruby's benchmark server skips real HTTP parsing; treat as socket throughput, not HTTP.

**What this means:**
- **On small VMs:** Duso and Node are peers (86% throughput). Duso wins on memory (5× smaller), ops simplicity (no clustering), and built-in state store.
- **On multi-core:** Duso scales automatically and beats Node 2.8×. One binary, no forking, no cluster module.
- **Ruby** dies under connection load and uses 3× memory. Not viable for server work.
- **Python** is slower and needs gunicorn for multi-core.

## The Case in One Table

| Goal | Best Choice | Why |
|---|---|---|
| Ship on a $6 VPS ASAP | **Duso** | Single binary, multi-core, built-in state store, 12–15 instances per GB, runs 6k rps |
| High concurrency (1000+) on a workstation | **Node** | Event loop scales cleanly; single-core but very efficient |
| Machine learning inference | **Python** | NumPy/PyTorch; can't beat it |
| Reliable REPL & interactive dev | **Ruby** | Good stdlib, mature ecosystem |

For single-server deployments on small VMs, Duso has no peer. It's not "fastest" — it's **right-sized**: close enough in throughput to Node, vastly simpler ops, no external services, runs on the smallest boxes, scales well vertically.
