package cli

import (
	"encoding/json"
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// NewFetchFunction creates the fetch(url, options) function.
//
// fetch() makes a single HTTP request following JavaScript's fetch API.
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
//     response = fetch("https://api.example.com/data")
//     if response.ok then
//       data = response.json()
//     end
//
//     response = fetch("https://api.example.com/submit", {
//       method = "POST",
//       headers = {["Content-Type"] = "application/json"},
//       body = format_json({name = "Alice"})
//     })
func NewFetchFunction() func(map[string]any) (any, error) {
	// Create a reusable HTTP client for fetch calls
	// This allows connection pooling across multiple fetch() calls
	client, err := script.NewHTTPClient(make(map[string]any))
	if err != nil {
		// If we can't create a client, return an error function
		return func(args map[string]any) (any, error) {
			return nil, fmt.Errorf("failed to initialize fetch: %w", err)
		}
	}

	return func(args map[string]any) (any, error) {
		// Get URL from first positional or named argument
		var url string

		if u, ok := args["0"]; ok {
			// Positional argument
			url = fmt.Sprintf("%v", u)
		} else if u, ok := args["url"]; ok {
			// Named argument
			url = fmt.Sprintf("%v", u)
		} else {
			return nil, fmt.Errorf("fetch() requires a URL")
		}

		// Get options from second positional or named argument
		var options map[string]any

		if opts, ok := args["1"]; ok {
			// Positional argument
			if optsMap, ok := opts.(map[string]any); ok {
				options = optsMap
			} else {
				options = make(map[string]any)
			}
		} else if opts, ok := args["options"]; ok {
			// Named argument
			if optsMap, ok := opts.(map[string]any); ok {
				options = optsMap
			} else {
				options = make(map[string]any)
			}
		} else {
			options = make(map[string]any)
		}

		// Build request object with defaults
		request := map[string]any{
			"method": "GET",
			"url":    url,
		}

		// Apply options to request
		if method, ok := options["method"]; ok && method != nil {
			request["method"] = fmt.Sprintf("%v", method)
		}

		if headers, ok := options["headers"]; ok && headers != nil {
			request["headers"] = headers
		}

		if body, ok := options["body"]; ok && body != nil {
			request["body"] = fmt.Sprintf("%v", body)
		}

		// Handle timeout - convert seconds to client timeout
		if timeout, ok := options["timeout"]; ok && timeout != nil {
			// Create a new client with the timeout for this request
			timeoutConfig := map[string]any{"timeout": timeout}
			reqClient, err := script.NewHTTPClient(timeoutConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create HTTP client with timeout: %w", err)
			}
			defer reqClient.Close()
			return buildFetchResponse(reqClient.Send(request))
		}

		// Send request
		responseData, err := client.Send(request)
		if err != nil {
			return nil, fmt.Errorf("fetch failed: %w", err)
		}

		return buildFetchResponse(responseData, nil)
	}
}

// buildFetchResponse wraps a raw HTTP response into a response object with methods
func buildFetchResponse(responseData map[string]any, err error) (any, error) {
	if err != nil {
		return nil, err
	}

	// Extract response fields
	status, _ := responseData["status"].(float64)
	body, _ := responseData["body"].(string)
	headers, _ := responseData["headers"].(map[string]any)

	// Create json() method
	jsonFn := script.NewGoFunction(func(args map[string]any) (any, error) {
		var result any
		if err := json.Unmarshal([]byte(body), &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		return result, nil
	})

	// Create text() method
	textFn := script.NewGoFunction(func(args map[string]any) (any, error) {
		return body, nil
	})

	// Determine .ok property (status < 400)
	ok := status < 400

	// Return response object with properties and methods
	return map[string]any{
		"status":  status,
		"ok":      ok,
		"body":    body,
		"headers": headers,
		"json":    jsonFn,
		"text":    textFn,
	}, nil
}
