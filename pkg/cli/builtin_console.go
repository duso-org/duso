package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/duso-org/duso/pkg/runtime"
)

// builtinPrint prints values to output
func builtinPrint(evaluator *Evaluator, args map[string]any) (any, error) {
	var parts []string
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			parts = append(parts, fmt.Sprintf("%v", val))
		} else {
			break
		}
	}

	ClearBusySpinner()
	output := strings.Join(parts, " ")
	fmt.Println(output)
	return nil, nil
}

// builtinError prints error messages to stderr
func builtinError(evaluator *Evaluator, args map[string]any) (any, error) {
	var parts []string
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			parts = append(parts, fmt.Sprintf("%v", val))
		} else {
			break
		}
	}

	ClearBusySpinner()
	output := strings.Join(parts, " ")
	fmt.Fprintln(os.Stderr, output)
	return nil, nil
}

// builtinWrite writes to output without newline
func builtinWrite(evaluator *Evaluator, args map[string]any) (any, error) {
	var parts []string
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			parts = append(parts, fmt.Sprintf("%v", val))
		} else {
			break
		}
	}

	ClearBusySpinner()
	output := strings.Join(parts, " ")
	fmt.Print(output)
	return nil, nil
}

// builtinDebug prints debug messages
func builtinDebug(evaluator *Evaluator, args map[string]any) (any, error) {
	var parts []string
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			parts = append(parts, fmt.Sprintf("%v", val))
		} else {
			break
		}
	}

	ClearBusySpinner()
	output := "[DEBUG] " + strings.Join(parts, " ")
	fmt.Println(output)
	return nil, nil
}

// builtinInput reads a line from stdin with optional prompt
func builtinInput(evaluator *Evaluator, args map[string]any) (any, error) {
	ClearBusySpinner()
	// Optional prompt argument
	if prompt, ok := args["0"]; ok {
		fmt.Fprint(os.Stdout, prompt)
	}

	// Check if stdin is disabled via sys datastore
	sysDs := runtime.GetDatastore("sys", nil)
	noStdinVal, _ := sysDs.Get("-no-stdin")
	if noStdinVal != nil {
		if noStdin, ok := noStdinVal.(bool); ok && noStdin {
			fmt.Println("warning: stdin disabled, input() returned ''")
			return "", nil
		}
	}

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err == io.EOF {
		return nil, nil // EOF returns nil
	}
	if err != nil {
		return nil, fmt.Errorf("input() error: %v", err)
	}

	// Remove the trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	// Also remove carriage return if on Windows
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	return line, nil
}
