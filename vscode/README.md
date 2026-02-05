# Duso VSCode Extension

Language Server Protocol (LSP) support for Duso in VSCode.

## Features

- **Syntax Highlighting** - Color coded syntax for Duso scripts
- **Error Diagnostics** - Real-time parse error detection
- **Hover Documentation** - View built-in function documentation
- **Go to Definition** - Jump to function and variable definitions (Ctrl+Click)
- **Find References** - Find all usages of variables and functions

## Installation

### From Source (Development)

```bash
cd vscode
npm install
code --extensionDevelopmentPath=. .
```

This opens a new VSCode window running the extension in development mode.

## Prerequisites

Ensure the `duso` binary is in your PATH:

```bash
which duso
duso --version
```

If duso is not found, the extension will fail to start. Make sure you've run `./build.sh` in the project root.

## Quick Test

1. Create a file `test.du`:
   ```duso
   print("Hello, Duso!")
   
   function greet(name) do
     print("Hello, " + name)
   end
   
   greet("World")
   ```

2. Open the file in VSCode - you should see syntax highlighting

3. Hover over `print` and `greet` to see documentation

4. Ctrl+Click on `greet("World")` to jump to the function definition

5. Create a syntax error to see diagnostics:
   ```duso
   print(1 2)  # Missing comma
   ```

## Features

### Error Diagnostics
- Syntax errors shown in real-time with red squiggles
- Error details in the Problems panel

### Hover Documentation
- Hover over built-in functions (print, len, map, etc.) to see docs
- Shows function signatures and descriptions

### Go to Definition
- Ctrl+Click on a function/variable to jump to its definition
- Works for same-file definitions in Phase 1

### Find References
- Right-click → Find All References to find all usages
- Highlights all matching identifiers

## Architecture

The extension launches `duso -lsp` which:
1. Reads LSP messages from VSCode over stdin
2. Parses documents and analyzes them
3. Sends diagnostics, hover info, and definitions back to VSCode

No separate server needed - it's built into the duso binary!

## Troubleshooting

If the server doesn't start:
1. Check Output panel: View → Output → Duso Language Server
2. Verify `duso` is in PATH: `which duso`
3. Rebuild: `./build.sh` in project root
4. Reload extension: Cmd+Shift+P → Developer: Reload Window

## Phase 1 Implementation

Currently implemented:
- ✅ Syntax highlighting
- ✅ Error diagnostics
- ✅ Hover documentation
- ✅ Go to definition (same-file)
- ✅ Find references (same-file)

Coming in Phase 2:
- Code completion
- Document symbols (outline)
- Rename refactoring
- Cross-file definitions
