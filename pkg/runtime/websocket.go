package runtime

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/duso-org/duso/pkg/core"
	"golang.org/x/net/websocket"
)

// Global registry of active WebSocket connections keyed by connection ID
var (
	wsConnRegistry = make(map[string]*WebSocketConnection)
	wsConnMutex    sync.RWMutex
	interruptChan  = make(chan struct{})
)

// WebSocketConfig holds configuration for WebSocket connections
type WebSocketConfig struct {
	ReadQueueSize         int
	WriteQueueSize        int
	DefaultReadTimeout    time.Duration
	IdleTimeout           time.Duration // 0 = no idle disconnect
	MaxMessageSize        int64          // 0 = unlimited
	MaxMessagesPerSecond  int            // 0 = unlimited
}

// DefaultWebSocketConfig returns sensible defaults
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		ReadQueueSize:        100,
		WriteQueueSize:       100,
		DefaultReadTimeout:   30 * time.Second,
		IdleTimeout:          300 * time.Second, // 5 minutes
		MaxMessageSize:       65536,             // 64KB
		MaxMessagesPerSecond: 0,                 // 0 = unlimited
	}
}

// WebSocketConnection represents an active WebSocket connection in Duso
type WebSocketConnection struct {
	ws                *websocket.Conn
	closed            bool
	mutex             sync.Mutex
	id                string
	config            WebSocketConfig
	readQ             chan string
	writeQ            chan string
	readDone          chan struct{}
	doneMu            sync.Once
	lastActivityTime  time.Time // Track idle timeout
	violationCount    int       // Rate limit violation counter
	lastViolationTime time.Time // For violation decay
	tokens            float64   // Token bucket for rate limiting
}

// NewWebSocketConnection creates a new WebSocket connection wrapper (server-side)
func NewWebSocketConnection(ws *websocket.Conn) *WebSocketConnection {
	return NewWebSocketConnectionWithConfig(ws, DefaultWebSocketConfig())
}

// NewWebSocketConnectionWithConfig creates a connection with custom config
func NewWebSocketConnectionWithConfig(ws *websocket.Conn, config WebSocketConfig) *WebSocketConnection {
	conn := &WebSocketConnection{
		ws:               ws,
		closed:           false,
		id:               generateUUIDv4(),
		config:           config,
		readQ:            make(chan string, config.ReadQueueSize),
		writeQ:           make(chan string, config.WriteQueueSize),
		readDone:         make(chan struct{}),
		lastActivityTime: time.Now(),
	}

	// Register in global registry
	RegisterConnection(conn)

	// Start background reader goroutine
	go conn.backgroundReader()

	// Start background writer goroutine
	go conn.backgroundWriter()

	// Start idle timeout monitor if configured
	if config.IdleTimeout > 0 {
		go conn.idleTimeoutMonitor()
	}

	return conn
}

// NewWebSocketClientConnection creates a client WebSocket connection
func NewWebSocketClientConnection(urlStr string, headers map[string]string) (*WebSocketConnection, error) {
	return NewWebSocketClientConnectionWithConfig(urlStr, headers, DefaultWebSocketConfig())
}

// NewWebSocketClientConnectionWithConfig creates a client connection with custom config
func NewWebSocketClientConnectionWithConfig(urlStr string, headers map[string]string, config WebSocketConfig) (*WebSocketConnection, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	// Convert http/https to ws/wss
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
		// Already correct
	default:
		return nil, fmt.Errorf("unsupported scheme: %s (use http, https, ws, or wss)", u.Scheme)
	}

	// Dial the WebSocket
	wsConfig, err := websocket.NewConfig(u.String(), u.String())
	if err != nil {
		return nil, fmt.Errorf("invalid WebSocket config: %w", err)
	}

	// Add custom headers
	for k, v := range headers {
		wsConfig.Header.Set(k, v)
	}

	ws, err := websocket.DialConfig(wsConfig)
	if err != nil {
		return nil, fmt.Errorf("WebSocket connection failed: %w", err)
	}

	conn := &WebSocketConnection{
		ws:               ws,
		closed:           false,
		id:               generateUUIDv4(),
		config:           config,
		readQ:            make(chan string, config.ReadQueueSize),
		writeQ:           make(chan string, config.WriteQueueSize),
		readDone:         make(chan struct{}),
		lastActivityTime: time.Now(),
	}

	// Register in global registry
	RegisterConnection(conn)

	// Start background reader goroutine
	go conn.backgroundReader()

	// Start background writer goroutine
	go conn.backgroundWriter()

	// Start idle timeout monitor if configured
	if config.IdleTimeout > 0 {
		go conn.idleTimeoutMonitor()
	}

	return conn, nil
}

// Accept accepts the WebSocket connection (protocol handshake already done by upgrade)
func (wsc *WebSocketConnection) Accept() error {
	wsc.mutex.Lock()
	defer wsc.mutex.Unlock()

	if wsc.closed {
		return fmt.Errorf("connection already closed")
	}

	// Connection is already accepted by the HTTP upgrade
	return nil
}

// Read checks the read queue, blocking with optional timeout if empty
// Returns message string on success (including empty string), error on disconnect
// If timeout is specified and expires, returns ("", nil) to indicate timeout
func (wsc *WebSocketConnection) Read(timeout *time.Duration) (string, error) {
	var timeoutChan <-chan time.Time
	if timeout != nil {
		timeoutChan = time.After(*timeout)
	}

	select {
	case msg := <-wsc.readQ:
		return msg, nil // Return message as-is, even if empty
	case <-timeoutChan:
		return "", nil // Timeout: return empty string (caller should check with explicit timeout check)
	case <-wsc.readDone:
		return "", fmt.Errorf("connection closed") // Connection closed: error indicates disconnect
	case <-interruptChan:
		return "", fmt.Errorf("interrupted") // Process interrupted (Ctrl+C)
	}
}

// Write queues a message to the write queue
// Returns number of bytes queued, or nil if queue is full
func (wsc *WebSocketConnection) Write(message string) any {
	wsc.mutex.Lock()
	if wsc.closed {
		wsc.mutex.Unlock()
		return nil
	}
	wsc.mutex.Unlock()

	select {
	case wsc.writeQ <- message:
		return float64(len(message))
	default:
		return nil // Queue full
	}
}

// Close closes the WebSocket connection
func (wsc *WebSocketConnection) Close() error {
	wsc.mutex.Lock()
	if wsc.closed {
		wsc.mutex.Unlock()
		return nil
	}
	wsc.closed = true
	wsc.mutex.Unlock()

	// Unregister from global registry
	UnregisterConnection(wsc.id)

	// Signal readDone to wake up any waiting readers
	wsc.doneMu.Do(func() {
		close(wsc.readDone)
	})

	return wsc.ws.Close()
}

// backgroundReader reads from the WebSocket and queues messages
func (wsc *WebSocketConnection) backgroundReader() {
	defer core.RecoverPanic(fmt.Sprintf("websocket_reader (id=%s)", wsc.id))
	defer func() {
		wsc.mutex.Lock()
		wsc.closed = true
		wsc.mutex.Unlock()
		wsc.doneMu.Do(func() {
			close(wsc.readDone)
		})
	}()

	for {
		var msg string
		err := websocket.Message.Receive(wsc.ws, &msg)
		if err != nil {
			return // Connection closed or error
		}

		// Update last activity time
		wsc.mutex.Lock()
		wsc.lastActivityTime = time.Now()
		wsc.mutex.Unlock()

		// Check message size limit
		if wsc.config.MaxMessageSize > 0 && int64(len(msg)) > wsc.config.MaxMessageSize {
			// Message too large - close connection
			return
		}

		// Check rate limit
		if wsc.checkRateLimit() {
			// Rate limit exceeded, close connection
			return
		}

		select {
		case wsc.readQ <- msg:
			// Queued successfully
		case <-wsc.readDone:
			return // Connection closing
		}
	}
}

// backgroundWriter drains the write queue and sends to WebSocket
func (wsc *WebSocketConnection) backgroundWriter() {
	defer core.RecoverPanic(fmt.Sprintf("websocket_writer (id=%s)", wsc.id))
	defer wsc.ws.Close()

	for msg := range wsc.writeQ {
		if err := websocket.Message.Send(wsc.ws, msg); err != nil {
			return // Send failed, connection is dead
		}

		// Update last activity time on successful write
		wsc.mutex.Lock()
		wsc.lastActivityTime = time.Now()
		wsc.mutex.Unlock()
	}
}

// IsConnected returns whether the connection is still open
func (wsc *WebSocketConnection) IsConnected() bool {
	wsc.mutex.Lock()
	defer wsc.mutex.Unlock()
	return !wsc.closed
}

// ID returns the unique identifier for this connection
func (wsc *WebSocketConnection) ID() string {
	return wsc.id
}

// RegisterConnection adds a connection to the global registry
func RegisterConnection(conn *WebSocketConnection) {
	wsConnMutex.Lock()
	defer wsConnMutex.Unlock()
	wsConnRegistry[conn.id] = conn
}

// UnregisterConnection removes a connection from the global registry
func UnregisterConnection(connID string) {
	wsConnMutex.Lock()
	defer wsConnMutex.Unlock()
	delete(wsConnRegistry, connID)
}

// GetConnection retrieves a connection by ID
func GetConnection(connID string) *WebSocketConnection {
	wsConnMutex.RLock()
	defer wsConnMutex.RUnlock()
	return wsConnRegistry[connID]
}

// SignalInterrupt closes the interrupt channel to wake up all blocked read operations
// This is called when Ctrl+C or similar signals are received
func SignalInterrupt() {
	select {
	case <-interruptChan:
		// Already closed, do nothing
	default:
		close(interruptChan)
	}
}

// CloseAllConnections closes all active WebSocket connections
// Used during server shutdown to ensure Ctrl+C interrupts waiting connections
func CloseAllConnections() {
	wsConnMutex.Lock()
	conns := make([]*WebSocketConnection, 0, len(wsConnRegistry))
	for _, conn := range wsConnRegistry {
		conns = append(conns, conn)
	}
	wsConnMutex.Unlock()

	for _, conn := range conns {
		// Force unblock any blocked Receive() calls by setting a 1-second read deadline
		// This allows backgroundReader goroutines to exit immediately during shutdown
		conn.ws.SetReadDeadline(time.Now().Add(1 * time.Second))
		conn.Close()
	}
}

// generateUUIDv4 generates a UUID v4 (random) string
func generateUUIDv4() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("ws_%d", time.Now().UnixNano())
	}

	// Set version 4 (random) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// idleTimeoutMonitor closes the connection if idle too long
func (wsc *WebSocketConnection) idleTimeoutMonitor() {
	defer core.RecoverPanic(fmt.Sprintf("websocket_idle_monitor (id=%s)", wsc.id))
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wsc.mutex.Lock()
			if wsc.closed {
				wsc.mutex.Unlock()
				return
			}
			if wsc.config.IdleTimeout > 0 {
				elapsed := time.Since(wsc.lastActivityTime)
				if elapsed >= wsc.config.IdleTimeout {
					wsc.closed = true
					wsc.mutex.Unlock()
					wsc.doneMu.Do(func() {
						close(wsc.readDone)
					})
					wsc.ws.Close()
					UnregisterConnection(wsc.id)
					return
				}
			}
			wsc.mutex.Unlock()
		case <-wsc.readDone:
			return
		}
	}
}

// checkRateLimit checks if we're over the rate limit using token bucket
// Returns true if message should be dropped, false if it's OK
func (wsc *WebSocketConnection) checkRateLimit() bool {
	if wsc.config.MaxMessagesPerSecond <= 0 {
		return false // Rate limiting disabled
	}

	wsc.mutex.Lock()
	defer wsc.mutex.Unlock()

	now := time.Now()

	// Initialize tokens on first call
	if wsc.lastViolationTime.IsZero() {
		wsc.tokens = float64(wsc.config.MaxMessagesPerSecond)
		wsc.lastViolationTime = now
		wsc.tokens-- // Consume 1 token for this message
		return false
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(wsc.lastViolationTime).Seconds()
	wsc.tokens += elapsed * float64(wsc.config.MaxMessagesPerSecond)

	// Cap tokens at max (prevents accumulating huge buffer)
	maxTokens := float64(wsc.config.MaxMessagesPerSecond) * 2
	if wsc.tokens > maxTokens {
		wsc.tokens = maxTokens
	}

	wsc.lastViolationTime = now

	// Try to consume 1 token
	if wsc.tokens >= 1.0 {
		wsc.tokens -= 1.0
		wsc.violationCount = 0 // Reset violation counter on success
		return false
	}

	// Over limit - count violation
	wsc.violationCount++
	if wsc.violationCount >= 10 {
		// Too many violations - close connection
		wsc.closed = true
		wsc.doneMu.Do(func() {
			close(wsc.readDone)
		})
		go wsc.ws.Close()
		UnregisterConnection(wsc.id)
		return true
	}

	// Drop this message
	return true
}

// IsWebSocketUpgrade checks if the request is a WebSocket upgrade request
func IsWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Upgrade")) == "websocket" &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// WebSocketHandler creates an http.Handler that performs WebSocket upgrade
// and calls the provided upgrade handler
func WebSocketHandler(upgradeHandler func(*WebSocketConnection, *http.Request) error) http.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		conn := NewWebSocketConnection(ws)
		// Get the underlying HTTP request
		upgradeHandler(conn, ws.Request())
	})
}
