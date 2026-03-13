# Duso Code Review - Issues Found

## HIGH PRIORITY ISSUES

### 1. **Context Cleanup Duplication in HTTP Handler**
**Severity: HIGH**
**File:** `pkg/runtime/http_server.go` (lines 952-984)

**Problem:**
The HTTP handler code attempts to manage both runtime and script context cleanups in a confusing pattern:

```go
// Line 953
SetRequestContextWithData(gid, ctx, contextData)  // runtime.SetRequestContextWithData
defer clearRequestContext(gid)  // runtime.clearRequestContext

// ... [code] ...

// Line 983
script.SetRequestContextWithData(gid, scriptCtx, contextData)  // script.SetRequestContextWithData
defer script.ClearRequestContext(gid)  // script.ClearRequestContext
```

Issues:
1. Two separate contexts are stored with the same goroutine ID
2. The second `SetRequestContextWithData` call overwrites the first context (same key in different maps)
3. Cleanup order (due to defer stack) is: script.ClearRequestContext (executed first), then runtime.clearRequestContext (executed second)
4. This is confusing and error-prone

**Recommendation:**
Once the RequestContext unification is done, simplify this to a single setup/cleanup pattern.

---

### 2. **Goroutine Leak Risk in Datastore Auto-save and Expiry**
**Severity: MEDIUM-HIGH**
**File:** `pkg/runtime/datastore.go` (lines 103-129)

**Problem:**
```go
// Start auto-save ticker if configured
if store.persistInterval > 0 {
    store.ticker = time.NewTicker(store.persistInterval)
    go func() {
        for {
            select {
            case <-store.ticker.C:
                _ = store.saveToDisk()
            case <-store.stopTicker:
                return
            }
        }
    }()
}

// Start expiry sweep ticker (1-second sweep)
expiryTicker := time.NewTicker(1 * time.Second)  // NOT STORED!
go func() {
    for {
        select {
        case <-expiryTicker.C:
            store.sweepExpiredKeys()
        case <-store.expiryStopTicker:
            expiryTicker.Stop()
            return
        }
    }()
}()
```

Issues:
1. **expiryTicker is not stored** in the DatastoreValue struct - it will be garbage collected if not referenced
2. Even if the goroutine continues running, it won't be cleanable because there's no way to access the ticker to stop it before cleanup
3. No cleanup mechanism when datastores are deleted from the registry
4. Multiple datastores could accumulate tickers that are never stopped

**Recommendation:**
```go
type DatastoreValue struct {
    // ... existing fields ...
    ticker        *time.Ticker
    expiryTicker  *time.Ticker  // ADD THIS
    stopTicker    chan bool
}

// And add a cleanup method
func (ds *DatastoreValue) Close() error {
    if ds.ticker != nil {
        ds.ticker.Stop()
    }
    if ds.expiryTicker != nil {
        ds.expiryTicker.Stop()
    }
    // Close channels
    return nil
}
```

---

### 3. **Unvalidated File Paths in HTTP Server Static File Serving**
**Severity: MEDIUM**
**File:** `pkg/runtime/http_server.go` (lines 1140-1150)

**Problem:**
The code serves files from three different filesystem layers (`/EMBED/`, `/STORE/`, OS-level) without path validation:

```go
// Try each attempt in order
for _, attempt := range attempts {
    fileBytes, err = s.FileReader(attempt)
    if err == nil {
        break
    }
}
```

This allows potential directory traversal attacks (e.g., `../../../etc/passwd`).

**Recommendation: Chroot Jail Pattern**
Implement single-point path validation before checking any filesystem:

```go
// sanitizePath prevents directory traversal across all filesystem types
func sanitizePath(requestPath string) (string, error) {
    // Normalize the path
    cleaned := filepath.Clean(requestPath)

    // Reject if it tries to escape (contains .. after cleaning or is absolute)
    if cleaned != filepath.Clean(requestPath) ||
       filepath.IsAbs(cleaned) ||
       strings.Contains(cleaned, "..") {
        return "", fmt.Errorf("invalid path: %s", requestPath)
    }

    // Ensure path doesn't start with /
    if strings.HasPrefix(cleaned, "/") {
        cleaned = cleaned[1:]
    }

    return cleaned, nil
}

// Then in request handling:
cleanPath, err := sanitizePath(requestPath)
if err != nil {
    http.Error(w, "Invalid path", 400)
    return
}

// Now safely try each filesystem
attempts := []string{
    cleanPath,
    "/STORE/" + cleanPath,
    "/EMBED/" + cleanPath,
}
```

This "chroot jail" approach validates once upfront before attempting any filesystem access, preventing traversal across all three filesystem layers.

---

## MEDIUM PRIORITY ISSUES

### 4. **Incomplete Error Handling in JSON/Encoding Operations**
**Severity: MEDIUM**
**File:** `pkg/runtime/http_server.go` (lines 231-232)

**Problem:**
```go
// Encode header and payload
headerJSON, _ := json.Marshal(header)
payloadJSON, _ := json.Marshal(tokenClaims)
```

Ignoring errors from `json.Marshal()` on JWT operations could silently produce incorrect tokens.

**Recommendation:**
```go
headerJSON, err := json.Marshal(header)
if err != nil {
    return "", fmt.Errorf("failed to marshal JWT header: %w", err)
}
// ... repeat for payloadJSON
```

---

### 5. **Missing Nil Checks in ExecuteScript**
**Severity: MEDIUM**
**File:** `pkg/script/execution.go` (lines 30-48)

**Problem:**
```go
if requestContext != nil && requestContext.Evaluator != nil {
    childEval = requestContext.Evaluator
} else {
    childEval = NewEvaluator()
}

// Copy custom functions from interpreter
if interpreter != nil {
    parentEval := interpreter.GetEvaluator()

    // Copy custom registered functions
    for name, fn := range parentEval.GetGoFunctions() {  // parentEval could be nil
        childEval.RegisterFunction(name, fn)
    }
}
```

If `interpreter.GetEvaluator()` returns nil, the code will panic on the `.GetGoFunctions()` call.

**Recommendation:**
```go
if interpreter != nil {
    parentEval := interpreter.GetEvaluator()
    if parentEval != nil {
        for name, fn := range parentEval.GetGoFunctions() {
            childEval.RegisterFunction(name, fn)
        }
    }
}
```

---

## SUMMARY

| Issue | Type | Severity | Files |
|-------|------|----------|-------|
| Context cleanup duplication | Correctness | HIGH | runtime/http_server.go |
| Goroutine leaks in datastore | Resource | MEDIUM-HIGH | runtime/datastore.go |
| Unvalidated file paths | Security | MEDIUM | runtime/http_server.go |
| Incomplete error handling in JWT | Correctness | MEDIUM | runtime/http_server.go |
| Missing nil check in ExecuteScript | Correctness | MEDIUM | script/execution.go |

---

## Recommended Action Plan

1. **High Priority:** Fix HTTP handler context cleanup pattern to simplify and reduce confusion
2. **Medium Priority:** Add path sanitization for file serving (chroot jail pattern) and fix datastore ticker cleanup
3. **Medium Priority:** Improve error handling in JWT operations and ExecuteScript nil checks
