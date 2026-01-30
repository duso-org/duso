# Duso Documentation

## Quick Lookup: Built-in Docs

Duso has comprehensive built-in documentation right in the CLI. Look up built-in functions and modules instantly:

```bash
duso -doc                # Index of all built-ins and modules
duso -doc string         # String functions (len, substr, upper, split, etc.)
duso -doc array          # Array functions (append, map, filter, sort, etc.)
duso -doc spawn          # How spawn() works for background execution
duso -doc parallel       # Parallel execution with parallel()
duso -doc claude         # Claude module for AI integration
duso -doc http           # HTTP module for requests
duso -doc datastore      # Thread-safe coordination primitive
```

No website hunting. No digging through files. Documentation is always one command away.

---

## Documentation by Audience

Choose your learning path below:

## 🔨 For Go Developers (Embedding Duso)

You want to embed Duso as a scripting or configuration layer in your Go applications.

**Start here:**
- [**Embedding Guide**](embedding/README.md) - Overview of embedding Duso in Go
- [**Getting Started**](embedding/GETTING_STARTED.md) - Quick tutorial with minimal example
- [**API Reference**](embedding/API_REFERENCE.md) - Go API documentation
- [**Custom Functions**](embedding/CUSTOM_FUNCTIONS.md) - Register Go functions in Duso
- [**Patterns**](embedding/PATTERNS.md) - Common use cases and design patterns
- [**Examples**](embedding/EXAMPLES.md) - More embedding examples with explanations

**Key insight:** You only use `pkg/script/`. Everything in `pkg/cli/` is CLI-specific and not needed when embedding.

---

## 🎯 For Script Writers (Using Duso CLI)

You want to write Duso scripts and run them with the `duso` command.

**Start here:**
- [**CLI User Guide**](cli/README.md) - Overview and feature guide
- [**Getting Started**](cli/GETTING_STARTED.md) - Quick tutorial (write your first script)
- [**File I/O**](cli/FILE_IO.md) - load(), save(), include() functions
- [**Claude Integration**](cli/CLAUDE_INTEGRATION.md) - conversation(), claude() functions
- [**Examples**](cli/EXAMPLES.md) - Links to relevant examples in `/examples/cli/`

**Key insight:** When you run `duso script.du`, you get all core language features PLUS file I/O and Claude integration.

---

## 📚 For Everyone (Language Reference)

These documents apply to both embedded and CLI use:

- [**Learning Duso**](learning-duso.md) - Complete guided tour with examples
- [**Built-in Functions Reference**](/docs/reference/index.md) - Quick reference for all built-in functions
- [**Internals**](internals.md) - Architecture, design decisions, and runtime details

---

## Quick Navigation by Topic

### Learning the Language
1. Read [Learning Duso](learning-duso.md) - Guided tour with examples
2. Reference [Built-in Functions](/docs/reference/index.md) - Quick lookup for functions
3. Look at [examples/core/](../examples/core/) - Runnable examples of language features
4. Try the language in your chosen context (embedded or CLI)

### Embedding Duso
1. [Embedding Getting Started](embedding/GETTING_STARTED.md)
2. [API Reference](embedding/API_REFERENCE.md) for Go API details
3. [Custom Functions](embedding/CUSTOM_FUNCTIONS.md) for extending Duso
4. [examples/go-embedding/](../examples/go-embedding/) for complete Go examples

### Using the CLI
1. [CLI Getting Started](cli/GETTING_STARTED.md)
2. [File I/O](cli/FILE_IO.md) for load/save/include
3. [Claude Integration](cli/CLAUDE_INTEGRATION.md) for AI functions
4. [examples/cli/](../examples/cli/) for complete example scripts

### Contributing
- See [CONTRIBUTING.md](/CONTRIBUTING.md) for guidelines on where to make changes

---

## One Language, Two Paths

**Core Language** (both embedded and CLI):
- All basic features: variables, functions, objects, loops, error handling
- All built-in functions: string, math, array, date/time, JSON
- Lexical scoping, closures, templates, multiline strings

**CLI Extensions** (CLI only):
- File I/O: `load()`, `save()`, `include()`
- Claude integration: `claude()`, `conversation()`

When you write a Duso script in the `core/` examples, it works both embedded in Go and run via the CLI. Scripts in `cli/` examples **only** work with the CLI (they need file I/O or Claude).

---

## Can't Find What You're Looking For?

- **"How do I embed Duso in my Go app?"** → [Embedding Guide](embedding/README.md)
- **"How do I use load() and save()?"** → [CLI File I/O](cli/FILE_IO.md)
- **"How do I write a function?"** → [Learning Duso § Functions](learning-duso.md#functions)
- **"How do I add custom Go functions?"** → [Custom Functions](embedding/CUSTOM_FUNCTIONS.md)
- **"What's the syntax for objects?"** → [Learning Duso § Objects](learning-duso.md#objects)
- **"Can I contribute?"** → [CONTRIBUTING.md](/CONTRIBUTING.md)
