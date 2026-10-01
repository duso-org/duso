package runtime

import (
	"fmt"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// builtinSendWebSocket sends a message to one or more WebSocket connections by ID
// Usage: send_websocket(conn_id, message) or send_websocket([conn_ids], message)
// Returns: bytes sent (number) or nil for a single ID; for an array of IDs, an
// array of those results, one per ID, even when the array has one element.
func builtinSendWebSocket(evaluator *Evaluator, args map[string]any) (any, error) {
	_, hasPos := args["0"]
	_, hasNamed := args["conn_id"]
	if !hasPos && !hasNamed {
		return nil, fmt.Errorf("send_websocket() requires a connection ID or array of IDs")
	}
	idArg := GetArg(args, 0, "conn_id")

	// Get message and stringify it
	msg := GetArg(args, 1, "message")
	if msg == nil {
		return nil, fmt.Errorf("send_websocket() requires a message")
	}
	message := fmt.Sprintf("%v", msg)

	ids := InterfaceToValue(idArg)
	switch {
	case ids.IsString():
		id := ids.AsString()
		if id == "" {
			return nil, fmt.Errorf("send_websocket() connection ID cannot be empty")
		}
		return sendWebSocketTo(id, message), nil
	case ids.IsArray():
		// One result per ID; a non-string or empty ID is nil, like an unknown one.
		arr := ids.AsArray()
		results := make([]any, len(arr))
		for i, id := range arr {
			if id.IsString() && id.AsString() != "" {
				results[i] = sendWebSocketTo(id.AsString(), message)
			}
		}
		return results, nil
	}
	// nil or a non-string ID matches no connection, same as an unknown ID
	return nil, nil
}

// sendWebSocketTo queues message on the connection with the given ID, returning
// bytes queued, or nil if the connection is unknown or its queue is full.
func sendWebSocketTo(connID, message string) any {
	conn := GetConnection(connID)
	if conn == nil {
		return nil
	}
	return conn.Write(message)
}

// builtinWebSocket establishes a WebSocket client connection
// Usage: websocket(url [, config])
// Returns: WebSocket connection object with read(), write(), close(), is_connected(), and id methods
func builtinWebSocket(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get URL from first positional or named argument
	var url string

	if u, ok := args["0"]; ok {
		url = fmt.Sprintf("%v", u)
	} else if u, ok := args["url"]; ok {
		url = fmt.Sprintf("%v", u)
	} else {
		return nil, fmt.Errorf("websocket() requires a URL")
	}

	if url == "" {
		return nil, fmt.Errorf("websocket() URL cannot be empty")
	}

	// Get options from second positional or named argument
	var headers map[string]string
	wsConfig := DefaultWebSocketConfig()

	if opts, ok := args["1"]; ok {
		if optsMap, ok := opts.(map[string]any); ok {
			headers = make(map[string]string)
			if headersOpt, ok := optsMap["headers"]; ok {
				if headerMap, ok := headersOpt.(map[string]any); ok {
					for k, v := range headerMap {
						headers[k] = fmt.Sprintf("%v", v)
					}
				}
			}
			// Parse WebSocket config options
			if readQSize, ok := optsMap["read_queue_size"].(float64); ok {
				wsConfig.ReadQueueSize = int(readQSize)
			}
			if writeQSize, ok := optsMap["write_queue_size"].(float64); ok {
				wsConfig.WriteQueueSize = int(writeQSize)
			}
			if readTimeout, ok := optsMap["read_timeout"].(float64); ok {
				wsConfig.DefaultReadTimeout = time.Duration(readTimeout) * time.Second
			}
			if idleTimeout, ok := optsMap["idle_timeout"].(float64); ok {
				wsConfig.IdleTimeout = time.Duration(idleTimeout) * time.Second
			}
			if maxMsgSize, ok := optsMap["max_message_size"].(float64); ok {
				wsConfig.MaxMessageSize = int64(maxMsgSize)
			}
			if maxMsgPerSec, ok := optsMap["max_messages_per_second"].(float64); ok {
				wsConfig.MaxMessagesPerSecond = int(maxMsgPerSec)
			}
		}
	} else if opts, ok := args["config"]; ok {
		if optsMap, ok := opts.(map[string]any); ok {
			headers = make(map[string]string)
			if headersOpt, ok := optsMap["headers"]; ok {
				if headerMap, ok := headersOpt.(map[string]any); ok {
					for k, v := range headerMap {
						headers[k] = fmt.Sprintf("%v", v)
					}
				}
			}
			// Parse WebSocket config options
			if readQSize, ok := optsMap["read_queue_size"].(float64); ok {
				wsConfig.ReadQueueSize = int(readQSize)
			}
			if writeQSize, ok := optsMap["write_queue_size"].(float64); ok {
				wsConfig.WriteQueueSize = int(writeQSize)
			}
			if readTimeout, ok := optsMap["read_timeout"].(float64); ok {
				wsConfig.DefaultReadTimeout = time.Duration(readTimeout) * time.Second
			}
			if idleTimeout, ok := optsMap["idle_timeout"].(float64); ok {
				wsConfig.IdleTimeout = time.Duration(idleTimeout) * time.Second
			}
			if maxMsgSize, ok := optsMap["max_message_size"].(float64); ok {
				wsConfig.MaxMessageSize = int64(maxMsgSize)
			}
			if maxMsgPerSec, ok := optsMap["max_messages_per_second"].(float64); ok {
				wsConfig.MaxMessagesPerSecond = int(maxMsgPerSec)
			}
		}
	} else {
		headers = make(map[string]string)
	}

	// Connect to WebSocket with config
	conn, err := NewWebSocketClientConnectionWithConfig(url, headers, wsConfig)
	if err != nil {
		return nil, err
	}

	// Return object with id and methods
	return map[string]any{
		"id": conn.ID(),
		"read": script.NewGoFunction(func(evaluator *Evaluator, args map[string]any) (any, error) {
			// Get optional timeout (positional or named)
			var timeout *time.Duration
			if t, ok := args["0"]; ok {
				if timeoutSec, ok := t.(float64); ok && timeoutSec > 0 {
					d := time.Duration(timeoutSec * float64(time.Second))
					timeout = &d
				}
			} else if t, ok := args["timeout"]; ok {
				if timeoutSec, ok := t.(float64); ok && timeoutSec > 0 {
					d := time.Duration(timeoutSec * float64(time.Second))
					timeout = &d
				}
			}

			msg, err := conn.Read(timeout)
			if err != nil {
				return nil, nil // Timeout or connection closed
			}
			return msg, nil // Actual message, including an empty one
		}),
		"write": script.NewGoFunction(func(evaluator *Evaluator, args map[string]any) (any, error) {
			msg, ok := args["0"]
			if !ok {
				return nil, fmt.Errorf("write() requires a message argument")
			}
			msgStr := fmt.Sprintf("%v", msg)
			return conn.Write(msgStr), nil // Returns bytes (number) or nil on queue full
		}),
		"close": script.NewGoFunction(func(evaluator *Evaluator, args map[string]any) (any, error) {
			return nil, conn.Close()
		}),
		"is_connected": script.NewGoFunction(func(evaluator *Evaluator, args map[string]any) (any, error) {
			return conn.IsConnected(), nil
		}),
	}, nil
}
