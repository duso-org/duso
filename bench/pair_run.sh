#!/bin/bash
# Run ONE language's delay server + client pair at ONE worker level.
# Usage: pair_run.sh mac|linux duso|node|python|ruby WORKERS
# Prints: client wall time, client peak RSS, server peak RSS.
# Appends a CSV row to cfetch_results.csv in the current directory:
#   machine,lang,workers,wall_ms,client_rss_mb,server_rss_mb,errors
OS=$1; L=$2; N=$3
DUSO_BIN=${DUSO:-duso}
ulimit -n 16384 2>/dev/null

case $L in
  duso)   SRV=("$DUSO_BIN" delay_server.du); CLI=("$DUSO_BIN" cfetch.du) ;;
  node)   SRV=(node delay_server.js);        CLI=(node cfetch.js) ;;
  python) SRV=(python3 delay_server.py);     CLI=(python3 cfetch.py) ;;
  ruby)   SRV=(ruby delay_server.rb);        CLI=(ruby cfetch.rb) ;;
  *) echo "unknown lang: $L"; exit 1 ;;
esac

SLOG=$(mktemp)
if [ "$OS" = mac ]; then
  /usr/bin/time -l "${SRV[@]}" >/dev/null 2>"$SLOG" &
  WRAP=$!
else
  systemd-run --scope --quiet -p MemoryMax=200M -p MemorySwapMax=0 \
    "${SRV[@]}" >/dev/null 2>"$SLOG" &
  WRAP=$!
fi
sleep 2

CLOG=$(mktemp)
if [ "$OS" = mac ]; then
  WORKERS=$N /usr/bin/time -l "${CLI[@]}" > "$CLOG" 2>&1
  CRSS=$(awk '/maximum resident/ {printf "%.1f", $1/1048576}' "$CLOG")
else
  WORKERS=$N systemd-run --scope --quiet -p MemoryMax=180M -p MemorySwapMax=0 \
    /usr/bin/time -v "${CLI[@]}" > "$CLOG" 2>&1
  CRSS=$(awk '/Maximum resident/ {printf "%.1f", $NF/1024}' "$CLOG")
fi
grep -E 'cfetch|illed|oom' "$CLOG"
echo "client rss: $CRSS MB"
WALL=$(grep -oE '[0-9.]+ms' "$CLOG" | head -1 | tr -d ms)
ERRS=$(grep -oE 'errors: [0-9]+' "$CLOG" | awk '{print $2}')
rm -f "$CLOG"

sleep 1
SPID=$(pgrep -P "$WRAP" | head -1)
[ -z "$SPID" ] && SPID=$WRAP
if [ "$OS" = linux ]; then
  if [ -d "/proc/$SPID" ] && grep -q delay "/proc/$SPID/cmdline" 2>/dev/null; then
    SRSS=$(awk '/VmHWM/ {printf "%.1f", $2/1024}' "/proc/$SPID/status")
    kill -TERM "$SPID" 2>/dev/null
  else
    SRSS="killed"
  fi
  kill -TERM "$WRAP" 2>/dev/null
else
  if [ "$L" = duso ]; then kill -INT "$SPID" 2>/dev/null; else kill -TERM "$SPID" 2>/dev/null; fi
  wait "$WRAP" 2>/dev/null
  SRSS=$(awk '/maximum resident/ {printf "%.1f", $1/1048576}' "$SLOG")
fi
echo "server peak rss: $SRSS MB"
echo "$OS,$L,$N,$WALL,$CRSS,$SRSS,$ERRS" >> cfetch_results.csv
rm -f "$SLOG"
