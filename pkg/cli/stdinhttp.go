// Package cli provides CLI-specific functions for Duso scripts.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// StdinHTTPServer provides HTTP access to a script's stdin/stdout.
// It allows remote clients (LLMs, tests, etc.) to:
// - Read accumulated script output (GET /)
// - Wait for and provide input to input() calls (GET /input, POST /input)
//
// This is a generic stdin/stdout transport that works for all scripts,
// not just debug mode. When -stdin-port is specified, the script's
// stdin/stdout is automatically exposed over HTTP.
type StdinHTTPServer struct {
	port          int
	bind          string
	server        *http.Server
	mu            sync.RWMutex
	outputBuffer  strings.Builder
	inputChan     chan []byte       // Channel for receiving input from POST /input
	inputWaitChan chan struct{}     // Signal that input() was called and we're waiting
	stdinReader   *bufio.Reader      // For reading input() lines from the input channel
	stdoutCapture *strings.Builder   // Buffer for capturing stdout output
	stdoutLock    sync.RWMutex
}

// NewStdinHTTPServer creates a new HTTP stdin/stdout server.
func NewStdinHTTPServer(port int, bind string) *StdinHTTPServer {
	return &StdinHTTPServer{
		port:          port,
		bind:          bind,
		inputChan:     make(chan []byte, 1),
		inputWaitChan: make(chan struct{}, 1),
		stdoutCapture: &strings.Builder{},
	}
}

// Start begins listening on the stdin/stdout port.
// This method blocks until the server is closed.
func (s *StdinHTTPServer) Start() error {
	addr := net.JoinHostPort(s.bind, strconv.Itoa(s.port))

	s.server = &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(s.routeRequest),
	}

	return s.server.ListenAndServe()
}

// routeRequest dispatches requests to the appropriate handler based on path and method
func (s *StdinHTTPServer) routeRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" && r.Method == "GET" {
		s.handleGetOutput(w, r)
	} else if r.URL.Path == "/input" && r.Method == "GET" {
		s.handleGetInput(w, r)
	} else if r.URL.Path == "/input" && r.Method == "POST" {
		s.handlePostInput(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "not found")
	}
}

// Stop shuts down the HTTP server.
func (s *StdinHTTPServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// GetOutputWriter returns a function compatible with interp.OutputWriter.
// It captures output to the HTTP buffer and also writes to stdout (tee pattern).
// This ensures output is visible in the terminal and can be captured with shell redirection.
func (s *StdinHTTPServer) GetOutputWriter() func(string) error {
	return func(msg string) error {
		// Write to stdout so output is visible in the terminal and can be redirected
		fmt.Println(msg)

		// Also capture to HTTP buffer for remote clients
		s.mu.Lock()
		s.outputBuffer.WriteString(msg)
		s.outputBuffer.WriteString("\n")
		s.mu.Unlock()

		return nil
	}
}

// GetInputReader returns a function compatible with interp.InputReader.
// It reads from the HTTP input channel and blocks waiting for POST /input.
func (s *StdinHTTPServer) GetInputReader() func(string) (string, error) {
	return func(prompt string) (string, error) {
		// Write prompt to stdout and buffer so it appears in terminal and HTTP
		fmt.Fprint(os.Stdout, prompt)

		// Also capture to HTTP buffer for remote clients
		s.mu.Lock()
		s.outputBuffer.WriteString(prompt)
		s.mu.Unlock()

		// Signal that we're waiting for input
		select {
		case s.inputWaitChan <- struct{}{}:
		default:
		}

		// Block until input is provided via POST /input
		data, ok := <-s.inputChan
		if !ok {
			return "", io.EOF
		}

		// Convert bytes to string and remove trailing newline if present
		line := string(data)
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		return line, nil
	}
}

// handleGetOutput handles GET / requests.
// Returns the accumulated stdout output.
func (s *StdinHTTPServer) handleGetOutput(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	s.mu.RLock()
	output := s.outputBuffer.String()
	s.mu.RUnlock()

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, output)
}

// handleGetInput handles GET /input requests.
// Blocks waiting for input() to be called, then returns the current output
// to give the caller context about what input is needed.
func (s *StdinHTTPServer) handleGetInput(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Block until input() is called
	<-s.inputWaitChan

	// Return current output buffer so caller sees the prompt/context
	s.mu.RLock()
	output := s.outputBuffer.String()
	s.mu.RUnlock()

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, output)
}

// handlePostInput handles POST /input requests.
// Receives input data and sends it to the waiting input() call.
func (s *StdinHTTPServer) handlePostInput(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Read input from request body
	data := make([]byte, 8192) // Max input line size
	n, err := r.Body.Read(data)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error reading input: %v", err)
		return
	}

	// Send input to the waiting GetInputReader()
	select {
	case s.inputChan <- data[:n]:
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "not waiting for input")
	}
}
