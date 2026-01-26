package cli

import (
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// NewContextFunction creates the context() builtin for HTTP handler scripts.
//
// context() returns an object with methods:
//   - .request() - Returns {method, path, headers, body, query}
//   - .response(data) - Sends HTTP response, takes {status, headers, body}
//
// Returns nil if called outside an HTTP request handler (useful for self-referential scripts).
//
// Example:
//
//	ctx = context()
//	if ctx then
//	  req = ctx.request()
//	  ctx.response({
//	      status = 200,
//	      headers = {Content-Type = "application/json"},
//	      body = format_json(data)
//	  })
//	else
//	  // Not in request handler - start server
//	  server = http_server({port = 8080})
//	  server.route("GET", "/", "myscript.du")
//	  server.start()
//	end
func NewContextFunction() func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
		// Get current goroutine ID
		gid := script.GetGoroutineID()

		// Retrieve request context from goroutine-local storage
		ctx, ok := script.GetRequestContext(gid)
		if !ok {
			// Not in a request handler - return nil (like a hook that's not active)
			return nil, nil
		}

		// Create request() method
		requestFn := script.NewGoFunction(func(requestArgs map[string]any) (any, error) {
			return ctx.GetRequest(), nil
		})

		// Create response() method
		responseFn := script.NewGoFunction(func(responseArgs map[string]any) (any, error) {
			// Extract response data - can be positional arg "0" or named arg "data"
			var data map[string]any

			if d, ok := responseArgs["0"]; ok {
				if dataMap, ok := d.(map[string]any); ok {
					data = dataMap
				} else {
					return nil, fmt.Errorf("response() requires a response object")
				}
			} else if d, ok := responseArgs["data"]; ok {
				if dataMap, ok := d.(map[string]any); ok {
					data = dataMap
				} else {
					return nil, fmt.Errorf("response() 'data' argument must be a response object")
				}
			} else {
				return nil, fmt.Errorf("response() requires a response object")
			}

			// Send the response
			return nil, ctx.SendResponse(data)
		})

		// Return context object with request() and response() methods
		return map[string]any{
			"request":  requestFn,
			"response": responseFn,
		}, nil
	}
}
