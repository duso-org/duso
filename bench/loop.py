#!/usr/bin/env python3
import time

start = time.time() * 1000

sum = 0
for i in range(1, 1001):
    for j in range(1, 1001):
        sum += i * j

elapsed = time.time() * 1000 - start

print(f"Loop sum (1000x1000) = {sum} in {elapsed:.1f}ms")
