#!/usr/bin/env python3
"""Python I/O-bound benchmark lite - 100 workers (not 1000)"""

import time
import requests
from concurrent.futures import ThreadPoolExecutor, as_completed

def worker(worker_id):
    """Worker that makes 5 HTTP requests"""
    results = []
    for i in range(1, 6):
        try:
            response = requests.get("https://httpbin.org/delay/1", timeout=15)
            if response.status_code == 200:
                results.append({"attempt": i, "status": response.status_code})
        except Exception as e:
            results.append({"attempt": i, "error": str(e)})
    return {"worker_id": worker_id, "requests": len(results)}

def main():
    print("=== Python I/O-Bound Benchmark (100 workers) ===")
    start = time.time() * 1000

    with ThreadPoolExecutor(max_workers=100) as executor:
        spawn_start = time.time() * 1000
        futures = [executor.submit(worker, i) for i in range(1, 101)]
        spawn_time = time.time() * 1000 - spawn_start

        wait_start = time.time() * 1000
        results = [f.result() for f in as_completed(futures)]
        wait_time = time.time() * 1000 - wait_start

    total_time = time.time() * 1000 - start

    print(f"Spawn 100 workers: {spawn_time:.1f}ms")
    print(f"Wait for completion: {wait_time:.1f}ms")
    print(f"Total time: {total_time:.1f}ms")

if __name__ == "__main__":
    main()
