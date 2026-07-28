#!/usr/bin/env python3
"""Concurrent fetch benchmark against delay_server.du.
WORKERS env sets the worker count (default 100), each making 5 requests."""
import os
import time
import requests
from concurrent.futures import ThreadPoolExecutor

WORKERS = int(os.environ.get("WORKERS", "100"))

errors = 0

def worker(_):
    global errors
    for _ in range(5):
        try:
            r = requests.get("http://127.0.0.1:8399/delay", timeout=60)
            if r.status_code != 200:
                errors += 1
        except Exception:
            errors += 1

start = time.time()
with ThreadPoolExecutor(max_workers=WORKERS) as ex:
    list(ex.map(worker, range(WORKERS)))
total = (time.time() - start) * 1000

print(f"cfetch {WORKERS} workers x 5 reqs: {total:.0f}ms, errors: {errors}")
