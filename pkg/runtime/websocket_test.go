package runtime

import (
	"errors"
	"testing"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// testConn builds a registered connection with queues but no socket or
// background goroutines, so sends land in writeQ where the test can see them.
func testConn(t *testing.T) *WebSocketConnection {
	t.Helper()
	conn := &WebSocketConnection{
		id:       generateUUIDv4(),
		readQ:    make(chan string, 4),
		writeQ:   make(chan string, 8),
		readDone: make(chan struct{}),
	}
	RegisterConnection(conn)
	t.Cleanup(func() { UnregisterConnection(conn.id) })
	return conn
}

// Arrays reach builtins as *[]Value. A []any-only type switch missed them, so
// every array send silently went nowhere and returned nil.
func TestSendWebSocketArray(t *testing.T) {
	conn := testConn(t)

	ids := []Value{script.NewString(conn.id)}
	out, err := builtinSendWebSocket(nil, map[string]any{"0": &ids, "1": "hi"})
	if err != nil {
		t.Fatalf("send_websocket failed: %v", err)
	}
	results, ok := out.([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("one-element array returned %#v, want a one-element array", out)
	}
	if results[0] != float64(2) {
		t.Errorf("result = %#v, want 2 bytes", results[0])
	}
	if got := <-conn.writeQ; got != "hi" {
		t.Errorf("queued %q, want %q", got, "hi")
	}

	mixed := []Value{script.NewString(conn.id), script.NewString("no-such-id"), script.NewNumber(42)}
	out, _ = builtinSendWebSocket(nil, map[string]any{"conn_id": &mixed, "message": "x"})
	results, _ = out.([]any)
	if len(results) != 3 || results[0] != float64(1) || results[1] != nil || results[2] != nil {
		t.Errorf("mixed array returned %#v, want [1, nil, nil]", out)
	}

	single, _ := builtinSendWebSocket(nil, map[string]any{"0": conn.id, "1": "abc"})
	if single != float64(3) {
		t.Errorf("single ID returned %#v, want 3", single)
	}

	missing, err := builtinSendWebSocket(nil, map[string]any{"0": nil, "1": "abc"})
	if err != nil || missing != nil {
		t.Errorf("nil ID returned (%#v, %v), want (nil, nil)", missing, err)
	}
}

// A timeout must be distinguishable from an empty message: the read bindings
// turn any error into nil, so it has to come back as an error.
func TestReadTimeoutIsNotEmptyMessage(t *testing.T) {
	conn := testConn(t)

	d := 10 * time.Millisecond
	if _, err := conn.Read(&d); !errors.Is(err, errReadTimeout) {
		t.Errorf("timeout returned err %v, want errReadTimeout", err)
	}

	conn.readQ <- ""
	msg, err := conn.Read(&d)
	if err != nil || msg != "" {
		t.Errorf("empty message returned (%q, %v), want (\"\", nil)", msg, err)
	}
}
