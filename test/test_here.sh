#!/bin/bash
# Acid test for /HERE/: launches test/here/main.du from /tmp, so the working
# directory, the app directory (test/here), and the module's directory
# (test/here/lib) are three different places. /HERE/ only has a distinguishable
# meaning when they differ - run from its own directory, every rule agrees.
#
# test/here/ holds decoy files named like the module's own, so a /HERE/ that
# leaks to appDir reads "APPDIR-..." and fails rather than passing by accident.
set -u
cd "$(dirname "$0")/.."
REPO=$PWD
DUSO=${DUSO:-$REPO/bin/duso}

echo "CWD-MARKER" > /tmp/here_cwd_marker.txt

# Launched from /tmp, by absolute path, so nothing about the invocation hints
# at either directory the script cares about.
(cd /tmp && "$DUSO" "$REPO/test/here/main.du")
code=$?

# Again under `duso debug`, which runs the entry script through a different
# path and used to resolve /HERE/ in modules to the entry script's directory.
# A failure there opens the debug REPL, which never exits on a closed stdin,
# so it gets a watchdog and its output is filtered down to the error.
out=/tmp/here_debug_out.txt
(cd /tmp && exec "$DUSO" debug "$REPO/test/here/main.du" </dev/null >"$out" 2>&1) &
pid=$!
for _ in $(seq 1 60); do
  kill -0 $pid 2>/dev/null || break
  sleep 1
done
if kill -0 $pid 2>/dev/null; then
  kill -9 $pid 2>/dev/null
  wait $pid 2>/dev/null
  dcode=1
else
  wait $pid
  dcode=$?
fi
if [ $dcode -eq 0 ]; then
  sed 's/^/debug: /' "$out"
else
  echo "✗ test_here failed under duso debug:"
  sed 's/\x1b\[[0-9;]*m//g; s/\(debug> \)\{1,\}//g' "$out" | grep . | head -20
  code=1
fi

rm -f /tmp/here_cwd_marker.txt "$out"
exit $code
