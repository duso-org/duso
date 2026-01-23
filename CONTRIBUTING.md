# Contributing to Duso

Thank you for your interest in contributing to Duso! This document explains how the project is organized and where to make different types of changes.

## Project Organization

Duso is organized around two core audiences: **Go developers embedding Duso** and **script writers using the CLI**.

### Package Structure

```
cmd/duso/              - CLI application (entry point)
pkg/
  ├── script/          - Core language (lexer, parser, evaluator, builtins)
  ├── anthropic/       - Claude API client
  └── cli/             - CLI-specific functions (file I/O, conversation API)
examples/
  ├── core/            - Language feature examples (work everywhere)
  └── cli/             - CLI-specific examples (file I/O, Claude)
docs/
  ├── embedding/       - Documentation for Go developers
  └── cli/             - Documentation for script writers
vscode/                - VSCode syntax highlighting extension
```

## Where to Make Changes

### Adding a Language Feature

**You want to add:** Operators, syntax, control flow, new built-in functions

**Files to modify:**
1. `pkg/script/token.go` - Add token type if needed
2. `pkg/script/lexer.go` - Add tokenization if needed
3. `pkg/script/parser.go` - Add parsing logic
4. `pkg/script/evaluator.go` - Add evaluation logic
5. `pkg/script/builtins.go` - If it's a built-in function
6. `docs/language-spec.md` - Document the feature
7. `examples/core/` - Add example demonstrating the feature

**Process:**
1. Make language changes in `pkg/script/`
2. Add example to `examples/core/`
3. Update `docs/language-spec.md`
4. Test in both embedded and CLI contexts

### Adding a CLI-Specific Feature

**You want to add:** File I/O enhancements, new Claude API patterns, CLI commands

**Files to modify:**
1. `pkg/cli/` - Implement the feature
2. `cmd/duso/main.go` - Register if needed
3. `docs/cli/` - Document the feature
4. `examples/cli/` - Add example

**Process:**
1. Implement in `pkg/cli/` or `cmd/duso/`
2. Register in CLI setup
3. Add example to `examples/cli/`
4. Update relevant docs in `docs/cli/`

### Adding Custom Go Functions

**You want embedders to:** Use a pre-built function in their apps

**Files to modify:**
1. `pkg/script/builtins.go` - Add the function
2. Provide registration code (or build it into core)
3. `docs/embedding/CUSTOM_FUNCTIONS.md` - Document patterns
4. `docs/language-spec.md` - If it's a core built-in

**Process:**
1. Implement in `pkg/script/builtins.go`
2. Add to appropriate test file
3. Document in language spec
4. Provide example

### Improving Documentation

**Files to modify:**
- `docs/language-spec.md` - Language syntax and semantics
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

### CLI Feature Requests

**Label:** `feature: cli`

These go in `pkg/cli/` and only affect CLI usage.

Example: "Add ability to set environment variables from script"

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

# Build the CLI
go build -o duso cmd/duso/main.go

# Run a test script
./duso examples/core/basic.du
```

### Testing

```bash
# Run tests
go test ./...

# Run a specific example
./duso examples/core/functions.du

# Test with verbose output
./duso -v examples/core/basic.du
```

### Building Examples

```bash
# Test an embedded example
cd examples/go-embedding
go run 01-hello-world.go
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

---

Thank you for contributing to Duso! 🎉
