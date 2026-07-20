# Duso Benchmark Suite

Three suites, run on two machines (a multi-core workstation and a small 1 vCPU
cloud VM). Feeds `docs/performance-report.md`. Run everything **one test at a
time** and read results between runs -- concurrent benchmarks contaminate each
other.

## Suites

**1. Basics** -- interpreter compute + baseline memory.
`fib.*`, `loop.*`, `sort.*` (duso/py/js/rb), `fib_builtin.du`.
Harness: `basics.sh mac|linux` (run from this directory; median of 3 +
peak RSS via `/usr/bin/time`). Override the duso binary with `DUSO=path`.

**2. Server** -- each language's minimal stdlib HTTP server under a neutral
load generator ([hey](https://github.com/rakyll/hey), single Go binary).
Servers: duso is `server.du` + `ping.du` + `delay.du` (route-per-handler, the
idiomatic layout -- the older single-file `delay_server.du` dispatched inside
the handler via context()/request() and is kept for reference); the others are
`delay_server.{js,py,rb}`. All serve `GET /delay` (respond after 1 s; tests
connection-holding + memory) and `GET /ping` (respond immediately; tests raw
throughput) on `127.0.0.1:8399`.
Harness: `hey_run.sh mac|linux LANG delay|ping CONCURRENCY` -- starts that
language's server, drives it, records rps / p50 / p99 / server peak RSS,
kills the server, appends a CSV row to `results.csv` in the cwd.

**3. Client orchestration** -- the language as API consumer: N concurrent
workers x 5 fetches each against its *own* language's delay server
(duso `spawn()` vs Promise.all vs ThreadPoolExecutor vs Thread.new).
Clients: `cfetch.{du,js,py,rb}` + `cfetch_worker.du`; worker count via
`WORKERS` env (default 100).
Harness: `pair_run.sh mac|linux LANG WORKERS` -- runs the server+client
pair, records wall time / client RSS / server peak RSS, appends to
`cfetch_results.csv` in the cwd.

## Running

```sh
cd bench
./basics.sh mac                      # compute + memory table
./hey_run.sh mac duso delay 500      # one server run
./hey_run.sh mac duso ping 100
./pair_run.sh mac duso 500           # one orchestration run
```

Escalate concurrency stepwise (100 -> 250 -> 500 -> 1000 -> 2000) and drop a
runtime once it blows up. On `linux`, both harnesses wrap the processes in
`systemd-run` cgroups (200 MB server / 180 MB client) so a runaway runtime is
killed instead of taking the box down -- **size those caps below the machine's
actual free RAM** (`free -m`), or the global OOM killer picks the victim.

## Before you trust a number

- **macOS listen backlog**: `sysctl kern.ipc.somaxconn` is 128 by default.
  Any Mac run with concurrency > 128 bursts connections against it; dropped
  SYNs retransmit on 1 s backoff and poison p99 and rps. Raise it first
  (`sudo sysctl -w kern.ipc.somaxconn=4096`, resets on reboot) or run the
  high-concurrency ladders on Linux only.
- **Stale servers**: a leftover process on 8399 makes new servers die at bind
  while the load generator happily measures the zombie. `pkill -f
  delay_server; lsof -i :8399` before each session (`pkill -9 -x duso16`
  etc. on the VM -- and beware `pkill -f` matching your own ssh command).
- **File descriptors**: `ulimit -n 16384` (harnesses do this).
- Surprising cross-runtime results deserve a cross-OS sanity check before
  they go in a report.
- Mac system Ruby is ancient (2.6); the VM's is 3.2. `delay_server.rb` is a
  raw socket loop (webrick left the stdlib), so its `/ping` rps is not an
  HTTP-server number.
- `DUSO_PPROF=127.0.0.1:6060 duso ...` exposes Go pprof for heap profiling
  during a run: `go tool pprof -top -inuse_space http://127.0.0.1:6060/debug/pprof/heap`.

## Artifacts

Raw appends (working files, in cwd): `results.csv` (server runs),
`cfetch_results.csv` (orchestration runs).

Curated results checked in from the 2026-07-19 session (duso v1.6.0-489,
pre-gzip-fix -- server numbers are stale once the http_server gzip patch
lands):

- `results-basics.csv` -- machine,lang,test,ms,peak_rss_mb
- `results-server.csv` -- machine,lang,endpoint,concurrency,rps,p50_ms,p99_ms,server_rss_mb,status
- `results-client.csv` -- machine,lang,workers,wall_ms,client_rss_mb,server_rss_mb,errors,status

`status` marks rows that are not clean measurements: `server_oom_killed` /
`client_oom_killed` (blew the VM cgroup cap), `macos_backlog_artifact` (see
above), `not_run_after_1000_blowup`. Machine is `mac` or `vm`/`linux`
(1 vCPU / 961 MB Ubuntu). All three are bar-graph-ready.
