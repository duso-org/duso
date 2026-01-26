# HTTP Server Examples

This directory contains examples of building HTTP servers with Duso.

## Quick Start

```bash
duso examples/http/server.du
```

Then visit:
- http://localhost:8080/ - Home page
- http://localhost:8080/hello - Simple greeting
- http://localhost:8080/hello?name=YourName - Greeting with parameter
- http://localhost:8080/api/time - JSON API

Press **Ctrl+C** to stop the server gracefully.

## Files

- **server.du** - Main server script that sets up routes and starts the server
- **handlers/** - Handler scripts for each route
  - **home.du** - Returns an HTML home page
  - **hello.du** - Returns a greeting, with optional query parameter
  - **time.du** - Returns JSON with timestamp

## How It Works

1. The main script creates an HTTP server on port 8080
2. Routes are registered with the handler script paths
3. `server.start()` blocks, listening for requests
4. Each incoming request spawns a fresh script instance
5. The handler script accesses the request via `context()` builtin
6. When Ctrl+C is pressed, the server shuts down gracefully
7. Cleanup code in the main script runs after shutdown

## Key Features Demonstrated

- **Multiline strings** - Using `"""..."""` for clean HTML/JSON
- **String templates** - Using `{{...}}` to embed variables
- **Request context** - Using `context()` to access request data
- **Query parameters** - Accessing `req.query["name"]`
- **Response handling** - Building response objects with status, headers, body
- **Graceful shutdown** - Cleanup code runs when server stops

## Exploring Further

Try modifying:
- Add new routes in `server.du`
- Create new handler scripts
- Use query parameters and request headers
- Return different content types (JSON, HTML, plain text, etc.)
