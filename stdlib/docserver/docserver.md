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
- **Hot reload**: Reads files from disk on each request (no caching)

## Accessing Files

- `/` → `README.md`
- `/docs/learning_duso.md` → `docs/learning_duso.md`
- `/path/to/file.md` → `path/to/file.md`

Files are loaded from the local filesystem first, then from embedded files if not found locally.

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
