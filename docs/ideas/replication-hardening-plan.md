# Plan: Datastore Replication Hardening

## Context

Datastore replication landed in `6b34279` / `6e7f02c` / `217649e`. A code review of
`pkg/runtime/datastore_replication.go`, `pkg/runtime/datastore_forward.go`, the Go tests in
`pkg/runtime/datastore_replication_test.go`, and `test/test_replication_config.du` found the
design sound and the small-value paths working end to end.

**Goal: ship the featureset as experimental and iterate over the following months.** That
sets the bar for this document. Experimental buys tolerance for missing capability, rough
edges, and limits users have to steer around. It does not buy silent wedges, invisible data
staleness, or a reference doc that promises more than the code delivers. Everything below is
sorted against that line.

---

# Part 1 — Gates the experimental release

Small, cheap, and each one is something a user cannot diagnose or work around on their own.
Estimated at well under a day of code; the tests are the longer pole.

## 1.1 Label the feature (no code)

`docs/reference/datastore_replication.md` currently opens with "for failover and continuous
backup" and carries no hedge of any kind. That is a durability promise, and right now it is
the largest gap between what the docs claim and what the code has been proven to do.

- Add an experimental banner at the top: interface may change, not yet validated against
  real remote servers, not yet recommended as the only copy of data.
- Add a "Known limitations" section carrying the items in Part 2 that ship unfixed.
- The latency table in that doc is reasoned estimates, not measurements. Say so inline, or
  drop the numbers until there are measured ones.

This is the highest value-per-minute item in the document.

## 1.2 Frames over ~4MB wedge a follower permanently — DONE

**Fixed.** `replMaxBatchBytes` bounds one coalesced message at 1MB, `since()` takes a byte
budget and reports the seq it actually reached, and the payload cap is now derived from the
store's `max_value_size` instead of being pinned at a snapshot chunk. The leader keeps the
small `replHandshakeLimit` until `lookup()` succeeds, then raises it, so an anonymous peer
still cannot make it allocate. The leader also advertises its limit in the welcome and the
follower takes the larger of the two, so a follower configured for smaller values than its
leader can still receive the leader's frames.

Covered by `TestReplRingSinceBoundsBatchSize` (unit), plus
`TestReplicationDeliversValueOverPayloadCap` and
`TestReplicationResumesFromLargeBacklogOfSmallWrites` over a real socket. The second one is
configured with a small `max_value_size` against the default 64MB ring on purpose: it fails
if the batch bound is removed even with the payload limit raised, which is what makes it a
regression test for the batch rather than for the cap.

Original analysis follows.

---

Both ends set `ws.MaxPayloadBytes = replSnapshotChunk + 1024` (4MB + 1KB). The snapshot
transfer respects that by chunking. The frame path does not, and `maxValueSize` defaults to
64MB.

Reproduced with a leader and follower over a real socket: one small write replicates, then
`leader.Set("big", <6MB string>)` produces

```
follower resumed at seq 1 (leader at seq 2, epoch 1, read-write)
stream ended: websocket: frame payload size exceeds limit (retrying in 1s)
```

repeating forever. The follower never advances past the write before the big one.

Two independent causes:

- A single frame larger than the cap can never be delivered.
- `serve()` coalesces every pending frame into one message (`datastore_replication.go:792`)
  with no size bound. The ring defaults to 64MB, so a follower catching up can be handed a
  64MB message built entirely out of small values.

**The second cause is why this gates the release.** A single 4MB value is unrealistic for
ordinary text and numeric web/API workloads and could be documented away. A 4MB *batch* is
not: a follower that blips for a second during a write burst, or whose apply loop falls
behind, gets everything it missed in one message. It triggers with small values, under
conditions the user cannot predict or steer around, and it never recovers.

It livelocks rather than failing slowly: `replFollowLoop` resets `backoff = replReconnectMin`
whenever `streamed` is true (line 894), which is always true here. So it reconnects every
500ms indefinitely, and `serve()` sees the cursor still inside the ring, resumes, and re-sends
the same oversized batch.

### Fix

Bound the batch. `since()` currently returns `[][]byte` with no seqs, so `serve()` cannot
advance `cursor` partially. Give it a byte budget and have it report the seq it actually
reached:

```go
if cursor < r.lastSeq {
    out := make([][]byte, 0, len(r.entries))
    last, total := cursor, 0
    for _, e := range r.entries {
        if e.seq <= cursor {
            continue
        }
        // len(out) > 0 keeps a single oversized frame deliverable on its own
        // rather than dropped: it goes out alone and the cap below covers it.
        if len(out) > 0 && total+len(e.frame) > maxBytes {
            break
        }
        out = append(out, e.frame)
        total += len(e.frame)
        last = e.seq
    }
    return out, last, false
}
```

`serve()` then loops naturally — the next `since()` returns immediately because
`cursor < lastSeq` still holds. One extra argument, three test call sites.

Then handle the single oversized frame by raising `MaxPayloadBytes` above `maxValueSize` on
both ends. **Watch the ordering:** `ws.MaxPayloadBytes` is set at the top of `serve()` before
the hello is read, so a 64MB limit would let an unauthenticated peer make the leader allocate
64MB. Keep it small for the handshake and raise it after `lookup()` succeeds.

**Cheaper stopgap, if the release cannot wait for the above:** on `ErrFrameTooLarge` the
follower resets its cursor to 0, which drives `needSnapshot` and self-heals via a full
resync. Roughly five lines, converts a permanent wedge into an expensive recovery. Not the
right long-term answer — every oversized frame costs every follower a full resync — but it is
honest and it recovers.

**Test:** a value over the payload cap, leader to follower, asserting the follower's cursor
catches up. That one test is what would have caught this.

## 1.3 The same cap silently kills the forward path

`replServeRequests` reads forwarded writes with the same 4MB cap. An oversized request fails
`Receive` and the goroutine simply returns. `serve()` never notices — its own writes still
succeed — so the connection stays up with nobody reading requests. Every later forwarded write
from that follower waits out `replForwardTimeout` (30s) and comes back "may or may not have
been applied".

The oversized-request half is fixed by the cap change in 1.2. The silent-death half is two
lines:

```go
go func() {
    defer core.RecoverPanic("datastore_replication_requests")
    defer ws.Close() // reader gone; the connection is no longer usable
    l.replServeRequests(ws, ds, state, done)
}()
```

That turns a permanent half-dead socket into a disconnect the follower already knows how to
handle. It also resolves the malformed-request case: a malformed request is answered with
`replReply{Err: ...}` and no ID (`datastore_forward.go:421`), and since IDs start at 1,
`deliver` finds no waiter and the caller eats the full 30s timeout. You cannot echo an ID you
failed to decode, so teardown is the honest answer.

## 1.4 `repl.epoch` data race

Plain `uint32`. Written by the follow loop (`replSession:970`, `replReceiveSnapshot:1093`),
read by `replStatus()` from script goroutines and by `replEpoch()` inside `encodeSnapshot` /
`appendWALHeader`, which run on the persist ticker goroutine. `-race` is clean on the current
tests only because nothing reads it concurrently there; a follower with `persist_interval` set
will trip it.

Gates the release because the epoch is what fences a superseded leader — a torn read weakens
the split-brain guard, and that is a data-integrity story rather than a rough edge.

**Fix:** `atomic.Uint32`.

## 1.5 Make promotion's `persist` dependency explicit

`replConfigure` reads provenance only under `if ds.persistPath != ""`. The WAL header carries
an epoch (`wal.go:286`) but nothing reads it back — `readWALHeader` checks magic, version and
flags only. A WAL-only follower promoted to leader therefore starts at epoch 1 instead of
`snapEpoch + 1`, and the split-brain guard in `serve()` can never fire for it.

For the experimental release the cheap correct move is to **reject the config**: require
`persist` whenever a `replicate_*` option is set. That is a few lines in `replConfigure` and
turns a silent correctness hole into a startup error. Reading the epoch out of the WAL header
is the better long-term answer and can iterate.

---

# Part 2 — Iterates after release

Document these in the "Known limitations" section from 1.1 and fix them as the subsystem
matures.

- **Blocking-op timeout race.** `ShiftWait` / `PopWait` forwarding gives the follower and the
  leader the same timeout, so they race. If the follower's timer wins, the leader may already
  have popped the item and the reply is discarded: item lost, caller told "may or may not".
  Giving the leader a slightly shorter timeout makes the window one-sided. Worth a doc note
  now either way — at-most-once forwarding of blocking ops is a real semantic.
- **Ambiguous outcomes on mid-flight disconnect.** Already reported honestly by the error
  text; just needs to be in the docs so it is not a surprise.
- **Swallowed decode failures.** `Push` / `Unshift` / `Increment` do `n, _ := toFloat64(v)` on
  the forwarded result. An unexpected reply shape returns 0 with a nil error, and a silent 0
  from `push()` reads as an empty array.
- **Reading the epoch from the WAL header,** removing the `persist` requirement from 1.5.
- **Measured latency numbers** against real remote servers, replacing the estimates.

---

# Part 3 — Test gaps

The existing Go tests are well targeted — the ring, gap/duplicate/corruption handling,
snapshot provenance, config rejection and the reply-ordering rule are covered where they can
actually be interrogated, and `TestReplicationOverSocket` picks the ops that catch a
double-apply (INCR, PUSH). What is missing is size, and the refusal paths.

Rough priority order:

1. A value over the payload cap, end to end. Catches 1.2, and belongs with that fix.
2. The epoch refusal in `serve()` over a socket. The newest code in the branch, currently
   untested.
3. Read-only secret end to end: leader grants read-only, the follower's write comes back
   refused. `TestReplGrantDefaults` only checks the local default before any connection
   exists.
4. The `encrypt_key` mismatch refusal.
5. The `snapshotMagic` prefix check from `217649e`.
6. A follower falling off the ring and recovering. `TestReplRingReportsFallenOffFollower`
   covers the ring's answer; nothing covers the leader sending the "reconnect for a snapshot"
   refusal and the follower actually resyncing.

`test/test_replication_config.du` covers the configuration surface well, and the header
comment explaining why a leader and a follower cannot be paired in one process is the right
call. One gap: every `replicate_from` case in it is a rejection, so no follower is ever
successfully constructed from a script, and `replication_status()`'s follower fields
(`connected`, `can_write`, `cursor`, `last_error`) have no script-level coverage. A follower
pointed at a dead port would exercise all of them plus the "not connected — the write was not
applied" message, without needing a second process.
