# String

Strings are text values. Duso provides powerful string features including templates, multiline strings, and a rich set of string functions.

## Creating Strings

Strings use double or single quotes:

```duso
s1 = "Hello, World"
s2 = 'Single quotes work too'
```

## Escape Sequences

Use escape sequences for special characters:

```duso
s = "Line 1\nLine 2"      // Newline
s = "Tab\there"            // Tab
s = "Quote=\"hi\""         // Quoted text
s = "Backslash=\\"         // Backslash
s = "Brace=\{"             // Literal brace
```

## Multiline Strings

For longer text, use triple quotes to preserve newlines:

```duso
doc = """
This is a multiline string.
Newlines are preserved.
No escaping needed!
"""
```

Both `"""..."""` and `'''...'''` work the same way.

## String Templates

Embed expressions in strings with `{{...}}` syntax:

```duso
name = "Alice"
age = 30
message = "{{name}} is {{age}} years old"
print(message)  // "Alice is 30 years old"
```

Templates evaluate any expression—arithmetic, comparisons, function calls:

```duso
nums = [1, 2, 3]
result = "Sum={{nums[0] + nums[1]}}"  // "Sum=3"
```

Templates can contain full conditional expressions:

```duso
age = 25
status = "{{age >= 18 ? "adult" : "minor"}}"  // "adult"

score = 85
grade = "Grade: {{score >= 90 ? "A" : score >= 80 ? "B" : "C"}}"  // "Grade: B"
```

Even complex if/then/else statements:

```duso
value = 42
result = """
The value is {{if value > 50 then "big" else "small" end}}
"""  // "The value is small"
```

Perfect for JSON, SQL, Markdown, and code generation:

```duso
json = """
{
  "name": "{{name}}",
  "age": {{age}},
  "status": "{{status}}"
}
"""
```

## Concatenation

Combine strings with the `+` operator:

```duso
greeting = "Hello" + ", " + "World"  // "Hello, World"
message = "Count=" + 42               // "Count=42"
```

## String Functions

Duso provides built-in functions for string manipulation:

- [`upper()`](upper.md) - Convert to uppercase
- [`lower()`](lower.md) - Convert to lowercase
- [`len()`](len.md) - Get length
- [`substr()`](substr.md) - Extract substring
- [`trim()`](trim.md) - Remove leading/trailing whitespace
- [`split()`](split.md) - Split into array
- [`join()`](join.md) - Join array into string
- [`contains()`](contains.md) - Check for substring
- [`replace()`](replace.md) - Replace all occurrences

## Truthiness

In conditions, non-empty strings are truthy:

```duso
if "hello" then print("true") end   // prints
if "" then print("true") end        // doesn't print
```

## Type Conversion

Convert to string with [`tostring()`](tostring.md):

```duso
s = tostring(42)    // "42"
s = tostring(true)  // "true"
```

