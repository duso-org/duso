# Contributing to Duso

Thank you for your interest in contributing to Duso! This document explains how the project is organized and where to make different types of changes.

## Before you send your code

We welcome your contributions! Before you open a PR, please:

- pick something **small** and see if there is a need/interest first
- don't pass us slop! test and re-test and review your code
- we may not want your contribution, and that should be ok
- don't be discouraged! sometimes it takes time to match the mindset and pace of a new open source project

## Project Organization

Duso is organized around two core audiences: **Go developers embedding Duso** and **script writers using the CLI**.

### Package Structure

```
cmd/duso/              - CLI application (entry point)
pkg/
  ├── script/          - Core language (lexer, parser, resolver, evaluator)
  ├── runtime/         - Every builtin, plus HTTP, datastore, websockets, SQL
  ├── cli/             - CLI-only builtins (file I/O, require, sys) and module resolution
  ├── core/            - Small shared helpers
  ├── lsp/             - Language server (completion, diagnostics, hover)
  └── version/         - Version constants
stdlib/                - Duso-source standard library (.du modules, embedded)
contrib/               - Duso-source community modules (.du, embedded)
examples/              - Duso example scripts (embedded)
go-embedding/          - Go embedding examples (not embedded in binary)
docs/
  ├── reference/       - Per-symbol reference docs
  ├── ideas/           - Proposed, not-yet-built designs
  └── cli/             - CLI help text
```

Everything under `docs/`, `stdlib/`, `contrib/` and `examples/` is copied into the binary
by `go generate ./cmd/duso` and served from `/EMBED/`. That embed is deliberate: a duso
binary carries the docs and libraries for exactly its own revision, so a user always reads
documentation that matches the code they are running. Keep it curated — anything stale in
`docs/` at release time is frozen into that release.

## Package Responsibilities

**Three-Layer Architecture:**

1. **`pkg/script/`** - Language Core (Embeddable)
   - Lexer, parser, resolver, evaluator
   - Type system and value representation
   - Environment and scope management
   - Circular dependency detection
   - No I/O, no HTTP, no builtins, no third-party dependencies

2. **`pkg/runtime/`** - Builtins and Orchestration (Embeddable)
   - Every builtin: 46 `pkg/runtime/builtin_*.go` files, registered in `pkg/runtime/register.go`
   - HTTP server (`HTTPServerValue`) and client (used by `fetch()`)
   - Datastore coordination (`DatastoreValue`) and replication
   - WebSockets, SQL, mail, images
   - Goroutine context management, request/concurrency primitives
   - By far the largest package

3. **`pkg/cli/`** - CLI Features (CLI-only)
   - CLI-only builtins in `builtin_*.go`: `load`, `require`, `sys`, `watch`, `console`
   - Module resolution (`pkg/cli/module_resolver.go`)
   - Virtual filesystem and embedded-FS access (`pkg/cli/vfs.go`)
   - Registration entry point (`pkg/cli/register.go`)

Note that builtins live in `pkg/runtime`, not `pkg/script`. `pkg/script` is the bare
language; embedding it alone gives you a sandbox with no `print()`.

## Where to Make Changes

### Adding a Language Feature

**You want to add:** Operators, syntax, control flow, new built-in functions

**Files to modify:**
1. `pkg/script/token.go` - Add token type if needed
2. `pkg/script/lexer.go` - Add tokenization if needed
3. `pkg/script/parser.go` - Add parsing logic
4. `pkg/script/evaluator.go` - Add evaluation logic
5. `pkg/runtime/builtin_*.go` - If it's a built-in function
6. `docs/learning-duso.md` - Document the feature
7. `docs/duso-primer.md` - Add a terse line if it changes how correct code is written
8. `examples/` - Add an example demonstrating the feature

**Process:**
1. Make language changes in `pkg/script/`
2. Add an example under `examples/`
3. Update `docs/learning-duso.md`
4. Run `duso lint` over any doc you touched — code blocks in `.md` are linted
5. Test in both embedded and CLI contexts

### Adding a Runtime Feature (Embeddable)

**You want to add:** HTTP server enhancements, datastore features, concurrency primitives, goroutine management

**Files to modify:**
1. `pkg/runtime/builtin_<name>.go` - Implement the builtin
2. `pkg/runtime/register.go` - Register it
3. `docs/reference/<name>.md` - Per-symbol documentation
4. `docs/internals.md` - Only if it changes the architecture
5. `examples/` - Add an example

**Process:**
1. Implement in `pkg/runtime/builtin_<name>.go`
2. Register in `pkg/runtime/register.go` — runtime builtins register themselves there;
   they do not need a `pkg/cli` wrapper
3. Add Go tests alongside the implementation
4. Write `docs/reference/<name>.md`
5. Test in both embedded and CLI contexts

**Examples of runtime features:**
- `pkg/runtime/builtin_http_server.go` + `pkg/runtime/http_server.go`
- `pkg/runtime/builtin_datastore.go` + `pkg/runtime/datastore.go`
- `pkg/runtime/goroutine_context.go` for context carried across `spawn()`/`run()`

### Adding a CLI-Specific Feature

**You want to add:** File I/O enhancements, module resolution improvements, new CLI
subcommands

**Files to modify:**
1. `pkg/cli/builtin_<name>.go` - Implement the feature
2. `pkg/cli/register.go` - Register it
3. `cmd/duso/main.go` - Only for a new subcommand
4. `docs/reference/<name>.md` - Document the builtin

**Process:**
1. Implement in `pkg/cli/`
2. Register in `pkg/cli/register.go`
3. Add an example under `examples/`
4. Document that the feature is CLI-only (not available when embedding `pkg/runtime`)

**Examples of CLI-only features:**
- File I/O (`load`, `save`, `include`)
- Module resolution (`require`)
- Environment variable access (`env`)

### Adding Custom Go Functions

**You want embedders to:** Use a pre-built function in their apps

**Files to modify:**
1. `pkg/runtime/builtin_*.go` - Add the function
2. `pkg/runtime/register.go` - Register it
3. `docs/reference/<name>.md` - Document it
4. `docs/learning-duso.md` - If it's a core built-in

**Process:**
1. Implement in `pkg/runtime/builtin_*.go`
2. Add to appropriate test file
3. Document in learning guide
4. Provide example

### Contributing a Module to the Registry

**You want to:** Share a Duso module that gets included in Duso binary distributions

**What you provide:** Pure Duso code - no Go required

**Files to modify:**
1. Create a repository: `duso-<modulename>` on GitHub
2. Add your module as `.du` files with Apache 2.0 license
3. Include documentation and examples
4. Submit for review

**Process:**
1. Create your module in a separate repository
   - Follow the naming convention: `duso-postgres`, `duso-helpers`, etc.
   - License under Apache 2.0 (copy from Duso's LICENSE file)
   - Include clear documentation and examples

2. Open an issue on the Duso repository requesting review
   - Include: repo URL, module description, use case
   - Duso team reviews for quality and standards

3. Once approved, module is added to `contrib/`
   - Module becomes available in all future Duso distributions
   - Code is baked into the binary at build time

4. Your module is frozen in time with each release
   - Duso can preserve working versions indefinitely
   - Users get a reliable, dependency-free way to use your module

**For more details:** See [contrib/README.md](/contrib/README.md)

**Context:** See [custom distributions](/docs/bundling-applications.md) to understand how modules are included in binary distributions.

### Improving Documentation

Each doc has one job. Put a change where it belongs rather than where it is easiest to add.

**Reference — the authority on behavior**
- `docs/reference/<name>.md` - One file per builtin, keyword and symbol. Every builtin has
  one, and it is what `duso doc TERM` prints. Parameters, return values, errors, examples.
- `docs/formal-language-spec.md` - Grammar and semantics, precisely stated

**Learning — prose for people**
- `docs/learning-duso.md` - Language syntax and semantics, explained at length
- `docs/duso-style-guide.md` - How idiomatic duso is written
- `docs/debugging-scripts.md`, `docs/files-and-modules.md`,
  `docs/virtual-filesystem.md`, `docs/bundling-applications.md` - Topic guides

**Primers — condensed, LLM-optimized**
- `docs/duso-primer.md` - The language, dense. Notes here are one line, never a
  subsection with examples; depth belongs in a topic guide or the reference
- `docs/datastore-primer.md` - `datastore()` in full
- `docs/http-primer.md` - `http_server()`, `fetch()`, `websocket()`

**Contributor-facing**
- `docs/internals.md` - Architecture, the evaluation pipeline, the Go embedding API
- `docs/ideas/` - Proposed designs that are **not built**. Say so at the top of the file.
  Delete the doc when the feature ships or the design is abandoned
- `docs/performance-report.md`, `docs/datastore-1.7-performance.md`,
  `docs/todo-comparison.md` - Measurements and comparisons

**Other**
- `docs/index.md` - Table of contents. A new doc that is not listed here is invisible
- `docs/cli/help.md` - Text printed by `duso help`
- `docs/installing.md`, `README.md` - Front door
- In-code comments - Implementation details

**Process:**
1. Identify unclear, missing, or wrong documentation
2. **Run it before you write it.** Do not document behavior you have not executed —
   check signatures and defaults against the source, not against another doc
3. Improve clarity, add examples
4. `duso lint <file>.md` — code blocks inside Markdown are linted, so broken examples
   are caught before they ship
5. Add new docs to `docs/index.md`
6. Update related docs so they don't contradict each other

**Remember the embed.** `docs/` is copied into the binary, so a doc that is wrong at
release time is frozen into that release and cannot be patched afterward. Prefer deleting
a stale doc over leaving it to mislead.

### Fixing a Bug

**Process:**
1. Identify which package contains the bug
2. Add test case demonstrating the bug
3. Fix the bug
4. Verify test passes
5. Check if documentation needs updating
6. Update examples if behavior changed

## Code Standards

### Style

- Follow Go conventions
- Use `camelCase` for functions/variables
- Use `snake_case` for Duso language functions
- Comment exported functions and complex logic

### Testing

- Add test cases for new features
- Test in both embedded and CLI contexts
- If you add a language feature, ensure it works when embedded
- If you add a CLI feature, ensure it doesn't break embedding

### Documentation

- Document all exported Go functions
- Add examples in `/examples/` for user-facing features
- Every builtin needs a `docs/reference/<name>.md`
- Never document behavior that isn't implemented — run it first
- Comment complex algorithms

## Pull Request Process

1. **Fork** the repository
2. **Create a branch** with a descriptive name: `feature/add-xyz`, `fix/issue-123`
3. **Make your changes** following the guidelines above
4. **Write/update tests** for your changes
5. **Update documentation** - Both code comments and user docs
6. **Add examples** if applicable
7. **Test thoroughly** - Both embedded and CLI usage
8. **Submit PR** with description of changes

## Issue Categories

### Language Feature Requests

**Label:** `feature: language`

These go in `pkg/script/` and may affect both embedding and CLI.

Example: "Add switch/case syntax"

### Runtime Feature Requests

**Label:** `feature: runtime`

These go in `pkg/runtime/` and can be used in embedded and CLI contexts.

Example: "Add request caching to fetch()", "Add timeout to datastore operations"

### CLI Feature Requests

**Label:** `feature: cli`

These go in `pkg/cli/` and only affect CLI usage.

Example: "Add ability to set environment variables from script", "Improve module resolution for monorepos"

### Documentation Issues

**Label:** `docs`

Unclear, missing, or outdated documentation.

Example: "The file I/O guide doesn't explain path resolution"

### Bug Reports

**Label:** `bug`

Something doesn't work as documented.

Provide:
- What you tried
- What you expected
- What happened instead
- Minimal reproduction case

## Development Workflow

### Setting Up

```bash
git clone https://github.com/duso-org/duso
cd duso

# For maintainers only: install git hooks for automatic versioning
./git-setup.sh

# Build the CLI
./build.sh

# Run a test script
duso examples/http/minimal.du
```

**For maintainers:** After cloning, run `./git-setup.sh` to install git hooks for automatic versioning based on commit messages (`feat:`, `fix:`, `major:` prefixes).

**Symlink for convenience** (so you can run `duso` from anywhere):
```bash
ln -s $(pwd)/bin/duso /usr/local/bin/duso
```

### Testing

```bash
# Run tests
go test ./...

# Run a specific example
duso examples/http/minimal.du

# Static analysis (works on .du files and on code blocks in .md files)
duso lint examples/http/minimal.du

# Step through a script in the debugger
duso debug examples/debug/basic.du
```

### Building Examples

```bash
# Test a Go embedding example
cd go-embedding
go run ./hello-world
```

## Important Principles

1. **Backward Compatibility** - Don't break existing scripts
2. **Core is Minimal** - Keep `pkg/script/` focused on language
3. **CLI is Optional** - Core features should work without CLI
4. **Clear Separation** - Embedded vs CLI concerns should be obvious
5. **Documentation** - Every user-facing feature needs good docs
6. **Examples** - Show, don't just tell

## Questions?

- Check [Documentation](/docs/index.md) for documentation navigation
- Look at existing code for patterns
- Review related issues and pull requests
- Ask in a discussion or new issue

## Code of Conduct

Be respectful, inclusive, and constructive. Welcome contributors and help them succeed.

Thank you for contributing to Duso!
