# sql() — Database Connections

The `sql()` builtin provides thread-safe, namespaced connections to MySQL-compatible databases (MySQL, MariaDB, TiDB). Connections are pooled globally by namespace, allowing multiple scripts to share the same pool.

## Basic Usage

Create a connection:

```duso
db = sql("myapp", {
  driver = "mysql",
  host = "localhost",
  port = 3306,
  database = "mydb",
  user = "root",
  password = "secret"
})
```

Retrieve an existing connection:

```duso
db = sql("myapp")
```

Execute queries:

```duso
// SELECT
rows = db.query("SELECT id, name FROM users WHERE id = ?", [42])

// INSERT/UPDATE/DELETE
n = db.exec("UPDATE users SET seen = ? WHERE id = ?", [timestamp(), 42])

// Health check
ok = db.ping()

// Close
db.close()
```

## Configuration

Pass a config object as the second argument to `sql()`:

```duso
db = sql("namespace", {
  driver = "mysql",           // "mysql", "mariadb", or "tidb"
  host = "localhost",         // default: "localhost"
  port = 3306,                // default: 3306 for MySQL, 4000 for TiDB
  database = "mydb",          // required
  user = "root",              // default: "root"
  password = "secret",        // default: "" (no password)
  max_open_conns = 25,        // optional, default: 25
  max_idle_conns = 5,         // optional, default: 5
  conn_max_lifetime = 300,    // optional, seconds, default: 300
  dsn = "...",                // optional: raw DSN string (overrides all above)
})
```

If the second argument is omitted, `sql()` retrieves an existing connection or throws an error if not found.

## Methods

### .query(sql, [values], [return_objects])

Execute a SELECT query and return rows.

**Parameters:**
- `sql` (string, required) — SQL query with `?` placeholders
- `values` (array or value, optional) — Query parameters. If a single non-array value is passed, it's auto-wrapped in an array.
- `return_objects` (boolean, optional) — Row format (default: true). If true, each row is an object with column names as keys. If false, each row is an array of column values in order.

**Returns:** Array of rows (empty if no matches). Each row is either an object or array depending on `return_objects`.

**Named arguments supported:**
```duso
rows = db.query(query="SELECT * FROM users WHERE id = ?", values=[42], return_objects=false)
```

**Examples:**

```duso
// Objects (default)
rows = db.query("SELECT id, name, email FROM users WHERE status = ?", ["active"])
for row in rows do
  print(row.name + ": " + row.email)
end

// Arrays
rows = db.query("SELECT id, name FROM users", [], false)
for row in rows do
  print(row[0] + ": " + row[1])
end

// With named params
rows = db.query(query="SELECT * FROM users WHERE id = ?", values=[42])

// Single value auto-wrap
rows = db.query("SELECT * FROM users WHERE id = ?", 42)  // auto-wrapped to [42]
```

### .exec(sql, [values])

Execute an INSERT, UPDATE, or DELETE query.

**Parameters:**
- `sql` (string, required) — SQL query with `?` placeholders
- `values` (array or value, optional) — Query parameters. Single non-array values are auto-wrapped.

**Returns:** Number of rows affected (as float64).

**Named arguments supported:**
```duso
n = db.exec(query="UPDATE users SET seen = ? WHERE id = ?", values=[timestamp(), 42])
```

**Examples:**

```duso
// INSERT
n = db.exec("INSERT INTO users (name, email) VALUES (?, ?)", ["alice", "alice@example.com"])
print("Inserted " + n + " row")

// UPDATE (atomic in a single statement)
n = db.exec("UPDATE users SET tokens = tokens + ? WHERE id = ?", [10, user_id])

// DELETE
n = db.exec("DELETE FROM users WHERE id = ?", 42)

// With named params
db.exec(query="INSERT INTO logs (message) VALUES (?)", values=["Event occurred"])
```

### .ping()

Test the database connection.

**Returns:** `true` if the connection is alive, `false` otherwise.

```duso
if db.ping() then
  print("Connected")
else
  print("Connection failed")
end
```

### .close()

Close the connection pool. Removes it from the global registry.

```duso
db.close()
```

## Type Mapping

Duso values are converted to/from SQL types as follows:

| Duso Type | SQL Type | Behavior |
|-----------|----------|----------|
| number | INT, BIGINT, FLOAT, DOUBLE | Numbers passed as-is. Returned as float64. |
| string | VARCHAR, TEXT, CHAR | Strings passed/returned as-is. |
| boolean | BOOL, TINYINT(1) | Booleans passed/returned as-is. |
| nil | NULL | Passed/returned as nil. |
| (Unix seconds) | DATETIME, TIMESTAMP | Pass as formatted strings (see below), returned as float64 Unix seconds. |

## Timestamps and Dates

### Using BIGINT Columns for Unix Timestamps

Store Unix seconds directly as numbers—no conversion needed:

```duso
db.exec("CREATE TABLE events (id INT, created_unix BIGINT)")
db.exec("INSERT INTO events VALUES (?, ?)", [1, timestamp()])

rows = db.query("SELECT created_unix FROM events")
ts = rows[0].created_unix  // returns float64 (Unix seconds)

// Math works directly
diff = timestamp() - ts  // time since event
```

### Using DATETIME/TIMESTAMP Columns

Format timestamps as strings before inserting:

```duso
db.exec("CREATE TABLE events (id INT, created_at DATETIME)")

ts = timestamp()
db.exec("INSERT INTO events VALUES (?, ?)", [1, format_time(ts, "YYYY-MM-DD HH:mm:ss")])

rows = db.query("SELECT created_at FROM events")
ts = rows[0].created_at  // returns float64 (Unix seconds)
```

When queried back with `parseTime=true` (default), DATETIME and TIMESTAMP columns are automatically converted to float64 Unix seconds, matching Duso's timestamp type.

## UUIDs

Store UUIDs as VARCHAR columns. Duso's `uuid()` function generates UUID7, which sorts naturally by creation time:

```duso
db.exec("CREATE TABLE users (id VARCHAR(36) PRIMARY KEY, name VARCHAR(255))")

id = uuid()
db.exec("INSERT INTO users VALUES (?, ?)", [id, "alice"])

rows = db.query("SELECT id FROM users WHERE id = ?", [id])
retrieved_id = rows[0].id  // matches the original UUID
```

UUID7 sortability ensures UUIDs inserted in sequence sort in insertion order—useful for time-ordered queries:

```duso
// Retrieve users created in the last hour, sorted by creation time
cutoff = uuid_from_timestamp(timestamp() - 3600)
rows = db.query("SELECT id, name FROM users WHERE id > ? ORDER BY id", [cutoff])
```

## Error Handling

SQL errors throw exceptions that can be caught with `try/catch`:

```duso
try
  db.query("SELECT * FROM nonexistent_table")
catch (err)
  print("Query failed: " + err)
end
```

Common errors:
- Syntax errors: "Error xxxx: You have an error in your SQL syntax"
- Constraint violations: "Error xxxx: Duplicate entry"
- Connection failures: "Error xxxx: connection refused"

## Connection Pooling

Connections are pooled globally by namespace. Multiple scripts in the same process can share a pool:

```duso
// Script A
db = sql("myapp", {driver = "mysql", host = "localhost", database = "prod", user = "root"})
db.exec("INSERT INTO logs VALUES (?, ?)", [1, "hello"])

// Script B (runs in same process)
db = sql("myapp")  // retrieves the same pool from Script A
rows = db.query("SELECT * FROM logs")
```

Pool settings:
- `max_open_conns` — Maximum concurrent connections (default: 25)
- `max_idle_conns` — Idle connections to keep alive (default: 5)
- `conn_max_lifetime` — Max lifetime per connection in seconds (default: 300)

Adjust for your workload:

```duso
db = sql("myapp", {
  driver = "mysql",
  host = "localhost",
  database = "prod",
  user = "root",
  max_open_conns = 100,    // High concurrency
  max_idle_conns = 10,     // Keep more idle
  conn_max_lifetime = 600  // 10 minutes
})
```

## Atomicity

SQL statements are atomic at the database level. Use single statements for operations that must be atomic:

```duso
// ATOMIC - safe under concurrency
db.exec("UPDATE users SET tokens = tokens + ? WHERE id = ?", [amount, user_id])

// NOT atomic - race condition risk
current = db.query("SELECT tokens FROM users WHERE id = ?", [user_id])[0].tokens
db.exec("UPDATE users SET tokens = ? WHERE id = ?", [current + amount, user_id])
```

For multi-statement transactions, keep the logic in SQL (views, stored procedures, or single queries that accomplish the full operation).

## Examples

### User Authentication

```duso
db = sql("auth", {driver = "mysql", host = "localhost", database = "app", user = "root"})

// Register user
db.exec(
  "INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)",
  [uuid(), "alice", hash_password("secret123")]
)

// Verify login
rows = db.query("SELECT id, password_hash FROM users WHERE username = ?", ["alice"])
if len(rows) > 0 and verify_password("secret123", rows[0].password_hash) then
  print("Login successful")
end
```

### Event Logging with Timestamps

```duso
db = sql("logging", {driver = "mysql", host = "localhost", database = "logs", user = "root"})

// Log event with Unix timestamp
db.exec(
  "INSERT INTO events (id, user_id, action, created_unix) VALUES (?, ?, ?, ?)",
  [uuid(), user_id, "login", timestamp()]
)

// Query events from last hour
cutoff = timestamp() - 3600
rows = db.query(
  "SELECT user_id, action FROM events WHERE created_unix > ? ORDER BY created_unix DESC",
  [cutoff]
)
```

### Atomic Counter Increment

```duso
db = sql("game", {driver = "mysql", host = "localhost", database = "game", user = "root"})

// Award tokens (atomic, safe for concurrent requests)
db.exec("UPDATE players SET tokens = tokens + ? WHERE id = ?", [10, player_id])

// Deduct tokens (atomic)
db.exec("UPDATE players SET tokens = tokens - ? WHERE id = ? AND tokens >= ?", [cost, player_id, cost])
```

## See Also

- [`datastore()`](/docs/reference/datastore.md) — In-memory key/value store for process-local coordination
- [`fetch()`](/docs/reference/fetch.md) — HTTP requests (for remote APIs)
- [`try/catch`](/docs/reference/try.md) — Error handling
