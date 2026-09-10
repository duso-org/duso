#!/bin/bash
# Tests that a static() mount cannot be talked into serving a file above its
# root.
#
# net/http percent-decodes before the handler runs, so "%2e%2e" reaches the
# static branch as a literal ".." having already cleared the normalization that
# catches a plain "/../". Joining that onto the mount root would fold it into an
# escape, so the remainder is inspected on its own first.
#
# The secret files here sit beside the mounted directory, never inside it: any
# request that returns their contents has escaped.
set -u
cd "$(dirname "$0")/.."

DUSO=${DUSO:-$PWD/bin/duso}
WORK=/tmp/test_http_traversal_work
PORT=8113
pass=0
fail=0

ok() { pass=$((pass + 1)); }
bad() { fail=$((fail + 1)); echo "✗ $1"; }

rm -rf "$WORK"; mkdir -p "$WORK/pub"
echo "INDEX" > "$WORK/pub/index.html"
echo "LOGO"  > "$WORK/pub/logo.txt"
# Legitimate names that merely resemble a traversal - these must still serve.
echo "DOTNAME"  > "$WORK/pub/..hidden.txt"
echo "MIDDOT"   > "$WORK/pub/a..b.txt"
# Outside the mount. Nothing may ever return these.
echo "SECRET"   > "$WORK/secret.env"
mkdir -p "$WORK/private"; echo "PRIVATE" > "$WORK/private/key.pem"

cat > "$WORK/server.du" <<'EOF'
server = http_server({port = 8113})
server.static("/", "/tmp/test_http_traversal_work/pub")
server.static("/files", "/tmp/test_http_traversal_work/pub")
server.static("/w*", "/tmp/test_http_traversal_work/pub")
server.start()
EOF

"$DUSO" "$WORK/server.du" >/dev/null 2>&1 &
pid=$!
sleep 2

# --path-as-is keeps curl from normalizing the traversal away client-side.
raw() { curl -s -m 10 --path-as-is "http://127.0.0.1:$PORT$1"; }

# $1 = url, $2 = description. Fails if the response contains a secret.
escapes() {
  got=$(raw "$1" | head -1)
  case "$got" in
    SECRET|PRIVATE) bad "$2: GET $1 escaped the mount, got [$got]" ;;
    *) ok ;;
  esac
}

# $1 = url, $2 = expected, $3 = description
serves() {
  got=$(raw "$1" | head -1)
  [ "$got" = "$2" ] && ok || bad "$3: GET $1 expected [$2], got [$got]"
}

# --- the mount still works ---
serves /              INDEX   "root mount serves its default file"
serves /logo.txt      LOGO    "root mount serves a file"
serves /files/logo.txt LOGO   "named mount serves a file"
serves /w/logo.txt    LOGO    "wildcard mount serves a file"

# --- names that look like traversals but are ordinary files ---
serves /..hidden.txt  DOTNAME "a leading-dots filename is not a traversal"
serves /a..b.txt      MIDDOT  "dots inside a filename are not a traversal"

# --- escapes, in every spelling that reached the handler ---
escapes '/%2e%2e/secret.env'            "encoded dots, root mount"
escapes '/%2e%2e/private/key.pem'       "encoded dots into a sibling directory"
escapes '/files/%2e%2e/secret.env'      "encoded dots, named mount"
escapes '/files../secret.env'           "bare dots left by trimming a prefix"
escapes '/w../secret.env'               "bare dots, wildcard mount"
escapes '/w/%2e%2e/secret.env'          "encoded dots, wildcard mount"
escapes '/%2e%2e%2fsecret.env'          "encoded dots and encoded slash"
escapes '/..%2fsecret.env'              "encoded slash only"
escapes '/%2e%2e/%2e%2e/etc/hostname'   "repeated climb"

kill -TERM $pid 2>/dev/null
wait $pid 2>/dev/null
rm -rf "$WORK"

echo ""
echo "Test Results: $pass passed, $fail failed"
[ "$fail" -eq 0 ] || exit 1
