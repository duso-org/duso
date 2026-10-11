# websocket()

Establish a client WebSocket connection to a server.

`websocket(url [, config])`

## Parameters

- `url` (string) - WebSocket URL (ws://, wss://, http://, or https://)
- `config` (optional, object) - Configuration object:
  - `headers` (object) - Custom headers to send in the upgrade request (e.g., authentication tokens)
  - `read_queue_size` (number) - Max queued incoming messages (default: 100)
  - `write_queue_size` (number) - Max queued outgoing messages (default: 100)
  - `read_timeout` (number) - Default read timeout in seconds (default: 30)
  - `idle_timeout` (number) - Disconnect connection if idle for N seconds (default: 300). Set to 0 to disable.
  - `max_message_size` (number) - Max message size in bytes (default: 65536 = 64KB). Set to 0 for unlimited.
  - `max_messages_per_second` (number) - Rate limit: max messages per second (default: 0 = unlimited). Set to 0 to disable.

## Returns

WebSocket connection object with methods and properties

## Connection Object

The returned object has the following methods and properties:

- `id` (string) - Unique identifier for this connection (UUID v4). Can be used with `send_websocket()` to send messages from other scripts.
- `read([timeout])` - Block until a message is received
  - `timeout` (optional, number) - Wait timeout in seconds. If omitted, blocks indefinitely.
  - Returns the message string, or `nil` on disconnect or timeout.
  - Messages that arrived before a disconnect are still returned first.
  - Supports both positional: `read(5)` and named: `read(timeout=5)` arguments.
- `write(message)` - Send a message to the server
- `close()` - Close the connection gracefully: further `write()` calls return `nil`, queued messages are sent (up to 5 seconds), then the socket closes
- `kill()` - Close the connection immediately, discarding queued messages
- `is_connected()` - Check if connection is still open (returns boolean)

## Configuration Limits

WebSocket connections enforce three types of limits to prevent abuse and manage resources:

### Idle Timeout

Close connections that haven't sent or received data for N seconds:

```duso
ws = websocket("ws://localhost:8080/ws", {
  idle_timeout = 300  // Close if idle 5 minutes (default)
})
```

Set to 0 to disable idle timeout. Useful for long-lived connections that send infrequent keepalives.

### Message Size Limit

Reject messages larger than N bytes, closing the connection:

```duso
ws = websocket("ws://localhost:8080/ws", {
  max_message_size = 65536  // 64KB default, closes on violation
})
```

Set to 0 for unlimited. The connection closes immediately if a message exceeds the limit.

### Rate Limiting

Enforce a maximum message rate (messages per second). Excess messages are silently dropped:

```duso
ws = websocket("ws://localhost:8080/ws", {
  max_messages_per_second = 100  // Max 100 messages/sec
})
```

Set to 0 to disable rate limiting (default). When rate limit is exceeded:
1. Excess messages are silently dropped (soft rejection)
2. After 10 dropped messages in succession, the connection closes (hard disconnect)
3. Connection resumes normal operation if the sender backs off

This approach allows transient spikes while protecting against sustained abuse.

## Examples

Basic WebSocket connection:

```duso
ws = websocket("ws://localhost:8080/chat")
print("Connected with ID: " + ws.id)

// Send a message
ws.write("Hello, server!")

// Receive message with timeout
msg = ws.read(timeout=10)
if msg then
  print("Received: " + msg)
end

// Close connection
ws.close()
```

WebSocket connection with custom headers and queue config:

```duso
ws = websocket("wss://api.example.com/stream", {
  headers = {
    "Authorization" = "Bearer token_123",
    "X-Client-ID" = "my-client"
  },
  read_queue_size = 200,      // Handle bursts of 200 messages
  write_queue_size = 200,     // Queue up to 200 outgoing messages
  default_read_timeout = 60   // Default to 60s timeout on read()
})

// Message loop
while ws.is_connected() do
  msg = ws.read()  // Uses default_read_timeout if no timeout specified
  if msg == nil then
    print("Timeout or disconnected")
    break
  end
  print("Got: " + msg)
end

ws.close()
```

Automatic URL scheme conversion:

```duso
// http automatically converts to ws
ws1 = websocket("http://localhost:8080/ws")
print("Connected via http->ws: " + ws1.id)
ws1.close()

// https automatically converts to wss
ws2 = websocket("https://api.example.com/ws")
print("Connected via https->wss: " + ws2.id)
ws2.close()
```

Discord bot using WebSocket client:

```duso
ws = websocket("wss://gateway.discord.gg/", {
  headers = {
    "User-Agent" = "DiscordBot (my-bot/1.0)"
  }
})

ws.write(format_json({
  op = 0,
  d = {
    token = "bot-token",
    properties = {
      os = "linux",
      browser = "duso",
      device = "duso"
    }
  }
}))

// Listen for messages
while ws.is_connected() do
  msg = ws.read(timeout=60)
  if msg then
    event = parse_json(msg)
    print("Got event: " + tostring(event.op))
  end
end
```

## Error Handling

WebSocket operations throw errors on:
- Invalid URL format
- Connection failure (network error, timeout)
- Invalid scheme (must be ws, wss, http, or https)

Use try/catch to handle connection errors:

```duso
try
  ws = websocket("ws://unreachable.example.com")
  print("Connected: " + ws.id)
catch (e)
  print("Failed to connect: " + e)
end
```

## Message Format

Messages are always text (UTF-8 strings). Binary frames are not supported in this version.

## Timeout Behavior

- `read()` with no timeout blocks indefinitely until a message arrives
- `read(timeout)` returns `nil` if timeout expires before a message arrives
- `read()` returns `nil` if the server closes the connection
- Timeout is specified in seconds as a floating-point number

## Connection Lifecycle

1. **Connect** — `websocket()` dials the URL and returns immediately on success
2. **Read/Write** — Use `read()` and `write()` in a message loop
3. **Disconnect** — Server closes, client calls `close()` or `kill()`, or network error
4. **Cleanup** — After `close()` or `kill()`, further `read()`/`write()` calls fail gracefully

## Performance Notes

- `read()` is blocking—it blocks the script until a message arrives or timeout expires
- For long-running connections, use `read()` in a loop rather than polling
- Each script has its own connection; connections are not shared across spawned scripts
- Use `datastore()` to coordinate multiple WebSocket connections
- Messages are queued in buffers: incoming messages in `read_queue`, outgoing in `write_queue`
- If `write_queue` fills up, `write()` returns nil (queue overflow)—handle gracefully or reduce send rate
- If `read_queue` fills up, incoming messages are dropped and connection closes with error

## See Also

- [send_websocket() - Send to specific connection](/docs/reference/send_websocket.md) - Send messages to any WebSocket connection (client or server) by ID
- [http_server() - WebSocket server](/docs/reference/http_server.md) - Create WebSocket servers
- [fetch() - HTTP client requests](/docs/reference/fetch.md) - Make regular HTTP requests
- [spawn() - Run scripts concurrently](/docs/reference/spawn.md) - Run message handlers in parallel
- [datastore() - Shared state](/docs/reference/datastore.md) - Coordinate between connections
