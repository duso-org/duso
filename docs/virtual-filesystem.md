# Virtual Filesystems: /EMBED/ and /STORE/

Duso provides two virtual filesystem prefixes that allow you to work with files without direct filesystem access. These are useful for sandboxing, embedding resources, and managing generated code.

## Overview

| Prefix | Purpose | Access | Persistence |
|--------|---------|--------|-------------|
| `/EMBED/` | Read embedded resources (stdlib, docs, etc.) | Read-only | N/A (built into binary) |
| `/STORE/` | Create and manage scripts at runtime | Read/write | In-memory (optional persistence) |

## /EMBED/ - Embedded Files

The `/EMBED/` prefix provides access to files embedded in the Duso binary during compilation. This includes:

- **stdlib/** - Standard library modules
- **docs/** - Documentation files
- **contrib/** - Contributed modules
- **examples/** - Example scripts
- **README.md** - Project README

### Reading Embedded Files

Use `load()` to read embedded files:

```duso
-- Load and display documentation
readme = load("/EMBED/README.md")
print(readme)

-- Load a standard library module
math_docs = load("/EMBED/docs/reference/math.md")

-- Use in other functions
copy_file("/EMBED/stdlib/http.du", "local_http.du")
```

### Module Resolution

When using `require()` or `include()`, the system automatically searches embedded files:

```duso
-- These all search /EMBED/stdlib automatically
http = require("http")
markdown = require("markdown")

-- This loads from /EMBED/docs/reference/print.md
doc_content = load("/EMBED/docs/reference/print.md")
```

### Use Cases

- Distributing pre-built libraries with your Duso binary
- Embedding documentation for `doc()` function
- Shipping default configuration files
- Storing resource files (templates, schemas, etc.)

## /STORE/ - Virtual Filesystem

The `/STORE/` prefix provides a virtual filesystem backed by the in-memory datastore. This allows scripts to create, modify, and execute code at runtime without touching the real filesystem.

### Creating and Managing Files

Use standard file operations with the `/STORE/` prefix:

```duso
-- Save a file
save("/STORE/myfile.txt", "Hello, World!")

-- Load a file
content = load("/STORE/myfile.txt")
print(content)  -- Hello, World!

-- Check if file exists
if file_exists("/STORE/myfile.txt")
  print("File exists!")
end

-- Append to a file
append_file("/STORE/myfile.txt", "\nLine 2")

-- Copy files
copy_file("/STORE/myfile.txt", "/STORE/backup.txt")

-- Move/rename files
move_file("/STORE/myfile.txt", "/STORE/renamed.txt")

-- Delete files
remove_file("/STORE/renamed.txt")
```

### Generating and Executing Scripts

A key use case for `/STORE/` is generating scripts at runtime and executing them:

```duso
-- Generate a script dynamically
code = """
  x = 10
  y = 20
  print("Sum:", x + y)
  exit(x + y)
"""

-- Save it to /STORE/
save("/STORE/generated.du", code)

-- Execute it
result = run("/STORE/generated.du")
print("Result:", result)  -- Result: 30
```

### Modules from /STORE/

You can create reusable modules in `/STORE/` and load them with `require()` or `include()`:

```duso
-- Create a helper module
save("/STORE/helpers.du", """
  {
    double = function(x)
      return x * 2
    end,

    square = function(x)
      return x * x
    end
  }
""")

-- Use it
helpers = require("/STORE/helpers")
print(helpers.double(5))   -- 10
print(helpers.square(4))   -- 16
```

### Working with Arrays in /STORE/

You can store complex data structures in `/STORE/` using the datastore, then access them through `/STORE/`:

```duso
-- Store configuration as JSON
config = {
  host = "localhost",
  port = 8080,
  debug = true
}
save("/STORE/config.json", format_json(config))

-- Load and parse configuration
loaded_config = parse_json(load("/STORE/config.json"))
print("Server:", loaded_config.host)  -- Server: localhost
```

## Search Order for Module Resolution

When using `require()` or `include()`, Duso searches for modules in this order:

1. Local filesystem (relative to script directory)
2. `/STORE/` virtual filesystem
3. `/EMBED/stdlib/` embedded stdlib
4. `/EMBED/contrib/` embedded contrib modules

Example:

```duso
-- If "helpers.du" exists in multiple locations, this priority applies:
helpers = require("helpers")
-- Searches in order:
-- 1. ./helpers.du
-- 2. /STORE/helpers.du
-- 3. /EMBED/stdlib/helpers.du
-- 4. /EMBED/contrib/helpers.du
```

## Security: The -no-files Flag

When executing untrusted code, use the `-no-files` flag to restrict filesystem access to virtual filesystems only:

```bash
# Restrict to /STORE/ and /EMBED/ only
duso -no-files agent-script.du
```

With `-no-files` enabled:

✅ **Allowed:**
- Read from `/EMBED/` (embedded files)
- Read from and write to `/STORE/` (virtual filesystem)
- All other Duso operations

❌ **Blocked:**
- Read from real filesystem
- Write to real filesystem
- Delete from real filesystem
- Any path that isn't `/STORE/` or `/EMBED/`

### Example: Safe Sandbox

```duso
-- This script can safely generate and execute code
-- without accessing the real filesystem

script_template = """
  print("Running generated script")
  x = {value = 42}
  exit(x)
"""

save("/STORE/generated.du", script_template)
result = run("/STORE/generated.du")
print("Generated script returned:", result)
```

Run with: `duso -no-files safe-sandbox.du`

## Datastore Backing

`/STORE/` is backed by the datastore system. Each file is stored as a key-value pair in the "vfs" namespace:

```duso
-- Direct datastore access
store = datastore("vfs")
store.set("myfile.txt", "Hello")
value = store.get("myfile.txt")

-- Equivalent to:
save("/STORE/myfile.txt", "Hello")
content = load("/STORE/myfile.txt")
```

### Persistence

By default, `/STORE/` is in-memory and lost when the script exits. For persistence, configure the datastore:

```duso
-- Save to disk periodically
store = datastore("vfs", {
  persist = "data.json",
  persist_interval = 60  -- Save every 60 seconds
})

store.set("persistent_data", {timestamp = "2024-01-01"})
-- Data is automatically saved to data.json
```

## Common Patterns

### Template Engine

```duso
-- Store templates in /STORE/
save("/STORE/email.template", """
  Subject: {{subject}}

  Dear {{name}},
  {{body}}
""")

-- Render templates dynamically
render_template = function(template_name, data)
  template = load("/STORE/" .. template_name)
  -- Simple string substitution
  for key, value in pairs(data)
    template = string.gsub(template, "{{" .. key .. "}}", value)
  end
  return template
end

result = render_template("email.template", {
  subject = "Hello",
  name = "Alice",
  body = "This is the email body"
})
print(result)
```

### Code Generation

```duso
-- Generate a function dynamically
generate_adder = function(x)
  code = """
    function(n)
      return n + """ .. x .. """
    end
  """
  save("/STORE/adder_" .. x .. ".du", code)
  return load("/STORE/adder_" .. x .. ".du")
end

-- Create specialized functions
add_10_fn = require("/STORE/adder_10")
result = add_10_fn(5)
print(result)  -- 15
```

### Logging to Virtual Filesystem

```duso
log = function(level, message)
  log_entry = format_json({
    timestamp = os.time(),
    level = level,
    message = message
  })
  append_file("/STORE/app.log", log_entry .. "\n")
end

log("INFO", "Application started")
log("ERROR", "Something went wrong")

-- Read logs
logs = load("/STORE/app.log")
print(logs)
```

## Important Notes

### Array Handling in Datastore

When working with arrays in `/STORE/` via the datastore:

```duso
store = datastore("vfs")

-- Initialize array
store.set("items", [])

-- Add items to array
store.push("items", "first")
store.push("items", "second")

-- Retrieve array
items = store.get("items")
print(items)  -- [first, second]
```

### Path Normalization

Paths in `/STORE/` are normalized:

```duso
-- These all refer to the same file
save("/STORE/myfile.txt", "content")
load("/STORE/myfile.txt")  -- Returns the same content

-- Directory structure is implicit (no mkdir needed)
save("/STORE/data/config.json", "{}")  -- Creates directory implicitly
```

### Deep Copying

Values stored in `/STORE/` are deep copied to prevent unintended mutations:

```duso
original = {value = 10}
save("/STORE/data.json", format_json(original))

-- Modifying original doesn't affect stored data
original.value = 20
stored = parse_json(load("/STORE/data.json"))
print(stored.value)  -- 10 (unchanged)
```

## See Also

- [File I/O Operations](/docs/cli/FILE_IO.md) - Detailed reference for load(), save(), etc.
- [Datastore Guide](/docs/reference/datastore.md) - Advanced datastore operations
- [Learning Duso](/docs/learning-duso.md) - Introduction to Duso concepts
- [-no-files Flag](/docs/cli/GETTING_STARTED.md) - Security and sandboxing
