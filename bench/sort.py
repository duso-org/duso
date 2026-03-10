#!/usr/bin/env python3
import time
import random

arr = [random.random() * 10000 for _ in range(10000)]

start = time.time() * 1000
sorted_arr = sorted(arr)
elapsed = time.time() * 1000 - start

print(f"Sort 10k random numbers in {elapsed:.1f}ms")
print(f"First 5: {sorted_arr[:5]}")
