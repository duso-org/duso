# Datastore Primer (LLM-Optimized)

Condensed reference for `datastore()`. Companion to `docs/duso-primer.md`; exhaustive
per-symbol docs in `docs/reference/datastore.md` and `docs/reference/datastore_replication.md`.

## What it is

A namespaced, thread-safe store shared by every script in the process. It covers three
jobs that would otherwise be three dependencies:

- **Coordination** — atomic counters, blocking waits, work queues. No locks in user code.
- **Document store** — objects as values, atomic deep-merge updates, predicate queries.
  Closer in shape to MongoDB than to a key/value cache.
- **Durable state** — snapshot + write-ahead log, TTL expiry, encryption at rest,
  leader/follower replication.

Everything is in-process and in-memory; persistence is a config option, not a server.

## Gotchas (read first)

- **Configure in exactly one place.** Pass `config` once (typically the startup script);
  open it everywhere else as a bare `datastore(ns)`. Passing config to an already-configured
  namespace **throws** — it would re-run recovery against a live store.
- **Crossing the boundary deep-copies.** Values written and read are independent copies,
  and **functions are stripped to `nil`**. Regex literals survive as plain strings.
- **`wait()` throws on timeout**; `shift_wait()`/`pop_wait()` return `nil` on timeout.
  Different failure modes on purpose — wrap `wait()` in try/catch.
- **A failed load is fatal.** If the snapshot or WAL can't be read, `datastore()` throws
  rather than starting empty. Starting empty would mean serving nothing and then
  overwriting the real data at the next snapshot.
- **`select()`'s `max=N` is not "the first N".** Map iteration order is non-deterministic,
  so it returns *any* N matches.
- **Values are capped at 64MB** by default (`max_value_size`). Recovery ignores the cap —
  lowering it never makes an existing store unopenable.

## Opening

```duso
store = datastore("app")                          // in-memory, nothing on disk

store = datastore("app", {                        // durable
  persist          = "/var/lib/app/db.dusnap",    // snapshot (duso binary format)
  persist_interval = 60,                          // background snapshot every 60s
  wal              = "/var/lib/app/db.duwal",     // write-ahead log
  wal_sync_interval = 0.1,                        // fsync cadence; 0 = every write
  max_value_size   = 67108864,                    // 64MB default; 0 disables
  encrypt_key      = env("DB_KEY"),               // base64 32 bytes -> AES-256-GCM at rest
  readonly         = false                        // true = load, never write; writes throw
})
```

Relative paths resolve against the working directory — use absolute paths in production.

## Key/value

```duso
store.set("k", value)                     // any duso value, including binary
store.set_once("k", value)                // true if set, false if key already existed
v = store.get("k")                        // nil if absent
old = store.swap("k", newValue)           // atomic exchange, returns the previous value
store.increment("hits")                   // delta defaults to 1; starts from 0
store.increment("hits", 5)                // returns the new value
store.decrement("credits", 2)
store.exists("k")                         // boolean
store.rename("old", "new")                // throws if old is missing or new exists
store.delete("k")
store.clear()
store.keys()                              // array of every key
```

## Objects: atomic deep merge

`update()` is the read-modify-write you would otherwise have to guard with a lock. One
lock, one merge, no lost updates between concurrent writers.

```duso
store.set("cfg", {
  version  = "1.0",
  features = {search = true, export = false},
  limits   = {rpm = 100}
})

store.update("cfg", {version = "1.1", features = {export = true}})  // worker 1
store.update("cfg", {limits = {rpm = 200}})                          // worker 2

// -> {version="1.1", features={search=true, export=true}, limits={rpm=200}}
```

A `nil` value deletes that key (shallow, at the merge level):

```duso
store.update("cfg", {features = {deprecated = nil}})
```

Creates an empty object if the key is absent. Throws if the key exists and is not an object.
Returns the merged object.

## Queries: select and count

```duso
// select(predicate [, max=N]) -> array. Return a value to include, nil to exclude.
done = store.select(function(key, value)
  if value.status == "done" then return value.count end
end)

// transform while filtering
rows = store.select(function(key, value)
  if value.count > 4 then return {name = key, doubled = value.count * 2} end
end)

// filter on the key
users = store.select(function(key, value)
  if starts_with(key, "user_") and value.active then return value end
end)

// stop early — "find any one match" without scanning the store
one = store.select(function(key, value)
  if value.status == "pending" then return value end
end, max = 1)

// count(predicate) -> number. Truthy return counts.
n = store.count(function(key, value) return value.status == "done" end)
```

Both accept the predicate positionally or by name (`select(predicate = fn)`).

`count()` is cheaper than `len(select(...))` — no result array is built and nothing is
deep-copied. `select()` results *are* deep-copied. Iteration locks briefly per key rather
than holding the store lock, so a large store stays responsive during a scan. A predicate
that throws propagates out of `select()` — wrap in try/catch.

There is no index: both are O(n) scans. That is fine for the store sizes duso targets, but
if a lookup is hot, maintain your own key (`"user_by_email:" + email`) and `get()` it.

## Arrays and queues

```duso
store.push("jobs", job)                   // append, creates the array; returns new length
store.unshift("jobs", job)                // prepend
job = store.shift("jobs")                 // FIFO dequeue, nil if empty
job = store.pop("jobs")                   // LIFO pop, nil if empty
job = store.shift_wait("jobs", 10)        // block up to 10s, nil on timeout
job = store.pop_wait("jobs", 10)          // LIFO variant
```

`push()`/`unshift()` size-check **the item**, not the resulting array — filling a queue
stays linear.

Drain an inbox atomically with `swap()`:

```duso
messages = store.swap("agent_1_inbox", [])
for msg in messages do process(msg) end
```

## Waiting

```duso
store.wait("status")                                       // until the value changes
store.wait("completed", 10)                                // until it equals 10
store.wait("completed", 10, timeout = 30)                  // THROWS on timeout
store.wait("temp", function(v) return v >= 20 end, 30)     // predicate form
```

Returns the current value on success. Condition-variable wakeups, not polling — a worker
swarm that increments a counter wakes the coordinator with zero overhead.

```duso
// coordinator
store = datastore("job_1")
store.set("completed", 0)
for i in range(10) do spawn("worker.du", {n = i}) end
store.wait("completed", 10, timeout = 300)

// worker.du
datastore("job_1").increment("completed", 1)
```

Prefer the predicate form over a polling loop for anything non-trivial.

## Expiry

```duso
store.set(sid, {user = "alice", created = now()})
store.expire(sid, 3600)                   // delete in 1 hour; re-calling resets the timer
```

Default TTL is 60 minutes. Throws if the key doesn't exist. Deadlines are absolute and
survive restarts on a persisted store: a key given an hour expires an hour after
`expire()` was called, not an hour after recovery.

## Durability

Writes go to the WAL before memory. Recovery is snapshot + replay of post-snapshot WAL
entries. Every mutation is logged — including `shift_wait()`, `rename()`, `clear()` and
`expire()` — so a consumed queue item stays consumed.

`wal_sync_interval` defers only the **fsync**, not the write itself. The kernel already has
the bytes:

| failure | default (`0.1`) | `wal_sync_interval = 0` |
|---|---|---|
| process crash: panic, OOM kill, `kill -9` | loses nothing | loses nothing |
| machine loss: power cut, kernel panic | loses up to 100ms | loses nothing |

Fsync-per-write costs roughly three orders of magnitude of write throughput (~300/sec vs
~400,000/sec on the same machine). Use `0` for ledgers and audit logs; the default
everywhere else.

A process that dies mid-write leaves a partial trailing record. That write was never
acknowledged, so recovery discards it and keeps everything before — normal, not an error.
Damage *inside* the log fails startup loudly.

Manual control:

```duso
store.save()                              // force a snapshot now
store.load()                              // force a reload from disk
```

On process exit — signal or the script simply finishing — the WAL is synced and a final
snapshot is written. With an HTTP server running this happens *after* in-flight requests
drain, so a write from a handler that was still finishing is included.

## Encryption at rest

```duso
store = datastore("app", {
  persist     = "/var/lib/app/db.dusnap",
  wal         = "/var/lib/app/db.duwal",
  encrypt_key = env("DB_KEY")             // base64 of exactly 32 bytes
})
```

Snapshot encrypted whole; each WAL entry encrypted with its own nonce. Memory holds
plaintext, so queries are unaffected. The key never touches disk. Rotation is manual
(export, change key, re-import). Backups are useless without the key.

## Replication (experimental)

Configuration only — no new builtins. One **leader** owns the data; any number of
**followers** apply its writes and serve reads locally at full speed.

```duso
// leader
store = datastore("app", {
  persist = "/var/lib/app/db.dusnap",
  wal     = "/var/lib/app/db.duwal",
  replicate_listen = "0.0.0.0:7777",
  replicate_secret = env("REPL_SECRET")
})

// follower — same namespace, that's how the leader knows which store is wanted
store = datastore("app", {
  persist = "/var/lib/app/db.dusnap",       // optional but recommended: resume instead of full resync
  wal     = "/var/lib/app/db.duwal",
  replicate_from   = "ws://db1.internal:7777",
  replicate_secret = env("REPL_SECRET")
})

user = store.get("user:42")                 // local, full speed
store.set("user:42", {name = "Ada"})        // forwarded to the leader, applied there, streamed back
```

Set `replicate_listen` **or** `replicate_from`, never both. Any other `replicate_*` option
without one of them is an error, not a silent no-op. Both sides must agree on `encrypt_key`.
Add `replicate_cert_file`/`replicate_key_file` on the leader for `wss://`.

Monitoring works on any store, replicated or not:

```duso
st = store.replication_status()
// leader:     {role="leader", epoch=1, listen=..., seq=40, followers=1, buffered_bytes=..., buffered_frames=...}
// follower:   {role="follower", epoch=1, leader=..., connected=true, cursor=40}   // + last_error when disconnected
// unreplicated: {role="standalone"}

if st.role == "follower" and not st.connected then
  // serving stale reads — leader unreachable
end
```

There is no `lag` field, deliberately: a follower is never told the leader's sequence, so
any lag it reported would be a guess. Compare `cursor` across replicas — that's the number
that decides which one to promote.

Know before relying on it:

- **Asynchronous only.** A read on a follower can miss a write the leader just accepted.
  Anything needing a strictly current view reads from the leader. Lag is typically well
  under a millisecond on a LAN, but it is not zero.
- **No automatic failover**, by design. Promotion is a config change plus a restart.
- **No sharding, no write scaling.** One leader takes all writes.
- **Only datastore contents replicate.** Not files, not `spawn()`ed process state, not
  `schedule()`d jobs, not live websocket connections.

A follower that never gets promoted is a live off-box backup — a complete copy, seconds
behind, in duso's own format. No dump job, no export window. Read it from another script
with `readonly = true` to avoid disturbing it.

See `docs/reference/datastore_replication.md` for TLS, security, epochs, resync behavior
and the full failover procedure.

## Atomicity summary

Atomic at the key level, so concurrent scripts never interleave on the same key:
`set`, `set_once`, `swap`, `update`, `increment`, `decrement`, `rename`, `push`, `shift`,
`shift_wait`, `pop`, `pop_wait`, `unshift`, `expire`.

There are no multi-key transactions. If two keys must move together, put them in one
object and use `update()`.

## See also

- `docs/duso-primer.md` — the language
- `docs/http-primer.md` — `http_server()`, `fetch()`, `websocket()`
- `docs/reference/datastore.md` — full per-method reference
- `docs/reference/datastore_replication.md` — full replication reference
- `docs/datastore-1.7-performance.md` — throughput numbers
