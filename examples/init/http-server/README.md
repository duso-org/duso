# HTTP Server Template

A complete REST API server demonstrating:

- **http_server()** - Create and configure a server
- **route()** - Register endpoint handlers
- **context()** - Access request/response data
- **JSON APIs** - Serve structured data
- **Error handling** - Try/catch for robustness
- **Query parameters** - Parse and use request data

## Running

```bash
duso http-server.du
```

Then in another terminal:

```bash
# Home page
curl http://localhost:8080/

# JSON greeting
curl "http://localhost:8080/api/hello?name=Alice"

# Echo endpoint
curl -X POST http://localhost:8080/api/echo \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, Duso!"}'
```

## Structure

- **http-server.du** - Server setup and route registration
- **handlers/** - Individual endpoint handlers
  - `home.du` - HTML home page
  - `hello.du` - JSON greeting with query params
  - `echo.du` - Echo back request data

## Learn More

- [http_server() documentation](/docs/reference/http_server.md)
- [Building HTTP servers](/docs/learning-duso.md#building-http-servers)
- [context() documentation](/docs/reference/context.md)
