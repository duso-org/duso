# Distributed Datastore: Local / Replicated / Remote Modes

## Summary

Evolve the `datastore()` builtin into a distributed-capable store while keeping the same duso API. Three configurable modes per namespace:

- **local** — current behavior, in-process only (default, no config change needed)
- **replicated** — this duso server participates as a Raft consensus node; data replicated across a cluster for durability and HA
- **remote** — thin RPC client only; routes reads/writes to the cluster that owns this namespace

Different datastores on the same server can use different modes:

```duso
// This server is a full Raft participant for access_tokens
tokens = datastore("access_tokens", {
  mode  = "replicated",
  peers = ["server1:7777", "server2:7777", "server3:7777"]
})

// This server is a thin client to a dedicated user DB cluster
users = datastore("users", {
  mode  = "remote",
  peers = ["db1:7777", "db2:7777", "db3:7777"]
})

// Local scratch, unchanged
temp = datastore("scratch")
```

**No external dependencies.** Go stdlib + golang.org/x packages only.

---

## RPC Port Model

**One global port per duso process** (default `:7777`). All replicated/remote datastores on a server share one TCP listener. Messages are routed by namespace name inside the protocol payload.

Isolation between clusters comes from the **peer list**, not the port. Example:

- `access_tokens` cluster: s1:7777, s2:7777, s3:7777
- `users` cluster: db1:7777, db2:7777, db3:7777

The DB machines only know about `users`. They reject/ignore Raft messages for namespaces they don't have registered. One firewall rule per server, one config option.

---

## Deployment Example

```duso
// users-db.du — runs on dedicated DB machines
store = datastore("users", {
  mode  = "replicated",
  peers = ["db1:7777", "db2:7777", "db3:7777"]
})
// participates in Raft cluster, serves client requests, nothing else

// app-server.du — runs on application servers
tokens = datastore("access_tokens", {
  mode  = "replicated",
  peers = ["s1:7777", "s2:7777", "s3:7777"]
})
users = datastore("users", {
  mode  = "remote",
  peers = ["db1:7777", "db2:7777", "db3:7777"]
})
```

---

## Critical Files to Modify

| File | Change |
|---|---|
| `pkg/runtime/datastore.go` | Extract interface, update `GetDatastore` to route by mode |
| `pkg/runtime/builtin_datastore.go` | Use `DatastoreBackend` interface instead of `*DatastoreValue` |
| `pkg/runtime/register.go` | No changes needed |

---

## Phase 1: Interface Extraction (Refactor Only)

**Goal:** Decouple all callers from the concrete `*DatastoreValue`. Zero behavior change.

### New file: `pkg/runtime/datastore_iface.go`

```go
type DatastoreBackend interface {
    Set(key string, value any) error
    SetOnce(key string, value any) bool
    Get(key string) (any, error)
    Swap(key string, newValue any) (any, error)
    Increment(key string, delta float64) (any, error)
    Push(key string, item any) (float64, error)
    Pop(key string) (any, error)
    Shift(key string) (any, error)
    ShiftWait(key string, timeout time.Duration) (any, error)
    PopWait(key string, timeout time.Duration) (any, error)
    Unshift(key string, item any) (float64, error)
    Wait(key string, expectedValue any, hasExpected bool, timeout time.Duration) (any, error)
    WaitWithPredicate(eval *Evaluator, key string, fn Value, timeout time.Duration) (any, error)
    Delete(key string) (any, error)
    Clear() error
    Keys() []string
    Exists(key string) bool
    Rename(oldKey, newKey string) error
    Expire(key string, ttlSeconds float64) error
    Save() error
    Load() error
    Shutdown() error
    IsReadonly() bool
    ReturnDeletedValue() bool
}
```

### `pkg/runtime/datastore.go`

- `GetDatastore` returns `DatastoreBackend` instead of `*DatastoreValue`
- `datastoreRegistry` becomes `map[string]DatastoreBackend`
- Routes on `config["mode"]`:

```go
func GetDatastore(namespace string, config map[string]any) DatastoreBackend {
    switch mode {
    case "replicated":
        store = newReplicatedDatastore(namespace, config)
    case "remote":
        store = newRemoteDatastore(namespace, config)
    default: // "local" or unset
        store = newLocalDatastore(namespace, config) // existing path, renamed
    }
}
```

### `pkg/runtime/builtin_datastore.go`

- `store *DatastoreValue` → `store DatastoreBackend`
- Readonly/returnDeletedValue via interface methods `IsReadonly()` / `ReturnDeletedValue()`
- No logic changes

---

## Phase 2: Raft Core (`pkg/raft/`)

Pure Go stdlib. ~1500 LOC.

```
pkg/raft/
  types.go     — LogEntry, AppendEntries/RequestVote req+resp, DSRequest/DSResponse
  log.go       — Persistent append-only binary log + metadata file
  rpc.go       — TCP listener + connection pool; length-prefixed JSON framing
  node.go      — Raft state machine (election, replication, commit, applyCh)
  snapshot.go  — Periodic full-state snapshot + log truncation
```

### Wire Protocol

Length-prefixed JSON over TCP:
```
[4 bytes big-endian: message length][JSON payload]
```

Message envelope:
```go
type Message struct {
    Type string          // "AppendEntries", "RequestVote", "DSRequest", "DSResponse", etc.
    Body json.RawMessage
}
```

Same listener, same port handles both Raft RPCs and DS client requests — discriminated by `Type`.

### Key Types

```go
type LogEntry struct {
    Term      uint64
    Index     uint64
    Namespace string
    Op        string  // "set", "delete", "push", "increment", etc.
    Key       string
    ValueJSON []byte
}

type DSRequest struct {
    ID        string
    Namespace string
    Op        string
    Key       string
    ValueJSON []byte
    Extra     map[string]any // delta for increment, ttl for expire, etc.
}

type DSResponse struct {
    ID        string
    ValueJSON []byte
    Error     string
    NotLeader bool
    Leader    string // redirect hint when NotLeader=true
}
```

### Log Storage (binary, no deps)

```
Per entry: [8b term][8b index][4b ns len][ns][4b op len][op][4b key len][key][4b val len][val]
```

Separate metadata file for `currentTerm` + `votedFor` (must persist across crashes).
Both files synced to disk (`file.Sync()`) on every write.

### Raft Node Goroutines

- `runElectionTimer()` — randomized timeout; resets on valid heartbeat; triggers `RequestVote` on fire
- `runHeartbeat()` — leader only; sends `AppendEntries` (empty = heartbeat) on tick (~50ms)
- `runApply()` — advances `lastApplied` → `commitIndex`, pushes entries to `applyCh`
- `handleRPC()` — dispatches incoming `Message` to correct handler

Client-facing:
- `Propose(entry LogEntry) (commitIndex uint64, err error)` — leader only; waits for majority ack
- Returns `ErrNotLeader + leaderID` if not leader (client retries with hint)

### Snapshots

- Triggered every N entries or T seconds (configurable)
- Full JSON dump of state machine + `{snapshotIndex, snapshotTerm}`
- On startup: load snapshot → replay log from `snapshotIndex+1`
- After snapshot: truncate log entries before `snapshotIndex`

---

## Phase 3: Replicated Datastore (`pkg/runtime/datastore_replicated.go`)

Implements `DatastoreBackend`. Wraps `LocalDatastoreValue` (state machine) + `raft.Node`.

### Write Path

```
script → ds.Set("k", v)
  → encode as LogEntry{Op:"set", Key:"k", ValueJSON:...}
  → raft.Propose(entry) — blocks until majority commit
  → applyLoop reads applyCh on every node
  → calls local.Set("k", v) → updates in-memory state + sync.Cond.Broadcast()
  → return to script
```

### Read Path

```
script → ds.Get("k")
  → reads local.Get("k") directly (stale reads ok by default)
  → if strong_reads=true and this node is follower: RPC to leader
```

### Blocking Ops (`wait`, `shift_wait`, `pop_wait`)

Work unchanged: after any node's `applyLoop` applies a write, it calls `sync.Cond.Broadcast()` on the local state machine. Local waiters wake up. No cross-node signaling needed — every node applies every commit.

### Config

```duso
store = datastore("access_tokens", {
  mode              = "replicated",
  peers             = ["server1:7777", "server2:7777", "server3:7777"],
  self              = "server1:7777",   // optional: auto-detected from hostname
  log_dir           = "/var/duso/raft", // Raft log + snapshot storage
  strong_reads      = false,            // optional: route reads to leader
  snapshot_interval = 300               // seconds (default 300)
})
```

---

## Phase 4: Remote Datastore (`pkg/runtime/datastore_remote.go`)

Thin client. Implements `DatastoreBackend`. Zero local state.

### Request Flow

```
script → ds.Set("k", v)
  → DSRequest{Op:"set", Namespace:"users", Key:"k", ValueJSON:...}
  → probe peers to find leader (cached; invalidated on NotLeader response)
  → send to leader, wait for DSResponse
  → decode, return to script
```

### Leader Discovery

1. On startup: probe each peer in order, first non-redirect is leader
2. Cache leader address
3. On `NotLeader=true`: use `Leader` hint, retry immediately
4. On connection failure: rotate to next peer

### Blocking Ops

Remote `wait`/`shift_wait`/`pop_wait`: client sends DSRequest with timeout to leader. Leader executes the blocking op locally (holds `sync.Cond`), responds when condition met or timeout expires.

### Config

```duso
store = datastore("users", {
  mode  = "remote",
  peers = ["db1:7777", "db2:7777", "db3:7777"]
})
```

---

## Phase 5: Global RPC Server (`pkg/raft/rpc_server.go`)

One `net.Listener` per duso process. Started lazily on first replicated or remote datastore creation.

Routes by message `Type` and `Namespace`:
- Raft RPCs → correct `raft.Node` instance (looked up by namespace)
- `DSRequest` → correct `DatastoreBackend` (looked up by namespace)

```go
var globalRPCServer *RPCServer  // in pkg/runtime/register.go

func ensureRPCServer(port int) *RPCServer { ... } // idempotent
```

CLI flag: `-raft-port` (default `7777`)

---

## Phased Delivery

| Phase | What ships | Key validation |
|---|---|---|
| 1 | Interface extraction | Existing scripts build and run unchanged |
| 2 | `pkg/raft/` core | Unit tests: election, replication, leader failover |
| 3 | `ReplicatedDatastore` + RPC server | 3-node cluster; write survives node kill |
| 4 | `RemoteDatastore` + leader discovery | Thin client reads/writes through cluster |
| 5 | Config polish, `-raft-port` flag, doc updates | `duso -doc datastore` covers all three modes |

---

## Verification

**Phase 1 — no regression:**
```bash
./build.sh
duso examples/http_server.du
```

**Phase 2 — Raft unit test sketch:**
```go
// pkg/raft/node_test.go
// 1. Start 3 in-process nodes
// 2. Wait for leader election
// 3. Propose 100 entries, verify all committed on all nodes
// 4. Kill leader
// 5. Verify new election completes in <2s
// 6. Propose 10 more entries
// 7. Restart killed node, verify it catches up
```

**Phase 3 — integration:**
```bash
# Three terminals, one per node
duso -c '
  store = datastore("test", {
    mode="replicated", self="localhost:7777",
    peers=["localhost:7777","localhost:7778","localhost:7779"],
    log_dir="/tmp/raft-7777"
  })
  store.set("x", 42)
  print(store.get("x"))
'
# Kill leader terminal, verify another takes over
# Restart it, verify it rejoins and sees x=42
```

**Phase 4 — remote client:**
```bash
# With cluster running from Phase 3:
duso -c '
  store = datastore("test", {
    mode  = "remote",
    peers = ["localhost:7777","localhost:7778","localhost:7779"]
  })
  print(store.get("x"))  // expects 42
  store.set("y", 99)
  print(store.get("y"))  // expects 99
'
```
