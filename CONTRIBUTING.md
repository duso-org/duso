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
  ├── script/          - Core language (lexer, parser, evaluator, builtins)
  ├── runtime/         - Runtime orchestration (HTTP, datastore, concurrency)
  ├── cli/             - CLI-specific functions (file I/O, Claude, module resolution)
  ├── anthropic/       - Claude API client (internal)
  └── markdown/        - Markdown rendering (internal)
examples/
  ├── core/            - Language feature examples (work everywhere)
  └── cli/             - CLI-specific examples (file I/O, Claude)
docs/
  ├── embedding/       - Documentation for Go developers
  └── cli/             - Documentation for script writers
vscode/                - VSCode syntax highlighting extension
```

## Package Responsibilities

**Three-Layer Architecture:**

1. **`pkg/script/`** - Language Core (Embeddable)
   - Lexer, parser, evaluator
   - Type system and value representation
   - Core built-in functions
   - Environment and scope management
   - No I/O, no HTTP, no external dependencies

2. **`pkg/runtime/`** - Orchestration Layer (Embeddable)
   - HTTP server (`HTTPServerValue`)
   - HTTP client (`HTTPClientValue`)
   - Datastore coordination (`DatastoreValue`)
   - Goroutine context management
   - Request/concurrency primitives
   - Can be used in embedded apps or CLI

3. **`pkg/cli/`** - CLI Features (CLI-only)
   - File I/O (load, save, include, require)
   - Claude integration (claude, conversation)
   - Module resolution and circular dependency detection
   - Script function wrappers for runtime features
   - Documentation lookup
   - Environment variable access

## Where to Make Changes

### Adding a Language Feature

**You want to add:** Operators, syntax, control flow, new built-in functions

**Files to modify:**
1. `pkg/script/token.go` - Add token type if needed
2. `pkg/script/lexer.go` - Add tokenization if needed
3. `pkg/script/parser.go` - Add parsing logic
4. `pkg/script/evaluator.go` - Add evaluation logic
5. `pkg/script/builtins.go` - If it's a built-in function
6. `docs/learning-duso.md` - Document the feature
7. `examples/core/` - Add example demonstrating the feature

**Process:**
1. Make language changes in `pkg/script/`
2. Add example to `examples/core/`
3. Update `docs/learning-duso.md`
4. Test in both embedded and CLI contexts

### Adding a Runtime Feature (Embeddable)

**You want to add:** HTTP server enhancements, datastore features, concurrency primitives, goroutine management

**Files to modify:**
1. `pkg/runtime/` - Implement the core feature
2. `pkg/cli/` - Create wrapper function(s) if needed for script use
3. `docs/internals.md` - Document the runtime architecture
4. `examples/core/` - Add example (works in embedded and CLI contexts)

**Process:**
1. Implement in `pkg/runtime/`
2. If needed for scripts, create wrapper in `pkg/cli/` to expose it
3. Register wrapper in `pkg/cli/register.go` (for CLI usage)
4. Update documentation
5. Add examples showing embedded and CLI usage
6. Test in both contexts

**Examples of runtime features:**
- `pkg/runtime/http_server.go` + `pkg/cli/http_server.go` wrapper
- `pkg/runtime/datastore.go` + `pkg/cli/datastore.go` wrapper
- `pkg/runtime/goroutine_context.go` + `pkg/cli/run.go` and `pkg/cli/spawn.go` wrappers

### Adding a CLI-Specific Feature

**You want to add:** File I/O enhancements, new Claude API patterns, module resolution improvements

**Files to modify:**
1. `pkg/cli/` - Implement the feature
2. `cmd/duso/main.go` - Register if needed
3. `docs/cli/` - Document the feature
4. `examples/cli/` - Add example (CLI-only)

**Process:**
1. Implement in `pkg/cli/`
2. Register in `pkg/cli/register.go`
3. Add example to `examples/cli/`
4. Update relevant docs in `docs/cli/`
5. Document that this feature is CLI-only (not embeddable)

**Examples of CLI-only features:**
- File I/O (`load`, `save`, `include`)
- Module resolution (`require`)
- Claude integration (`claude`, `conversation`)
- Environment variable access (`env`)

### Adding Custom Go Functions

**You want embedders to:** Use a pre-built function in their apps

**Files to modify:**
1. `pkg/script/builtins.go` - Add the function
2. Provide registration code (or build it into core)
3. `docs/embedding/CUSTOM_FUNCTIONS.md` - Document patterns
4. `docs/learning-duso.md` - If it's a core built-in

**Process:**
1. Implement in `pkg/script/builtins.go`
2. Add to appropriate test file
3. Document in learning guide
4. Provide example

### Contributing a Module to the Registry

**You want to:** Share a Duso module that gets included in Duso binary distributions

**What you provide:** Pure Duso code - no Go required

**Files to modify:**
1. Create a repository: `duso-<modulename>` on GitHub
2. Add your module as `.du` files with MIT license
3. Include documentation and examples
4. Submit for review

**Process:**
1. Create your module in a separate repository
   - Follow the naming convention: `duso-postgres`, `duso-helpers`, etc.
   - License under MIT (copy from Duso's LICENSE file)
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

**For more details:** See [contrib/README.md](contrib/README.md)

**Context:** See [custom distributions](docs/custom_distributions.md) to understand how modules are included in binary distributions.

### Improving Documentation

**Files to modify:**
- `docs/learning-duso.md` - Language syntax and semantics
- `docs/embedding/` - Guides for Go developers
- `docs/cli/` - Guides for script writers
- `README.md` - Project overview
- In-code comments - Implementation details

**Process:**
1. Identify unclear or missing documentation
2. Improve clarity, add examples
3. Test that examples actually work
4. Update related documentation files

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
- Update relevant documentation files
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

Example: "Add HTTP session management to http_client", "Add timeout to datastore operations"

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
duso examples/core/basic.du
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
duso examples/core/functions.du

# Test with verbose output
duso -v examples/core/basic.du
```

### Building Examples

```bash
# Test a Go embedding example
cd examples/go-embedding
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

- Check [docs/README.md](docs/README.md) for documentation navigation
- Look at existing code for patterns
- Review related issues and pull requests
- Ask in a discussion or new issue

## Code of Conduct

Be respectful, inclusive, and constructive. Welcome contributors and help them succeed.

Thank you for contributing to Duso!
