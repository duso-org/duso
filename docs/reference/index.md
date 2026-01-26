# Duso Built-in Functions Quick Reference

Run `duso -doc NAME` from a command line or `doc("NAME")` in a script for more info (eg. `doc("markdown")`).

## Strings

- `contains(str, text [, exact])` check if contains text
- `join(array, separator)` join array elements into single string
- `len(str)` number of charactes in string
- `lower(str)` convert to lowercase
- `replace(str, old, new [, exact])` replace all occurrences of old with new
- `split(str, separator)` split string into array by separator
- `substr(str, pos [, length])` get text, supports -length
- `trim(str)` remove leading and trailing whitespace
- `upper(str)` convert to uppercase

## Arrays

- `append(array, value)` add element to end of array, returns new array
- `filter(array, function)` keep only elements matching predicate
- `keys(object)` get array of all object keys
- `len(array | object | string)` get length or size
- `map(array, function)` transform each element with function
- `range(start, end [, step])` create array of numbers in sequence
- `reduce(array, function, initial_value)` combine array into single value
- `sort(array [, comparison_function])` sort array in ascending order
- `values(object)` get array of all object values

## Math

- `abs(n)` absolute value
- `ceil(n)` round up to nearest integer
- `clamp(value, min, max)` constrain value between min and max
- `floor(n)` round down to nearest integer
- `max(...ns)` find maximum value
- `min(...ns)` find minimum value
- `pow(base, exponent)` raise to power (exponentiation)
- `round(n)` round to nearest integer
- `sqrt(n)` square root

## I/O

- `http_client([options])` create HTTP client for making requests
- `http_server([config])` create HTTP server for handling requests
- `print(...args)` output values to stdout, separated by spaces

## Date & Time

- `now()` get current Unix timestamp
- `format_time(timestamp [, format])` format timestamp to string
- `parse_time(string [, format])` parse time string to timestamp

## JSON

- `format_json(value [, indent])` convert Duso value to JSON string
- `parse_json(str)` parse JSON string into Duso values

## Modules

- `context()` get runtime context (HTTP requests, spawned scripts, etc.) or nil if unavailable
- `require(module)` load module in isolated scope, return exports
- `include(filename)` execute script in current scope

## Types

- `tobool(value)` convert to boolean
- `tonumber(value)` convert to number
- `tostring(value)` convert to string
- `type(value)` get type name of variable

## Flow

- `parallel(...functions | array | object)` execute functions concurrently
- `run(script [, context])` execute script synchronously and return result (CLI-only)
- `spawn(script [, context])` run script in background goroutine (CLI-only)
- `exit(value)` exit program with optional return value
- `throw(message)` throw an error with call stack information

## Debugging

- `breakpoint([args...])` pause execution and enter debug mode (enable with DebugMode)
- `watch(expr, ...)` monitor expression values and break on changes (enable with DebugMode)

## System

**Only available in the `duso` binary.**

- `doc(str)` access documentation for modules and builtins
- `env(str)` read environment variable
- `input([prompt])` read line from stdin, optionally display prompt
- `load(filename)` read file contents as string
- `markdown(str)` render markdown to ANSI for colorful terminal display (-nocolor or $NOCOLOR will suppress this)
- `save(filename, str)` write string to file

# See Also

- [Individual function documentation](.) Detailed reference for each function
