#!/usr/bin/env python3
# Regex - re module (stdlib), single realistic ops on a real doc
import re
import time

with open('../docs/learning-duso.md', 'rb') as f:
    doc_bytes = f.read()
doc = doc_bytes.decode('utf8')

start = time.perf_counter()
has_datastore = re.search(r'datastore\(', doc) is not None
contains_ms = (time.perf_counter() - start) * 1000

start = time.perf_counter()
words = re.findall(r'\w+', doc)
find_ms = (time.perf_counter() - start) * 1000

print(f"contains() on {len(doc_bytes)}-byte doc in {contains_ms}ms (found={has_datastore})")
print(f"find() on {len(doc_bytes)}-byte doc in {find_ms}ms ({len(words)} matches)")
