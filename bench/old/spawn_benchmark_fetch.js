#!/usr/bin/env node
/**
 * Node.js I/O-bound benchmark - spawn 1000 concurrent workers making HTTP requests
 * Each worker makes 5 requests to httpbin.org/delay/1
 * Run: node spawn_benchmark_fetch.js
 */

const http = require('http');
const https = require('https');

function fetchWithTimeout(url, timeout = 15000) {
  return new Promise((resolve, reject) => {
    const protocol = url.startsWith('https') ? https : http;
    const request = protocol.get(url, { timeout }, (response) => {
      let data = '';
      response.on('data', chunk => data += chunk);
      response.on('end', () => resolve({ ok: response.statusCode === 200, status: response.statusCode }));
    });
    request.on('error', reject);
    request.on('timeout', () => { request.destroy(); reject(new Error('timeout')); });
  });
}

async function worker(workerId) {
  const results = [];
  for (let i = 1; i <= 5; i++) {
    try {
      const response = await fetchWithTimeout('https://httpbin.org/delay/1', 15000);
      results.push({ attempt: i, status: response.status });
    } catch (err) {
      results.push({ attempt: i, error: err.message });
    }
  }
  return { worker_id: workerId, requests: results.length };
}

async function main() {
  console.log("=== Node.js I/O-Bound Benchmark ===");
  console.log("Each worker makes 5 HTTP requests to httpbin.org/delay/1");
  console.log("");

  const start = Date.now();

  // Spawn 500 workers
  const spawnStart = Date.now();
  const promises = [];
  for (let i = 1; i <= 500; i++) {
    promises.push(worker(i));
  }
  const spawnTime = Date.now() - spawnStart;

  // Wait for all to complete
  const waitStart = Date.now();
  const results = await Promise.all(promises);
  const waitTime = Date.now() - waitStart;

  const totalTime = Date.now() - start;

  console.log(`Spawn 500 workers: ${spawnTime}ms`);
  console.log(`Wait for completion: ${waitTime}ms`);
  console.log(`Total time: ${totalTime}ms`);
  console.log(`Average per worker: ${(waitTime/500).toFixed(3)}ms`);
}

main().catch(console.error);
