# Duso Runtime Performance

A cross-language comparison of Duso 1.6 against Node.js, Python, and Ruby on two machines: a multi-core workstation and the small cloud VM Duso is designed for.

## Setup

Two test machines:

- **VM**: 1 vCPU / 961 MiB RAM Ubuntu 24.04 cloud instance (DigitalOcean, the $6/month tier most small servers actually deploy to). Duso v1.6.0, Node v20.19.6, Python 3.12.3, Ruby 3.2.3.
- **Workstation**: Intel i9-9880H Mac (8 cores / 16 threads, 16 GB). Duso v1.6.0, Node v22.20.0, Python 3.13.7, Ruby 2.6.10 (the macOS system Ruby -- older than the VM's, noted where it matters).

Methodology:

- Compute benchmarks are the median of 3 runs, each script self-timing its workload.
- Server benchmarks use [hey](https://github.com/rakyll/hey), a neutral Go load generator, against a minimal stdlib-only HTTP server written in each language. No language benchmarks against another language's server.
- Orchestration benchmarks pair each language's concurrent-fetch client with its own delay server, so each stack carries its full cost.
- Peak memory is max RSS (`/usr/bin/time -l` on macOS, `/usr/bin/time -v` and `/proc/<pid>/VmHWM` on Linux). On the VM, benchmark processes ran inside memory-capped cgroups (200 MB server / 180 MB client) so a runaway runtime kills the benchmark, not the box. "OOM-killed" rows mean the runtime blew that cap.
- Delay endpoints respond after 1 second server-side on localhost, eliminating internet variance. TLS is not in the loop.

All scripts live in `/bench` at the project root, one file per language. Raw numbers for graphing are in `bench/results-basics.csv`, `bench/results-server.csv`, and `bench/results-client.csv`.

## Single-threaded compute

Pure CPU in script code, no I/O. Milliseconds, lower is better.

| benchmark | machine | Duso | Python | Node | Ruby |
| --- | --- | --- | --- | --- | --- |
| fib(30) x 10,000 calls | VM | 113.7 | 11.4 | **4** | 59 |
| fib(30) x 10,000 calls | Mac | 68.5 | 9.2 | **3** | 24 |
| nested loop 1000x1000 | VM | 166.9 | 170.5 | **10** | 82 |
| nested loop 1000x1000 | Mac | 82.8 | 104.7 | **5** | 37 |
| sort 10,000 floats | VM | 3.9 | **1.7** | 9 | 4 |
| sort 10,000 floats | Mac | 1.7 | **1.3** | 6 | 2 |

The 1.6 interpreter is a large step up from the numbers in earlier versions of this report: on the same VM class, the nested loop went from 447 ms (v1.1) to 167 ms, and **Duso now edges out CPython on straight-line loop code on both machines**. Function-call-heavy code (fib is 10,000 script-level calls) still trails CPython -- calls are the next frontier. Node is a four-tier optimizing JIT; different sport, listed for scale.

`sort()` runs at native speed because it is a Go builtin. That distinction -- script loops vs Go primitives -- is the lead indicator for everything below.

## The builtin escape hatch

Duso 1.6 ships a `fibonacci()` builtin, included partly for humor, but it makes the architectural point precisely:

| workload | Duso script | Duso builtin |
| --- | --- | --- |
| fib(30) x 10,000 (Mac) | 68.5 ms | **1.05 ms** |
| fib(30) x 10,000 (VM) | 113.7 ms | **2.9 ms** |
| single fib(100000) (Mac) | 22.3 ms | **0.043 ms** (~500x) |
| single fib(100000) (VM) | 55.6 ms | **0.115 ms** (~480x) |

The 10,000-call builtin run beats Node's JIT executing the same workload as interpreted JavaScript. Duso's answer to "this hot path must run at native speed" is not a JIT -- it's "add it as a Go builtin to your own Duso build." A ~500x speedup is one `RegisterBuiltin` away, and 1.6's `RegisterBuiltinFast` signature passes your Go function evaluated values with zero marshalling overhead.

## Server benchmarks

Each language serves two endpoints from a minimal stdlib-only server: `/delay` (respond after 1 second -- can you *hold* concurrent connections?) and `/ping` (respond immediately -- raw request throughput). The load generator is `hey`, identical for all four.

The servers are what a developer would naturally write with no dependencies: Duso `http_server()`, Node `http.createServer`, Python `ThreadingHTTPServer` (thread per connection), Ruby a hand-rolled `TCPServer` + thread per connection (webrick left the stdlib in Ruby 3). One caveat on that Ruby server: it skips real HTTP parsing entirely, which flatters its `/ping` throughput below -- read its rows accordingly.

### Holding concurrent connections (`/delay`, 5 requests per connection)

| machine | concurrency | metric | Duso | Node | Python | Ruby |
| --- | --- | --- | --- | --- | --- | --- |
| VM | 250 | p99 (ms) | 1271 | 1195 | **1162** | OOM-killed |
| VM | 1000 | throughput (rps) | 650 | **844** | 792 | OOM-killed |
| VM | 1000 | p99 (ms) | 2156 | **1485** | 1829 | OOM-killed |
| VM | 1000 | server RSS (MB) | 125 | **65** | 50 | OOM-killed |
| Mac | 2000 | throughput (rps) | 1850 | **1854** | 1385 | 1410 |
| Mac | 2000 | p99 (ms) | **1270** | 1353 | 3037 | 2842 |
| Mac | 2000 | server RSS (MB) | 482 | 103 | **78** | 75 |

The headline failures: **Ruby's server was OOM-killed at 250 concurrent connections on the VM** -- Linux commits its per-thread stacks, so 100 held connections already cost 126 MB and 250 blew the cap. Python survives by thread-stack lazy allocation but its p99 degrades steadily (2.4-3 s at high concurrency on the Mac).

The honest negative: at these connection counts each in-flight Duso request holds a live interpreter environment (~225 KB), so Duso's `/delay` RSS grows faster than Node's event loop, which holds a closure and a timer per pending request. We verified this is live data, not GC laziness: capping the Go heap (`GOMEMLIMIT=64MiB`) left RSS unchanged and cost 24% throughput, and rebuilding without Go 1.26's Green Tea GC moved RSS only ~7%. Shrinking the per-request environment is on the 1.6.x list.

### Raw throughput (`/ping`, 10-second run)

| machine | concurrency | metric | Duso | Node | Python | Ruby* |
| --- | --- | --- | --- | --- | --- | --- |
| VM | 100 | rps | 1136 | **8704** | 2092 | 10097* |
| VM | 250 | rps | 1263 | **6741** | 2029 | OOM-killed |
| VM | 250 | p99 (ms) | 338 | **95** | 492 | -- |
| VM | 500 | rps | -- | **7402** | -- | -- |
| VM | 500 | p99 (ms) | -- | **121** | -- | -- |
| Mac | 100 | rps | 13368 | **33527** | 8690 | 46455* |
| Mac | 500 | rps | 15670 | (artifact†) | (artifact†) | 18899* |
| Mac | 500 | p99 (ms) | 220 | (artifact†) | (artifact†) | 31* |

\* Ruby's hand-rolled server does no real HTTP parsing; treat its rps as a socket-loop number, not an HTTP server number. It still dies at 250 connections on the VM.

#### Update: patched 1.6.x interpreter (2026-07-19, `/ping` only)

A follow-up optimization pass attacked the per-request cost identified above: the
parse-cache freshness stat (was one `stat(2)` per request, now at most one per
second), a per-execution fast path for reaching request context (was a
`runtime.Stack` parse plus globally-locked map lookups per builtin call, ~3.4 µs
each), removal of a redundant per-request context registration, and a
server-side GC target suited to small heaps (`GOGC` env still wins). Same
protocol, `hey` c=100 for 10 s, duso and node measured in the same session:

| machine | metric | Duso (patched) | Node (same session) | prior Duso |
| --- | --- | --- | --- | --- |
| VM | rps | 5,889 | 6,589 | 1,094 |
| VM | p99 (ms) | 33 | 34 | 403 |
| Mac, 1 core (`GOMAXPROCS=1`) | rps | 22,873 | 32,465 | -- |
| Mac, all cores | rps | 80,704 | 32,465 | 13,368 |

Two notes on reading this table. First, the VM is a shared instance and
absolute throughput drifts with neighbor load (node measured 8,704 rps in the
original session, 6,589 in this one against the same binary) -- only same-session
ratios are meaningful, which is why both columns were re-measured together.
Second, the headline above ("the event loop wins raw throughput on any core
count") no longer holds on the workstation: duso's zero-config multi-core
scaling now serves ~2.5x node's single event loop on trivial handlers, and the
single-core gap has closed from ~7x to 1.1x on the VM (89% of node) and 1.4x on
the Mac.

The `/delay`, memory, and orchestration tables in this report still reflect
v1.6.0; the full suite has not yet been re-run against the patched build. The
per-request allocation figure (~11 KB transient garbage per request remains
after these patches) suggests the `/delay` RSS numbers will move too.

† Initial Mac c=500 runs showed Node and Python collapsing (6,008 rps with a 6.1 s p99, and worse). Investigation traced this to a macOS measurement artifact, not runtime saturation: macOS caps the listen backlog at `kern.ipc.somaxconn` (128 by default), hey opens all 500 connections in one burst, and overflowed SYNs retransmit on exponential backoff, gating throughput and poisoning the tail. A Linux re-run at c=500 shows Node perfectly healthy (7,402 rps, p99 121 ms). Go drains its accept queue fast enough to sidestep the artifact, so Duso's c=500 numbers are real -- but they should not be read as "Node collapses at 500 connections." The affected rows are flagged in `results-server.csv`. The same caution applies to the Mac `/delay` p99 tails for Python and Ruby at c=2000, which faced the same connection burst against the same 128-slot backlog.

What the clean data actually says:

- **On trivial handlers, the event loop wins raw throughput -- on any core count.** Node's single V8 thread serves ~7x more no-op requests than Duso on the VM, and 2.5x more on the Mac at c=100. The cost is not interpretation: Duso parses scripts once and caches the AST, and walking a five-statement handler costs microseconds (the compute benchmarks put interpreted operations at ~80 ns each). The ~0.9 ms per request is environment construction and request/response plumbing -- the same allocation path behind the ~225 KB per-in-flight-request figure above. One root cause, two symptoms, one 1.6.x target. `/ping` -- a handler that does literally nothing -- is the maximally unflattering case for that fixed cost.
- **Duso's throughput is CPU-bound across all cores, and scales with them.** 15,670 rps at c=500 on the Mac is 16 threads x ~0.9 ms/request, with a steady 220 ms p99; the same interpreter on 1 vCPU manages 1,136. That happens with zero configuration -- no cluster mode, no process supervisor. Python and Ruby need gunicorn/Puma-style forking to use a second core at all.
- **The multi-core payoff is proportional to per-request work.** For handlers that do nothing, Node's one fast thread beats Duso's sixteen interpreted ones. As per-request work grows (datastore access, JSON, templates, native primitives), Node's total handler CPU stays capped at one core while Duso's grows with the machine -- and Duso's per-request work runs in native Go builtins. This report deliberately measures the no-work extreme; real handlers sit between it and the compute benchmarks above.

## Concurrent orchestration

The other direction: the language as the API *consumer*. Spawn N workers, each making 5 sequential requests to its own language's delay server -- Duso `spawn()` (goroutines), Node `Promise.all` (event loop), Python `ThreadPoolExecutor` (OS threads + GIL), Ruby `Thread.new` (OS threads + GVL). The theoretical floor is 5.0 seconds; everything above it is what concurrency actually costs. Both processes -- client and server -- run in the language being measured.

### VM (1 vCPU, 961 MB)

| workers | metric | Duso | Node | Python | Ruby |
| --- | --- | --- | --- | --- | --- |
| 100 | wall (s) | **5.3** | 5.3 | 5.7 | 5.4 |
| 100 | client+server RSS (MB) | **43** | 100 | 59 | 260 |
| 250 | wall (s) | 5.7 | **5.6** | 5.9 | client OOM-killed |
| 500 | wall (s) | 6.3 | **6.2** | 20.4 | client OOM-killed |
| 1000 | wall (s) | 8.2 | **6.9** | 58.7 | client OOM-killed |
| 1000 | client+server RSS (MB) | 221 | **132** | 117 | -- |

### Mac (16 threads, 16 GB)

| workers | metric | Duso | Node | Python | Ruby |
| --- | --- | --- | --- | --- | --- |
| 500 | wall (s) | **5.1** | 5.2 | 6.0 | 5.5 |
| 1000 | wall (s) | **5.2** | 5.3 | 268.0 | 6.1 |
| 2000 | wall (s) | **5.3** | 5.5 | not run | 8.2 |
| 2000 | client+server RSS (MB) | 504 | **251** | -- | 182 |

The cliffs:

- **Python falls off a scheduling cliff, not a memory cliff.** At 500 workers on the VM it takes 4x the floor; at 1000, 11.7x. On the Mac at 1000 workers -- 2000 real OS threads fighting the GIL -- it took **268 seconds against a 5-second floor and visibly pinned all 16 threads of the workstation for four and a half minutes**. Duso ran the same workload in 5.2 s; you wouldn't know it was running.
- **Ruby falls off a memory cliff.** 100 workers already cost 260 MB of combined RSS on the VM; at 250 the client blew its 180 MB cap and was killed. On a fresh 1 GB box without caps, this is the workload that invites the kernel OOM killer.
- **Duso and Node both just work**, at every level, on both machines. Node's combined footprint is smaller at high N (one event loop, no per-worker state); Duso's per-worker cost is ~100 KB (a live interpreter environment per spawned worker). Duso's code is also the simplest of the four -- `spawn()` a script, `store.wait()` for completion, no async/await coloring, no thread pool sizing.

## Memory footprint

Baseline resident memory, trivial script, both machines:

| runtime | VM (MB) | Mac (MB) |
| --- | --- | --- |
| **Duso** | **8.0** | **6.6** |
| Python | 9.8 | 8.6 |
| Ruby | 22.6 | 12.0 |
| Node | 38.8 | 30.4 |

Duso remains the smallest process at rest -- the number that matters most when a small VM runs many processes. During the compute benchmarks Duso peaked at 7-14 MB while Node sat at 33-45 MB.

The fuller picture from the concurrency data: Duso's footprint is the flattest at rest and at moderate concurrency (100-250 in-flight operations, its design center), while at 1000+ simultaneous in-flight operations Node's single-event-loop model is leaner. The per-unit costs, measured: Duso ~100 KB per spawned worker and ~225 KB per held server request; Node ~70 KB per pending promise; Python ~8 MB per thread *reserved* but lazily committed (its problem is the GIL, not RSS); Ruby ~1 MB+ per thread actually committed on Linux (its problem is both).

A note for readers comparing against earlier Duso reports: older numbers (24-29 MB at high worker counts) were sampled coarsely against real-internet endpoints, where latency staggering kept peak overlap low. Today's numbers are true peak RSS with every worker simultaneously live on localhost -- a stricter measure.

## Why an AST-walking interpreter holds this ground

Duso is a tree-walking AST interpreter -- the textbook-slowest interpreter design, with no bytecode and no JIT -- competing above its class:

- **The primitives are Go.** `sort()`, JSON, regex (RE2), HTTP client and server, the datastore, templates -- all native. Script code is glue between them, and the glue is rarely the bottleneck in a real application. Where it is, the `fibonacci()` experiment shows the escape hatch: a Go builtin is ~500x faster than script and one registration away.
- **The scheduler is Go's.** `spawn()` rides the goroutine scheduler -- M:N, work-stealing, all cores, no configuration. That is why 2000-way concurrency costs 0.3 s of overhead and why server throughput scales 14x from 1 vCPU to 16 threads without touching a config file. Duso didn't build a scheduler; it inherited a decade of Google's.
- **The trade is explicit.** Per-request and per-worker environment state costs real memory at extreme concurrency, and its construction costs real CPU (~0.9 ms/request) that an event loop's closure-per-request model avoids. Scripts themselves are parsed once and cached as ASTs -- the cost is env setup, not interpretation -- which is why it is one concrete optimization target rather than an architectural ceiling.

## What the data argues

| dimension | result |
| --- | --- |
| Tight script loops | Duso 1.6 beats CPython; trails JIT'd Node; rarely matters (glue code) |
| Function-call-heavy script | Trails CPython ~6-10x; next optimization frontier |
| Native primitives (sort, JSON, HTTP, datastore) | Go speed; `fibonacci()` shows the ~500x builtin path |
| Trivial-handler HTTP throughput | Node's event loop wins (7x on one core, 2.5x on sixteen); Duso's ~0.9 ms/request env-setup cost is the target |
| Multi-core scaling | Duso uses every core with zero config (1.1k rps on 1 vCPU -> 15.7k on 16 threads); alternatives need cluster/fork setups; payoff grows with per-request work |
| Holding 1000+ connections on 1 GB | Duso, Node, Python survive; Ruby OOM-killed at 250 |
| Orchestrating 1000 concurrent fetches | Duso and Node clean; Python 12-52x slower; Ruby OOM-killed |
| Memory at rest | Duso smallest (6.6-8 MB vs Node's 30-39 MB) |
| Memory at extreme concurrency | Node leaner at 1000+ in-flight; Duso ~100-225 KB per unit (1.6.x target) |

The runtime does not need to win every benchmark. Its case is the package: one small binary that beats CPython on script code, runs primitives at Go speed, scales across every core without configuration, holds thousands of concurrent operations without drama, and stays deployable on the smallest VPS tier alongside your other processes. On the two machines that bracket real deployments -- the $6 VM and the developer workstation -- the data supports exactly that.

## Reproducing

Everything is in `/bench`: `fib.*`, `loop.*`, `sort.*`, `fib_builtin.du` (compute); `delay_server.*` (the four servers); `cfetch.*` (the four orchestration clients, worker count via `WORKERS` env var); `hey` drives the server tests. Raw results: `results-basics.csv`, `results-server.csv`, `results-client.csv`. Measured 2026-07-19 on Duso v1.6.0 (Go 1.26).
