#!/usr/bin/env python3
"""Python I/O-bound benchmark with memory tracking"""

import time
import requests
import tracemalloc
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

    # Start memory tracking
    tracemalloc.start()

    start = time.time() * 1000
    peak_memory = 0

    with ThreadPoolExecutor(max_workers=100) as executor:
        spawn_start = time.time() * 1000
        futures = [executor.submit(worker, i) for i in range(1, 101)]
        spawn_time = time.time() * 1000 - spawn_start

        wait_start = time.time() * 1000

        # Track memory while waiting
        for future in as_completed(futures):
            current, peak = tracemalloc.get_traced_memory()
            peak_memory = max(peak_memory, peak)
            future.result()

        wait_time = time.time() * 1000 - wait_start

    total_time = time.time() * 1000 - start
    current, peak = tracemalloc.get_traced_memory()
    peak_memory = max(peak_memory, peak)
    tracemalloc.stop()

    print(f"Spawn 100 workers: {spawn_time:.1f}ms")
    print(f"Wait for completion: {wait_time:.1f}ms")
    print(f"Total time: {total_time:.1f}ms")
    print(f"Peak memory: {peak_memory / 1024 / 1024:.1f} MB")

if __name__ == "__main__":
    main()
