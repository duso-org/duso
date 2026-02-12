package lsp

import (
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// Position represents a position in a document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// DusoPositionToLSP converts Duso's 1-based Position to LSP 0-based Position
// Duso uses 1-based line and column numbers, LSP uses 0-based
func DusoPositionToLSP(pos script.Position) Position {
	if !pos.IsValid() {
		return Position{Line: 0, Character: 0}
	}
	return Position{
		Line:      pos.Line - 1,
		Character: pos.Column - 1,
	}
}

// LSPPositionToDuso converts LSP 0-based Position to Duso's 1-based Position
func LSPPositionToDuso(pos Position) script.Position {
	return script.Position{
		Line:   pos.Line + 1,
		Column: pos.Character + 1,
	}
}

// StringToURI converts a file path to a URI
func StringToURI(path string) string {
	// Simple conversion: file://path
	// On Windows, need file:///C:/path format
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	return "file://" + path
}

// URIToString converts a URI to a file path
func URIToString(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		path := uri[7:]
		// On Windows, file:///C:/path has an extra slash
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
		return path
	}
	return uri
}

// GetLineFromText returns the text of a specific line (0-based)
func GetLineFromText(text string, lineNum int) string {
	lines := strings.Split(text, "\n")
	if lineNum >= 0 && lineNum < len(lines) {
		return lines[lineNum]
	}
	return ""
}

// GetTextRange returns the text of a specific range
func GetTextRange(text string, r Range) string {
	lines := strings.Split(text, "\n")

	if r.Start.Line > len(lines)-1 {
		return ""
	}

	if r.Start.Line == r.End.Line {
		line := lines[r.Start.Line]
		// Convert UTF-16 code unit offset to byte offset
		start := utf16OffsetToByteOffset(line, r.Start.Character)
		end := utf16OffsetToByteOffset(line, r.End.Character)
		if start < 0 || end < 0 {
			return ""
		}
		if start > len(line) {
			start = len(line)
		}
		if end > len(line) {
			end = len(line)
		}
		if start > end {
			start, end = end, start
		}
		return line[start:end]
	}

	// Multi-line range
	var result strings.Builder

	// Add from start of first line
	firstLine := lines[r.Start.Line]
	start := utf16OffsetToByteOffset(firstLine, r.Start.Character)
	if start < 0 {
		start = 0
	}
	if start > len(firstLine) {
		start = len(firstLine)
	}
	result.WriteString(firstLine[start:])
	result.WriteByte('\n')

	// Add all intermediate lines
	for i := r.Start.Line + 1; i < r.End.Line; i++ {
		if i < len(lines) {
			result.WriteString(lines[i])
			result.WriteByte('\n')
		}
	}

	// Add to end of last line
	lastLine := lines[r.End.Line]
	end := utf16OffsetToByteOffset(lastLine, r.End.Character)
	if end < 0 {
		end = 0
	}
	if end > len(lastLine) {
		end = len(lastLine)
	}
	result.WriteString(lastLine[:end])

	return result.String()
}

// utf16OffsetToByteOffset converts a UTF-16 code unit offset to a byte offset in a line
// LSP uses UTF-16 code units for character positions
func utf16OffsetToByteOffset(line string, utf16Offset int) int {
	// Convert UTF-16 code unit offset to rune index, then to byte offset
	runes := []rune(line)
	utf16Pos := 0

	for i, r := range runes {
		if utf16Pos >= utf16Offset {
			// Convert rune index to byte offset
			return len(string(runes[:i]))
		}
		// Runes outside BMP (>= 0x10000) take 2 UTF-16 code units
		if r > 0xFFFF {
			utf16Pos += 2
		} else {
			utf16Pos += 1
		}
	}

	return len(line)
}

// byteOffsetToUTF16Offset converts a byte offset to UTF-16 code units
func byteOffsetToUTF16Offset(line string, byteOffset int) int {
	if byteOffset > len(line) {
		byteOffset = len(line)
	}

	runes := []rune(line[:byteOffset])
	utf16Pos := 0

	for _, r := range runes {
		if r > 0xFFFF {
			utf16Pos += 2
		} else {
			utf16Pos += 1
		}
	}

	return utf16Pos
}

// RangeContainsPosition checks if a range contains a position
func RangeContainsPosition(r Range, pos Position) bool {
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}

	if pos.Line == r.Start.Line && pos.Character < r.Start.Character {
		return false
	}

	if pos.Line == r.End.Line && pos.Character > r.End.Character {
		return false
	}

	return true
}

// FindNodeAtPosition walks an AST to find a node at the given position
// Returns the deepest node containing the position
func FindNodeAtPosition(node script.Node, pos script.Position) script.Node {
	if node == nil {
		return nil
	}

	// Check if position is within this node
	var nodeRange *Range
	var nodePos *script.Position

	// Extract position from different node types
	switch n := node.(type) {
	case *script.Program:
		// Program contains all statements, recurse into them
		var deepest script.Node
		for _, stmt := range n.Statements {
			if found := FindNodeAtPosition(stmt, pos); found != nil {
				deepest = found
			}
		}
		return deepest

	case *script.IfStatement:
		// For compound statements, don't require position match on the "if" keyword
		// Check condition and branches
		if found := FindNodeAtPosition(n.Condition, pos); found != nil {
			return found
		}
		for _, stmt := range n.Then {
			if found := FindNodeAtPosition(stmt, pos); found != nil {
				return found
			}
		}
		for _, elseif := range n.Elseifs {
			if found := FindNodeAtPosition(elseif.Condition, pos); found != nil {
				return found
			}
			for _, stmt := range elseif.Then {
				if found := FindNodeAtPosition(stmt, pos); found != nil {
					return found
				}
			}
		}
		for _, stmt := range n.Else {
			if found := FindNodeAtPosition(stmt, pos); found != nil {
				return found
			}
		}

	case *script.WhileStatement:
		// For compound statements, don't require position match on the "while" keyword
		if found := FindNodeAtPosition(n.Condition, pos); found != nil {
			return found
		}
		for _, stmt := range n.Body {
			if found := FindNodeAtPosition(stmt, pos); found != nil {
				return found
			}
		}

	case *script.ForStatement:
		// For compound statements, don't require position match on the "for" keyword
		// Try body first (where most code is), then loop bounds
		for _, stmt := range n.Body {
			if found := FindNodeAtPosition(stmt, pos); found != nil {
				return found
			}
		}
		// Then try loop initialization/bounds
		if found := FindNodeAtPosition(n.Start, pos); found != nil {
			return found
		}
		if found := FindNodeAtPosition(n.End, pos); found != nil {
			return found
		}
		if found := FindNodeAtPosition(n.Step, pos); found != nil {
			return found
		}
		if found := FindNodeAtPosition(n.Iterator, pos); found != nil {
			return found
		}

	case *script.FunctionDef:
		// For compound statements, don't require position match on the "function" keyword
		for _, stmt := range n.Body {
			if found := FindNodeAtPosition(stmt, pos); found != nil {
				return found
			}
		}

	case *script.ReturnStatement:
		// Recurse into the return value
		if n.Value != nil {
			if found := FindNodeAtPosition(n.Value, pos); found != nil {
				return found
			}
		}

	case *script.Identifier:
		nodePos = &n.Pos
		if positionMatch(nodePos, pos) {
			// Check if cursor is actually on this identifier
			// Identifier spans from start column to start column + name length
			if pos.Column >= n.Pos.Column && pos.Column < n.Pos.Column+len(n.Name) {
				return n
			}
		}

	case *script.AssignStatement:
		nodePos = &n.Pos
		if positionMatch(nodePos, pos) {
			// Try Value first (right side of assignment) to prioritize hover on function calls
			if found := FindNodeAtPosition(n.Value, pos); found != nil {
				return found
			}
			// Then try Target (left side)
			if found := FindNodeAtPosition(n.Target, pos); found != nil {
				return found
			}
			// Don't return self if children didn't match - let parent continue searching
		}

	case *script.CallExpr:
		nodePos = &n.Pos
		if positionMatch(nodePos, pos) {
			// For CallExpr, try arguments first (more likely to have nested calls)
			// Only use function if we're very close to it
			for _, arg := range n.Arguments {
				if found := FindNodeAtPosition(arg, pos); found != nil {
					return found
				}
			}
			// Only try function if arguments didn't match
			if found := FindNodeAtPosition(n.Func, pos); found != nil {
				return found
			}
			// Don't return self if children didn't match - let parent continue searching
		}

	case *script.BinaryExpr:
		nodePos = &n.Pos
		if positionMatch(nodePos, pos) {
			// Try both sides - right first since it's usually later
			if found := FindNodeAtPosition(n.Right, pos); found != nil {
				return found
			}
			if found := FindNodeAtPosition(n.Left, pos); found != nil {
				return found
			}
			// Don't return self if children didn't match - let parent continue searching
		}

	case *script.TryStatement:
		// Recurse into the try block
		for _, stmt := range n.Block {
			if found := FindNodeAtPosition(stmt, pos); found != nil {
				return found
			}
		}
		// Recurse into catch block if it exists
		for _, stmt := range n.CatchBlock {
			if found := FindNodeAtPosition(stmt, pos); found != nil {
				return found
			}
		}

	case *script.CompoundAssignStatement:
		nodePos = &n.Pos
		if positionMatch(nodePos, pos) {
			// Try Value first (right side)
			if found := FindNodeAtPosition(n.Value, pos); found != nil {
				return found
			}
			// Then try Target (left side)
			if found := FindNodeAtPosition(n.Target, pos); found != nil {
				return found
			}
		}

	case *script.PostIncrementStatement:
		nodePos = &n.Pos
		if positionMatch(nodePos, pos) {
			if found := FindNodeAtPosition(n.Target, pos); found != nil {
				return found
			}
		}
	}

	_ = nodeRange // Suppress unused warning

	return nil
}

// positionMatch checks if a node's position is on the same line as target
// For simple nodes (Identifier, etc), this will match precisely
// For compound statements, this just checks if we're in the right area
func positionMatch(nodePos *script.Position, pos script.Position) bool {
	if nodePos == nil || !nodePos.IsValid() {
		return false
	}
	// Only check line match - let recursion handle column specificity
	return nodePos.Line == pos.Line
}
