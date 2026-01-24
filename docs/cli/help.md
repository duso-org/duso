# Usage

```
duso [options] <script-path>
```

## Options

- `-v` - Enable verbose output
- `-doc NAME` - Display documentation for a module or builtin

## Examples

Run a script:

```
duso examples/basic.du
```

Run with verbose output:

```
duso -v ../scripts/my-script.du
```

View module documentation:

```
duso -doc http
duso -doc map
duso -doc claude
```

## More Information

For complete documentation, see the [Language Specification](../language-spec.md) and [CLI Guide](../cli/README.md).
