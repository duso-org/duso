#!/usr/bin/env python3
# JSON round-trip - json module (stdlib)
import json
import time

n = 5000
objs = []
for i in range(1, n + 1):
    objs.append({"id": i, "name": f"user-{i}", "active": i % 3 == 0, "score": i * 1.5, "tags": ["a", "b", "c"]})

start = time.perf_counter()
text = json.dumps(objs)
encoded = (time.perf_counter() - start) * 1000

start = time.perf_counter()
back = json.loads(text)
decoded = (time.perf_counter() - start) * 1000

print(f"JSON encode {n} objects in {encoded}ms, decode in {decoded}ms")
