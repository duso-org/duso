# shutdown()

Stop the process cleanly, from anywhere. Available in `duso` CLI only.

`shutdown([exit_code])`

## Parameters

- `exit_code` (optional, number) - Process exit status, a whole number from 0 to 255. Default 0.

## Returns

Nothing. `shutdown()` ends the current script instance the way `exit()` does, so no statement after it runs.

## shutdown() vs exit()

- **`exit(value)`** ends *this script instance* and hands a value back to whoever ran it. Inside an HTTP handler it ends only that handler; the server keeps serving.
- **`shutdown(code)`** ends *the process*.

That distinction is the reason `shutdown()` exists. A script blocked in `server.start()` has no statements left to reach — `exit()` from a request handler ends the handler and nothing else, so before `shutdown()` the only way to stop a running server was a signal from outside.

## What happens

In order:

1. Servers stop accepting new connections; in-flight handlers get up to 30 seconds to finish and send their responses.
2. WebSocket connections close.
3. Datastores sync their WAL and write a final snapshot.
4. The process exits with `exit_code`.

The calling script instance ends immediately, before any of that finishes. Everything else keeps running until its step arrives — which is what lets a handler that called `shutdown()` be part of the drain rather than deadlock against it.

## Examples

An admin route that stops the service:

```duso
// handlers/stop.du
ctx = context()
shutdown(0)
```

The client gets `204 No Content`. A response method like `res.text()` ends the handler by itself, so it can never run after `shutdown()` — a handler that shuts down does not send a body.

Stopping on a condition, with a non-zero status for the service manager:

```duso
config = parse_json(load("/etc/arla/config.json"))
if type(config) == "error" then
  print("config is unreadable, refusing to serve")
  shutdown(1)
end
```

From a scheduled job or a background worker:

```duso
// maintenance.du, run by schedule()
if file_exists("/var/lib/arla/STOP") then
  shutdown(0)
end
```

## Notes

- Safe to call from any script instance: the main script, a request handler, a `spawn()`, a scheduled job.
- Calling it twice is harmless. The first exit code wins; a later `shutdown(1)` will not change an in-progress `shutdown(0)`.
- A running server that is draining will not accept new connections, so clients see connection-refused for the length of the drain.
- The same sequence runs on `SIGINT`/`SIGTERM` (Ctrl+C, `systemctl stop`), which exits 0.
- `exec()` children are not terminated. A command still running keeps running.

## See Also

- [exit() - Return a value from the current script](/docs/reference/exit.md)
- [http_server() - HTTP server, and what start() does on shutdown](/docs/reference/http_server.md)
- [datastore() - Persistence and what is flushed](/docs/reference/datastore.md)
- [kill() - Terminate one spawned instance](/docs/reference/kill.md)
