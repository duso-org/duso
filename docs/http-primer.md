# HTTP Primer (LLM-Optimized)

Condensed reference for `http_server()`, `fetch()` and `websocket()`. Companion to
`docs/duso-primer.md`; exhaustive per-symbol docs in `docs/reference/http_server.md`,
`docs/reference/fetch.md` and `docs/reference/websocket.md`.

## Gotchas (read first)

- **One handler script per route.** `route()` names a file, and that file is the whole
  handler. Omitting the file so the server script handles the route itself is a one-endpoint
  convenience, not the way to build an app.
- **Response methods are terminal.** `res.json(...)`, `res.text(...)`, `res.redirect(...)`
  send the response and exit the handler immediately. Nothing after them runs.
- **A handler that sends nothing returns 204 No Content**, not an error.
- **`req.host` and `req.remote_addr` are not in `req.headers`.** They come off the request
  and the socket, not the request text.
- **The batteries are server config, not middleware.** TLS, CORS, JWT, uploads and every
  resource limit are keys in the `http_server({...})` object. There is no middleware chain
  to register, and no ordering to get wrong.
- **Each request gets a fresh evaluator instance** in its own goroutine. No shared state
  between handlers — use `datastore()` to share anything.
- **WebSocket routes must be registered as `"WS"`.** `"*"` (all methods) does not match an
  upgrade request.
- **WebSocket needs HTTP/1.1.** The upgrade handshake is forbidden over HTTP/2. Browsers
  and duso's own `websocket()` client don't offer `h2` in ALPN, so this only bites a custom
  client that insists on it.

## fetch()

```duso
r = fetch("https://api.example.com/users")
if r.ok then data = r.json() end            // ok == status < 400

r = fetch(url, {
  method  = "POST",                          // any method; default GET
  headers = {"Content-Type" = "application/json", Authorization = "Bearer " + token},
  body    = format_json(obj),
  timeout = 10                               // seconds
})

r.status                                     // number
r.ok                                         // status < 400
r.body                                       // raw string
r.json()                                     // parsed
r.text()                                     // same as .body
r.headers                                    // canonical-form keys ("Content-Type")
r.proto                                      // "HTTP/1.1" or "HTTP/2.0"
```

Blocking from the script's point of view — no async/await, no promise. Works in handler
scripts as well as at the top level, so a route can call out to another service.

**HTTP/2 is negotiated automatically** over TLS whenever the server supports it; check
`r.proto` to see what happened. Plaintext h2c is not supported — `http://` is always
HTTP/1.1. Connection pooling is automatic.

A non-2xx response is **not** an error — nothing throws. Check `.ok` or `.status`.
Network-level failures (DNS, refused, timeout) do throw; wrap in try/catch if the endpoint
is untrusted.

For header names with special characters use quoted keys: `{["Content-Type"] = "..."}`.

## http_server(): the shape

**One handler script per route.** That is the design. A server script wires up routes and
starts listening; each route names a file, and that file is the whole handler. Nothing is
shared between them, so a route is read, changed and reasoned about on its own.

```duso
// server.du — routing table and nothing else
server = http_server({port = 8080})

server.route("GET",  "/",           "handlers/home.du")
server.route("GET",  "/users/:id",  "handlers/user_show.du")
server.route("POST", "/users",      "handlers/user_create.du")
server.route("WS",   "/chat",       "handlers/chat.du")
server.static("/assets", "./public")

server.start()                                    // blocks
```

```duso
// handlers/user_show.du — one route, start to finish
ctx = context()
req = ctx.request()

user = datastore("users").get(req.params.id)
if user == nil then
  ctx.response().error(404, "no such user")
end

ctx.response().json(user)
```

The handler file runs top to bottom on each request, with no wrapper function to declare
and no export to register. `server.du` stays a readable routing table as the app grows.

Handler paths resolve against **the server script's own directory**, not the working
directory, so `duso /srv/app/server.du` finds `/srv/app/handlers/home.du` no matter where
it was launched from.

### The one-script form

Omit the third argument to `route()` and the server script handles the route itself. Then
`context()` — nil during setup, populated during a request — is what tells the two apart:

```duso
ctx = context()

if ctx == nil then
  server = http_server({port = 8080})
  server.route("GET", "/")                        // no handler file: this script
  server.start()
end

ctx.response().text("hello")
```

This is a convenience for a single-endpoint script — a health check, a webhook receiver, a
demo that has to fit in one file. It is **not** the pattern to grow an application in: every
route added this way piles another branch into one file, all sharing a top-level scope, and
the routing table stops being visible anywhere. Reach for a handler file per route as soon
as there is a second route.

## Config

Everything below is optional; every value shown is the default.

```duso
server = http_server({
  port    = 8080,
  address = "0.0.0.0",

  // TLS. HTTP/2 is negotiated automatically over ALPN when enabled.
  https     = false,
  cert_file = nil,                          // required when https = true
  key_file  = nil,
  cert_reload_interval = 86400,             // re-read the cert pair once a day

  // timeouts
  timeout                 = 30,             // socket read/write
  request_handler_timeout = 30,             // handler script execution
  idle_timeout            = 120,

  // resource limits — enforced at the HTTP level, before the handler runs
  max_body_size    = 10485760,              // 10MB  -> 413 Payload Too Large
  max_header_size  = 8192,                  // 8KB   -> 431 Request Header Fields Too Large
  max_headers      = 100,                   // -> 431
  max_form_fields  = 1000,                  // -> 400 Bad Request

  // static serving
  default              = ["index.html"],    // directory default file(s); nil disables
  directory            = false,             // directory listing when no default matches
  static_cache_control = "public, max-age=3600",
  cache_control        = "no-cache, no-store, must-revalidate",   // dynamic responses

  access_log = true,                        // Apache Combined Log Format to stderr

  cors = {
    enabled     = false,
    origins     = [],                       // "*" or ["https://app.example.com"]
    methods     = [],
    headers     = [],
    credentials = false,
    max_age     = 0
  },

  jwt = {
    enabled          = false,
    secret           = nil,                 // HS256
    rs256_private_key = nil,                // PEM
    rs256_public_key  = nil,
    required         = false                // reject every request without a valid token
  },

  uploads = {
    enabled  = false,
    max_size = 10240                        // KB per file
  },

  max_websocket_connections = 0,            // 0 = unlimited; -> 503 when exceeded
  websocket = {
    read_queue_size  = 100,
    write_queue_size = 100,
    read_timeout     = 30,
    idle_timeout     = 300,                 // 0 disables
    max_message_size = 65536,               // 0 = unlimited
    max_messages_per_second = 0             // 0 = unlimited
  }
})
```

The limits are real DOS protection: they are checked by the HTTP layer, so an oversized
body never reaches a handler and never costs a script instance.

## Routing

```duso
server.route(method, path, handler)         // handler = path to the script for this route

server.route("GET", "/api",          "handlers/api.du")        // exact
server.route("GET", "/users/:id",    "handlers/user.du")       // path param -> req.params.id
server.route("GET", "/api/*",        "handlers/api_any.du")    // wildcard prefix
server.route(["GET", "POST"], "/form", "handlers/form.du")     // array of methods
server.route("*", "/any",            "handlers/any.du")        // or nil — all HTTP methods, NOT WebSocket
server.route("WS", "/chat",          "handlers/chat.du")
server.static("/", "./public")
```

Methods are case-insensitive. Most specific wins: params beat wildcards, longer exact
paths beat shorter, exact beats wildcard.

`static()` routes bypass script execution entirely — served straight from the filesystem,
content type from the extension, 404 on miss, no handler timeout. Use them for assets.

## Request

```duso
ctx = context()
req = ctx.request()

req.method        // "GET"
req.path          // "/api/users"
req.proto         // "HTTP/1.1" | "HTTP/2.0"
req.host          // "shop.example.com:8080" — Host header, port included
req.remote_addr   // "203.0.113.7" or "2001:db8::1", no port, no brackets
req.headers       // object
req.query         // parsed ?a=1&b=2
req.cookies       // parsed Cookie header
req.form          // parsed POST/PUT form body
req.params        // path params from /users/:id
req.body          // raw string
req.files         // uploads (empty object, never nil)
req.jwt_claims    // verified claims, or nil
```

`query`, `cookies`, `form` and `params` are already-parsed objects — a repeated name
becomes an array. A missing key is `nil`, so guard before use:

```duso
sid = req.cookies.sid
session = sid ? datastore("sessions").get(sid) : nil
```

**Virtual hosting** falls out of `req.host` — one server, many hostnames:

```duso
tenant = split(split(req.host, ":")[0], ".")[0]     // "shop" of shop.example.com
```

**`req.remote_addr` cannot be spoofed** — it's the peer the server is actually talking to.
Behind a proxy it *is* the proxy, and the real client is in a header the proxy sets:

```duso
ip = req.headers["X-Forwarded-For"]        // may be "client, proxy1, proxy2"
if ip == nil then ip = req.remote_addr end
```

Headers *are* client-supplied. Only trust `X-Forwarded-For` when a proxy you control
overwrites it on every request.

## Response

```duso
res = ctx.response()

res.json(data, 200)                         // application/json
res.text("hi", 200)                         // text/plain
res.html("<h1>hi</h1>", 200)                // text/html
res.binary(bytes, "image/png", 200)         // any content type
res.file("./public/index.html", 200)        // from the filesystem
res.error(404, "Not Found")                 // JSON error body
res.redirect("/dashboard", 302)
res.response(body, 200, {"X-Custom" = "1"}) // generic
```

Status is optional and defaults to 200. **Every one of these sends and exits the handler.**

Every method takes an optional headers object as its **last** argument, and supplied
headers beat the defaults — no need to drop to `response()` to set one header:

```duso
res.redirect("/dashboard", 302, {
  "Set-Cookie" = "sid=" + token + "; Path=/; HttpOnly; Secure; SameSite=Lax"
})
res.json(doc, 200, {"Content-Type" = "application/vnd.api+json"})
res.json(doc, 200, headers = {"X-Trace" = id})       // also accepted by name
```

**An array value emits the header once per element** — which is what `Set-Cookie` needs
when one response sets a cookie and clears another:

```duso
res.json({ok = true}, 200, {
  "Set-Cookie" = ["sid=" + token + "; Path=/", "old_sid=; Path=/; Max-Age=0"]
})
```

`exit()` is the low-level alternative, and takes the same repeated-header arrays:

```duso
exit({status = 200, body = "...", headers = {"Content-Type" = "text/plain"}})
```

## JWT

Enable it in config, then use two methods that only exist inside handlers:

```duso
claims = req.verify_jwt()                              // configured key; nil if invalid/expired
claims = req.verify_jwt({public_key = load("partner.pem")})   // partner's RS256 key
claims = req.verify_jwt({secret = "other-secret"})            // alternative HS256 secret

token = sign_jwt({user_id = 123})                             // HS256 by default
token = sign_jwt({user_id = 123}, {algorithm = "RS256", expires_in = 7200})
```

The algorithm is auto-detected from the token header. `required = true` in config rejects
unauthenticated requests before the handler runs; leave it false to authenticate per route.
You manage the keys — load, cache and rotate PEM strings yourself.

## File uploads

```duso
// config: uploads = {enabled = true, max_size = 10240}

f = req.files.avatar
if f then
  f.filename                                // as sent by the client
  f.content_type                            // from the upload header or the extension
  f.size                                    // bytes
  if type(f.data) == "binary" then
    save_binary(f.data, "/STORE/uploads/" + f.filename)
  elseif type(f.data) == "string" then
    parsed = parse_json(f.data)
  end
end

for a in req.files.attachments do print(a.filename) end   // repeated field name -> array
```

Text MIME types (`text/*`, `application/json`, `application/xml`) arrive as **string**;
everything else as **binary**. Oversized files are silently skipped, not rejected — check
for the field rather than assuming it arrived.

## CORS

```duso
cors = {
  enabled = true,
  origins = ["https://app.example.com"],    // or "*"
  methods = ["GET", "POST", "PUT", "DELETE"],
  headers = ["Content-Type", "Authorization"],
  credentials = false,
  max_age = 3600
}
```

Preflight `OPTIONS` requests are answered with 204 automatically. You never write a CORS
handler.

## TLS and certificate renewal

An HTTPS server re-reads `cert_file` and `key_file` while running, so certbot rewriting the
same paths every 60-90 days needs **no reload hook and no restart**. Every
`cert_reload_interval` the server stats both files; if mtime or size moved it loads the
pair and swaps it in only when the cert parses and the key matches. In-flight handshakes
finish on the old cert.

A pair that fails to load leaves the running cert in place and retries in a minute — that's
the normal reading when the check lands between the renewer writing the cert and writing
the key. Certs that fail to load **at startup** are still fatal: the server refuses to
start rather than listen without TLS.

## Graceful shutdown

`start()` blocks until the process is asked to stop. On Ctrl+C or `systemctl stop`, in
order: stop accepting connections → in-flight handlers get up to 30s to finish and send →
WebSocket connections close → datastores flush. Then exit 0. A datastore write from a
handler that was still finishing is included in the final snapshot.

## WebSocket: server side

```duso
server.route("WS", "/chat", "handlers/chat.du")
server.route("WS", "/game/:id", "handlers/game.du")
```

Each connection spawns a persistent handler that runs until the client disconnects.

```duso
// handlers/chat.du
ctx  = context()
conn = ctx.connection()
req  = ctx.request()                        // full context from the upgrade request

if req.jwt_claims == nil then exit() end    // authenticate before accepting
conn.accept()                               // completes the upgrade

while true do
  msg = conn.read(timeout = 30)             // nil on timeout OR disconnect
  if msg == nil then break end
  conn.write("echo: " + msg)
end
```

Connection object: `accept()`, `read([timeout])`, `write(msg)`, `close()`,
`is_connected()`, and `id` (unique string).

`req.params`, `req.headers` and `req.jwt_claims` are all available and come from the
initial upgrade request — that's where you do auth, before `accept()`.

## Backpressure

Messages are queued in both directions (100 each by default) rather than dropped. **A full
write queue means `write()` returns `nil`** — the client is reading slower than you are
sending:

```duso
if conn.write(msg) == nil then
  conn.close()                              // client too slow; don't spin
end
```

A full *read* queue drops incoming messages and closes the connection. Size the queues for
your burst, and set `max_messages_per_second` to rate-limit abusive senders (excess
messages are dropped silently; 10 consecutive drops closes the connection).

## Broadcasting

`send_websocket()` reaches any connection by ID from anywhere — another handler, a spawned
script, a timer:

```duso
send_websocket(conn_id, "hello")            // returns bytes queued, nil if the queue is full
send_websocket([id1, id2], "broadcast")     // multiple at once
```

Combine with `datastore()` to keep the roster, since handlers share no state:

```duso
store = datastore("chat")
conn.accept()
store.push("conns", conn.id)

while true do
  msg = conn.read()
  if msg == nil then break end
  send_websocket(store.get("conns"), msg)
end
```

## WebSocket: client side

```duso
ws = websocket("wss://gateway.example.com/", {
  headers = {Authorization = "Bearer " + token},
  read_queue_size  = 200,
  write_queue_size = 200,
  read_timeout     = 60,                    // default for read() with no argument
  idle_timeout     = 300,                   // 0 disables
  max_message_size = 65536,                 // 0 = unlimited
  max_messages_per_second = 0
})

ws.id                                       // usable with send_websocket() from other scripts
ws.write(format_json({op = 1}))
msg = ws.read(10)                           // or read(timeout = 10); nil on timeout/disconnect
ws.is_connected()
ws.close()
```

`http://` and `https://` URLs are converted to `ws://` / `wss://` automatically.

`websocket()` **throws** on a bad URL, a bad scheme, or a failed connection — unlike
`read()`, which returns `nil`. Wrap the dial in try/catch:

```duso
try
  ws = websocket("ws://maybe-down.example.com")
catch (e)
  print("connect failed: " + e)
end
```

Messages are UTF-8 text only; binary frames are not supported. `read()` with no timeout
blocks indefinitely — use a timeout in any loop you expect to exit.

Each script owns its own connection; connections are not shared across `spawn()`. Use
`send_websocket()` with the id, or `datastore()`, to coordinate.

## See also

- `docs/duso-primer.md` — the language
- `docs/datastore-primer.md` — coordination and shared state between handlers
- `docs/reference/http_server.md` — full server reference
- `docs/reference/fetch.md` — full client reference
- `docs/reference/websocket.md` / `docs/reference/send_websocket.md`
