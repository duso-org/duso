package cli

import (
	"reflect"

	"github.com/duso-org/duso/pkg/runtime"
	"github.com/duso-org/duso/pkg/script"
)

// NewContextFunction creates the context() builtin.
//
// context() returns an object with methods:
//   - .request() - Returns {method, path, headers, body, query, params, form}
//   - .response() - Returns response helpers (json, text, html, error, redirect, file)
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
//	  res = ctx.response()
//	  res.json({hello = "world"})
//	else
//	  // Not in request handler - start server
//	  server = http_server({port = 8080})
//	  server.route("GET", "/", "myscript.du")
//	  server.start()
//	end
func NewContextFunction() func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		// Get current goroutine ID
		gid := runtime.GetGoroutineID()

		// Retrieve context getter from goroutine-local storage
		getter, ok := runtime.GetContextGetter(gid)
		if !ok {
			// No context getter registered - return nil
			return nil, nil
		}

		// Call the getter to retrieve the actual context (lazy evaluation)
		ctxAny := getter()
		if ctxAny == nil {
			// Getter returned nil - return nil
			return nil, nil
		}

		// The context can be either script.RequestContext or runtime.RequestContext
		// Try to cast to runtime.RequestContext first (for HTTP handlers)
		if ctx, ok := ctxAny.(*runtime.RequestContext); ok {
			return buildContextObject(ctx), nil
		}

		// Try to cast to script.RequestContext (for run/spawn)
		// Since we can't directly compare package types, use reflection
		ctxVal := reflect.ValueOf(ctxAny)
		if ctxVal.Kind() == reflect.Ptr && ctxVal.Elem().Kind() == reflect.Struct {
			// It's a pointer to a struct - call GetRequest and GetResponse via reflection
			return buildContextObjectFromAny(ctxAny), nil
		}

		return nil, nil
	}
}

// buildContextObject builds the context object from a runtime.RequestContext
func buildContextObject(ctx *runtime.RequestContext) map[string]any {
	// Create request() method
	requestFn := script.NewGoFunction(func(evaluator *script.Evaluator, requestArgs map[string]any) (any, error) {
		return ctx.GetRequest(), nil
	})

	// Create response() method that returns response helpers
	responseFn := script.NewGoFunction(func(evaluator *script.Evaluator, responseArgs map[string]any) (any, error) {
		return ctx.GetResponse(), nil
	})

	// Create callstack() method
	callstackFn := script.NewGoFunction(func(evaluator *script.Evaluator, callstackArgs map[string]any) (any, error) {
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

	// Return context object with request(), response(), and callstack() methods
	return map[string]any{
		"request":   requestFn,
		"response":  responseFn,
		"callstack": callstackFn,
	}
}

// buildContextObjectFromAny builds the context object from any RequestContext-like object
// This handles both script.RequestContext and runtime.RequestContext via reflection
func buildContextObjectFromAny(ctxAny any) map[string]any {
	ctxVal := reflect.ValueOf(ctxAny)

	// Create request() method via reflection
	requestFn := script.NewGoFunction(func(evaluator *script.Evaluator, requestArgs map[string]any) (any, error) {
		getRequestMethod := ctxVal.MethodByName("GetRequest")
		if !getRequestMethod.IsValid() {
			return nil, nil
		}
		result := getRequestMethod.Call([]reflect.Value{})
		if len(result) > 0 {
			return result[0].Interface(), nil
		}
		return nil, nil
	})

	// Create response() method via reflection
	responseFn := script.NewGoFunction(func(evaluator *script.Evaluator, responseArgs map[string]any) (any, error) {
		getResponseMethod := ctxVal.MethodByName("GetResponse")
		if !getResponseMethod.IsValid() {
			return nil, nil
		}
		result := getResponseMethod.Call([]reflect.Value{})
		if len(result) > 0 {
			return result[0].Interface(), nil
		}
		return nil, nil
	})

	// Create callstack() method via reflection
	callstackFn := script.NewGoFunction(func(evaluator *script.Evaluator, callstackArgs map[string]any) (any, error) {
		frameField := ctxVal.Elem().FieldByName("Frame")
		if !frameField.IsValid() {
			return []any{}, nil
		}

		// Build array of frames by walking the chain
		frames := make([]any, 0)
		framePtr := frameField.Interface().(*script.InvocationFrame)

		for framePtr != nil {
			frameObj := map[string]any{
				"filename": framePtr.Filename,
				"line":     float64(framePtr.Line),
				"col":      float64(framePtr.Col),
				"reason":   framePtr.Reason,
			}

			// Add details if present
			if framePtr.Details != nil {
				for k, v := range framePtr.Details {
					frameObj[k] = v
				}
			}

			frames = append(frames, frameObj)
			framePtr = framePtr.Parent
		}

		return frames, nil
	})

	// Return context object with request(), response(), and callstack() methods
	return map[string]any{
		"request":   requestFn,
		"response":  responseFn,
		"callstack": callstackFn,
	}
}
