# keys()

Get an array of all keys in an object.

## Signature

```duso
keys(object)
```

## Parameters

- `object` (object) - The object to get keys from

## Returns

Array of keys (strings) in the order they appear

## Examples

Get object keys:

```duso
config = {timeout = 30, retries = 3, debug = false}
k = keys(config)
print(k)                        // [timeout retries debug]
```

Iterate over keys:

```duso
user = {name = "Alice", age = 30, city = "NYC"}
for key in keys(user) do
  print(key)
end
```

Check for key existence:

```duso
settings = {darkMode = true, fontSize = 14}
all_keys = keys(settings)
if contains(join(all_keys, ","), "darkMode") then
  print("Dark mode setting exists")
end
```

## See Also

- [values() - Get object values](./values.md)
- [len() - Get object size](./len.md)
