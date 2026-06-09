# load()

Read the contents of a file as a string. Available in `duso` CLI only.


`load(filename)`

```

## Parameters

- `filename` (string) - Path to the file, relative to the script's directory

## Returns

String containing the file contents

## Examples

Read a text file:

```duso
content = load("data.txt")
print(content)
```

Read and parse JSON:

```duso
config_str = load("config.json")
config = parse_json(config_str)
print(config.timeout)
```

Read multiple files:

```duso
file1 = load("input1.txt")
file2 = load("input2.txt")
combined = file1 + file2
```

## Path Resolution

Bare paths resolve against the entry script's directory (appDir); explicit prefixes (`/HERE/`, `/CWD/`, `/EMBED/`, `/STORE/`) and absolute paths use those roots directly. See [Files, Modules, and Paths](/docs/files-and-modules.md#path-roots) for the full table.

```duso
content = load("data/input.txt")    // appDir/data/input.txt
local   = load("/HERE/sibling.txt") // next to the current .du file
log_txt = load("/CWD/app.log")      // the process's working directory
```

## Working with Structured Data

Load and process JSON:
```duso
json_str = load("data.json")
data = parse_json(json_str)
for item in data do
  print(item.name)
end
```

Load and process CSV:
```duso
csv_str = load("records.csv")
lines = split(csv_str, "\n")
for line in lines do
  fields = split(line, ",")
  print(fields)
end
```

## See Also

- [Files, Modules, and Paths](/docs/files-and-modules.md) - Path roots, file operations overview
- [save() - Write files](/docs/reference/save.md)
- [include() - Execute scripts](/docs/reference/include.md)
- [parse_json() - Parse JSON strings](/docs/reference/parse_json.md)
