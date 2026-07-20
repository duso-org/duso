#!/usr/bin/env node
// Concurrent fetch benchmark against delay_server.du.
// WORKERS env sets the worker count (default 100), each making 5 requests.
const http = require('http');

const WORKERS = parseInt(process.env.WORKERS || '100');

let errors = 0;

function get(url) {
  return new Promise((resolve, reject) => {
    const req = http.get(url, { timeout: 60000 }, res => {
      res.resume();
      res.on('end', () => res.statusCode === 200 ? resolve() : reject(new Error('' + res.statusCode)));
    });
    req.on('error', reject);
    req.on('timeout', () => { req.destroy(); reject(new Error('timeout')); });
  });
}

async function worker() {
  for (let i = 0; i < 5; i++) {
    try { await get('http://127.0.0.1:8399/delay'); } catch (e) { errors++; }
  }
}

(async () => {
  const start = Date.now();
  await Promise.all(Array.from({ length: WORKERS }, worker));
  console.log(`cfetch ${WORKERS} workers x 5 reqs: ${Date.now() - start}ms, errors: ${errors}`);
})();
