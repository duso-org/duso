package runtime

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

// WebSocketConnection represents an active WebSocket connection in Duso
type WebSocketConnection struct {
	ws     *websocket.Conn
	closed bool
	mutex  sync.Mutex
}

// NewWebSocketConnection creates a new WebSocket connection wrapper
func NewWebSocketConnection(ws *websocket.Conn) *WebSocketConnection {
	return &WebSocketConnection{
		ws:     ws,
		closed: false,
	}
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

// Receive blocks until a message is received or connection closes
// Returns the message string, or error on disconnect
func (wsc *WebSocketConnection) Receive() (string, error) {
	wsc.mutex.Lock()
	if wsc.closed {
		wsc.mutex.Unlock()
		return "", fmt.Errorf("connection closed")
	}
	wsc.mutex.Unlock()

	var msg string
	err := websocket.Message.Receive(wsc.ws, &msg)

	if err != nil {
		wsc.mutex.Lock()
		wsc.closed = true
		wsc.mutex.Unlock()

		// Return error on EOF/disconnect
		if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "closed") {
			return "", nil // Nil equivalent for disconnect
		}
		return "", fmt.Errorf("websocket receive error: %w", err)
	}

	return msg, nil
}

// Send sends a message to the WebSocket client
func (wsc *WebSocketConnection) Send(message string) error {
	wsc.mutex.Lock()
	defer wsc.mutex.Unlock()

	if wsc.closed {
		return fmt.Errorf("connection closed")
	}

	return websocket.Message.Send(wsc.ws, message)
}

// Close closes the WebSocket connection
func (wsc *WebSocketConnection) Close() error {
	wsc.mutex.Lock()
	defer wsc.mutex.Unlock()

	if wsc.closed {
		return nil
	}

	wsc.closed = true
	return wsc.ws.Close()
}

// IsConnected returns whether the connection is still open
func (wsc *WebSocketConnection) IsConnected() bool {
	wsc.mutex.Lock()
	defer wsc.mutex.Unlock()
	return !wsc.closed
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
