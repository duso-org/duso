#!/usr/bin/env python3
import time

def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

start = time.time() * 1000
result = fib(30)
elapsed = time.time() * 1000 - start

print(f"fib(30) = {result} in {elapsed:.1f}ms")
