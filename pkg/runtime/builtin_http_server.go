package runtime

import (
	"fmt"
	"strings"
	"time"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// builtinHTTPServer returns a stateful HTTP server object with methods:
//   - .route(method, path, handler_script_path) - Register a route
//   - .start() - Start the server in a background goroutine
//
// http_server() returns a stateful HTTP server object with methods:
//   - .route(method, path, handler_script_path) - Register a route
//   - .start() - Start the server in a background goroutine
//
// Configuration options:
//   - port (number) - Port to listen on (default: 8080)
//   - address (string) - Bind address (default: "0.0.0.0")
//   - https (boolean) - Enable HTTPS (default: false)
//   - cert_file (string) - Path to TLS certificate
//   - key_file (string) - Path to TLS private key
//   - timeout (number) - Read/write timeout in seconds (default: 30)
//
// Example:
//
//	server = http_server({port = 8080})
//	server.route("GET", "/hello", "handlers/hello.du")
//	server.start()
//	print("Server started on port 8080")
func builtinHTTPServer(evaluator *Evaluator, args map[string]any) (any, error) {
		// Get config from first positional or named argument
		var config map[string]any

		if cfg, ok := args["0"]; ok {
			// Positional argument
			if cfgMap, ok := cfg.(map[string]any); ok {
				config = cfgMap
			} else {
				return nil, fmt.Errorf("http_server() argument must be a config object")
			}
		} else if cfg, ok := args["config"]; ok {
			// Named argument
			if cfgMap, ok := cfg.(map[string]any); ok {
				config = cfgMap
			} else {
				return nil, fmt.Errorf("http_server() 'config' argument must be a config object")
			}
		} else {
			// Empty config
			config = make(map[string]any)
		}

		// Initialize server with defaults
		server := &HTTPServerValue{
			Port:                   8080,                  // default
			Address:                "0.0.0.0",             // default
			Timeout:                30 * time.Second,      // default socket timeout
			RequestHandlerTimeout:  30 * time.Second,      // default handler script timeout
			FileReader:             globalInterpreter.FileReader,     // Use host's FileReader capability
			FileStatter:            globalInterpreter.FileStatter,    // Use host's FileStatter capability
			Interpreter:            globalInterpreter,                // Store interpreter for optional script path
		}

		// Parse port
		if port, ok := config["port"]; ok {
			if portNum, ok := port.(float64); ok {
				server.Port = int(portNum)
			}
		}

		// Parse address
		if addr, ok := config["address"]; ok {
			server.Address = fmt.Sprintf("%v", addr)
		}

		// Parse HTTPS config
		if https, ok := config["https"]; ok {
			if httpsFlag, ok := https.(bool); ok && httpsFlag {
				server.TLSEnabled = true
			}
		}
		if certFile, ok := config["cert_file"]; ok {
			server.CertFile = fmt.Sprintf("%v", certFile)
		}
		if keyFile, ok := config["key_file"]; ok {
			server.KeyFile = fmt.Sprintf("%v", keyFile)
		}

		// Parse timeout in seconds
		if timeout, ok := config["timeout"]; ok {
			if timeoutSecs, ok := timeout.(float64); ok {
				server.Timeout = time.Duration(timeoutSecs) * time.Second
			}
		}

		// Parse request handler timeout in seconds
		if handlerTimeout, ok := config["request_handler_timeout"]; ok {
			if timeoutSecs, ok := handlerTimeout.(float64); ok {
				server.RequestHandlerTimeout = time.Duration(timeoutSecs) * time.Second
			}
		}

		// Create route() method
		routeFn := script.NewGoFunction(func(evaluator *script.Evaluator, routeArgs map[string]any) (any, error) {
			// Get the directory of the calling script for path resolution
			scriptDir := ""
			scriptFilePath := ""
			if server.Interpreter != nil {
				scriptFilePath = server.Interpreter.GetFilePath()
				if scriptFilePath != "" {
					scriptDir = core.Dir(scriptFilePath)
				}
			}

			// Get method (can be nil, string, or []string)
			methodArg := routeArgs["0"]

			path, ok := routeArgs["1"].(string)
			if !ok {
				return nil, fmt.Errorf("route() requires method, path, and optional handler arguments")
			}

			// Handler path is optional - defaults to current script
			handlerPath := ""
			if handlerArg, ok := routeArgs["2"]; ok {
				if handlerStr, ok := handlerArg.(string); ok {
					handlerPath = handlerStr
				}
			}

			// If no handler path provided, use current script
			if handlerPath == "" {
				if server.Interpreter != nil {
					handlerPath = server.Interpreter.GetFilePath()
				}
				if handlerPath == "" {
					return nil, fmt.Errorf("route() handler path required when script path unknown")
				}
					}
	
			// Register the route
			err := server.Route(methodArg, path, handlerPath)
				// Set scriptDir for ALL routes (both current script and external handlers)
			// This allows static files to be resolved relative to the handler script's directory
			if err == nil && scriptDir != "" {
				server.routeMutex.Lock()
				for key, route := range server.routes {
					if strings.HasSuffix(key, " "+path) {
						route.ScriptDir = scriptDir
					}
				}
				server.routeMutex.Unlock()
			}
			return nil, err
		})

		// Create start() method
		startFn := script.NewGoFunction(func(evaluator *script.Evaluator, startArgs map[string]any) (any, error) {
			return nil, server.Start()
		})

		// Return server object with methods
	return map[string]any{
		"route": routeFn,
		"start": startFn,
	}, nil
}
