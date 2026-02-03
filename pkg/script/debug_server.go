package script

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// DebugServer provides HTTP access to debug a running duso script.
// It allows external clients (tests, Claude Code) to:
// - Read current debug state (breakpoint location, variables, output)
// - Execute debug commands (eval, continue, etc.)
// - Provide input to script's input() calls
type DebugServer struct {
	interp        *Interpreter
	port          int
	bind          string
	server        *http.Server
	mu            sync.RWMutex
	currentEvent  *DebugEvent
	outputBuffer  strings.Builder
	inputChan     chan []byte
	stdinWrapper  *StdinWrapper
	stdoutWrapper *StdoutWrapper
}

// NewDebugServer creates a new debug HTTP server.
func NewDebugServer(interp *Interpreter, port int, bind string,
	stdin *StdinWrapper, stdout *StdoutWrapper) *DebugServer {
	return &DebugServer{
		interp:        interp,
		port:          port,
		bind:          bind,
		inputChan:     make(chan []byte, 1),
		stdinWrapper:  stdin,
		stdoutWrapper: stdout,
	}
}

// Start begins listening on the debug port.
func (ds *DebugServer) Start() error {
	addr := net.JoinHostPort(ds.bind, strconv.Itoa(ds.port))

	mux := http.NewServeMux()
	mux.HandleFunc("/", ds.handleRequest)

	ds.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return ds.server.ListenAndServe()
}

// Stop shuts down the debug server.
func (ds *DebugServer) Stop() error {
	if ds.server != nil {
		return ds.server.Close()
	}
	return nil
}

// SetEvent stores the current debug event for HTTP clients to read.
func (ds *DebugServer) SetEvent(event *DebugEvent) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.currentEvent = event
}

// GetInputChannel returns the channel for providing input to the script.
func (ds *DebugServer) GetInputChannel() chan []byte {
	return ds.inputChan
}

// GetOutputBuffer returns the output buffer for capturing script output.
func (ds *DebugServer) GetOutputBuffer() *strings.Builder {
	return &ds.outputBuffer
}

// handleRequest processes GET and POST requests to the debug server.
func (ds *DebugServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch r.Method {
	case "GET":
		ds.handleGet(w)
	case "POST":
		ds.handlePost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed\n")
	}
}

// handleGet returns the current debug state and captured output.
func (ds *DebugServer) handleGet(w http.ResponseWriter) {
	ds.mu.RLock()
	event := ds.currentEvent
	ds.mu.RUnlock()

	// Start with state summary
	var state strings.Builder

	if event == nil {
		state.WriteString("running\n")
	} else {
		// Format location
		loc := event.FilePath
		if event.Position.Line > 0 {
			loc = fmt.Sprintf("%s:%d", loc, event.Position.Line)
			if event.Position.Column > 0 {
				loc = fmt.Sprintf("%s:%d", loc, event.Position.Column)
			}
		}

		state.WriteString(fmt.Sprintf("paused at %s\n", loc))

		// Error message if present
		if event.Message != "" {
			state.WriteString(fmt.Sprintf("error: %s\n", event.Message))
		}

		// Call stack
		if len(event.CallStack) > 0 {
			state.WriteString("\nCall stack:\n")
			for i := len(event.CallStack) - 1; i >= 0; i-- {
				frame := event.CallStack[i]
				state.WriteString(fmt.Sprintf("  at %s", frame.FunctionName))
				if frame.FilePath != "" {
					state.WriteString(fmt.Sprintf(" (%s:%d", frame.FilePath, frame.Position.Line))
					if frame.Position.Column > 0 {
						state.WriteString(fmt.Sprintf(":%d", frame.Position.Column))
					}
					state.WriteString(")")
				}
				state.WriteString("\n")
			}
		}

		// Local variables (from environment variables map)
		if event.Env != nil {
			state.WriteString("\nLocal variables:\n")
			// Iterate through environment variables (from its variables map)
			// Since Environment doesn't export GetAll, we'll show minimal info
			state.WriteString("  (debug scope variables available)\n")
		}
	}

	// Captured output
	output := ds.stdoutWrapper.GetCapturedOutput()
	if output != "" {
		state.WriteString("\nOutput:\n")
		state.WriteString(output)
	}

	fmt.Fprint(w, state.String())
}

// handlePost processes debug commands from the HTTP client.
// Commands can be:
// - "c" or "continue": Resume execution
// - Script code: Evaluate in the breakpoint's scope (like the console REPL)
func (ds *DebugServer) handlePost(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	defer r.Body.Close()
	body := make([]byte, 4096)
	n, err := r.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error reading request: %v\n", err)
		return
	}

	data := body[:n]
	command := strings.TrimSpace(string(data))

	// Get current debug event
	ds.mu.RLock()
	event := ds.currentEvent
	ds.mu.RUnlock()

	// Handle continue command
	if command == "c" || command == "continue" {
		// Signal resume
		if event != nil && event.ResumeChan != nil {
			select {
			case event.ResumeChan <- true:
			default:
			}
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK\n")
		return
	}

	// Evaluate script command in breakpoint scope
	if command != "" && event != nil && event.Env != nil {
		// Capture output before eval
		prevOutput := ds.stdoutWrapper.GetCapturedOutput()

		result, evalErr := ds.interp.EvalInEnvironment(command, event.Env)
		if evalErr != nil {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Error: %v\n", evalErr)
		} else {
			// Return the output generated during eval
			w.WriteHeader(http.StatusOK)
			newOutput := ds.stdoutWrapper.GetCapturedOutput()
			if len(newOutput) > len(prevOutput) {
				// Only return the new output generated
				fmt.Fprint(w, newOutput[len(prevOutput):])
			} else {
				// If no new output, return the result (though it's usually empty)
				fmt.Fprint(w, result)
			}
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK\n")
}

// formatValue converts a Value to a readable string for display.
func formatValue(val Value) string {
	if val.IsNil() {
		return "nil"
	}

	switch v := val.Data.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case []any:
		// Simple array formatting
		parts := make([]string, len(v))
		for i, item := range v {
			if itemVal, ok := item.(Value); ok {
				parts[i] = formatValue(itemVal)
			} else {
				parts[i] = fmt.Sprintf("%v", item)
			}
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]any:
		// Simple object formatting
		var pairs []string
		for k, val := range v {
			if itemVal, ok := val.(Value); ok {
				pairs = append(pairs, fmt.Sprintf("%s: %s", k, formatValue(itemVal)))
			} else {
				pairs = append(pairs, fmt.Sprintf("%s: %v", k, val))
			}
		}
		return "{" + strings.Join(pairs, ", ") + "}"
	default:
		return fmt.Sprintf("%v", val.Data)
	}
}
