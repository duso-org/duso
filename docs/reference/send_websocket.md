# send_websocket()

Send a message to one or more WebSocket connections by ID. Used to send messages from outside the connection's handler (cross-handler messaging, orchestration, broadcasting, etc.).

`send_websocket(conn_id, message)` or `send_websocket([conn_ids], message)`

## Parameters

- `conn_id` (string or array) - Unique connection ID(s). Either a single ID (string) or array of IDs (array of strings)
- `message` (any type) - Message to send. Automatically converted to string if not already

## Returns

**Single connection:** Number of bytes queued (number), or `nil` if connection not found or write queue is full

**Array of IDs:** Array of results (one per ID, even for a one-element array), with each element being bytes queued (number) or `nil`. An empty array returns `[]`

## Description

Every WebSocket connection has a unique ID that persists for the lifetime of the connection. `send_websocket()` allows any script to send a message to a specific connection by referencing its ID, without needing direct access to the connection object itself.

Messages are queued in the connection's write buffer. If the buffer is full (client can't keep up), `send_websocket()` returns `nil` instead of queuing, preventing slow clients from blocking the sender.

**For broadcasting to multiple connections**, pass the array of IDs in one call, or loop over them and call `send_websocket()` for each one (see examples).

## Examples

Broadcasting to all connected clients:

```duso
// Server setup
server = http_server({port = 8080})
server.route("WS", "/chat", "handlers/chat.du")

// Orchestrator script - broadcast to all
store = datastore("connections")

function broadcast_message(msg)
  conn_ids = store.get("all_connections")
  if conn_ids then
    for id in conn_ids do
      bytes = send_websocket(id, msg)
      if bytes == nil then
        print("Failed to queue message to " + id)
      end
    end
  end
end

server.start()
```

```duso
// handlers/chat.du - register connection and handle messages
ctx = context()
conn = ctx.connection()
conn_id = conn.id

conn.accept()

// Register this connection globally
store = datastore("connections")
store.push("all_connections", conn_id)

// Listen for messages
while true do
  msg = conn.read()
  if msg == nil then break end
  
  // Echo to all users (via orchestrator)
  broadcast_message("User said: " + msg)
end
```

One-way system notifications (automatic stringification):

```duso
// System alert sent from any script (objects automatically converted to string)
user_conn_id = "user123_conn"

// Duso automatically stringifies objects
send_websocket(user_conn_id, {
  type = "alert",
  message = "System maintenance at 3 PM",
  severity = "info"
})

// Or explicit JSON
send_websocket(user_conn_id, format_json({
  type = "alert",
  message = "System maintenance at 3 PM"
}))
```

Broadcast to multiple connections at once:

```duso
store = datastore("room_chat")
conn_ids = ["user1_conn", "user2_conn", "user3_conn"]  // Connection IDs for room
store.set("lobby_conns", conn_ids)

// Send to all at once - returns array of results
results = send_websocket(conn_ids, "New user joined!")

// Check for queue overflows
for result in results do
  if result == nil then
    print("Failed to send to a connection")
  end
end
```

Selective messaging to specific users:

```duso
store = datastore("user_connections")

// Simulate storing a user's connection ID
store.set("user_alice", "alice_conn_123")

conn_id = store.get("user_alice")
if conn_id then
  // Send text message
  bytes = send_websocket(conn_id, "Your order shipped!")
  if bytes then
    print("Notification sent: " + tostring(bytes) + " bytes")
  end
  
  // Send object (automatically stringified)
  bytes = send_websocket(conn_id, {status = "ready", order_id = 123})
  if bytes then
    print("Object sent automatically stringified")
  end
else
  print("User not connected")
end
```

## Queue Overflow Behavior

If a connection's write queue is full (client is too slow to receive), `send_websocket()` returns `nil`:

```duso
store = datastore("user_connections")
conn_id = "user_123_conn"
msg = "broadcast message"

bytes = send_websocket(conn_id, msg)
if bytes == nil then
  // Queue full - message was not sent
  // Client is too slow, consider:
  // - Dropping the message
  // - Logging and retrying later
  // - Closing the connection
  print("Failed to send to " + conn_id)
  store.delete("user_" + conn_id)  // Mark as disconnected
end
```

## Connection Lifecycle

- **Both client and server connections are registered** in the global registry when created:
  - Client connections via `websocket()` (client-side)
  - Server connections via `http_server()` WebSocket handlers (server-side)
- Connection ID is available as `conn.id` (server-side) or `ws.id` (client-side)
- Connection is automatically unregistered when closed
- After unregistration, `send_websocket()` silently returns `nil` (connection not found)
- `send_websocket()` works with both client and server connections by ID

## Errors

`send_websocket()` does not throw errors. It returns:
- Bytes sent (number) if successfully queued
- `nil` if connection not found or queue is full

Use these return values to detect failed sends and handle appropriately.

## See Also

- [websocket() - Client connections](/docs/reference/websocket.md)
- [http_server() - WebSocket server](/docs/reference/http_server.md)
- [datastore() - Coordinate connections](/docs/reference/datastore.md)
