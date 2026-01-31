# upper()

Convert a string to uppercase.

## Signature

```duso
upper(string)
```

## Parameters

- `string` (string) - The string to convert

## Returns

String converted to uppercase

## Examples

Basic conversion:

```duso
print(upper("hello"))           // "HELLO"
print(upper("Hello World"))     // "HELLO WORLD"
```

Numbers and special characters:

```duso
print(upper("abc123!@#"))       // "ABC123!@#"
```

Variable strings:

```duso
text = "duso scripting language"
title = upper(text)
print(title)                    // "DUSO SCRIPTING LANGUAGE"
```

## See Also

- [lower() - Convert to lowercase](/docs/reference/lower.md)
