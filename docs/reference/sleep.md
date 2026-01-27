# sleep()

Pause execution for a specified duration.

## Signature

```duso
sleep(seconds)
```

## Parameters

- `seconds` (number, optional) - The number of seconds to sleep. Defaults to 1 second if not provided.

## Returns

Nothing (nil)

## Examples

Sleep for 1 second (default):

```duso
print("Starting...")
sleep()
print("Done!")                      // Prints after 1 second
```

Sleep for a specific duration:

```duso
print("Waiting 2 seconds...")
sleep(2)
print("Resuming")                   // Prints after 2 seconds
```

Sleep with fractional seconds:

```duso
sleep(0.5)                          // Sleep for half a second
sleep(2.5)                          // Sleep for 2.5 seconds
```

Using sleep in a loop:

```duso
for i = 0; i < 3; i = i + 1 {
    print("Tick " + string(i))
    sleep(1)
}
```

## See Also

- [time() - Get current Unix timestamp](./time.md)
