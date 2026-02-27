# Documentation

## Getting Started

- [Learning Duso](/docs/learning-duso.md)

## Guides

- [Files and Modules](/docs/files-and-modules.md)
- [Debugging Scripts](/docs/debugging-scripts.md)
- [Virtual Filesystem](/docs/virtual-filesystem.md)
- [Distribution](/docs/distribution.md)
- [Custom Distributions](/docs/custom-distributions.md)
- [Internals](/docs/internals.md)

## Libraries

### Standard Library (stdlib)

- [ansi](/stdlib/ansi/ansi.md) ANSI color and terminal styling utilities
- [markdown](/stdlib/markdown/markdown.md) Markdown parser with HTML and ANSI terminal output
- [docserver](/stdlib/docserver/docserver.md) Embedded documentation server with caching
- doccli CLI documentation viewer

### Community Libraries (contrib)

- [azure-ai](/contrib/azure-ai/azure.md) Azure OpenAI API for accessing GPT-4, Claude, and other hosted models
- [claude](/contrib/claude/claude.md) Anthropic Claude API integration with multi-turn conversations and tools
- [couchdb](/contrib/couchdb/couchdb.md) CouchDB database client with CRUD operations and Mango queries
- [deepseek](/contrib/deepseek/deepseek.md) DeepSeek LLM API idiomatic interface
- [groq](/contrib/groq/groq.md) Groq ultra-fast inference API
- [ollama](/contrib/ollama/ollama.md) Local LLMs through Ollama's OpenAI-compatible API
- [openai](/contrib/openai/openai.md) OpenAI API with options-based - [svgraph](/contrib/svgraph/svgraph.md) SVG chart and graph generation
- [zlm](/contrib/zlm/zlm.md) "Zero Language Model" for testing LLM-scale scenarios without burning tokens

## Language Keywords

- [`var`](/docs/reference/var.md) Variable declaration
- [`raw`](/docs/reference/raw.md) Raw string/template modifier

- [`if`](/docs/reference/if.md) Conditional statement
- [`then`](/docs/reference/if.md) Part of if statement
- [`else`](/docs/reference/if.md) Else branch of if statement
- [`elseif`](/docs/reference/if.md) Additional condition in if statement
- [`end`](/docs/reference/end.md) Closes function, if, while, for, try blocks

- [`for`](/docs/reference/for.md) Loop with iteration
- [`in`](/docs/reference/for.md) Part of for loop (iteration)
- [`while`](/docs/reference/while.md) Loop while condition is true
- [`do`](/docs/reference/while.md) Part of while loop (optional)
- [`break`](/docs/reference/break.md) Exit loop early
- [`continue`](/docs/reference/continue.md) Skip to next iteration

- [`function`](/docs/reference/function.md) Define a function
- [`return`](/docs/reference/return.md) Return from function

- [`try`](/docs/reference/try.md) Try-catch error handling
- [`catch()`](/docs/reference/catch.md) Catch errors from try block

- [`and`](/docs/reference/if.md) Logical AND
- [`or`](/docs/reference/if.md) Logical OR
- [`not`](/docs/reference/if.md) Logical NOT


## Built-in Functions

### Strings

- [`contains(str, pattern)`](/docs/reference/contains.md) Check if contains pattern (supports regex)
- [`find(str, pattern)`](/docs/reference/find.md) Find all matches, returns array of {text, pos, len} objects (supports regex)
- [`join(array, sep)`](/docs/reference/join.md) Join array elements into single string
- [`len(value)`](/docs/reference/len.md) Get the length of arrays, objects, or strings
- [`lower(str)`](/docs/reference/lower.md) Convert to lowercase
- [`repeat(str, count)`](/docs/reference/string.md) Repeat string multiple times
- [`replace(str, pattern, replacement)`](/docs/reference/replace.md) Replace all matches of pattern with replacement string or function result (supports regex)
- [`split(str, sep)`](/docs/reference/split.md) Split string into array by separator
- [`substr(str, pos, length)`](/docs/reference/substr.md) Get text, supports negative length
- [`template(str)`](/docs/reference/template.md) Create reusable template function from string with {{expression}} syntax
- [`trim(str)`](/docs/reference/trim.md) Remove leading and trailing whitespace
- [`upper(str)`](/docs/reference/upper.md) Convert to uppercase

### Arrays & Objects

- [`deep_copy(value)`](/docs/reference/deep_copy.md) Deep copy of arrays/objects; functions removed (safety for scope boundaries)
- [`filter(array, fn)`](/docs/reference/filter.md) Keep only elements matching predicate
- [`keys(obj)`](/docs/reference/keys.md) Get array of all object keys
- [`map(array, fn)`](/docs/reference/map.md) Transform each element with function
- [`pop(array)`](/docs/reference/pop.md) Remove and return last element
- [`push(array, values...)`](/docs/reference/push.md) Add elements to end, returns new length
- [`range(start, end, step)`](/docs/reference/range.md) Create array of numbers in sequence
- [`reduce(array, fn, init)`](/docs/reference/reduce.md) Combine array into single value
- [`shift(array)`](/docs/reference/shift.md) Remove and return first element
- [`sort(array, fn)`](/docs/reference/sort.md) Sort array in ascending order
- [`unshift(array, values...)`](/docs/reference/unshift.md) Add elements to beginning, returns new length
- [`values(obj)`](/docs/reference/values.md) Get array of all object values

### Math

- [`abs(n)`](/docs/reference/abs.md) Absolute value
- [`ceil(n)`](/docs/reference/ceil.md) Round up to nearest integer
- [`clamp(val, min, max)`](/docs/reference/clamp.md) Constrain value between min and max
- [`floor(n)`](/docs/reference/floor.md) Round down to nearest integer
- [`max(nums...)`](/docs/reference/max.md) Find maximum value
- [`min(nums...)`](/docs/reference/min.md) Find minimum value
- [`pow(base, exp)`](/docs/reference/pow.md) Raise to power (exponentiation)
- [`random()`](/docs/reference/random.md) Get random float between 0 and 1 (seeded per invocation)
- [`round(n)`](/docs/reference/round.md) Round to nearest integer
- [`sqrt(n)`](/docs/reference/sqrt.md) Square root
- [`acos(x)`](/docs/reference/acos.md) Inverse cosine (arccosine), x between -1 and 1, returns radians
- [`asin(x)`](/docs/reference/asin.md) Inverse sine (arcsine), x between -1 and 1, returns radians
- [`atan(x)`](/docs/reference/atan.md) Inverse tangent (arctangent), returns radians
- [`atan2(y, x)`](/docs/reference/atan2.md) Inverse tangent with quadrant correction, returns radians
- [`sin(angle)`](/docs/reference/sin.md) Sine of angle in radians
- [`cos(angle)`](/docs/reference/cos.md) Cosine of angle in radians
- [`tan(angle)`](/docs/reference/tan.md) Tangent of angle in radians
- [`exp(x)`](/docs/reference/exp.md) E raised to the power x
- [`log(x)`](/docs/reference/log.md) Logarithm base 10
- [`ln(x)`](/docs/reference/ln.md) Natural logarithm (base e)
- [`pi()`](/docs/reference/pi.md) Mathematical constant π (3.14159...)

### File I/O

- [`load(path)`](/docs/reference/load.md) Read file contents as string
- [`save(path, content)`](/docs/reference/save.md) Write string to file (create/overwrite)
- [`append_file(path, content)`](/docs/reference/append_file.md) Append content to file (create if needed)
- [`copy_file(src, dst)`](/docs/reference/copy_file.md) Copy file (supports /EMBED/ for embedded files)
- [`move_file(src, dst)`](/docs/reference/move_file.md) Move file from source to destination
- [`rename_file(old, new)`](/docs/reference/rename_file.md) Rename or move a file
- [`remove_file(path)`](/docs/reference/remove_file.md) Delete a file
- [`list_dir(path)`](/docs/reference/list_dir.md) List directory contents with {name, is_dir}
- [`list_files(path)`](/docs/reference/list_files.md) List files in directory recursively
- [`make_dir(path)`](/docs/reference/make_dir.md) Create directory (including parent directories)
- [`remove_dir(path)`](/docs/reference/remove_dir.md) Remove empty directory
- [`file_exists(path)`](/docs/reference/file_exists.md) Check if file or directory exists
- [`file_type(path)`](/docs/reference/file_type.md) Get file type ("file" or "directory")
- [`current_dir()`](/docs/reference/current_dir.md) Get current working directory

### Network & HTTP

- [`fetch(url, options)`](/docs/reference/fetch.md) Make HTTP requests (JavaScript-style fetch API)
- [`http_server(config)`](/docs/reference/fetch.md) Create HTTP server for handling requests

### I/O

- [`input(prompt)`](/docs/reference/input.md) Read line from stdin, optionally display prompt
- [`print(values...)`](/docs/reference/print.md) Output values to stdout, separated by spaces
- [`write(values...)`](/docs/reference/print.md) Output values to stdout without newline at the end

### Date & Time

- [`format_time(timestamp, format)`](/docs/reference/format_time.md) Format timestamp to string
- [`now()`](/docs/reference/now.md) Get current Unix timestamp
- [`parse_time(str, format)`](/docs/reference/parse_time.md) Parse time string to timestamp
- [`sleep(duration)`](/docs/reference/sleep.md) Pause execution for duration in seconds (default: 1)

### JSON

- [`format_json(value, indent)`](/docs/reference/format_json.md) Convert value to JSON string
- [`parse_json(str)`](/docs/reference/parse_json.md) Parse JSON string

### Modules

- [`include(path)`](/docs/reference/include.md) Execute script in current scope
- [`require(path)`](/docs/reference/require.md) Load module in isolated scope, return exports

### Types

- [`tobool(value)`](/docs/reference/tobool.md) Convert to boolean
- [`tonumber(value)`](/docs/reference/tonumber.md) Convert to number
- [`tostring(value)`](/docs/reference/tostring.md) Convert to string
- [`type(value)`](/docs/reference/type.md) Get type name of variable

### Flow & Concurrency

- [`context()`](/docs/reference/context.md) Get runtime context for a scripts or nil if unavailable
- [`exit(value)`](/docs/reference/exit.md) Exit script with optional return value
- [`parallel(fns)`](/docs/reference/parallel.md) Execute functions concurrently
- [`parse(source, metadata)`](/docs/reference/parse.md) Parse code string into code or error value (never throws)
- [`run(script, context)`](/docs/reference/run.md) Execute script synchronously and return result (CLI-only)
- [`spawn(script, context)`](/docs/reference/spawn.md) Run script in background goroutine and return numeric process ID (CLI-only)

### Debugging

- [`breakpoint(args...)`](/docs/reference/breakpoint.md) Pause execution and enter debug mode (enable with -debug)
- [`watch(exprs...)`](/docs/reference/watch.md) Monitor expression values and break on changes (enable with -debug)
- [`throw(msg)`](/docs/reference/try.md) Throw an error with call stack information

### System

- [`datastore(namespace, config)`](/docs/reference/datastore.md) Access a named thread-safe in-memory key/value store with optional persistence
- [`doc(topic)`](/docs/reference/doc.md) Access documentation for modules and builtins
- [`env(name)`](/docs/reference/env.md) Read environment variable
- [`uuid()`](/docs/reference/uuid.md) Generate RFC 9562 UUID v7 (time-ordered, sortable unique identifier)

## Embedding

Learn about [Embedding Duso](/docs/embedding/) in your Go applications.
