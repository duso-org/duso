# datastore()

Create a thread-safe in-memory key/value store with optional JSON persistence. Perfect for coordinating work between spawned scripts.

## Signature

```duso
datastore(namespace [, config])
```

## Parameters

- `namespace` (string) - Namespace identifier. Multiple scripts access the same store via same namespace
- `config` (optional, object) - Configuration object:
  - `persist` (string) - Path to JSON file for auto-save and auto-load
  - `persist_interval` (number) - Auto-save interval in seconds (only if persist configured)

## Returns

Datastore object with methods

## Methods

### Key-Value Operations
- `set(key, value)` - Store any Duso value (thread-safe)
- `set_once(key, value)` - Atomically set value only if key doesn't already exist. Returns true if set, false if key already existed. Useful for cache initialization under concurrent load
- `get(key)` - Retrieve value by key (returns nil if not found)
- `swap(key, newValue)` - Atomically exchange key's value and return the old value. Useful for atomic consume/replace patterns
- `increment(key [, delta])` - Atomically add delta to number. Delta defaults to 1 if not provided. Starts at 0 if key doesn't exist. Returns the new value
- `decrement(key [, delta])` - Atomically subtract delta from number. Delta defaults to 1 if not provided. Starts at 0 if key doesn't exist. Returns the new value
- `exists(key)` - Check if key exists in store. Returns true/false
- `rename(oldKey, newKey)` - Atomically rename a key. Returns error if oldKey doesn't exist or newKey already exists
- `delete(key)` - Remove a key
- `clear()` - Remove all keys

### Array Operations
- `push(key, item)` - Atomically append to array. Creates array if key doesn't exist. Returns new length
- `shift(key)` - Atomically remove and return first element from array (FIFO dequeue). Returns nil if array is empty
- `shift_wait(key [, timeout])` - Block until array has items, atomically remove and return first element. Returns nil if timeout exceeded
- `pop(key)` - Atomically remove and return last element from array (LIFO pop). Returns nil if array is empty
- `pop_wait(key [, timeout])` - Block until array has items, atomically remove and return last element. Returns nil if timeout exceeded
- `unshift(key, item)` - Atomically prepend item to array. Creates array if key doesn't exist. Returns new length

### Wait & Blocking
- `wait(key [, expectedValue] [, timeout])` - Block until key changes (no expectedValue) or equals expectedValue. Optional timeout in seconds
- `wait_for(key, predicate [, timeout])` - Block until predicate(value) returns true. For arrays, predicate receives length. Optional timeout in seconds

### Expiration
- `expire(key, ttlSeconds)` - Set time-to-live for a key in seconds. Key automatically deleted when TTL expires. Re-calling resets the timer. Default TTL is 60 minutes. Returns error if key doesn't exist

### Persistence
- `save()` - Explicitly save to disk (requires persist configured)
- `load()` - Explicitly load from disk (requires persist configured)

### Inspection
- `keys()` - Get array of all keys in the store

## Context

Datastores are namespaced globally - all scripts in the same process accessing the same namespace share the same store. This enables coordination patterns without shared memory.

## Examples

### Worker Swarm Coordination

Orchestrate multiple spawned scripts:

```duso
// Setup
store = datastore("swarm_job_123")
store.set("worker_count", 0)
store.set("completed", 0)

// Spawn 5 workers
for i = 1, 5 do
  spawn("worker.du", {job_id = "swarm_job_123", worker_id = i})
end

// Wait for all to finish
store.wait("completed", 5)
print("All workers done!")
```

```duso
// worker.du - spawned worker script
ctx = context()
job_id = ctx.request().job_id
worker_id = ctx.request().worker_id

store = datastore(job_id)
store.increment("worker_count", 1)

// Do work...
print("Worker " + format_json(worker_id) + " working...")

// Mark done
store.increment("completed", 1)
```

### Append to Shared Array

Collect results from multiple workers atomically:

```duso
store = datastore("results")

for i = 1, 3 do
  spawn("collector.du", {job = "results"})
end

// Wait until 3 items collected
store.wait_for("items", fn(len) => len == 3)
print("All results: " + format_json(store.get("items")))
```

```duso
// collector.du
store = datastore("results")
store.push("items", {worker = 1, result = 42})
```

### Persistent Coordination State

Save state to disk for recovery:

```duso
store = datastore("app_state", {
  persist = "state.json",
  persist_interval = 60  // Auto-save every 60 seconds
})

// Load from disk if exists, or start fresh
store.set("session_id", "sess_123")
store.increment("request_count", 1)

// On shutdown, save() is called automatically
// Manual save if paranoid:
store.save()
```

### Custom Predicate for Wait

Wait until condition is met on value:

```duso
store = datastore("metrics")
store.set("temperature", 25)

// Background script updates temperature
spawn("temperature_monitor.du")

// Wait until temperature drops below threshold
threshold = 20
store.wait_for("temperature", fn(temp) => temp < threshold)
print("Temperature is now safe")
```

### Atomic Counters with Increment/Decrement

Maintain counters safely with default and custom deltas:

```duso
store = datastore("counters")
store.set("requests", 0)
store.set("active", 0)

// Increment by 1 (default) - returns new value
count = store.increment("requests")
print(count)  // 1

// Increment by custom amount
count = store.increment("requests", 10)
print(count)  // 11

// Increment and track
store.increment("active")
print(store.get("active"))  // 1

// Decrement by 1 (default) - returns new value
count = store.decrement("active")
print(count)  // 0

// Decrement by custom amount
count = store.decrement("requests", 5)
print(count)  // 6
```

### Work Queue with shift_wait (Blocking Consumer)

Distribute work atomically with blocking consumer:

```duso
// Producer
store = datastore("work_queue")
store.push("jobs", {id = 1, task = "process_data"})

// Worker (blocks until job available)
store = datastore("work_queue")
while true do
  job = store.shift_wait("jobs", 5)  // Wait up to 5 seconds for job
  if job == nil then
    print("No jobs - timeout after 5 seconds")
    break
  end
  print("Got job: " + format_json(job))
end
```

No race conditions—`shift_wait()` atomically waits for items and removes them in one operation.

### Work Queue with Non-Blocking shift

Simple non-blocking pattern for polling:

```duso
// Producer
store = datastore("work_queue")
for i = 1, 10 do
  store.push("jobs", {id = i, data = "job_" + i})
end

// Worker (non-blocking, checks periodically)
store = datastore("work_queue")
while true do
  job = store.shift("jobs")  // Returns nil if empty
  if job == nil then break end
  print("Processing: " + format_json(job))
end
```

### Session Expiration with TTL

Implement session timeouts using automatic expiration:

```duso
store = datastore("sessions")

// Create session
session_id = "sess_abc123"
store.set(session_id, {user = "alice", created = now()})
store.expire(session_id, 3600)  // Expire in 1 hour

// On each request, refresh the session
store.expire(session_id, 3600)  // Reset the 1-hour timer

// Check if session still exists
if store.exists(session_id) then
  print("Session active")
else
  print("Session expired")
end
```

### Atomic Inbox with Swap

Agent receives messages and consumes them atomically:

```duso
// Orchestrator sends messages
store = datastore("agents")
agent_id = "agent_1"
store.push(agent_id + "_inbox", {msg = "hello"})
store.push(agent_id + "_inbox", {msg = "world"})

// Agent consumes all messages atomically
messages = store.swap(agent_id + "_inbox", [])
for msg in messages do
  print(msg.msg)
end
```

## Atomicity

All operations are atomic at the key level. Multiple operations on same key from different scripts won't interfere:

**Value Operations**
- `set(key, value)` - Atomic write
- `set_once(key, value)` - Atomic read-check-write
- `swap(key, newValue)` - Atomic read-old-write-new-return-old
- `increment(key [, delta])` - Atomic read-add-write
- `decrement(key [, delta])` - Atomic read-subtract-write
- `rename(oldKey, newKey)` - Atomic move operation

**Array Operations**
- `push(key, item)` - Atomic append
- `shift(key)` - Atomic remove-first
- `shift_wait(key [, timeout])` - Atomic wait-and-remove-first
- `pop(key)` - Atomic remove-last
- `pop_wait(key [, timeout])` - Atomic wait-and-remove-last
- `unshift(key, item)` - Atomic prepend

**Lifecycle**
- `expire(key, ttlSeconds)` - Atomic TTL set (re-calling resets timer atomically)

Example: Two scripts calling `swap()` on same key won't lose values - one gets old value, other gets its previous old value.

## Wait Semantics

**wait(key)** - Blocks until value changes from initial state (detects new appends, value updates)

**wait(key, expectedValue)** - Blocks until key equals expectedValue (useful for status flags)

**wait_for(key, predicate)** - Blocks until predicate returns true

For **arrays**, predicates receive the array **length** (as number), not the array itself:

```duso
store.wait_for("items", fn(len) => len >= 10)  // len is a number
```

For non-arrays, predicates receive the value:

```duso
store.wait_for("status", fn(val) => val == "complete")  // val is a string
```

## Persistence

If `persist` is configured:

- **Auto-load**: Datastore loads from disk when first created (if file exists)
- **Auto-save**: If `persist_interval` set, saves every N seconds in background
- **Manual save**: Call `store.save()` for paranoid writes
- **Shutdown**: On process exit (Ctrl+C), final save happens

JSON format preserves all Duso types (arrays, objects, numbers, strings, booleans).

## Timeout on Wait

All wait methods support optional timeout (last parameter):

```duso
// Wait up to 5 seconds for value to equal "done"
store.wait("status", "done", 5)

// Wait up to 10 seconds for predicate
store.wait_for("items", fn(len) => len > 0, 10)
```

Returns error if timeout exceeded without condition met.

## Thread Safety

- All operations are thread-safe
- Multiple goroutines (spawned scripts) can safely access same namespace
- No race conditions on read or write
- Condition variables efficiently wake up waiting goroutines on writes

## Concurrency Pattern

Ideal for agent swarms and worker coordination:

```
Main Script
  ├─ Creates datastore("job_id")
  ├─ Spawns 10 workers
  ├─ Calls store.wait("completed", 10)  [blocks]
  │
  └─ Workers (concurrent)
      ├─ Each calls store.increment("completed", 1)
      ├─ One worker's increment broadcasts
      └─ Main script wakes up when all 10 done
```

Zero-overhead signaling - no polling, just efficient condition variable wakeups.

## Notes

- **Namespacing is global**: Calling `datastore("name")` multiple times returns the same cached instance. Pass config only on first call
- **Persistence is opt-in**: No `persist` config = in-memory only. With `persist` = auto-load on first creation and optional auto-save
- **Process lifetime**: Datastores persist in memory for the lifetime of the Duso process only. Restart requires re-creating with `persist` config to reload
- **Namespace collision**: Collision between swarms will cause them to share state (usually a bug - use unique namespaces)
- **Script functions**: Not yet supported as predicates in `wait_for()` (only Go functions work)
- **No ACID**: Simple last-write-wins semantics, no transactions
- **Array deletes**: Not supported (just clear() and rebuild if needed)

## See Also

- [spawn() - Run script asynchronously](/docs/reference/spawn.md)
- [run() - Run script synchronously](/docs/reference/run.md)
- [context() - Access request context](/docs/reference/context.md)
