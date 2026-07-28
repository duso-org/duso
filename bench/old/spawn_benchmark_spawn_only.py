#!/usr/bin/env python3
"""Python spawn benchmark - measure spawn time only (like Duso)
This measures how long it takes to submit 1000 processes,
not how long it takes for them to complete.
Run: python3 spawn_benchmark_spawn_only.py
"""

import time
import multiprocessing
from threading import Thread

def worker(worker_id):
    """Worker function that counts to 1M"""
    sum = 0
    for i in range(1, 1000001):
        sum = sum + i % 10
    return {"worker_id": worker_id, "sum": sum}

def main():
    print("=== Python Spawn Benchmark (spawn time only) ===")

    # Start workers in a thread pool so we can measure submit time
    start = time.time() * 1000  # milliseconds

    with multiprocessing.Pool(processes=None) as pool:
        # Measure time to submit all tasks (async_apply)
        async_results = []
        for i in range(1, 1001):
            async_results.append(pool.apply_async(worker, (i,)))

    spawn_time = time.time() * 1000 - start
    print(f"Submit 1000 worker tasks: {spawn_time:.1f}ms")
    print(f"Submitted {len(async_results)} tasks")

if __name__ == "__main__":
    main()
