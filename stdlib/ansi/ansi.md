# ansi - ANSI Color and Styling Utilities

Provides semantic style functions for terminal colors and text formatting.

## Usage

```duso
ansi = require("ansi")

print("{{ansi.error('Error message')}}")
print("{{ansi.warning('Warning message')}}")
print("{{ansi.title('Section Title')}}")
print("{{ansi.code('some_function()')}}")
```

## Semantic Styles

- **error** - Bold red
- **warning** - Bold yellow
- **title** - Bold cyan
- **code** - Green (for code snippets)
- **blockquote** - Gray italic
- **comment** - Gray
- **link** - Blue underlined

## Convenience Functions

- **highlight** - Black text on bright yellow background (for emphasis)
- **success** - Bold green (for positive messages)
- **info** - Cyan (for informational messages)

## Building Custom Styles

Use `combine()` to build ANSI codes with multiple attributes including background colors:

```duso
ansi = require("ansi")

myStyle = ansi.combine(
  fg="magenta",
  bg="black",
  bold=true,
  underline=true
)

print("{{myStyle}}Custom styled text{{ansi.clear}}")

// Background color examples
alert = ansi.combine(fg="white", bg="red", bold=true)
highlight = ansi.combine(fg="black", bg="bright_yellow")

print("{{alert}}Critical alert{{ansi.clear}}")
print("{{highlight}}Important{{ansi.clear}}")
```

### combine() Parameters

- `fg` - Foreground color name
- `bg` - Background color name
- `bold` - Bold text (boolean)
- `dim` - Dim/faint text (boolean)
- `italic` - Italic text (boolean)
- `underline` - Underlined text (boolean)
- `blink` - Blinking text (boolean)
- `reverse` - Reverse video (boolean)

### Color Names

**Standard colors:** black, red, green, yellow, blue, magenta, cyan, white, gray

**Bright colors:** bright_red, bright_green, bright_yellow, bright_blue, bright_magenta, bright_cyan, bright_white

## Constants

- `ansi.clear` - Reset all formatting to default
