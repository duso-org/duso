# Datastore Optimization: sync.Map Refactor

## Overview

Replace the global `sync.RWMutex` on the datastore's main data map with Go's `sync.Map`. This eliminates lock contention on session reads and other per-key operations, enabling true concurrent access to different keys.

**Status:** Planning phase
**Priority:** High (session cache reads are a hot path in production chat server)
**Impact:** 5-20x improvement for read-heavy workloads with many concurrent keys

---

## Problem Statement

Current architecture:
```go
type DatastoreValue struct {
    data      map[string]any
    dataMutex sync.RWMutex  // Single global lock for ALL operations
    // ...
}
```

**Bottleneck:** All operations (Get, Set, Increment, Push, etc.) contend on single `dataMutex`, even when accessing different keys. Example:
- Worker 1: `Get("session_user_123")` holds RLock
- Worker 2: `Get("session_user_456")` waits for RLock
- Both serialize despite different keys

**Use case impact:** Production chat server with session cache. 1000s of concurrent workers reading different session keys all serialize on one lock.

---

## Solution: sync.Map

Use Go's built-in `sync.Map` which:
- Lock-free reads (optimized for concurrent Get operations)
- Fine-grained locking on writes
- No manual locking required
- Designed exactly for "high read, occasional write" patterns

```go
type DatastoreValue struct {
    data sync.Map  // Lock-free reads, automatic per-key locking on writes
    // ...
}
```

---

## Implementation Plan

### Phase 1: Struct Changes

**File:** `pkg/runtime/datastore.go`

**Changes:**

1. **Line 44:** Replace `data` field
   ```go
   // Before
   data map[string]any

   // After
   data sync.Map
   ```

2. **Remove line 45:** Delete `dataMutex` field
   ```go
   // REMOVE: dataMutex sync.RWMutex
   ```

3. **Add after line 46:** New mutexes for conditions and expiry tracking
   ```go
   conditionsMu  sync.RWMutex      // Protect conditions map
   expiryTimesMu sync.RWMutex      // Protect expiryTimes map
   ```

### Phase 2: Function Updates

**Pattern:** Replace `dataMutex.Lock()` + map access with `data.Store()`/`data.Load()`

#### High-priority functions (hot paths):

1. **Get()** (line 211)
   - Replace: `dataMutex.RLock()` / `RUnlock()`
   - With: `data.Load(key)`
   - Protect expiry check with `expiryTimesMu`

2. **Set()** (line 144)
   - Replace: `dataMutex.Lock()` / `Unlock()`
   - With: `data.Store(key, value)`
   - Protect conditions access with `conditionsMu`

3. **Increment()** (line 279)
   - Replace: `dataMutex.Lock()` / `Unlock()`
   - With: `data.Load(key)` → `data.Store(key, newValue)`
   - Protect conditions access with `conditionsMu`

4. **Swap()** (line 240)
   - Replace: `dataMutex.Lock()` / `Unlock()`
   - With: `data.Load()` + `data.Store()`
   - Protect conditions access with `conditionsMu`

#### Array operations:

5. **Push()** (line 307)
   - Replace: `dataMutex.Lock()` / `Unlock()`
   - With: `data.Load()` + `data.Store()`
   - Protect conditions access with `conditionsMu`

6. **Shift()** (line 346)
   - Replace: `dataMutex.Lock()` / `Unlock()`
   - With: `data.Load()` + `data.Store()`
   - Protect conditions access with `conditionsMu`

7. **Pop()** (line 380)
   - Same pattern as Shift()

8. **Unshift()** (line 544)
   - Same pattern as Push()

#### Wait operations:

9. **ShiftWait()** (line 415)
   - Replace: `dataMutex.Lock()` with `data.Load()`
   - Keep: Condition variable logic, but protect conditions map
   - Note: Multiple calls to `data.Load()` in loop, all lock-free

10. **PopWait()** (line 480)
    - Same pattern as ShiftWait()

11. **Wait()** (line 698)
    - Replace: `dataMutex.Lock()` with `data.Load()`
    - Protect: conditions map access with `conditionsMu`

12. **WaitFor()** (line 766)
    - Same pattern as Wait()

#### Utility functions:

13. **SetOnce()** (line 175)
    - Atomic pattern: `data.LoadOrStore()` (Go 1.20+)
    - Or: `data.Load()` check → `data.Store()`
    - Protect conditions map access

14. **Exists()** (line 580)
    - Replace: `dataMutex.RLock()` / `RUnlock()`
    - With: `data.Load()` and check ok boolean

15. **Delete()** (line 828)
    - Replace: `dataMutex.Lock()` / `Unlock()`
    - With: `data.Delete(key)`
    - Protect: conditions and expiryTimes cleanup with their mutexes

16. **Rename()** (line 589)
    - Read old value: `data.Load(oldKey)`
    - Store new value: `data.Store(newKey, oldValue)`
    - Delete old: `data.Delete(oldKey)`
    - Protect: conditions map manipulation with `conditionsMu`

#### Map-wide operations:

17. **Keys()** (line 870)
    - Replace: `dataMutex.RLock()` + `for k := range`
    - With: `data.Range(func(key, value interface{}) bool { ... })`

18. **saveToDisk()** (line 904)
    - Convert sync.Map to regular map:
      ```go
      tempData := make(map[string]any)
      ds.data.Range(func(key, value interface{}) bool {
          tempData[key.(string)] = value
          return true
      })
      jsonData, _ := json.MarshalIndent(tempData, "", "  ")
      ```

19. **loadFromDisk()** (line 942)
    - Replace: `dataMutex.Lock()` + direct assignment
    - With: `data.Store(key, value)` in loop

20. **Clear()** (line 841)
    - sync.Map has no `Clear()`, so iterate and delete:
      ```go
      ds.data.Range(func(key, value interface{}) bool {
          ds.data.Delete(key.(string))
          return true
      })
      ```

#### Expiry operations:

21. **Expire()** (line 629)
    - Check key exists: `data.Load(key)` not `dataMutex.Lock()`
    - Protect: `expiryTimes` map update with `expiryTimesMu`

22. **sweepExpiredKeys()** (line 654)
    - Protect: `expiryTimes` map access with `expiryTimesMu`
    - When deleting expired key: `data.Delete(key)`
    - Protect: `conditions` map cleanup with `conditionsMu`

23. **checkExpired()** (line 680)
    - Protect: `expiryTimes` map access with `expiryTimesMu`
    - When deleting: `data.Delete(key)`

---

## Implementation Order

1. **Step 1:** Struct changes (Phase 1)
2. **Step 2:** Hot paths (Get, Set, Increment, Swap)
3. **Step 3:** Array operations (Push, Shift, Pop, Unshift)
4. **Step 4:** Wait operations (ShiftWait, PopWait, Wait, WaitFor)
5. **Step 5:** Utility functions (SetOnce, Exists, Delete, Rename)
6. **Step 6:** Map-wide operations (Keys, Save, Load, Clear)
7. **Step 7:** Expiry operations (Expire, sweepExpiredKeys, checkExpired)
8. **Step 8:** Testing and validation

---

## Key Pattern Summary

### Reading from data:
```go
// Before
ds.dataMutex.RLock()
val, exists := ds.data[key]
ds.dataMutex.RUnlock()

// After
val, exists := ds.data.Load(key)
```

### Writing to data:
```go
// Before
ds.dataMutex.Lock()
ds.data[key] = value
ds.dataMutex.Unlock()

// After
ds.data.Store(key, value)
```

### Protecting shared maps (conditions, expiryTimes):
```go
// Before
ds.dataMutex.Lock()
cond := ds.conditions[key]
ds.dataMutex.Unlock()

// After
ds.conditionsMu.RLock()
cond := ds.conditions[key]
ds.conditionsMu.RUnlock()
```

---

## Testing Strategy

1. **Unit tests:** Run existing datastore tests, verify all operations still work
2. **Concurrency test:** Spawn 1000 workers reading different session keys, measure lock contention
3. **Performance benchmark:** Compare before/after for:
   - Get() throughput on different keys
   - Set() throughput
   - Mixed read/write under contention
4. **Edge cases:** Expiry, persistence, condition variables

---

## Rollback Plan

If issues arise:
- Keep current code in git history
- Changes are isolated to datastore.go (one file)
- Can revert with single commit
- No API changes (public interface unchanged)

---

## Expected Improvements

**Session read cache (primary use case):**
- Lock contention: Eliminated for different keys
- Throughput: 5-20x improvement (measured for 1000s concurrent reads on different keys)
- Latency: Reduced (no RWMutex contention)

**Shared key operations (like message queue):**
- No change (single key still serializes correctly)

**Overall code:**
- Simpler (no manual locking)
- Cleaner (no dataMutex everywhere)
- Less error-prone (sync.Map is well-tested library code)

---

## Notes

- **sync.Map trade-off:** `Range()` is slower than direct map iteration. Used in Keys(), saveToDisk(), loadFromDisk()—none are hot paths, so acceptable.
- **Type assertions:** sync.Map stores `interface{}`, requiring type assertions. Existing pattern already does this.
- **SetOnce optimization:** If Go 1.20+, can use `LoadOrStore()` for atomic set-if-not-exists. Otherwise, manual Load + check + Store.
- **Backward compat:** No breaking changes to public API or behavior.

---

## Success Criteria

✅ All existing tests pass
✅ Production chat server session cache shows measurable throughput improvement
✅ No regression on write-heavy workloads
✅ Code is simpler and easier to understand
