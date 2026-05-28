# Valid Module

Minimal validation helpers. Just the essentials: checking object structure and format patterns.

For everything else (type checks, length bounds, comparisons), write the conditional directly—it's clearer and shorter.

## Usage

```duso
valid = require("valid")

ctx = context()
req = ctx.request()
res = ctx.response()

// Check required fields exist
if not valid.has(req.body, ["email", "password"]) then
  res.error(400, "Missing required fields")
end

// Check format
if not valid.email(req.body.email) then
  res.error(400, "Invalid email")
end

// Everything else: just write it
if len(req.body.password) < 8 then
  res.error(400, "Password too short")
end

username = req.body.username or "anonymous"
```

## Functions

### Field Validation
- `has(obj, schema)` - Check object fields/structure. Schema can be array of field names or nested object

### Format Checks
- `email(str)` - Looks like an email
- `url(str)` - Looks like a URL
- `uuid(str)` - Valid UUID characters (hex digits and dashes)

### Context-Aware Parsing
- `as_bool(str)` - Parse string as boolean (true/yes/on/1 → true, else false)
- `as_int(str)` - Parse string as integer, return nil if invalid
- `as_num(str)` - Parse string as number, return nil if invalid

## Examples

```duso
valid = require("valid")

// Flat list of required fields
if not valid.has(req.body, ["id", "name", "email"]) then
  res.error(400, "Missing required fields")
end

// Nested structure with object schema
user_data = {
  id = 123,
  name = "Alice",
  contact = {
    work_email = "alice@work.com",
    phone = "555-1234"
  }
}

if not valid.has(user_data, {id=true, name=true, contact={work_email=true}}) then
  res.error(400, "Invalid user structure")
end

// Format checks
if not valid.email(req.body.email) then
  res.error(400, "Invalid email format")
end

if not valid.url(req.body.website) then
  res.error(400, "Invalid URL")
end

// Type/length checks: just write them
if type(req.body.age) != "number" then
  res.error(400, "Age must be a number")
end

if len(req.body.password) < 8 then
  res.error(400, "Password too short")
end

// Defaults: use or operator
email = req.body.email or "unknown@example.com"
username = req.body.username or "anonymous"

// Parse from config/env/params (all come as strings)
debug = valid.as_bool(env("DEBUG"))  // env vars
port = valid.as_int(req.query.port) or 8080  // query params
timeout = valid.as_num(config.timeout) or 30  // config file

// Config from environment with defaults
db_host = env("DB_HOST") or "localhost"
db_port = valid.as_int(env("DB_PORT")) or 5432
db_ssl = valid.as_bool(env("DB_SSL"))
```
