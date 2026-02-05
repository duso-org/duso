package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Transport defines the interface for LSP transports
type Transport interface {
	Start(server *Server) error
	Stop() error
}

// StdioTransport implements LSP over stdin/stdout
type StdioTransport struct {
	server   *Server
	input    io.Reader
	output   io.Writer
	mu       sync.Mutex
	stopChan chan struct{}
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		input:    os.Stdin,
		output:   os.Stdout,
		stopChan: make(chan struct{}),
	}
}

// Start begins accepting connections over stdin
func (t *StdioTransport) Start(server *Server) error {
	t.server = server

	// Create a message reader
	reader := NewMessageReader(t.input)

	// Handle messages in a loop
	for {
		select {
		case <-t.stopChan:
			return nil
		default:
		}

		msg, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Handle message
		go t.handleMessage(msg)
	}
}

// Stop stops the transport
func (t *StdioTransport) Stop() error {
	close(t.stopChan)
	return nil
}

// handleMessage processes an LSP message
func (t *StdioTransport) handleMessage(msg *Message) {
	response := t.server.HandleMessage(msg)
	if response != nil {
		t.sendMessage(response)
	}
}

// sendMessage sends a message to the client
func (t *StdioTransport) sendMessage(msg *Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Encode the message
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Write headers
	fmt.Fprintf(t.output, "Content-Length: %d\r\n\r\n", len(data))

	// Write body
	_, err = t.output.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// TCPTransport implements LSP over TCP
type TCPTransport struct {
	server   *Server
	address  string
	listener net.Listener
	mu       sync.Mutex
	stopChan chan struct{}
}

// NewTCPTransport creates a new TCP transport
func NewTCPTransport(port string) *TCPTransport {
	return &TCPTransport{
		address:  "localhost:" + port,
		stopChan: make(chan struct{}),
	}
}

// Start begins listening for TCP connections
func (t *TCPTransport) Start(server *Server) error {
	t.server = server

	var err error
	t.listener, err = net.Listen("tcp", t.address)
	if err != nil {
		return err
	}

	fmt.Printf("LSP server listening on %s\n", t.address)

	// Accept connections
	for {
		select {
		case <-t.stopChan:
			return nil
		default:
		}

		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.stopChan:
				return nil
			default:
				return err
			}
		}

		// Handle connection
		go t.handleConnection(conn)
	}
}

// Stop stops the transport
func (t *TCPTransport) Stop() error {
	close(t.stopChan)
	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}

// handleConnection handles a single TCP connection
func (t *TCPTransport) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := NewMessageReader(conn)
	writer := NewMessageWriter(conn)

	for {
		select {
		case <-t.stopChan:
			return
		default:
		}

		msg, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading message: %v\n", err)
			}
			return
		}

		response := t.server.HandleMessage(msg)
		if response != nil {
			if err := writer.Write(response); err != nil {
				fmt.Printf("Error writing message: %v\n", err)
				return
			}
		}
	}
}

// MessageReader reads LSP messages
type MessageReader struct {
	reader *bufio.Reader
}

// NewMessageReader creates a new message reader
func NewMessageReader(r io.Reader) *MessageReader {
	return &MessageReader{
		reader: bufio.NewReader(r),
	}
}

// Read reads a single LSP message
func (mr *MessageReader) Read() (*Message, error) {
	// Read headers
	headers := make(map[string]string)

	for {
		line, err := mr.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			// Empty line marks end of headers
			break
		}

		// Parse header
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Get content length
	contentLenStr, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	contentLen, err := strconv.Atoi(contentLenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %w", err)
	}

	// Read content
	content := make([]byte, contentLen)
	_, err = io.ReadFull(mr.reader, content)
	if err != nil {
		return nil, err
	}

	// Parse message
	var msg Message
	if err := json.Unmarshal(content, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// MessageWriter writes LSP messages
type MessageWriter struct {
	writer *bufio.Writer
	mu     sync.Mutex
}

// NewMessageWriter creates a new message writer
func NewMessageWriter(w io.Writer) *MessageWriter {
	return &MessageWriter{
		writer: bufio.NewWriter(w),
	}
}

// Write writes a single LSP message
func (mw *MessageWriter) Write(msg *Message) error {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	// Encode message
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Write headers
	if _, err := fmt.Fprintf(mw.writer, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}

	// Write content
	if _, err := mw.writer.Write(data); err != nil {
		return err
	}

	// Flush
	return mw.writer.Flush()
}

// Message represents an LSP message
type Message struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleMessage processes an LSP message and returns a response
func (s *Server) HandleMessage(msg *Message) *Message {
	// Skip responses
	if msg.Result != nil || msg.Error != nil {
		return nil
	}

	// Route by method
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return nil
	case "shutdown":
		return s.handleShutdown(msg)
	case "exit":
		return nil
	case "textDocument/didOpen":
		return s.handleDidOpen(msg)
	case "textDocument/didChange":
		return s.handleDidChange(msg)
	case "textDocument/didClose":
		return s.handleDidClose(msg)
	case "textDocument/hover":
		return s.handleHover(msg)
	case "textDocument/definition":
		return s.handleDefinition(msg)
	case "textDocument/references":
		return s.handleReferences(msg)
	case "textDocument/completion":
		return s.handleCompletion(msg)
	default:
		// Unknown method
		if msg.ID != nil {
			errCode := -32601 // Method not found
			return &Message{
				Jsonrpc: "2.0",
				ID:      msg.ID,
				Error: &JSONRPCError{
					Code:    errCode,
					Message: "Method not found",
				},
			}
		}
		return nil
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(msg *Message) *Message {
	response := s.Initialize()

	return &Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result:  response,
	}
}

// handleShutdown handles the shutdown request
func (s *Server) handleShutdown(msg *Message) *Message {
	if err := s.Shutdown(); err != nil {
		errCode := -32603 // Internal error
		return &Message{
			Jsonrpc: "2.0",
			ID:      msg.ID,
			Error: &JSONRPCError{
				Code:    errCode,
				Message: err.Error(),
			},
		}
	}

	return &Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result:  nil,
	}
}

// handleDidOpen handles the textDocument/didOpen notification
func (s *Server) handleDidOpen(msg *Message) *Message {
	var params DidOpenParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	diagnostics := s.DidOpen(params)

	// Send diagnostics notification
	return &Message{
		Jsonrpc: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params:  json.RawMessage(mustMarshal(&PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: diagnostics,
		})),
	}
}

// handleDidChange handles the textDocument/didChange notification
func (s *Server) handleDidChange(msg *Message) *Message {
	var params DidChangeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	diagnostics := s.DidChange(params)

	// Send diagnostics notification
	return &Message{
		Jsonrpc: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params:  json.RawMessage(mustMarshal(&PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: diagnostics,
		})),
	}
}

// handleDidClose handles the textDocument/didClose notification
func (s *Server) handleDidClose(msg *Message) *Message {
	var params DidCloseParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil
	}

	s.DidClose(params)
	return nil
}

// handleHover handles the textDocument/hover request
func (s *Server) handleHover(msg *Message) *Message {
	var params HoverParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		errCode := -32700 // Parse error
		return &Message{
			Jsonrpc: "2.0",
			ID:      msg.ID,
			Error: &JSONRPCError{
				Code:    errCode,
				Message: err.Error(),
			},
		}
	}

	hover := s.Hover(params)

	return &Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result:  hover,
	}
}

// handleDefinition handles the textDocument/definition request
func (s *Server) handleDefinition(msg *Message) *Message {
	var params DefinitionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		errCode := -32700 // Parse error
		return &Message{
			Jsonrpc: "2.0",
			ID:      msg.ID,
			Error: &JSONRPCError{
				Code:    errCode,
				Message: err.Error(),
			},
		}
	}

	location := s.Definition(params)

	return &Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result:  location,
	}
}

// handleReferences handles the textDocument/references request
func (s *Server) handleReferences(msg *Message) *Message {
	var params ReferenceParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		errCode := -32700 // Parse error
		return &Message{
			Jsonrpc: "2.0",
			ID:      msg.ID,
			Error: &JSONRPCError{
				Code:    errCode,
				Message: err.Error(),
			},
		}
	}

	locations := s.References(params)

	return &Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result:  locations,
	}
}

// handleCompletion handles the textDocument/completion request
func (s *Server) handleCompletion(msg *Message) *Message {
	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		errCode := -32700 // Parse error
		return &Message{
			Jsonrpc: "2.0",
			ID:      msg.ID,
			Error: &JSONRPCError{
				Code:    errCode,
				Message: err.Error(),
			},
		}
	}

	items := s.Completion(params)

	return &Message{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result:  items,
	}
}

// mustMarshal marshals data and panics on error
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// StringWithTimeout reads a string with a timeout
func StringWithTimeout(r *bufio.Reader, timeout time.Duration) (string, error) {
	// This is a helper for potential future use
	return r.ReadString('\n')
}
