# toregex()

Convert a string pattern into a regex value for pattern matching (advanced/dynamic use only).

`toregex(pattern)`

## Parameters

- `pattern` (string) - A regular expression pattern to compile

## Returns

A regex value that can be used with `find()`, `replace()`, `contains()`, `starts_with()`, and `ends_with()`.

## Description

**Most of the time, use `~pattern~` literals instead.** This function is only needed for dynamic patterns.

Duso has two ways to create regex values:

### Preferred: Static patterns with `~...~` syntax

Use `~pattern~` for patterns you know at write time. These are compiled at parse time and are the recommended approach:

```duso
text = "Phone: 555-1234"

// Find digits
matches = find(text, ~\d+~)
print("Found:", matches[0].text)

// Replace word boundaries
result = replace(text, ~\b\d+\b~, "XXXX")
print("Masked:", result)
```

### Advanced: Dynamic patterns with `toregex()`

Only use `toregex()` when the pattern comes from user input, a database, configuration, or other runtime source:

```duso
// Pattern from user input
pattern_str = input("Enter regex pattern: ")
pattern = toregex(pattern_str)
```

The `toregex()` function compiles the string at runtime, which is slower but necessary when the pattern isn't known until execution time.

## Examples

**Use `~...~` literals for normal pattern matching:**

```duso
text = "Email: user@example.com"

// Find email-like pattern
matches = find(text, ~\w+@\w+\.\w+~)
print("Found:", matches[0].text)

// Replace numbers
result = replace(text, ~\d+~, "X")

// Check if contains pattern
if contains(text, ~@~) then
  print("Contains @")
end
```

**Use `toregex()` for dynamic patterns (user input, configuration, etc.):**

```duso
// Get pattern from user
user_pattern = input("Enter a regex pattern: ")
pattern = toregex(user_pattern)

// Use the pattern
text = "Find numbers: 123 and 456"
matches = find(text, pattern)
print("Found", len(matches), "matches")
```

**Storing patterns from configuration:**

```duso
// Patterns loaded from config or database (not known at write time)
patterns = {
  phone = toregex("\\d{3}-\\d{3}-\\d{4}"),
  zipcode = toregex("\\d{5}(-\\d{4})?")
}

text = "Call 555-123-4567 or visit 90210"

// Use stored patterns
for name in keys(patterns) do
  if contains(text, patterns[name]) then
    print("Found", name, "format")
  end
end
```

## See Also

- [regex - Regex data type](/docs/reference/regex.md)
- [find() - Find pattern matches](/docs/reference/find.md)
- [replace() - Replace pattern matches](/docs/reference/replace.md)
- [contains() - Check if pattern exists](/docs/reference/contains.md)
