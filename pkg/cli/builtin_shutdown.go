package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/runtime"
	"github.com/duso-org/duso/pkg/script"
)

// shutdown() ends the process, cleanly.
//
// exit() ends the current script instance and hands back a value - inside an
// HTTP handler it ends only that handler, which is why a script parked in
// server.start() could not previously be stopped from anywhere but a signal.
// shutdown() ends everything: in-flight requests drain, WebSockets close,
// datastores flush, then the process exits.
//
// It lives in the CLI rather than the runtime because ending the process is the
// CLI's business. The runtime knows how to drain and flush; it has no business
// deciding a Go program is over.

var (
	exitCodeMutex sync.Mutex
	exitCode      int
	exitRequested bool
)

// builtinShutdown stops the process with an optional exit code (default 0).
//
// Usage: shutdown() or shutdown(1)
func builtinShutdown(evaluator *script.Evaluator, args map[string]any) (any, error) {
	code := 0
	if raw, ok := args["0"]; ok {
		c, err := shutdownCode(raw)
		if err != nil {
			return nil, err
		}
		code = c
	} else if raw, ok := args["code"]; ok {
		c, err := shutdownCode(raw)
		if err != nil {
			return nil, err
		}
		code = c
	}

	requestShutdown(code)

	// End this script instance the way exit() does, so nothing after
	// shutdown() runs while the process is on its way down.
	return nil, &script.ExitExecution{Values: make([]any, 0)}
}

// shutdownCode validates the exit code. Duso numbers are floats, so a
// fractional or out-of-range code is a mistake worth naming rather than
// silently truncating into something a service manager will misread.
func shutdownCode(raw any) (int, error) {
	n, ok := raw.(float64)
	if !ok {
		return 0, fmt.Errorf("shutdown() exit code must be a number")
	}
	if n != float64(int(n)) || n < 0 || n > 255 {
		return 0, fmt.Errorf("shutdown() exit code must be a whole number from 0 to 255")
	}
	return int(n), nil
}

// requestShutdown records the exit code and runs the shutdown on its own
// goroutine.
//
// The goroutine is not an optimization. shutdown() is usually called from an
// HTTP handler, and draining waits for in-flight connections to go idle - a
// handler blocked inside the drain is a connection that never does, so blocking
// here would stall every shutdown for the full drain timeout. Backgrounding it
// lets the calling handler unwind, which is what allows the drain to finish.
func requestShutdown(code int) {
	exitCodeMutex.Lock()
	if !exitRequested {
		// First caller wins: a second shutdown(2) racing a shutdown(0) does not
		// get to change the answer half way through.
		exitRequested = true
		exitCode = code
	}
	exitCodeMutex.Unlock()

	go func() {
		defer core.RecoverPanic("shutdown")
		runtime.GracefulShutdown()
		os.Exit(ExitCode(0))
	}()
}

// ExitCode returns the code shutdown() asked for, or fallback if it was never
// called. The CLI's own exit paths route through this, so a shutdown() from a
// handler still decides the process's exit status even though the main script
// finished by other means.
func ExitCode(fallback int) int {
	exitCodeMutex.Lock()
	defer exitCodeMutex.Unlock()
	if exitRequested {
		return exitCode
	}
	return fallback
}
