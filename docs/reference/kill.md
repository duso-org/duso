# kill()

Terminate a spawned process by sending it a cancellation signal.

## Signature

```duso
kill(pid)
```

## Parameters

- `pid` (number) - The process ID returned by spawn()

## Returns

- `true` if the process was found and signaled
- Error if the PID doesn't exist or is invalid

## Behavior

- Sends a cancellation signal to the spawned goroutine
- The process exits gracefully at the next execution checkpoint
- The spawned script cannot ignore the signal; it will exit when it checks the context
- Returns immediately without waiting for the process to actually exit

## How It Works

Internally, each spawned process is given a cancellable context. When `kill()` is called:
1. The context is cancelled
2. The spawned script exits on its next iteration or checkpoint
3. Any resources held by the script are cleaned up via deferred cleanup handlers

## Examples

Kill a worker after it's done some work:

```duso
pid = spawn("worker.du")
sleep(5)
kill(pid)
print("Worker killed")
```

Kill multiple workers:

```duso
pids = []
for i = 1, 5 do
  push(pids, spawn("worker.du"))
end

sleep(10)

for pid in pids do
  kill(pid)
end

print("All workers killed")
```

Handle kill errors:

```duso
pid = spawn("worker.du")

if err = kill(999) then
  print("Error:", err)
end

if kill(pid) then
  print("Successfully killed process", pid)
end
```

## Error Cases

- **PID not found**: Returns error if the PID doesn't correspond to a spawned process
- **Invalid PID**: Returns error if PID is not a valid number

```duso
if err = kill(invalid_pid) then
  print("Error:", err)
end
```

## Notes

- The process ID returned by `spawn()` is the unique identifier for that process
- Each spawned process gets its own PID from an incrementing counter
- Once a process exits (whether by `kill()` or naturally), its PID cannot be reused
- Attempting to kill the same PID twice will error on the second call (PID already cleaned up)

## See Also

- [spawn() - Run script in background](/docs/reference/spawn.md)
- [run() - Run script synchronously](/docs/reference/run.md)
- [context() - Access request context](/docs/reference/context.md)
