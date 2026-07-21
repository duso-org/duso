# Duso Primer (LLM-Optimized)

Condensed reference for writing correct Duso code. For prose explanations see `docs/learning-duso.md`; for exhaustive per-symbol docs run `duso -doc TERM` or see `docs/reference/`.

## Gotchas (read first)

- **Reserved words cannot be shadowed anywhere**: keywords (`if`, `for`, `while`, `function`, `return`, `var`, `true`, `false`, `nil`, ...) and *all builtin function names* (`print`, `len`, `type`, `now`, `map`, ...) are forbidden as variable names, function names, parameters, loop variables, or catch variables. This applies even inside nested scopes.
  - Exception: object properties CAN use reserved names (needed for JSON interop) — access them via `self` inside methods: `self.now`, `self.type`, not bare `now`/`type`.
- **`obj()` / `arr()` copy is shallow**: nested structures are shared by reference. Use `deep_copy()` for independent nested copies.
- **Crossing a process boundary (`spawn()`, `run()`, `datastore()`) auto deep-copies**, and **functions are stripped to `nil`** in the copy — closures can't survive across isolated scopes. Regex literals survive but become plain strings; reconstruct with `~pattern~` in the receiving script if needed.
- **Assignment (`x = 1`) walks the scope chain** and mutates the nearest existing binding; use `var x = 1` to force a new local that shadows an outer variable.
- **Comments**: `//` to end of line; `/* ... */` block comments, and block comments nest.
- Arrays are 0-indexed. Negative indices allowed in `substr()` (from end).

## Types

`number` (float, no separate int), `string`, `boolean`, `array`, `object`, `function`, `binary` (immutable, ref-shared bytes), `code` (compiled via `parse()`), `error`, `nil`.

```duso
type(42)          // "number"
type("x")         // "string"
type([1,2])       // "array"
type({a=1})       // "object"
type(parse("1"))  // "code"
type(parse("@"))  // "error"
```

## Variables & literals

```duso
name = "Alice"
age = 30
active = true
skills = ["Go", "Rust"]
config = {timeout = 30, retries = 3}
```

## Control flow

```duso
if age >= 18 then
  print("adult")
elseif age >= 13 then
  print("teen")
else
  print("child")
end

status = age >= 18 ? "adult" : "minor"   // ternary

for i = 0, 4 do print(i) end             // inclusive range 0..4
for item in items do print(item) end     // array iteration

count = 0
while count < 5 do
  count = count + 1
end

for i = 1, 10 do
  if i == 2 then continue end
  if i == 8 then break end
end
```

## Functions

```duso
function greet(name, greeting = "Hello")   // default params
  return greeting + ", " + name
end

double = function(x) return x * 2 end      // function as value

// closures capture outer scope at definition time; captured vars stay live
function makeCounter()
  var count = 0
  return function()
    count = count + 1
    return count
  end
end
```

Calling: positional, named, or mixed args: `configure(30, retries = 5)`.

## Arrays

```duso
nums = [10, 20, 30]
nums[0]                 // 10, mutable indexing
nums[10] = 99           // sparse: any index OK — auto-extends, gap slots are nil
len(nums)               // 11 now: counts through highest index, iteration includes the nils
push(nums, 40)           // mutate in place, returns new length
pop(nums)                // remove/return last
shift(nums)              // remove/return first
unshift(nums, 0)         // prepend
nums = sort(nums)                       // returns new array (no mutation), ascending
nums = sort(nums, function(a,b) return a > b end)  // custom comparator
map(nums, function(x) return x*2 end)
filter(nums, function(x) return x%2==0 end)
reduce(nums, function(acc,x) return acc+x end, 0)
range(1, 5)              // [1,2,3,4,5]
range(0, 10, 2)          // step
range(5, 0, -1)          // descending
a()                      // shallow copy (constructor pattern)
a(4, 5)                  // shallow copy + append 4,5
```

## Objects

```duso
person = {name = "Alice", age = 30}
person.name              // dot access (preferred for known properties)
person[key]              // bracket access — for variable/computed keys only
person.age = 31           // mutate

h = {"Content-Type" = "application/json"}   // quote odd string keys, NOT ["..."]
h["Content-Type"]                           // non-identifier keys need brackets to read

keys(config)              // array of key names
values(config)            // array of values

// methods: dot-call auto-binds; properties visible as bare vars inside methods
agent = {
  name = "Alice",
  greet = function(msg) return msg + ", " + name end
}
agent.greet("Hi")

// object-as-blueprint/constructor pattern (no class keyword)
Counter = { count = 0, increment = function() count = count+1 end }
c1 = Counter()   // shallow copy, independent state, shared method defs
c2 = Counter()
c1.increment()   // c2 unaffected

// object properties named like reserved words: use self inside methods
data = { now = 123, describe = function() return self.now end }
```

## Copying

```duso
copy = original()              // shallow: nested structures shared
copy = deep_copy(original)     // deep: fully independent
```

## Strings

```duso
'single' == "double"                  // '...' and "..." are equivalent
"{{name}} is {{age}}"                 // template interpolation — works in ALL string forms
"{{upper(name)}}: {{x ? 1 : 0}}"      // any expression: function calls, math, ternary
                                      // NOT statements: no {{if ...}}, no {{for ...}}
"""
  multiline, indentation auto-stripped
  {{name}} works here too
"""                                   // '''...''' is the single-quote equivalent

upper(s)
lower(s)
len(s)
substr(s, start, len)   // substr(s, -5) from end
split(s, " ")           // split(s, "") -> chars
join(arr, "-")
trim(s)
replace(s, "old", "new", ignore_case = true)
contains(s, "sub")
starts_with(s, "pre")   // ends_with(s, "suf") for suffix

find(s, ~\w+~)          // regex match -> array of {text, pos, len}
```

## Regex

Delimited `~...~` (Go regexp syntax).

```duso
contains(email, ~\w+@\w+\.\w+~)
find(text, ~\d+~)
replace(text, ~\d+~, "X")
replace(text, ~\w+~, function(text, pos, n) return upper(text) end)
contains("HELLO", ~hello~, true)   // 3rd arg = case-insensitive
```

## Error handling

```duso
try
  data = load("config.json")
catch (e)              // e is type "error" or whatever throw() passed
  print("Failed: " + e)
end

throw("simple string")
throw({code = "INVALID_ID", message = "...", status_code = 400})  // any type
```

`parse()` and `catch` return values of type `"error"` on failure — check with `type(x) == "error"`.

## Modules

```duso
mod = require("claude")     // cached, load stdlib/contrib module
include("helpers.du")       // execute in current scope, no namespace
```

Custom module dirs: `duso -lib-path ./my_modules script.du`, then `require("my_modules/utils")`.

Built-in LLM provider modules: `claude`, `openai`, `azure-ai`, `groq`, `deepseek`, `ollama`. Pattern: `mod.prompt("...")`, `mod.session({...})` for multi-turn / tool use.

## Execution model: isolated script instances

Every script instance — the main/top-level script, each `spawn()`, each `run()`, and each HTTP handler invoked by `http_server()` — runs as its own isolated instance with its own memory. Nothing is shared implicitly: no globals, no shared heap, no pointers across instances.

- Data passed **into** an instance (via `spawn()`/`run()` args, or an incoming HTTP request) and data passed **out** (return value, `exit()`, HTTP response) is always **deep-copied** across that boundary. Functions are stripped to `nil` on the way across since closures can't survive into another instance's isolated scope (see Serialization Contracts below).
- The only sanctioned way to **share mutable state** between instances — e.g. coordinating a worker swarm, or state visible to every HTTP handler — is `datastore()`. It's a **thread-safe** key/value store built for exactly this: concurrent instances read/write without races, with atomic ops (`increment`, `push`) and blocking waits (`wait()`) for coordination, no manual locking needed.
- This means: `run()` is a blocking call to a fresh isolated instance, not a function call into the same memory space; `spawn()` launches a fresh isolated instance in the background; an `http_server()` handler is a fresh isolated instance per request. Think "separate process/goroutine with a copy of its input," not "shared-memory thread."

## Processes: run / spawn / context

```duso
result = run("processor.du", {data = [1,2,3]})   // synchronous, blocking
pid = spawn("worker.du", {data = things})         // async fire-and-forget

// worker.du:
ctx = context()                 // nil if this is the top-level/standalone invocation
req = ctx.request()             // data passed in
exit({status = "done"})         // return value (also used to send HTTP responses)
```

Gate pattern (single script as both launcher and handler):
```duso
ctx = context()
if ctx == nil then
  result = run("child.du", {})
else
  exit({status = "done"})
end
```

## Datastore (thread-safe coordination, no locks)

```duso
store = datastore("job_123")               // in-memory by default
store = datastore("job_123", {disk = true}) // optional persistence
store.set("completed", 0)
store.increment("completed", 1)            // atomic
store.wait("completed", 5)                  // block until value == 5
store.wait("completed", 5, timeout = 30)    // THROWS on timeout (wrap in try/catch)
store.wait("temp", function(v) return v >= 20 end, 30)  // predicate wait
store.push("jobs", job)                     // atomic array append
job = store.pop("jobs")                     // non-blocking; nil if empty
job = store.shift_wait("jobs", 10)          // FIFO queue: block up to 10s, nil on timeout
job = store.pop_wait("jobs", 10)            // LIFO variant, same semantics
```

## HTTP client & server

```duso
response = fetch("https://api.example.com/users")
if response.ok then data = response.json() end
fetch(url, {method = "POST", headers = {Authorization = "Bearer TOKEN"}, body = format_json(obj)})

server = http_server({port = 8080})
server.route("GET", "/", "handlers/home.du")   // or self-referential: server.route("GET","/")
server.start()   // blocks

// in a route handler script:
ctx = context()
req = ctx.request()
res = ctx.response()
res.html("Hello")
res.json({success = true})
```

## JSON / time / math

```duso
data = parse_json(json_str)
json = format_json(obj)            // format_json(obj, 2) pretty-prints

ts = now()
format_time(ts, "YYYY-MM-DD")
parse_time("2026-01-22")

abs(-42)
sqrt(16)
pow(2, 3)
floor(3.7)              // also ceil(), round()
sin(r)                  // radians; also cos(), tan(), atan2(y, x), pi()
exp(x)
log(x)                  // base 10; ln(x) for natural log
```

## Files & virtual filesystems

```duso
content = load("data.txt")
save("output.txt", "text")
append_file("log.txt", "line\n")
copy_file(src, dst)     // also move_file(), remove_file()
file_exists(p)          // also file_type()
list_dir("./data")
make_dir("./nested/path")
remove_dir("./empty")
```

`/EMBED/` — read-only, baked into binary. `/STORE/` — read/write, backed by datastore. Run untrusted (e.g. LLM-generated) scripts sandboxed to these only: `duso -no-files`.

## SQL

`sql()` gives thread-safe, namespaced, pooled connections. Drivers: `"mysql"`/`"mariadb"`/`"tidb"`, or `"postgres"`/`"pg"`.

```duso
db = sql("myapp", {driver = "mysql", host = "localhost", database = "mydb", user = "root", password = "secret"})
db = sql("myapp")   // retrieve existing pooled connection (config omitted)

rows = db.query("SELECT id, name FROM users WHERE id = ?", [42])   // -> array of objects
rows = db.query("SELECT id, name FROM users", [], false)            // -> array of arrays (positional)
n = db.exec("UPDATE users SET seen = ? WHERE id = ?", [timestamp(), 42])  // rows affected
db.ping()
db.close()
```
MySQL placeholders are `?`; Postgres uses `$1, $2, ...`.

## WebSocket

```duso
ws = websocket("wss://example.com/socket", {headers = {Authorization = "Bearer TOKEN"}})
ws.write("hello")
msg = ws.read(5)          // blocks up to 5s, nil on timeout/disconnect
ws.id                     // connection ID, usable from other instances
ws.close()
ws.is_connected()

send_websocket(conn_id, "message")        // send from outside the owning handler
send_websocket([id1, id2], "broadcast")   // to multiple connections
```

## CSV

```duso
rows = parse_csv(csv_string)              // -> array of arrays; parse_csv(str, "\t") for TSV
csv_string = format_csv(rows)             // array of arrays -> CSV string
```

## Crypto, hashing & encoding

```duso
hash("sha256", data)                       // sha256/sha512/sha1/md5, string or binary
hmac("sha256", data, key)
hash_password(pw)                          // bcrypt, hash_password(pw, cost)
verify_password(pw, hash)                  // constant-time compare
sig = sign_ec(data, private_key_pem)      // ES256; sign_rsa(data, key) for RSA
verify_ec(data, sig, public_key_pem)       // also verify_rsa(), verify_ed25519()
encode_base64(data)                        // decode_base64(str) to reverse
uuid()                                     // UUIDv7
```

## Images

PNG/JPEG/GIF. `load_image()`/`save_image()`, `scale_image()` (fit/fill/stretch), `crop_image()`, `rotate_image()`, `flip_image_x/y()`, `grayscale_image()`, `adjust_image_opacity()`, `composite_image()`, `convert_image()`. See `duso -doc image` for the full set.

## Markdown

```duso
markdown_html(md)    // -> HTML
markdown_ansi(md)    // -> terminal-colored text
markdown_text(md)    // -> plain text
```

## sys()

Introspect how the current process was invoked (CLI flags, runtime info) — CLI only: `sys("key")`. See `duso -doc sys`.

## Binary type

```duso
image = load_binary("avatar.png")   // ref-shared, not copied
len(image)                          // size in bytes
image.filename                      // metadata
save_binary(image, "copy.png")
```

## CLI essentials

```bash
duso script.du                 # run a file
duso -c 'print("hi")'          # inline
duso -repl                     # interactive
duso -read                     # ls/less-style browser over embedded docs, examples, and modules — start here
duso -doc TERM                 # docs for a builtin/keyword
duso -lint file.du             # static analysis
duso -init DIR                 # scaffold a project
duso -no-files -no-stdin       # sandbox untrusted code
```

## Parallel

```duso
results = parallel([
  function() return claude.prompt("A") end,
  function() return claude.prompt("B") end
])
```
Each entry runs concurrently; an error in one yields `nil` at that index.
