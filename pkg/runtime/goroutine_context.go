package runtime

import (
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/duso-org/duso/pkg/script"
)

// RequestContext holds context data for a handler script
// Used for HTTP requests, spawn() calls, run() calls - anything that needs context
type RequestContext struct {
	Request    *http.Request           // HTTP request (if HTTP handler), nil otherwise
	Writer     http.ResponseWriter     // HTTP response writer (if HTTP handler), nil otherwise
	Data       any                     // Generic context data (used by spawn/run)
	closed     bool
	mutex      sync.Mutex
	bodyCache  []byte                   // Cache request body since it can only be read once
	bodyCached bool
	Frame      *script.InvocationFrame // Root invocation frame for this context
	ExitChan   chan any                 // Channel to receive exit value from script
}

// Global goroutine-local storage for request contexts
var (
	requestContexts = make(map[uint64]*RequestContext)
	contextMutex    sync.RWMutex
)

// GetGoroutineID extracts the current goroutine ID from the stack trace
func GetGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	stackTrace := string(buf[:n])

	// Parse "goroutine 123 [running]:"
	lines := strings.Split(stackTrace, "\n")
	if len(lines) > 0 {
		line := lines[0]
		if strings.HasPrefix(line, "goroutine ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				if id, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					return id
				}
			}
		}
	}
	return 0
}

// setRequestContext stores a request context in goroutine-local storage
func setRequestContext(gid uint64, ctx *RequestContext) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	requestContexts[gid] = ctx
}

// SetRequestContextWithData stores a request context with optional spawned context data
func SetRequestContextWithData(gid uint64, ctx *RequestContext, spawnedData any) {
	contextMutex.Lock()
	defer contextMutex.Unlock()

	// Store spawned data in the Data field (generic context)
	ctx.Data = spawnedData

	requestContexts[gid] = ctx
}

// GetRequestContext retrieves a request context from goroutine-local storage
func GetRequestContext(gid uint64) (*RequestContext, bool) {
	contextMutex.RLock()
	defer contextMutex.RUnlock()
	ctx, ok := requestContexts[gid]
	return ctx, ok
}

// ClearRequestContext removes a request context from goroutine-local storage
func ClearRequestContext(gid uint64) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	delete(requestContexts, gid)
}

// clearRequestContext removes a request context from goroutine-local storage (lowercase version)
func clearRequestContext(gid uint64) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	delete(requestContexts, gid)
}
