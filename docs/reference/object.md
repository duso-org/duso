# Object

Objects are key-value maps. Keys are identifiers and values can be any type, including functions.

## Creating Objects

```duso
person = {
  name = "Alice",
  age = 30,
  city = "Portland"
}

config = {
  timeout = 30,
  retries = 3,
  verbose = true
}

empty = {}
```

## Accessing Values

Access with dot notation or brackets:

```duso
obj = {name = "Alice", age = 30}
print(obj.name)          // "Alice"
print(obj["name"])       // "Alice"
```

## Modifying Values

Change existing values:

```duso
obj = {count = 0}
obj.count = 5
obj["count"] = 10
```

## Length

Get the number of keys with [`len()`](len.md):

```duso
obj = {a = 1, b = 2, c = 3}
print(len(obj))  // 3
```

## Extracting Keys and Values

- [`keys()`](keys.md) - Get array of all keys
- [`values()`](values.md) - Get array of all values

```duso
obj = {name = "Alice", age = 30}
print(keys(obj))    // [name age]
print(values(obj))  // [Alice 30]
```

## Iteration

Loop through object keys with `for...in`:

```duso
config = {host = "localhost", port = 8080}
for key in config do
  print(key)
end
```

## Objects as Constructors

Objects can be called like functions to create new copies with field overrides:

```duso
Config = {timeout = 30, retries = 3}
config1 = Config()
config2 = Config(timeout = 60)  // Override specific field
```

This pattern is useful for creating instances from blueprints.

## Objects with Methods

Objects can contain functions that act as methods:

```duso
agent = {
  name = "Alice",
  greet = function(msg)
    print(msg + ", I am " + name)
  end
}

agent.greet("Hello")  // "Hello, I am Alice"
```

## Truthiness

In conditions, non-empty objects are truthy:

```duso
if {a = 1} then print("true") end  // prints
if {} then print("true") end       // doesn't print
```

