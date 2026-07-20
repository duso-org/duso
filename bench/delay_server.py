#!/usr/bin/env python3
"""Python benchmark server: GET /delay responds after 1 second,
GET /ping responds immediately. Stdlib only."""
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def do_GET(self):
        if self.path.startswith("/delay"):
            time.sleep(1)
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.send_header("Content-Length", "2")
        self.end_headers()
        self.wfile.write(b"ok")

    def log_message(self, *args):
        pass

ThreadingHTTPServer.request_queue_size = 1024
ThreadingHTTPServer(("127.0.0.1", 8399), Handler).serve_forever()
