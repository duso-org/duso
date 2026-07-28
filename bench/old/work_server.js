#!/usr/bin/env node
// Node benchmark server: GET /work does the same per-request computation as
// work.du (20 calls of score(50), sum of squares), GET /ping responds
// immediately. Stdlib only.
const http = require('http');

function score(n) {
  let s = 0;
  for (let i = 1; i <= n; i++) {
    s = s + i * i;
  }
  return s;
}

http.createServer((req, res) => {
  if (req.url.startsWith('/work')) {
    let total = 0;
    for (let j = 1; j <= 20; j++) {
      total = total + score(50);
    }
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end(String(total));
  } else {
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end('ok');
  }
}).listen(8399, '127.0.0.1', 1024);
