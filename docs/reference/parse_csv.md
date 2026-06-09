# parse_csv()

Parse a CSV string into an array of arrays.


`parse_csv(str [, delimiter])`

```

## Parameters

- `str` (string) - A CSV string
- `delimiter` (optional, string) - Field delimiter, default is `,`. Use `"\t"` for TSV

## Returns

Array of arrays, where each inner array is a row of fields

## Examples

Parse basic CSV:

```duso
csv = """
  name,age,city
  Alice,30,NYC
  Bob,25,LA
"""
records = parse_csv(csv)
print(records[0])               // ["name", "age", "city"]
print(records[1][0])            // "Alice"
```

Parse with quoted fields:

```duso
csv = """
  name,description
  Alice,"Hello, world"
  Bob,"Test, data"
"""
records = parse_csv(csv)
print(records[1][1])            // "Hello, world" (comma preserved)
```

Parse TSV (tab-separated):

```duso
tsv = """
  name\tage\tcity
  Alice\t30\tNYC
"""
records = parse_csv(tsv, delimiter="\t")
print(records[0])               // ["name", "age", "city"]
```

Process records with string templates:

```duso
csv = load("users.csv")
records = parse_csv(csv)
for i = 1, len(records) - 1 do
  row = records[i]
  print("{{row[0]}} is {{row[1]}} years old")
end
```

## Features

- Correctly handles quoted fields with embedded delimiters
- Supports escaped quotes within quoted fields
- Handles newlines within quoted fields
- Returns empty array for empty input
- Works with any single-character delimiter

## See Also

- [format_csv() - Format arrays to CSV](/docs/reference/format_csv.md)
- [load() - Read files](/docs/reference/load.md)
