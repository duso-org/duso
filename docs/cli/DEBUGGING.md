# Debugging and Interactive Testing

Duso provides multiple debugging and interaction modes for developing and testing scripts.

## Interactive Debugging with `-debug`

Use `-debug` to enable the interactive debugger. When your script calls `breakpoint()`, execution pauses and you get an interactive REPL at that point.

### Basic Breakpoint Usage

```duso
x = 5
y = 10
print("About to hit breakpoint")
breakpoint()  // Execution pauses here
print("After breakpoint")
```

Run with `-debug`:

```bash
duso -debug script.du
```

Output:

```
About to hit breakpoint

script.du:4
    1 | x = 5
    2 | y = 10
    3 | print("About to hit breakpoint")

    4 | breakpoint()
      ^
    5 | print("After breakpoint")

Call stack:
  at main (script.du:4)

Type 'c' to continue, or inspect variables.
debug>
```

### Debug Commands

At the `debug>` prompt, you can:

- **`c`** - Continue execution
- **Any expression** - Evaluate and print result

### Inspecting Variables

```
debug> x
5

debug> y
10

debug> x + y
15

debug> {sum = x + y, product = x * y}
{sum = 15, product = 50}
```

### Watch Points

Use `watch()` to pause at a specific condition with a message:

```duso
for i = 1, 100 do
  if i == 50 then
    watch("Hit iteration 50")  // Pause with message
  end
end
```

### Conditional Breakpoints

Use `if` to conditionally break:

```duso
for i = 1, 100 do
  if i == 50 then
    breakpoint()  // Only pauses at i=50
  end
end
```

## HTTP-Based Interaction with `-stdin-port`

The `-stdin-port` option exposes script stdin/stdout over a simple HTTP API. This enables:

- **LLM agents** to interact with running scripts
- **Remote terminals** to drive scripts over HTTP
- **Testing frameworks** to send input and capture output
- **Interactive debugging** over HTTP instead of direct terminal access

### Quick Start

Terminal 1 - Start script with HTTP stdin/stdout:

```bash
duso -stdin-port 9999 script.du
```

Output:

```
HTTP stdin/stdout transport listening on http://localhost:9999
  GET /        - Read accumulated output
  GET /input   - Block until input() is called, returns accumulated output
  POST /input  - Send input in request body to waiting input() call
```

Terminal 2 - Interact via HTTP using duso's `fetch()`:

```bash
# Get accumulated output
duso -c 'print(fetch("http://localhost:9999/").body)'

# Wait for input prompt (blocks until script calls input())
duso -c 'print(fetch("http://localhost:9999/input").body)'

# Send input when script is waiting
duso -c 'fetch("http://localhost:9999/input", {method = "POST", body = "your input here"})'

# Get updated output
duso -c 'print(fetch("http://localhost:9999/").body)'
```

### HTTP Endpoints

#### `GET /` - Read Output

Returns all accumulated stdout as plain text.

```bash
duso -c 'print(fetch("http://localhost:9999/").body)'
```

Response (200 OK):
```
Starting
Processing...
Result: 42
```

Non-blocking. Returns immediately.

#### `GET /input` - Wait for Input Prompt

Blocks until the running script calls `input()`. Returns accumulated output up to that point.

```bash
duso -c 'print(fetch("http://localhost:9999/input").body)'
```

**Blocks until:**
- Script calls `input()` with a prompt
- Timeout is reached (returns error after ~60 seconds)

Response (200 OK) when input is needed:
```
Starting
What is your name?
```

Use this to know when the script needs input.

#### `POST /input` - Send Input

Send data to a waiting `input()` call. Unblocks the script.

```bash
duso -c 'fetch("http://localhost:9999/input", {method = "POST", body = "Alice"})'
```

Response (200 OK):
```
ok
```

The running script's `input()` call returns the body text (with newline appended).

### Interactive Script Example

Create `interactive.du`:

```duso
print("What is your name?")
name = input("Name: ")
print("Hello, " + name + "!")

print("What is your age?")
age = input("Age: ")
print("You are " + age + " years old.")
```

Terminal 1:
```bash
duso -stdin-port 9999 interactive.du
```

Terminal 2 - Interact via HTTP:

```bash
# Wait for input prompt
duso -c 'print(fetch("http://localhost:9999/input").body)'
# Output:
# What is your name?
# Name:

# Send name
duso -c 'fetch("http://localhost:9999/input", {method = "POST", body = "Alice"})'

# Get response and next prompt
duso -c 'print(fetch("http://localhost:9999/input").body)'
# Output:
# What is your name?
# Name: Hello, Alice!
# What is your age?
# Age:

# Send age
duso -c 'fetch("http://localhost:9999/input", {method = "POST", body = "30"})'

# Get final output
duso -c 'print(fetch("http://localhost:9999/").body)'
# Output:
# What is your name?
# Name: Hello, Alice!
# What is your age?
# Age: You are 30 years old.
```

## Combining `-debug` and `-stdin-port`

Run with both flags for debug mode accessible over HTTP:

```bash
duso -debug -stdin-port 9999 script.du
```

Now you can:
- Hit `breakpoint()` and inspect variables
- Use HTTP to fetch output and variable states
- Send `c` to continue (via POST /input)
- All without direct terminal access

Example:

Terminal 1:
```bash
duso -debug -stdin-port 9999 script.du
```

Terminal 2:
```bash
# Script hits breakpoint, wait for prompt
duso -c 'print(fetch("http://localhost:9999/input").body)'

# Inspect variable x
duso -c 'fetch("http://localhost:9999/input", {method = "POST", body = "x"})'

# Get the response (x = 42)
duso -c 'print(fetch("http://localhost:9999/input").body)'

# Continue execution
duso -c 'fetch("http://localhost:9999/input", {method = "POST", body = "c"})'

# Get output after breakpoint
duso -c 'print(fetch("http://localhost:9999/").body)'
```

## LLM Agent Interaction

The `-stdin-port` option is designed for LLM agents to drive scripts. An LLM can:

1. Start a script with `-stdin-port`
2. Periodically fetch output with `GET /`
3. When input is needed, receive context from `GET /input`
4. Send input with `POST /input`
5. Repeat until script completes

### Self-Aware Script Example

Create `self-aware.du`:

```duso
claude = require("claude")

print("I am Claude, running in duso.")
user_request = input("What would you like me to do? ")

session = claude.session({
  system = "You are Claude, an AI assistant. You're helpful and curious."
})

response = session.prompt(user_request)
print(response)
```

Start script:
```bash
duso -stdin-port 8000 self-aware.du
```

LLM interaction (pseudocode):

```
while script_running:
  output = GET http://localhost:8000/

  // Check if input is needed
  input_prompt = GET http://localhost:8000/input  (blocks)

  // LLM reads prompt and output, decides what to send
  input_text = llm.decide_input(output, input_prompt)

  // Send input
  POST http://localhost:8000/input with body=input_text

  // Continue loop
```

This enables true bidirectional interaction between an LLM and a duso script.

## Testing Patterns

### Test with curl (or duso fetch)

For simple testing, use curl or duso's `fetch()`:

```bash
# Start script
duso -stdin-port 9999 quiz.du &

# Wait for first prompt
curl http://localhost:9999/input

# Send answer
curl -X POST http://localhost:9999/input -d "Paris"

# Get output
curl http://localhost:9999/
```

Or with duso fetch (as shown earlier):

```bash
duso -c 'print(fetch("http://localhost:9999/input").body)'
duso -c 'fetch("http://localhost:9999/input", {method="POST", body="Paris"})'
duso -c 'print(fetch("http://localhost:9999/").body)'
```

### Automated Testing Script

Create `test_quiz.du`:

```duso
// Start quiz script with HTTP transport
spawn_result = spawn("quiz.du", {stdin_port = 9999})

sleep(1)  // Wait for startup

// Test interaction
responses = [
  "Paris",
  "4",
  "Shakespeare"
]

for answer in responses do
  // Wait for prompt
  output = fetch("http://localhost:9999/input").body
  print("Prompt: " + output)

  // Send answer
  fetch("http://localhost:9999/input", {
    method = "POST",
    body = answer
  })

  sleep(0.5)
end

// Get final output
final = fetch("http://localhost:9999/").body
print("\\nFinal output:\\n" + final)
```

## Error Handling

### Script Crashes

If the script crashes:
- `GET /` returns the error message
- `GET /input` returns error
- `POST /input` receives "script crashed" error

### Network Timeouts

HTTP requests have reasonable timeouts:
- `GET /` - returns immediately
- `GET /input` - blocks up to ~60 seconds
- `POST /input` - responds immediately

### Port Already in Use

If port is in use:

```bash
# Find process using port 9999
lsof -i :9999

# Or use a different port
duso -stdin-port 9998 script.du
```

## Notes

- `-stdin-port` captures ALL stdin/stdout (including `print()`, `input()`, debug output)
- Output is accumulated and persists in memory while script runs
- When script exits, the HTTP server stops
- All output is available at `/` endpoint until script terminates
- For large scripts with lots of output, HTTP response size grows
- Useful for testing, remote interaction, and LLM agent orchestration

## See Also

- [input() - Read user input](/docs/reference/input.md)
- [breakpoint() - Pause execution for debugging](/docs/reference/breakpoint.md)
- [watch() - Conditional breakpoint with message](/docs/reference/watch.md)
- [fetch() - Make HTTP requests](/docs/reference/fetch.md)
