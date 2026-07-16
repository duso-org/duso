# Documentation

## Getting Started

- [Installing Duso](/docs/installing.md)
- [Learning Duso](/docs/learning-duso.md)

## Guides

- [Performance Report vs. Node, Python, Ruby](/docs/performance-report.md)
- [How to Bundle Applications](/docs/bundling-applications.md)
- [Files and Modules](/docs/files-and-modules.md)
- [Debugging Scripts](/docs/debugging-scripts.md)
- [Virtual Filesystem](/docs/virtual-filesystem.md)

## Libraries

### Standard Library (stdlib)

- [ansi](/stdlib/ansi/ansi.md) ANSI color and terminal styling utilities
- [docserver](/stdlib/docserver/docserver.md) Embedded documentation server with caching

### Community Libraries (contrib)

- [azure-ai](/contrib/azure-ai/azure.md) Azure OpenAI API for accessing GPT-4, Claude, and other hosted models
- [claude](/contrib/claude/claude.md) Anthropic Claude API integration with multi-turn conversations and tools
- [couchdb](/contrib/couchdb/couchdb.md) CouchDB database client with CRUD operations and Mango queries
- [discord](/contrib/discord/discord.md) Discord API integration with webhooks and Gateway client
- [slack](/contrib/slack/slack.md) Slack API integration with webhooks and Socket Mode client
- [deepseek](/contrib/deepseek/deepseek.md) DeepSeek LLM API idiomatic interface
- [gemini](/contrib/gemini/gemini.md) Google Gemini API with OpenAI-compatible interface
- [grok](/contrib/grok/grok.md) xAI Grok LLM API with OpenAI-compatible interface
- [groq](/contrib/groq/groq.md) Groq ultra-fast inference API
- [ollama](/contrib/ollama/ollama.md) Local LLMs through Ollama's OpenAI-compatible API
- [openai](/contrib/openai/openai.md) OpenAI API with options-based interface
- [phospher](/contrib/phospher/phospher.md) Phospher Icons SVG inline icon fetcher
- [stripe](/contrib/stripe/stripe.md) Stripe payment processing API
- [svgraph](/contrib/svgraph/svgraph.md) SVG chart and graph generation
- [zlm](/contrib/zlm/zlm.md) "Zero Language Model" for testing LLM-scale scenarios without burning tokens

## Language Keywords

### Variables

- [`var`](/docs/reference/var.md) Variable declaration
- [`raw`](/docs/reference/raw.md) Raw string/template modifier

### Syntax

- [`comments`](/docs/reference/comments.md) Using comments in code

### Logic

- [`if then elseif else end`](/docs/reference/if.md) Conditional statements
- [`end`](/docs/reference/end.md) Block terminator keyword
- [`and or not`](/docs/reference/if.md) Logical AND, OR, and NOT

### Loops

- [`for in`](/docs/reference/for.md) Loop with iteration
- [`while do`](/docs/reference/while.md) Loop while condition is true
- [`break`](/docs/reference/break.md) Exit loop early
- [`continue`](/docs/reference/continue.md) Skip to next iteration

### Functions

- [`function`](/docs/reference/function.md) Define a function
- [`return`](/docs/reference/return.md) Return from function

### Exceptions

- [`try`](/docs/reference/try.md) Try-catch error handling
- [`catch()`](/docs/reference/catch.md) Catch errors from try block
- [`throw(msg)`](/docs/reference/throw.md) Throw an error with call stack information

### Types

- [`array`](/docs/reference/array.md) Ordered list of values
- [`binary`](/docs/reference/binary.md) Immutable binary data (files, images, etc.)
- [`boolean`](/docs/reference/boolean.md) True or false
- [`code`](/docs/reference/code.md) Pre-parsed source code value
- [`error`](/docs/reference/error.md) Error value with message and stack trace
- [`nil`](/docs/reference/nil.md) Null/undefined value
- [`number`](/docs/reference/number.md) Floating-point number
- [`object`](/docs/reference/object.md) Key-value data structure
- [`regex`](/docs/reference/regex.md) Compiled regular expression pattern
- [`string`](/docs/reference/string.md) Text value

## Built-in Functions

### Strings

- [`contains(str, pattern)`](/docs/reference/contains.md) Check if contains pattern (supports regex)
- [`find(str, pattern)`](/docs/reference/find.md) Find all matches, returns array of {text, pos, len} objects (supports regex)
- [`join(array, sep)`](/docs/reference/join.md) Join array elements into single string
- [`len(value)`](/docs/reference/len.md) Get the length of arrays, objects, or strings
- [`lower(str)`](/docs/reference/lower.md) Convert to lowercase
- [`pad_left(str, width, char)`](/docs/reference/pad_left.md) Pad on the left to reach desired width
- [`pad_right(str, width, char)`](/docs/reference/pad_right.md) Pad on the right to reach desired width
- [`toregex(pattern)`](/docs/reference/toregex.md) Convert string pattern to regex (for dynamic patterns; use ~...~ syntax for static patterns)
- [`repeat(str, count)`](/docs/reference/repeat.md) Repeat string multiple times
- [`starts_with(str, prefix)`](/docs/reference/starts_with.md) Check if string starts with prefix
- [`ends_with(str, suffix)`](/docs/reference/ends_with.md) Check if string ends with suffix
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
- [`load_binary(path)`](/docs/reference/load_binary.md) Read file as immutable binary data
- [`save(path, content)`](/docs/reference/save.md) Write string to file (create/overwrite)
- [`save_binary(binary, path)`](/docs/reference/save_binary.md) Write binary data to file
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
- [`watch(path, timeout)`](/docs/reference/watch.md) Monitor file or directory for changes

### Image Processing

- [`load_image(path)`](/docs/reference/load_image.md) Load image from file
- [`save_image(image, path)`](/docs/reference/save_image.md) Save image to file
- [`scale_image(image, max_x, max_y, mode)`](/docs/reference/scale_image.md) Scale image with fit/fill/stretch modes
- [`crop_image(image, x, y, width, height)`](/docs/reference/crop_image.md) Extract rectangular region from image
- [`rotate_image(image, degrees)`](/docs/reference/rotate_image.md) Rotate by 90, 180, or 270 degrees
- [`flip_image_x(image)`](/docs/reference/flip_image_x.md) Flip horizontally (left-right mirror)
- [`flip_image_y(image)`](/docs/reference/flip_image_y.md) Flip vertically (top-bottom mirror)
- [`grayscale_image(image)`](/docs/reference/grayscale_image.md) Convert to grayscale
- [`convert_image(image, format)`](/docs/reference/convert_image.md) Convert between PNG, JPEG, and GIF formats
- [`composite_image(base, overlay, x, y, blend)`](/docs/reference/composite_image.md) Layer overlay on base with blend modes
- [`set_image_opacity(image, opacity)`](/docs/reference/set_image_opacity.md) Set absolute opacity (0.0-1.0)
- [`adjust_image_opacity(image, factor)`](/docs/reference/adjust_image_opacity.md) Multiply opacity by factor

### Network & HTTP

- [`fetch(url, options)`](/docs/reference/fetch.md) Make HTTP requests (JavaScript-style fetch API)
- [`http_server(config)`](/docs/reference/http_server.md) Create HTTP server for handling requests
- [`websocket(url, config)`](/docs/reference/websocket.md) Create WebSocket client connection

### I/O

- [`input(prompt)`](/docs/reference/input.md) Read line from stdin, optionally display prompt
- [`print(values...)`](/docs/reference/print.md) Output values to stdout, separated by spaces
- [`write(values...)`](/docs/reference/write.md) Output values to stdout without newline at the end
- [`busy(message)`](/docs/reference/busy.md) Display a loading/busy message to stderr

### Date & Time

- [`format_time(timestamp, format)`](/docs/reference/format_time.md) Format timestamp to string
- [`now()`](/docs/reference/now.md) Get current Unix timestamp in local timezone
- [`timestamp(timezone)`](/docs/reference/timestamp.md) Get current Unix timestamp in UTC or a specific timezone/offset
- [`timer()`](/docs/reference/timer.md) Get current time with sub-second precision for benchmarking
- [`parse_time(str, format)`](/docs/reference/parse_time.md) Parse time string to timestamp
- [`sleep(duration)`](/docs/reference/sleep.md) Pause execution for duration in seconds (default: 1)

### Encoding

- [`encode_base64(str|binary)`](/docs/reference/encode_base64.md) Encode string or binary to base64
- [`decode_base64(str)`](/docs/reference/decode_base64.md) Decode base64 string to binary
- [`format_csv(array, delimiter)`](/docs/reference/format_csv.md) Format array of arrays to CSV string
- [`parse_csv(str, delimiter)`](/docs/reference/parse_csv.md) Parse CSV string to array of arrays
- [`format_json(value, indent)`](/docs/reference/format_json.md) Convert value to JSON string (stringifies binary, functions, errors)
- [`parse_json(str)`](/docs/reference/parse_json.md) Parse JSON string
- [`markdown_html(text, options)`](/docs/reference/markdown_html.md) Render markdown to HTML
- [`markdown_ansi(text, theme)`](/docs/reference/markdown_ansi.md) Render markdown to ANSI terminal output with colors
- [`markdown_text(text)`](/docs/reference/markdown_text.md) Render markdown to plain text

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
- [`run(script, context)`](/docs/reference/run.md) Execute script synchronously and return result
- [`spawn(script, context)`](/docs/reference/spawn.md) Run script in background goroutine and return numeric process ID
- [`kill(pid)`](/docs/reference/kill.md) Terminate a spawned process by PID

### Debugging

- [`breakpoint(args...)`](/docs/reference/breakpoint.md) Pause execution and enter debug mode (enable with -debug)
- [`watch(exprs...)`](/docs/reference/watch.md) Monitor expression values and break on changes (enable with -debug)
- [`throw(msg)`](/docs/reference/try.md) Throw an error with call stack information

### Testing

- [`assert(condition, message)`](/docs/reference/assert.md) Check a condition and throw an error if false

### System & Data Storage

- [`sql(namespace, config)`](/docs/reference/sql.md) Create or retrieve a MySQL-compatible database connection pool
- [`datastore(namespace, config)`](/docs/reference/datastore.md) Access a named thread-safe in-memory key/value store with optional persistence
- [`sys(key)`](/docs/reference/sys.md) Access system information and CLI configuration values
- [`doc(topic)`](/docs/reference/doc.md) Access documentation for modules and builtins
- [`env(name)`](/docs/reference/env.md) Read environment variable
- [`uuid()`](/docs/reference/uuid.md) Generate RFC 9562 UUID v7 (time-ordered, sortable unique identifier)

### Security

- [`hash(algo, data)`](/docs/reference/hash.md) Compute cryptographic hash of string or binary (sha256, sha512, sha1, md5)
- [`hash_password(password, cost)`](/docs/reference/hash_password.md) Hash password with bcrypt for secure storage
- [`verify_password(password, hash)`](/docs/reference/verify_password.md) Verify password against bcrypt hash
- [`hmac(algo, data, key)`](/docs/reference/hmac.md) Compute HMAC for message authentication (sha256, sha512, sha1, md5)
- [`sign_rsa(data, private_key_pem)`](/docs/reference/sign_rsa.md) Sign data with RSA private key (SHA256-PKCS1v15)
- [`verify_rsa(data, signature, public_key_pem)`](/docs/reference/verify_rsa.md) Verify RSA signature
- [`rsa_from_jwk(n, e)`](/docs/reference/rsa_from_jwk.md) Convert JWK modulus and exponent to PEM-encoded RSA public key
- [`sign_ec(data, private_key_pem)`](/docs/reference/sign_ec.md) Sign data with EC private key (ES256, P-256 curve)
- [`verify_ec(data, signature, public_key_pem)`](/docs/reference/verify_ec.md) Verify EC signature (ES256, P-256 curve)
- [`ec_from_jwk(x, y)`](/docs/reference/ec_from_jwk.md) Convert JWK x,y coordinates to PEM-encoded EC public key (P-256)
- [`verify_ed25519(data, signature, public_key_pem)`](/docs/reference/verify_ed25519.md) Verify Ed25519 signature
