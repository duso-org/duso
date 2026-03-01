package runtime

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// CORSConfig holds CORS (Cross-Origin Resource Sharing) settings
type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string // ["*"] or specific origins
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// JWTConfig holds JWT (JSON Web Token) settings
type JWTConfig struct {
	Enabled  bool
	Secret   string
	Required bool
}

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
	ShowDirectoryListing  bool              // Show directory listing when no default file found
	DefaultFiles          []string          // Default filenames to try in order (e.g., index.html, index.md)
	CORS                  CORSConfig        // CORS configuration
	JWT                   JWTConfig         // JWT configuration
	routes                map[string]*Route // key: "METHOD /path"
	sortedRouteKeys       []string          // Routes sorted by path length (descending)
	routeMutex            sync.RWMutex
	server                *http.Server
	Interpreter           *script.Interpreter // Interpreter for getting current script path
	FileReader            func(string) ([]byte, error)
	FileStatter           func(string) int64 // Returns mtime, 0 if error
	DirReader             func(string) ([]map[string]any, error) // Lists directory contents, supports /EMBED/ and /STORE/
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
	IsStatic    bool           // True if this is a static file route
	StaticDir   string         // Directory to serve files from (for static routes)
}

// base64urlEncode encodes data using base64url encoding (no padding)
func base64urlEncode(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	// Remove padding and replace URL-unsafe characters
	encoded = strings.TrimRight(encoded, "=")
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	return encoded
}

// base64urlDecode decodes base64url-encoded data (with or without padding)
func base64urlDecode(data string) ([]byte, error) {
	// Add padding if needed
	switch len(data) % 4 {
	case 1:
		return nil, fmt.Errorf("illegal base64url string")
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	// Replace URL-safe characters with standard base64
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")

	return base64.StdEncoding.DecodeString(data)
}

// verifyJWT verifies a JWT token and returns claims if valid
func (s *HTTPServerValue) verifyJWT(tokenString string) (map[string]any, error) {
	if !s.JWT.Enabled || s.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT not configured")
	}

	// Split token into three parts: header.payload.signature
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerStr := parts[0]
	payloadStr := parts[1]
	signatureStr := parts[2]

	// Decode and verify header
	headerBytes, err := base64urlDecode(headerStr)
	if err != nil {
		return nil, fmt.Errorf("invalid header encoding: %w", err)
	}

	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid header JSON: %w", err)
	}

	// Check algorithm
	if alg, ok := header["alg"].(string); !ok || alg != "HS256" {
		return nil, fmt.Errorf("unsupported algorithm: expected HS256")
	}

	// Verify signature using HMAC-SHA256
	h := hmac.New(sha256.New, []byte(s.JWT.Secret))
	h.Write([]byte(headerStr + "." + payloadStr))
	expectedSigBytes := h.Sum(nil)

	// Decode the signature from the token
	decodedSigBytes, err := base64urlDecode(signatureStr)
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %w", err)
	}

	// Compare signatures using constant-time comparison
	if !hmac.Equal(decodedSigBytes, expectedSigBytes) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode payload
	payloadBytes, err := base64urlDecode(payloadStr)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %w", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("invalid payload JSON: %w", err)
	}

	// Check expiration if present
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	return claims, nil
}

// buildSignJWTFunction creates a sign_jwt function bound to a JWT secret
func buildSignJWTFunction(jwtSecret string) script.Value {
	return script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		claims, ok := args["0"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("sign_jwt() requires claims object as first argument")
		}

		// Get optional expires_in in seconds (default 3600 = 1 hour)
		expiresIn := 3600.0
		if exp, ok := args["1"]; ok {
			if expNum, ok := exp.(float64); ok {
				expiresIn = expNum
			}
		} else if exp, ok := args["expires_in"]; ok {
			if expNum, ok := exp.(float64); ok {
				expiresIn = expNum
			}
		}

		// Clone claims and add expiration
		tokenClaims := make(map[string]any)
		for k, v := range claims {
			tokenClaims[k] = v
		}
		tokenClaims["exp"] = float64(time.Now().Unix()) + expiresIn

		// Build header
		header := map[string]string{
			"alg": "HS256",
			"typ": "JWT",
		}

		// Encode header and payload
		headerJSON, _ := json.Marshal(header)
		payloadJSON, _ := json.Marshal(tokenClaims)

		headerB64 := base64urlEncode(headerJSON)
		payloadB64 := base64urlEncode(payloadJSON)
		message := headerB64 + "." + payloadB64

		// Sign with HMAC-SHA256
		signature := hmac.New(sha256.New, []byte(jwtSecret))
		signature.Write([]byte(message))
		signatureB64 := base64urlEncode(signature.Sum(nil))

		// Return token
		token := message + "." + signatureB64
		return token, nil
	})
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

// StaticRoute registers a static file route (thread-safe).
// Serves files from staticDir for requests matching the path prefix.
func (s *HTTPServerValue) StaticRoute(path, staticDir string) error {
	s.routeMutex.Lock()
	defer s.routeMutex.Unlock()

	// Initialize routes map if nil
	if s.routes == nil {
		s.routes = make(map[string]*Route)
	}

	// For static routes, register with GET method (most common use case)
	// Also register for HEAD to support conditional requests
	methods := []string{"GET", "HEAD"}

	for _, method := range methods {
		key := method + " " + path
		s.routes[key] = &Route{
			Method:      method,
			Path:        path,
			HandlerPath: "",
			ScriptDir:   "",
			PathParams:  nil,
			PathRegex:   nil,
			IsStatic:    true,
			StaticDir:   staticDir,
		}
	}

	// Rebuild sorted route keys
	s.rebuildSortedRoutes()

	return nil
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
			IsStatic:    false,
			StaticDir:   "",
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
	// If no routes have been registered, default to serving static files from current directory
	if len(s.routes) == 0 {
		s.StaticRoute("/", ".")
	}

	mux := http.NewServeMux()

	// Register catch-all handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS preflight if enabled
		if s.CORS.Enabled {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			isAllowed := false
			if len(s.CORS.AllowedOrigins) > 0 {
				for _, allowedOrigin := range s.CORS.AllowedOrigins {
					if allowedOrigin == "*" {
						isAllowed = true
						break
					}
					if origin == allowedOrigin {
						isAllowed = true
						break
					}
				}
			}

			if isAllowed || (len(s.CORS.AllowedOrigins) == 0) {
				// Set CORS headers
				if len(s.CORS.AllowedOrigins) > 0 && s.CORS.AllowedOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if isAllowed && origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}

				if len(s.CORS.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(s.CORS.AllowedMethods, ", "))
				}
				if len(s.CORS.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(s.CORS.AllowedHeaders, ", "))
				}
				if s.CORS.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if s.CORS.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", s.CORS.MaxAge))
				}

				// Handle OPTIONS preflight request
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
		}

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

		// Handle static file routes directly
		if route.IsStatic {
			// Construct request path relative to static directory
			requestPath := r.URL.Path
			var filePath string
			if route.Path == "/" {
				filePath = requestPath
			} else {
				if !strings.HasPrefix(requestPath, route.Path) {
					http.NotFound(w, r)
					return
				}
				filePath = strings.TrimPrefix(requestPath, route.Path)
			}
			filePath = strings.TrimPrefix(filePath, "/")

			fullPath := core.Join(route.StaticDir, filePath)

			// 1. Try to serve as a file
			if _, err := s.FileReader(fullPath); err == nil {
				response := map[string]any{
					"status":   200,
					"filename": fullPath,
				}
				s.sendHTTPResponse(w, response, "")
				return
			}

			// 2. Check if it's a directory
			if s.DirReader == nil {
				http.Error(w, "Server configuration error: DirReader not initialized for static file serving", 500)
				return
			}
			entries, err := s.DirReader(fullPath)
			if err == nil && entries != nil {
				// It's a directory - try default files in order
				for _, defaultFile := range s.DefaultFiles {
					defaultPath := core.Join(fullPath, defaultFile)
					if _, errFile := s.FileReader(defaultPath); errFile == nil {
						response := map[string]any{
							"status":   200,
							"filename": defaultPath,
						}
						s.sendHTTPResponse(w, response, "")
						return
					}
				}
				// No default files found, show directory listing or 404
				if s.ShowDirectoryListing {
					html := fmt.Sprintf("<pre>Directory: %s\n\n", strings.ReplaceAll(requestPath, "<", "&lt;"))

					for _, entryMap := range entries {
						name, _ := entryMap["name"].(string)
						// Skip parent directory entry
						if name == ".." {
							continue
						}
						isDir, _ := entryMap["is_dir"].(bool)
						path := requestPath
						if !strings.HasSuffix(path, "/") {
							path += "/"
						}
						path += name
						if isDir {
							path += "/"
						}
						html += fmt.Sprintf("<a href=\"%s\">%s</a>\n", path, name)
					}
					html += "</pre>"

					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(200)
					w.Write([]byte(html))
					return
				}
				http.NotFound(w, r)
				return
			}
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
	// TODO: Increment HTTP request counter
	// IncrementHTTPProcs() - metrics system disabled

	// Verify JWT if enabled (before spawning handler)
	var jwtClaims map[string]any
	var jwtSecret string = s.JWT.Secret

	if s.JWT.Enabled {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Extract bearer token
			const bearerPrefix = "Bearer "
			if strings.HasPrefix(authHeader, bearerPrefix) {
				token := authHeader[len(bearerPrefix):]
				claims, err := s.verifyJWT(token)
				if err != nil {
					if s.JWT.Required {
						// Token required but invalid
						http.Error(w, fmt.Sprintf("Invalid or expired token: %v", err), http.StatusUnauthorized)
						return
					}
					// Token invalid but not required - continue without claims
				} else {
					jwtClaims = claims
				}
			}
		} else if s.JWT.Required {
			// Token required but not provided
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
	}

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
		JWTClaims:  jwtClaims,
		JWTSecret:  jwtSecret,
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
	if route.ScriptDir != "" && !core.IsAbsoluteOrSpecial(route.HandlerPath) {
		// Only resolve if the handler path is relative (not absolute or special prefix)
		// and doesn't already start with the script directory (to avoid doubling)
		if !strings.HasPrefix(route.HandlerPath, route.ScriptDir+"/") {
			resolvedHandlerPath = script.ResolveScriptPathFromDir(route.HandlerPath, route.ScriptDir)
		}
	}

	// Update frame to use resolved path (so scriptDir is correct for load/save/etc)
	frame.Filename = resolvedHandlerPath

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

		// Create request() function
		requestFn := script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
			return ctx.GetRequest(), nil
		})

		// Create response() function
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

		// Create script RequestContext for ExecuteScript
		// Note: HTTP request/response are passed via contextData through the context getter
		scriptCtx := &script.RequestContext{
			Data:     contextData,
			Frame:    ctx.Frame,
			ExitChan: ctx.ExitChan,
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
			s.sendHTTPResponse(w, responseMap, route.ScriptDir)
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
		".html":     "text/html; charset=utf-8",
		".htm":      "text/html; charset=utf-8",
		".txt":      "text/plain; charset=utf-8",
		".md":       "text/markdown; charset=utf-8",
		".markdown": "text/markdown; charset=utf-8",
		".json":     "application/json; charset=utf-8",
		".xml":      "application/xml; charset=utf-8",
		".css":      "text/css; charset=utf-8",
		".js":       "application/javascript; charset=utf-8",
		".mjs":      "application/javascript; charset=utf-8",
		".ts":       "text/typescript; charset=utf-8",
		".tsx":      "text/typescript; charset=utf-8",
		".py":       "text/x-python; charset=utf-8",
		".go":       "text/x-go; charset=utf-8",
		".rs":       "text/x-rust; charset=utf-8",
		".rb":       "text/x-ruby; charset=utf-8",
		".java":     "text/x-java; charset=utf-8",
		".c":        "text/x-c; charset=utf-8",
		".cpp":      "text/x-c++src; charset=utf-8",
		".h":        "text/x-c; charset=utf-8",
		".sh":       "text/x-shellscript; charset=utf-8",
		".bash":     "text/x-shellscript; charset=utf-8",
		".du":       "text/x-duso; charset=utf-8",
		".yaml":     "text/yaml; charset=utf-8",
		".yml":      "text/yaml; charset=utf-8",
		".toml":     "text/x-toml; charset=utf-8",
		".ini":      "text/plain; charset=utf-8",
		".cfg":      "text/plain; charset=utf-8",
		".conf":     "text/plain; charset=utf-8",
		".png":      "image/png",
		".jpg":      "image/jpeg",
		".jpeg":     "image/jpeg",
		".gif":      "image/gif",
		".svg":      "image/svg+xml",
		".webp":     "image/webp",
		".ico":      "image/x-icon",
		".pdf":      "application/pdf",
		".zip":      "application/zip",
		".gz":       "application/gzip",
		".mp3":      "audio/mpeg",
		".mp4":      "video/mp4",
		".webm":     "video/webm",
		".wav":      "audio/wav",
	}

	// Find extension
	dotIdx := strings.LastIndex(filename, ".")
	if dotIdx == -1 {
		return "text/plain; charset=utf-8"
	}

	ext := strings.ToLower(filename[dotIdx:])
	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	// Default to text/plain for unknown types
	// This allows viewing unusual formats in the browser rather than forcing download
	return "text/plain; charset=utf-8"
}

// sendHTTPResponse is a helper that sends HTTP response from a data map
func (s *HTTPServerValue) sendHTTPResponse(w http.ResponseWriter, data map[string]any, scriptDir string) {
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
			var fileBytes []byte
			var err error

			// Get scriptDir from response if available
			responseScriptDir := ""
			if sd, ok := data["scriptDir"]; ok {
				responseScriptDir, _ = sd.(string)
			}

			// Waterfall: as-is, /STORE/, /STORE/+scriptDir, /EMBED/, /EMBED/+scriptDir
			attempts := []string{
				filenameStr,
				"/STORE/" + filenameStr,
				"/EMBED/" + filenameStr,
			}

			// Add scriptDir variants if available
			if responseScriptDir != "" {
				scriptDirFile := core.Join(responseScriptDir, filenameStr)
				attempts = append(attempts,
					"/STORE/" + scriptDirFile,
					"/EMBED/" + scriptDirFile,
				)
			}

			// Try each attempt in order
			for _, attempt := range attempts {
				fileBytes, err = s.FileReader(attempt)
				if err == nil {
					break
				}
			}

			if err != nil {
				w.WriteHeader(404)
				_, _ = w.Write([]byte("File not found"))
				return
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

		// Include JWT claims if available
		if rc.JWTClaims != nil {
			result["jwt_claims"] = rc.JWTClaims
		}

		return result
	}

	// For spawn/run contexts, return the Data field as-is
	if rc.Data != nil {
		return script.DeepCopyAny(rc.Data)
	}

	return nil
}

// GetResponse returns an object with response helper methods for use in HTTP handler scripts
// This is HTTP-specific and includes sign_jwt if JWT is configured
func (rc *RequestContext) GetResponse() map[string]any {
	// Create response helper object with methods
	respMethods := map[string]any{
		// json(data [, status]) - Send JSON response and exit
		"json": script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

			// Return response data as exit value (same as exit() does)
			responseData := map[string]any{
				"status": status,
				"body":   string(jsonBytes),
				"headers": map[string]any{
					"Content-Type": "application/json",
				},
			}
			return nil, &script.ExitExecution{Values: []any{responseData}}
		}),

		// text(data [, status]) - Send plain text response and exit
		"text": script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

			// Return response data as exit value (same as exit() does)
			responseData := map[string]any{
				"status": status,
				"body":   fmt.Sprintf("%v", data),
				"headers": map[string]any{
					"Content-Type": "text/plain; charset=utf-8",
				},
			}
			return nil, &script.ExitExecution{Values: []any{responseData}}
		}),

		// html(data [, status]) - Send HTML response and exit
		"html": script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

			// Return response data as exit value (same as exit() does)
			responseData := map[string]any{
				"status": status,
				"body":   fmt.Sprintf("%v", data),
				"headers": map[string]any{
					"Content-Type": "text/html; charset=utf-8",
				},
			}
			return nil, &script.ExitExecution{Values: []any{responseData}}
		}),

		// error(status [, message]) - Send error response and exit
		"error": script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

			// Return response data as exit value (same as exit() does)
			responseData := map[string]any{
				"status": status,
				"body":   body,
				"headers": map[string]any{
					"Content-Type": "text/plain; charset=utf-8",
				},
			}
			return nil, &script.ExitExecution{Values: []any{responseData}}
		}),

		// redirect(url [, status]) - Send redirect response and exit
		"redirect": script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

			// Return response data as exit value (same as exit() does)
			responseData := map[string]any{
				"status": status,
				"headers": map[string]any{
					"Location": fmt.Sprintf("%v", url),
				},
			}
			return nil, &script.ExitExecution{Values: []any{responseData}}
		}),

		// file(path [, status]) - Send file response and exit
		"file": script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

			filename := fmt.Sprintf("%v", path)

			gid := script.GetGoroutineID()
			var scriptDir string
			if ctx, ok := GetRequestContext(gid); ok && ctx.Frame != nil && ctx.Frame.Filename != "" {
				scriptDir = core.Dir(ctx.Frame.Filename)
			}

			// Return response data as exit value (same as exit() does)
			// Include scriptDir so HTTP server can do full path resolution waterfall
			responseData := map[string]any{
				"status":    status,
				"filename":  filename,
				"scriptDir": scriptDir,
			}
			return nil, &script.ExitExecution{Values: []any{responseData}}
		}),

		// response(data, status [, headers]) - Generic response and exit
		"response": script.NewGoFunction(func(evaluator *script.Evaluator, args map[string]any) (any, error) {
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

			// Return response data as exit value (same as exit() does)
			responseData := map[string]any{
				"status":  status,
				"body":    fmt.Sprintf("%v", data),
				"headers": headers,
			}
			return nil, &script.ExitExecution{Values: []any{responseData}}
		}),

		// sign_jwt(claims [, expires_in]) - Sign and return a JWT token (HTTP context only)
		"sign_jwt": buildSignJWTFunction(rc.JWTSecret),
	}

	return respMethods
}
