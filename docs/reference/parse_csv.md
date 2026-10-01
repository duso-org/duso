# parse_csv()

Parse a CSV string into an array of arrays.

`parse_csv(str [, delimiter] [, quotes])`

## Parameters

- `str` (string) - A CSV string
- `delimiter` (optional, string) - Field delimiter, default is `,`. Use `"\t"` for TSV
- `quotes` (optional, boolean) - Honor CSV quoting, default `true`. Set `quotes = false` to split plainly on newlines and the delimiter, treating `"` as an ordinary character

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

Parse plain TSV whose fields contain `"` characters:

```duso
tsv = "item\tsize\nhose\t5\" pipe"
records = parse_csv(tsv, "\t", quotes = false)
print(records[1][1])            // 5" pipe
```

With the default `quotes = true`, that input is a parse error (`bare " in non-quoted-field`), because CSV quoting rules apply whatever the delimiter is. Plain TSV has no quoting, so use `quotes = false` for data from database exports and similar tools. Spreadsheet TSV exports that quote fields need the default.

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

- Correctly handles quoted fields with embedded delimiters (default mode)
- Supports escaped quotes within quoted fields (default mode)
- Handles newlines within quoted fields (default mode)
- Returns empty array for empty input
- Works with any single-character delimiter
- `quotes = false` mode: no quote handling, `\r\n` line endings accepted, blank lines skipped; a field can't contain the delimiter or a newline

## See Also

- [format_csv() - Format arrays to CSV](/docs/reference/format_csv.md)
- [load() - Read files](/docs/reference/load.md)
