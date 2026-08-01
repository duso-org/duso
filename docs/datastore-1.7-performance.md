# Datastore Performance: 1.6.21 → 1.7

Duso 1.7 rebuilt the datastore's write-ahead log and its on-disk value encoding.
This document measures what that changed.

Every number here was produced on **one machine, on one day, from both binaries
built from source**, running the same benchmark scripts. That matters: the
figures in `performance-report.md` were taken on an older Intel Mac, so comparing
against them measures the hardware as much as the software.

## TL;DR

- **Queue operations are ~200× faster and write ~2,000× less to disk.** This is
  the headline and everything else is a footnote to it.
- **Encrypted writes roughly doubled**, and the cost of turning encryption on
  fell from 44% of write throughput to 19%.
- **Loading a 100k-key store from disk is 3.4–3.8× faster.**
- **Pure in-memory performance is unchanged** — which is the correct result, and
  serves as the control group for everything above.

## Method

| | |
|---|---|
| Machine | Apple M1, 8 cores, macOS (Darwin 25.5.0) |
| Go | 1.26.5 (darwin/arm64) |
| Baseline | `v1.6.21-512`, built from the tag on this machine |
| Current | `v1.6.34-526-2` (1.7 development build) |
| Scripts | `test/create_datastore.du` + `test/bench_datastore.du` |

Six configurations: in-memory, WAL-backed, and WAL + AES-256-GCM encrypted, each
at 10,000 and 100,000 records. Records are six-field objects with a nested array.
`wal_sync_interval = 1` (batched), `persist_interval = 30`.

Reproduce:

```bash
# In-memory (create runs the benchmarks itself, since nothing persists)
duso test/create_datastore.du -config 'size=10000,persist=false,wal=false'

# WAL-backed, and WAL + encrypted — two processes, so recovery is exercised
rm -f /tmp/loadtest.dusnap /tmp/loadtest.duwal
duso test/create_datastore.du -config 'size=10000'
duso test/bench_datastore.du  -config 'size=10000'

rm -f /tmp/loadtest.dusnap /tmp/loadtest.duwal
duso test/create_datastore.du -config 'size=10000,encrypted=true'
duso test/bench_datastore.du  -config 'size=10000,encrypted=true'
```

Delete the two files between differently-configured runs.

## The headline: queues

`push()` / `shift()` / `pop_wait()` are how Duso scripts coordinate — spawned
workers share nothing but the datastore. They were also, until 1.7, by far the
most expensive thing it could do.

The old log recorded the **resulting value** of each operation rather than the
operation itself, so every `push` wrote the entire array to disk. Filling a queue
was quadratic.

**2,000 items pushed then drained, WAL-backed:**

| | 1.6.21 | 1.7 | |
|---|---:|---:|---|
| PUSH | 1,442/s | **300,075/s** | 208× |
| SHIFT | 1,507/s | **365,230/s** | 242× |
| WAL written | 328 MB | **167 KB** | 2,011× less |
| Bytes per operation | 85,979 | **42** | |

Encrypted, same test:

| | 1.6.21 | 1.7 | |
|---|---:|---:|---|
| PUSH | 1,406/s | **220,433/s** | 157× |
| SHIFT | 1,377/s | **248,168/s** | 180× |
| WAL written | 328 MB | **277 KB** | 1,214× less |
| Bytes per operation | 86,079 | **70** | |

**The ratio is not a constant.** 42 bytes per operation is flat regardless of
queue depth; the old cost grew with it. At 2,000 items the gap is ~200×; at 5,000
it measured ~480×. A deeper queue widens it further.

In pure in-memory mode both versions complete 2,000 queue operations in 1–2ms,
which is at the edge of the timer's resolution — the honest statement is that
there is no measurable difference, as expected, since no log is involved.

## Full results

### In-memory

| Metric | 1.6.21 10k | 1.7 10k | 1.6.21 100k | 1.7 100k |
|---|---:|---:|---:|---:|
| Insert | 408,670/s | 415,524/s | 449,367/s | 417,855/s |
| GET | 763,600/s | 734,695/s | 709,520/s | 654,690/s |
| SET | 369,646/s | 366,894/s | 300,880/s | 424,301/s |
| Atomic UPDATE | — | 354,533/s | — | 399,651/s |
| Atomic increment | 2,977,006/s | 2,642,747/s | 2,818,487/s | 2,616,648/s |
| SELECT predicate | 22ms | 21ms | 246ms | 255ms |
| COUNT predicate | 14ms | 14ms | 152ms | 149ms |
| SELECT max=100 | 1ms | 1ms | 2ms | 2ms |

Flat within noise, a few points slightly down. **This is the control group.** The
in-memory path was not touched, and it shows — which is what licenses reading the
persisted-mode differences below as real.

### WAL-backed

| Metric | 1.6.21 10k | 1.7 10k | 1.6.21 100k | 1.7 100k |
|---|---:|---:|---:|---:|
| Insert | 147,288/s | **199,812/s** | 149,898/s | **197,460/s** |
| Load from disk | 51ms | **19ms** | 478ms | **125ms** |
| GET | 708,222/s | 821,092/s | 616,226/s | 819,384/s |
| SET | 134,325/s | **184,148/s** | 138,798/s | **197,530/s** |
| Atomic UPDATE | — | 186,986/s | — | 202,680/s |
| Atomic increment | 377,260/s | 444,325/s | 396,616/s | 448,323/s |
| SELECT predicate | 21ms | 22ms | 245ms | 232ms |
| COUNT predicate | 13ms | 13ms | 149ms | 143ms |
| SELECT max=100 | 1ms | 2ms | 4ms | 2ms |

### WAL + encrypted (AES-256-GCM)

| Metric | 1.6.21 10k | 1.7 10k | 1.6.21 100k | 1.7 100k |
|---|---:|---:|---:|---:|
| Insert | 80,116/s | **163,425/s** | 81,136/s | **160,485/s** |
| Load from disk | 53ms | **19ms** | 502ms | **142ms** |
| GET | 696,288/s | 853,611/s | 605,414/s | 832,492/s |
| SET | 75,132/s | **150,376/s** | 83,036/s | **160,655/s** |
| Atomic UPDATE | — | 153,619/s | — | 161,175/s |
| Atomic increment | 145,966/s | **300,984/s** | 150,887/s | **321,128/s** |
| SELECT predicate | 22ms | 21ms | 245ms | 233ms |
| COUNT predicate | 13ms | 13ms | 149ms | 144ms |
| SELECT max=100 | 1ms | 1ms | 2ms | 2ms |

## Why it moved

**The log records operations, not results.** A `shift` is now 15 bytes on disk —
the fact that a shift happened — rather than a copy of everything left in the
array. That single change is the queue result above, and it is also why inserts
and updates got cheaper across every persisted mode.

**The encoding is Duso's own, not `gob`.** Records are 2–7× smaller than the
equivalent gob encoding, and self-contained rather than part of a stateful
stream. Smaller records mean less to write, less to read back, and less to
encrypt.

**Encryption became nearly free.** In 1.6.21, enabling it cost 44% of SET
throughput (134,325 → 75,132). In 1.7 it costs 19% (184,148 → 150,376). AES work
is proportional to bytes, and there are far fewer bytes.

**Loading is 3.4–3.8× faster at 100k** — 478ms → 125ms plain, 502ms → 142ms
encrypted. That is the codec replacing gob on the snapshot path, with nothing
else plausibly contributing.

One result is **not fully explained**: GET improved 15–35% in persisted modes
while staying flat in-memory. GET never touches the log, so the likely cause is
reduced allocation and GC pressure from writes no longer producing large
short-lived buffers. That is a hypothesis, not a measurement.

## Caveats

- **Atomic UPDATE has no baseline.** That benchmark was added after 1.6.21, so
  the 1.7 column stands alone.
- **In-memory queue numbers are below timer resolution** at 2,000 items and are
  reported as "no measurable difference" rather than as rates.
- **`wal_sync_interval = 1`**, so these are batched writes, not fsync-per-write.
  Fully durable mode (`0`) is slower in both versions.
- Single run per configuration, not a best-of-N.

## What else changed

Performance was not the motivation — correctness was. The old log never recorded
`shift_wait()`, `pop_wait()`, `rename()`, `clear()` or `expire()` at all, so a
crash could resurrect consumed queue jobs, undo a rename, restore cleared data,
and lose every TTL. `delete()` recovered as *key present with value nil* rather
than absent. Binary values could not be persisted at all — they failed with a raw
Go error the moment a WAL was configured.

All of that is fixed. The speed is a consequence of fixing it properly.
