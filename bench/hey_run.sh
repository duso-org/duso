#!/bin/bash
# Run hey against ONE language's server, ONE endpoint, ONE concurrency level.
# Usage: hey_run.sh mac|linux duso|node|python|ruby delay|ping CONCURRENCY
# Appends a CSV row to results.csv in the current directory:
#   machine,lang,endpoint,concurrency,rps,p50_ms,p99_ms,server_rss_mb,non200,total_reqs
OS=$1; L=$2; EP=$3; C=$4
DUSO_BIN=${DUSO:-duso}
HEY=${HEY:-hey}
ulimit -n 16384 2>/dev/null

case $L in
  duso)   SRV=("$DUSO_BIN" server.du) ;;
  node)   SRV=(node delay_server.js) ;;
  python) SRV=(python3 delay_server.py) ;;
  ruby)   SRV=(ruby delay_server.rb) ;;
  *) echo "unknown lang: $L"; exit 1 ;;
esac

SLOG=$(mktemp); HLOG=$(mktemp)
if [ "$OS" = mac ]; then
  /usr/bin/time -l "${SRV[@]}" >/dev/null 2>"$SLOG" &
  WRAP=$!
else
  systemd-run --scope --quiet -p MemoryMax=200M -p MemorySwapMax=0 \
    "${SRV[@]}" >/dev/null 2>"$SLOG" &
  WRAP=$!
fi
sleep 2

if [ "$EP" = delay ]; then
  "$HEY" -c "$C" -n $((C * 5)) -t 60 "http://127.0.0.1:8399/delay" > "$HLOG" 2>&1
else
  "$HEY" -c "$C" -z 10s -t 60 "http://127.0.0.1:8399/ping" > "$HLOG" 2>&1
fi

RPS=$(awk '/Requests\/sec/ {print $2}' "$HLOG")
P50=$(awk '$1 ~ /^50%/ {printf "%.1f", $3 * 1000}' "$HLOG")
P99=$(awk '$1 ~ /^99%/ {printf "%.1f", $3 * 1000}' "$HLOG")
TOTAL=$(awk '/responses/ {s += $2} END {print s}' "$HLOG")
NON200=$(awk '/responses/ {if ($1 != "[200]") s += $2} END {print s + 0}' "$HLOG")

sleep 1
SPID=$(pgrep -P "$WRAP" | head -1)
[ -z "$SPID" ] && SPID=$WRAP
RSS=""
if [ "$OS" = linux ]; then
  if [ -d "/proc/$SPID" ] && grep -q -e delay -e server "/proc/$SPID/cmdline" 2>/dev/null; then
    RSS=$(awk '/VmHWM/ {printf "%.1f", $2/1024}' "/proc/$SPID/status")
    kill -TERM "$SPID" 2>/dev/null
  else
    RSS="killed"
  fi
  kill -TERM "$WRAP" 2>/dev/null
else
  if [ "$L" = duso ]; then kill -INT "$SPID" 2>/dev/null; else kill -TERM "$SPID" 2>/dev/null; fi
  wait "$WRAP" 2>/dev/null
  RSS=$(awk '/maximum resident/ {printf "%.1f", $1/1048576}' "$SLOG")
fi

echo "$OS,$L,$EP,$C,$RPS,$P50,$P99,$RSS,$NON200,$TOTAL" | tee -a results.csv
rm -f "$SLOG" "$HLOG"
