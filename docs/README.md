# Duso Documentation

Welcome! This documentation is organized by audience. Choose your path below:

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

- [**Language Specification**](language-spec.md) - Complete reference for syntax, types, operators, functions
- [**Implementation Notes**](implementation-notes.md) - Design decisions and architecture overview

---

## Quick Navigation by Topic

### Learning the Language
1. Read [language-spec.md](language-spec.md) - Start with the overview section
2. Look at [examples/core/](../examples/core/) - Runnable examples of language features
3. Try the language in your chosen context (embedded or CLI)

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
- See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on where to make changes

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
- **"How do I write a function?"** → [language-spec.md § Functions](language-spec.md#functions)
- **"How do I add custom Go functions?"** → [Custom Functions](embedding/CUSTOM_FUNCTIONS.md)
- **"What's the syntax for objects?"** → [language-spec.md § Objects](language-spec.md#objects)
- **"Can I contribute?"** → [CONTRIBUTING.md](../CONTRIBUTING.md)
