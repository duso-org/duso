package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// HTTPServerValue represents an HTTP server in Duso.
// It manages routes and spawns handler scripts for incoming requests.
type HTTPServerValue struct {
	Port                   int
	Address                string        // bind address (default "0.0.0.0")
	TLSEnabled             bool
	CertFile               string
	KeyFile                string
	Timeout                time.Duration // Socket-level read/write timeout
	RequestHandlerTimeout  time.Duration // Handler script execution timeout
	routes                 map[string]*Route // key: "METHOD /path"
	sortedRouteKeys        []string          // Routes sorted by path length (descending)
	routeMutex             sync.RWMutex
	server                 *http.Server
	Interpreter            *script.Interpreter // Interpreter for getting current script path
	ParentEval             *script.Evaluator    // Parent evaluator to copy functions from
	FileReader             func(string) ([]byte, error)
	FileStatter            func(string) int64 // Returns mtime, 0 if error
	startedChan            chan error         // Channel to communicate startup errors
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
	case []script.Value:
		// Duso array as []script.Value - convert to strings
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
	// Increment HTTP request counter
	IncrementHTTPProcs()

	// Create invocation frame for this HTTP route
	// Note: For phase 1, we create script.InvocationFrame since that's what DebugEvent expects
	frame := &script.InvocationFrame{
		Filename: route.HandlerPath,
		Line:     1,
		Col:      1,
		Reason:   "http_route",
		Details: map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
		},
		Parent: nil,
	}

	// Create request context with exit channel
	ctx := &RequestContext{
		Request:  r,
		Writer:   w,
		closed:   false,
		Frame:    frame,
		ExitChan: make(chan any, 1),
	}

	// Create fresh evaluator (child of parent)
	childEval := script.NewEvaluator(&strings.Builder{})

	// Copy registered functions and settings from parent evaluator
	if s.ParentEval != nil {
		for name, fn := range s.ParentEval.GetGoFunctions() {
			childEval.RegisterFunction(name, fn)
		}
		// Copy debug and stdin settings from parent
		childEval.DebugMode = s.ParentEval.DebugMode
		childEval.NoStdin = s.ParentEval.NoStdin
	}

	// Parse handler script
	// Read file using provided reader (with fallback to embedded files)
	if s.FileReader == nil {
		if !ctx.closed {
			http.Error(w, "Server not properly configured: no file reader", 500)
		}
		return
	}

	// Try reading handler with fallback: local -> /EMBED/{path} -> /EMBED/{scriptDir}/{path}
	var fileBytes []byte
	var err error

	// Try 1: Local file
	fileBytes, err = s.FileReader(route.HandlerPath)
	if err != nil {
		// Try 2: Embedded file
		fileBytes, err = s.FileReader("/EMBED/" + route.HandlerPath)
		if err != nil {
			// Try 3: Embedded file with script directory
			if s.Interpreter != nil && s.Interpreter.GetScriptDir() != "" && s.Interpreter.GetScriptDir() != "." {
				fileBytes, err = s.FileReader("/EMBED/" + s.Interpreter.GetScriptDir() + "/" + route.HandlerPath)
			}
			if err != nil {
				if !ctx.closed {
					http.Error(w, fmt.Sprintf("Failed to load handler: %v", err), 500)
				}
				return
			}
		}
	}
	source := string(fileBytes)

	// Tokenize and parse
	lexer := script.NewLexer(source)
	tokens := lexer.Tokenize()
	parser := script.NewParserWithFile(tokens, route.HandlerPath)
	program, err := parser.Parse()
	if err != nil {
		if !ctx.closed {
			http.Error(w, fmt.Sprintf("Handler script parse error: %v", err), 500)
		}
		return
	}

	// Create timeout context for handler execution
	handlerCtx, cancel := context.WithTimeout(context.Background(), s.RequestHandlerTimeout)
	defer cancel()

	// Execute handler script with timeout
	evalDone := make(chan error, 1)
	go func() {
		// Register request context in THIS goroutine
		gid := GetGoroutineID()
		setRequestContext(gid, ctx)
		defer clearRequestContext(gid)

		_, err := childEval.Eval(program)
		evalDone <- err
	}()

	// Wait for evaluation or timeout
	select {
	case err = <-evalDone:
		// Evaluation completed
	case <-handlerCtx.Done():
		// Timeout occurred
		if !ctx.closed {
			http.Error(w, "Handler script timeout exceeded", 504)
		}
		return
	}

	// Check for exit() call
	var exitValue any
	if exitErr, ok := err.(*script.ExitExecution); ok {
		// Script called exit() - capture first value if present
		if len(exitErr.Values) > 0 {
			exitValue = exitErr.Values[0]
		}
	} else if bpErr, ok := err.(*script.BreakpointError); ok {
		// Debug breakpoint - queue it for the main process
		debugEvent := &script.DebugEvent{
			Error:           bpErr,
			FilePath:        bpErr.FilePath,
			Position:        bpErr.Position,
			CallStack:       bpErr.CallStack,
			InvocationStack: frame,
			Env:             bpErr.Env,
		}
		if s.Interpreter != nil {
			s.Interpreter.QueueDebugEvent(debugEvent)
		}
		// Return 500 since debug paused execution
		if !ctx.closed {
			http.Error(w, "Debug breakpoint - check terminal for REPL", 500)
		}
		return
	} else if err != nil {
		// Regular error - queue as debug event if in debug mode
		if childEval.DebugMode {
			debugEvent := &script.DebugEvent{
				Error:           err,
				Message:         err.Error(),
				FilePath:        route.HandlerPath,
				InvocationStack: frame,
				Env:             childEval.GetEnv(),
			}
			// Extract position info if available
			if dusoErr, ok := err.(*script.DusoError); ok {
				debugEvent.FilePath = dusoErr.FilePath
				debugEvent.Position = dusoErr.Position
				debugEvent.CallStack = dusoErr.CallStack
			}
			if s.Interpreter != nil {
				s.Interpreter.QueueDebugEvent(debugEvent)
			}
			// Return 500 since debug paused execution
			if !ctx.closed {
				http.Error(w, fmt.Sprintf("Handler error - check terminal for REPL: %v", err), 500)
			}
			return
		}
		// Not in debug mode - return error normally
		if !ctx.closed {
			http.Error(w, fmt.Sprintf("Handler script error: %v", err), 500)
		}
		return
	}

	// Send response based on exit value or default
	if exitValue != nil {
		// Script called exit() with a value - treat as response
		if responseMap, ok := exitValue.(map[string]any); ok {
			s.sendHTTPResponse(w, responseMap)
		}
	} else {
		// No exit value - send 204 No Content if response not already sent
		if !ctx.closed {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// getContentType returns the MIME type for a file based on extension
func getContentType(filename string) string {
	mimeTypes := map[string]string{
		".html":  "text/html",
		".htm":   "text/html",
		".txt":   "text/plain",
		".json":  "application/json",
		".xml":   "application/xml",
		".css":   "text/css",
		".js":    "application/javascript",
		".mjs":   "application/javascript",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".svg":   "image/svg+xml",
		".webp":  "image/webp",
		".ico":   "image/x-icon",
		".pdf":   "application/pdf",
		".zip":   "application/zip",
		".gz":    "application/gzip",
		".mp3":   "audio/mpeg",
		".mp4":   "video/mp4",
		".webm":  "video/webm",
		".wav":   "audio/wav",
	}

	// Find extension
	dotIdx := strings.LastIndex(filename, ".")
	if dotIdx == -1 {
		return "application/octet-stream"
	}

	ext := strings.ToLower(filename[dotIdx:])
	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	return "application/octet-stream"
}

// sendHTTPResponse is a helper that sends HTTP response from a data map
func (s *HTTPServerValue) sendHTTPResponse(w http.ResponseWriter, data map[string]any) {
	// Extract status (default 200)
	status := 200
	if st, ok := data["status"]; ok {
		if statusNum, ok := st.(float64); ok {
			status = int(statusNum)
		}
	}

	// Extract headers
	if headers, ok := data["headers"]; ok {
		if headerMap, ok := headers.(map[string]any); ok {
			for k, v := range headerMap {
				w.Header().Set(k, fmt.Sprintf("%v", v))
			}
		}
	}

	// Check for filename (binary file serving)
	if filename, ok := data["filename"]; ok {
		if filenameStr, ok := filename.(string); ok {
			// Read the file
			fileBytes, err := s.FileReader(filenameStr)
			if err != nil {
				// Try with /EMBED/ prefix
				fileBytes, err = s.FileReader("/EMBED/" + filenameStr)
				if err != nil {
					w.WriteHeader(500)
					_, _ = w.Write([]byte(fmt.Sprintf("Failed to read file: %v", err)))
					return
				}
			}

			// Determine content type
			contentType := getContentType(filenameStr)

			// Allow explicit type override
			if t, ok := data["type"]; ok {
				if typeStr, ok := t.(string); ok {
					contentType = typeStr
				}
			}

			// Set content type header (unless already set in headers)
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", contentType)
			}

			// Send response
			w.WriteHeader(status)
			_, _ = w.Write(fileBytes)
			return
		}
	}

	// Extract body (fallback if no filename)
	body := ""
	if b, ok := data["body"]; ok {
		body = fmt.Sprintf("%v", b)
	}

	// Send response
	w.WriteHeader(status)
	if body != "" {
		_, _ = w.Write([]byte(body))
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

// GetRequest returns the request data for the context() builtin
// For spawn/run contexts, returns the Data field as-is
// For HTTP contexts, returns parsed HTTP request data
func (rc *RequestContext) GetRequest() any {
	// If generic context data was provided (spawn/run), return it as-is
	if rc.Data != nil {
		return rc.Data
	}

	// HTTP handler - parse and return HTTP request data
	if rc.Request == nil {
		return nil
	}

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
