# sys datastore

Access system-wide metrics and runtime statistics. The `sys` datastore is a special read-only namespace that tracks server metrics, process counts, and memory usage.

## Overview

The `sys` datastore is automatically initialized when your script runs. It provides metrics about:
- Server startup time
- Process counts (HTTP requests, spawned scripts, run() calls)
- Goroutine monitoring
- Memory allocation statistics
- Configuration passed via `-config` flag

## Accessing sys

```duso
sys = datastore("sys")
server_start = sys.get("server_start")
config = sys.get("config")
```

## Metrics

All metrics are read-only and computed on-demand (zero overhead if not accessed).

### Process Counters

- `http_procs` (number) - Total count of HTTP requests processed by `http_server()`
- `spawn_procs` (number) - Total count of scripts spawned via `spawn()`
- `run_procs` (number) - Total count of scripts executed via `run()`

### Goroutine Monitoring

- `active_goroutines` (number) - Current number of active goroutines
- `peak_goroutines` (number) - Highest goroutine count reached since startup

### Memory Metrics

- `heap_alloc` (number) - Current bytes allocated in heap
- `total_alloc` (number) - Total bytes ever allocated (cumulative)
- `heap_sys` (number) - Bytes obtained from OS for heap (includes unused reserved memory)
- `num_gc` (number) - Number of garbage collections run
- `peak_heap_alloc` (number) - Highest heap allocation reached since startup

### Server Info

- `server_start` (number) - Unix timestamp when server started
- `datastore_count` (number) - Number of active datastores in registry

### Configuration

- `config` (object) - Configuration passed via `-config` flag at startup (or nil if none provided)
- `listen_ports` (array) - Ports where `http_server()` instances are listening

## Examples

### Monitor HTTP Request Volume

Track request throughput in real-time:

```duso
sys = datastore("sys")

http_before = sys.get("http_procs")
print("HTTP requests before: " + format_json(http_before))

// Simulate some HTTP requests...
sleep(5)

http_after = sys.get("http_procs")
print("HTTP requests after: " + format_json(http_after))
print("Processed: " + format_json(http_after - http_before) + " requests")
```

### Check Memory Usage

Monitor memory consumption:

```duso
sys = datastore("sys")

heap = sys.get("heap_alloc")
heap_mb = heap / (1024 * 1024)
print("Heap allocated: {{heap_mb}} MB")

peak = sys.get("peak_heap_alloc")
peak_mb = peak / (1024 * 1024)
print("Peak heap: {{peak_mb}} MB")

gcs = sys.get("num_gc")
print("Garbage collections: {{gcs}}")
```

### Access Configuration

Read config passed from command line:

```bash
duso -config "port=8080, debug=true, max_workers=10" script.du
```

```duso
sys = datastore("sys")
config = sys.get("config")

port = config.port
debug = config.debug
workers = config.max_workers

print("Server on port {{port}}")
print("Debug mode: {{debug}}")
print("Max workers: {{workers}}")
```

### Uptime Tracking

Calculate how long server has been running:

```duso
sys = datastore("sys")

start_time = sys.get("server_start")
now_time = now()
uptime_seconds = now_time - start_time

print("Uptime: {{uptime_seconds}} seconds")
```

### Goroutine Leak Detection

Monitor goroutine count to detect leaks:

```duso
sys = datastore("sys")

initial = sys.get("active_goroutines")
print("Starting goroutines: {{initial}}")

// Do work...
spawn("task1.du")
spawn("task2.du")
spawn("task3.du")

sleep(2)

final = sys.get("active_goroutines")
print("Ending goroutines: {{final}}")

if final > initial + 5 then
  print("WARNING: Possible goroutine leak!")
end
```

### Configuration from -config Flag

Pass configuration as Duso object syntax:

```bash
duso -config "server_id=5, allow_admin=false, timeout=30" app.du
```

Configuration values are parsed using Duso syntax, so you can use:
- Numbers: `port=8080`
- Strings: `name="production"`
- Booleans: `debug=false`
- Arrays: `hosts=["10.0.0.1", "10.0.0.2"]`
- Objects: `db={host="localhost", port=5432}`

## Read-Only

The `sys` datastore is **read-only** to scripts. Attempting to write will error:

```duso
sys = datastore("sys")
sys.set("test", "value")  // Error: datastore("sys") is read-only
```

Internal runtime code can write to `sys` (incrementing counters, storing config), but scripts can only read.

## Configuration Not Allowed

The `sys` datastore does not accept configuration options:

```duso
sys = datastore("sys", {persist = "sys.json"})  // Error: does not accept configuration
```

Use a different namespace if you need persistent storage:

```duso
my_store = datastore("myapp", {persist = "data.json"})  // OK
```

## Metrics Update Timing

Metrics are computed on-demand when accessed:

- **Counters** (`http_procs`, `spawn_procs`, `run_procs`) - Updated in real-time as events occur
- **Goroutine stats** (`active_goroutines`, `peak_goroutines`) - Updated when accessed
- **Memory stats** (`heap_alloc`, `num_gc`, etc) - Updated when accessed via `runtime.ReadMemStats()`
- **Server info** (`server_start`, `datastore_count`) - Static or updated when accessed

This design means zero overhead if metrics aren't being read.

## Notes

- **Process Lifetime**: Counters are cumulative over the lifetime of the Duso process. Restart resets to zero
- **Per-Port Counters**: Use `port_8080`, `port_8081`, etc to get request count by port (populated by `http_server()`)
- **Peak Values**: Peak goroutines and peak heap allocation are high-water marks - they never decrease during runtime
- **Memory Units**: All memory values are in bytes. Divide by `1024` for KB, `1024*1024` for MB

## See Also

- [datastore() - Thread-safe key/value store](/docs/reference/datastore.md)
- [spawn() - Run script asynchronously](/docs/reference/spawn.md)
- [run() - Run script synchronously](/docs/reference/run.md)
- [http_server() - Create HTTP server](/docs/reference/http_server.md)
