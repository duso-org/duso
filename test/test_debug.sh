#!/bin/bash
# Tests `duso debug`: uncaught errors end the script, try/catch still works,
# breakpoints resume in place, and a closed stdin never hangs the REPL.
#
# Needs a shell script to drive it - debug mode is a CLI flag and the REPL reads
# real stdin. The scripts it runs are in test/debug/.
set -u
cd "$(dirname "$0")"

DUSO=${DUSO:-../bin/duso}
OUT=/tmp/test_debug_out.txt
pass=0
fail=0

ok() { pass=$((pass + 1)); }
bad() { fail=$((fail + 1)); echo "✗ $1"; }

# run SCRIPT INPUT - runs under duso debug with INPUT on stdin ("" = closed),
# sets $code and $out. A REPL that never exits is killed after 15s.
run() {
  if [ -z "$2" ]; then
    "$DUSO" debug "debug/$1" </dev/null >"$OUT" 2>&1 &
  else
    printf "$2" | "$DUSO" debug "debug/$1" >"$OUT" 2>&1 &
  fi
  local pid=$!
  for _ in $(seq 1 30); do
    kill -0 $pid 2>/dev/null || break
    sleep 0.5
  done
  if kill -0 $pid 2>/dev/null; then
    kill -9 $pid 2>/dev/null
    wait $pid 2>/dev/null
    code=hung
  else
    wait $pid
    code=$?
  fi
  out=$(sed 's/\x1b\[[0-9;]*m//g' "$OUT")
}

# An uncaught error in a function ends the script, even with stdin closed
run fn_error.du ""
[ "$code" = "1" ] && ok || bad "fn_error.du: expected exit 1, got $code"
echo "$out" | grep -q "SHOULD NOT RUN" && bad "fn_error.du: script kept running after the error" || ok
echo "$out" | grep -q "boom from fail()" && ok || bad "fn_error.du: error message not shown"

run try_catch.du ""
[ "$code" = "0" ] && echo "$out" | grep -q "debug try/catch ok" \
  && ok || bad "try_catch.du (exit $code): $(echo "$out" | grep -v '^debug>' | tail -3)"

run breakpoints.du 'c\nc\nc\nc\n'
[ "$code" = "0" ] && echo "$out" | grep -q "debug breakpoints ok" \
  && ok || bad "breakpoints.du (exit $code): $(echo "$out" | grep -v '^debug>' | tail -3)"
n=$(echo "$out" | grep -c "BREAKPOINT: bp in")
[ "$n" = "4" ] && ok || bad "breakpoints.du: expected 4 breakpoints to open the REPL, saw $n"

# Same script with stdin closed: every breakpoint gives up waiting and resumes
run breakpoints.du ""
[ "$code" = "0" ] && echo "$out" | grep -q "debug breakpoints ok" \
  && ok || bad "breakpoints.du with stdin closed (exit $code)"

run locals.du 'local_x\nc\n'
echo "$out" | grep -q "^debug> 42$" && ok || bad "locals.du: REPL did not see local_x = 42"
[ "$code" = "1" ] && ok || bad "locals.du: expected exit 1 after 'c', got $code"

rm -f "$OUT"

echo ""
if [ $fail -gt 0 ]; then
  echo "test_debug: $pass passed, $fail failed"
  exit 1
fi
echo "✓ test_debug passed ($pass checks)"
