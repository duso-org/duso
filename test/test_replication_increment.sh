#!/bin/bash
# Tests object-delta increment() and lazy expiry across a real leader and
# follower, then that each side's files recover to the same state.
#
# Needs a shell script to drive it - a namespace is one store per process, so a
# leader and a follower of the same store need two duso processes. The checks
# themselves are in the test_replication_increment_*.du scripts.
set -u
cd "$(dirname "$0")"

DUSO=${DUSO:-../bin/duso}
export REPL_DIR=/tmp/test_replication_increment_work
export REPL_PORT=${REPL_PORT:-17991}
pass=0
fail=0

ok() { pass=$((pass + 1)); }
bad() { fail=$((fail + 1)); echo "✗ $1"; }

rm -rf "$REPL_DIR"; mkdir -p "$REPL_DIR"

"$DUSO" test_replication_increment_leader.du > "$REPL_DIR/leader.out" 2>&1 &
leader=$!
sleep 0.5

out=$("$DUSO" test_replication_increment_follower.du 2>&1)
echo "$out" | grep -q '^follower ok$' && ok || bad "follower: $(echo "$out" | grep -v '^datastore' | tail -3)"

wait $leader 2>/dev/null
grep -q '^leader ok$' "$REPL_DIR/leader.out" && ok || bad "leader: $(grep -v '^datastore' "$REPL_DIR/leader.out" | tail -3)"

for side in leader follower; do
  out=$(REPL_SIDE=$side "$DUSO" test_replication_increment_readback.du 2>&1)
  echo "$out" | grep -q "^$side recovered ok$" && ok || bad "$side recovery: $(echo "$out" | tail -3)"
done

rm -rf "$REPL_DIR"

echo ""
if [ $fail -gt 0 ]; then
  echo "test_replication_increment: $pass passed, $fail failed"
  exit 1
fi
echo "✓ test_replication_increment passed ($pass checks)"
