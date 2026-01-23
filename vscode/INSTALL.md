# Install Duso Syntax Highlighter in VS Code

## Quick Install (Manual)

1. **Copy the extension to VS Code:**
   ```bash
   cp -r /Users/dbalmer/Projects/shannan/fulcrum/vscode ~/.vscode/extensions/duso-0.0.1
   ```

2. **Restart VS Code** (close and reopen)

3. **Create or open a `.du` file** to see syntax highlighting in action

## Or: Create a Symlink (Development)

For faster iteration during development:

```bash
ln -s /Users/dbalmer/Projects/shannan/fulcrum/vscode ~/.vscode/extensions/duso
```

Then reload VS Code window with `Cmd+Shift+P` → "Developer: Reload Window"

## Verify Installation

1. Open the example file: `/Users/dbalmer/Projects/shannan/fulcrum/vscode/example.du`
2. You should see:
   - Keywords highlighted in blue (if, then, function, etc.)
   - Strings in green
   - Template expressions `{{...}}` with special highlighting
   - Comments in gray
   - Numbers in different color

## Convert Example Scripts

To use the syntax highlighter with existing examples:

```bash
cd /Users/dbalmer/Projects/shannan/fulcrum/script/examples

# Rename or copy .du files to .du
for f in *.du; do
  cp "$f" "${f%.du}.du"
done
```

Then open any `.du` file in VS Code for syntax highlighting.

## Troubleshooting

**Syntax highlighting not showing?**
1. Check file extension is `.du`
2. Verify extension installed to correct location: `~/.vscode/extensions/duso*`
3. Restart VS Code completely (close all windows)
4. Try opening the `example.du` file to verify it works there

**Want to debug the grammar?**
1. Install "TextMate Language Grammar" support (should be default)
2. Open Command Palette: `Cmd+Shift+P`
3. Search "Developer: Inspect Editor Tokens and Scopes"
4. Click on text to see what scope is applied
