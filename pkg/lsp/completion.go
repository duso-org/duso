package lsp

import (
	"github.com/duso-org/duso/pkg/script"
)

// CompletionItem represents a completion suggestion
type CompletionItem struct {
	Label       string `json:"label"`
	Kind        int    `json:"kind"`        // 1=Text, 2=Method, 3=Function, 4=Constructor, 5=Field, 6=Variable, 7=Class, 8=Interface, 9=Module, 10=Property, 11=Unit, 12=Value, 13=Enum, 14=Keyword, 15=Snippet, 16=Color, 17=Reference, 18=Folder, 19=EnumMember, 20=Constant, 21=Struct, 22=Event, 23=Operator, 24=TypeParameter
	Detail      string `json:"detail,omitempty"`
	Description string `json:"documentation,omitempty"`
}

// ProvideCompletion returns completion suggestions for a position
func ProvideCompletion(server *Server, doc *Document, pos Position, entry *ASTCacheEntry) []*CompletionItem {
	if entry == nil || entry.AST == nil {
		return nil
	}

	// Convert LSP position to Duso position
	dusoPos := LSPPositionToDuso(pos)

	// Collect variables defined before cursor position
	variables := CollectVariablesBefore(entry.AST, dusoPos)

	// Get built-in functions
	builtins := GetBuiltinFunctions()

	// Combine into completion items
	var items []*CompletionItem

	// Add keywords first (kind 14 = Keyword)
	keywords := []string{
		"if", "then", "else", "elseif", "end",
		"while", "do",
		"for", "in",
		"function",
		"return", "break", "continue",
		"try", "catch",
		"and", "or", "not",
		"var", "raw",
		"true", "false", "nil",
	}
	for _, name := range keywords {
		items = append(items, &CompletionItem{
			Label:  name,
			Kind:   14,
			Detail: "Keyword",
		})
	}

	// Add built-in functions (kind 3 = Function)
	for _, name := range builtins {
		items = append(items, &CompletionItem{
			Label:  name,
			Kind:   3,
			Detail: "Built-in function",
		})
	}

	// Add variables (kind 6 = Variable)
	seen := make(map[string]bool)
	for _, name := range variables {
		if !seen[name] {
			items = append(items, &CompletionItem{
				Label:  name,
				Kind:   6,
				Detail: "Variable",
			})
			seen[name] = true
		}
	}

	return items
}

// CollectVariablesBefore walks the AST and collects all variable names assigned before the cursor position
func CollectVariablesBefore(node script.Node, pos script.Position) []string {
	var variables []string

	switch n := node.(type) {
	case *script.Program:
		// Walk through statements in order until we reach the cursor position
		for _, stmt := range n.Statements {
			if stmt == nil {
				continue
			}
			// Get statement position (if available)
			stmtPos := getNodePosition(stmt)
			if stmtPos != nil && (stmtPos.Line > pos.Line || (stmtPos.Line == pos.Line && stmtPos.Column >= pos.Column)) {
				// We've reached or passed the cursor position
				break
			}
			// Collect variables from this statement
			collectVariablesFromNode(stmt, &variables)
		}

	default:
		collectVariablesFromNode(node, &variables)
	}

	return variables
}

// collectVariablesFromNode recursively collects variable names from a node
func collectVariablesFromNode(node script.Node, variables *[]string) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *script.AssignStatement:
		// Extract variable name from target
		if ident, ok := n.Target.(*script.Identifier); ok {
			*variables = append(*variables, ident.Name)
		}
		// Also recurse into value
		collectVariablesFromNode(n.Value, variables)

	case *script.CompoundAssignStatement:
		if ident, ok := n.Target.(*script.Identifier); ok {
			*variables = append(*variables, ident.Name)
		}

	case *script.ForStatement:
		// Loop iterator is a variable
		if n.Iterator != nil {
			if ident, ok := n.Iterator.(*script.Identifier); ok {
				*variables = append(*variables, ident.Name)
			}
		}
		// Recurse into body
		for _, stmt := range n.Body {
			collectVariablesFromNode(stmt, variables)
		}
		collectVariablesFromNode(n.Start, variables)
		collectVariablesFromNode(n.End, variables)
		collectVariablesFromNode(n.Step, variables)

	case *script.IfStatement:
		for _, stmt := range n.Then {
			collectVariablesFromNode(stmt, variables)
		}
		for _, elseif := range n.Elseifs {
			for _, stmt := range elseif.Then {
				collectVariablesFromNode(stmt, variables)
			}
		}
		for _, stmt := range n.Else {
			collectVariablesFromNode(stmt, variables)
		}

	case *script.WhileStatement:
		for _, stmt := range n.Body {
			collectVariablesFromNode(stmt, variables)
		}

	case *script.FunctionDef:
		// Function parameters are variables
		for _, param := range n.Parameters {
			if param != nil {
				*variables = append(*variables, param.Name)
			}
		}
		// Recurse into body
		for _, stmt := range n.Body {
			collectVariablesFromNode(stmt, variables)
		}

	case *script.TryStatement:
		for _, stmt := range n.Block {
			collectVariablesFromNode(stmt, variables)
		}
		if n.CatchVar != "" {
			*variables = append(*variables, n.CatchVar)
		}
		for _, stmt := range n.CatchBlock {
			collectVariablesFromNode(stmt, variables)
		}
	}
}

// getNodePosition extracts position from a node (if available)
func getNodePosition(node script.Node) *script.Position {
	switch n := node.(type) {
	case *script.AssignStatement:
		return &n.Pos
	case *script.CompoundAssignStatement:
		return &n.Pos
	case *script.PostIncrementStatement:
		return &n.Pos
	case *script.IfStatement:
		return &n.Pos
	case *script.WhileStatement:
		return &n.Pos
	case *script.ForStatement:
		return &n.Pos
	case *script.TryStatement:
		return &n.Pos
	case *script.ReturnStatement:
		return &n.Pos
	case *script.BreakStatement:
		return &n.Pos
	case *script.ContinueStatement:
		return &n.Pos
	case *script.FunctionDef:
		return &n.Pos
	case *script.CallExpr:
		return &n.Pos
	case *script.BinaryExpr:
		return &n.Pos
	}
	return nil
}

// GetBuiltinFunctions returns a list of all built-in function names
func GetBuiltinFunctions() []string {
	return []string{
		// Core I/O
		"print",
		"input",

		// Type checking/conversion
		"type",
		"tostring",
		"tonumber",
		"tobool",

		// String operations
		"upper",
		"lower",
		"substr",
		"trim",
		"split",
		"join",
		"contains",
		"find",
		"replace",

		// Collections
		"len",
		"keys",
		"values",
		"push",
		"pop",
		"shift",
		"unshift",
		"sort",

		// Functional
		"map",
		"filter",
		"reduce",
		"range",

		// Math
		"abs",
		"floor",
		"ceil",
		"round",
		"min",
		"max",
		"sqrt",
		"pow",
		"clamp",
		"random",

		// JSON
		"parse_json",
		"format_json",

		// Date/Time
		"now",
		"format_time",
		"parse_time",

		// Concurrency/Control
		"sleep",
		"parallel",
		"breakpoint",
		"watch",

		// System
		"exit",
		"throw",

		// Utility
		"uuid",
		"template",

		// CLI-specific functions
		"fetch",
		"http_server",
		"run",
		"doc",
		"datastore",
		"load",
		"save",
		"context",
		"env",
		"require",
		"include",
		"spawn",
		"current_dir",
		"list_dir",
		"list_files",
		"file_exists",
		"file_type",
		"make_dir",
		"remove_dir",
		"copy_file",
		"move_file",
		"rename_file",
		"remove_file",
		"append_file",
	}
}
