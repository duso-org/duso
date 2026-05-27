# Embedding Examples

Complete, runnable Go applications demonstrating Duso embedding.

All examples are in `/go-embedding/`.

## 01 - Hello World

**File:** `/go-embedding/01-hello-world.go`

The simplest possible Duso embedding:

- Create an interpreter
- Execute a basic script
- Display output

**Demonstrates:**
- Importing the Duso package
- Creating an interpreter
- Executing Duso code
- Basic print output

**Run it:**
```bash
go run /go-embedding/01-hello-world.go
```

---

## 02 - Custom Functions

**File:** `/go-embedding/02-custom-functions.go`

Extend Duso with your own Go functions:

- Register custom functions
- Call them from Duso scripts
- Handle arguments and return values
- Error handling

**Demonstrates:**
- RegisterFunction()
- Type conversion (float64, string, etc.)
- Named arguments
- Error handling in custom functions

**Run it:**
```bash
go run /go-embedding/02-custom-functions.go
```

---

## 03 - Configuration DSL

**File:** `/go-embedding/03-config-dsl.go`

Use Duso as a configuration language:

- Load config from a .du file
- Parse and validate configuration
- Access nested values
- Use in Go code

**Demonstrates:**
- Loading scripts from files
- GetVariable() to access script variables
- Type conversion (map[string]any)
- Configuration patterns

**Real-world use case:** Configuration files, app settings

**Run it:**
```bash
go run /go-embedding/03-config-dsl.go
```

---

## 04 - Task Orchestration

**File:** `/go-embedding/04-task-scripting.go`

Coordinate multi-step workflows:

- Define tasks in Duso
- Call Go functions from task scripts
- Orchestrate complex operations
- Aggregate results

**Demonstrates:**
- Multiple custom functions
- Duso control flow (loops, if/else)
- Calling Duso functions from Go (Call())
- Complex data structures (arrays of objects)

**Real-world use case:** ETL pipelines, job orchestration, agent workflows

**Run it:**
```bash
go run /go-embedding/04-task-scripting.go
```

---

## Which Example Should I Read?

**"I want to learn embedding basics"**
→ Start with **01-hello-world**, then **02-custom-functions**

**"I want to use Duso for configuration"**
→ Read **03-config-dsl**

**"I want to coordinate workflows"**
→ Read **04-task-scripting**

**"I want complete working code"**
→ All examples are self-contained and runnable

---

## Example Templates

### Template: Simple Script Execution

```go
package main

import (
    "fmt"
    "github.com/duso-org/duso/pkg/script"
)

func main() {
    interp := script.NewInterpreter(false)

    result, err := interp.Execute(`
        x = 5
        y = 10
        print(x + y)
    `)

    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

### Template: Custom Function

```go
package main

import (
    "fmt"
    "github.com/duso-org/duso/pkg/script"
)

func main() {
    interp := script.NewInterpreter(false)

    interp.RegisterFunction("double", func(args map[string]any) (any, error) {
        x := args["x"].(float64)
        return x * 2, nil
    })

    interp.Execute(`result = double(x = 21)`)
    fmt.Println(interp.GetVariable("result"))  // 42
}
```

### Template: Configuration

```go
package main

import (
    "os"
    "github.com/duso-org/duso/pkg/script"
)

func main() {
    interp := script.NewInterpreter(false)

    config, _ := os.ReadFile("config.du")
    interp.Execute(string(config))

    settings := interp.GetVariable("settings")
    // Use settings...
}
```

---

## Building All Examples

```bash
cd /go-embedding
go build -o ../bin/01-hello-world 01-hello-world.go
go build -o ../bin/02-custom-functions 02-custom-functions.go
go build -o ../bin/03-config-dsl 03-config-dsl.go
go build -o ../bin/04-task-scripting 04-task-scripting.go

# Run them
../bin/01-hello-world
../bin/02-custom-functions
../bin/03-config-dsl
../bin/04-task-scripting
```

---

## Next Steps

1. **Run an example** that matches your use case
2. **Modify it** to do something similar to what you need
3. **Read the guides:**
   - [API Reference](/docs/embedding/api-reference.md) for available methods
   - [Custom Functions](/docs/embedding/custom-functions.md) for more patterns
   - [Patterns](/docs/embedding/patterns.md) for design guidance

## See Also

- [Getting Started](/docs/embedding/getting-started.md) - Tutorial format
- [API Reference](/docs/embedding/api-reference.md) - Function documentation
- [Custom Functions](/docs/embedding/custom-functions.md) - Function registration guide
- [Patterns](/docs/embedding/patterns.md) - Design patterns and use cases
