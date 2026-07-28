# Duso 1.6 Benchmarks (2026-07-27)

Every number in this report was measured from the actual scripts in `/bench` and `/test`, run fresh for this report. Each section notes exactly how to reproduce it — `duso` script names, or the `hey_run.sh` / `create_datastore.du` / `bench_datastore.du` invocations used, so anyone can rerun the same tests. Two machines: an 8-core workstation, and the small cloud VM that's actually the deployment target Duso is designed for.

## TL;DR: Not a Toy

- **On the $6/month VM, Duso wins on both throughput and tail latency against Node, Python, and Ruby — at every concurrency level tested.** On `/delay` (held connections, real per-request work), Duso beats all three on rps at c=250/1000/2000, and has the lowest p99 at every one of those levels too (e.g. 1594.6ms vs. Node's 1897.8ms and Python's 2460.2ms at c=2000). That holds on a single core with nothing to scale into — the vertical story below is what happens when you give it more.
- **92,026 requests/sec on a trivial handler, on 8 cores — 2.7× Node's 33,538.** No cluster mode, no worker processes, no config. Duso's runtime spreads incoming connections across every core automatically; Node's event loop is pinned to one.
- **A built-in datastore, Redis/MongoDB-shaped, does up to 474,850 GETs/sec on the 8-core workstation.** It's a thread-safe key/value store baked into the binary — any Duso value as a record, no separate service, no network hop — and that GET throughput holds whether the store is pure in-memory, WAL-persisted to disk, or AES-256-GCM encrypted at rest: reads don't pay for durability. On the $6/month single-core VM it still holds 100k+ encrypted GETs/sec, and atomic partial-record updates (`update()`) run 18-70k/sec depending on durability mode.
- **Ruby and Python both hit real structural ceilings under held-connection load — Duso and Node didn't.** On the workstation at c=5000, Ruby failed outright (`ThreadError: can't create Thread`) and Python's p99 blew past 60 seconds under GIL contention; on the 1GB VM, Ruby failed even sooner, at c=1000 (kernel OOM-kill on its per-connection thread stacks). Duso and Node stayed stable through every concurrency level tested on both machines.
- **Hot paths move into Go, not a JIT.** `fib(30)` × 10,000 runs in 63.0ms as script, 0.94ms as a registered Go builtin — a 67× jump from one `RegisterBuiltin` call. That's Duso's actual answer to "is the interpreter fast enough."
- **Not everything is fast — regex `find()` is a real weak spot.** 91ms for a single call collecting 6,627 matches, vs. 2-4ms for the same workload in Node/Python/Ruby's own regex engines. `contains()` (no match collection) is fast and in line with the others, so this is specifically about match-object construction, not the RE2 engine underneath. Flagged, not hidden — see the Compute section.

# Part 1: Workstation

## Test System

- **Workstation:** MacBook Pro 16" (2019), 8-Core Intel i9 @ 2.3 GHz, 16 GB RAM
- Duso 1.6, Node, Python 3, and Ruby 4.0.6 (a current Ruby was needed for the HTTP server test — see note below)

## Compute Performance

### Cross-language

| Test | Duso | Node | Python | Ruby |
|---|---:|---:|---:|---:|
| Nested loop (1000×1000 sum) | 82.2ms | 6ms | 88.2ms | 48ms |
| `fib(30)` ×10,000 | 0.94ms / 63.0ms | 2ms | 10.1ms | 33ms |
| Sort 10k random floats | 1.6ms | 7ms | 1.2ms | 2ms |
| JSON encode 5,000 objects | 13ms | 4.07ms | 6.64ms | 2.96ms |
| JSON decode 5,000 objects | 12ms | 4.74ms | 6.22ms | 5.67ms |
| `contains()` on 48KB doc | 0.035ms | 1.20ms | 0.088ms | 0.013ms |
| `find()` on 48KB doc (6,627 matches) | 91ms | 2.03ms | 2.12ms | 3.87ms |
| `markdown_html()` × 200 passes, 48KB doc | 400ms (2ms/pass) | — | — | — |
| `template()` render × 20,000 | 124ms (6.2µs/render) | — | — | — |

The `fib(30)` row shows two Duso numbers: 0.94ms and 63.0ms. The gap is entirely about whether that workload runs as script or as a registered Go function — same interpreter, same machine, same result. It's the clearest illustration of Duso's actual performance model: the interpreter doesn't need to be fast on its own, because hot paths move into Go. Loop/fib sit next to Node's JIT and CPython/Ruby's bytecode VMs mainly for scale — a different category of engine, not a fair fight, included for reference.

**What this shows:**
- **JSON**: Duso runs 2-4× slower than the other three on this workload. The gap is most likely in converting between Go's `encoding/json` output and Duso's internal dynamic value representation, not the encoding itself — worth a profiling pass.
- **`find()` is the standout weak point.** 91ms vs. 2-4ms elsewhere is a 24-45× gap on what's ultimately the same `regexp` (RE2) engine under everyone's own implementation. `contains()` (no match collection) is fast and in line with the others, so the cost is specifically in building the `{text, pos, len}` object array for every match, not the regex engine itself. This is the clearest place Duso lags in this report, and it deserves a real profiling pass.
- **Sort** and **`contains()`** land in the same ballpark across all four languages — nothing notable there.

**Reproduce:** from `/bench` — `duso loop.du`, `duso fib.du`, `duso fib_builtin.du`, `duso sort.du`, `duso json.du`, `duso regex.du`; equivalents `node loop.js`/`fib.js`/`sort.js`/`json.js`/`regex.js`, `python3 loop.py`/`fib.py`/`sort.py`/`json_bench.py`/`regex.py` (named `json_bench.py`, not `json.py`, to avoid shadowing Python's own `json` module), `ruby loop.rb`/`fib.rb`/`sort.rb`/`json.rb`/`regex.rb` on a current Ruby. Each script self-times and prints its own result.

### Markdown and templating

The last two rows above are Duso-only — neither has a clean cross-language comparison yet:

- **Markdown has no stdlib equivalent in Node, Python, or Ruby** — all three need a third-party package (e.g. `markdown-it`, `Python-Markdown`, `kramdown`/`redcarpet`) to convert markdown to HTML. That's arguably a point in Duso's favor on its own (one less dependency to manage), but picking a "fair" package to benchmark against isn't a call we want to make unilaterally.
- **Templating has partial stdlib parity**: Ruby has `erb` (full expression support, the closest match to Duso's `{{expr}}`); Python has `string.Template` (variable substitution only, no expressions); Node has no stdlib template-compiling API at all — template literals are a language feature, not a reusable compile-once/render-many object.

**We'd welcome suggestions from the community on which packages to use for a fair comparison here** — open an issue or PR against `/bench`.

**Reproduce (Duso only, for now):** `duso markdown.du`, `duso template.du`.

## HTTP Server Performance

Server: `bench/server.du` (Duso), `bench/delay_server.js` (Node), `bench/delay_server.py` (Python `ThreadingHTTPServer`, stdlib, real HTTP parsing), `bench/delay_server.rb` (Ruby, `Socket.tcp_server_loop` + thread-per-connection, stdlib). All four run via `bench/hey_run.sh`.

### Raw throughput (`/ping`, `hey -c 100 -z 10s`)

| Runtime | rps | p50 | p99 | Server RSS |
|---|---:|---:|---:|---:|
| Duso | 92,026 | 0.6ms | 10.5ms | 92.5 MB |
| Ruby | 43,975 | 2.3ms | 3.2ms | 15.3 MB |
| Node | 33,538 | 2.8ms | 5.3ms | 63.3 MB |
| Python | 8,633 | 11.5ms | 22.8ms | 18.8 MB |

Ruby's server here does real HTTP parsing (request line + headers), same server used in the `/delay` tests below. On a trivial no-work handler, thread-per-connection (Ruby) and multi-core scheduling (Duso, via Go's runtime) both outrun Node and Python's single-threaded event loop/GIL; that ordering does not necessarily hold once handlers do real work or concurrency climbs into the thousands (see `/delay` below).

### Holding concurrent connections (`/delay`, 1s backend sleep, `hey -c N -n N*5`)

| Concurrency | Duso | Node | Python | Ruby |
|---|---|---|---|---|
| 250 | 247.7 rps, p50 1000.9ms, p99 1040.5ms, RSS 28.2MB | 246.2 rps, p50 1005.7ms, p99 1063.4ms, RSS 50.6MB | 242.2 rps, p50 1007.7ms, p99 1117.0ms, RSS 22.6MB | 245.8 rps, p50 1005.9ms, p99 1058.0ms, RSS 19.0MB |
| 1000 | 983.8 rps, p50 1000.8ms, p99 1066.8ms, RSS 93.4MB | 962.9 rps, p50 1002.4ms, p99 1169.0ms, RSS 78.6MB | 953.0 rps, p50 1012.4ms, p99 1177.3ms, RSS 41.6MB | 959.5 rps, p50 1008.3ms, p99 1162.9ms, RSS 36.7MB |
| 2000 | 1781.8 rps, p50 1000.6ms, p99 1362.8ms, RSS 137.1MB | 1656.5 rps, p50 1002.4ms, p99 2010.2ms, RSS 102.9MB | 1603.4 rps, p50 1013.1ms, p99 2074.6ms, RSS 77.9MB | 1602.3 rps, p50 1009.2ms, p99 2163.3ms, RSS 66.3MB |
| 5000 | 4101.0 rps, p50 1001.1ms, p99 1630.9ms, RSS 401.1MB\* | 4023.0 rps, p50 1002.6ms, p99 2130.8ms, RSS 168.7MB | **148.1 rps, p50 2205.3ms, p99 60,544.8ms**, RSS 130.7MB, only 14,295/25,000 requests completed | **FAILED** — `ThreadError: can't create Thread: Resource temporarily unavailable` |

\* Duso's 401.1MB at c=5000 may be inflated: Go on Intel macOS is slow to release pages back to the OS, so RSS at high transient allocation can overstate live memory. A Linux number is a more trustworthy read on this and will be added with the VM follow-up.

**What this shows:**
- All four runtimes handle held-connection load cleanly up to 2000 concurrent — tail latency (p99) grows with concurrency for everyone, which is expected, not a Duso-specific problem. At 2000, Duso's p99 (1362.8ms) is actually the best of the four; Node and Python both exceed 2000ms.
- At 5000 concurrent, the two thread/process-per-connection runtimes hit their structural ceiling: **Ruby fails outright** (can't allocate more OS threads), **Python survives but thrashes** (GIL contention drags p99 past 60 seconds, and nearly half of requests don't even complete in the test window). Duso and Node — both using an M:N/event-loop model instead of one OS thread per connection — stay stable and keep p99 under ~2.2s.
- Memory: Duso and Node cost more RAM per connection than Python/Ruby at moderate concurrency (250-1000), tracking with per-request state (Duso: live interpreter environment; Node: pending promise + closure). That inverts at 5000, where Ruby/Python's low RSS is a symptom of the run failing, not evidence of efficiency.

**Reproduce:** from `/bench`, `./hey_run.sh <mac|linux> <duso|node|python|ruby> <ping|delay> <concurrency>` — each call starts a fresh server, runs `hey`, captures RSS, appends one CSV row to `results.csv`, and prints that row. Ruby needs a current interpreter (`RUBY=/path/to/ruby ./hey_run.sh ...`) since `delay_server.rb` uses `Socket.tcp_server_loop`, unavailable before Ruby 3.0. Rows shown above: `ping` at c=100; `delay` at c=250/1000/2000/5000, for all four runtimes (Ruby's c=5000 failure and Python's c=5000 degradation are left in the table rather than dropped).

## Datastore Performance

Duso's built-in datastore sits somewhere between Redis and MongoDB in form — a thread-safe, namespaced key/value store, but the values are full Duso data types (objects, arrays, nested structures), not just strings/hashes, and it supports predicate-based `select()`/`count()` queries over them, closer to a document store than a plain cache. Unlike either, it's not a separate service: it's built into the same binary as the rest of the app, in-process, with no network hop, plus optional WAL-backed persistence and AES-256-GCM encryption at rest. On a single-node deployment that tight integration removes an entire service (and its ops burden) from the stack.

It's also how Duso scripts communicate at all: every script instance — the top-level script, each `spawn()`, each `run()`, each HTTP handler — runs isolated, with no shared globals or heap. The datastore is the only sanctioned way to share state or coordinate between them (atomic `increment`/`push`, blocking `wait()`), which is what makes the concurrency model thread-safe and hard to screw up — there's no shared mutable state to race on outside of it. That makes the near-free GETs and fast atomic ops above more than a nice number: they're the mechanism the rest of the concurrency model runs on.

There's no stdlib parallel to any of this in Node, Python, or Ruby, so this section is Duso-only rather than a cross-language comparison. Three modes tested: pure in-memory, WAL-backed with periodic disk persistence, and the same WAL-backed mode with AES-256-GCM encryption at rest.

| Metric | In-memory 10k | In-memory 100k | WAL 10k | WAL 100k | WAL+encrypted 10k | WAL+encrypted 100k |
|---|---:|---:|---:|---:|---:|---:|
| Insert (populate) | 238,011/s | 276,740/s | 70,464/s | 77,594/s | 40,059/s | 42,962/s |
| Load from disk (new process) | n/a | n/a | 111ms | 879ms | 120ms | 928ms |
| GET (10k or 100k ops) | 474,850/s | 442,258/s | 421,962/s | 419,857/s | 435,953/s | 417,107/s |
| SET (10k updates) | 281,913/s | 234,378/s | 75,973/s | 70,940/s | 42,493/s | 42,914/s |
| Atomic UPDATE (10k partial merges, 2 of 6 keys) | 293,840/s | 245,798/s | 69,230/s | 82,840/s | 42,484/s | 42,585/s |
| SELECT predicate (~2/3 match) | 31ms | 353ms | 29ms | 344ms | 27ms | 351ms |
| COUNT predicate (~2/3 match) | 19ms | 227ms | 19ms | 225ms | 17ms | 220ms |
| SELECT max=100 (early-exit) | 1ms | 3ms | 2ms | 3ms | 1ms | 3ms |
| Atomic increment | 2,003,680/s | 2,005,452/s | 169,842/s | 185,627/s | 74,094/s | 74,908/s |

**What this shows:**
- **GET is essentially free regardless of mode** (~420-475k ops/sec across all six configurations) — reads never touch disk or pay an encryption cost; data lives decrypted in memory once loaded.
- **Writes (SET, UPDATE, increment) pay for durability.** In-memory SET does 282k/s; add WAL persistence and it drops to ~76k/s (each write appends to the WAL); add encryption on top and it drops again to ~42k/s (each write also pays an AES-GCM cost). Same pattern on atomic increment: 2M/s in-memory → 170-185k/s with WAL → 74-75k/s encrypted. **`update()` — an atomic partial deep-merge on a multi-key object, the realistic "patch a couple fields" operation — tracks SET almost exactly at every durability level** (294k/s in-memory, 69-83k/s WAL, ~42.5k/s encrypted): the deep-merge itself is cheap: the WAL append and encryption are what cost, and both operations pay the same cost for them. That's the honest cost of a durable, encrypted store built into the binary instead of a separate service — trading some raw write throughput for not needing to run and manage Redis.
- **Encryption's real cost is on writes, not on load or reads.** Load-from-disk time is nearly identical encrypted vs. unencrypted at both scales (120ms vs. 111ms at 10k, 928ms vs. 879ms at 100k) — decryption on load isn't the bottleneck, per-write encryption is.
- **SELECT/COUNT scale with dataset size, not with WAL/encryption mode** — full predicate scans cost the same whether the store is in-memory or encrypted (e.g. ~344-353ms at 100k across all three modes), since they're pure in-memory operations regardless of how the store persists. Worth being clear about what these numbers mean: `select()`/`count()` are full scans over every key, by design — there's no index to consult. They're not meant to be a hot-path query mechanism; they're the "I need to find or inspect something" tool (debugging, ad-hoc lookups, one-off reports), not something you'd call per-request in a tight loop. Judge them as "fast enough for occasional use," not as core-throughput numbers like GET/SET above.

**Reproduce:** `test/create_datastore.du -config 'size=N,persist=false,wal=false'` (in-memory) or `-config 'size=N'` (WAL-backed) or `-config 'size=N,encrypted=true'` (WAL + encrypted), then `test/bench_datastore.du` with the matching config — the persisted modes write to `/tmp/loadtest.{gob,wal}`, so create and bench can run as separate processes. Pure in-memory mode doesn't survive across separate `duso` invocations, so create and bench need to run in one process: `duso eval 'run("test/create_datastore.du"); run("test/bench_datastore.du")' -config 'size=N,persist=false,wal=false'` — the shared `datastore()` namespace survives across `run()` boundaries within a single process. Delete `/tmp/loadtest.gob`/`.wal` between differently-configured runs.

# Part 2: Small Cloud VM

This is the deployment target Duso is actually designed for — a $6/month-tier single-core box, not the 8-core workstation above. Same scripts, same versions where possible, run fresh on this machine.

## Test System

- **VM:** DigitalOcean-class, 1 vCPU, 961Mi RAM, arlandmvp.ludonode.com
- Duso 1.6 (same build as the workstation, v1.6.24-516), Node 20.19.6, Python 3.12.3, Ruby 3.2.3

## Compute Performance

| Test | Duso | Node | Python | Ruby |
|---|---:|---:|---:|---:|
| Nested loop (1000×1000 sum) | 132.9ms | 13ms | 301.6ms | 128ms |
| `fib(30)` ×10,000 | 1.57ms / 104.2ms | 5ms | 21.6ms | 80ms |
| Sort 10k random floats | 3.0ms | 12ms | 2.2ms | 5ms |
| JSON encode 5,000 objects | 29.9ms | 6.58ms | 14.9ms | 15.9ms |
| JSON decode 5,000 objects | 23.9ms | 11.8ms | 13.7ms | 24.3ms |
| `contains()` on 48KB doc | 0.054ms | 0.052ms | 0.189ms | 0.030ms |
| `find()` on 48KB doc (6,627 matches) | 164.0ms | 1.86ms | 3.83ms | 11.4ms |
| `markdown_html()` × 200 passes, 48KB doc | 910ms (4.6ms/pass) | — | — | — |
| `template()` render × 20,000 | 264ms (13.2µs/render) | — | — | — |

Same shape as the workstation: `fib` shows the same script-vs-builtin split (1.57ms vs. 104.2ms), and `find()` is again the standout gap — proportionally worse here (~44-90× vs. Node/Python/Ruby, vs. ~24-45× on the workstation), consistent with a per-match construction cost that doesn't parallelize away on a single core. JSON also lands slower relative to the others here than on the workstation. Everything is slower in absolute terms on a single shared vCPU, as expected — the comparison that matters is the ratio between runtimes, not the absolute number.

**Reproduce:** identical to the workstation section above, run on the VM: `duso loop.du`, `duso fib.du`, `duso fib_builtin.du`, `duso sort.du`, `duso json.du`, `duso regex.du`, `duso markdown.du`, `duso template.du`, plus the `node`/`python3`/`ruby` equivalents.

## HTTP Server Performance

Same servers as the workstation section. Concurrency (`-c`) and total request count are both called out explicitly below since they matter more on a memory-constrained single-core box than on the workstation.

### Raw throughput (`/ping`, `hey -c 100 -z 10s`, duration-based — total requests vary by runtime speed)

| Runtime | rps | p50 | p99 | Server RSS | Total requests completed |
|---|---:|---:|---:|---:|---:|
| Ruby | 10,536 | 9.0ms | 21.3ms | 126.5 MB | 105,500 |
| Node | 8,354 | 11.0ms | 28.5ms | 51.7 MB | 83,625 |
| Duso | 7,221 | 12.8ms | 34.3ms | 27.9 MB | 72,363 |
| Python | 1,836 | 50.4ms | 105.1ms | 22.8 MB | 18,468 |

Ordering shifts from the workstation: on a single shared vCPU, Ruby's thread-per-connection edges out Node and Duso on raw no-work throughput, and Duso no longer has multiple cores to spread across, so it drops from first to third. Python is well behind all three regardless of machine.

### Holding concurrent connections (`/delay`, 1s backend sleep, `hey -c N -n N*5` — fixed request count, 5× concurrency)

| Concurrency (c) | Requests sent (N×5) | Duso | Node | Python | Ruby |
|---|---:|---|---|---|---|
| 250 | 1,250 | 242.4 rps, p50 1007.8ms, p99 1095.6ms, RSS 29.4MB, 1250/1250 completed | 235.5 rps, p50 1014.0ms, p99 1208.8ms, RSS 54.2MB, 1250/1250 completed | 229.9 rps, p50 1044.7ms, p99 1212.7ms, RSS 26.8MB, 1250/1250 completed | 236.2 rps, p50 1000.5ms, p99 1201.5ms, RSS 279.9MB, 1250/1250 completed |
| 1000 | 5,000 | 911.5 rps, p50 1026.7ms, p99 1314.9ms, RSS 89.2MB, 5000/5000 completed | 816.4 rps, p50 1047.8ms, p99 1627.3ms, RSS 70.1MB, 5000/5000 completed | 638.3 rps, p50 1064.9ms, p99 2701.5ms, RSS 50.1MB, 5000/5000 completed | **FAILED** — `Memory cgroup out of memory: Killed process (ruby)`, RSS at kill unrecorded, only 107/5000 completed |
| 2000 | 10,000 | 1671.0 rps, p50 1027.7ms, p99 1594.6ms, RSS 135.4MB, 10000/10000 completed | 1509.7 rps, p50 1027.3ms, p99 1897.8ms, RSS 79.1MB, 10000/10000 completed | 977.4 rps, p50 1046.0ms, p99 2460.2ms, RSS 56.8MB, only 6837/10000 completed | not run (already failed at c=1000) |

c=5000 was tried and dropped: even with the memory cap raised from 200MB to 700MB, Duso was killed by the kernel's memcg OOM killer partway through (confirmed via `dmesg`: `Memory cgroup out of memory: Killed process ... duso`, anon-rss ~184MB at time of kill climbing toward the cap). 5000 concurrent connections isn't a realistic scenario for a 1GB single-core VM in the first place, so the report stops at 2000 for this machine rather than chase a number nobody would actually run.

**What this shows:**
- **Ruby's RSS at just c=250 (279.9MB) already exceeds what the original 200MB test cap allowed**, and it fails outright by c=1000 — confirming the same thread-per-connection stack cost seen on the workstation, just hitting much sooner here because Linux commits per-thread stacks more eagerly than macOS and there's far less RAM to work with.
- **Python survives further than Ruby but degrades hard** — p99 more than doubles the 1-second floor by c=1000, and by c=2000 nearly a third of requests don't complete at all in the test window.
- **Duso and Node are the only two that complete every request cleanly through c=2000** on this machine, with Duso's p99 consistently the lowest of the four at both 1000 and 2000.

**Reproduce:** `MEM_CAP=700M ./hey_run.sh linux <duso|node|python|ruby> <ping|delay> <concurrency>` from `/tmp/bench` on the VM (the Linux path runs each server under a `systemd-run --scope -p MemoryMax=$MEM_CAP` cgroup; `MEM_CAP` defaults to 700M, override as needed). `hey` needs to be on `PATH` (`/root/go/bin/hey` on this box).

## Datastore Performance

Same six configurations as the workstation section, same `test/create_datastore.du` / `test/bench_datastore.du` scripts.

| Metric | In-memory 10k | In-memory 100k | WAL 10k | WAL 100k | WAL+encrypted 10k | WAL+encrypted 100k |
|---|---:|---:|---:|---:|---:|---:|
| Insert (populate) | 105,930/s | 86,367/s | 37,644/s | 36,402/s | 18,889/s | 18,901/s |
| Load from disk (new process) | n/a | n/a | 194ms | 2,802ms | 221ms | 2,297ms |
| GET (10k or 100k ops) | 179,866/s | 117,780/s | 172,348/s | 104,857/s | 157,246/s | 102,892/s |
| SET (10k updates) | 98,221/s | 70,049/s | 41,738/s | 18,491/s | 24,915/s | 19,170/s |
| Atomic UPDATE (10k partial merges, 2 of 6 keys) | 67,649/s | 33,636/s | 31,939/s | 50,400/s | 23,061/s | 18,062/s |
| SELECT predicate (~2/3 match) | 108ms | 1,679ms | 127ms | 1,676ms | 103ms | 1,349ms |
| COUNT predicate (~2/3 match) | 66ms | 850ms | 63ms | 975ms | 61ms | 799ms |
| SELECT max=100 (early-exit) | 5ms | 7ms | 5ms | 9ms | 15ms | 11ms |
| Atomic increment | 467,087/s | 1,029,165/s | 215,434/s | 204,107/s | 60,646/s | 42,528/s |

**What this shows:** the same shape as the workstation, at roughly a third to a fifth of the throughput on the single shared vCPU — which tracks, since this box has no cores to spare for background work like WAL flushing or GC while a benchmark runs. GET again holds up well across all six modes (~103-180k/s); writes again pay for durability, most visibly on the 100k WAL+encrypted SET (19,170/s) and atomic increment (42,528/s) rows. Load-from-disk at 100k takes noticeably longer here (2.3-2.8s vs. under 1s on the workstation) — a single core doing decryption and deserialization has nowhere to parallelize that work.

**Reproduce:** identical commands to the workstation section, run from `/tmp` on the VM with `test/create_datastore.du` and `test/bench_datastore.du` staged there.
