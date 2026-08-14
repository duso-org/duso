# Datastore Replication

Stream a datastore's writes to one or more standby servers, for failover and continuous backup.

One server is the **leader**: it owns the data and serves the stream. Any number of **followers** connect to it, apply its writes as they happen, and serve reads locally at full speed. A write issued on a follower is forwarded to the leader and applied there, so the same code runs unchanged on either.

This is configuration only — there are no new builtins. A datastore becomes a leader or a follower based on the `replicate_*` options passed to [`datastore()`](/docs/reference/datastore.md).

## Configuration

| Option | Side | Description |
|---|---|---|
| `replicate_listen` | leader | Address to serve the replication stream on, e.g. `"0.0.0.0:7777"`. Setting this makes the store a leader |
| `replicate_from` | follower | The leader's URL, e.g. `"ws://db1.internal:7777"` or `"wss://..."`. Setting this makes the store a follower |
| `replicate_secret` | both | Shared secret. Required. On a follower, whichever secret it was given |
| `replicate_readonly_secret` | leader | A second secret granting the stream but not write forwarding |
| `replicate_buffer` | leader | Bytes of recent writes kept in memory so a briefly disconnected follower can resume without a full resync. Default 64MB |
| `replicate_cert_file` | leader | TLS certificate for the listener. Followers then connect with `wss://` |
| `replicate_key_file` | leader | TLS private key. Must be set together with `replicate_cert_file` |
| `replicate_ca_file` | follower | PEM file of the certificate authority to trust for `wss://`. Omit to use the system trust store |

Set `replicate_listen` or `replicate_from`, never both. Using any other `replicate_*` option without one of them is an error rather than a silent no-op — a store you believed was replicating but never was is the worst outcome available here.

Both sides must agree on `encrypt_key`: either both set the same key or neither sets one. The stream carries the leader's frames byte for byte, encryption included, so a mismatch is rejected at connect time instead of failing on every record.

### Leader

```duso
store = datastore("app", {
  persist = "/var/lib/app/db.dusnap",
  wal     = "/var/lib/app/db.duwal",

  replicate_listen = "0.0.0.0:7777",
  replicate_secret = env("REPL_SECRET")
})
```

### Follower

```duso
store = datastore("app", {
  persist = "/var/lib/app/db.dusnap",
  wal     = "/var/lib/app/db.duwal",

  replicate_from   = "ws://db1.internal:7777",
  replicate_secret = env("REPL_SECRET")
})

// Reads are local and full speed
user = store.get("user:42")

// Writes are forwarded to db1.internal, applied there, and streamed back
store.set("user:42", {name = "Ada"})
```

The namespace must match on both sides — it is how the leader knows which store a follower is asking for. One leader can replicate several namespaces over one `replicate_listen` address.

Giving a follower its own `persist` and `wal` is optional but recommended: without them a restart re-downloads the leader's entire state, and with them it resumes from where it left off.

### TLS

There is no `tls` flag. The transport is decided by two settings on two machines: the scheme in the follower's `replicate_from`, and whether the leader has `replicate_cert_file`/`replicate_key_file` set.

```duso
// leader
store = datastore("app", {
  replicate_listen    = "0.0.0.0:7777",
  replicate_secret    = env("REPL_SECRET"),
  replicate_cert_file = "/etc/duso/repl-cert.pem",
  replicate_key_file  = "/etc/duso/repl-key.pem"
})

// follower
store = datastore("app", {
  replicate_from    = "wss://db1.internal:7777",
  replicate_secret  = env("REPL_SECRET"),
  replicate_ca_file = "/etc/duso/internal-ca.pem"
})
```

Certificates are verified, and there is no option to skip verification. An internal leader will not have a publicly-trusted certificate, so point `replicate_ca_file` at the CA that issued it, or at the certificate itself if it is self-signed. Omit `replicate_ca_file` only when the leader's certificate chains to a public CA.

**The two sides must agree, and nothing negotiates it for them.** A mismatch fails to connect rather than falling back, and duso names the fix:

```
ws://  → TLS leader:       bad status — the leader is serving TLS;
                           change replicate_from to wss://db1.internal:7777
wss:// → plaintext leader: tls: first record does not look like a TLS handshake —
                           the leader is not serving TLS; either set
                           replicate_cert_file and replicate_key_file on the
                           leader, or change replicate_from to ws://...
```

Failing closed is the point: because there is no negotiation, a `wss://` follower cannot be talked down to plaintext by anything on the network.

## Security

**Use `wss://` for any link you do not physically control.** Plain `ws://` is appropriate only over loopback or an already-private network — WireGuard, a VPN, an internal VPC. Duso prints a startup warning if you configure plaintext replication to or from a non-loopback address.

What each configuration actually protects:

| | `ws://` | `ws://` + `encrypt_key` | `wss://` |
|---|---|---|---|
| Record values | plaintext | AES-256-GCM | TLS |
| `replicate_secret` | **plaintext** | **plaintext** | TLS |
| Record sizes and write rate | visible | visible | hidden |
| A fake leader feeding a follower | possible | bodies fail to decrypt | prevented |
| Tampering with a record in flight | possible | detected | prevented |

Points worth being precise about:

- **`replicate_secret` crosses the wire in the handshake.** Over `ws://` it is in the clear, and anyone who reads it gets a continuous live copy of the store. `encrypt_key` does not help — the secret is in the handshake, not in a record.
- **The secret is only sent after the transport is established.** A `wss://` follower will not leak it to a plaintext endpoint, even one impersonating your leader.
- **`encrypt_key` protects record values on the wire even over `ws://`**, because frames are shipped exactly as they were written. Frame headers are never encrypted, so sizes and timing stay visible.
- **The frame checksum is CRC32C, not a MAC.** It catches corruption, not an attacker. Over plain `ws://` without `encrypt_key`, an active attacker can rewrite a record and recompute the checksum.
- **The secret is a write credential, not just a read one.** Because followers forward writes, anyone holding `replicate_secret` and able to reach the port can both read the whole store continuously and mutate it. Treat it like a database password.
- **The listener has no connection limit.** Someone holding the secret can force repeated snapshot resyncs, each of which stalls leader writes.

Secrets are compared in constant time. Generate one with real entropy — `openssl rand -base64 32` — and keep it out of your scripts with `env()`.

## How it works

Duso's write-ahead log is already an operation log: every mutation produces a framed, checksummed, sequence-numbered record. Replication ships those exact frames.

```
leader: store.increment("n", 1)
  ── under the store lock ────────────────
     n: 10 → 11
     frame = [len | crc | seq 4021 | INCR | "n" | 1]
  ───────────┬───────────────────────────
             │  same bytes, two destinations
     ┌───────┴────────┐
     ▼                ▼
  wal file      followers
```

Two consequences worth knowing:

**The stream never reads the WAL file.** Frames are forked at the point they are created, not tailed from disk. Snapshotting, truncation and `save()` are invisible to replication, which is why enabling it changes nothing about how the log file behaves.

**Followers apply frames through the same code as crash recovery.** "Replaying my log after a crash" and "streaming from the leader" are one function with two sources.

Because the log records *operations* rather than resulting state, a `shift()` on a million-item queue is a ~20 byte frame, not a million-item array. It also means order is load-bearing: a follower applies every sequence number exactly once, in order, and a gap ends the session rather than being papered over.

### Catching up

When a follower connects it sends the sequence number it last applied. The leader either:

- **resumes the stream**, if that record is still in `replicate_buffer`; or
- **sends a full snapshot**, then streams from the watermark it covers.

A resync holds the store's lock for the length of the snapshot encode, which pauses leader writes. Both outcomes are logged so you can tell them apart:

```
duso: datastore "app": follower resumed at seq 41022 (leader at seq 41031, epoch 3)
duso: datastore "app": follower at seq 0 needs a full resync (leader at seq 41031, epoch 3)
```

Frequent resyncs mean `replicate_buffer` is too small for your write rate and reconnect frequency.

### What followers deliberately don't do

- **They don't expire keys on their own.** Expiry is the leader's decision and arrives as a record like any other write. A follower running its own sweep would delete keys the leader still holds.
- **They never apply their own writes.** A forwarded write is applied on the leader and comes back on the stream like any other. The follower applying it locally would place it at a sequence position that does not exist, and two replicas writing the same key would diverge for good.
- **`save()` and `load()` are not forwarded.** They act on this machine's files, not on the data. `load()` in particular would overwrite replicated state while the cursor kept pointing at the leader's sequence.

## Writes on a follower

A follower accepts every mutation the leader does. It does not apply them itself — it sends the operation to the leader, the leader applies it and computes the result, and the follower returns that result once the write has come back on the stream and been applied locally.

```duso
// This is the same code on a leader and on a follower.
n = store.increment("visits", 1)     // leader computes it, you get the number
job = store.shift_wait("jobs", 30)   // blocks on the leader until an item exists
```

Two consequences worth knowing:

**Read-your-writes holds.** By the time a forwarded call returns, the write has already been applied to local state — the reply is delivered after the frame it refers to. A read on the next line cannot miss it.

**Other nodes' writes are not instant.** You see them when their frames reach you, typically well under a millisecond on a LAN, but not zero. A follower is not a strongly consistent view of everyone else's writes.

### What it costs

| | latency |
|---|---|
| local write on the leader | ~2.5µs |
| forwarded, same AZ | ~0.5ms |
| forwarded, cross-AZ | ~2ms |
| forwarded, cross-region | 30–100ms |

That is 200–800× per call — but per-call is the wrong lens. One goroutine gets ~2k forwarded writes/sec sequentially, while a hundred concurrent ones get roughly the leader's own apply rate. A concurrent app barely notices; a serial loop over many writes falls off a cliff and should run on the leader.

Blocking forwards (`shift_wait`, `pop_wait`) park a goroutine on the leader for their duration. A follower is capped at 256 in-flight writes; past that it is told so rather than being allowed to exhaust the leader.

### Read-only replicas

A leader can hold two secrets. Which one a follower presents decides whether it may forward writes:

```duso
// leader
store = datastore("app", {
  replicate_listen          = "0.0.0.0:7777",
  replicate_secret          = env("REPL_RW"),   // may forward writes
  replicate_readonly_secret = env("REPL_RO")    // stream only
})
```

Follower config does not change — it sets `replicate_secret` to whichever one it was given. The leader reports the grant in its welcome, so a read-only replica fails writes locally and immediately rather than paying a round trip to be refused:

```
datastore("app"): this replica has read-only access to ws://db1:7777 — the write was not applied
```

`replication_status()` reports it as `can_write`, and the leader names it in its connection log:

```
duso: datastore "app": follower at seq 0 needs a full resync (leader at seq 1, epoch 1, read-only)
```

The check runs on both ends. The follower's is a convenience that saves a round trip; the leader's is the actual boundary.

**This is blast-radius control, not a trust boundary.** It stops the application on a replica — a bug, a bad deploy, a compromised process — from mutating the shared store. It does not constrain whoever controls that machine, because that box already holds a complete copy of the data and could stand it up as its own leader. Use it to keep a backup host from ever writing to production by accident; don't use it to replicate to a host you don't trust.

Promotion is unaffected: a read-only follower has the full dataset and is promotable like any other. Doing so makes it writable, which is correct — the grant described a connection, never the node.

### When a forward cannot complete

Three failures, and they say different things on purpose:

- **Not connected** — "the write was not applied". Definite.
- **Timed out** (30s for ordinary writes) — "may or may not have been applied".
- **Connection dropped in flight** — "may or may not have been applied".

The last two are genuinely ambiguous: the leader may have applied the write before the socket died, and nothing on the follower can tell. Duso says so rather than guessing. If that matters for a given operation, make it idempotent or issue it against the leader.


## Failover

Promotion is a config change and a restart: replace `replicate_from` with `replicate_listen`.

```duso
// was: replicate_from = "ws://db1.internal:7777"
replicate_listen = "0.0.0.0:7777"
```

The promoted store keeps everything it had replicated, starts accepting writes, and logs the new term:

```
duso: datastore "app" promoted to replication leader, epoch 3
```

**Duso does not decide when to promote.** There is no consensus and no automatic failover. Two followers that independently concluded the leader was dead would both promote, and two divergent operation logs cannot be merged — one of them has to be discarded wholesale. Avoiding that requires a majority quorum, which duso does not implement.

So the decision belongs outside duso, in one of two shapes:

- **An operator promotes.** Someone gets paged, checks, edits the config, restarts. Minutes of write downtime, zero infrastructure.
- **Something that already has a quorum promotes.** A floating IP, a load-balancer health check, Consul, etcd, a cloud lock. Whatever wins the lock restarts duso with the new config. Typically 10–30 seconds of write downtime. **This is the recommended production setup.**

Promote the follower with the **highest sequence number** — promoting a laggard discards more. Each follower prints its position when the stream ends, which is exactly when the leader has just died:

```
duso: datastore "app": replication stopped at seq 41031 (epoch 3) — this is the
      position to compare when choosing a replica to promote
```

### Epochs

Each leadership term has an **epoch**, recorded in the snapshot and WAL headers. It is bumped exactly once: when a store that was last written as a follower starts as a leader. An ordinary leader restart keeps its epoch.

That single rule is what makes a mistaken failover loud. A follower that has seen epoch 3 will refuse to stream from a leader still at epoch 2:

```
duso: datastore "app": follower has seen epoch 3 but this leader is epoch 2 —
      this leader has been superseded and must be reconfigured as a follower
```

Both sides log it and the follower keeps its newer data instead of silently reverting.

Epochs **detect** split-brain; they cannot prevent it. Nothing without consensus can. What they buy is that the failure mode is "one node refuses to serve and says why" rather than "two nodes diverge quietly for a week."

### What failover costs you

Replication is asynchronous. A leader acknowledges a write as soon as it is applied and logged locally — it does not wait for followers.

**Writes the old leader acknowledged but had not yet shipped are lost when another node is promoted.** This is the same bargain Postgres asynchronous replication makes. If that is unacceptable for a given store, replication is not the right tool for it on its own.

### After a failover

Point the remaining followers at the new leader. A follower that was *ahead* of the promoted node — it had received records the new leader never did — holds data from a timeline that no longer exists. An operation log has no partial rollback, so it takes a full snapshot resync:

```
duso: datastore "app": follower at seq 41044 needs a full resync (leader at seq 41031, epoch 3)
```

That is correct and automatic, but it means promotion costs a full state transfer on other nodes. Worth knowing before it happens at 3am.

The old leader rejoins as a follower — replace its `replicate_listen` with `replicate_from`.

## Monitoring

`replication_status()` reports what a store is doing. It works on every datastore, so the same script can run on a leader, a follower and a laptop without knowing which:

```duso
st = store.replication_status()

if st.role == "follower" and not st.connected then
  // serving stale reads — the leader is unreachable
end
```

**Leader:**

```duso
{role = "leader", epoch = 1, listen = "127.0.0.1:7991", seq = 40,
 followers = 1, buffered_bytes = 630, buffered_frames = 40}
```

**Follower:**

```duso
{role = "follower", epoch = 1, leader = "ws://db1:7777",
 connected = true, cursor = 40}
```

When a follower is disconnected it also carries `last_error` with the most recent failure:

```duso
{role = "follower", epoch = 1, leader = "ws://db1:7777", connected = false,
 cursor = 40, last_error = "websocket.Dial ...: connection refused"}
```

**Unreplicated:** `{role = "standalone"}`.

Two fields worth explaining:

- `followers` counts connections that completed the handshake, not connections that were opened — a brute-force attempt against `replicate_secret` does not inflate it.
- There is no `lag` field. A follower is never told the leader's current sequence, so any lag it reported would be a guess. Compare `cursor` across replicas instead; that's the number that decides which one to promote.

Everything else replication does is reported on stderr — connects, resyncs, refusals, and the position a stream stopped at. Configuration errors are ordinary duso errors thrown from `datastore()`, so `try`/`catch` works on them; failures in the background connection are not, because there is no script frame to throw into.

## Continuous backup

A follower that never gets promoted is a live off-box backup. It holds a complete copy, seconds behind, in duso's own snapshot format — no dump job, no export window, no separate tooling.

```duso
// backup.du — on a machine that does nothing else
datastore("app", {
  persist          = "/backup/app.dusnap",
  persist_interval = 300,
  wal              = "/backup/app.duwal",
  replicate_from   = "wss://db1.example.com:7777",
  replicate_secret = env("REPL_SECRET"),
  encrypt_key      = env("DB_KEY")
})
```

To read the backup without disturbing it, open the files from another script with `readonly = true`.

## Cost

**On the leader.** Building the frame is work the WAL already does; replication adds a slice append and a wakeup. Writes to the socket happen on separate goroutines and are coalesced, so a burst of writes is one send rather than one per record. Nothing in the write path blocks on a slow follower — a wedged replica falls off `replicate_buffer` and is disconnected, rather than becoming a wedged leader.

**Bandwidth** is the constraint worth sizing:

```
bytes/sec ≈ write_rate × average_value_size × follower_count
```

**On followers.** Reads cost nothing extra — they are local map reads at single-node speed. Applying a frame costs about what the original write cost.

**Replication lag** is typically well under a millisecond on a LAN. It is not zero: a read on a follower can miss a write the leader accepted moments ago. Anything needing a strictly current view has to read from the leader.

**The spiky cost is resync**, because the snapshot encode holds the store lock. Size `replicate_buffer` so that ordinary restarts and network blips resume from the stream instead.

## What isn't replicated

Only datastore contents. Not files your app writes, not `spawn()`ed process state, not `schedule()`d jobs, not in-flight websocket connections. A `sql()` connection points at an external database that has its own replication story.

## Limits

- No sharding and no write scaling. One leader takes all writes
- No automatic failover, by design — see [Failover](#failover)
- Asynchronous only: no option to make a write wait for a follower to acknowledge it

## See Also

- [datastore() - Create a datastore](/docs/reference/datastore.md)
- [env() - Read environment variables](/docs/reference/env.md)
