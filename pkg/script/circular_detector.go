package script

import (
	"errors"
	"fmt"
	"strings"
)

// CircularDetector tracks module loading stack to detect circular dependencies
type CircularDetector struct {
	stack []string // Paths currently being loaded
}

// Push adds a path to the detector's loading stack
func (c *CircularDetector) Push(path string) error {
	for i, existingPath := range c.stack {
		if existingPath == path {
			// Found circular dependency - build error message
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
	c.stack = append(c.stack, path)
	return nil
}

// Pop removes the most recent path from the detector's loading stack
func (c *CircularDetector) Pop() {
	if len(c.stack) > 0 {
		c.stack = c.stack[:len(c.stack)-1]
	}
}
