package main

import (
	"fmt"
	"github.com/duso-org/duso/pkg/script"
)

// 03-config-dsl: Use Duso as a configuration language
//
// This example demonstrates:
// - Loading configuration from a Duso script
// - Parsing nested objects
// - Accessing values from Go
// - Configuration validation patterns
//
// Real-world use case: Config files, app settings, feature flags
//
// Run: go run 03-config-dsl.go
func main() {
	interp := script.NewInterpreter(false)

	// Configuration written in Duso (could be loaded from file)
	configScript := `
		// Database configuration
		database = {
			host = "localhost",
			port = 5432,
			name = "myapp",
			pool = {
				min = 5,
				max = 20
			}
		}

		// Server configuration
		server = {
			port = 8080,
			debug = true,
			timeout = 30
		}

		// Feature flags
		features = {
			newUI = true,
			betaAPI = false,
			analytics = true
		}
	`

	// Execute the configuration script
	err := interp.Execute(configScript)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Read configuration into Go
	dbConfigObj := interp.GetVariable("database")
	dbConfig := dbConfigObj.(map[string]any)

	serverConfigObj := interp.GetVariable("server")
	serverConfig := serverConfigObj.(map[string]any)

	featuresObj := interp.GetVariable("features")
	features := featuresObj.(map[string]any)

	// Access nested values
	poolObj := dbConfig["pool"].(map[string]any)

	// Display configuration
	fmt.Println("=== Database Configuration ===")
	fmt.Printf("Host: %v\n", dbConfig["host"])
	fmt.Printf("Port: %v\n", dbConfig["port"])
	fmt.Printf("Database: %v\n", dbConfig["name"])
	fmt.Printf("Pool Min: %v\n", poolObj["min"])
	fmt.Printf("Pool Max: %v\n", poolObj["max"])

	fmt.Println("\n=== Server Configuration ===")
	fmt.Printf("Port: %v\n", serverConfig["port"])
	fmt.Printf("Debug: %v\n", serverConfig["debug"])
	fmt.Printf("Timeout: %v\n", serverConfig["timeout"])

	fmt.Println("\n=== Features ===")
	fmt.Printf("New UI: %v\n", features["newUI"])
	fmt.Printf("Beta API: %v\n", features["betaAPI"])
	fmt.Printf("Analytics: %v\n", features["analytics"])

	// Validate configuration
	port := int(serverConfig["port"].(float64))
	if port < 1 || port > 65535 {
		fmt.Println("ERROR: Invalid port number")
		return
	}

	fmt.Println("\nConfiguration loaded and validated successfully!")
}
