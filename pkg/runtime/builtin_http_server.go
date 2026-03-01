package runtime

import (
	"fmt"
	"strings"
	"time"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// builtinHTTPServer returns a stateful HTTP server object with methods:
//   - .route(method, path, handler_script_path) - Register a route with a handler script
//   - .static(path, directory) - Serve static files from a directory
//   - .start() - Start the server (blocks until Ctrl+C)
//
// http_server() returns a stateful HTTP server object with methods:
//   - .route(method, path, handler_script_path) - Register a route with a handler script
//   - .static(path, directory) - Serve static files from a directory
//   - .start() - Start the server (blocks until Ctrl+C)
//
// Configuration options:
//   - port (number) - Port to listen on (default: 8080)
//   - address (string) - Bind address (default: "0.0.0.0")
//   - https (boolean) - Enable HTTPS (default: false)
//   - cert_file (string) - Path to TLS certificate
//   - key_file (string) - Path to TLS private key
//   - timeout (number) - Read/write timeout in seconds (default: 30)
//
// Examples:
//
// Static file server:
//
//	server = http_server({port = 8080})
//	server.static("/", "./public")
//	server.start()
//
// Handler-based server:
//
//	server = http_server({port = 8080})
//	server.route("GET", "/hello", "handlers/hello.du")
//	server.start()
//
// Quick testing with -c:
//
//	duso -c 'server = http_server({port = 8080}); server.static("/", "."); server.start()'
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
		Port:                  8080,                          // default
		Address:               "0.0.0.0",                     // default
		Timeout:               30 * time.Second,              // default socket timeout
		RequestHandlerTimeout: 30 * time.Second,              // default handler script timeout
		DefaultFiles:          []string{"index.html"},        // default
		FileReader:            globalInterpreter.FileReader,  // Use host's FileReader capability
		FileStatter:           globalInterpreter.FileStatter, // Use host's FileStatter capability
		DirReader:             globalInterpreter.DirReader,   // Use host's DirReader capability
		Interpreter:           globalInterpreter,             // Store interpreter for optional script path
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

	// Parse directory listing setting
	if directory, ok := config["directory"]; ok {
		if dirFlag, ok := directory.(bool); ok {
			server.ShowDirectoryListing = dirFlag
		}
	}

	// Parse default files
	if defaultFiles, ok := config["default"]; ok {
		switch v := defaultFiles.(type) {
		case nil:
			// nil clears defaults
			server.DefaultFiles = nil
		case string:
			if v == "" {
				// Empty string clears defaults
				server.DefaultFiles = nil
			} else {
				// Parse comma-separated list of default files
				server.DefaultFiles = nil // Clear first
				parts := strings.Split(v, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part != "" {
						server.DefaultFiles = append(server.DefaultFiles, part)
					}
				}
			}
		case []any:
			if len(v) == 0 {
				// Empty array clears defaults
				server.DefaultFiles = nil
			} else {
				// Array of strings
				server.DefaultFiles = nil // Clear first
				for _, item := range v {
					if itemStr, ok := item.(string); ok {
						server.DefaultFiles = append(server.DefaultFiles, itemStr)
					}
				}
			}
		}
	}

	// Parse CORS config
	if corsRaw, ok := config["cors"]; ok {
		if corsMap, ok := corsRaw.(map[string]any); ok {
			// Parse enabled flag
			if enabled, ok := corsMap["enabled"].(bool); ok {
				server.CORS.Enabled = enabled
			}

			// Parse allowed origins (string or array)
			if origins, ok := corsMap["origins"]; ok {
				switch o := origins.(type) {
				case string:
					server.CORS.AllowedOrigins = []string{o}
				case []any:
					for _, origin := range o {
						if originStr, ok := origin.(string); ok {
							server.CORS.AllowedOrigins = append(server.CORS.AllowedOrigins, originStr)
						}
					}
				}
			}

			// Parse allowed methods (string or array)
			if methods, ok := corsMap["methods"]; ok {
				switch m := methods.(type) {
				case string:
					server.CORS.AllowedMethods = []string{strings.ToUpper(m)}
				case []any:
					for _, method := range m {
						if methodStr, ok := method.(string); ok {
							server.CORS.AllowedMethods = append(server.CORS.AllowedMethods, strings.ToUpper(methodStr))
						}
					}
				}
			}

			// Parse allowed headers (string or array)
			if headers, ok := corsMap["headers"]; ok {
				switch h := headers.(type) {
				case string:
					server.CORS.AllowedHeaders = []string{h}
				case []any:
					for _, header := range h {
						if headerStr, ok := header.(string); ok {
							server.CORS.AllowedHeaders = append(server.CORS.AllowedHeaders, headerStr)
						}
					}
				}
			}

			// Parse credentials flag
			if credentials, ok := corsMap["credentials"].(bool); ok {
				server.CORS.AllowCredentials = credentials
			}

			// Parse max age
			if maxAge, ok := corsMap["max_age"].(float64); ok {
				server.CORS.MaxAge = int(maxAge)
			}
		}
	}

	// Parse JWT config
	if jwtRaw, ok := config["jwt"]; ok {
		if jwtMap, ok := jwtRaw.(map[string]any); ok {
			// Parse enabled flag
			if enabled, ok := jwtMap["enabled"].(bool); ok {
				server.JWT.Enabled = enabled
			}

			// Parse secret
			if secret, ok := jwtMap["secret"].(string); ok {
				server.JWT.Secret = secret
			}

			// Parse required flag
			if required, ok := jwtMap["required"].(bool); ok {
				server.JWT.Required = required
			}
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

	// Create static() method
	staticFn := script.NewGoFunction(func(evaluator *script.Evaluator, staticArgs map[string]any) (any, error) {
		// Get path (first positional arg)
		path, ok := staticArgs["0"].(string)
		if !ok {
			return nil, fmt.Errorf("static() requires path and directory arguments")
		}

		// Get directory (second positional arg)
		dir, ok := staticArgs["1"].(string)
		if !ok {
			return nil, fmt.Errorf("static() requires path and directory arguments")
		}

		// Resolve relative paths to absolute paths using current working directory
		// This ensures "." refers to the cwd where duso was invoked, not the script directory
		absDir, err := core.Abs(dir)
		if err != nil {
			// If resolution fails, use the original path
			absDir = dir
		}

		// Register the static route
		return nil, server.StaticRoute(path, absDir)
	})

	// Return server object with methods
	return map[string]any{
		"route":  routeFn,
		"static": staticFn,
		"start":  startFn,
	}, nil
}
