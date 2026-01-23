package main

import (
	"fmt"
	"github.com/duso-org/duso/pkg/script"
)

// 04-task-scripting: Orchestrate workflows using Duso scripts
//
// This example demonstrates:
// - Registering multiple custom functions
// - Defining workflows in Duso
// - Calling Duso functions from Go
// - Processing results
// - Control flow in scripts
//
// Real-world use case: ETL pipelines, task orchestration, agent workflows
//
// Run: go run 04-task-scripting.go
func main() {
	interp := script.NewInterpreter(false)

	// Register functions that represent tasks
	interp.RegisterFunction("fetchData", func(args map[string]any) (any, error) {
		source := args["source"].(string)
		// Simulate fetching data
		return map[string]any{
			"rows": float64(100),
			"from": source,
		}, nil
	})

	interp.RegisterFunction("processData", func(args map[string]any) (any, error) {
		data := args["data"].(map[string]any)
		rows := int(data["rows"].(float64))
		// Simulate processing
		processed := map[string]any{
			"count": float64(rows),
			"valid": float64(rows * 95 / 100), // 95% valid
		}
		return processed, nil
	})

	interp.RegisterFunction("saveResult", func(args map[string]any) (any, error) {
		result := args["result"].(map[string]any)
		// Simulate saving
		return map[string]any{
			"success": true,
			"stored":  result["valid"],
		}, nil
	})

	// Define a workflow in Duso
	workflow := `
		function runWorkflow()
			print("Starting workflow...")

			// Step 1: Fetch data
			print("Step 1: Fetching data...")
			data = fetchData(source = "database")
			print("Fetched " + data.rows + " rows")

			// Step 2: Process data
			print("Step 2: Processing data...")
			processed = processData(data = data)
			print("Valid records: " + processed.valid)

			// Step 3: Validate
			print("Step 3: Validating...")
			if processed.valid >= 90 then
				print("Validation passed")
				validated = true
			else
				print("Validation failed")
				validated = false
			end

			// Step 4: Save result (only if validated)
			if validated then
				print("Step 4: Saving result...")
				result = saveResult(result = processed)
				print("Saved " + result.stored + " records")
				return {success = true, message = "Workflow complete"}
			else
				return {success = false, message = "Validation failed"}
			end
		end
	`

	// Execute the workflow definition
	interp.Execute(workflow)

	// Call the Duso function from Go
	result, err := interp.Call("runWorkflow")
	if err != nil {
		fmt.Println("Error running workflow:", err)
		return
	}

	// Process the result
	if resultMap, ok := result.(map[string]any); ok {
		fmt.Println("\n=== Workflow Result ===")
		fmt.Printf("Success: %v\n", resultMap["success"])
		fmt.Printf("Message: %v\n", resultMap["message"])
	}
}
