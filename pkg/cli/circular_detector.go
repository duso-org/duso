package cli

import (
	"errors"
	"fmt"
	"strings"
)

// CircularDetector tracks the module loading stack to detect circular dependencies.
// When a module loads another module (via require or include), we push the path.
// If we encounter a path that's already on the stack, we have a circular dependency.
type CircularDetector struct {
	stack []string // Paths currently being loaded
}

// Push adds a path to the loading stack.
// Returns an error if the path is already being loaded (circular dependency).
func (c *CircularDetector) Push(path string) error {
	// Check if this path is already in the stack
	for i, existingPath := range c.stack {
		if existingPath == path {
			// Found a circular dependency - build error message showing the cycle
			var cycleStr strings.Builder
			cycleStr.WriteString("circular dependency detected\n")
			for j := i; j < len(c.stack); j++ {
				cycleStr.WriteString(fmt.Sprintf("  %s\n", c.stack[j]))
				cycleStr.WriteString("  → ")
			}
			cycleStr.WriteString(fmt.Sprintf("%s (circular)\n", path))

			return errors.New(cycleStr.String())
		}
	}

	// Not found, add to stack
	c.stack = append(c.stack, path)
	return nil
}

// Pop removes the most recent path from the loading stack.
// Call this after finishing loading a module.
func (c *CircularDetector) Pop() {
	if len(c.stack) > 0 {
		c.stack = c.stack[:len(c.stack)-1]
	}
}
