# Duso vs Lua: Key Differences

Duso draws ideas from many other scripting languages. It looks the most like lua, but there are some important differences:

## Language Features

- **String templates** Duso uses `{{expr}}` syntax for embedding expressions in strings; Lua doesn't have native templates
- **Multi-line strings** Duso's `"""..."""` syntax preserves newlines and automatically strips matching indentation; Lua uses `[[...]]` without indentation handling
- **Comments** Duso uses `//` for single-line and `/* */` for multi-line (with nesting)
- **Named function arguments** Duso supports optional `func(name = value)` syntax
- **For loops with ranges** Duso uses `for i = 1, 5 do...end` like Lua, but also supports `for item in array` for iteration
- **Object methods** Duso methods auto-bind the object as context at invocation; Lua requires explicit `self` parameter
- **Objects as constructors** Duso objects and arrays can be called to create shallow copies with optional overrides; Lua doesn't have this pattern
- **Variable scoping** Duso allows optional `var` keyword for local variables; Lua uses `local` keyword
- **Concurreny model** Is completely different, intuitively thread-safe (see `spawn()` and `run()`, and `parallel()`)
- **Integrated debugging** Duso has an integrated debugger that handles concurrent processes and can work over HTTP

## Data Types & Numbers

- **Numbers** Duso uses floating-point numbers exclusively internally
- **Arrays** Are 0-indexed in Duso (1-indexed in Lua)
- **Regular Expressions** are more standard (based on Go's) and denoted with `~...~`; lua has its own limited set and uses regular strings
- **Deep copy** Duso has `deep_copy()` built-in that safely removes functions for scope boundaries (eg. calling `spawn()` or `exit()`)

## Functional Programming

- **Higher-order functions** Both support closures and lexical scoping similarly
- **Array functions** Duso has `map()`, `filter()`, `reduce()` as builtins; Lua doesn't have these standard
- **Closure safety** Duso automatically handles closure safety in concurrent contexts

# Duso Keywords

- `if` Conditional statement
- `then` Part of if statement
- `else` Else branch of if statement
- `elseif` Additional condition in if statement
- `end` Closes function, if, while, for, try blocks
- `while` Loop while condition is true
- `do` Part of while loop (optional)
- `for` Loop with iteration
- `in` Part of for loop (iteration)
- `function` Define a function
- `return` Return from function
- `break` Exit loop early
- `continue` Skip to next iteration
- `try` Try-catch error handling
- `catch` Catch errors from try block
- `and` Logical AND
- `or` Logical OR
- `not` Logical NOT
- `var` Variable declaration
- `raw` Raw string/template modifier

## Boolean & Null Literals

- `true` Boolean true
- `false` Boolean false
- `nil` Null/nil value

## Operators

### Arithmetic Operators

- `+` Addition
- `-` Subtraction
- `*` Multiplication
- `/` Division
- `%` Modulo

### Comparison Operators

- `==` Equal
- `!=` Not Equal
- `<` Less Than
- `>` Greater Than
- `<=` Less Than or Equal
- `>=` Greater Than or Equal

### Assignment Operators

- `=` Simple assign
- `+=` Add-assign
- `-=` Subtract-assign
- `*=` Multiply-assign
- `/=` Divide-assign
- `%=` Modulo-assign

### Post-fix Increment/Decrement Operators

- `++` Increment
- `--` Decrement

## Delimiters

- `(...)` Function calls, grouping expressions
- `[...]` Array indexing, array literals
- `{...}` Object literals, code blocks
- `,` Separator for arguments, array elements
- `.` Property access
- `:` Object key-value separator
- `?` Ternary conditional operator
- `~...~` Regex pattern delimiter

# Duso Built-in Functions

Duso comes ready-to-run with its own runtime written in Go. It has a wide array of built-in functions along with a few that are hugely convenient (`datatstore()`, `http_server()`, and `fetch()`)

## Strings

- `contains(str, pattern [, ignore_case])` check if contains pattern (supports regex with ~pattern~ syntax)
- `find(str, pattern [, ignore_case])` find all matches, returns array of {text, pos, len} objects (supports regex)
- `join(array, separator)` join array elements into single string
- `len(str)` number of charactes in string
- `lower(str)` convert to lowercase
- `repeat(str, count)` repeat string multiple times
- `replace(str, pattern, replacement [, ignore_case])` replace all matches of pattern with replacement string or function result (supports regex)
- `split(str, separator)` split string into array by separator
- `substr(str, pos [, length])` get text, supports -length
- `template(str)` create reusable template function from string with {{expression}} syntax
- `trim(str)` remove leading and trailing whitespace
- `upper(str)` convert to uppercase

## Arrays & Objects

- `deep_copy(value)` deep copy of arrays/objects; functions removed (safety for scope boundaries)
- `filter(array, function)` keep only elements matching predicate
- `keys(object)` get array of all object keys
- `len(array | object | string)` get length or size
- `map(array, function)` transform each element with function
- `pop(array)` remove and return last element
- `push(array, value...)` add elements to end, returns new length
- `range(start, end [, step])` create array of numbers in sequence
- `reduce(array, function, initial_value)` combine array into single value
- `shift(array)` remove and return first element
- `sort(array [, comparison_function])` sort array in ascending order
- `unshift(array, value...)` add elements to beginning, returns new length
- `values(object)` get array of all object values

## Math

### Basic Operations

- `abs(n)` absolute value
- `ceil(n)` round up to nearest integer
- `clamp(value, min, max)` constrain value between min and max
- `floor(n)` round down to nearest integer
- `max(...ns)` find maximum value
- `min(...ns)` find minimum value
- `pow(base, exponent)` raise to power (exponentiation)
- `random()` get random float between 0 and 1 (seeded per invocation)
- `round(n)` round to nearest integer
- `sqrt(n)` square root

### Trigonometric Functions

All trigonometric functions work with angles in radians. Use `pi()` for π.

- `acos(x)` inverse cosine (arccosine), x between -1 and 1, returns radians
- `asin(x)` inverse sine (arcsine), x between -1 and 1, returns radians
- `atan(x)` inverse tangent (arctangent), returns radians
- `atan2(y, x)` inverse tangent with quadrant correction, returns radians
- `sin(angle)` sine of angle in radians
- `cos(angle)` cosine of angle in radians
- `tan(angle)` tangent of angle in radians

### Exponential & Logarithmic Functions

- `exp(x)` e raised to the power x
- `log(x)` logarithm base 10
- `ln(x)` natural logarithm (base e)
- `pi()` mathematical constant π (3.14159...)

### Utilities

- `uuid()` generate RFC 9562 UUID v7 (time-ordered, sortable unique identifier)

## I/O

- `busy(message)` display animated spinner with status message
- `fetch(url [, options])` make HTTP requests (JavaScript-style fetch API)
- `input([prompt])` read line from stdin, optionally display prompt
- `http_server([config])` create HTTP server for handling requests
- `print(...args)` output values to stdout, separated by spaces
- `write(...args)` output values to stdout without newline at the end

## File Operations

- `load(filename)` read file contents as string
- `save(filename, str)` write string to file (create/overwrite)
- `append_file(path, content)` append content to file (create if needed)
- `copy_file(src, dst)` copy file (supports /EMBED/ for embedded files)
- `move_file(src, dst)` move file from source to destination
- `rename_file(old, new)` rename or move a file
- `remove_file(path)` delete a file
- `list_dir(path)` list directory contents with {name, is_dir}
- `make_dir(path)` create directory (including parent directories)
- `remove_dir(path)` remove empty directory
- `file_exists(path)` check if file or directory exists
- `file_type(path)` get file type ("file" or "directory")
- `current_dir()` get current working directory

## Date & Time

- `format_time(timestamp [, format])` format timestamp to string
- `now()` get current Unix timestamp
- `parse_time(string [, format])` parse time string to timestamp
- `sleep([duration])` pause execution for duration in seconds (default: 1)

## JSON

- `format_json(value [, indent])` convert value to JSON string
- `parse_json(str)` parse JSON string

## Modules and Flow

- `include(filename)` execute script in current scope
- `require(module)` load module in isolated scope, return exports

## Types

- `tobool(value)` convert to boolean
- `tonumber(value)` convert to number
- `tostring(value)` convert to string
- `type(value)` get type name of variable

## Flow & Concurrency

- `context()` get runtime context for a scripts or nil if unavailable
- `exit(value)` exit script with optional return value
- `parallel(...functions | array | object)` execute functions concurrently
- `run(script [, context])` execute script synchronously and return result (CLI-only)
- `spawn(script [, context])` run script in background goroutine and return numeric process ID (CLI-only)

## Errors and Debugging

- `breakpoint([args...])` pause execution and enter debug mode (enable with `-debug`)
- `throw(message)` throw an error with call stack information
- `watch(expr, ...)` monitor expression values and break on changes (enable with `-debug`)

## System

- `datastore(namespace [, config])` access a named thread-safe in-memory key/value store with optional persistence
- `doc(str)` access documentation for modules and builtins
- `env(str)` read environment variable
- `sys(key)` access system information and CLI flags

# See Also

- `duso -read` List and browse any embedded doc or file
- `duso -doc TOPIC` for details with examples for keywords,
  built-ins, and modules (eg. ansi, markdown, claude)
- `duso -docserver` to start a local web server with all docs
- [Learning Duso](/docs/learning-duso.md) - Tutorial and examples
- [Internals](/docs/internals.md) - Deep dive into Duso's architecture
