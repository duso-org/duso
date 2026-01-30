# Duso Language Extension for VS Code

Syntax highlighting for [Duso](https://github.com/duso-org/duso) - a lightweight scripting language designed for AI agents and automation.

## Quick Start

### Option 1: Copy to Extensions (Simple)
```bash
cp -r /path/to/duso/vscode ~/.vscode/extensions/duso
```
Then restart VS Code.

### Option 2: Symlink for Development (Hot Reload)
```bash
ln -s /path/to/duso/vscode ~/.vscode/extensions/duso
```
Then reload VS Code window with `Cmd+Shift+P` → `Developer: Reload Window`

## Features

**Syntax Highlighting for:**
- **Keywords**: `if`, `then`, `elseif`, `else`, `end`, `while`, `do`, `for`, `in`, `function`, `return`, `break`, `continue`, `try`, `catch`, `var`, `raw`
- **Builtins**: 45+ functions including `print`, `len`, `append`, `map`, `filter`, `split`, `replace`, `load`, `save`, `parse_json`, `format_json`, and more
- **String templates**: `{{expression}}` syntax highlighted distinctly
- **Regex patterns**: `~pattern~` syntax for pattern matching
- **Multiline strings**: Triple-quoted strings `"""..."""` and `'''...'''`
- **Comments**: `//` line comments and `/* */` block comments (nestable)
- **Constants**: `true`, `false`, `nil`
- **Operators**: All arithmetic, comparison, and logical operators
- **Numbers**: Integer and floating-point literals

## Example File

Create a file named `example.du`:

```duso
// Define a function
function greet(name)
  return "Hello, {{name}}!"
end

// Work with data
items = ["apple", "banana", "cherry"]
for item in items do
  print(item)
end

// Use builtins
numbers = [1, 2, 3, 4, 5]
doubled = map(numbers, function(x) return x * 2 end)
print(doubled)

// Multiline strings with templates
prompt = """
You are a helpful assistant.
User: {{user_input}}
Please respond in JSON format.
"""

// Error handling
try
  result = greet("World")
  print(result)
catch error
  print("Error: " + error)
end
```

## Development

To modify the syntax highlighting:

1. Edit `syntaxes/script.tmLanguage.json`
2. Reload VS Code window: `Cmd+Shift+P` → `Developer: Reload Window`
3. Changes appear immediately (if using symlink option)

To test scope detection:
1. Open any `.du` file
2. `Cmd+Shift+P` → `Developer: Inspect Editor Tokens and Scopes`
3. Click on text to see assigned scopes

## Resources

- **Repository**: https://github.com/duso-org/duso
- **Documentation**: https://duso.rocks
- **Issues**: https://github.com/duso-org/duso/issues
