# values()

Get an array of all values in an object.

## Signature

```duso
values(object)
```

## Parameters

- `object` (object) - The object to get values from

## Returns

Array of values in the order their keys appear

## Examples

Get object values:

```duso
config = {timeout = 30, retries = 3, debug = false}
vals = values(config)
print(vals)                     // [30 3 false]
```

Iterate over values:

```duso
scores = {alice = 95, bob = 87, charlie = 92}
for score in values(scores) do
  print(score)
end
```

Sum all values:

```duso
expenses = {rent = 1200, food = 400, utilities = 150}
total = 0
for amount in values(expenses) do
  total = total + amount
end
print(total)                    // 1750
```

## See Also

- [keys() - Get object keys](./keys.md)
- [len() - Get object size](./len.md)
