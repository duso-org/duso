# Custom Duso Distributions

One of Duso's superpowers is that **you don't need to be a Go developer** to create your own custom binary with your own modules baked in.

The Duso binary is a self-contained, frozen unit. Once built, it works forever - no external dependencies, no version conflicts. By creating a custom distribution, you can add your own Duso modules to stdlib and build your own binary.

## Why Custom Distributions?

- **Add your own modules** - Custom helpers, tools, workflows specific to your needs
- **Include third-party modules** - Add community modules to your binary
- **Freeze versions** - Everything is baked in at build time
- **Share with your team** - One binary, everyone has the same tools
- **Production-ready** - Your custom duso binary is a cohesive, dependencies-free unit

## Getting Started (No Go Knowledge Required)

### 1. Clone Duso

```bash
git clone https://github.com/duso-org/duso.git
cd duso
```

### 2. Add Your Module to `stdlib/`

Create a new module directory:

```bash
mkdir -p stdlib/mymodule
```

Create your module file (just Duso code):

```bash
cat > stdlib/mymodule/mymodule.du << 'EOF'
// My custom module - pure Duso code
function greet(name)
  return "Hello, " + name + "!"
end

function multiply(a, b)
  return a * b
end

// Export functions as an object
{
  greet = greet,
  multiply = multiply,
  version = "1.0.0"
}
EOF
```

### 3. Build Your Custom Duso

```bash
# From the repo root, run the build process:
go generate ./cmd/duso
go build -o my-duso ./cmd/duso
```

That's it. Your custom duso binary is ready:

```bash
./my-duso << 'EOF'
mod = require("mymodule")
print(mod.greet("World"))
print(mod.multiply(6, 7))
EOF
```

Output:
```
Hello, World!
42
```

## Example: Custom Distribution with Organization Tools

Create a module with helpers your team uses:

```bash
mkdir -p stdlib/teamtools

cat > stdlib/teamtools/teamtools.du << 'EOF'
// Team tools module

function format_timestamp(dt)
  // Format a date nicely
  return dt
end

function log_event(level, message)
  // Format a log entry
  return "[" + level + "] " + message
end

function is_valid_email(email)
  // Basic email validation
  return contains(email, "@")
end

{
  format_timestamp = format_timestamp,
  log_event = log_event,
  is_valid_email = is_valid_email,
  version = "1.0.0"
}
EOF
```

Build:

```bash
go generate ./cmd/duso
go build -o duso-team ./cmd/duso
```

Now your team can use it:

```bash
./duso-team << 'EOF'
tools = require("teamtools")
print(tools.log_event("INFO", "Process started"))
print(tools.is_valid_email("user@example.com"))
EOF
```

## Directory Structure

After adding modules, your stdlib looks like:

```
stdlib/
├── http/              # Built-in HTTP module
│   ├── http.du
│   └── http.md
├── mymodule/          # Your custom module
│   └── mymodule.du
└── teamtools/         # Another custom module
    └── teamtools.du
```

## Building and Distributing

### Quick Local Build

```bash
# During development
go generate ./cmd/duso
go build -o duso-dev ./cmd/duso
./duso-dev test.du
```

### Release Build

```bash
# Clean build for distribution
go generate ./cmd/duso
go build -o duso-myorg ./cmd/duso

# Binary is complete and ready to share
ls -lh duso-myorg
```

### Sharing with Your Team

```bash
# Build your custom distribution
go generate ./cmd/duso
go build -o duso-custom ./cmd/duso

# Upload to GitHub releases, your artifact server, S3, etc.
# Everyone on the team downloads the same binary
# Everyone has the same modules, same versions, same behavior
```

## Updating Modules

To update a module in your custom distribution:

1. Edit the `.du` file in `stdlib/yourmodule/`
2. Rebuild: `go generate ./cmd/duso && go build -o duso-custom ./cmd/duso`
3. Done - new binary includes your updates

## Naming Your Distribution

Give your custom distribution a recognizable name so it's clear it's different from the standard duso:

```bash
# Organization-specific
go build -o duso-acme ./cmd/duso

# Purpose-specific
go build -o duso-datapipeline ./cmd/duso

# Team-specific
go build -o duso-dataeng ./cmd/duso
```

This makes it immediately clear to users that they're using a custom distribution with organization-specific tools, not the standard Duso binary.

## Important Notes

- **Pure Duso only** - You only write Duso code in stdlib modules. No Go required.
- **Build time binding** - Modules are baked into the binary at build time
- **Binary size** - Custom modules typically add kilobytes; http module adds ~2-3MB
- **Zero runtime dependencies** - Once built, your duso binary needs nothing else
- **Frozen in time** - duso-v1.0 will work the same way forever

## Optional: Add Documentation

Document your custom modules with a README:

```bash
cat > stdlib/mymodule/README.md << 'EOF'
# My Module

Custom tools for my workflow.

## Usage

```duso
mod = require("mymodule")
result = mod.greet("Alice")
```

## Functions

- `greet(name)` - Greet someone
- `multiply(a, b)` - Multiply two numbers
EOF
```

## Next Steps

- Explore [stdlib/](../stdlib/) to see how built-in modules are structured
- Read the [Language Spec](language-spec.md) for Duso syntax
- Check the [CLI Guide](cli/README.md) for script features
- See [Contributing](/CONTRIBUTING.md) if you want to contribute modules back to core Duso

## What's Possible

With custom distributions, you can create:

- **Organization distributions** - Shared tools for your team
- **Domain-specific distributions** - Duso configured for a specific task (data processing, automation, etc.)
- **Educational distributions** - Duso bundled with teaching modules
- **Integrated distributions** - Duso with industry-standard modules pre-loaded
- **Frozen snapshots** - A version of Duso from a specific date, guaranteed to work unchanged

The only limit is your imagination.
