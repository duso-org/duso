# Duso Contributed Modules

Community-contributed modules for Duso, vetted and curated by the Duso team.

Contributed modules are:
- Quality-checked and tested by the Duso team
- Licensed under MIT
- Available in custom Duso distributions
- Can be frozen and preserved if maintained by community members

## Contributing a Module

To contribute a module to the curated registry:

1. **Create a repository** on GitHub (or similar) with your module

   - Follow the naming convention: `duso-<modulename>`
   - Example: `duso-postgres`, `duso-claude`, `duso-helpers`

2. **Implement your module** as a `.du` file

   - See [examples](../stdlib/) for structure
   - Include documentation in a README

3. **License it under MIT**

   - Copy the [LICENSE](../LICENSE) file to your repo
   - Include the license header in your code

4. **Submit for review**

   - Open an issue on the Duso repository
   - Include: repo URL, module name, description, use case
   - Duso team will review for quality and standards

5. **Get accepted**

   - Once approved, we add it to the curated registry
   - Your module becomes available to include in Duso distributions

## Using Contributed Modules

If you're maintaining a custom Duso distribution, you can include contributed modules.

They work exactly like stdlib modules:

```duso
mymodule = require("mymodule")
result = mymodule.somefunction()
```

See [Custom Distributions](/docs/CUSTOM_DISTRIBUTIONS.md) for how to build Duso with contrib modules included.

## Standards

All contributed modules must:

- **Be MIT licensed** - Same license as Duso
- **Have clear documentation** - README with examples and API
- **Be production-ready** - Well-tested, no known bugs
- **Follow Duso conventions** - Use Duso idioms and patterns
- **Have a single maintainer or team** - Clear point of contact for maintenance

## Included Modules

### claude

Access Anthropic's Claude API directly from Duso scripts.

```duso
claude = require("claude")

// One-shot query
response = claude.prompt("What is Duso?")

// Multi-turn conversation
chat = claude.session(system = "You are helpful")
response1 = chat.prompt("First question")
response2 = chat.prompt("Follow-up")
chat.close()
```

See [claude/claude.md](claude/claude.md) for full documentation.

### couchdb

Simple CouchDB client with basic CRUD and Mango query support.

```duso
couchdb = require("couchdb")

db = couchdb.connect("http://localhost:5984", "duso")

// Create
db.put({_id = "user_1", name = "Alice", age = 30})

// Read
doc = db.get("user_1")

// Query with Mango
results = db.query({age = {$gt = 25}})

// Update
doc.age = 31
db.put(doc)

// Delete
db.delete("user_1", doc._rev)
```

See [couchdb/couchdb.md](couchdb/couchdb.md) for full documentation and examples.

## Examples of Contributed Modules

- **Database clients** - PostgreSQL, MySQL, MongoDB connectors
- **API integrations** - Popular API wrappers (Stripe, Twilio, etc.)
- **Utilities** - JSON processing, string manipulation, date handling
- **Domain-specific** - Machine learning helpers, data science tools, etc.

## Module Repository Structure

Your `duso-<modulename>` repository should look like:

```
duso-mymodule/
├── mymodule.du           # Main module file
├── mymodule.md           # Full documentation
├── examples/
│   ├── basic.du
│   └── readme.md
└── readme.md             # Getting started
```

## Questions?

See [CONTRIBUTING.md](/CONTRIBUTING.md) for more details on contributing to Duso.
