#!/usr/bin/env python3
"""Python benchmark suite - timing in milliseconds
Run: python3 benchmark.py
"""

import time

def benchmark(name, fn):
    start = time.time() * 1000  # Convert to milliseconds
    result = fn()
    elapsed = time.time() * 1000 - start
    print(f"{name}: {elapsed:.1f}ms")
    return result

# Benchmark 1: Arithmetic operations
def test_arithmetic():
    sum = 0
    for i in range(1, 1000001):
        sum = sum + i % 10
    return sum

# Benchmark 2: Array push (tight loop)
def test_array_push():
    arr = []
    for i in range(1, 100001):
        arr.append(i)
    return len(arr)

# Benchmark 3: Nested loops
def test_nested_loops():
    sum = 0
    for i in range(1, 101):
        for j in range(1, 101):
            sum = sum + i * j
    return sum

# Benchmark 4: String operations
def test_string_concat():
    str = ""
    for i in range(1, 10001):
        str = str + "x"
    return len(str)

# Benchmark 5: Function calls (recursion)
def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

def test_recursion():
    return fib(25)

# Benchmark 6: Array filtering and mapping
def test_functional():
    nums = list(range(1, 10001))
    filtered = [x for x in nums if x % 2 == 0]
    mapped = [x * 2 for x in filtered]
    return len(mapped)

# Run all benchmarks
print("=== Python Benchmarks ===")
benchmark("Arithmetic (1M ops)", test_arithmetic)
benchmark("Array push (100k)", test_array_push)
benchmark("Nested loops (100x100)", test_nested_loops)
benchmark("String concat (10k)", test_string_concat)
benchmark("Recursion fib(25)", test_recursion)
benchmark("Functional chain", test_functional)
