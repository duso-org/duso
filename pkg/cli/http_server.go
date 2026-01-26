package cli

import (
	"fmt"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// NewHTTPServerFunction creates the http_server(config) builtin.
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
func NewHTTPServerFunction(interp *script.Interpreter) func(map[string]any) (any, error) {
	return func(args map[string]any) (any, error) {
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
		server := &script.HTTPServerValue{
			Port:        8080,                  // default
			Address:     "0.0.0.0",             // default
			Timeout:     30 * time.Second,      // default
			FileReader:  readFile,              // Use cli's readFile function
			FileStatter: getFileMtime,          // Use cli's getFileMtime function
			Interpreter: interp,                // Store interpreter for optional script path
		}

		// Get parent evaluator from interpreter if available
		if interp != nil && interp.GetEvaluator() != nil {
			server.ParentEval = interp.GetEvaluator()
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

		// Create route() method
		routeFn := script.NewGoFunction(func(routeArgs map[string]any) (any, error) {
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

			return nil, server.Route(methodArg, path, handlerPath)
		})

		// Create start() method
		startFn := script.NewGoFunction(func(startArgs map[string]any) (any, error) {
			return nil, server.Start()
		})

		// Return server object with methods
		return map[string]any{
			"route": routeFn,
			"start": startFn,
		}, nil
	}
}
