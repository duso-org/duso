package main

import (
	"fmt"
	"github.com/duso-org/duso/pkg/script"
)

// 02-custom-functions: Register Go functions callable from Duso
//
// This example demonstrates:
// - Registering custom Go functions
// - Calling them from Duso scripts
// - Handling named arguments
// - Returning values
//
// Run: go run 02-custom-functions.go
func main() {
	interp := script.NewInterpreter(false)

	// Register a custom function: add(a, b) -> a + b
	interp.RegisterFunction("add", func(args map[string]any) (any, error) {
		a := args["a"].(float64)
		b := args["b"].(float64)
		return a + b, nil
	})

	// Register another function: greet(name) -> "Hello, <name>"
	interp.RegisterFunction("greet", func(args map[string]any) (any, error) {
		name := args["name"].(string)
		return "Hello, " + name + "!", nil
	})

	// Register a function that returns an object
	interp.RegisterFunction("person", func(args map[string]any) (any, error) {
		return map[string]any{
			"name": args["name"].(string),
			"age":  args["age"].(float64),
		}, nil
	})

	// Execute script that uses custom functions
	interp.Execute(`
		// Call custom functions from Duso
		sum = add(a = 10, b = 20)
		print("10 + 20 = " + sum)

		greeting = greet(name = "Alice")
		print(greeting)

		// Create an object using custom function
		user = person(name = "Bob", age = 30)
		print(user.name + " is " + user.age + " years old")
	`)
}
