# postgres()

Connect to and query a PostgreSQL database with connection pooling and transaction support.

## Signature

```duso
postgres(resource_name [, config])
```

## Parameters

- `resource_name` (string) - Resource identifier for connection pooling. Subsequent calls with the same name reuse the connection pool
- `config` (optional, object) - Connection and pool configuration:
  - `host` (string) - Database host, default `"localhost"`
  - `port` (number) - Database port, default `5432`
  - `database` (string) - Database name
  - `user` (string) - Database user
  - `password` (string) - Database password
  - `pool_size` (number) - Maximum connections in pool, default `10`
  - `idle_timeout` (number) - Idle timeout in seconds
  - `connection_timeout` (number) - Connection timeout in seconds
  - `sslmode` (string) - SSL mode: `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full` (default: `disable`)
  - `sslcert` (string) - Path to client certificate file
  - `sslkey` (string) - Path to client key file
  - `sslrootcert` (string) - Path to root CA certificate file
  - `return_objects` (bool) - Default return format for queries: `true` for objects, `false` for arrays (default: `false`)

## Returns

Database connection object with methods.

## Methods

### Queries

#### query(sql, ...params [, {return_objects = bool}])

Execute a SELECT query. Returns array of rows (arrays by default, objects if `return_objects` is true).

```duso
db = postgres("mydb", {host = "localhost", database = "myapp", user = "app", password = "secret"})

// Returns array of arrays
rows = db.query("SELECT id, name FROM users WHERE active = $1", true)
for row in rows do
  print("ID: {{row[0]}}, Name: {{row[1]}}")
end

// Returns array of objects
users = db.query("SELECT id, name, email FROM users WHERE id = $1", 123, {return_objects = true})
for user in users do
  print("Name: {{user.name}}, Email: {{user.email}}")
end
```

**Parameters:**
- `sql` (string) - SQL query with `$1`, `$2`, etc. placeholders
- `...params` - Query parameters bound to placeholders
- `return_objects` (optional) - Override default format for this query

**Returns:** Array of rows (each row is array or object depending on format)

**Error handling:** Returns object `{error = "message"}` instead of throwing on SQL errors (query errors, constraint violations, etc.)

#### exec(sql, ...params)

Execute INSERT, UPDATE, DELETE, or DDL. Returns number of rows affected.

```duso
count = db.exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Alice", "alice@example.com")
print("Inserted {{count}} row(s)")

count = db.exec("UPDATE users SET active = $1 WHERE id = $2", false, 123)
print("Updated {{count}} row(s)")

count = db.exec("DELETE FROM users WHERE inactive_since < $1", "2024-01-01")
print("Deleted {{count}} row(s)")
```

**Parameters:**
- `sql` (string) - SQL statement with `$1`, `$2`, etc. placeholders
- `...params` - Parameters bound to placeholders

**Returns:** Number of rows affected (0 if no rows matched/updated/deleted)

**Error handling:** Returns object `{error = "message"}` instead of throwing on SQL errors

### Transactions

#### begin()

Start a transaction. Returns transaction object with same methods as db object, plus commit/rollback.

```duso
tx = db.begin()

// Execute operations within transaction
tx.exec("INSERT INTO ledger (account, amount) VALUES ($1, $2)", "checking", -100)
tx.exec("INSERT INTO ledger (account, amount) VALUES ($1, $2)", "savings", 100)

// Check intermediate state
rows = tx.query("SELECT SUM(amount) FROM ledger", {return_objects = true})
print("Balance: {{rows[0]}}")

// Commit or rollback
if rows[0].sum > 0 then
  tx.commit()
  print("Transaction committed")
else
  tx.rollback()
  print("Transaction rolled back")
end
```

**Returns:** Transaction object

**Transaction methods:**
- `query(sql, ...params)` - Query within transaction (same as db.query)
- `exec(sql, ...params)` - Execute within transaction (same as db.exec)
- `commit()` - Commit the transaction
- `rollback()` - Rollback the transaction

## Type Mapping

Duso's number type is float64, so SQL types are converted as follows:

| SQL Type | Duso Type | Notes |
|----------|-----------|-------|
| INTEGER, SERIAL, BIGINT | number | Converted to float64 |
| REAL, DOUBLE PRECISION | number | |
| NUMERIC, DECIMAL | number | Converted to float64 |
| VARCHAR, TEXT, CHAR | string | |
| BOOLEAN | boolean | |
| DATE, TIME, TIMESTAMP | string | ISO 8601 format |
| BYTEA | string | Converted from byte slice |
| NULL | nil | |
| JSON, JSONB | object/array | Parsed into Duso objects |

## Connection Pooling

Calls to `postgres("resource_name")` with the same resource name reuse the connection pool:

```duso
// First call creates pool
db1 = postgres("main", {host = "localhost", database = "myapp", user = "app", password = "secret"})

// Second call reuses pool
db2 = postgres("main")

// Both use same connection pool under the hood
db1.query("SELECT 1")
db2.query("SELECT 2")
```

Config can be updated on subsequent calls:

```duso
db = postgres("main", {host = "localhost", pool_size = 10})
db = postgres("main", {pool_size = 20})  // Updates pool_size for existing connection
```

## Error Handling

Query/exec operations return error objects instead of throwing for better error handling without try/catch:

```duso
// Check for error
result = db.query("SELECT * FROM invalid_table")
if result.error then
  print("Query failed: {{result.error}}")
else
  for row in result do
    print(row)
  end
end

// Connection errors throw (config validation, connection failures)
db = postgres("prod", {host = "invalid.host", database = "myapp"})  // May throw
```

## Examples

### Basic CRUD

```duso
db = postgres("mydb", {
  host = "localhost",
  database = "test",
  user = "postgres",
  password = "admin",
  sslmode = "disable"
})

// Create
db.exec("CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(100), email VARCHAR(100))")

// Insert
db.exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Alice", "alice@example.com")
db.exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Bob", "bob@example.com")

// Read
users = db.query("SELECT id, name, email FROM users ORDER BY id", {return_objects = true})
for user in users do
  print("{{user.id}}: {{user.name}} ({{user.email}})")
end

// Update
db.exec("UPDATE users SET email = $1 WHERE id = $2", "alice.new@example.com", 1)

// Delete
db.exec("DELETE FROM users WHERE id = $1", 2)
```

### Parameterized Queries

Always use parameterized queries with `$1`, `$2`, etc. to prevent SQL injection:

```duso
user_id = 123
search_term = "john"

// Safe - parameters are properly escaped
results = db.query(
  "SELECT * FROM users WHERE id = $1 AND name LIKE $2",
  user_id,
  "{{search_term}}%"
)

// Unsafe - string interpolation is vulnerable
// results = db.query("SELECT * FROM users WHERE id = {{user_id}}")
```

### Transactions with Rollback

```duso
tx = db.begin()

try
  tx.exec("INSERT INTO accounts (name, balance) VALUES ($1, $2)", "checking", 1000)
  tx.exec("INSERT INTO accounts (name, balance) VALUES ($1, $2)", "savings", 2000)
  tx.commit()
  print("Account created successfully")
catch err
  tx.rollback()
  print("Failed to create account: {{err}}")
end
```

## Performance Notes

- Connection pooling is per resource name, reused across all calls
- Prepared statements are not currently cached, queries are executed directly
- Use transactions for multi-statement operations for atomicity and consistency
- Consider increasing `pool_size` if experiencing "connection busy" errors under concurrent load

## See Also

- [`datastore()`](/docs/reference/datastore.md) - For in-memory key-value data
- [`fetch()`](/docs/reference/fetch.md) - For HTTP requests
