package script

// Position represents a location in source code
type Position struct {
	Line   int
	Column int
}

// NoPos represents an unknown or invalid position
var NoPos = Position{Line: 0, Column: 0}

// IsValid returns true if the position has valid line/column information
func (p Position) IsValid() bool {
	return p.Line > 0
}
