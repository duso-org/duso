# busy()

Display an animated spinner with a status message. Useful for long-running operations to show progress to the user.

## Signature

```duso
busy(message)
```

## Parameters

- `message` - String message to display with the spinner (required)

## Returns

`nil`

## Examples

Basic usage with sleep:

```duso
busy("Processing data")
sleep(2)
print("Done!")
```

Sequential messages (spinner automatically clears on new message):

```duso
busy("Step 1: Downloading")
sleep(2)
busy("Step 2: Extracting")
sleep(2)
busy("Step 3: Installing")
sleep(2)
print("Complete!")
```

With other operations:

```duso
write("Operation: ")
busy("fetching results")
data = fetch("https://api.example.com/data")
print("Success!")
```

## How It Works

- Displays a message followed by an animated Braille spinner character
- Uses stderr for output (doesn't pollute stdout for redirection)
- Hides the terminal cursor while animating for cleaner display
- Automatically clears the spinner when:
  - A new `busy()` call is made
  - `print()`, `write()`, `error()`, or `debug()` is called
  - The script completes or calls `input()`

## Notes

- The spinner animation uses Braille Unicode characters: ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
- Only works properly in interactive terminals; safe to use in scripts that may be piped
- The message should be concise (typically under 50 characters for best visual appearance)
- Multiple consecutive `busy()` calls replace the previous message seamlessly
- No explicit clearing is needed; just call another function or `busy()` with a new message

## See Also

- [print()](/docs/reference/print.md) - Output with newline
- [write()](/docs/reference/write.md) - Output without newline
- [sleep()](/docs/reference/sleep.md) - Pause execution
