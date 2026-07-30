#!/bin/bash
# Tests the CLI's argument parsing (getPositionalArgs/parseCliFlags in cmd/duso/main.go),
# specifically the "subcommand file(s) arg(s), then flags" ordering rule. Can't be
# tested from inside a .du script since it's about how argv gets sliced up before
# any script starts running - needs to invoke the duso binary itself repeatedly.
set -u
cd "$(dirname "$0")/.."

DUSO=${DUSO:-./bin/duso}
pass=0
fail=0

check() {
  local desc="$1" expected_exit="$2"; shift 2
  actual_exit=0
  "$@" >/tmp/test_cli_args_out 2>&1 || actual_exit=$?
  if [ "$actual_exit" -eq "$expected_exit" ]; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    echo "✗ $desc: expected exit $expected_exit, got $actual_exit"
    cat /tmp/test_cli_args_out
  fi
}

# --- lint: files must come before flags ---
check "lint: file then flag (clean)" 0 "$DUSO" lint docs/reference/schedule.md -ignore-warnings
check "lint: two files then flag (clean)" 0 "$DUSO" lint docs/reference/schedule.md docs/reference/unschedule.md -ignore-warnings
check "lint: no flag at all (clean, just warnings)" 0 "$DUSO" lint docs/reference/schedule.md
check "lint: flag before file is NOT supported (errors, doesn't silently eat the file)" 1 "$DUSO" lint -ignore-warnings docs/reference/schedule.md

# --- eval: code arg with a trailing flag ---
check "eval: code, no flags" 0 "$DUSO" eval 'print(1+1)'
check "eval: code then -no-color" 0 "$DUSO" eval 'print("hi")' -no-color

# --- doc: positional topic ---
check "doc: positional topic" 0 "$DUSO" doc len

# --- extract: two positional args ---
rm -rf /tmp/test_cli_extract_dst
check "extract: source and dest" 0 "$DUSO" extract stdlib /tmp/test_cli_extract_dst
rm -rf /tmp/test_cli_extract_dst

# --- init: positional project name still reaches the interactive prompt ---
rm -rf /tmp/test_cli_init_proj
init_out=$("$DUSO" init /tmp/test_cli_init_proj < /dev/null 2>&1)
if echo "$init_out" | grep -q "Select a template"; then
  pass=$((pass + 1))
else
  fail=$((fail + 1))
  echo "✗ init: positional project name didn't reach the template prompt"
  echo "$init_out"
fi
rm -rf /tmp/test_cli_init_proj

# --- plain script execution, with and without a trailing flag ---
check "script exec: no flags" 0 "$DUSO" test/test_schedule_spec.du
check "script exec: script then -no-color" 0 "$DUSO" test/test_schedule_spec.du -no-color

# --- -config still reaches the script via sys("-config") ---
echo 'if sys("-config").port != 8080 then throw("config not passed through") end' > /tmp/test_cli_config.du
check "-config reaches script via sys(\"-config\")" 0 "$DUSO" /tmp/test_cli_config.du -config 'port=8080'
rm -f /tmp/test_cli_config.du

# --- arbitrary, duso-doesn't-know-about-them script flags reach sys() too -
# this is the actual "natural CLI script" case, not just duso's own known flags ---
cat > /tmp/test_cli_custom_flags.du << 'DUEOF'
if sys("-verbose") != true then throw("boolean custom flag not passed through") end
if sys("-output") != "report.txt" then throw("string-value custom flag not passed through") end
if sys("-count") != 5 then throw("numeric-value custom flag not passed through") end
DUEOF
check "arbitrary custom flags reach sys()" 0 "$DUSO" /tmp/test_cli_custom_flags.du -verbose -output report.txt -count 5
rm -f /tmp/test_cli_custom_flags.du

# --- double-dash flags normalize to the same single-dash sys() key as single-dash ---
cat > /tmp/test_cli_dash_norm.du << 'DUEOF'
if sys("-verbose") != true then throw("--verbose should be readable as sys(\"-verbose\")") end
if sys("-output") != "report.txt" then throw("--output should be readable as sys(\"-output\")") end
DUEOF
check "double-dash flags normalize to single-dash sys() keys" 0 "$DUSO" /tmp/test_cli_dash_norm.du --verbose --output report.txt
rm -f /tmp/test_cli_dash_norm.du

rm -f /tmp/test_cli_args_out

echo ""
echo "CLI arg tests: $pass passed, $fail failed"
if [ "$fail" -gt 0 ]; then
  exit 1
fi
echo "✓ test_cli_args passed"
