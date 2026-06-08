# format_csv()

Format an array of arrays into a CSV string.

## Signature

```duso
format_csv(array [, delimiter])
```

## Parameters

- `array` (array) - Array of arrays, where each inner array is a row of fields
- `delimiter` (optional, string) - Field delimiter, default is `,`. Use `"\t"` for TSV

## Returns

CSV string with rows separated by newlines

## Examples

Format basic data:

```duso
data = [
  ["name", "age", "city"],
  ["Alice", "30", "NYC"],
  ["Bob", "25", "LA"]
]
csv = format_csv(data)
print(csv)
// Output:
// name,age,city
// Alice,30,NYC
// Bob,25,LA
```

Format as TSV:

```duso
data = [
  ["name", "age", "city"],
  ["Alice", "30", "NYC"]
]
tsv = format_csv(data, delimiter="\t")
print(tsv)
// Output:
// name	age	city
// Alice	30	NYC
```

Quote fields with commas:

```duso
data = [
  ["name", "description"],
  ["Product A", "Price: $5.00, includes tax"]
]
csv = format_csv(data)
print(csv)
// Output:
// name,description
// Product A,"Price: $5.00, includes tax"
```

Build from objects:

```duso
users = [
  {name = "Alice", age = 30},
  {name = "Bob", age = 25}
]
rows = [["name", "age"]]
for user in users do
  rows = push(rows, [user.name, tostring(user.age)])
end
csv = format_csv(rows)
save("users.csv", csv)
```

Round-trip with parse_csv:

```duso
original = """
  id,title
  1,Report
  2,Summary
"""
records = parse_csv(original)
csv = format_csv(records)
print(csv == original)          // true (round-trip preserves format)
```

## Features

- Automatically quotes fields containing delimiters or quotes
- Escapes quotes within quoted fields
- Handles special characters correctly
- Round-trip compatible with parse_csv()

## See Also

- [parse_csv() - Parse CSV strings](/docs/reference/parse_csv.md)
- [save() - Write file](/docs/reference/save.md)
