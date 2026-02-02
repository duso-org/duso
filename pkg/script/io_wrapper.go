package script

import (
	"io"
	"strings"
	"sync"
)

// StdoutWrapper captures output while writing to the real stdout.
// When capture is enabled, all writes go to both the real stdout and a capture buffer.
// This allows HTTP debug server to collect script output while still displaying it.
type StdoutWrapper struct {
	real          io.Writer
	captureBuffer *strings.Builder
	captureMutex  sync.Mutex
}

// NewStdoutWrapper creates a new stdout wrapper around the provided writer.
func NewStdoutWrapper(real io.Writer) *StdoutWrapper {
	return &StdoutWrapper{
		real: real,
	}
}

// Write writes to both the real stdout and capture buffer (if enabled).
func (sw *StdoutWrapper) Write(p []byte) (n int, err error) {
	// Write to real stdout
	realN, realErr := sw.real.Write(p)

	// Also capture if enabled
	sw.captureMutex.Lock()
	if sw.captureBuffer != nil {
		sw.captureBuffer.Write(p)
	}
	sw.captureMutex.Unlock()

	return realN, realErr
}

// EnableCapture starts capturing output to the provided buffer.
func (sw *StdoutWrapper) EnableCapture(buf *strings.Builder) {
	sw.captureMutex.Lock()
	defer sw.captureMutex.Unlock()
	sw.captureBuffer = buf
}

// DisableCapture stops capturing output.
func (sw *StdoutWrapper) DisableCapture() {
	sw.captureMutex.Lock()
	defer sw.captureMutex.Unlock()
	sw.captureBuffer = nil
}

// GetCapturedOutput returns the current captured output as a string.
func (sw *StdoutWrapper) GetCapturedOutput() string {
	sw.captureMutex.Lock()
	defer sw.captureMutex.Unlock()
	if sw.captureBuffer == nil {
		return ""
	}
	return sw.captureBuffer.String()
}

// StdinWrapper provides input from either HTTP requests or the real stdin.
// In HTTP debug mode, input is provided via a channel from HTTP requests.
// In console mode, input comes from the real stdin.
type StdinWrapper struct {
	real      io.Reader
	inputChan chan []byte
	readMutex sync.Mutex
}

// NewStdinWrapper creates a new stdin wrapper around the provided reader.
func NewStdinWrapper(real io.Reader) *StdinWrapper {
	return &StdinWrapper{
		real: real,
	}
}

// Read reads from either the input channel (HTTP mode) or real stdin (console mode).
// In HTTP mode, this blocks waiting for input from HTTP requests.
func (sw *StdinWrapper) Read(p []byte) (n int, err error) {
	sw.readMutex.Lock()
	inputChan := sw.inputChan
	sw.readMutex.Unlock()

	// If HTTP mode is enabled, wait for input from channel
	if inputChan != nil {
		data, ok := <-inputChan
		if !ok {
			// Channel closed
			return 0, io.EOF
		}
		// Copy received data into buffer
		n = copy(p, data)
		return n, nil
	}

	// Otherwise, read from real stdin
	return sw.real.Read(p)
}

// EnableHTTPMode activates HTTP input mode with the provided input channel.
func (sw *StdinWrapper) EnableHTTPMode(inputChan chan []byte) {
	sw.readMutex.Lock()
	defer sw.readMutex.Unlock()
	sw.inputChan = inputChan
}

// DisableHTTPMode deactivates HTTP input mode.
func (sw *StdinWrapper) DisableHTTPMode() {
	sw.readMutex.Lock()
	defer sw.readMutex.Unlock()
	sw.inputChan = nil
}

// ProvideInput sends input data to the waiting Read() call.
// Safe to call from HTTP handler goroutines.
func (sw *StdinWrapper) ProvideInput(data []byte) error {
	sw.readMutex.Lock()
	inputChan := sw.inputChan
	sw.readMutex.Unlock()

	if inputChan == nil {
		return nil // HTTP mode not enabled, ignore
	}

	select {
	case inputChan <- data:
		return nil
	default:
		// Channel buffer full or closed, silently drop
		return nil
	}
}
