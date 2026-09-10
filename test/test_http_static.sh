#!/bin/bash
# Tests which requests a static() mount answers.
#
# A static mount names a directory, so it has to serve everything below it, at a
# segment boundary - /assets serves /assets/logo.txt but must not bleed into
# /assetsfoo. The wildcard spellings were the only ones that ever worked, so
# they are pinned here too: they are what existing code is written against.
set -u
cd "$(dirname "$0")/.."

DUSO=${DUSO:-$PWD/bin/duso}
WORK=/tmp/test_http_static_work
PORT=8110
pass=0
fail=0

ok() { pass=$((pass + 1)); }
bad() { fail=$((fail + 1)); echo "✗ $1"; }

rm -rf "$WORK"; mkdir -p "$WORK/pub"
echo "LOGO" > "$WORK/pub/logo.txt"
echo "INDEX" > "$WORK/pub/index.html"
echo "SOLO" > "$WORK/solo.txt"

cat > "$WORK/api.du" <<'EOF'
ctx = context()
ctx.response().text("DYNAMIC")
EOF

cat > "$WORK/api_wild.du" <<'EOF'
ctx = context()
ctx.response().text("DYNAMIC-WILDCARD")
EOF

cat > "$WORK/server.du" <<'EOF'
server = http_server({port = 8110})
server.static("/assets", "/tmp/test_http_static_work/pub")
server.static("/docs*", "/tmp/test_http_static_work/pub")
server.static("/help/*", "/tmp/test_http_static_work/pub")
server.static("/README", "/tmp/test_http_static_work/solo.txt")
server.static("/", "/tmp/test_http_static_work/pub")
server.route("GET", "/assets/api", "api.du")
server.route("GET", "/assets/api/*", "api_wild.du")
server.start()
EOF

"$DUSO" "$WORK/server.du" >/dev/null 2>&1 &
pid=$!
sleep 2

# $1 = url, $2 = expected body, $3 = description
check() {
  got=$(curl -s -m 10 "http://127.0.0.1:$PORT$1" | head -1)
  [ "$got" = "$2" ] && ok || bad "$3: GET $1 expected [$2], got [$got]"
}

# --- a bare mount serves the directory below it ---
check /assets/logo.txt LOGO      "bare mount, file below"
check /assets          INDEX     "bare mount, mount itself serves the default file"

# --- and stops at the segment boundary ---
check /assetsfoo "404 page not found" "bare mount must not bleed into a sibling path"

# --- a longer dynamic route still wins over the mount it sits under ---
# A mount is a prefix match, so it has to rank with the wildcards by length.
# Ranked with the exact routes instead, /assets would outrank /assets/api/* and
# swallow it, because the category is compared before the length.
check /assets/api       DYNAMIC          "exact dynamic route under a static mount is not shadowed"
check /assets/api/thing DYNAMIC-WILDCARD "wildcard dynamic route under a static mount is not shadowed"

# --- the wildcard spellings that existing code uses ---
check /docs/logo.txt LOGO  "trailing-* mount, file below"
check /docs/         INDEX "trailing-* mount, directory"
check /docs          INDEX "trailing-* mount, no trailing slash"
check /help/logo.txt LOGO  "/*-suffixed mount, file below"

# --- a single file mount answers its own URL and nothing under it ---
check /README     SOLO                  "single-file mount"
check /README/x   "404 page not found"  "single-file mount serves nothing below itself"

# --- "/" is the document root: it serves the whole tree ---
# It is the shortest prefix there is, so it sorts last and catches only what no
# other route claimed. Everything above this line still wins against it.
check /          INDEX "root mount serves its default file"
check /logo.txt  LOGO  "root mount serves files below it"
check /nope.txt  "404 page not found" "root mount 404s a file it does not have"

kill -TERM $pid 2>/dev/null
wait $pid 2>/dev/null
rm -rf "$WORK"

echo ""
echo "Test Results: $pass passed, $fail failed"
[ "$fail" -eq 0 ] || exit 1
