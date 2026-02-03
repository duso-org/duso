package runtime

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// TestExtractPathParams_NoParams tests path without parameters
func TestExtractPathParams_NoParams(t *testing.T) {
	params, regex, err := extractPathParams("/users")
	if err != nil {
		t.Errorf("extractPathParams failed: %v", err)
	}

	if params != nil {
		t.Errorf("Expected nil params for path without parameters, got %v", params)
	}

	if regex != nil {
		t.Errorf("Expected nil regex for path without parameters, got %v", regex)
	}
}

// TestExtractPathParams_SingleParam tests path with single parameter
func TestExtractPathParams_SingleParam(t *testing.T) {
	params, regex, err := extractPathParams("/users/:id")
	if err != nil {
		t.Errorf("extractPathParams failed: %v", err)
	}

	if len(params) != 1 || params[0] != "id" {
		t.Errorf("Expected ['id'], got %v", params)
	}

	if regex == nil {
		t.Errorf("Expected regex, got nil")
	}

	// Test regex matching
	if !regex.MatchString("/users/123") {
		t.Errorf("Regex should match /users/123")
	}

	if regex.MatchString("/users/") {
		t.Errorf("Regex should not match /users/")
	}
}

// TestExtractPathParams_MultipleParams tests path with multiple parameters
func TestExtractPathParams_MultipleParams(t *testing.T) {
	params, regex, err := extractPathParams("/users/:id/tokens/:token")
	if err != nil {
		t.Errorf("extractPathParams failed: %v", err)
	}

	if len(params) != 2 || params[0] != "id" || params[1] != "token" {
		t.Errorf("Expected ['id', 'token'], got %v", params)
	}

	if regex == nil {
		t.Errorf("Expected regex, got nil")
	}

	if !regex.MatchString("/users/123/tokens/abc") {
		t.Errorf("Regex should match /users/123/tokens/abc")
	}
}

// TestMatchPathPattern_NoParams tests pattern matching without parameters
func TestMatchPathPattern_NoParams(t *testing.T) {
	route := &Route{
		Path:      "/users",
		PathRegex: nil,
	}

	params := matchPathPattern(route, "/users")
	if params != nil {
		t.Errorf("Expected nil params for no-param route, got %v", params)
	}
}

// TestMatchPathPattern_WithParams tests pattern matching with parameters
func TestMatchPathPattern_WithParams(t *testing.T) {
	paramNames := []string{"id"}
	regex := regexp.MustCompile(`^/users/([^/]+)$`)

	route := &Route{
		Path:       "/users/:id",
		PathRegex:  regex,
		PathParams: paramNames,
	}

	params := matchPathPattern(route, "/users/123")
	if params == nil {
		t.Errorf("Expected params, got nil")
	}

	if params["id"] != "123" {
		t.Errorf("Expected id='123', got %v", params["id"])
	}
}

// TestMatchPathPattern_NoMatch tests pattern matching with no match
func TestMatchPathPattern_NoMatch(t *testing.T) {
	paramNames := []string{"id"}
	regex := regexp.MustCompile(`^/users/([^/]+)$`)

	route := &Route{
		Path:       "/users/:id",
		PathRegex:  regex,
		PathParams: paramNames,
	}

	params := matchPathPattern(route, "/posts/123")
	if params != nil {
		t.Errorf("Expected nil for non-matching path, got %v", params)
	}
}

// TestIsValidHTTPMethod tests HTTP method validation
func TestIsValidHTTPMethod(t *testing.T) {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE", "CONNECT", "*"}
	for _, method := range validMethods {
		if !isValidHTTPMethod(method) {
			t.Errorf("Method %s should be valid", method)
		}
	}

	invalidMethods := []string{"INVALID", "get", "post", ""}
	for _, method := range invalidMethods {
		if isValidHTTPMethod(method) && method != "" {
			t.Errorf("Method %s should be invalid", method)
		}
	}
}

// TestHTTPServerRoute_BasicRegistration tests basic route registration
func TestHTTPServerRoute_BasicRegistration(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route("GET", "/users", "handlers/users.du")
	if err != nil {
		t.Errorf("Route registration failed: %v", err)
	}

	// Verify route was registered
	key := "GET /users"
	if _, exists := server.routes[key]; !exists {
		t.Errorf("Route not registered: %s", key)
	}
}

// TestHTTPServerRoute_MultipleMethodsString tests route with multiple methods as string
func TestHTTPServerRoute_MultipleMethodsString(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	// Register with "GET" string
	err := server.Route("GET", "/api/data", "handlers/data.du")
	if err != nil {
		t.Errorf("GET route registration failed: %v", err)
	}

	// Verify GET route exists
	if _, exists := server.routes["GET /api/data"]; !exists {
		t.Errorf("GET route not registered")
	}
}

// TestHTTPServerRoute_WildcardMethod tests route with wildcard method
func TestHTTPServerRoute_WildcardMethod(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route("*", "/catch-all", "handlers/catch.du")
	if err != nil {
		t.Errorf("Wildcard route registration failed: %v", err)
	}

	if _, exists := server.routes["* /catch-all"]; !exists {
		t.Errorf("Wildcard route not registered")
	}
}

// TestHTTPServerRoute_NilMethod tests route with nil method (all methods)
func TestHTTPServerRoute_NilMethod(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route(nil, "/any", "handlers/any.du")
	if err != nil {
		t.Errorf("Nil method route registration failed: %v", err)
	}

	if _, exists := server.routes["* /any"]; !exists {
		t.Errorf("Nil method route not registered")
	}
}

// TestHTTPServerRoute_InvalidMethod tests route with invalid method
func TestHTTPServerRoute_InvalidMethod(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route("INVALID", "/path", "handlers/path.du")
	if err == nil {
		t.Errorf("Should error on invalid method")
	}

	if !strings.Contains(err.Error(), "invalid HTTP method") {
		t.Errorf("Error should mention invalid method, got: %v", err)
	}
}

// TestHTTPServerRoute_CaseInsensitive tests that methods are case-insensitive
func TestHTTPServerRoute_CaseInsensitive(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route("get", "/users", "handlers/users.du")
	if err != nil {
		t.Errorf("Lowercase method registration failed: %v", err)
	}

	// Should be registered as uppercase "GET"
	if _, exists := server.routes["GET /users"]; !exists {
		t.Errorf("Method not converted to uppercase")
	}
}

// TestHTTPServerRoute_WithPathParams tests route with path parameters
func TestHTTPServerRoute_WithPathParams(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route("GET", "/users/:id", "handlers/user.du")
	if err != nil {
		t.Errorf("Parameterized route registration failed: %v", err)
	}

	route := server.routes["GET /users/:id"]
	if route == nil {
		t.Errorf("Parameterized route not found")
		return
	}

	if len(route.PathParams) != 1 || route.PathParams[0] != "id" {
		t.Errorf("Expected PathParams=['id'], got %v", route.PathParams)
	}

	if route.PathRegex == nil {
		t.Errorf("Expected PathRegex to be set")
	}
}

// TestHTTPServerFindMatchingRoute_ExactMatch tests finding exact route match
func TestHTTPServerFindMatchingRoute_ExactMatch(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	server.Route("GET", "/users", "handlers/users.du")
	server.rebuildSortedRoutes()

	route, params := server.findMatchingRoute("GET", "/users")
	if route == nil {
		t.Errorf("Route not found")
	}

	if route.HandlerPath != "handlers/users.du" {
		t.Errorf("Wrong handler path: %s", route.HandlerPath)
	}

	if params != nil {
		t.Errorf("Expected nil params, got %v", params)
	}
}

// TestHTTPServerFindMatchingRoute_PatternMatch tests finding parameterized route
func TestHTTPServerFindMatchingRoute_PatternMatch(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	server.Route("GET", "/users/:id", "handlers/user.du")
	server.rebuildSortedRoutes()

	route, params := server.findMatchingRoute("GET", "/users/123")
	if route == nil {
		t.Errorf("Route not found")
	}

	if params == nil {
		t.Errorf("Expected params, got nil")
	}

	if params["id"] != "123" {
		t.Errorf("Expected id='123', got %v", params["id"])
	}
}

// TestHTTPServerFindMatchingRoute_WildcardMethod tests wildcard method matching
func TestHTTPServerFindMatchingRoute_WildcardMethod(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	server.Route("*", "/catch-all", "handlers/catch.du")
	server.rebuildSortedRoutes()

	// Should match any method
	route, _ := server.findMatchingRoute("GET", "/catch-all")
	if route == nil {
		t.Errorf("Wildcard route not found for GET")
	}

	route, _ = server.findMatchingRoute("POST", "/catch-all")
	if route == nil {
		t.Errorf("Wildcard route not found for POST")
	}
}

// TestHTTPServerFindMatchingRoute_NoMatch tests no matching route
func TestHTTPServerFindMatchingRoute_NoMatch(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	server.Route("GET", "/users", "handlers/users.du")
	server.rebuildSortedRoutes()

	route, _ := server.findMatchingRoute("GET", "/posts")
	if route != nil {
		t.Errorf("Route should not be found for /posts")
	}
}

// TestHTTPServerFindMatchingRoute_MostSpecificFirst tests that most specific route is matched first
func TestHTTPServerFindMatchingRoute_MostSpecificFirst(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	// Register general route first
	server.Route("GET", "/users", "handlers/users.du")
	// Register specific route
	server.Route("GET", "/users/admin", "handlers/admin.du")
	server.rebuildSortedRoutes()

	route, _ := server.findMatchingRoute("GET", "/users/admin")
	if route == nil {
		t.Errorf("Route not found")
	}

	if route.HandlerPath != "handlers/admin.du" {
		t.Errorf("Expected more specific route /users/admin, got %s", route.HandlerPath)
	}
}

// TestGetContentType tests MIME type detection
func TestGetContentType(t *testing.T) {
	testCases := []struct {
		filename string
		expected string
	}{
		{"file.html", "text/html"},
		{"file.json", "application/json"},
		{"file.xml", "application/xml"},
		{"file.css", "text/css"},
		{"file.js", "application/javascript"},
		{"file.png", "image/png"},
		{"file.jpg", "image/jpeg"},
		{"file.pdf", "application/pdf"},
		{"file.zip", "application/zip"},
		{"file.unknown", "application/octet-stream"},
		{"noextension", "application/octet-stream"},
	}

	for _, tc := range testCases {
		result := getContentType(tc.filename)
		if result != tc.expected {
			t.Errorf("getContentType(%q): expected %q, got %q", tc.filename, tc.expected, result)
		}
	}
}

// TestHTTPServerRoute_MethodArray tests route with method as []string
func TestHTTPServerRoute_MethodArray(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	methods := []string{"GET", "POST"}
	err := server.Route(methods, "/api/data", "handlers/data.du")
	if err != nil {
		t.Errorf("Method array route registration failed: %v", err)
	}

	// Verify both routes were registered
	if _, exists := server.routes["GET /api/data"]; !exists {
		t.Errorf("GET route not registered")
	}

	if _, exists := server.routes["POST /api/data"]; !exists {
		t.Errorf("POST route not registered")
	}
}

// TestSendHTTPResponse_StatusCode tests response status code handling
func TestSendHTTPResponse_StatusCode(t *testing.T) {
	server := &HTTPServerValue{}
	recorder := httptest.NewRecorder()

	data := map[string]any{
		"status": 201.0,
		"body":   "Created",
	}

	server.sendHTTPResponse(recorder, data)

	if recorder.Code != 201 {
		t.Errorf("Expected status 201, got %d", recorder.Code)
	}
}

// TestSendHTTPResponse_DefaultStatus tests default status code
func TestSendHTTPResponse_DefaultStatus(t *testing.T) {
	server := &HTTPServerValue{}
	recorder := httptest.NewRecorder()

	data := map[string]any{
		"body": "OK",
	}

	server.sendHTTPResponse(recorder, data)

	if recorder.Code != 200 {
		t.Errorf("Expected default status 200, got %d", recorder.Code)
	}
}

// TestSendHTTPResponse_Headers tests response headers
func TestSendHTTPResponse_Headers(t *testing.T) {
	server := &HTTPServerValue{}
	recorder := httptest.NewRecorder()

	data := map[string]any{
		"body": "OK",
		"headers": map[string]any{
			"X-Custom-Header": "value",
			"Content-Type":    "text/plain",
		},
	}

	server.sendHTTPResponse(recorder, data)

	if recorder.Header().Get("X-Custom-Header") != "value" {
		t.Errorf("Custom header not set")
	}

	if recorder.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Content-Type header not set")
	}
}

// TestSendHTTPResponse_Body tests response body
func TestSendHTTPResponse_Body(t *testing.T) {
	server := &HTTPServerValue{}
	recorder := httptest.NewRecorder()

	bodyText := "Response body"
	data := map[string]any{
		"body": bodyText,
	}

	server.sendHTTPResponse(recorder, data)

	if recorder.Body.String() != bodyText {
		t.Errorf("Expected body %q, got %q", bodyText, recorder.Body.String())
	}
}

// TestSendHTTPResponse_NoBody tests response with empty body
func TestSendHTTPResponse_NoBody(t *testing.T) {
	server := &HTTPServerValue{}
	recorder := httptest.NewRecorder()

	data := map[string]any{}

	server.sendHTTPResponse(recorder, data)

	if recorder.Body.String() != "" {
		t.Errorf("Expected empty body, got %q", recorder.Body.String())
	}
}

// TestRequestContextSendResponse_Success tests sending response
func TestRequestContextSendResponse_Success(t *testing.T) {
	recorder := httptest.NewRecorder()
	rc := &RequestContext{
		Writer: recorder,
		closed: false,
	}

	data := map[string]any{
		"status": 200.0,
		"body":   "OK",
	}

	err := rc.SendResponse(data)
	if err != nil {
		t.Errorf("SendResponse failed: %v", err)
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	if recorder.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got %q", recorder.Body.String())
	}

	if !rc.closed {
		t.Errorf("Context should be marked as closed")
	}
}

// TestRequestContextSendResponse_AlreadyClosed tests sending response twice
func TestRequestContextSendResponse_AlreadyClosed(t *testing.T) {
	recorder := httptest.NewRecorder()
	rc := &RequestContext{
		Writer: recorder,
		closed: true,
	}

	data := map[string]any{
		"body": "OK",
	}

	err := rc.SendResponse(data)
	if err == nil {
		t.Errorf("Should error when context already closed")
	}

	if !strings.Contains(err.Error(), "already closed") {
		t.Errorf("Error should mention already closed, got: %v", err)
	}
}

// TestRequestContextGetRequest_HTTPRequest tests parsing HTTP request
func TestRequestContextGetRequest_HTTPRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/path?key=value", nil)
	rc := &RequestContext{
		Request: req,
	}

	result := rc.GetRequest()
	data := result.(map[string]any)

	if data["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", data["method"])
	}

	if data["path"] != "/path" {
		t.Errorf("Expected path /path, got %v", data["path"])
	}

	if query, ok := data["query"].(map[string]any); ok {
		if query["key"] != "value" {
			t.Errorf("Expected query key=value, got %v", query["key"])
		}
	} else {
		t.Errorf("Expected query to be a map")
	}
}

// TestRequestContextGetRequest_HTTPHeaders tests HTTP header parsing
func TestRequestContextGetRequest_HTTPHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom", "value")

	rc := &RequestContext{
		Request: req,
	}

	result := rc.GetRequest()
	data := result.(map[string]any)
	headers := data["headers"].(map[string]any)

	if headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type header not parsed correctly")
	}

	if headers["X-Custom"] != "value" {
		t.Errorf("Custom header not parsed correctly")
	}
}

// TestRequestContextGetRequest_WithPathParams tests path parameters in request
func TestRequestContextGetRequest_WithPathParams(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rc := &RequestContext{
		Request:    req,
		PathParams: map[string]any{"id": "123"},
	}

	result := rc.GetRequest()
	data := result.(map[string]any)

	if params, ok := data["params"].(map[string]any); ok {
		if params["id"] != "123" {
			t.Errorf("Expected id=123, got %v", params["id"])
		}
	} else {
		t.Errorf("Expected params in request data")
	}
}

// TestRequestContextGetRequest_WithData tests generic context data
func TestRequestContextGetRequest_WithData(t *testing.T) {
	rc := &RequestContext{
		Data: map[string]any{"spawn_key": "spawn_value"},
	}

	result := rc.GetRequest()
	data := result.(map[string]any)

	if data["spawn_key"] != "spawn_value" {
		t.Errorf("Expected spawn context data in request")
	}
}

// TestHTTPServerRoute_DoubleSlash tests route with double slashes
func TestHTTPServerRoute_DoubleSlash(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route("GET", "//double", "handlers/double.du")
	if err != nil {
		t.Errorf("Double slash route registration should work: %v", err)
	}
}

// TestHTTPServerRoute_RootPath tests route for root path
func TestHTTPServerRoute_RootPath(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route("GET", "/", "handlers/root.du")
	if err != nil {
		t.Errorf("Root route registration failed: %v", err)
	}

	route, _ := server.findMatchingRoute("GET", "/")
	if route == nil {
		t.Errorf("Root route not found")
	}
}

// TestHTTPServerRoute_EmptyStringMethod tests empty string method (wildcard)
func TestHTTPServerRoute_EmptyStringMethod(t *testing.T) {
	server := &HTTPServerValue{
		Port:  8080,
		routes: make(map[string]*Route),
	}

	err := server.Route("", "/any", "handlers/any.du")
	if err != nil {
		t.Errorf("Empty string method should be wildcard: %v", err)
	}

	if _, exists := server.routes["* /any"]; !exists {
		t.Errorf("Empty string method should register as wildcard")
	}
}
