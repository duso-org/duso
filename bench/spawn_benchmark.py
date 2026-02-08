#!/usr/bin/env python3
"""Python spawn benchmark - multiprocessing 1000 workers
Each worker counts to 1M
Run: python3 spawn_benchmark.py
"""

import time
import multiprocessing

def worker(worker_id):
    """Worker function that counts to 1M"""
    sum = 0
    for i in range(1, 1000001):
        sum = sum + i % 10
    return {"worker_id": worker_id, "sum": sum}

def main():
    print("=== Python Spawn Benchmark ===")

    start = time.time() * 1000  # milliseconds

    # Spawn 1000 worker processes
    with multiprocessing.Pool(processes=None) as pool:
        results = pool.starmap(worker, [(i,) for i in range(1, 1001)])

    spawn_time = time.time() * 1000 - start
    print(f"Spawn and complete 1000 workers: {spawn_time:.1f}ms")
    print(f"Completed {len(results)} workers")

if __name__ == "__main__":
    main()
