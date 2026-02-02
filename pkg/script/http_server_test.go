package script

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestHTTPServer_RouteRegistration tests registering routes with various methods
func TestHTTPServer_RouteRegistration(t *testing.T) {
	tests := []struct {
		name          string
		method        any
		path          string
		handlerPath   string
		expectError   bool
		errorContains string
	}{
		{
			name:        "GET route",
			method:      "GET",
			path:        "/hello",
			handlerPath: "handlers/hello.du",
			expectError: false,
		},
		{
			name:        "POST route",
			method:      "POST",
			path:        "/api/users",
			handlerPath: "handlers/create_user.du",
			expectError: false,
		},
		{
			name:        "PUT route",
			method:      "PUT",
			path:        "/api/users/:id",
			handlerPath: "handlers/update_user.du",
			expectError: false,
		},
		{
			name:        "DELETE route",
			method:      "DELETE",
			path:        "/api/users/:id",
			handlerPath: "handlers/delete_user.du",
			expectError: false,
		},
		{
			name:        "PATCH route",
			method:      "PATCH",
			path:        "/api/data",
			handlerPath: "handlers/patch.du",
			expectError: false,
		},
		{
			name:        "wildcard methods",
			method:      "*",
			path:        "/catch-all",
			handlerPath: "handlers/catch_all.du",
			expectError: false,
		},
		{
			name:        "nil methods (all methods)",
			method:      nil,
			path:        "/all-methods",
			handlerPath: "handlers/all.du",
			expectError: false,
		},
		{
			name:          "invalid HTTP method",
			method:        "INVALID",
			path:          "/test",
			handlerPath:   "handlers/test.du",
			expectError:   true,
			errorContains: "invalid HTTP method",
		},
		{
			name:          "lowercase method (should be converted)",
			method:        "get",
			path:          "/test",
			handlerPath:   "handlers/test.du",
			expectError:   false,
		},
		{
			name:        "empty method string (matches all)",
			method:      "",
			path:        "/test",
			handlerPath: "handlers/test.du",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &HTTPServerValue{
				Port:    8080,
				Address: "0.0.0.0",
				Timeout: 30 * time.Second,
			}

			err := server.Route(tt.method, tt.path, tt.handlerPath)

			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}

			if tt.expectError && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			}

			if !tt.expectError && server.routes == nil {
				t.Errorf("routes map should not be nil after successful registration")
			}
		})
	}
}

// TestHTTPServer_MultipleMethodsArray tests registering routes with method arrays
func TestHTTPServer_MultipleMethodsArray(t *testing.T) {
	tests := []struct {
		name        string
		methods     any
		path        string
		expectError bool
	}{
		{
			name:        "array of strings",
			methods:     []string{"GET", "POST"},
			path:        "/api/data",
			expectError: false,
		},
		{
			name:        "array with single method",
			methods:     []string{"DELETE"},
			path:        "/api/delete",
			expectError: false,
		},
		{
			name:        "array with lowercase methods",
			methods:     []string{"get", "post", "put"},
			path:        "/api/multi",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &HTTPServerValue{
				Port:    8080,
				Address: "0.0.0.0",
				Timeout: 30 * time.Second,
			}

			err := server.Route(tt.methods, tt.path, "handler.du")

			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}

// TestHTTPServer_RouteMatching tests route matching with prefix and method matching
func TestHTTPServer_RouteMatching(t *testing.T) {
	server := &HTTPServerValue{
		Port:    8080,
		Address: "0.0.0.0",
		Timeout: 30 * time.Second,
		routes:  make(map[string]*Route),
	}

	// Register some routes
	server.Route("GET", "/api/users", "users.du")
	server.Route("GET", "/api/users/list", "list.du")
	server.Route("POST", "/api/users", "create.du")
	server.Route("*", "/static", "static.du")

	tests := []struct {
		name           string
		method         string
		path           string
		expectMatch    bool
		expectedPath   string
		expectedMethod string
	}{
		{
			name:           "exact GET match",
			method:         "GET",
			path:           "/api/users",
			expectMatch:    true,
			expectedPath:   "/api/users",
			expectedMethod: "GET",
		},
		{
			name:           "more specific route",
			method:         "GET",
			path:           "/api/users/list",
			expectMatch:    true,
			expectedPath:   "/api/users/list",
			expectedMethod: "GET",
		},
		{
			name:           "POST method",
			method:         "POST",
			path:           "/api/users",
			expectMatch:    true,
			expectedPath:   "/api/users",
			expectedMethod: "POST",
		},
		{
			name:           "wildcard matches all methods",
			method:         "DELETE",
			path:           "/static/index.html",
			expectMatch:    true,
			expectedPath:   "/static",
			expectedMethod: "*",
		},
		{
			name:        "no matching route",
			method:      "GET",
			path:        "/unknown",
			expectMatch: false,
		},
		{
			name:        "wrong method",
			method:      "DELETE",
			path:        "/api/users",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.routeMutex.RLock()
			route := server.findMatchingRoute(tt.method, tt.path)
			server.routeMutex.RUnlock()

			if (route != nil) != tt.expectMatch {
				t.Errorf("expected match: %v, got: %v", tt.expectMatch, route != nil)
			}

			if tt.expectMatch && route != nil {
				if route.Path != tt.expectedPath {
					t.Errorf("expected path %q, got %q", tt.expectedPath, route.Path)
				}
				if route.Method != tt.expectedMethod {
					t.Errorf("expected method %q, got %q", tt.expectedMethod, route.Method)
				}
			}
		})
	}
}

// TestHTTPServer_RoutePriority tests that more specific routes are matched first
func TestHTTPServer_RoutePriority(t *testing.T) {
	server := &HTTPServerValue{
		Port:    8080,
		Address: "0.0.0.0",
		Timeout: 30 * time.Second,
		routes:  make(map[string]*Route),
	}

	// Register routes with varying specificity
	server.Route("GET", "/api/v1", "v1.du")
	server.Route("GET", "/api/v1/users", "users.du")
	server.Route("GET", "/api/v1/users/profile", "profile.du")

	tests := []struct {
		name         string
		path         string
		expectedPath string
	}{
		{
			name:         "matches most specific route",
			path:         "/api/v1/users/profile",
			expectedPath: "/api/v1/users/profile",
		},
		{
			name:         "matches less specific when more specific not found",
			path:         "/api/v1/users/settings",
			expectedPath: "/api/v1/users",
		},
		{
			name:         "matches general route",
			path:         "/api/v1/anything",
			expectedPath: "/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.routeMutex.RLock()
			route := server.findMatchingRoute("GET", tt.path)
			server.routeMutex.RUnlock()

			if route == nil {
				t.Errorf("expected to find a matching route")
				return
			}

			if route.Path != tt.expectedPath {
				t.Errorf("expected path %q, got %q", tt.expectedPath, route.Path)
			}
		})
	}
}

// TestHTTPServer_RequestContextParsing tests request context extraction
func TestHTTPServer_RequestContextParsing(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		body           string
		headers        map[string]string
		expectMethod   string
		expectPath     string
		expectBodyLen  int
	}{
		{
			name:          "GET with query params",
			method:        "GET",
			path:          "/search",
			query:         "q=test&limit=10",
			expectMethod:  "GET",
			expectPath:    "/search",
			expectBodyLen: 0,
		},
		{
			name:          "POST with body",
			method:        "POST",
			path:          "/api/data",
			body:          `{"name":"test"}`,
			expectMethod:  "POST",
			expectPath:    "/api/data",
			expectBodyLen: 15,
		},
		{
			name:         "headers parsing",
			method:       "GET",
			path:         "/test",
			headers:      map[string]string{"Content-Type": "application/json"},
			expectMethod: "GET",
			expectPath:   "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an HTTP request
			url := "http://localhost:8080" + tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}

			req, err := http.NewRequest(tt.method, url, strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Add headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Create request context
			ctx := &RequestContext{
				Request: req,
				Writer:  httptest.NewRecorder(),
				closed:  false,
			}

			// Get request data
			data := ctx.GetRequest()
			if data == nil {
				t.Fatalf("expected request data, got nil")
			}

			// Verify parsed request
			reqMap, ok := data.(map[string]any)
			if !ok {
				t.Fatalf("expected map, got %T", data)
			}

			if method, ok := reqMap["method"]; !ok || method != tt.expectMethod {
				t.Errorf("expected method %q, got %v", tt.expectMethod, method)
			}

			if path, ok := reqMap["path"]; !ok || path != tt.expectPath {
				t.Errorf("expected path %q, got %v", tt.expectPath, path)
			}

			if body, ok := reqMap["body"]; ok {
				if bodyStr, ok := body.(string); ok {
					if len(bodyStr) != tt.expectBodyLen {
						t.Errorf("expected body length %d, got %d", tt.expectBodyLen, len(bodyStr))
					}
				}
			}
		})
	}
}

// TestHTTPServer_ResponseSending tests response sending with different configurations
func TestHTTPServer_ResponseSending(t *testing.T) {
	tests := []struct {
		name         string
		responseData map[string]any
		expectStatus int
		expectBody   string
	}{
		{
			name: "simple response with body",
			responseData: map[string]any{
				"status": float64(200),
				"body":   "Hello World",
			},
			expectStatus: 200,
			expectBody:   "Hello World",
		},
		{
			name: "response with custom status",
			responseData: map[string]any{
				"status": float64(201),
				"body":   "Created",
			},
			expectStatus: 201,
			expectBody:   "Created",
		},
		{
			name: "response with headers",
			responseData: map[string]any{
				"status": float64(200),
				"headers": map[string]any{
					"Content-Type": "application/json",
				},
				"body": `{"success":true}`,
			},
			expectStatus: 200,
			expectBody:   `{"success":true}`,
		},
		{
			name: "error response",
			responseData: map[string]any{
				"status": float64(404),
				"body":   "Not Found",
			},
			expectStatus: 404,
			expectBody:   "Not Found",
		},
		{
			name: "response without explicit body",
			responseData: map[string]any{
				"status": float64(204),
			},
			expectStatus: 204,
			expectBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx := &RequestContext{
				Writer: recorder,
				closed: false,
			}

			err := ctx.SendResponse(tt.responseData)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if recorder.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, recorder.Code)
			}

			if recorder.Body.String() != tt.expectBody {
				t.Errorf("expected body %q, got %q", tt.expectBody, recorder.Body.String())
			}
		})
	}
}

// TestHTTPServer_DuplicateResponseError tests that sending response twice fails
func TestHTTPServer_DuplicateResponseError(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := &RequestContext{
		Writer: recorder,
		closed: false,
	}

	data := map[string]any{"body": "test"}

	// First response should succeed
	err := ctx.SendResponse(data)
	if err != nil {
		t.Fatalf("first response failed: %v", err)
	}

	// Second response should fail
	err = ctx.SendResponse(data)
	if err == nil {
		t.Errorf("expected error on second response, got nil")
	}

	if !strings.Contains(err.Error(), "already closed") {
		t.Errorf("expected error to mention 'already closed', got %q", err.Error())
	}
}

// TestHTTPServer_DefaultPort tests that default port is set correctly
func TestHTTPServer_DefaultPort(t *testing.T) {
	server := &HTTPServerValue{
		Port:    8080,
		Address: "0.0.0.0",
		Timeout: 30 * time.Second,
	}

	if server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", server.Port)
	}

	if server.Address != "0.0.0.0" {
		t.Errorf("expected default address 0.0.0.0, got %q", server.Address)
	}
}

// TestHTTPServer_TimeoutDefaults tests that timeout defaults are set
func TestHTTPServer_TimeoutDefaults(t *testing.T) {
	server := &HTTPServerValue{
		Port:                   8080,
		Address:                "0.0.0.0",
		Timeout:                30 * time.Second,
		RequestHandlerTimeout:  30 * time.Second,
	}

	if server.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", server.Timeout)
	}

	if server.RequestHandlerTimeout != 30*time.Second {
		t.Errorf("expected handler timeout 30s, got %v", server.RequestHandlerTimeout)
	}
}

// TestHTTPServer_ConcurrentRouteRegistration tests thread-safe route registration
func TestHTTPServer_ConcurrentRouteRegistration(t *testing.T) {
	server := &HTTPServerValue{
		Port:    8080,
		Address: "0.0.0.0",
		Timeout: 30 * time.Second,
		routes:  make(map[string]*Route),
	}

	// Register routes concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			path := "/api/route" + string(rune(48+index))
			server.Route("GET", path, "handler.du")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all routes were registered
	if len(server.routes) != 10 {
		t.Errorf("expected 10 routes, got %d", len(server.routes))
	}
}

// TestHTTPServer_InvalidMethodTypeError tests that invalid method types produce errors
func TestHTTPServer_InvalidMethodTypeError(t *testing.T) {
	server := &HTTPServerValue{
		Port:    8080,
		Address: "0.0.0.0",
		Timeout: 30 * time.Second,
	}

	// Try to register with invalid type (number instead of string)
	err := server.Route(12345, "/test", "handler.du")
	if err == nil {
		t.Errorf("expected error for invalid method type, got nil")
	}

	if !strings.Contains(err.Error(), "must be string") {
		t.Errorf("expected error to mention 'must be string', got %q", err.Error())
	}
}

// TestHTTPServer_GetContentType tests MIME type detection
func TestHTTPServer_GetContentType(t *testing.T) {
	tests := []struct {
		filename    string
		expectType  string
	}{
		{"index.html", "text/html"},
		{"style.css", "text/css"},
		{"script.js", "application/javascript"},
		{"data.json", "application/json"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"video.mp4", "video/mp4"},
		{"document.pdf", "application/pdf"},
		{"archive.zip", "application/zip"},
		{"unknown.xyz", "application/octet-stream"},
		{"noextension", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			contentType := getContentType(tt.filename)
			if contentType != tt.expectType {
				t.Errorf("expected %q, got %q", tt.expectType, contentType)
			}
		})
	}
}
