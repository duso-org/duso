# random()

Get a random floating-point number between 0 and 1.

## Signature

```duso
random()
```

## Parameters

None

## Returns

Number between 0 (inclusive) and 1 (exclusive)

## Examples

Basic random number:

```duso
print(random())                 // 0.8374629...
print(random())                 // 0.2918472...
```

Random integers in a range:

```duso
// Random integer between 0 and 9
n = floor(random() * 10)
print(n)                        // 7

// Random integer between 1 and 100
n = floor(random() * 100) + 1
print(n)                        // 42
```

Random selection from array:

```duso
colors = ["red", "green", "blue", "yellow"]
idx = floor(random() * len(colors))
print(colors[idx])              // "blue"
```

Probability-based logic:

```duso
if random() < 0.5 then
  print("Heads")
else
  print("Tails")
end
```

## See Also

- [floor() - Round down](/docs/reference/floor.md)
- [min() - Find minimum](/docs/reference/min.md)
- [max() - Find maximum](/docs/reference/max.md)
