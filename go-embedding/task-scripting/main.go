package main

import (
	"fmt"
	"github.com/duso-org/duso/pkg/script"
)

// task-scripting: Orchestrate workflows using Duso scripts
//
// This example demonstrates:
// - Registering multiple custom functions
// - Defining workflows in Duso that use those functions
// - Executing the workflow
// - Capturing results via GetOutput()
//
// Real-world use case: ETL pipelines, task orchestration, agent workflows
//
// Run: go run ./task-scripting
func main() {
	interp := script.NewInterpreter(false)

	// Register functions that represent tasks
	interp.RegisterFunction("fetchData", func(args map[string]any) (any, error) {
		source := args["0"].(string)
		// Simulate fetching data
		return map[string]any{
			"rows": float64(100),
			"from": source,
		}, nil
	})

	interp.RegisterFunction("processData", func(args map[string]any) (any, error) {
		data := args["0"].(map[string]any)
		rows := int(data["rows"].(float64))
		// Simulate processing
		processed := map[string]any{
			"count": float64(rows),
			"valid": float64(rows * 95 / 100), // 95% valid
		}
		return processed, nil
	})

	interp.RegisterFunction("saveResult", func(args map[string]any) (any, error) {
		result := args["0"].(map[string]any)
		// Simulate saving
		return map[string]any{
			"success": true,
			"stored":  result["valid"],
		}, nil
	})

	// Define a workflow in Duso
	workflow := `
		print("Starting workflow...")

		// Step 1: Fetch data
		print("Step 1: Fetching data...")
		data = fetchData("database")
		print("Fetched " + data.rows + " rows")

		// Step 2: Process data
		print("Step 2: Processing data...")
		processed = processData(data)
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
			result = saveResult(processed)
			print("Saved " + result.stored + " records")
			print("Workflow complete: success")
		else
			print("Workflow complete: validation failed")
		end
	`

	// Execute the workflow
	_, err := interp.Execute(workflow)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
