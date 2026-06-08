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
- **Code values** First-class code type created by `parse()`, can be executed or persisted
- **Error values** First-class error type with message and stack trace, returned by `parse()` or caught in try/catch
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
- `ends_with(str, suffix [, ignore_case])` check if string ends with suffix
- `find(str, pattern [, ignore_case])` find all matches, returns array of {text, pos, len} objects (supports regex)
- `join(array, separator)` join array elements into single string
- `len(str)` number of charactes in string
- `lower(str)` convert to lowercase
- `pad_left(str, width [, char])` pad on the left to reach desired width
- `pad_right(str, width [, char])` pad on the right to reach desired width
- `repeat(str, count)` repeat string multiple times
- `replace(str, pattern, replacement [, ignore_case])` replace all matches of pattern with replacement string or function result (supports regex)
- `split(str, separator)` split string into array by separator
- `starts_with(str, prefix [, ignore_case])` check if string starts with prefix
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

- [Math functions](/docs/reference/math.md) - Basic arithmetic, trigonometry, exponential/logarithmic, and utility functions.

### Utilities

- `uuid()` generate RFC 9562 UUID v7 (time-ordered, sortable unique identifier)

## I/O

- `busy(message)` display animated spinner with status message
- `input([prompt])` read line from stdin, optionally display prompt
- `print(...args)` output values to stdout, separated by spaces
- `write(...args)` output values to stdout without newline at the end

## HTTP

- [HTTP functions](/docs/reference/http.md) - Full-featured HTTP client and server with routing, CORS, JWT authentication, WebSocket support, and file uploads.

## File Operations

- [File functions](/docs/reference/file.md) - Read, write, copy, move, and monitor files and directories on disk.

## Image Processing

- [Image functions](/docs/reference/image.md) - Load, save, resize, crop, transform, and compose images with effects and blend modes.

## Date & Time

- `format_time(timestamp [, format])` format timestamp to string
- `now()` get current Unix timestamp in local timezone
- `timestamp([timezone])` get current Unix timestamp in UTC or a specific timezone/offset
- `timer()` get current time with sub-second precision for benchmarking
- `parse_time(string [, format])` parse time string to timestamp
- `sleep([duration])` pause execution for duration in seconds (default: 1)

## JSON

- `format_json(value [, indent])` convert value to JSON string (stringifies binary, functions, errors)
- `parse_json(str)` parse JSON string

## CSV

- `format_csv(array [, delimiter])` format array of arrays to CSV string
- `parse_csv(str [, delimiter])` parse CSV string to array of arrays

## Encoding

- `encode_base64(str | binary)` encode string or binary to base64
- `decode_base64(str)` decode base64 string to binary
- `markdown_html(text, options)` render markdown to HTML
- `markdown_ansi(text, theme)` render markdown to ANSI terminal output with colors

## Security

- `hash(algo, data)` compute cryptographic hash of string or binary (sha256, sha512, sha1, md5)
- `hash_password(password [, cost])` hash password with bcrypt for secure storage
- `verify_password(password, hash)` verify password against bcrypt hash
- `sign_rsa(data, private_key_pem)` sign data with RSA private key (SHA256-PKCS1v15)
- `verify_rsa(data, signature, public_key_pem)` verify RSA signature
- `rsa_from_jwk(n, e)` convert JWK modulus and exponent to PEM-encoded RSA public key
- `sign_ec(data, private_key_pem)` sign data with EC private key (ES256, P-256 curve)
- `verify_ec(data, signature, public_key_pem)` verify EC signature (ES256, P-256 curve)
- `ec_from_jwk(x, y)` convert JWK x,y coordinates to PEM-encoded EC public key (P-256)

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
- `parse(source [, metadata])` parse code string into code or error value (never throws)
- `run(script | code [, context])` execute script or code value synchronously and return result
- `spawn(script | code [, context])` run script or code value in background goroutine and return numeric process ID
- `kill(pid)` terminate a spawned process by PID

## Errors and Debugging

- `assert(condition [, message])` check a condition and throw an error if false (essential for testing)
- `breakpoint([args...])` pause execution and enter debug mode (enable with `-debug`)
- `throw(message)` throw an error with call stack information
- `watch(expr, ...)` monitor expression values and break on changes (enable with `-debug`)

## System & Data Storage

- `sql(namespace [, config])` create or retrieve a MySQL-compatible database connection pool
- `datastore(namespace [, config])` access a named thread-safe in-memory key/value store with optional persistence
- `doc(str)` access documentation for modules and builtins
- `env(str)` read environment variable
- `sys(key)` access system information and CLI flags
- `uuid()` generate RFC 9562 UUID v7 (time-ordered, sortable unique identifier)

# See Also

- `duso -read` List and browse any embedded doc or file
- `duso -doc TOPIC` for details with examples for keywords,
  built-ins, and modules (eg. ansi, markdown, claude)
- `duso -docserver` to start a local web server with all docs
- [Learning Duso](/docs/learning-duso.md) - Tutorial and examples
- [Internals](/docs/internals.md) - Deep dive into Duso's architecture
