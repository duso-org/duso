package main

import (
	"fmt"
	"github.com/duso-org/duso/pkg/script"
)

// hello-world: The simplest Duso embedding
//
// This example demonstrates:
// - Creating an interpreter
// - Executing a basic script
// - Handling errors
//
// Run: go run ./hello-world
func main() {
	// Create a new interpreter
	interp := script.NewInterpreter(false) // false = no debug output

	// Execute a simple Duso script
	_, err := interp.Execute(`
		name = "World"
		message = "Hello, " + name + "!"
		print(message)
	`)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
