package runtime

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HTTPClientValue represents a stateful HTTP client in Duso.
// It wraps Go's net/http.Client and provides methods for sending requests.
type HTTPClientValue struct {
	client  *http.Client
	BaseURL string            // Optional base URL for relative requests
	Headers map[string]string // Default headers for all requests
	Config  map[string]any    // Full config map (base_url, timeout, etc)
}

// NewHTTPClient creates a new HTTP client from Duso configuration.
func NewHTTPClient(config map[string]any) (*HTTPClientValue, error) {
	hc := &HTTPClientValue{
		client:  &http.Client{},
		Headers: make(map[string]string),
		Config:  config,
	}

	// Parse base_url
	if baseURL, ok := config["base_url"]; ok && baseURL != nil {
		hc.BaseURL = fmt.Sprintf("%v", baseURL)
	}

	// Parse timeout in seconds
	if timeout, ok := config["timeout"]; ok && timeout != nil {
		var timeoutSecs float64
		switch v := timeout.(type) {
		case float64:
			timeoutSecs = v
		case int:
			timeoutSecs = float64(v)
		default:
			return nil, fmt.Errorf("timeout must be a number (seconds): got %v", v)
		}
		hc.client.Timeout = time.Duration(timeoutSecs*1000) * time.Millisecond
	}

	// Parse default headers
	if headers, ok := config["headers"]; ok && headers != nil {
		if headerMap, ok := headers.(map[string]any); ok {
			for k, v := range headerMap {
				if v != nil {
					hc.Headers[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	return hc, nil
}

// Send executes an HTTP request and returns a response object.
// Request object structure: {method, url, body, headers, query}
// Response object structure: {status, body, headers}
func (hc *HTTPClientValue) Send(requestObj map[string]any) (map[string]any, error) {
	// Extract request fields
	method, ok := requestObj["method"].(string)
	if !ok {
		method = "GET"
	}

	url, ok := requestObj["url"].(string)
	if !ok {
		return nil, fmt.Errorf("request must have a 'url' field")
	}

	// Apply base_url if URL is relative
	if hc.BaseURL != "" && !isAbsoluteURL(url) {
		url = hc.BaseURL + url
	}

	body := ""
	if bodyVal, ok := requestObj["body"]; ok {
		body = fmt.Sprintf("%v", bodyVal)
	}

	// Build HTTP request
	var req *http.Request
	var err error
	if body != "" {
		req, err = http.NewRequest(method, url, io.NopCloser(nil))
		if err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}
		// Set body manually
		req.Body = io.NopCloser(&StringReader{s: body, offset: 0})
		req.ContentLength = int64(len(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}
	}

	// Add default headers
	for k, v := range hc.Headers {
		req.Header.Set(k, v)
	}

	// Add request-specific headers
	if headers, ok := requestObj["headers"]; ok {
		if headerMap, ok := headers.(map[string]any); ok {
			for k, v := range headerMap {
				req.Header.Set(k, fmt.Sprintf("%v", v))
			}
		}
	}

	// Auto-set Content-Type if body is provided and not already set
	if req.Body != nil && req.Header.Get("Content-Type") == "" {
		fmt.Fprintf(os.Stderr, "DEBUG: Setting Content-Type for POST with body\n")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Execute request
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Build response object
	responseHeaders := make(map[string]any)
	for k, vv := range resp.Header {
		if len(vv) == 1 {
			responseHeaders[k] = vv[0]
		} else {
			// For multiple values, return as array
			arr := make([]any, len(vv))
			for i, v := range vv {
				arr[i] = v
			}
			responseHeaders[k] = arr
		}
	}

	return map[string]any{
		"status":  float64(resp.StatusCode),
		"body":    string(respBody),
		"headers": responseHeaders,
	}, nil
}

// Close closes the HTTP client's idle connections (for cleanup).
func (hc *HTTPClientValue) Close() error {
	hc.client.CloseIdleConnections()
	return nil
}

// isAbsoluteURL checks if a URL is absolute (http:// or https://)
func isAbsoluteURL(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}

// StringReader is a simple reader for strings.
type StringReader struct {
	s      string
	offset int
}

func (sr *StringReader) Read(p []byte) (n int, err error) {
	if sr.offset >= len(sr.s) {
		return 0, io.EOF
	}
	n = copy(p, sr.s[sr.offset:])
	sr.offset += n
	return n, nil
}
