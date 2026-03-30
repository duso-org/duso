# WebSocket External Send

## Problem

WebSocket handler scripts own the connection exclusively. No other code (HTTP handlers, spawned scripts) can send messages to a connected socket.

## Design

### Registry scoped to server instance

Each `HTTPServerValue` gets its own connection registry:

```go
type HTTPServerValue struct {
    // ...existing fields...
    wsConns   map[string]*WebSocketConnection
    wsConnsMu sync.RWMutex
}
```

Connections auto-register on upgrade, auto-deregister on close. IDs are auto-assigned (atomic uint64 counter).

### Script API

**In websocket handlers:**
```python
conn = context().connection
conn.id  # auto-assigned connection id
```

**In any handler on the same server:**
```python
server = context().server
server.ws_send(id, message)       # send to one socket
server.ws_broadcast(message)      # send to all sockets
server.ws_connections()           # list active connection ids
```

App-level mappings (user->socket, rooms, etc.) are managed in userland via datastore.

### Lock discipline

Never hold registry lock during I/O. Copy pointer out, release, then send:

```go
func (s *HTTPServerValue) WsSend(id string, msg string) error {
    s.wsConnsMu.RLock()
    conn := s.wsConns[id]
    s.wsConnsMu.RUnlock()

    if conn == nil {
        return fmt.Errorf("connection not found: %s", id)
    }
    return conn.Send(msg)  // conn has its own mutex
}

func (s *HTTPServerValue) WsBroadcast(msg string) {
    s.wsConnsMu.RLock()
    conns := make([]*WebSocketConnection, 0, len(s.wsConns))
    for _, c := range s.wsConns {
        conns = append(conns, c)
    }
    s.wsConnsMu.RUnlock()

    for _, conn := range conns {
        conn.Send(msg)
    }
}
```

Two independent lock scopes (registry mutex, per-connection mutex), no nesting, no I/O under lock.

### Cross-server

Each server instance is isolated. Cross-server messaging (if needed) bridges through the datastore, which is already global/namespaced.

## Implementation touches

1. Add `id` field + atomic counter to `WebSocketConnection`
2. Add `wsConns` map + mutex to `HTTPServerValue`
3. Register in `handleWebSocketRequest`, deregister in `Close()`
4. Expose `context().server` with `ws_send`, `ws_broadcast`, `ws_connections` methods
5. Expose `conn.id` in `GetWSConnection()`
