# CouchDB Module for Duso

A simple CouchDB client for Duso supporting basic CRUD operations and Mango queries. No authentication yet.

## Quick Start

```duso
couchdb = require("couchdb")

// Connect to database
db = couchdb.connect("http://localhost:5984", "duso")

// Create a document
db.put({_id = "doc1", name = "Alice", age = 30})

// Read it back
doc = db.get("doc1")
print(doc)

// Query with Mango
results = db.query({age = {$gt = 25}})
print(results)
```

## Prerequisites

- CouchDB running locally or accessible at a URL
- Default: `http://localhost:5984`
- Start locally: `brew services start couchdb` or `docker run -d -p 5984:5984 couchdb`

## API

### Top-Level Functions

#### `couchdb.connect(url, db_name)`

Create a connection object to a database.

**Arguments:**
- `url` (string, optional): CouchDB server URL. Defaults to `"http://localhost:5984"`
- `db_name` (string, required): Database name

**Returns:** Connection object with CRUD methods

**Example:**
```duso
db = couchdb.connect("http://localhost:5984", "my_db")
```

#### `couchdb.create_database(url, db_name)`

Create a new database and return a connection to it.

**Arguments:**
- `url` (string, optional): CouchDB server URL
- `db_name` (string, required): Name for new database

**Returns:** Connection object

**Example:**
```duso
db = couchdb.create_database("http://localhost:5984", "new_db")
```

#### `couchdb.list_databases(url)`

List all databases on the server.

**Arguments:**
- `url` (string, optional): CouchDB server URL

**Returns:** Array of database names

**Example:**
```duso
dbs = couchdb.list_databases("http://localhost:5984")
print(dbs)  // ["_replicator", "_users", "my_db", ...]
```

#### `couchdb.server_info(url)`

Get server information.

**Arguments:**
- `url` (string, optional): CouchDB server URL

**Returns:** Server info object

**Example:**
```duso
info = couchdb.server_info("http://localhost:5984")
print(info.couchdb)  // "Welcome"
```

### Connection Methods

Once you have a connection object, use these methods:

#### `db.get(doc_id)`

Retrieve a single document by ID.

**Arguments:**
- `doc_id` (string): Document ID

**Returns:** Document object with `_id` and `_rev` fields

**Throws:** Error if document not found

**Example:**
```duso
doc = db.get("user_001")
print(doc.name)
```

#### `db.put(doc)`

Create or update a document.

**Arguments:**
- `doc` (object): Document with required `_id` field. Include `_rev` if updating existing document.

**Returns:** Result object with new `_rev`

**Note:** CouchDB uses optimistic locking. To update, you must:
1. Get the document
2. Modify it
3. Put it back (it includes `_rev` from the get)

**Example:**
```duso
// Create new
db.put({_id = "user_001", name = "Alice", age = 30})

// Update existing
doc = db.get("user_001")
doc.age = 31
db.put(doc)  // Includes _rev from get
```

#### `db.delete(doc_id, rev)`

Delete a document.

**Arguments:**
- `doc_id` (string): Document ID
- `rev` (string): Current revision (from `_rev` field)

**Returns:** Result object

**Note:** You must get the document first to obtain its `_rev`:

**Example:**
```duso
doc = db.get("user_001")
db.delete("user_001", doc._rev)
```

#### `db.query(selector, options)`

Query documents using Mango selector language.

**Arguments:**
- `selector` (object): Mango query selector
- `options` (object, optional):
  - `sort` (array): Sort fields. Example: `[{field = "asc"}]`
  - `limit` (number): Max documents to return
  - `skip` (number): Number of documents to skip
  - `fields` (array): Only return these fields

**Returns:** Array of matching documents

**Mango Selector Syntax:**
```duso
// Simple equality
{name = "Alice"}

// Comparison operators (use quoted keys for $ operators)
{age = {"$gt" = 30}}        // greater than
{age = {"$gte" = 30}}       // greater than or equal
{age = {"$lt" = 30}}        // less than
{age = {"$lte" = 30}}       // less than or equal
{age = {"$eq" = 30}}        // equal
{age = {"$ne" = 30}}        // not equal

// Array operators
{tags = {"$in" = ["red", "blue"]}}     // any value in array
{tags = {"$nin" = ["red"]}}            // none of values
{tags = {"$all" = ["red", "blue"]}}    // all values present

// Logical operators
{"$and" = [{name = "Alice"}, {age = {"$gt" = 25}}]}
{"$or" = [{name = "Alice"}, {name = "Bob"}]}
{"$nor" = [{status = "deleted"}]}
{"$not" = {status = "inactive"}}

// Complex queries
{
  "$and" = [
    {age = {"$gte" = 25}},
    {age = {"$lt" = 65}},
    {status = "active"}
  ]
}
```

**Example:**
```duso
// Find users over 30
results = db.query({age = {"$gt" = 30}})

// With sorting and limit
results = db.query(
  {age = {"$gt" = 30}},
  {sort = [{age = "desc"}], limit = 10}
)

// Only specific fields
results = db.query(
  {status = "active"},
  {fields = ["name", "email"]}
)
```

#### `db.create_index(fields, index_name)`

Create an index to speed up queries.

**Arguments:**
- `fields` (array): Array of field objects to index. Example: `[{age = "asc"}]`
- `index_name` (string, optional): Index name. Defaults to `"auto"`

**Returns:** Index creation result

**Note:** For best query performance, create indexes on fields you frequently query.

**Example:**
```duso
// Create index on age field
db.create_index([{age = "asc"}])

// Compound index on multiple fields
db.create_index([{status = "asc"}, {created = "asc"}])
```

#### `db.bulk(docs)`

Insert or update multiple documents at once.

**Arguments:**
- `docs` (array): Array of document objects (each must have `_id`)

**Returns:** Array of results (one per document)

**Example:**
```duso
docs = [
  {_id = "doc1", name = "Alice"},
  {_id = "doc2", name = "Bob"},
  {_id = "doc3", name = "Charlie"}
]
results = db.bulk(docs)
```

#### `db.info()`

Get database statistics and information.

**Returns:** Info object with document count, data size, etc.

**Example:**
```duso
info = db.info()
print(info.doc_count)      // Number of documents
print(info.data_size)      // Size in bytes
```

#### `db.delete_db()`

Delete the entire database.

**Warning:** This is destructive and irreversible.

**Returns:** Deletion result

**Example:**
```duso
db.delete_db()
```

## Error Handling

All operations throw errors on failure:

```duso
try
  db.put({_id = "doc1", name = "Alice"})
catch (e)
  print("Error: " + e)
end
```

Common errors:
- `404 not_found`: Document or database doesn't exist
- `409 conflict`: Document revision mismatch (use fresh `_rev`)
- `400 bad_request`: Invalid query or bad selector syntax

## Examples

See `examples/basic.du` for a complete working example.

### Session Pattern (Multiple Operations)

```duso
couchdb = require("couchdb")
db = couchdb.connect("http://localhost:5984", "app_db")

// Create some documents
for i = 1, 5 do
  db.put({
    _id = "user_" + i,
    name = "User " + i,
    created = now()
  })
end

// Query them
users = db.query({}, {limit = 10})
for user in users do
  print(user.name)
end
```

### Updating Workflow

```duso
// Get current state
doc = db.get("doc_id")

// Modify
doc.field = "new value"
doc.updated_at = now()

// Save back (includes _rev from get)
result = db.put(doc)

// If conflict, fetch fresh and retry
if result.error == "conflict" then
  doc = db.get("doc_id")  // Get latest
  doc.field = "new value"  // Reapply changes
  db.put(doc)
end
```

## Limitations & Future Work

- **No authentication** yet (basic auth planned)
- **No attachments** support
- **No views** (only Mango queries)
- **No replication** primitives
- **No change feeds**

## Testing

Run the example:

```bash
duso contrib/couchdb/examples/basic.du
```

Requires CouchDB running locally on port 5984.
