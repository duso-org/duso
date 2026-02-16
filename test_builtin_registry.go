package main

import (
	"fmt"
	"github.com/duso-org/duso/pkg/runtime"
	"github.com/duso-org/duso/pkg/script"
)

func main() {
	// Register builtins
	runtime.RegisterBuiltins()

	// Create interpreter
	interp := script.NewInterpreter(false)

	// Test: print and deep_copy
	output, err := interp.Execute(`
x = [1, 2, 3]
y = deep_copy(x)
print("Copied:", y)
print("Deep copy works!")
`)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Output: %s\n", output)
	fmt.Println("✓ Builtin registry approach works!")

	// Test: env builtin
	fmt.Println("\nTesting env() builtin...")
	output2, err := interp.Execute(`
path = env("PATH")
print("✓ env() works! Got PATH:", path != "")
`)

	if err != nil {
		fmt.Printf("Error testing env(): %v\n", err)
		return
	}

	fmt.Printf("env() test output: %s\n", output2)

	// Test: print with various types
	fmt.Println("\nTesting print() with various types...")
	_, err = interp.Execute(`
x = 42
y = "hello"
z = true
arr = [1, 2, 3]
print("Number:", x)
print("String:", y)
print("Boolean:", z)
print("Array:", arr)
print("Multiple:", x, y, z)
`)

	if err != nil {
		fmt.Printf("Error testing print(): %v\n", err)
		return
	}

	fmt.Printf("print() test passed!\n")

	// Test: string functions
	fmt.Println("\nTesting string functions...")
	_, err = interp.Execute(`
text = "Hello World"
print("Original:", text)
print("Upper:", upper(text))
print("Lower:", lower(text))
print("Substr:", substr(text, 0, 5))
print("Trim:", trim("  spaces  "))
`)

	if err != nil {
		fmt.Printf("Error testing strings: %v\n", err)
		return
	}

	fmt.Printf("string functions test passed!\n")

	// Test: math functions
	fmt.Println("\nTesting math functions...")
	_, err = interp.Execute(`
print("floor(3.7):", floor(3.7))
print("ceil(3.2):", ceil(3.2))
print("round(3.5):", round(3.5))
print("abs(-5):", abs(-5))
print("min(3, 1, 4):", min(3, 1, 4))
print("max(3, 1, 4):", max(3, 1, 4))
print("sqrt(16):", sqrt(16))
print("pow(2, 3):", pow(2, 3))
print("clamp(5, 1, 3):", clamp(5, 1, 3))
`)

	if err != nil {
		fmt.Printf("Error testing math: %v\n", err)
		return
	}

	fmt.Printf("math functions test passed!\n")

	// Test: date functions
	fmt.Println("\nTesting date functions...")
	_, err = interp.Execute(`
ts = now()
print("now():", ts > 0)
formatted = format_time(ts, "date")
print("format_time(now, date):", formatted != "")
parsed = parse_time("2024-01-15")
print("parse_time works:", parsed > 0)
`)

	if err != nil {
		fmt.Printf("Error testing date: %v\n", err)
		return
	}

	fmt.Printf("date functions test passed!\n")

	// Test: type functions
	fmt.Println("\nTesting type functions...")
	_, err = interp.Execute(`
print("len([1,2,3]):", len([1, 2, 3]))
print("len('hello'):", len("hello"))
print("type(42):", type(42))
print("type('hi'):", type("hi"))
print("tonumber('123'):", tonumber("123"))
print("tostring(456):", tostring(456))
print("tobool(1):", tobool(1))
`)

	if err != nil {
		fmt.Printf("Error testing type: %v\n", err)
		return
	}

	fmt.Printf("type functions test passed!\n")

	// Test: array functions
	fmt.Println("\nTesting array functions...")
	_, err = interp.Execute(`
arr = [1, 2, 3]
print("Initial:", arr)
push(arr, 4)
print("After push:", arr)
pop(arr)
print("After pop:", arr)
print("join:", join(arr, ","))
print("split:", split("a,b,c", ","))
print("range:", range(1, 3))
`)

	if err != nil {
		fmt.Printf("Error testing array: %v\n", err)
		return
	}

	fmt.Printf("array functions test passed!\n")

	// Test: functional operations
	fmt.Println("\nTesting functional operations...")
	_, err = interp.Execute(`
arr = [1, 2, 3, 4, 5]
doubled = map(arr, function(x) return x * 2 end)
print("map (double):", doubled)
evens = filter(arr, function(x) return x % 2 == 0 end)
print("filter (evens):", evens)
sum = reduce(arr, function(acc, x) return acc + x end, 0)
print("reduce (sum):", sum)
`)

	if err != nil {
		fmt.Printf("Error testing functional: %v\n", err)
		return
	}

	fmt.Printf("functional operations test passed!\n")

	// Test: regex operations
	fmt.Println("\nTesting regex operations...")
	_, err = interp.Execute(`
text = "Hello World 123"
print("contains 'Hello':", contains(text, "Hello"))
print("contains 'world' (case-insensitive):", contains(text, "world", ignore_case = true))
matches = find(text, "[0-9]+")
print("find numbers count:", len(matches))
replaced = replace(text, "[0-9]+", "###")
print("replace numbers:", replaced)
`)

	if err != nil {
		fmt.Printf("Error testing regex: %v\n", err)
		return
	}

	fmt.Printf("regex operations test passed!\n")

	// Test: template operations
	fmt.Println("\nTesting template operations...")
	_, err = interp.Execute(`
tmpl = template("Hello {{name}}, you have {{count}} messages")
result = tmpl(name = "Alice", count = 5)
print("Template result:", result)
`)

	if err != nil {
		fmt.Printf("Error testing template: %v\n", err)
		return
	}

	fmt.Printf("template operations test passed!\n")

	// Test: system functions
	fmt.Println("\nTesting system functions...")
	_, err = interp.Execute(`
print("uuid():", len(uuid()) > 0)
sleep(0.01)
print("sleep(0.01) completed")
`)

	if err != nil {
		fmt.Printf("Error testing system: %v\n", err)
		return
	}

	fmt.Printf("system functions test passed!\n")

	// Test: fetch function
	fmt.Println("\nTesting fetch function...")
	_, err = interp.Execute(`
response = fetch("https://httpbin.org/get")
print("fetch status:", response.status)
print("fetch ok:", response.ok)
print("fetch has body:", len(response.body) > 0)
text_result = response.text()
print("response.text() works:", len(text_result) > 0)
data = response.json()
print("response.json() works:", type(data) == "object")

response_post = fetch("https://httpbin.org/post", {method = "POST", body = "test=data"})
print("fetch POST status:", response_post.status)
print("fetch POST ok:", response_post.ok)
`)

	if err != nil {
		fmt.Printf("Error testing fetch: %v\n", err)
		return
	}

	fmt.Printf("fetch function test passed!\n")

	// Test: datastore function
	fmt.Println("\nTesting datastore function...")
	_, err = interp.Execute(`
store = datastore("test")
store.set("key1", "value1")
val = store.get("key1")
print("datastore set/get:", val == "value1")
store.set("counter", 0)
store.increment("counter")
store.increment("counter", 5)
count = store.get("counter")
print("datastore increment:", count == 6)
store.push("items", "item1")
store.push("items", "item2")
items = store.get("items")
print("datastore push:", len(items) == 2)
`)

	if err != nil {
		fmt.Printf("Error testing datastore: %v\n", err)
		return
	}

	fmt.Printf("datastore function test passed!\n")
}
