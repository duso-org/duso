# Datastore Value Codec, WAL v1, and Snapshot v1

Status: plan, not yet implemented. Supersedes the current gob-based WAL and
persist-file formats.

## Why

The current WAL is **state-based**: `WALEntry{Key, Value}` records the resulting
value, not the operation that produced it (`pkg/runtime/datastore.go:35`). That
choice produces four separate problems.

### 1. Array operations are O(n) per op, O(n²) to drain

`Push` writes the entire new array (`datastore.go:455`); `Shift` writes the
entire remaining array (`datastore.go:495`). Measured on this repo at v1.6.32,
5,000 `push` + 5,000 `shift` on a single queue of small objects:

| mode | push | shift | WAL bytes written |
|---|---:|---:|---:|
| in-memory | 978,880/s | 1,319,378/s | — |
| WAL, `wal_sync_interval=5` | **635/s** | **667/s** | **2.1 GB** |

For reference, WAL-backed `SET` benchmarks at ~76,000/s. Queue operations are
~100× slower than `SET` at the same durability setting, purely because of the
log format. `push()`/`pop_wait()` are documented as the coordination primitive
for spawned workers, so this is the hot path, not a corner.

### 2. Five mutations are never logged at all

Auditing `writeWAL` call sites against the mutation methods:

| mutation | logged? | consequence |
|---|---|---|
| `shift_wait()` / `pop_wait()` | no | consumed jobs resurrect after a crash |
| `rename()` | no | rename lost |
| `clear()` | no | cleared data resurrects |
| `expire()` | no | all TTLs lost across restart |
| `delete()` | as `{key, nil}` | replays as *key present with nil*, not absent |

The `delete()` encoding is ambiguous with `set(key, nil)`: `replayWAL` does
`ds.data[entry.Key] = entry.Value` unconditionally, so a deleted key comes back
existing-with-nil. `exists()`, `keys()`, `select()` and `count()` then disagree
with the pre-crash state.

Separately, `saveToDisk` encodes `ds.data` only — `expiryTimes` is never
persisted. Set a TTL, let a snapshot happen, restart, and the deadline is gone
and the key lives forever.

### 3. gob cannot represent duso's value model

`ValueToInterface` wraps binary, function, code, error and regex values as
`*ValueRef` (`pkg/script/value.go:455-465`), and `DeepCopyAny` passes them
through untouched, so they reach `ds.data` as opaque Go pointers. Observed:

```
SET binary   FAILED: gob: type not registered for interface: script.ValueRef
SET function FAILED: gob: type not registered for interface: script.ValueRef
```

The same script succeeds on a memory-only store (`type=binary len=12` round-trips
cleanly). So `store.set("avatar", load_binary("x.png"))` works or hard-fails
depending on whether a WAL is configured, and the failure leaks a raw Go error
into script-land.

### 4. gob is the wrong shape for self-contained records

Records must be independently decodable to support torn-tail recovery, log
tailing, and eventual replication. A `gob.Encoder` is a *stateful stream*: type
descriptors are emitted once at the head. Encoding each record with a fresh
encoder makes it self-contained but re-emits those descriptors every time.
Measured:

| | bytes |
|---|---:|
| `SHIFT` record, fresh encoder per record | 83 |
| `SHIFT` record, amortized in a shared stream | 15.7 |

~67 bytes of pure overhead per record, landing hardest on the small hot ops.

There is also a latent framing hazard: `openWALForWrites` opens `O_APPEND` and
creates a new `gob.Encoder`. Recovery normally truncates first, but if
`truncateWAL` fails (currently logged as a warning, not fatal) a second gob
stream is appended onto a non-empty file, which does not decode.

## Scope decisions (settled)

| | types | handling |
|---|---|---|
| Persist fully | nil, bool, number, string, array, object, binary, error, regex | encode directly |
| Persist as source | code | store `Source` + `Metadata`, re-parse on load |
| Not persisted | function | replaced with nil / omitted, **no error thrown** |

Rationale for the boundary cases:

- **`code`** — `CodeValue{Source, Program, Metadata}` already retains its source
  (`value.go:165`), so the AST is rebuildable by re-parsing. If the re-parse
  fails on load (source written under a different grammar), surface a clear
  error naming the key rather than degrading silently.
- **`error`** — `ErrorValue{Message Value, Stack string}` is fully
  reconstructible. `Message` is itself a `Value` and recurses through the codec;
  `Stack` is already a formatted string, not live pointers. Nothing is lost.
- **`regex`** — `RegexValue{Pattern, Compiled}`; store the pattern, recompile on
  load. There is no flags field.
- **`function`** — follows the existing `deep_copy` builtin precedent exactly
  (`builtin_deepcopy.go:18`), which is observable behavior today:
  - top-level function → `nil`
  - function in an **array** → `nil` (position preserved; dropping would shift indices)
  - function as an **object value** → **key omitted entirely** (line 33)

  Enforced at the datastore write boundary (`Set`/`Push`/`Swap`/`Update`/…), *not*
  inside `DeepCopy`/`DeepCopyAny` — those run at every scope boundary and
  functions legitimately cross them as arguments (`select(predicate)`,
  `wait(key, fn)`). Nothing internal stores functions today; `scheduleRecord`
  is plain data only (`builtin_schedule.go:408`).
- **`binary`** — stored **inline**, see below.

### Binary: inline, not a blob store

A content-addressed blob sidecar is not hard to build; it is hard to own.
Reference counting is the problem: a blob becomes unreferenced when a key is
overwritten by `set`, removed by `delete`, dropped by `clear`, aged out by
`expire`, shifted out of an array, or replaced by `update` — and the same blob
may be referenced from several keys or from inside nested arrays. Getting it
wrong leaks blobs or deletes live ones. It also breaks the deployment story:
persistence today is "one file, scp it," and a blob directory makes it "a file
plus a directory that must stay mutually consistent," dragging the
encryption-at-rest design along with it.

Inline costs less than it appears. The binary is already resident in `ds.data`,
so a 100MB binary is already 100MB of RAM on a box that idles at 5MB — the
memory ceiling binds before the disk ceiling does. Under the op-log it is
written to the WAL once, not rewritten per operation as arrays are today. For
the realistic range (avatars, thumbnails, uploads at 10KB–5MB) inline is fine.

Two guardrails:

1. A **configurable max value size** with a clear error, so nobody discovers
   500MB-through-fsync in production. Belongs with the other caps in
   `docs/ideas/limits.md`.
2. **Reserve a `BINARY_REF` tag now, unused.** A future blob store then becomes a
   purely additive format change — old readers reject it cleanly via the version
   field, new readers handle both.

`BinaryValue.Metadata` is a `map[string]Value`, so the codec encodes it
recursively with no special handling. If a storage id lands there later,
ignoring or regenerating it on load is a two-line rule.

## Part 1 — the value codec

One codec, shared by the WAL, the snapshot, and eventually the replication wire.
Tag byte + payload, recursive. Varints are LEB128 unsigned
(`encoding/binary.Uvarint`).

Tags are allocated in reserved ranges rather than sequentially, so the category
of an unknown tag is inferable and future additions never collide with
experimental use. **13 of 256 used.**

```
0x00        NIL         —

            --- scalars, 0x01-0x0F (5 used, 10 free) ---
0x01        FALSE       —                            bool folds into the tag
0x02        TRUE        —
0x03        NUMBER      f64, 8 bytes little-endian
0x04        INT         zigzag varint
0x05        STRING      varint len + utf8 bytes

            --- containers, 0x10-0x1F (2 used, 14 free) ---
0x10        ARRAY       varint count + N × value
0x11        OBJECT      varint count + N × (varint keylen + key + value)

            --- duso runtime types, 0x20-0x2F (4 used, 12 free) ---
0x20        BINARY      varint len + bytes + value(metadata object)
0x21        ERROR       value(message) + varint len + stack string
0x22        REGEX       varint len + pattern
0x23        CODE        varint len + source + value(metadata object)

            --- indirect / reference, 0x30-0x3F (1 used, 15 free) ---
0x30        BINARY_REF  RESERVED — not emitted, reject on decode

0x40-0xEF   reserved, future core types
0xF0-0xFE   reserved, experimental / vendor
0xFF        invalid — never a valid tag
```

`NIL` keeps `0x00` rather than reserving it as an invalid sentinel: it is the
most common value, and record-level CRC already catches the zero-filled-region
case that a sentinel would otherwise guard against. `0xFF` is reserved as
invalid at the top end, which costs nothing.

**`INT` selection rule.** Duso is float64-only, but the values in practice are
overwhelmingly small integers (ids, counts, indexes). Emit `INT` when:

```go
v == math.Trunc(v) && !math.IsInf(v, 0) && math.Abs(v) <= 1<<53
```

NaN falls through naturally (`NaN == Trunc(NaN)` is false); `IsInf` needs the
explicit guard. `{id=1, count=42}` drops from 16 bytes of numbers to 2. This is
the single largest size lever in the format.

**Object keys are encoded in sorted order.** Identical values then produce
identical bytes, which gives free checksums, dedup, and replica-divergence
comparison later. Costs an O(k log k) sort per object on the write path —
negligible for typical objects, and painful to retrofit.

**Placement.** `pkg/script/codec.go`, next to `Value` and `DeepCopyAny`, since it
operates on the `any` trees that package produces. Consumed by `pkg/runtime`.

## Part 2 — WAL v1

**File header, 16 bytes, written once:**

```
magic     [8]  "DUSOWAL\x00"
version   u16   1
flags     u16   bit0 = bodies encrypted
reserved  u32   0
```

Carrying the encrypted flag in the header means replay no longer needs the
config to know the framing — one replay path instead of today's two.

**Record frame:**

```
length    u32   body byte count
crc32c    u32   over body bytes
body      [length]
```

When encryption is enabled the **body is AES-GCM sealed** and `length`/`crc32c`
describe the sealed bytes, so torn-record detection works without the key.

**Record body:**

```
seq       varint
op        u8
          ...then exactly the fields the opcode table specifies, in order
```

**The opcode determines the record shape.** There is no field-presence bitmask:
the opcode table below is normative, and a decoder reads fields according to it.
This keeps records minimal and removes an entire class of malformed-record
handling (a presence mask can contradict its opcode; an opcode cannot contradict
itself).

Field encodings, when the opcode calls for them:

```
key       varint len + bytes
key2      varint len + bytes
num       f64, 8 bytes LE
value     codec-encoded
```

A `SHIFT` record is 7 bytes of body plus the 8-byte frame — 15 bytes total,
against an entire array copy today. `store.increment("hits", 1)` at seq 7:

```
07                        seq varint = 7
06                        op = INCR
04 68 69 74 73            key: len 4 + "hits"
00 00 00 00 00 00 F0 3F   num = 1.0
```

15 bytes of body, 23 with the frame.

**Growth lands on the opcode space, which is where the room is.** Because shape
follows opcode, adding an optional field to an existing operation means a new
opcode rather than a new flag bit. The path-access proposal in
`docs/ideas/new-datastore-ops.md` (`increment(["stats", "counters.requests"])`)
is the likely first case, and it would land in the reserved
`0x60-0x7F` future-data-ops range. With 242 free codes that is affordable, and
it is why the ranges are partitioned rather than assigned sequentially.

Compatibility is resolved by the **file format version plus the duso version**,
not by per-record negotiation.

**Opcodes.** Allocated in reserved ranges, same reasoning as the type tags.
**14 of 256 used.**

```
0x00        invalid — never a valid opcode

            --- key-value ops, 0x01-0x2F (9 used, 38 free) ---
```

| code | op | fields | notes |
|---:|---|---|---|
| `0x01` | `SET` | key, value | |
| `0x02` | `SET_ONCE` | key, value | logged only on success, so replay stays a blind apply |
| `0x03` | `DELETE` | key | own opcode — resolves the `set(key,nil)` ambiguity |
| `0x04` | `SWAP` | key, value | new value only; the caller's return is local |
| `0x05` | `UPDATE` | key, value | **the patch, not the merged result** |
| `0x06` | `INCR` | key, num | `decrement` folds in as a negative delta |
| `0x07` | `RENAME` | key, key2 | |
| `0x08` | `EXPIRE` | key, num | **absolute deadline, unix milliseconds** |
| `0x09` | `EXPIRED` | key | sweeper removed an expired key |

```
            --- array / queue ops, 0x30-0x4F (4 used, 28 free) ---
```

| code | op | fields | notes |
|---:|---|---|---|
| `0x30` | `PUSH` | key, value | |
| `0x31` | `UNSHIFT` | key, value | |
| `0x32` | `SHIFT` | key | also emitted by `shift_wait()` |
| `0x33` | `POP` | key | also emitted by `pop_wait()` |

```
            --- store-wide ops, 0x50-0x5F (1 used, 15 free) ---
0x50        CLEAR       —

0x60-0x7F   reserved, future data ops (indexes, sets)
0x80-0xBF   reserved, cluster / replication control
0xC0-0xFE   reserved, experimental / vendor — control/metadata only, never data
0xFF        invalid — never a valid opcode
```

`0x00` and `0xFF` are both reserved invalid, so a zero-filled or `0xFF`-filled
region fails opcode validation immediately instead of decoding as a plausible
operation — cheap robustness layered on top of the CRC.

**Unknown-opcode policy, which is what the ranges actually buy.** Because records
are length-framed, a reader can always skip a record it doesn't understand. Whether
it *should* is decided by the high bit alone — one test, no table lookup:

```
op <  0x80   data.      Unknown opcode is FATAL.
op >= 0x80   non-data.  Unknown opcode is SKIPPABLE.
```

Skipping an unrecognized data op would silently diverge state, so that case must
fail loudly and name the opcode. Skipping an unrecognized control record cannot
affect key-value state at all, so an older reader can ignore membership, term, or
handshake records written by a newer or cluster-aware node.

The cost of the single-bit rule is a hard constraint on the vendor range: **it
may not carry data mutations.** Anything at or above `0x80` is by definition safe
to ignore, so a vendor data op would be silently dropped by stock duso — and a
log that only one build can correctly apply is a worse outcome than not having
the extension point.

This matters on a single node too, not just in a cluster: write a WAL under a
newer duso, roll back to an older binary, and replay hits opcodes that did not
exist when that binary was built. The rule turns that from an opaque failure into
either a clean fatal error naming the opcode, or a safe skip.

Every opcode is a pure function of (current state, record) — no clocks, no
randomness during replay. That property is what makes replay deterministic now
and makes the log usable as a replication stream later. It is also why `EXPIRE`
must carry a deadline rather than a TTL: replaying a 60-second TTL three hours
after a crash must not resurrect the key.

`EXPIRED` applies identically to `DELETE` but records provenance, so a log reader
can distinguish a user deletion from a TTL expiry — useful for debugging and for
a replica that wants to account for expiry traffic separately.

## Part 3 — Snapshot v1

Currently a bare `gob.Encode(map[string]any)` with no header at all, so it has
the same "can't tell what version this is" problem the WAL had.

```
magic       [8]  "DUSOSNAP"
version     u16   1
flags       u16   bit0 = body encrypted
reserved    u32   0
seq         u64   WAL watermark this snapshot covers
keycount    u32
            N × (varint keylen + key + codec value)
expirycount u32
            N × (varint keylen + key + i64 unix millis deadline)
```

The header stays plaintext when encryption is on (so version and flags remain
readable); everything after it is sealed as one body, matching today's behavior.

The `seq` watermark earns its keep beyond tidiness: it makes WAL truncation
incremental and safe rather than all-or-nothing, and it is how a replica later
says "I have snapshot@seq, stream me from there." The expiry section fixes the
deadline-loss bug independently of anything else here.

**Naming.** The file is no longer gob, so the documented extension should change
(`.dsnap`). The extension is user-supplied in config (`persist = ".../db.gob"`),
so nothing is forced — this is a docs and examples change, not a code one.

## Part 4 — Expiry handling

- `expire(key, ttl)` → log `EXPIRE` with an **absolute** deadline in unix millis.
- Sweeper deletion (`sweepExpiredKeys`, `datastore.go:856`) → log `EXPIRED`.
  Applies identically to `DELETE`, but keeps the provenance distinguishable when
  reading a log.
- Lazy `checkExpired` (`datastore.go:879`) → **does not log.** It fires inside
  read paths like `Get`, and reads are currently free (~420k/s regardless of
  durability mode); logging there would make reads write to disk. The key is
  hidden locally and the sweeper produces the authoritative record within a
  second.
- Snapshot carries the expiry map; replay rebuilds both `expiryTimes` and
  `expiryHeap`.

The cost of this split is a sub-second window where the log lags observed local
state. That is acceptable because the sweeper is guaranteed to catch up and
writes the same deletion either way — and keeping the log authoritative for key
removal is what stops clock skew from becoming a correctness problem if this
ever becomes a replication stream.

## Part 5 — Migration

Read-and-convert, per the decision to phase out legacy readers in a later
release.

- **WAL:** if the first 8 bytes are not `"DUSOWAL\x00"`, use the legacy reader
  (both the plaintext gob-stream and the length-prefixed encrypted variants).
  Replay it, then write v1 from the next truncate onward.
- **Snapshot:** if the first 8 bytes are not `"DUSOSNAP"`, decode as a legacy bare
  gob map. The next `save()` writes v1.
- Legacy `delete()` records are `{key, nil}` and remain indistinguishable from
  `set(key, nil)`. Legacy replay preserves today's behavior (assign nil); this
  cannot be recovered retroactively and should be noted in release notes.
- Mark both legacy readers deprecated with a target removal release.

## Part 6 — Implementation phases

1. **Codec** (`pkg/script/codec.go`) + full round-trip tests. Standalone; no
   format change ships yet. — **done**
2. **Snapshot v1** — header, `seq`, expiry section, codec bodies, legacy reader.
   — **done** (`pkg/runtime/snapshot.go`). `walSeq` is carried and persisted but
   stays 0 until phase 3 populates it. Load failures now warn on stderr instead
   of silently yielding an empty store.
3. **WAL v1** — framing, CRC, seq, opcodes; log the five currently-unlogged
   mutations; `DELETE` gets its own opcode; legacy reader.
4. **Torn-tail recovery** — replay to the last intact record, truncate the
   remainder, continue. Replaces today's behavior of discarding the entire WAL
   on a single bad byte.
5. **Value-size limit** — coordinate with `docs/ideas/limits.md`.
6. **Docs** — `docs/reference/datastore.md` (durability section, binary support,
   TTL survival), extension rename, release notes.

## Part 7 — Test matrix

**Codec**
- round-trip every type, including nesting and empty array/object
- unicode object keys; key sort determinism (same value → identical bytes)
- numeric edges: NaN, ±Inf, −0.0, 2^53 boundary either side, very large/small floats
- `INT` vs `NUMBER` selection at the boundary
- function elision: top-level → nil, in array → nil, object value → key omitted
- `code` round-trip and re-parse; parse failure on load surfaces a keyed error
- `binary` round-trip including metadata — specifically the case that fails today
- `error` round-trip (message value + stack), `regex` round-trip and recompile
- `BINARY_REF` rejected cleanly on decode

**WAL**
- every opcode replays to the correct state
- `delete()` vs `set(key, nil)` produce distinguishable recovery
- `shift_wait()` / `pop_wait()` durability — consumed items stay consumed
- `rename()`, `clear()` survive restart
- TTL survives restart **and** survives a snapshot+truncate cycle
- expired-before-crash key does not resurrect past the first sweep
- torn tail: truncate mid-record, confirm prior records replay and the tail is dropped
- CRC corruption detected rather than silently applied
- encrypted and plaintext paths through the same replay code
- opcode `0x00` and `0xFF` rejected as invalid
- unknown opcode in a data range is fatal; unknown opcode in the cluster-control
  range is skipped without disturbing key-value state
- every opcode round-trips at exactly the byte length its shape implies —
  `CLEAR` carries no key at all, `SHIFT` carries key only, `RENAME` carries two
  keys — so a shape/table drift shows up as a size assertion failure

**Migration**
- a v1.6-written gob WAL and persist file both load correctly, plaintext and encrypted
- first write after migration produces a valid v1 file

**Performance regression**
- 5,000 push + 5,000 shift stays O(1) per op and writes kilobytes, not gigabytes
- `GET` throughput unchanged (must remain ~420k/s across durability modes)
- `SET`/`UPDATE` throughput no worse than today

## Non-goals

- Blob store for large binaries — tag reserved, deferred.
- Any replication or multi-node behavior. This plan is the prerequisite (a
  deterministic, self-contained, seq-numbered op log), not the feature.

  One requirement to carry forward when that work starts: with record shape
  driven by opcode and compatibility resolved by file-version-plus-duso-version,
  a cluster needs an explicit **version signal between nodes** — when a producer
  node restarts on a newer duso that emits opcodes a consumer doesn't know, the
  consumer must learn that and restart rather than hit an unknown opcode
  mid-stream and fail fatally. The reserved cluster-control opcode range
  (`0x80-0xBF`) is where that handshake would live, and the skippable-unknown
  rule for that range is what lets an older consumer receive it at all.
- Non-blocking snapshots. `saveToDisk` currently blocks writes for its duration
  (879ms at 100k keys). The `seq` watermark makes an incremental fix possible
  later; it is not attempted here.
