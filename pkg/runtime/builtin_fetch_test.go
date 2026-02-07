package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// TestFetch_SuccessfulGET tests basic GET request to a test server
func TestFetch_SuccessfulGET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": server.URL}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Expected response object, got %T", result)
	}

	if status, ok := respObj["status"].(float64); !ok || status != 200 {
		t.Errorf("Expected status 200, got %v", respObj["status"])
	}

	if body, ok := respObj["body"].(string); !ok || body != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got %v", respObj["body"])
	}
}

// TestFetch_ResponseBody tests accessing .body property
func TestFetch_ResponseBody(t *testing.T) {
	expectedBody := "Test response body"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": server.URL}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	body, ok := respObj["body"].(string)
	if !ok || body != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, body)
	}
}

// TestFetch_ResponseStatus tests accessing .status property
func TestFetch_ResponseStatus(t *testing.T) {
	tests := []struct {
		statusCode int
		name       string
	}{
		{http.StatusOK, "200 OK"},
		{http.StatusCreated, "201 Created"},
		{http.StatusBadRequest, "400 BadRequest"},
		{http.StatusNotFound, "404 NotFound"},
		{http.StatusInternalServerError, "500 InternalServerError"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			interp := &script.Interpreter{}
			fn := NewFetchFunction(interp)
			args := map[string]any{"0": server.URL}

			result, err := fn(nil, args)
			if err != nil {
				t.Fatalf("fetch() failed: %v", err)
			}

			respObj := result.(map[string]any)
			status, ok := respObj["status"].(float64)
			if !ok || int(status) != tc.statusCode {
				t.Errorf("Expected status %d, got %v", tc.statusCode, respObj["status"])
			}
		})
	}
}

// TestFetch_ResponseHeaders tests accessing .headers property
func TestFetch_ResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "CustomValue")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": server.URL}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	headers, ok := respObj["headers"].(map[string]any)
	if !ok {
		t.Fatalf("Expected headers map, got %T", respObj["headers"])
	}

	if len(headers) == 0 {
		t.Errorf("Expected headers to be populated")
	}
}

// TestFetch_JSONMethod tests calling .json() method
func TestFetch_JSONMethod(t *testing.T) {
	expectedData := map[string]any{
		"name":  "Alice",
		"age":   30.0,
		"email": "alice@example.com",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data, _ := json.Marshal(expectedData)
		w.Write(data)
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": server.URL}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	jsonVal := respObj["json"]

	// Extract the function from the Value
	var jsonFn script.GoFunction
	if fnVal, ok := jsonVal.(script.Value); ok && fnVal.Type == script.VAL_FUNCTION {
		var ok bool
		jsonFn, ok = fnVal.Data.(script.GoFunction)
		if !ok {
			t.Fatalf("Expected GoFunction, got %T", fnVal.Data)
		}
	} else {
		t.Fatalf("Expected Value with VAL_FUNCTION type, got %T", jsonVal)
	}

	// Call .json() method
	parsed, err := jsonFn(nil, make(map[string]any))
	if err != nil {
		t.Fatalf("json() method failed: %v", err)
	}

	parsedObj, ok := parsed.(map[string]any)
	if !ok {
		t.Fatalf("Expected parsed JSON object, got %T", parsed)
	}

	if name, ok := parsedObj["name"].(string); !ok || name != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", parsedObj["name"])
	}

	if age, ok := parsedObj["age"].(float64); !ok || age != 30.0 {
		t.Errorf("Expected age 30, got %v", parsedObj["age"])
	}
}

// TestFetch_TextMethod tests calling .text() method
func TestFetch_TextMethod(t *testing.T) {
	expectedText := "Plain text response"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedText))
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": server.URL}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	textVal := respObj["text"]

	// Extract the function from the Value
	var textFn script.GoFunction
	if fnVal, ok := textVal.(script.Value); ok && fnVal.Type == script.VAL_FUNCTION {
		var ok bool
		textFn, ok = fnVal.Data.(script.GoFunction)
		if !ok {
			t.Fatalf("Expected GoFunction, got %T", fnVal.Data)
		}
	} else {
		t.Fatalf("Expected Value with VAL_FUNCTION type, got %T", textVal)
	}

	// Call .text() method
	text, err := textFn(nil, make(map[string]any))
	if err != nil {
		t.Fatalf("text() method failed: %v", err)
	}

	if textStr, ok := text.(string); !ok || textStr != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, text)
	}
}

// TestFetch_OKProperty tests the .ok property (status < 400)
func TestFetch_OKProperty(t *testing.T) {
	tests := []struct {
		statusCode int
		expectOK   bool
		name       string
	}{
		{http.StatusOK, true, "200 OK"},
		{http.StatusCreated, true, "201 Created"},
		{http.StatusBadRequest, false, "400 BadRequest"},
		{http.StatusNotFound, false, "404 NotFound"},
		{http.StatusInternalServerError, false, "500 InternalServerError"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			interp := &script.Interpreter{}
			fn := NewFetchFunction(interp)
			args := map[string]any{"0": server.URL}

			result, err := fn(nil, args)
			if err != nil {
				t.Fatalf("fetch() failed: %v", err)
			}

			respObj := result.(map[string]any)
			ok, _ := respObj["ok"].(bool)
			if ok != tc.expectOK {
				t.Errorf("Expected ok=%v, got %v", tc.expectOK, ok)
			}
		})
	}
}

// TestFetch_MissingURL tests error when URL is missing
func TestFetch_MissingURL(t *testing.T) {
	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{} // No URL

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for missing URL, got nil")
	}
}

// TestFetch_InvalidURL tests error with invalid URL
func TestFetch_InvalidURL(t *testing.T) {
	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": "not-a-valid-url!@#$"}

	_, err := fn(nil, args)
	if err == nil {
		t.Errorf("Expected error for invalid URL, got nil")
	}
}

// TestFetch_JSONParseError tests error when calling .json() on non-JSON response
func TestFetch_JSONParseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": server.URL}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	jsonVal := respObj["json"]

	// Extract the function from the Value
	var jsonFn script.GoFunction
	if fnVal, ok := jsonVal.(script.Value); ok && fnVal.Type == script.VAL_FUNCTION {
		var ok bool
		jsonFn, ok = fnVal.Data.(script.GoFunction)
		if !ok {
			t.Fatalf("Expected GoFunction, got %T", fnVal.Data)
		}
	} else {
		t.Fatalf("Expected Value with VAL_FUNCTION type, got %T", jsonVal)
	}

	// Call .json() on non-JSON response
	_, err = jsonFn(nil, make(map[string]any))
	if err == nil {
		t.Errorf("Expected error for JSON parse failure, got nil")
	}
}

// TestFetch_POSTRequest tests POST request with body
func TestFetch_POSTRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{
		"0": server.URL,
		"1": map[string]any{
			"method": "POST",
			"body":   `{"name": "Alice"}`,
		},
	}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	status, _ := respObj["status"].(float64)
	if int(status) != http.StatusCreated {
		t.Errorf("Expected status 201, got %v", respObj["status"])
	}
}

// TestFetch_PUTRequest tests PUT request
func TestFetch_PUTRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"updated": true}`))
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{
		"0": server.URL,
		"1": map[string]any{
			"method": "PUT",
			"body":   `{"id": 1}`,
		},
	}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	status, _ := respObj["status"].(float64)
	if int(status) != http.StatusOK {
		t.Errorf("Expected status 200, got %v", respObj["status"])
	}
}

// TestFetch_DELETERequest tests DELETE request
func TestFetch_DELETERequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{
		"0": server.URL,
		"1": map[string]any{
			"method": "DELETE",
		},
	}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	status, _ := respObj["status"].(float64)
	if int(status) != http.StatusNoContent {
		t.Errorf("Expected status 204, got %v", respObj["status"])
	}
}

// TestFetch_WithHeaders tests sending custom request headers
func TestFetch_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer token123" {
			t.Errorf("Expected Authorization header, got %q", auth)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authenticated"))
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{
		"0": server.URL,
		"1": map[string]any{
			"headers": map[string]any{
				"Authorization": "Bearer token123",
			},
		},
	}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	body, _ := respObj["body"].(string)
	if body != "authenticated" {
		t.Errorf("Expected 'authenticated', got %q", body)
	}
}

// TestFetch_WithTimeout tests timeout option
func TestFetch_WithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{
		"0": server.URL,
		"1": map[string]any{
			"timeout": 0.5, // 500ms timeout
		},
	}

	_, err := fn(nil, args)
	// Should timeout
	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}
}

// TestFetch_EmptyResponse tests handling empty response body
func TestFetch_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		// No body
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": server.URL}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	body, _ := respObj["body"].(string)
	if body != "" {
		t.Errorf("Expected empty body, got %q", body)
	}
}

// TestFetch_LargeResponse tests handling large response body
func TestFetch_LargeResponse(t *testing.T) {
	largeData := make([]map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		largeData[i] = map[string]any{
			"id":   float64(i),
			"name": fmt.Sprintf("Item %d", i),
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data, _ := json.Marshal(largeData)
		w.Write(data)
	}))
	defer server.Close()

	interp := &script.Interpreter{}
	fn := NewFetchFunction(interp)
	args := map[string]any{"0": server.URL}

	result, err := fn(nil, args)
	if err != nil {
		t.Fatalf("fetch() failed: %v", err)
	}

	respObj := result.(map[string]any)
	jsonVal := respObj["json"]

	// Extract the function from the Value
	var jsonFn script.GoFunction
	if fnVal, ok := jsonVal.(script.Value); ok && fnVal.Type == script.VAL_FUNCTION {
		var ok bool
		jsonFn, ok = fnVal.Data.(script.GoFunction)
		if !ok {
			t.Fatalf("Expected GoFunction, got %T", fnVal.Data)
		}
	} else {
		t.Fatalf("Expected Value with VAL_FUNCTION type, got %T", jsonVal)
	}

	parsed, err := jsonFn(nil, make(map[string]any))
	if err != nil {
		t.Fatalf("json() method failed: %v", err)
	}

	if parsedSlice, ok := parsed.([]any); !ok || len(parsedSlice) != 1000 {
		t.Errorf("Expected array of 1000 items, got %T with length %v", parsed, len(parsedSlice))
	}
}
