package cli

import (
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// NewHTTPClientFunction creates the http_client(config) function.
//
// http_client() returns a stateful HTTP client object with methods:
//   - .send(request) - Execute HTTP request, returns {status, body, headers}
//   - .close() - Close idle connections
//
// Configuration options:
//   - base_url (string) - Base URL for relative requests
//   - timeout (number) - Timeout in seconds
//   - headers (object) - Default headers for all requests
//
// Example:
//     client = http_client({
//         base_url = "https://api.example.com",
//         timeout = 30,
//         headers = {Authorization = "Bearer token123"}
//     })
//
//     response = client.send({
//         method = "GET",
//         url = "/data",
//         headers = {Accept = "application/json"}
//     })
//
//     print(response.status)   // 200
//     print(response.body)     // JSON response body
func NewHTTPClientFunction() func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		// Get config from first positional or named argument
		var config map[string]any

		if cfg, ok := args["0"]; ok {
			// Positional argument
			if cfgMap, ok := cfg.(map[string]any); ok {
				config = cfgMap
			} else {
				return nil, fmt.Errorf("http_client() argument must be a config object")
			}
		} else if cfg, ok := args["config"]; ok {
			// Named argument
			if cfgMap, ok := cfg.(map[string]any); ok {
				config = cfgMap
			} else {
				return nil, fmt.Errorf("http_client() 'config' argument must be a config object")
			}
		} else {
			// Empty config
			config = make(map[string]any)
		}

		// Create HTTP client
		client, err := script.NewHTTPClient(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP client: %w", err)
		}

		// Create send function wrapped as a Duso value
		sendFn := script.NewGoFunction(func(sendArgs map[string]any) (any, error) {
			// Get request object
			var request map[string]any

			if req, ok := sendArgs["0"]; ok {
				if reqMap, ok := req.(map[string]any); ok {
					request = reqMap
				} else {
					return nil, fmt.Errorf("send() argument must be a request object")
				}
			} else if req, ok := sendArgs["request"]; ok {
				if reqMap, ok := req.(map[string]any); ok {
					request = reqMap
				} else {
					return nil, fmt.Errorf("send() 'request' argument must be a request object")
				}
			} else {
				return nil, fmt.Errorf("send() requires a request object")
			}

			return client.Send(request)
		})

		// Create close function wrapped as a Duso value
		closeFn := script.NewGoFunction(func(closeArgs map[string]any) (any, error) {
			return nil, client.Close()
		})

		// Return as Duso object with send() and close() methods
		return map[string]any{
			"send":  sendFn,
			"close": closeFn,
		}, nil
	}
}
