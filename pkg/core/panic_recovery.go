package core

import (
	"fmt"
	"os"
	"runtime/debug"
)

// RecoverPanic recovers from panics in goroutines and logs them safely.
// It prints the panic message, stack trace, and goroutine info to stderr.
// This should be called at the beginning of any goroutine function:
//
//	go func() {
//		defer RecoverPanic("goroutine description")
//		// actual work
//	}()
func RecoverPanic(context string) {
	if r := recover(); r != nil {
		// Format panic message
		var panicMsg string
		switch v := r.(type) {
		case string:
			panicMsg = v
		case error:
			panicMsg = v.Error()
		default:
			panicMsg = fmt.Sprintf("%v", v)
		}

		// Log panic to stderr with full context
		fmt.Fprintf(os.Stderr, "PANIC in %s: %s\n", context, panicMsg)
		fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
	}
}
