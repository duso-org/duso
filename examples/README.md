# Duso Examples

This directory contains example scripts demonstrating Duso's features organized by use case.

## Core Language Examples (`core/`)

These examples demonstrate the **Duso language itself** and work in any Duso environment—whether embedded in a Go application or run via the CLI.

**Language Fundamentals:**
- `basic.du` - Variables, assignments, and simple operations
- `variables.du` - Variable scoping and closures
- `functions.du` - Function definition and calling

**Data Structures:**
- `arrays.du` - Array operations and iteration
- `structures.du` - Objects as constructor blueprints
- `methods.du` - Object methods with implicit property access

**Operators & Control Flow:**
- `break-continue.du` - Loop control statements
- `ternary-test.du` - Conditional expressions with the ternary operator
- `coercion.du` - Type coercion and truthiness rules

**String Handling:**
- `strings.du` - String operations
- `multiline.du` - Multiline string syntax (triple quotes)
- `templates.du` - String templates with `{{expr}}` syntax
- `template-test.du` - Template evaluation examples

**Built-in Functions:**
- `builtins.du` - Comprehensive showcase of all built-in functions
- `find_replace.du` - String search and replace
- `sort_custom.du` - Custom comparison functions with sort()
- `dates.du` - Date and time functions (now, format_time, parse_time)
- `test_json.du` - JSON parsing and formatting

**Advanced Features:**
- `test_var.du` - Variable scope with `var` keyword and closures
- `benchmark.du` - Performance testing example
- `colors.du` - ANSI color constants (helper for terminal output)

## CLI Examples (`cli/`)

These examples use **CLI-specific features** available only when running scripts with the `duso` command:

**File I/O:**
- `file-io.du` - Reading and writing files with `load()` and `save()`
- `multi-file.du` - Loading and executing other scripts with `include()`
- `with-include.du` - Using `include()` for code organization

**Claude Integration:**
- `fun.du` - Simple interactive conversation with Claude
- `comedy_writer.du` - Single-shot Claude call with analysis
- `spy_workflow.du` - Multi-step agent workflow
- `spy_vs_guard.du` - Comparison workflow
- `panel/` - Expert panel analysis using multiple Claude calls
- `self-taught.du` - Self-improving agent architecture

## Running Examples

**From the command line:**
```bash
# Run a core example (works with duso CLI)
duso core/basic.du

# Run a CLI example (requires duso CLI)
duso cli/file-io.du
```

**In embedded Go code:**
```go
// Works with any example from core/
content, _ := ioutil.ReadFile("examples/core/basic.du")
result, _ := interp.Execute(string(content))

// Will NOT work with cli/ examples (no load/save/include functions)
```

## Quick Navigation

- **Getting Started?** Start with `core/basic.du`
- **Learning the language?** Work through `core/` examples in order
- **Exploring CLI features?** Check out `cli/file-io.du` and `cli/` examples
- **Embedding Duso?** Look at `core/` examples—these all work in embedded contexts
