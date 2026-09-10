#!/bin/bash
# Tests how http_server() resolves the paths it is handed: /HERE/ in route()
# and static(), a bare path, and an omitted handler.
#
# Has to live outside a .du script - the cases only differ when the working
# directory, appDir, and the module's own directory are three different places,
# so the test has to control the directory duso is launched from.
#
# Every /HERE/ case has a decoy of the same name sitting in appDir. If /HERE/
# ever resolves to appDir the decoy answers and the assertion fails, instead of
# the test passing because both paths happened to reach the same file.
set -u
cd "$(dirname "$0")/.."

DUSO=${DUSO:-$PWD/bin/duso}
WORK=/tmp/test_http_paths_work
pass=0
fail=0

ok() { pass=$((pass + 1)); }
bad() { fail=$((fail + 1)); echo "✗ $1"; }

rm -rf "$WORK"; mkdir -p "$WORK/app/public" "$WORK/lib/public"

# --- fixtures -------------------------------------------------------------
# lib/ is the module's directory: what /HERE/ must mean.
# app/ is appDir: what a bare path must mean, and where the decoys live.

cat > "$WORK/lib/handler.du" <<'EOF'
ctx = context()
ctx.response().text("HERE-HANDLER")
EOF
echo "HERE-STATIC" > "$WORK/lib/public/hello.txt"

cat > "$WORK/app/handler.du" <<'EOF'
ctx = context()
ctx.response().text("APPDIR-HANDLER")
EOF
echo "APPDIR-STATIC" > "$WORK/app/public/hello.txt"

# The module registers its routes with /HERE/, from inside a function, called
# by a script in another directory. /HERE/ is lexical, so it has to name lib/.
cat > "$WORK/lib/routes.du" <<'EOF'
function setup(server)
  server.route("GET", "/here", "/HERE/handler.du")
  server.static("/here-static/*", "/HERE/public")
end
return {setup = setup}
EOF

cat > "$WORK/app/main.du" <<'EOF'
routes = require("/tmp/test_http_paths_work/lib/routes.du")
server = http_server({port = 8097})
routes.setup(server)
server.route("GET", "/bare", "handler.du")
server.static("/bare-static/*", "public")
print("serving")
server.start()
EOF

# The minimal.du pattern: no handler argument at all.
cat > "$WORK/app/self.du" <<'EOF'
ctx = context()
if ctx == nil then
  server = http_server({port = 8098})
  server.route("GET", "/")
  print("serving")
  server.start()
end
ctx.response().text("SELF-HANDLER")
EOF

get() { curl -s -m 10 "http://127.0.0.1:$1$2"; }

# --- /HERE/ and bare paths, launched from a third directory ----------------
# cwd is the repo root, appDir is app/, the module lives in lib/ - all different.

"$DUSO" "$WORK/app/main.du" >/dev/null 2>&1 &
pid=$!
sleep 2

got=$(get 8097 /here)
[ "$got" = "HERE-HANDLER" ] \
  && ok || bad "route() /HERE/ handler: expected HERE-HANDLER, got [$got]"

got=$(get 8097 /here-static/hello.txt)
[ "$got" = "HERE-STATIC" ] \
  && ok || bad "static() /HERE/ dir: expected HERE-STATIC, got [$got]"

# The other half of the contract: bare still means appDir, not the module dir.
got=$(get 8097 /bare)
[ "$got" = "APPDIR-HANDLER" ] \
  && ok || bad "route() bare handler: expected APPDIR-HANDLER, got [$got]"

got=$(get 8097 /bare-static/hello.txt)
[ "$got" = "APPDIR-STATIC" ] \
  && ok || bad "static() bare dir: expected APPDIR-STATIC, got [$got]"

kill -TERM $pid 2>/dev/null
wait $pid 2>/dev/null

# --- an omitted handler, launched by a relative path -----------------------
# The script path duso was invoked with is relative to the working directory.
# Resolving it a second time against appDir lands at app/app/self.du, so this
# only shows up when the working directory is not appDir.

(cd "$WORK" && exec "$DUSO" app/self.du >/dev/null 2>&1) &
pid=$!
disown $pid 2>/dev/null    # bash announces signalled jobs otherwise
sleep 2

got=$(get 8098 /)
[ "$got" = "SELF-HANDLER" ] \
  && ok || bad "route() with no handler, relative launch: got [$got]"

kill -TERM $pid 2>/dev/null
pkill -f "app/self.du" 2>/dev/null
sleep 0.5

# Same script, launched by an absolute path - the case that always worked.
"$DUSO" "$WORK/app/self.du" >/dev/null 2>&1 &
pid=$!
sleep 2

got=$(get 8098 /)
[ "$got" = "SELF-HANDLER" ] \
  && ok || bad "route() with no handler, absolute launch: got [$got]"

kill -TERM $pid 2>/dev/null
wait $pid 2>/dev/null

rm -rf "$WORK"

echo ""
echo "Test Results: $pass passed, $fail failed"
[ "$fail" -eq 0 ] || exit 1
