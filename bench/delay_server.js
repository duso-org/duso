#!/usr/bin/env node
// Node benchmark server: GET /delay responds after 1 second,
// GET /ping responds immediately. Stdlib only.
const http = require('http');

http.createServer((req, res) => {
  const respond = () => {
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end('ok');
  };
  if (req.url.startsWith('/delay')) {
    setTimeout(respond, 1000);
  } else {
    respond();
  }
}).listen(8399, '127.0.0.1', 1024);
