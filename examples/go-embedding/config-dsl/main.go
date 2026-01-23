package main

import (
	"fmt"
	"github.com/duso-org/duso/pkg/script"
)

// config-dsl: Use Duso as a configuration language
//
// This example demonstrates:
// - Writing configuration in Duso
// - Executing the config script
// - Reading the output via GetOutput()
// - Parsing config values
//
// Real-world use case: Config files, app settings, feature flags
//
// Run: go run ./config-dsl
func main() {
	interp := script.NewInterpreter(false)

	// Configuration written in Duso (could be loaded from file)
	configScript := `
		// Database configuration
		config = {
			database = {
				host = "localhost",
				port = 5432,
				name = "myapp"
			},
			server = {
				port = 8080,
				debug = true,
				timeout = 30
			},
			features = {
				newUI = true,
				betaAPI = false
			}
		}

		// Print configuration for debugging
		print("Configuration loaded successfully")
	`

	// Execute the configuration script
	_, err := interp.Execute(configScript)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
