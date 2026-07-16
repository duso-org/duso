# regex

A compiled regular expression pattern that can be used for pattern matching.

## Creating Regex Values

Regex values are created using `~pattern~` syntax:

```duso
digit_pattern = ~\d+~
email_pattern = ~^[\w.-]+@[\w.-]+\.\w+$~
word_pattern = ~\w+~
```

For dynamic patterns (from user input or configuration), use the `toregex()` function:

```duso
user_input = input("Enter a regex pattern: ")
pattern = toregex(user_input)
```

## Using Regex Values

Regex values work with pattern matching functions:

```duso
text = "Price: 42 dollars"

// find() - Find all matches
matches = find(text, ~\d+~)
// Result: [{text: "42", pos: 7, len: 2}]

// contains() - Check if pattern exists
if contains(text, ~\d+~) then
  print("Contains a number")
end

// replace() - Replace matches
result = replace(text, ~\d+~, "X")
// Result: "Price: X dollars"

// starts_with() - Check pattern at start
if starts_with(text, ~Price~) then
  print("Starts with Price")
end

// ends_with() - Check pattern at end
if ends_with(text, ~dollars$~) then
  print("Ends with dollars")
end
```

## Type Checking

Check if a value is a regex:

```duso
value = ~\d+~
if type(value) == "regex" then
  print("This is a regex pattern")
end
```

## String Literals vs Regex Patterns

Duso distinguishes between plain strings (treated as literal text) and regex patterns:

```duso
// Plain string - matches literal text
find(text, "[1]")      // Finds the literal string "[1]"

// Regex pattern - matches the pattern
find(text, ~\[1\]~)    // Finds "[" followed by "1" followed by "]"
```

This prevents accidental regex interpretation of special characters.

## See Also

- [`toregex()` - Create regex from dynamic string](/docs/reference/toregex.md)
- [`find()` - Find pattern matches](/docs/reference/find.md)
- [`replace()` - Replace pattern matches](/docs/reference/replace.md)
- [`contains()` - Check if pattern exists](/docs/reference/contains.md)
