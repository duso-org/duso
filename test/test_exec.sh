#!/bin/bash
# Tests the -allow-exec flag itself: whether duso starts, what it warns about,
# and what exec() can reach. Can't live inside a .du script, since the answers
# are decided before any script runs - it needs to invoke the binary repeatedly.
#
# Also drives test/test_exec.du, which is the exec() behavior suite and needs a
# specific allowlist to run against.
set -u
cd "$(dirname "$0")/.."

DUSO=${DUSO:-./bin/duso}
OUT=/tmp/test_exec_out
pass=0
fail=0

# check DESC EXPECTED_EXIT -- cmd...
check() {
  local desc="$1" expected_exit="$2"; shift 2
  local actual_exit=0
  "$@" >"$OUT" 2>&1 || actual_exit=$?
  if [ "$actual_exit" -eq "$expected_exit" ]; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    echo "✗ $desc: expected exit $expected_exit, got $actual_exit"
    cat "$OUT"
  fi
}

# expect_output DESC PATTERN -- cmd...  (runs the command, greps its output)
expect_output() {
  local desc="$1" pattern="$2"; shift 2
  "$@" >"$OUT" 2>&1
  if grep -q -- "$pattern" "$OUT"; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    echo "✗ $desc: output did not match '$pattern'"
    cat "$OUT"
  fi
}

# --- the behavior suite ---

DUSO_TEST_SECRET=shhh \
  check "exec() behavior suite" 0 \
  "$DUSO" test/test_exec.du -allow-exec "echo,false,cat,sleep,date -u,sh -c"

# --- exec() is off unless asked for ---

echo 'exec("echo", {args = ["x"]})' > /tmp/test_exec_denied.du
check "no flag: exec() throws" 1 "$DUSO" /tmp/test_exec_denied.du
expect_output "no flag: error names the flag" "allow-exec" "$DUSO" /tmp/test_exec_denied.du

# --- boot-time resolution ---

check "unresolvable command refuses to boot" 1 \
  "$DUSO" /tmp/test_exec_denied.du -allow-exec "echo,definitely_not_a_real_binary"
expect_output "and says which one" "definitely_not_a_real_binary not found" \
  "$DUSO" /tmp/test_exec_denied.du -allow-exec "echo,definitely_not_a_real_binary"

check "resolvable list boots" 0 "$DUSO" /tmp/test_exec_denied.du -allow-exec "echo"
check "spaces around entries are trimmed" 0 "$DUSO" /tmp/test_exec_denied.du -allow-exec " echo , cat "
check "empty entries are skipped" 0 "$DUSO" /tmp/test_exec_denied.du -allow-exec "echo,,cat,"

# --- warnings ---

expect_output "allowing a shell warns" "a shell can run anything" \
  "$DUSO" /tmp/test_exec_denied.du -allow-exec "echo,sh"

echo 'r = exec("date", {args = ["+%s"]}) print(r.ok)' > /tmp/test_exec_any.du
expect_output "bare flag warns loudly" "no command list" \
  "$DUSO" /tmp/test_exec_any.du -allow-exec
check "bare flag allows anything" 0 "$DUSO" /tmp/test_exec_any.du -allow-exec
expect_output "all-comma list is the same as bare" "no command list" \
  "$DUSO" /tmp/test_exec_any.du -allow-exec ",,,"

# --- the call is logged, the environment never is ---

echo 'exec("sh", {args = ["-c", "true"], env = {MY_SECRET = "hunter2"}})' > /tmp/test_exec_log.du
expect_output "argv is logged" "exec: .*sh -c true" \
  "$DUSO" /tmp/test_exec_log.du -allow-exec "sh"
"$DUSO" /tmp/test_exec_log.du -allow-exec "sh" >"$OUT" 2>&1
if grep -q "hunter2" "$OUT"; then
  fail=$((fail + 1))
  echo "✗ environment leaked into the log"
else
  pass=$((pass + 1))
fi

# --- cleanup reaches the whole process tree ---
# Both of these use odd sleep durations as markers so pgrep can't hit anything
# else on the machine. The grandchild is the interesting one: signalling only
# the child we started would leave the backgrounded sleep running.

no_survivors() {
  local desc="$1" pattern="$2"
  sleep 3   # SIGTERM, then the 2s grace, then SIGKILL
  if pgrep -f "$pattern" >/dev/null; then
    fail=$((fail + 1))
    echo "✗ $desc: '$pattern' still running"
    pkill -f "$pattern"
  else
    pass=$((pass + 1))
  fi
}

cat > /tmp/test_exec_group.du <<'EOF'
try
  exec("sh", {args = ["-c", "sleep 47 & sleep 47"], timeout = 0.5})
catch (e)
end
EOF
"$DUSO" /tmp/test_exec_group.du -allow-exec "sh -c" >"$OUT" 2>&1
no_survivors "timeout kills the whole process group" "sleep 47"

# kill(pid) on a spawned instance has to tear down its child too, or a killed
# worker leaves a process behind every time.
cat > /tmp/test_exec_killed.du <<'EOF'
try
  exec("sleep", {args = ["53"]})
catch (e)
  print("child: " + e)
end
EOF
cat > /tmp/test_exec_killer.du <<'EOF'
pid = spawn("/tmp/test_exec_killed.du", {})
sleep(1)
kill(pid)
sleep(1)
EOF
"$DUSO" /tmp/test_exec_killer.du -allow-exec "sleep" >"$OUT" 2>&1
if grep -q "child: .*cancelled" "$OUT"; then
  pass=$((pass + 1))
else
  fail=$((fail + 1))
  echo "✗ kill() surfaces as a cancellation, not a timeout"
  cat "$OUT"
fi
no_survivors "kill(pid) tears down the running child" "sleep 53"

rm -f /tmp/test_exec_denied.du /tmp/test_exec_any.du /tmp/test_exec_log.du \
      /tmp/test_exec_group.du /tmp/test_exec_killed.du /tmp/test_exec_killer.du "$OUT"

echo ""
echo "Test Results: $pass passed, $fail failed"
[ "$fail" -eq 0 ] || exit 1
