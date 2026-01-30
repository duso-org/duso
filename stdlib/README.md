# Duso Standard Library

The Duso stdlib provides essential modules for common tasks. These modules are distributed with duso and can be loaded using `require()`.

## Available Modules

### http
HTTP client for making requests and managing connections.

```duso
http = require("http")
response = http.fetch("https://api.example.com/data")
data = parse_json(response)
```

- **Status**: Stable (v1.0+)
- **Documentation**: [http.md](http/http.md)
- **Examples**: [http/examples/](http/examples/)

## Module Organization

Each module lives in its own directory with:
- `modulename.du` - The main module file
- `modulename.md` - Full documentation
- `examples/` - Working example scripts
- Sub-modules: `submodule.du` (loaded via `require("modulename/submodule")`)

## Stability Levels

**v0.x (Experimental)**
- API may change
- Breaking changes allowed
- Use at your own risk

**v1.0+ (Stable)**
- API is stable
- Follows semantic versioning
- Breaking changes only in major versions
- Safe for production use

## Contributing to Stdlib

Want to add a module or improve existing ones? See [CONTRIBUTING.md](/CONTRIBUTING.md#contributing-to-stdlib) for guidelines.

### Quick Start for Contributors

1. Open an RFC issue proposing the module
2. Get design feedback from maintainers
3. Implement module in `stdlib/modulename/`
4. Include examples and comprehensive docs
5. Submit PR

### Module Template

```
stdlib/mymodule/
├── mymodule.du          # Main module
├── mymodule.md          # Documentation
├── subfeature.du        # Optional sub-modules
├── examples/
│   ├── basic.du
│   ├── advanced.du
│   └── README.md
└── [tests]
```

## See Also

- [Language Reference](/docs/language-spec.md)
- [File I/O Guide](/docs/cli/FILE_IO.md)
- [Contributing Guide](/CONTRIBUTING.md)
