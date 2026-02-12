# Duso Language Reference

Complete reference of all keywords, operators, and special syntax in the Duso language.

## Keywords

| Keyword    | Purpose                                      |
|------------|----------------------------------------------|
| `if`       | Conditional statement                        |
| `then`     | Part of if statement                         |
| `else`     | Else branch of if statement                  |
| `elseif`   | Additional condition in if statement         |
| `end`      | Closes function, if, while, for, try blocks |
| `while`    | Loop while condition is true                 |
| `do`       | Part of while loop (optional)                |
| `for`      | Loop with iteration                          |
| `in`       | Part of for loop (iteration)                 |
| `function` | Define a function                            |
| `return`   | Return from function                         |
| `break`    | Exit loop early                              |
| `continue` | Skip to next iteration                       |
| `try`      | Try-catch error handling                     |
| `catch`    | Catch errors from try block                  |
| `and`      | Logical AND                                  |
| `or`       | Logical OR                                   |
| `not`      | Logical NOT                                  |
| `var`      | Variable declaration                         |
| `raw`      | Raw string/template modifier                 |

## Boolean & Null Literals

| Literal  | Type              |
|----------|-------------------|
| `true`   | Boolean true      |
| `false`  | Boolean false     |
| `nil`    | Null/nil value    |

## Operators

### Arithmetic Operators

| Operator       | Symbol | Example   |
|----------------|--------|-----------|
| Addition       | `+`    | `a + b`   |
| Subtraction    | `-`    | `a - b`   |
| Multiplication | `*`    | `a * b`   |
| Division       | `/`    | `a / b`   |
| Modulo         | `%`    | `a % b`   |

### Comparison Operators

| Operator               | Symbol |
|------------------------|--------|
| Equal                  | `==`   |
| Not Equal              | `!=`   |
| Less Than              | `<`    |
| Greater Than           | `>`    |
| Less Than or Equal     | `<=`   |
| Greater Than or Equal  | `>=`   |

### Assignment Operators

| Operator          | Symbol | Meaning              |
|-------------------|--------|----------------------|
| Simple assign     | `=`    | `x = 5`              |
| Add-assign        | `+=`   | `x += 5` (x = x + 5) |
| Subtract-assign   | `-=`   | `x -= 5` (x = x - 5) |
| Multiply-assign   | `*=`   | `x *= 5` (x = x * 5) |
| Divide-assign     | `/=`   | `x /= 5` (x = x / 5) |
| Modulo-assign     | `%=`   | `x %= 5` (x = x % 5) |

### Post-fix Increment/Decrement Operators

| Operator  | Symbol | Example |
|-----------|--------|---------|
| Increment | `++`   | `x++`   |
| Decrement | `--`   | `x--`   |

## Delimiters

| Delimiter   | Purpose                                          |
|-------------|--------------------------------------------------|
| `(` `)`     | Function calls, grouping expressions             |
| `[` `]`     | Array indexing, array literals                   |
| `{` `}`     | Object literals, code blocks                     |
| `,`         | Separator for arguments, array elements          |
| `.`         | Property access                                  |
| `:`         | Object key-value separator                       |
| `?`         | Ternary conditional operator                     |
| `~` `~`     | Regex pattern delimiter                          |

## Summary

- **18 Keywords** — Control flow, declarations, and logic operators
- **3 Literals** — Boolean and null values
- **5 Arithmetic Operators** — Basic math operations
- **6 Comparison Operators** — Value comparisons
- **6 Assignment Operators** — Variable assignments with operations
- **2 Post-fix Operators** — Increment and decrement
- **8 Delimiters** — Syntax structure and grouping

## Examples

### Control Flow
```duso
if x > 5 then
    print("x is greater than 5")
elseif x == 5 then
    print("x equals 5")
else
    print("x is less than 5")
end
```

### Loops
```duso
-- While loop
while x < 10 do
    x += 1
end

-- For loop
for i = 1, 10 do
    print(i)
end

-- For-in loop
for item in list do
    print(item)
end
```

### Functions
```duso
function add(a, b)
    return a + b
end

result = add(3, 5)
```

### Error Handling
```duso
try
    result = risky_operation()
catch e
    print("Error: " .. e)
end
```

### Operators in Action
```duso
-- Arithmetic
sum = 10 + 5      -- 15
diff = 10 - 5     -- 5
prod = 10 * 5     -- 50
quot = 10 / 5     -- 2
remainder = 10 % 3  -- 1

-- Comparison
if a == b and c != d or not e then
    -- do something
end

-- Compound assignment
x = 10
x += 5    -- x is now 15
x *= 2    -- x is now 30

-- Increment/Decrement
count = 0
count++   -- count is now 1
count--   -- count is back to 0
```
