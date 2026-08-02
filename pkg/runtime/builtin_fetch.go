package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// builtinFetch makes a single HTTP request following JavaScript's fetch API.
// Returns a response object with properties and methods:
//   - .status (number) - HTTP status code
//   - .ok (boolean) - true if status < 400
//   - .body (string) - Response body as string
//   - .headers (object) - Response headers
//   - .json() - Method to parse body as JSON
//   - .text() - Method to get body as string (same as .body)
//
// Options:
//   - method (string) - HTTP method, default "GET"
//   - headers (object) - Request headers
//   - body (string) - Request body
//   - timeout (number) - Request timeout in seconds
//
// Example:
//
//	response = fetch("https://api.example.com/data")
//	if response.ok then
//	  data = response.json()
//	end
//
//	response = fetch("https://api.example.com/submit", {
//	  method = "POST",
//	  headers = {["Content-Type"] = "application/json"},
//	  body = format_json({name = "Alice"})
//	})
func builtinFetch(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get URL from first positional or named argument
	var url string

	if u, ok := args["0"]; ok {
		url = fmt.Sprintf("%v", u)
	} else if u, ok := args["url"]; ok {
		url = fmt.Sprintf("%v", u)
	} else {
		return nil, fmt.Errorf("fetch() requires a URL")
	}

	// Get options from second positional or named argument
	var options map[string]any

	if opts, ok := args["1"]; ok {
		if optsMap, ok := opts.(map[string]any); ok {
			options = optsMap
		} else {
			options = make(map[string]any)
		}
	} else if opts, ok := args["options"]; ok {
		if optsMap, ok := opts.(map[string]any); ok {
			options = optsMap
		} else {
			options = make(map[string]any)
		}
	} else {
		options = make(map[string]any)
	}

	// Get HTTP method (default: GET)
	method := "GET"
	if m, ok := options["method"]; ok && m != nil {
		method = fmt.Sprintf("%v", m)
	}

	// Get request body
	var body io.Reader
	if b, ok := options["body"]; ok && b != nil {
		body = strings.NewReader(fmt.Sprintf("%v", b))
	}

	// Derive request context from the spawned process's cancellation context (if any),
	// so kill(pid) aborts an in-flight fetch() immediately instead of leaving it running.
	var procCtx context.Context = context.Background()
	if reqCtx, ok := script.CurrentRequestContext(evaluator); ok && reqCtx.ProcessCtx != nil {
		procCtx = reqCtx.ProcessCtx
	}

	// Optional script-supplied timeout: wrap procCtx so cancelling procCtx (kill())
	// makes the derived timeout fire immediately too.
	ctx := procCtx
	if timeout, ok := options["timeout"]; ok && timeout != nil {
		var timeoutSecs float64
		switch v := timeout.(type) {
		case float64:
			timeoutSecs = v
		case int:
			timeoutSecs = float64(v)
		}
		if timeoutSecs > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(procCtx, time.Duration(timeoutSecs*1000)*time.Millisecond)
			defer cancel()
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("fetch() invalid request: %w", err)
	}

	// Add request headers
	if headers, ok := options["headers"]; ok {
		if headerMap, ok := headers.(map[string]any); ok {
			for k, v := range headerMap {
				req.Header.Set(k, fmt.Sprintf("%v", v))
			}
		}
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch() failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fetch() failed to read response: %w", err)
	}

	// Build response headers map
	responseHeaders := make(map[string]any)
	for k, vv := range resp.Header {
		if len(vv) == 1 {
			responseHeaders[k] = vv[0]
		} else {
			arr := make([]any, len(vv))
			for i, v := range vv {
				arr[i] = v
			}
			responseHeaders[k] = arr
		}
	}

	// Return response object via buildFetchResponse
	return buildFetchResponse(map[string]any{
		"status":  float64(resp.StatusCode),
		"proto":   resp.Proto,
		"body":    string(respBody),
		"headers": responseHeaders,
	}, nil)
}

// buildFetchResponse wraps a raw HTTP response into a response object with methods
func buildFetchResponse(responseData map[string]any, err error) (any, error) {
	if err != nil {
		return nil, err
	}

	// Extract response fields
	status, _ := responseData["status"].(float64)
	proto, _ := responseData["proto"].(string)
	body, _ := responseData["body"].(string)
	headers, _ := responseData["headers"].(map[string]any)

	// Create json() method
	jsonFn := NewGoFunction(func(evaluator *Evaluator, args map[string]any) (any, error) {
		result, err := decodeJSON([]byte(body))
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		return result, nil
	})

	// Create text() method
	textFn := NewGoFunction(func(evaluator *Evaluator, args map[string]any) (any, error) {
		return body, nil
	})

	// Determine .ok property (status < 400)
	ok := status < 400

	// Return response object with properties and methods
	return map[string]any{
		"status":  status,
		"proto":   proto,
		"ok":      ok,
		"body":    body,
		"headers": headers,
		"json":    jsonFn,
		"text":    textFn,
	}, nil
}
