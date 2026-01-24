# exit()

Exit the program with a status code.

## Signature

```duso
exit([code])
```

## Parameters

- `code` (optional, number) - Exit code. Defaults to 0 (success)

## Returns

Never returns (program exits)

## Examples

Exit with success:

```duso
print("Done!")
exit()                          // Exit with code 0
```

Exit with error code:

```duso
if error_condition then
  print("An error occurred")
  exit(1)                       // Exit with code 1
end
```

Conditional exit:

```duso
result = perform_action()
if result == false then
  print("Action failed")
  exit(1)
end
print("Success")
exit(0)
```

## Notes

- Exit code 0 typically indicates success
- Non-zero codes indicate various errors
- Use meaningful exit codes for different failure scenarios

## See Also

- [print() - Output text](./print.md)
