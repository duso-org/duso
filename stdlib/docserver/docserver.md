# Docserver

Launch an HTTP server that serves documentation files as HTML with markdown rendering.

## Usage

```bash
duso -docserver
```

This starts an HTTP server on `http://localhost:5150` that:
- Serves the README.md as the homepage
- Converts markdown files to HTML
- Applies the Pico CSS framework for styling
- Serves static assets (images, fonts) directly

The server URL is copied to your clipboard automatically.

## Features

- **Markdown rendering**: All `.md` files are automatically converted to HTML
- **Static file serving**: Images (`.png`, `.jpg`, `.gif`, etc.) and fonts are served as binary files
- **Clean styling**: Uses Pico CSS for a minimal, responsive design
- **Performance caching**: Rendered HTML is cached in-memory for instant subsequent requests
- **Development mode**: Disable caching for doc editing with `-config docserver_dev=true`

## Accessing Files

- `/` → `README.md`
- `/docs/learning-duso.md` → `docs/learning-duso.md`
- `/path/to/file.md` → `path/to/file.md`

Files are loaded from the local filesystem first, then from embedded files if not found locally.

## Caching

The docserver automatically caches rendered HTML in-memory for performance:

**Normal mode (caching enabled):**
```bash
duso -docserver
```
- First request to a path: markdown is rendered and stored in cache
- Subsequent requests: served from cache (instant response)
- Cache is cleared on server restart

**Development mode (caching disabled):**
```bash
duso -config "docserver_dev=true" -docserver
```
- All requests render fresh markdown from disk
- Perfect for editing documentation and seeing changes immediately
- No cache means you always see the latest version

### How It Works

The caching uses Duso's thread-safe `datastore()` with a `"docserver"` namespace:
- **Pattern**: Simple key-value caching where the request path is the key
- **Storage**: In-memory only (no persistence to disk)
- **Thread-safe**: Multiple concurrent requests are handled safely
- **Automatic**: No configuration needed - just use dev mode when editing docs

This demonstrates a practical pattern for semi-production web servers: use `datastore` for coordinating state across concurrent request handlers.

## Customization

Edit `stdlib/docserver/docserver.du` to:
- Change the port (default: 5150)
- Modify the CSS styling
- Add custom route handlers
- Change the markdown formatter

## Implementation

The docserver is implemented as a self-referential HTTP handler script that:
1. Sets up routes on startup
2. Handles both static files and markdown rendering
3. Uses the `md-lite` module for markdown parsing
