#!/bin/bash
# Monitor memory usage of a process

pid=$1
name=$2

echo "Monitoring memory for: $name (PID: $pid)"

peak_mem=0
while kill -0 $pid 2>/dev/null; do
  mem=$(ps -p $pid -o rss= 2>/dev/null)
  if [ ! -z "$mem" ]; then
    mem_mb=$((mem / 1024))
    if [ $mem_mb -gt $peak_mem ]; then
      peak_mem=$mem_mb
    fi
  fi
  sleep 0.1
done

echo "Peak memory: ${peak_mem} MB"
