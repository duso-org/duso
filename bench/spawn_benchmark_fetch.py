#!/usr/bin/env python3
"""Python I/O-bound benchmark - spawn 1000 concurrent workers making HTTP requests
Each worker makes 5 requests to httpbin.org/delay/1
Run: python3 spawn_benchmark_fetch.py
"""

import time
import requests
from concurrent.futures import ThreadPoolExecutor, as_completed

def worker(worker_id):
    """Worker that makes 5 HTTP requests"""
    results = []
    for i in range(1, 6):
        try:
            # Request with 1-second delay
            response = requests.get("https://httpbin.org/delay/1", timeout=15)
            if response.status_code == 200:
                results.append({"attempt": i, "status": response.status_code})
        except Exception as e:
            results.append({"attempt": i, "error": str(e)})
    return {"worker_id": worker_id, "requests": len(results)}

def main():
    print("=== Python I/O-Bound Benchmark ===")
    print("Each worker makes 5 HTTP requests to httpbin.org/delay/1")
    print("")

    start = time.time() * 1000  # milliseconds

    # Spawn 1000 workers using thread pool
    spawn_start = time.time() * 1000
    with ThreadPoolExecutor(max_workers=1000) as executor:
        # Submit all tasks
        futures = [executor.submit(worker, i) for i in range(1, 1001)]

        spawn_time = time.time() * 1000 - spawn_start

        # Wait for all to complete
        wait_start = time.time() * 1000
        results = [f.result() for f in as_completed(futures)]
        wait_time = time.time() * 1000 - wait_start

    total_time = time.time() * 1000 - start

    print(f"Spawn 1000 workers: {spawn_time:.1f}ms")
    print(f"Wait for completion: {wait_time:.1f}ms")
    print(f"Total time: {total_time:.1f}ms")
    print(f"Average per worker: {wait_time/1000:.3f}ms")

if __name__ == "__main__":
    main()
