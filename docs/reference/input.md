# input()

Read a line of text from standard input (user input).

## Signature

```duso
input([prompt])
```

## Parameters

- `prompt` (optional, string) - Message to display before reading input

## Returns

String containing the line read (without trailing newline)

## Examples

Simple input:

```duso
name = input()
print("Hello, " + name)
```

With prompt:

```duso
age = input("Enter your age: ")
age_num = tonumber(age)
if age_num >= 18 then
  print("You are an adult")
end
```

Multiple inputs:

```duso
print("Enter your information:")
name = input("Name: ")
email = input("Email: ")
print("Stored: " + name + " (" + email + ")")
```

## See Also

- [print() - Output text](./print.md)
