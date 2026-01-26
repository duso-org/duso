package script

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// HTTPServerValue represents an HTTP server in Duso.
// It manages routes and spawns handler scripts for incoming requests.
type HTTPServerValue struct {
	Port            int
	Address         string        // bind address (default "0.0.0.0")
	TLSEnabled      bool
	CertFile        string
	KeyFile         string
	Timeout         time.Duration
	routes          map[string]*Route // key: "METHOD /path"
	sortedRouteKeys []string          // Routes sorted by path length (descending)
	routeMutex      sync.RWMutex
	server          *http.Server
	Interpreter     *Interpreter // Interpreter for getting current script path
	ParentEval      *Evaluator    // Parent evaluator to copy functions from
	FileReader      func(string) ([]byte, error)
	FileStatter     func(string) int64 // Returns mtime, 0 if error
	startedChan     chan error         // Channel to communicate startup errors
}

// Route represents a registered HTTP route
type Route struct {
	Method      string
	Path        string
	HandlerPath string
}

// isValidHTTPMethod checks if a method is a valid HTTP method
func isValidHTTPMethod(method string) bool {
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
		"TRACE":   true,
		"CONNECT": true,
		"*":       true,
	}
	return validMethods[method]
}

// RequestContext holds the request-response context for a handler script
type RequestContext struct {
	Request    *http.Request
	Writer     http.ResponseWriter
	closed     bool
	mutex      sync.Mutex
	bodyCache  []byte // Cache request body since it can only be read once
	bodyCached bool
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

// GetRequestContext retrieves a request context from goroutine-local storage
func GetRequestContext(gid uint64) (*RequestContext, bool) {
	contextMutex.RLock()
	defer contextMutex.RUnlock()
	ctx, ok := requestContexts[gid]
	return ctx, ok
}

// clearRequestContext removes a request context from goroutine-local storage
func clearRequestContext(gid uint64) {
	contextMutex.Lock()
	defer contextMutex.Unlock()
	delete(requestContexts, gid)
}

// Route registers a new route (thread-safe).
// method can be: string ("GET", "get", "", "*"), nil, or []string for multiple methods
func (s *HTTPServerValue) Route(methodArg any, path, handlerPath string) error {
	s.routeMutex.Lock()
	defer s.routeMutex.Unlock()

	// Initialize routes map if nil
	if s.routes == nil {
		s.routes = make(map[string]*Route)
	}

	// Parse and validate method argument
	var methods []string

	switch m := methodArg.(type) {
	case nil:
		methods = []string{"*"} // nil = all methods
	case string:
		if m == "" || m == "*" {
			methods = []string{"*"} // "" or "*" = all methods
		} else {
			m = strings.ToUpper(m)
			if !isValidHTTPMethod(m) {
				return fmt.Errorf("invalid HTTP method: %q (valid: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT)", m)
			}
			methods = []string{m}
		}
	case []string:
		methods = make([]string, len(m))
		for i, item := range m {
			item = strings.ToUpper(item)
			if !isValidHTTPMethod(item) {
				return fmt.Errorf("invalid HTTP method: %q (valid: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT)", item)
			}
			methods[i] = item
		}
	case []Value:
		// Duso array as []Value - convert to strings
		methods = make([]string, len(m))
		for i, val := range m {
			if !val.IsString() {
				return fmt.Errorf("method array must contain strings, got %v", val.Type)
			}
			item := strings.ToUpper(val.AsString())
			if !isValidHTTPMethod(item) {
				return fmt.Errorf("invalid HTTP method: %q (valid: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT)", item)
			}
			methods[i] = item
		}
	case []interface{}:
		// Duso array as []interface{} - convert each to string
		methods = make([]string, len(m))
		for i, elem := range m {
			str, ok := elem.(string)
			if !ok {
				return fmt.Errorf("method array must contain strings, got %T", elem)
			}
			item := strings.ToUpper(str)
			if !isValidHTTPMethod(item) {
				return fmt.Errorf("invalid HTTP method: %q (valid: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT)", item)
			}
			methods[i] = item
		}
	default:
		return fmt.Errorf("method must be string, nil, or []string, got %T", m)
	}

	// Register route for each method
	for _, method := range methods {
		key := method + " " + path
		s.routes[key] = &Route{
			Method:      method,
			Path:        path,
			HandlerPath: handlerPath,
		}
	}

	// Rebuild sorted route keys
	s.rebuildSortedRoutes()

	return nil
}

// rebuildSortedRoutes sorts routes by path length (longest first) for prefix matching
func (s *HTTPServerValue) rebuildSortedRoutes() {
	keys := make([]string, 0, len(s.routes))
	for k := range s.routes {
		keys = append(keys, k)
	}

	// Sort by path length descending (most specific first)
	sort.Slice(keys, func(i, j int) bool {
		// Extract path from key (format: "METHOD /path")
		pathI := strings.SplitN(keys[i], " ", 2)[1]
		pathJ := strings.SplitN(keys[j], " ", 2)[1]
		return len(pathI) > len(pathJ)
	})

	s.sortedRouteKeys = keys
}

// findMatchingRoute finds the best matching route using prefix matching.
// Returns the most specific matching route, or nil if no match found.
// Must be called with routeMutex held (RLock).
func (s *HTTPServerValue) findMatchingRoute(method, path string) *Route {
	// Try exact method first, then wildcard
	for _, routeKey := range s.sortedRouteKeys {
		parts := strings.SplitN(routeKey, " ", 2)
		if len(parts) != 2 {
			continue
		}
		routeMethod := parts[0]
		routePath := parts[1]

		// Check if path matches (prefix match)
		if strings.HasPrefix(path, routePath) {
			// Check if method matches (exact or wildcard)
			if routeMethod == method || routeMethod == "*" {
				return s.routes[routeKey]
			}
		}
	}

	// If no exact method match found, try wildcard again
	for _, routeKey := range s.sortedRouteKeys {
		parts := strings.SplitN(routeKey, " ", 2)
		if len(parts) != 2 {
			continue
		}
		routeMethod := parts[0]
		routePath := parts[1]

		if routeMethod == "*" && strings.HasPrefix(path, routePath) {
			return s.routes[routeKey]
		}
	}

	return nil
}

// Start launches the HTTP server and blocks until the process receives a termination signal.
// This allows the script to handle cleanup code after the server stops.
// Returns an error if the server fails to bind to the port.
func (s *HTTPServerValue) Start() error {
	mux := http.NewServeMux()

	// Register catch-all handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Find matching route using prefix matching (most specific first)
		s.routeMutex.RLock()
		route := s.findMatchingRoute(r.Method, r.URL.Path)
		s.routeMutex.RUnlock()

		if route == nil {
			http.NotFound(w, r)
			return
		}

		// Handle request
		s.handleRequest(w, r, route)
	})

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Address, s.Port),
		Handler:      mux,
		ReadTimeout:  s.Timeout,
		WriteTimeout: s.Timeout,
	}

	// Channel to receive startup errors from server goroutine
	serverErr := make(chan error, 1)

	// Launch server in background goroutine
	go func() {
		var err error
		if s.TLSEnabled {
			err = s.server.ListenAndServeTLS(s.CertFile, s.KeyFile)
		} else {
			err = s.server.ListenAndServe()
		}

		// Send error (will be non-nil only if server fails to start, not on graceful shutdown)
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for server to start successfully (with timeout to catch binding errors)
	select {
	case err := <-serverErr:
		// Server failed to bind (port already in use, etc.)
		return err
	case <-time.After(100 * time.Millisecond):
		// Server started successfully, proceed to wait for signals
		break
	}

	// Set up signal handling - wait for Ctrl+C or termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal arrives
	<-sigChan

	// Gracefully shutdown the server
	// Use a timeout context so shutdown doesn't hang forever
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	return nil
}

// handleRequest processes an incoming HTTP request
func (s *HTTPServerValue) handleRequest(w http.ResponseWriter, r *http.Request, route *Route) {
	// Create request context
	ctx := &RequestContext{
		Request: r,
		Writer:  w,
		closed:  false,
	}

	// Store in goroutine-local storage
	gid := GetGoroutineID()
	setRequestContext(gid, ctx)
	defer clearRequestContext(gid)

	// Create fresh evaluator (child of parent)
	childEval := NewEvaluator(&strings.Builder{})

	// Copy registered functions from parent evaluator
	if s.ParentEval != nil {
		for name, fn := range s.ParentEval.goFunctions {
			childEval.RegisterFunction(name, fn)
		}
	}

	// Parse handler script
	// Read file using provided reader
	if s.FileReader == nil {
		if !ctx.closed {
			http.Error(w, "Server not properly configured: no file reader", 500)
		}
		return
	}

	fileBytes, err := s.FileReader(route.HandlerPath)
	if err != nil {
		if !ctx.closed {
			http.Error(w, fmt.Sprintf("Failed to load handler: %v", err), 500)
		}
		return
	}
	source := string(fileBytes)

	// Tokenize and parse
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		if !ctx.closed {
			http.Error(w, fmt.Sprintf("Handler script parse error: %v", err), 500)
		}
		return
	}

	// Execute handler script
	_, err = childEval.Eval(program)
	if err != nil {
		// If no response sent yet, send 500
		if !ctx.closed {
			http.Error(w, fmt.Sprintf("Handler script error: %v", err), 500)
		}
		return
	}

	// If handler didn't send response, send 204 No Content
	if !ctx.closed {
		w.WriteHeader(http.StatusNoContent)
	}
}

// GetRequest returns the request data as a map for the context() builtin
func (rc *RequestContext) GetRequest() map[string]any {
	// Parse headers
	headers := make(map[string]any)
	for k, vv := range rc.Request.Header {
		if len(vv) == 1 {
			headers[k] = vv[0]
		} else {
			arr := make([]any, len(vv))
			for i, v := range vv {
				arr[i] = v
			}
			headers[k] = arr
		}
	}

	// Parse query params
	query := make(map[string]any)
	for k, vv := range rc.Request.URL.Query() {
		if len(vv) == 1 {
			query[k] = vv[0]
		} else {
			arr := make([]any, len(vv))
			for i, v := range vv {
				arr[i] = v
			}
			query[k] = arr
		}
	}

	// Read body (cache it since it can only be read once)
	body := ""
	if !rc.bodyCached {
		bodyBytes, err := io.ReadAll(rc.Request.Body)
		if err == nil {
			rc.bodyCache = bodyBytes
			rc.bodyCached = true
			body = string(bodyBytes)
		}
	} else {
		body = string(rc.bodyCache)
	}

	return map[string]any{
		"method":  rc.Request.Method,
		"path":    rc.Request.URL.Path,
		"headers": headers,
		"query":   query,
		"body":    body,
	}
}

// SendResponse sends an HTTP response and marks the context as closed
func (rc *RequestContext) SendResponse(data map[string]any) error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	if rc.closed {
		return fmt.Errorf("context already closed: response already sent")
	}

	// Extract status (default 200)
	status := 200
	if s, ok := data["status"]; ok {
		if statusNum, ok := s.(float64); ok {
			status = int(statusNum)
		}
	}

	// Extract headers
	if headers, ok := data["headers"]; ok {
		if headerMap, ok := headers.(map[string]any); ok {
			for k, v := range headerMap {
				rc.Writer.Header().Set(k, fmt.Sprintf("%v", v))
			}
		}
	}

	// Extract body
	body := ""
	if b, ok := data["body"]; ok {
		body = fmt.Sprintf("%v", b)
	}

	// Send response
	rc.Writer.WriteHeader(status)
	if body != "" {
		_, _ = rc.Writer.Write([]byte(body))
	}

	rc.closed = true
	return nil
}

