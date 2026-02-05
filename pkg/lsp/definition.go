package lsp

import (
	"github.com/duso-org/duso/pkg/script"
)

// DefinitionParams represents parameters for the definition request
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// TextDocumentIdentifier identifies a document
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// ProvideDefinition returns the definition location for an identifier
func ProvideDefinition(doc *Document, pos Position, entry *ASTCacheEntry) *Location {
	if entry == nil || entry.AST == nil {
		return nil
	}

	// Convert LSP position to Duso position
	dusoPos := LSPPositionToDuso(pos)

	// Find the node at the position
	node := FindNodeAtPosition(entry.AST, dusoPos)
	if node == nil {
		return nil
	}

	// Get the identifier name
	var identName string
	switch n := node.(type) {
	case *script.Identifier:
		identName = n.Name
	case *script.CallExpr:
		// For function calls, look at the function being called
		if ident, ok := n.Func.(*script.Identifier); ok {
			identName = ident.Name
		}
	default:
		return nil
	}

	if identName == "" {
		return nil
	}

	// Search for the definition in the same file
	// Phase 1: Same-file only
	location := findDefinitionInAST(entry.AST, identName, doc.URI)
	return location
}

// findDefinitionInAST searches for a function or variable definition
func findDefinitionInAST(ast *script.Program, name string, uri string) *Location {
	if ast == nil {
		return nil
	}

	// Walk statements to find function definition or first assignment
	var firstAssignment *script.AssignStatement
	var assignmentFound bool

	for _, stmt := range ast.Statements {
		// Check for function definition
		if funcDef, ok := stmt.(*script.FunctionDef); ok {
			if funcDef.Name == name {
				return &Location{
					URI: uri,
					Range: Range{
						Start: DusoPositionToLSP(funcDef.Pos),
						End:   DusoPositionToLSP(funcDef.Pos),
					},
				}
			}
		}

		// Track first assignment for variables
		if !assignmentFound {
			if assign, ok := stmt.(*script.AssignStatement); ok {
				if ident, ok := assign.Target.(*script.Identifier); ok {
					if ident.Name == name {
						firstAssignment = assign
						assignmentFound = true
					}
				}
			}
		}
	}

	// Return first assignment if found
	if firstAssignment != nil {
		return &Location{
			URI: uri,
			Range: Range{
				Start: DusoPositionToLSP(firstAssignment.Pos),
				End:   DusoPositionToLSP(firstAssignment.Pos),
			},
		}
	}

	// No definition found in source (could be a built-in or undefined)
	return nil
}

// FindReferences finds all references to an identifier in the AST
// Phase 1: Same-file only
func FindReferences(ast *script.Program, identName string, uri string) []*Location {
	if ast == nil {
		return nil
	}

	var locations []*Location
	visitNodesForIdentifier(ast, identName, uri, &locations)
	return locations
}

// visitNodesForIdentifier recursively visits all nodes and collects references
func visitNodesForIdentifier(node script.Node, identName string, uri string, locations *[]*Location) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *script.Program:
		for _, stmt := range n.Statements {
			visitNodesForIdentifier(stmt, identName, uri, locations)
		}

	case *script.Identifier:
		if n.Name == identName {
			*locations = append(*locations, &Location{
				URI: uri,
				Range: Range{
					Start: DusoPositionToLSP(n.Pos),
					End:   DusoPositionToLSP(n.Pos),
				},
			})
		}

	case *script.IfStatement:
		visitNodesForIdentifier(n.Condition, identName, uri, locations)
		for _, stmt := range n.Then {
			visitNodesForIdentifier(stmt, identName, uri, locations)
		}
		for _, elseif := range n.Elseifs {
			visitNodesForIdentifier(elseif.Condition, identName, uri, locations)
			for _, stmt := range elseif.Then {
				visitNodesForIdentifier(stmt, identName, uri, locations)
			}
		}
		for _, stmt := range n.Else {
			visitNodesForIdentifier(stmt, identName, uri, locations)
		}

	case *script.WhileStatement:
		visitNodesForIdentifier(n.Condition, identName, uri, locations)
		for _, stmt := range n.Body {
			visitNodesForIdentifier(stmt, identName, uri, locations)
		}

	case *script.ForStatement:
		if n.Var == identName {
			// Don't count loop variable declaration as a reference
		}
		visitNodesForIdentifier(n.Start, identName, uri, locations)
		visitNodesForIdentifier(n.End, identName, uri, locations)
		if n.Step != nil {
			visitNodesForIdentifier(n.Step, identName, uri, locations)
		}
		if n.Iterator != nil {
			visitNodesForIdentifier(n.Iterator, identName, uri, locations)
		}
		for _, stmt := range n.Body {
			visitNodesForIdentifier(stmt, identName, uri, locations)
		}

	case *script.FunctionDef:
		for _, stmt := range n.Body {
			visitNodesForIdentifier(stmt, identName, uri, locations)
		}

	case *script.AssignStatement:
		// Check if this is defining the identifier
		if ident, ok := n.Target.(*script.Identifier); ok {
			if ident.Name != identName {
				visitNodesForIdentifier(n.Target, identName, uri, locations)
			}
		} else {
			visitNodesForIdentifier(n.Target, identName, uri, locations)
		}
		visitNodesForIdentifier(n.Value, identName, uri, locations)

	case *script.ReturnStatement:
		visitNodesForIdentifier(n.Value, identName, uri, locations)

	case *script.CallExpr:
		visitNodesForIdentifier(n.Func, identName, uri, locations)
		for _, arg := range n.Arguments {
			visitNodesForIdentifier(arg, identName, uri, locations)
		}
		for _, arg := range n.NamedArgs {
			visitNodesForIdentifier(arg, identName, uri, locations)
		}

	case *script.BinaryExpr:
		visitNodesForIdentifier(n.Left, identName, uri, locations)
		visitNodesForIdentifier(n.Right, identName, uri, locations)

	case *script.TryStatement:
		for _, stmt := range n.Block {
			visitNodesForIdentifier(stmt, identName, uri, locations)
		}
		for _, stmt := range n.CatchBlock {
			visitNodesForIdentifier(stmt, identName, uri, locations)
		}
	}
}

// isRelevantIdentifier checks if a node is relevant for identifier matching
// This filters out loop variables and parameters
func isRelevantIdentifier(node script.Node, identName string) bool {
	switch n := node.(type) {
	case *script.ForStatement:
		return n.Var != identName
	default:
		return true
	}
}

// ProvideRename provides a rename for an identifier
// Returns all locations that would need to be updated
func ProvideRename(ast *script.Program, identName string, uri string) []*Location {
	// For Phase 1, same as references
	// Phase 2: Add validation to ensure new name doesn't conflict
	return FindReferences(ast, identName, uri)
}

// ValidateName checks if a name is valid for renaming
func ValidateName(name string) error {
	if len(name) == 0 {
		return &LSPError{Code: -32600, Message: "name cannot be empty"}
	}

	// Check first character is letter or underscore
	if !isValidNameStart(rune(name[0])) {
		return &LSPError{Code: -32600, Message: "name must start with letter or underscore"}
	}

	// Check remaining characters
	for _, ch := range name[1:] {
		if !isValidNameChar(ch) {
			return &LSPError{Code: -32600, Message: "name contains invalid characters"}
		}
	}

	return nil
}

func isValidNameStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isValidNameChar(ch rune) bool {
	return isValidNameStart(ch) || (ch >= '0' && ch <= '9')
}

// IsKeyword checks if a name is a reserved keyword
func IsKeyword(name string) bool {
	keywords := map[string]bool{
		"if":       true,
		"else":     true,
		"elseif":   true,
		"while":    true,
		"for":      true,
		"in":       true,
		"function": true,
		"return":   true,
		"break":    true,
		"continue": true,
		"true":     true,
		"false":    true,
		"nil":      true,
		"var":      true,
		"try":      true,
		"catch":    true,
		"require":  true,
	}
	return keywords[name]
}

// LSPError represents an LSP error
type LSPError struct {
	Code    int
	Message string
}

func (e *LSPError) Error() string {
	return e.Message
}
