package runtime

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
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
	Port                  int
	Address               string // bind address (default "0.0.0.0")
	TLSEnabled            bool
	CertFile              string
	KeyFile               string
	Timeout               time.Duration     // Socket-level read/write timeout
	RequestHandlerTimeout time.Duration     // Handler script execution timeout
	routes                map[string]*Route // key: "METHOD /path"
	sortedRouteKeys       []string          // Routes sorted by path length (descending)
	routeMutex            sync.RWMutex
	server                *http.Server
	Interpreter           *script.Interpreter // Interpreter for getting current script path
	FileReader            func(string) ([]byte, error)
	FileStatter           func(string) int64 // Returns mtime, 0 if error
	startedChan           chan error         // Channel to communicate startup errors
}

// gzipResponseWriter wraps http.ResponseWriter to compress with gzip
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.gzipWriter.Write(b)
}

func (w *gzipResponseWriter) Flush() error {
	return w.gzipWriter.Flush()
}

func (w *gzipResponseWriter) Close() error {
	return w.gzipWriter.Close()
}

// Route represents a registered HTTP route
type Route struct {
	Method      string
	Path        string
	HandlerPath string
	ScriptDir   string         // Directory of the script that registered this route (for handler path resolution)
	PathParams  []string       // Parameter names extracted from path pattern (e.g., ["id", "token"])
	PathRegex   *regexp.Regexp // Compiled regex for matching (nil if no params)
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

// extractPathParams extracts parameter names and creates a regex pattern from a path like "/users/:id/tokens/:token"
// Returns the parameter names and compiled regex pattern
func extractPathParams(path string) ([]string, *regexp.Regexp, error) {
	// Find all :paramName occurrences
	paramPattern := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := paramPattern.FindAllStringSubmatch(path, -1)

	if len(matches) == 0 {
		// No parameters - return nil for regex
		return nil, nil, nil
	}

	// Extract parameter names
	var paramNames []string
	for _, match := range matches {
		paramNames = append(paramNames, match[1])
	}

	// Convert path pattern to regex
	// /users/:id/tokens/:token → ^/users/([^/]+)/tokens/([^/]+)$
	regexPattern := paramPattern.ReplaceAllString(regexp.QuoteMeta(path), `([^/]+)`)
	regexPattern = "^" + regexPattern + "$"

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile path pattern regex: %w", err)
	}

	return paramNames, regex, nil
}

// matchPathPattern matches a request path against a route pattern and extracts parameters
// Returns the extracted parameters or nil if no match
func matchPathPattern(route *Route, requestPath string) map[string]any {
	if route.PathRegex == nil {
		// No parameters - use simple prefix matching
		return nil
	}

	// Use regex to match and extract
	matches := route.PathRegex.FindStringSubmatch(requestPath)
	if matches == nil {
		return nil
	}

	// Extract parameter values (matches[0] is full string, matches[1:] are captured groups)
	params := make(map[string]any)
	for i, paramName := range route.PathParams {
		if i+1 < len(matches) {
			params[paramName] = matches[i+1]
		}
	}

	return params
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

	// Extract path parameters
	pathParams, pathRegex, err := extractPathParams(path)
	if err != nil {
		return err
	}

	// Register route for each method
	for _, method := range methods {
		key := method + " " + path
		s.routes[key] = &Route{
			Method:      method,
			Path:        path,
			HandlerPath: handlerPath,
			ScriptDir:   "",
			PathParams:  pathParams,
			PathRegex:   pathRegex,
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

// findMatchingRoute finds the best matching route using pattern matching.
// Returns the route and extracted path parameters, or (nil, nil) if no match found.
// Must be called with routeMutex held (RLock).
func (s *HTTPServerValue) findMatchingRoute(method, path string) (*Route, map[string]any) {
	// Try exact method first, then wildcard
	for _, routeKey := range s.sortedRouteKeys {
		parts := strings.SplitN(routeKey, " ", 2)
		if len(parts) != 2 {
			continue
		}
		routeMethod := parts[0]
		route := s.routes[routeKey]

		// Check if path matches
		var params map[string]any
		if route.PathRegex != nil {
			// Pattern matching with parameters
			params = matchPathPattern(route, path)
			if params == nil {
				continue
			}
		} else {
			// Prefix matching (no parameters)
			if !strings.HasPrefix(path, route.Path) {
				continue
			}
		}

		// Check if method matches (exact or wildcard)
		if routeMethod == method || routeMethod == "*" {
			return route, params
		}
	}

	// If no exact method match found, try wildcard again
	for _, routeKey := range s.sortedRouteKeys {
		parts := strings.SplitN(routeKey, " ", 2)
		if len(parts) != 2 {
			continue
		}
		routeMethod := parts[0]
		route := s.routes[routeKey]

		if routeMethod != "*" {
			continue
		}

		// Check if path matches
		var params map[string]any
		if route.PathRegex != nil {
			params = matchPathPattern(route, path)
			if params == nil {
				continue
			}
		} else {
			if !strings.HasPrefix(path, route.Path) {
				continue
			}
		}

		return route, params
	}

	return nil, nil
}

// Start launches the HTTP server and blocks until the process receives a termination signal.
// This allows the script to handle cleanup code after the server stops.
// Returns an error if the server fails to bind to the port.
func (s *HTTPServerValue) Start() error {
	mux := http.NewServeMux()

	// Register catch-all handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w = &gzipResponseWriter{ResponseWriter: w, gzipWriter: gz}
		}

		// Find matching route using pattern matching (most specific first)
		s.routeMutex.RLock()
		route, pathParams := s.findMatchingRoute(r.Method, r.URL.Path)
		s.routeMutex.RUnlock()

		if route == nil {
			http.NotFound(w, r)
			return
		}

		// Handle request
		s.handleRequest(w, r, route, pathParams)
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
func (s *HTTPServerValue) handleRequest(w http.ResponseWriter, r *http.Request, route *Route, pathParams map[string]any) {
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
		Request:    r,
		Writer:     w,
		closed:     false,
		PathParams: pathParams,
		Frame:      frame,
		ExitChan:   make(chan any, 1),
		FileReader: s.FileReader,
	}

	// Parse handler script
	// Read file using provided reader (with fallback to embedded files)
	if s.FileReader == nil {
		if !ctx.closed {
			http.Error(w, "Server not properly configured: no file reader", 500)
		}
		return
	}

	// Resolve handler path relative to the script that registered the route
	resolvedHandlerPath := route.HandlerPath
	if route.ScriptDir != "" {
		resolvedHandlerPath = script.ResolveScriptPath(route.HandlerPath, filepath.Join(route.ScriptDir, "dummy.du"))
	}

	// Parse with caching (HTTP handlers are called repeatedly, avoid re-parsing each request)
	if s.Interpreter == nil {
		if !ctx.closed {
			http.Error(w, "Handler execution requires interpreter", 500)
		}
		return
	}

	program, err := s.Interpreter.ParseScript(resolvedHandlerPath)
	if err != nil {
		if !ctx.closed {
			http.Error(w, fmt.Sprintf("Handler script parse error: %v", err), 500)
		}
		return
	}

	// Create timeout context for handler execution
	handlerCtx, cancel := context.WithTimeout(context.Background(), s.RequestHandlerTimeout)
	defer cancel()

	// Execute handler script with timeout using unified ExecuteScript
	resultChan := make(chan *script.ScriptExecutionResult, 1)
	go func() {
		// Register request context in THIS goroutine
		gid := GetGoroutineID()

		// Create request() and response() functions to pass as context data
		requestFn := script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
			return ctx.GetRequest(), nil
		})

		responseFn := script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
			return ctx.GetResponse(), nil
		})

		// Create context data with request/response functions
		contextData := map[string]any{
			"request":  requestFn,
			"response": responseFn,
		}

		// Register request context with context data (MUST be after creating contextData)
		SetRequestContextWithData(gid, ctx, contextData)
		defer clearRequestContext(gid)

		// Set up context getter for context() builtin
		// The getter returns the data object (request/response functions)
		SetContextGetter(gid, func() any {
			rtCtx, ok := GetRequestContext(gid)
			if !ok {
				return nil
			}
			return rtCtx.Data
		})
		defer ClearContextGetter(gid)

		// Convert runtime RequestContext to script RequestContext for ExecuteScript
		scriptCtx := &script.RequestContext{
			Request:    ctx.Request,
			Writer:     ctx.Writer,
			Data:       contextData,
			Frame:      ctx.Frame,
			ExitChan:   ctx.ExitChan,
		}

		// Use unified ExecuteScript for statement-by-statement execution with breakpoint handling
		result := script.ExecuteScript(program, s.Interpreter, frame, scriptCtx, handlerCtx)
		resultChan <- result
	}()

	// Wait for execution or timeout
	var result *script.ScriptExecutionResult
	select {
	case result = <-resultChan:
		// Execution completed
	case <-handlerCtx.Done():
		// Timeout occurred
		if !ctx.closed {
			http.Error(w, "Handler script timeout exceeded", 504)
		}
		return
	}

	// Process result (ExecuteScript has already handled breakpoints and errors)
	if result.Error != nil {
		// Script had a non-recoverable error
		if !ctx.closed {
			http.Error(w, fmt.Sprintf("Handler script error: %v", result.Error), 500)
		}
		return
	}

	// Send response based on exit value or default
	exitValue := result.Value
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
		".html": "text/html",
		".htm":  "text/html",
		".txt":  "text/plain",
		".json": "application/json",
		".xml":  "application/xml",
		".css":  "text/css",
		".js":   "application/javascript",
		".mjs":  "application/javascript",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".ico":  "image/x-icon",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".gz":   "application/gzip",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".wav":  "audio/wav",
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

// SendResponse stores response data for handler processing
// (instead of writing immediately, uses same path as exit() via sendHTTPResponse)
func (rc *RequestContext) SendResponse(data map[string]any) error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	if rc.closed {
		return fmt.Errorf("context already closed: response already sent")
	}

	// Store response data for handler to process through sendHTTPResponse
	rc.ResponseData = data
	rc.closed = true
	return nil
}

// GetRequest returns the request data for the context() builtin
// For spawn/run contexts, returns the Data field as-is
// For HTTP contexts, returns parsed HTTP request data
func (rc *RequestContext) GetRequest() any {
	// HTTP handler - parse and return HTTP request data (check this FIRST)
	if rc.Request != nil {
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

		// Parse form data FIRST (before reading body, since ParseForm reads the body)
		formData := make(map[string]any)
		contentType := rc.Request.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// URL-encoded form data
			if err := rc.Request.ParseForm(); err == nil {
				for k, vv := range rc.Request.Form {
					if len(vv) == 1 {
						formData[k] = vv[0]
					} else {
						arr := make([]any, len(vv))
						for i, v := range vv {
							arr[i] = v
						}
						formData[k] = arr
					}
				}
			}
		} else if strings.Contains(contentType, "multipart/form-data") {
			// Multipart form data
			if err := rc.Request.ParseMultipartForm(32 << 20); err == nil { // 32MB max
				if rc.Request.MultipartForm != nil && rc.Request.MultipartForm.Value != nil {
					for k, vv := range rc.Request.MultipartForm.Value {
						if len(vv) == 1 {
							formData[k] = vv[0]
						} else {
							arr := make([]any, len(vv))
							for i, v := range vv {
								arr[i] = v
							}
							formData[k] = arr
						}
					}
				}
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

		result := map[string]any{
			"method":  rc.Request.Method,
			"path":    rc.Request.URL.Path,
			"headers": headers,
			"query":   query,
			"form":    formData,
			"body":    body,
		}

		// Include path params if available
		if rc.PathParams != nil {
			result["params"] = rc.PathParams
		}

		return result
	}

	// For spawn/run contexts, return the Data field as-is
	if rc.Data != nil {
		return script.DeepCopyAny(rc.Data)
	}

	return nil
}
