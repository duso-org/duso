# File I/O Operations

Read, write, and include files with Duso CLI.

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
        data = append(data, record)
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

## Best Practices

1. **Path Organization** - Keep related files in directories
2. **Naming** - Use clear names: `config.du`, `helpers.du`, etc.
3. **Error Handling** - Check if files exist before loading
4. **UTF-8** - Files are read/written as UTF-8
5. **Relative Paths** - Scripts are always relative to script directory
6. **Documentation** - Comment what each included script does

## See Also

- [Getting Started](GETTING_STARTED.md) - Quick tutorial
- [Claude Integration](CLAUDE_INTEGRATION.md) - Using Claude API
- [Language Reference](../language-spec.md) - Complete language spec
