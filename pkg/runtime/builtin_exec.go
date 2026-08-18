package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// defaultMaxExecOutput caps each captured stream. Past this, output is dropped
// and .truncated is set - a command that prints a gigabyte shouldn't be able to
// take the process down with it.
const defaultMaxExecOutput = 1 << 20 // 1 MB

// execTermGrace is how long a timed-out child gets between SIGTERM and SIGKILL.
// It runs in the background, so it never delays the script.
const execTermGrace = 2 * time.Second

// builtinExec runs an external command and returns a result object.
// Mirrors fetch(): the transport-level failure throws, the unhappy-but-normal
// outcome comes back as an object with .ok on it.
//
// Result:
//   - .ok (boolean) - the command exited 0
//   - .code (number) - exit status
//   - .stdout, .stderr (string) - captured output, capped by max_output
//   - .truncated (boolean) - either stream hit the cap
//
// Options:
//   - args (array) - arguments, passed to the command directly
//   - env (object) - environment variables for the child
//   - inherit_env (boolean) - pass duso's whole environment through, default false
//   - dir (string) - working directory
//   - timeout (number) - seconds; no timeout by default, same as fetch()
//   - max_output (number) - bytes per stream, default 1 MB
//   - stdin (string) - written to the child's stdin
//
// A shell is never involved. Arguments go to the process as given, so a value
// interpolated from a request can't turn into a second command - the only way
// to reach a shell is to ask for one by name, and only if the operator allowed
// it: exec("sh", {args = ["-c", "..."]}).
//
// Requires -allow-exec at startup. See exec_allow.go.
//
// Example:
//
//	result = exec("systemctl", {args = ["restart", "arland"]})
//	if not result.ok then
//	  print("restart failed: " + result.stderr)
//	end
func builtinExec(evaluator *Evaluator, args map[string]any) (any, error) {
	// Command from first positional or named argument
	var command string
	if c, ok := args["0"]; ok {
		command = fmt.Sprintf("%v", c)
	} else if c, ok := args["command"]; ok {
		command = fmt.Sprintf("%v", c)
	} else {
		return nil, fmt.Errorf("exec() requires a command")
	}

	// Options from second positional or named argument
	options := map[string]any{}
	if opts, ok := args["1"]; ok {
		if optsMap, ok := opts.(map[string]any); ok {
			options = optsMap
		}
	} else if opts, ok := args["options"]; ok {
		if optsMap, ok := opts.(map[string]any); ok {
			options = optsMap
		}
	}

	argv, err := execArgv(options["args"])
	if err != nil {
		return nil, err
	}

	// Permission check happens before anything else touches the system, and
	// gives us the absolute path resolved back at startup.
	path, err := resolveExec(command, argv)
	if err != nil {
		return nil, err
	}

	// Derive from the spawned process's context (same as fetch) so kill(pid)
	// tears down an in-flight child instead of leaving it running.
	var procCtx context.Context = context.Background()
	if reqCtx, ok := script.CurrentRequestContext(evaluator); ok && reqCtx.ProcessCtx != nil {
		procCtx = reqCtx.ProcessCtx
	}
	ctx := procCtx
	if timeoutSecs := execNumber(options["timeout"], 0); timeoutSecs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(procCtx, time.Duration(timeoutSecs*float64(time.Second)))
		defer cancel()
	}

	maxOutput := int(execNumber(options["max_output"], defaultMaxExecOutput))
	if maxOutput < 0 {
		maxOutput = 0
	}

	cmd := exec.Command(path, argv...)
	if dir, ok := options["dir"]; ok && dir != nil {
		cmd.Dir = fmt.Sprintf("%v", dir)
	}
	cmd.Env, err = execEnv(options)
	if err != nil {
		return nil, err
	}
	if in, ok := options["stdin"]; ok && in != nil {
		cmd.Stdin = strings.NewReader(fmt.Sprintf("%v", in))
	}
	// With Stdin left nil, Go hands the child /dev/null. That matters: a child
	// that reads stdin would otherwise sit on an open pipe forever, and with no
	// default timeout, forever means forever.

	stdout := &capWriter{max: maxOutput}
	stderr := &capWriter{max: maxOutput}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	setProcessGroup(cmd)

	// argv is logged on every call. The environment never is - that's where the
	// secrets are, and it's the reason inherit_env defaults to false.
	logExecCall(path, argv)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("exec() could not start %s: %w", command, err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		var exitErr *exec.ExitError
		switch {
		case err == nil:
		case errors.As(err, &exitErr):
			// A non-zero exit is a result, not a failure. The script decides.
		default:
			return nil, fmt.Errorf("exec() failed: %w", err)
		}
		return buildExecResult(cmd.ProcessState.ExitCode(), stdout, stderr), nil

	case <-ctx.Done():
		// Hand control back to the script now. Cleaning up the child is best
		// effort and happens behind us: a process ignoring SIGTERM must not be
		// able to hang the very call the timeout exists to unblock.
		go func() {
			signalProcessGroup(cmd, false)
			select {
			case <-done:
			case <-time.After(execTermGrace):
				signalProcessGroup(cmd, true)
				<-done
			}
		}()

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("exec() timed out: %s", command)
		}
		return nil, fmt.Errorf("exec() cancelled: %s", command)
	}
}

// buildExecResult assembles the result object once the child is done.
func buildExecResult(code int, stdout, stderr *capWriter) map[string]any {
	out, outTruncated := stdout.result()
	errOut, errTruncated := stderr.result()
	return map[string]any{
		"ok":        code == 0,
		"code":      float64(code),
		"stdout":    out,
		"stderr":    errOut,
		"truncated": outTruncated || errTruncated,
	}
}

// capWriter collects up to max bytes and discards the rest, so a child can keep
// writing without blocking on a full pipe. Locked because a timed-out child is
// still writing after the script has its control back.
type capWriter struct {
	mu        sync.Mutex
	buf       []byte
	max       int
	truncated bool
}

func (w *capWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if room := w.max - len(w.buf); room > 0 {
		if len(p) <= room {
			w.buf = append(w.buf, p...)
		} else {
			w.buf = append(w.buf, p[:room]...)
			w.truncated = true
		}
	} else if len(p) > 0 {
		w.truncated = true
	}
	return len(p), nil // always claim the full write; dropping isn't the child's problem
}

func (w *capWriter) result() (string, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return string(w.buf), w.truncated
}

// execArgv converts the args option into strings. Scalars convert the way
// tostring() converts them; anything structural is a mistake worth naming,
// since "[object]" as an argument would be a baffling thing to debug.
func execArgv(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}

	// An array arrives either as a duso array ref or as a plain slice,
	// depending on how it reached the builtin. Both are the same thing here.
	var list []any
	switch v := raw.(type) {
	case []any:
		list = v
	case *[]Value:
		list = make([]any, len(*v))
		for i, item := range *v {
			list[i] = ValueToInterface(item)
		}
	default:
		return nil, fmt.Errorf("exec() args must be an array")
	}

	argv := make([]string, 0, len(list))
	for i, item := range list {
		switch v := item.(type) {
		case string:
			argv = append(argv, v)
		case float64, bool:
			argv = append(argv, InterfaceToValue(v).String())
		default:
			return nil, fmt.Errorf("exec() args[%d] must be a string, number, or boolean", i)
		}
	}
	return argv, nil
}

// execEnv builds the child's environment.
//
// inherit_env is false by default because duso's own environment is where the
// database URL and the API keys live, and a command that doesn't need them
// shouldn't see them. But an empty environment breaks most real tools, and a
// default that breaks everything just gets turned off - so a small set of
// variables that say where the machine keeps things, and nothing about who it
// talks to, passes through.
func execEnv(options map[string]any) ([]string, error) {
	var env []string

	if inherit, ok := options["inherit_env"].(bool); ok && inherit {
		env = os.Environ()
	} else {
		for _, name := range []string{"PATH", "HOME", "TMPDIR", "LANG", "TZ"} {
			if val, ok := os.LookupEnv(name); ok {
				env = append(env, name+"="+val)
			}
		}
	}

	if raw, ok := options["env"]; ok && raw != nil {
		extra, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("exec() env must be an object")
		}
		for name, val := range extra {
			env = append(env, fmt.Sprintf("%s=%v", name, val))
		}
	}
	return env, nil
}

// execNumber reads a numeric option, falling back when it's absent or not a number.
func execNumber(raw any, fallback float64) float64 {
	switch v := raw.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return fallback
}

// logExecCall records what ran. Argv only - never the environment.
func logExecCall(path string, argv []string) {
	if len(argv) == 0 {
		fmt.Fprintf(os.Stderr, "exec: %s\n", path)
		return
	}
	fmt.Fprintf(os.Stderr, "exec: %s %s\n", path, strings.Join(argv, " "))
}
