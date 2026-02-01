# Duso Examples

This directory contains example scripts demonstrating Duso's features organized by use case.

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

