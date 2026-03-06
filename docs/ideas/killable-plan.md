# Making Blocking Builtins Killable via kill(pid)

## Overview

Currently, `kill(pid)` doesn't work reliably on spawned processes that are blocked in I/O, synchronization, or sleep operations. This document outlines all blocking calls in Duso builtins and proposes a strategy to make them killable.

## Current Status

`kill()` was recently implemented using Go's context cancellation. However, most blocking operations don't check the cancellation context, so they continue blocking even after `kill()` is called.

## Blocking Calls Inventory

### Runtime Builtins

| Builtin | Blocking Call | File | Line | Has Timeout | Context-Aware |
|---------|---------------|------|------|-------------|---|
| `sleep(seconds)` | `time.Sleep()` | builtin_system.go | 42 | ❌ No | ❌ No |
| `input([prompt])` | `bufio.Reader.ReadString()` | builtin_console.go | 36 | ❌ No | ❌ No |
| `fetch(url, opts)` | `client.Do()` + `ReadAll()` | builtin_fetch.go | 113, 120 | ✅ Yes* | ❌ No |
| `parallel(fns)` | `sync.WaitGroup.Wait()` | builtin_parallel.go | 82, 121 | ❌ No | ❌ No |
| `http_server().Start()` | `server.ListenAndServe()` | http_server.go | 712-714 | ❌ No | ❌ No |
| `http_server().Start()` | `<-sigChan` (signal channel) | http_server.go | 738 | ❌ No | ❌ No |

### CLI Builtins

| Builtin | Blocking Call | File | Line | Has Timeout | Context-Aware |
|---------|---------------|------|------|-------------|---|
| `input([prompt])` | `bufio.Reader.ReadString()` | cli/builtin_console.go | 122 | ❌ No | ❌ No |
| `busy(msg)` | `time.Sleep()` | cli/builtin_busy.go | 125 | ❌ No | ❌ No |
| `watch(path, timeout)` | `time.Sleep()` in loop | cli/builtin_watch.go | 146 | ✅ Yes | ❌ No |

### Datastore Methods

| Method | Blocking Call | File | Line | Has Timeout | Context-Aware |
|--------|---------------|------|------|-------------|---|
| `.shift_wait(key, timeout)` | `sync.Cond.Wait()` | datastore.go | 461, 471 | ✅ Yes | ❌ No |
| `.pop_wait(key, timeout)` | `sync.Cond.Wait()` | datastore.go | ~530 | ✅ Yes | ❌ No |
| `.wait(key [, val])` | `sync.Cond.Wait()` | datastore.go | 746, 756 | ✅ Yes | ❌ No |
| `.wait_for(key, pred)` | `sync.Cond.Wait()` | datastore.go | ~810 | ✅ Yes | ❌ No |

**Total: 13 blocking operations across 10 API functions**

*Has its own timeout mechanism, but doesn't use process context

## Implementation Strategy

### Priority Tiers

**Tier 1 (Easiest, High Impact):**
- `sleep()` - Replace `time.Sleep()` with `select` + context
- `watch()` - Already has polling loop, add context check
- `busy()` - Same as sleep()
- `input()` - Add context polling with non-blocking mode

**Tier 2 (Medium Difficulty):**
- `fetch()` - Add context to request via `req.WithContext()`
- `http_server().Start()` - Add goroutine to listen to context cancellation

**Tier 3 (Hardest):**
- `parallel()` - Use context-aware goroutine coordination
- Datastore `.wait*()` methods - Refactor from `sync.Cond` to channels (architectural change)

### Implementation Patterns

#### Pattern 1: Replace time.Sleep with Context-Aware Wait

```go
// Before
time.Sleep(duration)

// After
select {
case <-time.After(duration):
    // Sleep completed normally
case <-procCtx.Done():
    // Process was killed
    return
}
```

**Files to update:**
- `builtin_system.go` - `sleep()`
- `builtin_busy.go` - busy spinner loop
- `builtin_watch.go` - polling loop (already uses time.Sleep)

#### Pattern 2: Add Context to HTTP Client

```go
// Before
req, _ := http.NewRequest(method, url, body)
client.Do(req)

// After
req, _ := http.NewRequestWithContext(procCtx, method, url, body)
client.Do(req)
```

**Files to update:**
- `builtin_fetch.go` - fetch() request

#### Pattern 3: Interrupt Blocking Listen with Goroutine

```go
// Before
server.ListenAndServe()  // Blocks forever

// After
go func() {
    <-procCtx.Done()
    server.Shutdown(context.Background())
}()
server.ListenAndServe()
```

**Files to update:**
- `http_server.go` - Start() method

#### Pattern 4: Refactor sync.Cond to Channels (Most Complex)

Datastore wait operations use `sync.Cond` which doesn't support context cancellation. Options:

**Option A:** Use select with context (requires architecture change)
```go
select {
case val := <-waitChannel:
    return val
case <-procCtx.Done():
    return nil  // Killed
}
```

**Option B:** Keep timeout but add periodic context checks (simpler, less efficient)

**Files to update:**
- `datastore.go` - ShiftWait, PopWait, Wait, WaitFor methods

## Performance Impact

### Context Checking Overhead: **MINIMAL**

1. **Creating context per spawn:** ~microseconds (one-time cost)
2. **Checking context.Done():** O(1) non-blocking channel read, ~nanoseconds
3. **Select statement penalty:** Negligible (~microseconds in select, but only at blocking points)

Go's context implementation is highly optimized and used extensively in production code.

### Adding Periodic Context Checks

For operations like `watch()` that already sleep periodically:
- No additional overhead—just add context check alongside existing sleep
- Already paying sleep latency cost

### Refactoring sync.Cond to Channels

Potential performance impact if done poorly, but:
- Datastore wait operations are I/O-bound (waiting for values to change)
- CPU overhead of channels vs condition variables is negligible compared to blocking cost
- Can optimize by keeping fast path for common case (no timeout)

### Estimated Impact: < 1% in most scenarios

The primary cost is design complexity, not performance. Go's concurrency primitives are heavily optimized.

## Risks & Considerations

### Risk 1: Goroutine Leaks
**Concern:** Spawning cleanup goroutines (e.g., timeout goroutines in datastore wait)
**Mitigation:** Ensure all goroutines are properly cancelled and cleaned up

### Risk 2: Resource Cleanup
**Concern:** Killing mid-operation (e.g., during file read) could leak file handles
**Mitigation:** Rely on Go's defer cleanup, ensure all file operations use proper cleanup

### Risk 3: Partial State Changes
**Concern:** Killing during datastore operation (e.g., between checking and modifying)
**Mitigation:** Datastore operations already use locks for atomicity—context check won't break this

### Risk 4: API Compatibility
**Concern:** Adding context parameters changes function signatures
**Mitigation:** Context is thread-local (via goroutine ID), no signature changes needed

## Phased Rollout Plan

### Phase 1: Tier 1 Operations (Low Risk, High Value)
- [ ] `sleep()` - context-aware wait
- [ ] `watch()` - context check in loop
- [ ] `busy()` - context-aware sleep
- Estimated effort: 2-3 hours
- Testing: Simple—verify kill stops operation

### Phase 2: Tier 2 Operations (Medium Risk/Effort)
- [ ] `fetch()` - context-aware request
- [ ] `http_server().Start()` - graceful shutdown on kill
- Estimated effort: 2-3 hours
- Testing: Test network interruption scenarios

### Phase 3: Tier 3 Operations (High Risk/Effort)
- [ ] `parallel()` - context-aware goroutine wait
- [ ] Datastore wait methods - channel-based refactor
- Estimated effort: 4-6 hours
- Testing: Complex—verify atomic operation semantics

## Testing Strategy

### Unit Tests
- Verify kill() interrupts operation within reasonable time (~100ms)
- Verify killed process actually exits (no resource leaks)
- Verify cleanup handlers run (defers, channel closes, etc.)

### Integration Tests
- Spawned process calling `sleep()` → kill → exits
- Spawned process in HTTP server → kill → server stops
- Spawned process doing datastore wait → kill → exits
- Multiple spawned processes → kill some → others continue

### Edge Cases
- Kill during exception handling
- Kill during nested operations (sleep in loop in spawned process)
- Kill process that's already exiting naturally
- Kill already-dead process (should error gracefully)

## Timeline Estimate

- **Phase 1:** 2-3 hours (start here)
- **Phase 2:** 2-3 hours (after phase 1 complete)
- **Phase 3:** 4-6 hours (after phases 1 & 2 complete)
- **Total:** 8-12 hours of development

## Success Criteria

✅ All blocking operations check process cancellation context
✅ `kill(pid)` terminates process within 100ms (or timeout) of call
✅ No resource leaks (files, goroutines, channels)
✅ All existing tests still pass
✅ Documentation updated with kill() guarantees

## References

- Go context package: https://pkg.go.dev/context
- Go HTTP context: https://pkg.go.dev/net/http#Request.WithContext
- Graceful shutdown patterns: https://pkg.go.dev/net/http#Server.Shutdown
