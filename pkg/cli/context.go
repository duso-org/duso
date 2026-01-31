package cli

import (
	"github.com/duso-org/duso/pkg/runtime"
	"github.com/duso-org/duso/pkg/script"
)

// NewContextFunction creates the context() builtin.
//
// context() returns an object with methods:
//   - .request() - Returns {method, path, headers, body, query}
//   - .callstack() - Returns array of invocation frames
//
// Returns nil if called outside a handler context (HTTP request, run(), or spawn()).
//
// Use exit() to return a value from handlers (exit() value becomes HTTP response or run() return value).
//
// Example:
//
//	ctx = context()
//	if ctx then
//	  req = ctx.request()
//	  exit({
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
		gid := runtime.GetGoroutineID()

		// Retrieve request context from goroutine-local storage
		ctx, ok := runtime.GetRequestContext(gid)
		if !ok {
			// Not in a request handler - return nil (like a hook that's not active)
			return nil, nil
		}

		// Create request() method
		requestFn := script.NewGoFunction(func(requestArgs map[string]any) (any, error) {
			return ctx.GetRequest(), nil
		})

		// Create callstack() method
		callstackFn := script.NewGoFunction(func(callstackArgs map[string]any) (any, error) {
			// Build array of frames by walking the chain
			frames := make([]any, 0)
			frame := ctx.Frame

			for frame != nil {
				frameObj := map[string]any{
					"filename": frame.Filename,
					"line":     float64(frame.Line),
					"col":      float64(frame.Col),
					"reason":   frame.Reason,
				}

				// Add details if present
				if frame.Details != nil {
					for k, v := range frame.Details {
						frameObj[k] = v
					}
				}

				frames = append(frames, frameObj)
				frame = frame.Parent
			}

			return frames, nil
		})

		// Return context object with request() and callstack() methods
		return map[string]any{
			"request":   requestFn,
			"callstack": callstackFn,
		}, nil
	}
}
