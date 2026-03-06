package runtime

import (
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// builtinKill terminates a spawned process by PID.
//
// kill(pid)
// - pid (number) - The process ID returned by spawn()
// - Returns true if the process was found and signaled
// - Returns error if the PID doesn't exist
//
// The spawned script will exit gracefully when it checks the cancellation signal
// in the next iteration of the execution loop.
//
// Example:
//
//	pid = spawn("worker.du")
//	sleep(5)
//	if kill(pid) then
//	  print("Process killed")
//	end
func builtinKill(evaluator *script.Evaluator, args map[string]any) (any, error) {
	// Get PID argument (positional or named)
	var pidArg any
	if p, ok := args["0"]; ok {
		pidArg = p
	} else if p, ok := args["pid"]; ok {
		pidArg = p
	} else {
		return nil, fmt.Errorf("kill() requires a pid argument")
	}

	// Convert to int64
	pidNum, ok := pidArg.(float64)
	if !ok {
		return nil, fmt.Errorf("kill() pid must be a number, got %T", pidArg)
	}

	pid := int64(pidNum)

	// Look up the cancel function
	procMutex.RLock()
	cancel, exists := spawnedProcs[pid]
	procMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("kill() no process with PID %d", pid)
	}

	// Signal the context cancellation
	cancel()

	return true, nil
}
