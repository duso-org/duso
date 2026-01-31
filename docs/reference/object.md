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

Objects can contain functions that act as methods. When you call a method with the dot notation (`obj.method()`), the object is automatically bound as `self`, and the method can access the object's properties as variables.

```duso
agent = {
  name = "Alice",
  age = 30,
  greet = function(msg)
    print(msg + ", I am " + name + " (age " + age + ")")
  end
}

agent.greet("Hello")  // "Hello, I am Alice (age 30)"
```

The method accesses `name` and `age` from the object it's called on. This same function can work with different objects:

### Modifying Properties in Methods

Methods can modify the object's properties through self binding:

```duso
counter = {
  count = 0,
  increment = function()
    count = count + 1
  end,
  get_count = function()
    return count
  end
}

counter.increment()
counter.increment()
print(counter.get_count())  // 2
```

### Methods Calling Other Methods

Methods can call other methods on the same object. When one method calls another through a variable, self is automatically preserved:

```duso
calculator = {
  value = 10,
  add = function(x)
    value = value + x
  end,
  multiply = function(x)
    value = value * x
  end,
  double = function()
    multiply(2)  // Calls the multiply method on the same object
  end,
  get = function()
    return value
  end
}

calculator.add(5)      // value = 15
calculator.double()    // value = 30
print(calculator.get())  // 30
```

### Method Reuse Through Objects

Because methods use dynamic self binding rather than static closures, the same method can work with different objects. This is useful with the constructor pattern:

```duso
Counter = {
  count = 0,
  increment = function()
    count = count + 1
  end,
  get = function()
    return count
  end
}

c1 = Counter()
c2 = Counter()

c1.increment()
c1.increment()
c2.increment()

print(c1.get())  // 2
print(c2.get())  // 1
```

Each instance has its own `count`, and the same `increment` method works on both.

## Truthiness

In conditions, non-empty objects are truthy:

```duso
if {a = 1} then print("true") end  // prints
if {} then print("true") end       // doesn't print
```

