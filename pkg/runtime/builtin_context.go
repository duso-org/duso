package runtime

// builtinContext returns the data object passed to spawn() or run() by the caller,
// or nil if called outside a spawned/run context.
//
// context() returns the data object passed to spawn() or run() by the caller,
// or nil if called outside a spawned/run context.
//
// For HTTP handlers, the HTTP server should pass request() and response() functions
// as part of the data object passed to the handler script.
//
// Example:
//
//	ctx = context()
//	if ctx then
//	  // ctx is the data object passed by spawn() or run()
//	  print(ctx.my_data)
//	else
//	  // Not in a spawned/run context
//	  print("Running standalone")
//	end
func builtinContext(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get current goroutine ID
	gid := GetGoroutineID()

	// Retrieve context getter from goroutine-local storage
	getter, ok := GetContextGetter(gid)
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

	// For spawn/run: context getter returns the data object directly
	// For HTTP handlers: context getter should return the data object (request/response functions passed by HTTP server)
	// In all cases, just return what the getter gives us
	return ctxAny, nil
}
