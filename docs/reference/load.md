# load()

Read the contents of a file as a string. Available in `duso` CLI only.

## Signature

```duso
load(filename)
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

Files are resolved relative to the directory containing the script:

```bash
# Directory structure:
# project/
#   ├── script.du
#   └── data/
#       └── input.txt

# In script.du:
content = load("data/input.txt")  // Relative to script.du directory
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

- [save() - Write files](./save.md)
- [include() - Execute scripts](./include.md)
- [parse_json() - Parse JSON strings](./parse_json.md)
- [CLI File I/O Guide](../cli/FILE_IO.md)
