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

rm -f /tmp/here_cwd_marker.txt
exit $code
