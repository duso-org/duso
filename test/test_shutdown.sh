#!/bin/bash
# Tests process shutdown: datastores flush and HTTP requests drain, on both the
# signal path and ordinary script completion.
#
# Has to live outside a .du script - it needs to send real signals to the duso
# binary and inspect what reached disk after the process is gone.
set -u
cd "$(dirname "$0")/.."

DUSO=${DUSO:-./bin/duso}
WORK=/tmp/test_shutdown_work
pass=0
fail=0

ok() { pass=$((pass + 1)); }
bad() { fail=$((fail + 1)); echo "✗ $1"; }

rm -rf "$WORK"; mkdir -p "$WORK"

# --- a datastore with persist and no persist_interval ---
# The snapshot ticker only runs when an interval is set, so shutdown is the only
# thing that ever writes this file. It used to write nothing at all.

cat > "$WORK/persist_only.du" <<'EOF'
store = datastore("shutdown_persist", {persist = "/tmp/test_shutdown_work/p.dusnap"})
store.set("k", "value")
EOF

"$DUSO" "$WORK/persist_only.du" >/dev/null 2>&1
[ -f "$WORK/p.dusnap" ] && ok || bad "normal script exit: snapshot never written"

cat > "$WORK/read.du" <<'EOF'
store = datastore("shutdown_persist", {persist = "/tmp/test_shutdown_work/p.dusnap"})
print(tostring(store.get("k")))
EOF
[ "$("$DUSO" "$WORK/read.du" 2>/dev/null | tail -1)" = "value" ] \
  && ok || bad "normal script exit: value did not survive"

# Same store, but the process is signalled while it sits there
rm -f "$WORK/p.dusnap"
cat > "$WORK/persist_wait.du" <<'EOF'
store = datastore("shutdown_persist", {persist = "/tmp/test_shutdown_work/p.dusnap"})
store.set("k", "value")
print("ready")
sleep(60)
EOF

"$DUSO" "$WORK/persist_wait.du" >/dev/null 2>&1 &
pid=$!
sleep 2
kill -TERM $pid
wait $pid 2>/dev/null; code=$?

[ "$code" -eq 0 ] && ok || bad "SIGTERM: expected exit 0, got $code"
[ -f "$WORK/p.dusnap" ] && ok || bad "SIGTERM: snapshot never written"
[ "$("$DUSO" "$WORK/read.du" 2>/dev/null | tail -1)" = "value" ] \
  && ok || bad "SIGTERM: value did not survive"

# --- a script with nothing to drain still exits promptly ---
# The drain has a 30s cap; a plain sleeping script must not pay it.

echo 'sleep(60)' > "$WORK/sleeper.du"
"$DUSO" "$WORK/sleeper.du" >/dev/null 2>&1 &
pid=$!
sleep 1
start=$(date +%s)
kill -TERM $pid
wait $pid 2>/dev/null
elapsed=$(( $(date +%s) - start ))
[ "$elapsed" -lt 5 ] && ok || bad "SIGTERM on a plain script took ${elapsed}s"

# --- an in-flight HTTP request finishes before the process exits ---
# This is the whole point of draining: the client gets its response instead of a
# cut connection, so a non-idempotent handler is never left ambiguous.

cat > "$WORK/slow_handler.du" <<'EOF'
ctx = context()
res = ctx.response()
sleep(3)
res.text("finished")
EOF
cat > "$WORK/server.du" <<'EOF'
server = http_server({port = 8096})
server.route("GET", "/slow", "slow_handler.du")
print("serving")
server.start()
EOF

"$DUSO" "$WORK/server.du" >/dev/null 2>&1 &
pid=$!
sleep 2
curl -s -m 20 http://127.0.0.1:8096/slow > "$WORK/response.txt" 2>&1 &
sleep 1
kill -TERM $pid
wait $pid 2>/dev/null; code=$?
wait 2>/dev/null

[ "$(cat "$WORK/response.txt")" = "finished" ] \
  && ok || bad "in-flight request was cut: [$(cat "$WORK/response.txt")]"
[ "$code" -eq 0 ] && ok || bad "server SIGTERM: expected exit 0, got $code"

# --- a datastore written by a handler mid-drain still lands ---
# Datastores flush after the drain, so a write from a handler that was still
# running when the signal arrived has to be in the snapshot.

cat > "$WORK/write_handler.du" <<'EOF'
ctx = context()
res = ctx.response()
store = datastore("shutdown_handler")   // configured by the server script
sleep(3)
store.set("written_during_drain", "yes")
res.text("ok")
EOF
cat > "$WORK/server2.du" <<'EOF'
store = datastore("shutdown_handler", {persist = "/tmp/test_shutdown_work/h.dusnap"})
server = http_server({port = 8095})
server.route("GET", "/write", "write_handler.du")
print("serving")
server.start()
EOF

"$DUSO" "$WORK/server2.du" >/dev/null 2>&1 &
pid=$!
sleep 2
curl -s -m 20 http://127.0.0.1:8095/write >/dev/null 2>&1 &
sleep 1
kill -TERM $pid
wait $pid 2>/dev/null
wait 2>/dev/null

cat > "$WORK/read2.du" <<'EOF'
store = datastore("shutdown_handler", {persist = "/tmp/test_shutdown_work/h.dusnap"})
print(tostring(store.get("written_during_drain")))
EOF
[ "$("$DUSO" "$WORK/read2.du" 2>/dev/null | tail -1)" = "yes" ] \
  && ok || bad "handler write during drain did not reach the snapshot"

# --- shutdown() from a plain script ---

cat > "$WORK/shutdown_script.du" <<'EOF'
store = datastore("shutdown_builtin", {persist = "/tmp/test_shutdown_work/s.dusnap"})
store.set("k", "value")
shutdown(3)
print("THIS MUST NOT PRINT")
EOF

out=$("$DUSO" "$WORK/shutdown_script.du" 2>&1); code=$?
[ "$code" -eq 3 ] && ok || bad "shutdown(3): expected exit 3, got $code"
case "$out" in
  *"MUST NOT PRINT"*) bad "shutdown() let the next statement run" ;;
  *) ok ;;
esac
[ -f "$WORK/s.dusnap" ] && ok || bad "shutdown(): datastore not flushed"

# A bad code is a mistake worth naming, not something to truncate into a status
# a service manager will misread.
"$DUSO" eval 'shutdown(1.5)' >/dev/null 2>&1 && bad "shutdown(1.5) was accepted" || ok

# --- shutdown() from a handler, while the main script is parked in start() ---
# This is the case that has no other answer: exit() would end only the handler,
# and the script blocked in start() cannot be reached any other way.

cat > "$WORK/stop_handler.du" <<'EOF'
ctx = context()
res = ctx.response()
shutdown(0)
EOF
cat > "$WORK/drain_handler.du" <<'EOF'
ctx = context()
res = ctx.response()
store = datastore("shutdown_from_handler")
sleep(3)
store.set("finished_during_drain", "yes")
res.text("slow done")
EOF
cat > "$WORK/server3.du" <<'EOF'
store = datastore("shutdown_from_handler", {persist = "/tmp/test_shutdown_work/f.dusnap"})
server = http_server({port = 8092})
server.route("GET", "/slow", "drain_handler.du")
server.route("GET", "/stop", "stop_handler.du")
print("serving")
server.start()
EOF

"$DUSO" "$WORK/server3.du" >/dev/null 2>&1 &
pid=$!
sleep 2
curl -s -m 20 http://127.0.0.1:8092/slow > "$WORK/slow.txt" 2>&1 &
sleep 1
start=$(date +%s)
status=$(curl -s -o /dev/null -w "%{http_code}" -m 20 http://127.0.0.1:8092/stop)
wait $pid 2>/dev/null; code=$?
elapsed=$(( $(date +%s) - start ))
wait 2>/dev/null

[ "$code" -eq 0 ] && ok || bad "shutdown() from handler: expected exit 0, got $code"
[ "$status" = "204" ] && ok || bad "shutdown() from handler: expected 204, got $status"
# Blocking in the drain would stall for the full 30s drainTimeout, because the
# calling handler is itself a connection the drain is waiting on.
[ "$elapsed" -lt 10 ] && ok || bad "shutdown() from handler stalled ${elapsed}s"
[ "$(cat "$WORK/slow.txt")" = "slow done" ] \
  && ok || bad "shutdown() cut off a concurrent request: [$(cat "$WORK/slow.txt")]"

cat > "$WORK/read3.du" <<'EOF'
store = datastore("shutdown_from_handler", {persist = "/tmp/test_shutdown_work/f.dusnap"})
print(tostring(store.get("finished_during_drain")))
EOF
[ "$("$DUSO" "$WORK/read3.du" 2>/dev/null | tail -1)" = "yes" ] \
  && ok || bad "shutdown(): write from a draining handler was lost"

rm -rf "$WORK"

echo ""
echo "Test Results: $pass passed, $fail failed"
[ "$fail" -eq 0 ] || exit 1
