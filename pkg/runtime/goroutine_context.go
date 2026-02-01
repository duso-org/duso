package runtime

import (
	"encoding/json"
	"fmt"
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
	PathParams map[string]any           // Extracted path parameters from route pattern (e.g., {id: "123"})
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

// GetResponse returns an object with response helper methods for use in handler scripts
func (rc *RequestContext) GetResponse() map[string]any {
	ctx := rc // Capture context for closures

	// Create response helper object with methods
	return map[string]any{
		// json(data [, status]) - Send JSON response
		"json": script.NewGoFunction(func(args map[string]any) (any, error) {
			data, ok := args["0"]
			if !ok {
				return nil, fmt.Errorf("json() requires data argument")
			}

			status := 200.0
			if s, ok := args["1"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			} else if s, ok := args["status"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			}

			// Convert data to JSON
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal JSON: %w", err)
			}

			return ctx.SendResponse(map[string]any{
				"status": status,
				"body":   string(jsonBytes),
				"headers": map[string]any{
					"Content-Type": "application/json",
				},
			}), nil
		}),

		// text(data [, status]) - Send plain text response
		"text": script.NewGoFunction(func(args map[string]any) (any, error) {
			data, ok := args["0"]
			if !ok {
				return nil, fmt.Errorf("text() requires data argument")
			}

			status := 200.0
			if s, ok := args["1"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			} else if s, ok := args["status"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			}

			return ctx.SendResponse(map[string]any{
				"status": status,
				"body":   fmt.Sprintf("%v", data),
				"headers": map[string]any{
					"Content-Type": "text/plain; charset=utf-8",
				},
			}), nil
		}),

		// html(data [, status]) - Send HTML response
		"html": script.NewGoFunction(func(args map[string]any) (any, error) {
			data, ok := args["0"]
			if !ok {
				return nil, fmt.Errorf("html() requires data argument")
			}

			status := 200.0
			if s, ok := args["1"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			} else if s, ok := args["status"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			}

			return ctx.SendResponse(map[string]any{
				"status": status,
				"body":   fmt.Sprintf("%v", data),
				"headers": map[string]any{
					"Content-Type": "text/html; charset=utf-8",
				},
			}), nil
		}),

		// error(status [, message]) - Send error response
		"error": script.NewGoFunction(func(args map[string]any) (any, error) {
			status := 500.0
			if s, ok := args["0"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			} else if s, ok := args["status"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			}

			message := ""
			if m, ok := args["1"]; ok {
				message = fmt.Sprintf("%v", m)
			} else if m, ok := args["message"]; ok {
				message = fmt.Sprintf("%v", m)
			}

			body := fmt.Sprintf("%v", int(status))
			if message != "" {
				body = message
			}

			return ctx.SendResponse(map[string]any{
				"status": status,
				"body":   body,
				"headers": map[string]any{
					"Content-Type": "text/plain; charset=utf-8",
				},
			}), nil
		}),

		// redirect(url [, status]) - Send redirect response
		"redirect": script.NewGoFunction(func(args map[string]any) (any, error) {
			url, ok := args["0"]
			if !ok {
				return nil, fmt.Errorf("redirect() requires url argument")
			}

			status := 302.0
			if s, ok := args["1"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			} else if s, ok := args["status"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			}

			return ctx.SendResponse(map[string]any{
				"status": status,
				"headers": map[string]any{
					"Location": fmt.Sprintf("%v", url),
				},
			}), nil
		}),

		// file(path [, status]) - Send file response
		"file": script.NewGoFunction(func(args map[string]any) (any, error) {
			path, ok := args["0"]
			if !ok {
				return nil, fmt.Errorf("file() requires path argument")
			}

			status := 200.0
			if s, ok := args["1"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			} else if s, ok := args["status"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			}

			return ctx.SendResponse(map[string]any{
				"status":   status,
				"filename": fmt.Sprintf("%v", path),
			}), nil
		}),

		// response(data, status [, headers]) - Generic response
		"response": script.NewGoFunction(func(args map[string]any) (any, error) {
			data, ok := args["0"]
			if !ok {
				return nil, fmt.Errorf("response() requires data argument")
			}

			status := 200.0
			if s, ok := args["1"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			} else if s, ok := args["status"]; ok {
				if statusNum, ok := s.(float64); ok {
					status = statusNum
				}
			}

			headers := make(map[string]any)
			if h, ok := args["2"]; ok {
				if headerMap, ok := h.(map[string]any); ok {
					headers = headerMap
				}
			} else if h, ok := args["headers"]; ok {
				if headerMap, ok := h.(map[string]any); ok {
					headers = headerMap
				}
			}

			return ctx.SendResponse(map[string]any{
				"status":  status,
				"body":    fmt.Sprintf("%v", data),
				"headers": headers,
			}), nil
		}),
	}
}
