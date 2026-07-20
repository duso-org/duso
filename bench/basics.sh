#!/bin/bash
# Basic benchmark harness: median of 3 timings + peak RSS (one extra run)
# Usage: basics.sh mac|linux   (run from the bench directory)
OS=$1

run_one() {  # cmd... -> prints "<median_ms> <peak_rss_mb>"
  local vals=()
  for i in 1 2 3; do
    v=$("$@" 2>/dev/null | grep -oE 'in [0-9.]+ms' | grep -oE '[0-9.]+')
    vals+=("$v")
  done
  med=$(printf '%s\n' "${vals[@]}" | sort -n | sed -n 2p)
  if [ "$OS" = "mac" ]; then
    rss=$(/usr/bin/time -l "$@" 2>&1 >/dev/null | awk '/maximum resident/ {printf "%.1f", $1/1048576}')
  else
    rss=$(/usr/bin/time -v "$@" 2>&1 >/dev/null | awk '/Maximum resident/ {printf "%.1f", $NF/1024}')
  fi
  echo "$med $rss"
}

DUSO=${DUSO:-duso}

for t in fib loop sort; do
  echo "== $t =="
  echo "duso   $(run_one $DUSO $t.du)"
  echo "python $(run_one python3 $t.py)"
  echo "node   $(run_one node $t.js)"
  echo "ruby   $(run_one ruby $t.rb)"
done

echo "== fib_builtin =="
echo "duso   $(run_one $DUSO fib_builtin.du)"

echo "== baseline rss (MB) =="
if [ "$OS" = "mac" ]; then
  T() { /usr/bin/time -l "$@" 2>&1 >/dev/null | awk '/maximum resident/ {printf "%.1f\n", $1/1048576}'; }
else
  T() { /usr/bin/time -v "$@" 2>&1 >/dev/null | awk '/Maximum resident/ {printf "%.1f\n", $NF/1024}'; }
fi
echo "duso   $(T $DUSO -c 'print(1)')"
echo "python $(T python3 -c 'print(1)')"
echo "node   $(T node -e 'console.log(1)')"
echo "ruby   $(T ruby -e 'puts 1')"
