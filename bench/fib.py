#!/usr/bin/env python3
import time

def fib(n):
    if n <= 1:
        return n
    a, b = 0, 1
    for i in range(2, n + 1):
        a, b = b, a + b
    return b

start = time.time() * 1000
result = None
for i in range(10000):
    result = fib(30)
elapsed = time.time() * 1000 - start

print(f"fib(30) iterative x10000 = {result} in {elapsed:.1f}ms")
