package runtime

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// builtinPrint prints values to stdout with newline
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

	output := strings.Join(parts, " ")
	fmt.Println(output)
	return nil, nil
}

// builtinInput reads a line from stdin with optional prompt
func builtinInput(evaluator *Evaluator, args map[string]any) (any, error) {
	// Optional prompt argument
	if prompt, ok := args["0"]; ok {
		fmt.Fprint(os.Stdout, prompt)
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
