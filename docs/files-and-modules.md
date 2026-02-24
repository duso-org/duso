# File I/O Operations

Read, write, and include files with Duso CLI.

> **Virtual Filesystems:** This guide covers standard file I/O. For information about the `/EMBED/` (embedded files) and `/STORE/` (virtual filesystem) prefixes, see the [Virtual Filesystems Guide](/docs/virtual-filesystem.md).

## load(filename)

Load the contents of a file as a string.

```duso
content = load("data.txt")
print(content)
```

**Parameters:**
- `filename` (string) - Path to file (relative to script's directory)

**Returns:**
- `string` - File contents

**Example - Read JSON:**

```duso
jsonText = load("config.json")
config = parse_json(jsonText)
print(config.host)
```

**Example - Read CSV:**

```duso
csv = load("data.csv")
lines = split(csv, "\n")

for line in lines do
    fields = split(line, ",")
    print(fields[0])
end
```

**Paths:**
- Relative paths are relative to the script's directory
- `./file.txt` - Current directory
- `../file.txt` - Parent directory
- `data/file.txt` - Subdirectory

---

## input([prompt])

Read a line of text from standard input (stdin).

```duso
name = input("What is your name? ")
print("Hello, " + name)
```

**Parameters:**
- `prompt` (string, optional) - Text to display before reading input. If not provided, reads without prompting.

**Returns:**
- `string` - The input line (with trailing newline stripped)

**Example - Interactive Script:**

```duso
print("Welcome to the survey!")
name = input("Enter your name: ")
age = tonumber(input("Enter your age: "))
city = input("Enter your city: ")

response = {
    name = name,
    age = age,
    city = city
}

save("survey.json", format_json(response))
print("Response saved!")
```

**Example - Reading Multiple Lines:**

```duso
lines = []
print("Enter lines of text (blank line to stop):")

while true do
    line = input("> ")
    if line == "" then
        break
    end
    push(lines, line)
end

save("notes.txt", join(lines, "\n"))
```

**Notes:**
- The prompt is optional - you can call `input()` without arguments
- The returned string has the trailing newline removed
- Use `tonumber()` to convert numeric input: `age = tonumber(input("Age: "))`
- Useful for interactive scripts and user prompts
- Blocks until the user enters input and presses Enter

---

## save(filename, content)

Write content to a file.

```duso
save("output.txt", "Hello, World!")
```

**Parameters:**
- `filename` (string) - Path to file
- `content` (string) - Content to write

**Returns:**
- Nothing (use in statements, not expressions)

**Example - Save JSON:**

```duso
data = {name = "Alice", age = 30}
json = format_json(data)
save("user.json", json)
```

**Example - Generate Report:**

```duso
report = """
Report: {{format_time(now())}}

Results:
- Count: {{count}}
- Average: {{average}}
- Status: {{status}}
"""

save("report.txt", report)
print("Report saved")
```

**Paths:**
- Paths are relative to script's directory
- Creates parent directories if needed (on most systems)
- Overwrites existing files

---

## list_files(pattern)

List files matching a wildcard pattern.

```duso
scripts = list_files("*.du")
for script in scripts do
  print(script)
end
```

**Parameters:**
- `pattern` (string) - Wildcard pattern to match files (relative to script's directory)

**Returns:**
- `array` - File paths matching the pattern

**Wildcard Syntax:**
- `*` - Matches any characters
- `?` - Matches exactly one character
- `**` - Not supported (recursive wildcard)

**Examples:**

```duso
// List all text files
texts = list_files("*.txt")

// List with single-character wildcard
temp = list_files("temp_?.log")

// List files in subdirectory
data = list_files("data/*.json")

// List all files
all = list_files("*")
```

**Other File Pattern Functions:**

These functions also support wildcard patterns:
- `copy_file("*.txt", "backup/")` - Copy matching files
- `move_file("*.log", "archived/")` - Move matching files
- `remove_file("*.tmp")` - Delete matching files

**See Also:**
- [list_dir() - List directory contents](/docs/reference/list_dir.md)
- [File pattern reference](/docs/reference/list_files.md)

---

## include(filename)

Load and execute another Duso script in the current environment.

```duso
include("helpers.du")
result = helper_function()
```

**Parameters:**
- `filename` (string) - Path to .du script file

**Returns:**
- Nothing (executes script, sharing variables and functions)

**Use Cases:**

### Shared Functions

`helpers.du`:
```duso
function formatMoney(amount)
    return "$" + round(amount * 100) / 100
end

function formatPercent(decimal)
    return round(decimal * 10000) / 100 + "%"
end
```

`main.du`:
```duso
include("helpers.du")

price = 19.99
taxRate = 0.08

print("Price: " + formatMoney(price))
print("Tax: " + formatPercent(taxRate))
```

### Shared Configuration

`config.du`:
```duso
settings = {
    apiUrl = "https://api.example.com",
    timeout = 30,
    debug = true
}

colors = {
    primary = "#007bff",
    danger = "#dc3545"
}
```

`app.du`:
```duso
include("config.du")

print("Connecting to " + settings.apiUrl)
print("Using color: " + colors.primary)
```

### Modular Organization

```
project/
├── main.du           (main script)
├── config.du         (configuration)
├── utils/
│   ├── string.du     (string utilities)
│   ├── math.du       (math utilities)
│   └── http.du       (HTTP utilities)
└── data.json
```

`main.du`:
```duso
include("config.du")
include("utils/string.du")
include("utils/http.du")

// Now all functions are available
response = http_get(url = settings.apiUrl)
text = http_response_body(response)
processed = string_uppercase(text)
```

---

## require(moduleName)

Load a module in an isolated scope and return its exports.

```duso
math = require("math")
result = math.add(2, 3)
```

**Parameters:**
- `moduleName` (string) - Name/path of module (searches script dir and DUSO_LIB)

**Returns:**
- The module's exported value (typically an object with functions)

**Key Differences from include():**

| Feature | `include()` | `require()` |
|---------|------------|------------|
| Scope | Current scope (shared) | Isolated scope (private) |
| Variables leak | Yes - visible in caller | No - invisible to caller |
| Exports | Returns nil | Returns last expression |
| Caching | No - re-executes every time | Yes - executed once, cached |
| Use case | Config files, helpers | Reusable libraries, APIs |

**Path Resolution:**

When you call `require("math")` or `require("utils/helpers")`, Duso searches:

1. **User-provided paths** (absolute or `~/...`)
   - `require("/usr/local/duso/math")` - Absolute path
   - `require("~/duso/lib")` - Home directory

2. **Relative to script directory**
   - `require("modules/math")` - Subdirectory

3. **DUSO_LIB environment variable**
   - `export DUSO_LIB=/usr/local/duso/lib:~/.duso/modules`
   - Searches each directory in order

4. **Extension fallback**
   - If file is not found, tries adding `.du` extension
   - `require("math")` finds `math.du`

**Module Pattern:**

A module exports its API by returning a value (last expression):

```duso
// mymath.du
function add(a, b)
  return a + b
end

function multiply(a, b)
  return a * b
end

return {
  add = add,
  multiply = multiply
}
```

**Using the module:**

```duso
math = require("mymath")
print(math.add(2, 3))        -- 5
print(math.multiply(4, 5))   -- 20
```

**Module Caching:**

Once loaded, a module is cached. Subsequent requires return the cached value without re-executing:

```duso
math = require("mymath")
math2 = require("mymath")  -- Returns cached value, doesn't re-execute
```

This means:
- Expensive initialization happens only once
- Side effects (like file I/O) happen only on first require
- All requires return the same value

**Complete Module Example:**

`utils.du`:
```duso
// Private helper (not exported)
function _normalize(value)
  return value / 100
end

// Public functions
function percentToDecimal(percent)
  return _normalize(percent)
end

function decimalToPercent(decimal)
  return decimal * 100
end

// Export public API
return {
  percentToDecimal = percentToDecimal,
  decimalToPercent = decimalToPercent
}
```

Usage:
```duso
utils = require("utils")
print(utils.percentToDecimal(50))  -- 0.5
print(utils.decimalToPercent(0.5)) -- 50
// _normalize is NOT accessible - it's private to the module
```

**Circular Dependency Detection:**

If modules have circular dependencies, Duso detects and reports them:

```duso
// a.du
require("b")

// b.du
require("a")

// Running either will error: "circular dependency detected"
```

---

## Complete Examples

### Reading and Processing CSV

```duso
// Read CSV file
csv = load("data.csv")
lines = split(csv, "\n")

// Parse header
header = split(lines[0], ",")

// Parse data rows
data = []
for i = 1, len(lines) do
    if i < len(lines) then
        row = split(lines[i], ",")
        record = {}
        for j = 0, len(header) do
            if j < len(header) then
                record[header[j]] = row[j]
            end
        end
        push(data, record)
    end
end

// Save as JSON
output = format_json(data)
save("output.json", output)

print("Converted " + len(data) + " rows")
```

### Generating Documentation

```duso
// Load template
template = load("README-template.md")

// Fill in variables
readme = """
# {{projectName}}

{{template}}

## Configuration

Last updated: {{format_time(now())}}
Version: {{version}}
"""

save("README.md", readme)
print("Generated README.md")
```

### Multi-Script Workflow

`step1-fetch.du`:
```duso
// Fetch data and save
result = {data = "fetched"}
save("step1-output.json", format_json(result))
```

`step2-process.du`:
```duso
include("step1-fetch.du")

// Load previous result
result = parse_json(load("step1-output.json"))

// Process
processed = {processed = true}
save("step2-output.json", format_json(processed))
```

`main.du`:
```duso
include("step2-process.du")

// Load final result
final = parse_json(load("step2-output.json"))
print(format_json(final))
```

---

## Setting Up DUSO_LIB

To enable module discovery beyond the script directory and current path, configure the `DUSO_LIB` environment variable:

```bash
# Single directory
export DUSO_LIB=~/.duso/modules

# Multiple directories (colon-separated on Unix, semicolon on Windows)
export DUSO_LIB=~/.duso/modules:/usr/local/duso/lib:./vendor/modules

# Create the directory structure
mkdir -p ~/.duso/modules
```

**Example: Create a shared module library**

Create `~/.duso/modules/http.du`:
```duso
function get(url)
  // Implementation
  return response
end

function post(url, data)
  // Implementation
  return response
end

return {get = get, post = post}
```

Now use it from any script:
```duso
http = require("http")   // Found via DUSO_LIB
response = http.get("https://api.example.com")
```

**DUSO_LIB Resolution Order:**

When you call `require("moduleName")`, Duso searches in this order:

1. Absolute paths or `~/...` (user-provided paths)
2. Relative to the script's directory
3. Each directory in `DUSO_LIB` (left to right)
4. Error if not found

**Best for DUSO_LIB:**
- Shared library modules used across projects
- Vendor dependencies
- Standard utilities library

**Best for local `modules/` directory:**
- Project-specific modules
- Private utilities
- Configuration modules

---

## Best Practices

### File I/O
1. **Path Organization** - Keep related files in directories
2. **Naming** - Use clear names: `config.du`, `helpers.du`, etc.
3. **Error Handling** - Check if files exist before loading
4. **UTF-8** - Files are read/written as UTF-8
5. **Relative Paths** - Scripts are always relative to script directory
6. **Documentation** - Comment what each included script does

### Modules
1. **Use `require()` for libraries** - Isolated scope prevents pollution
2. **Use `include()` for configuration** - When variables need to leak into scope
3. **Module exports** - Return an object with public functions/values
4. **Avoid circular dependencies** - Design module dependencies as DAG (directed acyclic graph)
5. **Version modules** - Consider including version info in module exports
6. **DUSO_LIB for shared code** - Put reusable libraries in ~/.duso/modules
7. **Omit .du extension** - `require("mylib")` is cleaner than `require("mylib.du")`
8. **Document module API** - Clear comments about what each export does

## See Also

- [Getting Started](/docs/cli/GETTING_STARTED.md) - Quick tutorial
- [Virtual Filesystems](/docs/virtual-filesystem.md) - /EMBED/ and /STORE/ prefixes
- [Claude Integration](/docs/cli/CLAUDE_INTEGRATION.md) - Using Claude API
- [Language Reference](/docs/learning-duso.md) - Complete language spec
