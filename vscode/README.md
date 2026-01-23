# Duso Language Extension for VS Code

Syntax highlighting for the Duso scripting language (`.du` files).

## Installation

1. Copy this folder to your VS Code extensions directory:
   - **macOS**: `~/.vscode/extensions/`
   - **Linux**: `~/.vscode/extensions/`
   - **Windows**: `%USERPROFILE%\.vscode\extensions\`

   Or rename to `duso-0.0.1` first.

2. Restart VS Code

3. Open any `.du` file and syntax highlighting will be applied automatically

## Features

- **Keywords**: `if`, `then`, `else`, `elseif`, `end`, `while`, `do`, `for`, `in`, `function`, `return`, `try`, `catch`, `structure`
- **String templates**: `{{expression}}` syntax highlighted distinctly
- **Multiline strings**: Triple-quoted strings `"""..."""` and `'''...'''`
- **Comments**: `--` line comments
- **Constants**: `true`, `false`, `nil`
- **Operators**: All arithmetic, comparison, and logical operators
- **Structures**: Capitalized structure names
- **Functions**: Function definitions and calls

## Example File

Create a file named `example.du`:

```duso
// Define a structure
structure Config
  timeout = 30
  retries = 3
end

// Create instances
config = Config(timeout = 60)

// Functions
function greet(name)
  return "Hello {{name}}!"
end

// Multiline strings
prompt = """
You are a helpful assistant.
Please respond in JSON format.
"""

// Arrays and control flow
items = ["apple", "banana", "cherry"]
for item in items do
  print(item)
end

// Try/catch
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
2. Reload VS Code window (Cmd+R / Ctrl+Shift+F5)

## TextMate Scope Reference

For reference on TextMate scope names used in this extension:
- `keyword.control` - Control flow keywords
- `keyword.operator` - Operators
- `entity.name.type` - Structure names
- `entity.name.function` - Function names
- `string.quoted.double` - Double-quoted strings
- `string.quoted.single` - Single-quoted strings
- `string.multiline` - Multiline strings
- `meta.template` - Template expressions `{{...}}`
- `constant.numeric` - Numbers
- `constant.language` - `true`, `false`, `nil`
- `comment.line` - Comments
- `punctuation` - Brackets, parentheses, etc.
