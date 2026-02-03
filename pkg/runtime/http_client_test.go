package runtime

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewHTTPClient_Basic tests creating a basic HTTP client
func TestNewHTTPClient_Basic(t *testing.T) {
	config := make(map[string]any)

	client, err := NewHTTPClient(config)
	if err != nil {
		t.Errorf("NewHTTPClient failed: %v", err)
	}

	if client == nil {
		t.Errorf("Expected client, got nil")
	}

	if client.client == nil {
		t.Errorf("Internal HTTP client not initialized")
	}
}

// TestNewHTTPClient_WithBaseURL tests creating client with base URL
func TestNewHTTPClient_WithBaseURL(t *testing.T) {
	config := map[string]any{
		"base_url": "https://api.example.com",
	}

	client, err := NewHTTPClient(config)
	if err != nil {
		t.Errorf("NewHTTPClient failed: %v", err)
	}

	if client.BaseURL != "https://api.example.com" {
		t.Errorf("Expected base_url 'https://api.example.com', got %q", client.BaseURL)
	}
}

// TestNewHTTPClient_WithTimeout tests creating client with timeout
func TestNewHTTPClient_WithTimeout(t *testing.T) {
	config := map[string]any{
		"timeout": 30.0, // 30 seconds
	}

	client, err := NewHTTPClient(config)
	if err != nil {
		t.Errorf("NewHTTPClient failed: %v", err)
	}

	expectedTimeout := 30 * time.Second
	if client.client.Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, client.client.Timeout)
	}
}

// TestNewHTTPClient_WithTimeout_Integer tests timeout with integer value
func TestNewHTTPClient_WithTimeout_Integer(t *testing.T) {
	config := map[string]any{
		"timeout": 15, // 15 seconds as int
	}

	client, err := NewHTTPClient(config)
	if err != nil {
		t.Errorf("NewHTTPClient failed: %v", err)
	}

	expectedTimeout := 15 * time.Second
	if client.client.Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, client.client.Timeout)
	}
}

// TestNewHTTPClient_InvalidTimeout tests invalid timeout value
func TestNewHTTPClient_InvalidTimeout(t *testing.T) {
	config := map[string]any{
		"timeout": "invalid",
	}

	_, err := NewHTTPClient(config)
	if err == nil {
		t.Errorf("Should error on invalid timeout")
	}

	if !strings.Contains(err.Error(), "timeout must be a number") {
		t.Errorf("Error should mention timeout must be number, got: %v", err)
	}
}

// TestNewHTTPClient_WithHeaders tests creating client with default headers
func TestNewHTTPClient_WithHeaders(t *testing.T) {
	config := map[string]any{
		"headers": map[string]any{
			"Authorization": "Bearer token123",
			"X-Custom":      "value",
		},
	}

	client, err := NewHTTPClient(config)
	if err != nil {
		t.Errorf("NewHTTPClient failed: %v", err)
	}

	if client.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Authorization header not set")
	}

	if client.Headers["X-Custom"] != "value" {
		t.Errorf("Custom header not set")
	}
}

// TestHTTPClientSend_GET tests sending a GET request
func TestHTTPClientSend_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("GET response"))
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"method": "GET",
		"url":    server.URL,
	}

	resp, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	if resp["status"] != 200.0 {
		t.Errorf("Expected status 200, got %v", resp["status"])
	}

	if resp["body"] != "GET response" {
		t.Errorf("Expected body 'GET response', got %v", resp["body"])
	}
}

// TestHTTPClientSend_POST tests sending a POST request
func TestHTTPClientSend_POST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.WriteHeader(201)
		w.Write([]byte("POST response"))
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"method": "POST",
		"url":    server.URL,
		"body":   "request body",
	}

	resp, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	if resp["status"] != 201.0 {
		t.Errorf("Expected status 201, got %v", resp["status"])
	}
}

// TestHTTPClientSend_DefaultMethod tests default method is GET
func TestHTTPClientSend_DefaultMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"url": server.URL,
	}

	_, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}
}

// TestHTTPClientSend_MissingURL tests that missing URL errors
func TestHTTPClientSend_MissingURL(t *testing.T) {
	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"method": "GET",
	}

	_, err := client.Send(req)
	if err == nil {
		t.Errorf("Should error on missing URL")
	}

	if !strings.Contains(err.Error(), "url") {
		t.Errorf("Error should mention url, got: %v", err)
	}
}

// TestHTTPClientSend_Headers tests request headers
func TestHTTPClientSend_Headers(t *testing.T) {
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(200)
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"url": server.URL,
		"headers": map[string]any{
			"X-Custom": "test-value",
		},
	}

	_, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	if receivedHeaders.Get("X-Custom") != "test-value" {
		t.Errorf("Custom header not sent")
	}
}

// TestHTTPClientSend_DefaultHeaders tests default headers from config
func TestHTTPClientSend_DefaultHeaders(t *testing.T) {
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(200)
	}))
	defer server.Close()

	config := map[string]any{
		"headers": map[string]any{
			"Authorization": "Bearer token",
		},
	}
	client, _ := NewHTTPClient(config)

	req := map[string]any{
		"url": server.URL,
	}

	_, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	if receivedHeaders.Get("Authorization") != "Bearer token" {
		t.Errorf("Default header not sent")
	}
}

// TestHTTPClientSend_HeaderOverride tests that request headers override defaults
func TestHTTPClientSend_HeaderOverride(t *testing.T) {
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(200)
	}))
	defer server.Close()

	config := map[string]any{
		"headers": map[string]any{
			"X-Default": "default_value",
		},
	}
	client, _ := NewHTTPClient(config)

	req := map[string]any{
		"url": server.URL,
		"headers": map[string]any{
			"X-Default": "override_value",
		},
	}

	_, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	if receivedHeaders.Get("X-Default") != "override_value" {
		t.Errorf("Request header should override default")
	}
}

// TestHTTPClientSend_WithBaseURL tests using base URL for relative requests
func TestHTTPClientSend_WithBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			t.Errorf("Expected path /api/users, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	config := map[string]any{
		"base_url": server.URL,
	}
	client, _ := NewHTTPClient(config)

	req := map[string]any{
		"url": "/api/users",
	}

	_, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}
}

// TestHTTPClientSend_AbsoluteURL tests that absolute URLs bypass base URL
func TestHTTPClientSend_AbsoluteURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	config := map[string]any{
		"base_url": "https://other.com",
	}
	client, _ := NewHTTPClient(config)

	req := map[string]any{
		"url": server.URL, // Absolute URL
	}

	_, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}
}

// TestHTTPClientSend_ResponseHeaders tests response headers parsing
func TestHTTPClientSend_ResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom", "custom-value")
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"url": server.URL,
	}

	resp, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	headers := resp["headers"].(map[string]any)
	if headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type not in response headers")
	}

	if headers["X-Custom"] != "custom-value" {
		t.Errorf("Custom header not in response headers")
	}
}

// TestHTTPClientSend_MultiValueResponseHeaders tests multi-value response headers
func TestHTTPClientSend_MultiValueResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "cookie1=value1")
		w.Header().Add("Set-Cookie", "cookie2=value2")
		w.WriteHeader(200)
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"url": server.URL,
	}

	resp, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	headers := resp["headers"].(map[string]any)
	setCookie := headers["Set-Cookie"]

	if arr, ok := setCookie.([]any); ok {
		if len(arr) != 2 {
			t.Errorf("Expected 2 Set-Cookie values, got %d", len(arr))
		}
	} else {
		t.Errorf("Expected Set-Cookie as array for multi-value header")
	}
}

// TestHTTPClientSend_ErrorResponse tests error status codes
func TestHTTPClientSend_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"url": server.URL,
	}

	resp, err := client.Send(req)
	if err != nil {
		t.Errorf("Send should not error on 404 response: %v", err)
	}

	if resp["status"] != 404.0 {
		t.Errorf("Expected status 404, got %v", resp["status"])
	}

	if resp["body"] != "Not Found" {
		t.Errorf("Expected body 'Not Found', got %v", resp["body"])
	}
}

// TestHTTPClientSend_RequestWithBody tests sending request body
func TestHTTPClientSend_RequestWithBody(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		w.WriteHeader(200)
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	requestBody := "test request body"
	req := map[string]any{
		"method": "POST",
		"url":    server.URL,
		"body":   requestBody,
	}

	_, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	if receivedBody != requestBody {
		t.Errorf("Expected body %q, got %q", requestBody, receivedBody)
	}
}

// TestHTTPClientClose tests closing the HTTP client
func TestHTTPClientClose(t *testing.T) {
	client, _ := NewHTTPClient(make(map[string]any))

	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestHTTPClientSend_InvalidURL tests invalid URL handling
func TestHTTPClientSend_InvalidURL(t *testing.T) {
	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"url": "invalid://[bad url",
	}

	_, err := client.Send(req)
	if err == nil {
		t.Errorf("Should error on invalid URL")
	}
}

// TestHTTPClientSend_ResponseBody tests response body reading
func TestHTTPClientSend_ResponseBody(t *testing.T) {
	expectedBody := `{"key": "value", "number": 42}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"url": server.URL,
	}

	resp, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	if resp["body"] != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, resp["body"])
	}
}

// TestIsAbsoluteURL tests URL classification
func TestIsAbsoluteURL(t *testing.T) {
	testCases := []struct {
		url      string
		expected bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"http://example.com/path", true},
		{"https://api.example.com:8080", true},
		{"/relative/path", false},
		{"relative/path", false},
		{"ftp://example.com", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isAbsoluteURL(tc.url)
		if result != tc.expected {
			t.Errorf("isAbsoluteURL(%q): expected %v, got %v", tc.url, tc.expected, result)
		}
	}
}

// TestStringReader tests the StringReader helper
func TestStringReader(t *testing.T) {
	sr := &StringReader{
		s:      "hello",
		offset: 0,
	}

	buf := make([]byte, 3)
	n, err := sr.Read(buf)

	if n != 3 {
		t.Errorf("Expected 3 bytes read, got %d", n)
	}

	if string(buf) != "hel" {
		t.Errorf("Expected 'hel', got %q", string(buf))
	}

	if err != nil {
		t.Errorf("First read should not error")
	}

	// Read remainder
	buf2 := make([]byte, 10)
	n2, err := sr.Read(buf2)
	if n2 != 2 {
		t.Errorf("Expected 2 bytes in second read, got %d", n2)
	}

	if string(buf2[:n2]) != "lo" {
		t.Errorf("Expected 'lo', got %q", string(buf2[:n2]))
	}

	// Read when at end should return EOF
	n3, err := sr.Read(buf2)
	if n3 != 0 || err != io.EOF {
		t.Errorf("Expected EOF, got n=%d err=%v", n3, err)
	}
}

// TestHTTPClientSend_ContentType tests Content-Type header setting
func TestHTTPClientSend_ContentType(t *testing.T) {
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(200)
	}))
	defer server.Close()

	client, _ := NewHTTPClient(make(map[string]any))

	req := map[string]any{
		"method": "POST",
		"url":    server.URL,
		"body":   "some data",
	}

	_, err := client.Send(req)
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}

	// Note: There's a debug print in http_client.go, but Content-Type should be set
	if receivedContentType == "" {
		t.Errorf("Content-Type should be set for POST with body")
	}
}

// TestNewHTTPClient_AllOptions tests creating client with all options
func TestNewHTTPClient_AllOptions(t *testing.T) {
	config := map[string]any{
		"base_url": "https://api.example.com",
		"timeout":  45.0,
		"headers": map[string]any{
			"Authorization": "Bearer token",
			"User-Agent":    "MyClient/1.0",
		},
	}

	client, err := NewHTTPClient(config)
	if err != nil {
		t.Errorf("NewHTTPClient failed: %v", err)
	}

	if client.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL not set correctly")
	}

	if client.client.Timeout != 45*time.Second {
		t.Errorf("Timeout not set correctly")
	}

	if len(client.Headers) != 2 {
		t.Errorf("Headers not set correctly, got %d headers", len(client.Headers))
	}
}
