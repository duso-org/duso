package script

// SymbolInfo tracks a definition (function or variable)
type SymbolInfo struct {
	Name     string
	Scope    *LintScope
	Position Position
	Used     bool
	Kind     string // "function", "parameter", "variable"
}

// LintScope represents a lexical scope (function, block, etc.)
type LintScope struct {
	Parent     *LintScope
	Symbols    map[string]*SymbolInfo
	IsFunction bool
}

// LintDiagnostic represents a linting issue
type LintDiagnostic struct {
	Message string
	Severity int // 0=error, 1=warning
	Line    int
	Column  int
}

// LintAnalyzer performs static analysis on a Duso AST
type LintAnalyzer struct {
	program      *Program
	filename     string
	diagnostics  []*LintDiagnostic
	currentScope *LintScope
	builtins     map[string]bool // Known builtin functions
}

// NewLintAnalyzer creates a new analyzer
func NewLintAnalyzer(program *Program, filename string) *LintAnalyzer {
	analyzer := &LintAnalyzer{
		program:      program,
		filename:     filename,
		diagnostics: []*LintDiagnostic{},
		currentScope: &LintScope{
			Parent:     nil,
			Symbols:    make(map[string]*SymbolInfo),
			IsFunction: false,
		},
		builtins: make(map[string]bool),
	}

	// Populate builtins from the global registry
	builtinsMap := CopyBuiltins()
	for name := range builtinsMap {
		analyzer.builtins[name] = true
		analyzer.currentScope.Symbols[name] = &SymbolInfo{
			Name: name,
			Kind: "builtin",
			Used: true,
		}
	}

	return analyzer
}

// Analyze performs all linting checks
func (a *LintAnalyzer) Analyze() []*LintDiagnostic {
	// Pre-pass: collect top-level definitions (functions and variables at module level)
	// This allows forward references within the same scope
	for _, node := range a.program.Statements {
		a.collectTopLevelDefinitions(node)
	}

	// Regular pass: check uses and report issues
	a.walkNodes(a.program.Statements)

	// Report unused definitions
	a.reportUnusedDefinitions(a.currentScope)

	return a.diagnostics
}

// collectTopLevelDefinitions collects function and variable definitions at the current scope level
func (a *LintAnalyzer) collectTopLevelDefinitions(node Node) {
	switch n := node.(type) {
	case *FunctionDef:
		// Warn if shadowing a builtin (but allow it)
		if a.builtins[n.Name] {
			a.addDiagnosticAt("'"+n.Name+"' shadows a builtin function", 1, n.Pos)
		}
		a.defineSymbol(n.Name, "function", n.Pos)
	case *AssignStatement:
		if ident, ok := n.Target.(*Identifier); ok {
			if !a.symbolExists(ident.Name) {
				// Warn if shadowing a builtin (but allow it)
				if a.builtins[ident.Name] {
					a.addDiagnosticAt("'"+ident.Name+"' shadows a builtin function", 1, ident.Pos)
				}
				a.defineSymbol(ident.Name, "variable", ident.Pos)
			} else {
				// Warn if shadowing a builtin (but allow it)
				if a.builtins[ident.Name] {
					a.addDiagnosticAt("'"+ident.Name+"' shadows a builtin function", 1, ident.Pos)
				}
			}
		}
	}
}

// walkNodes walks a list of nodes and processes them, checking for unreachable code
func (a *LintAnalyzer) walkNodes(nodes []Node) {
	for i, node := range nodes {
		a.walkNode(node)

		// Check if this statement always exits (return/break/continue)
		// If so, any following statements are unreachable
		if a.alwaysExits(node) && i < len(nodes)-1 {
			// Mark remaining statements as unreachable
			for j := i + 1; j < len(nodes); j++ {
				if pos := a.getNodePos(nodes[j]); pos.Line > 0 {
					a.addDiagnosticAt("unreachable code", 1, pos)
				}
			}
			break
		}
	}
}

// walkNode walks a single node
func (a *LintAnalyzer) walkNode(node Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *FunctionDef:
		a.handleFunctionDef(n)
	case *AssignStatement:
		a.handleAssignStatement(n)
	case *Identifier:
		a.handleIdentifier(n)
	case *ForStatement:
		a.handleForStatement(n)
	case *IfStatement:
		a.handleIfStatement(n)
	case *WhileStatement:
		a.handleWhileStatement(n)
	case *TryStatement:
		a.handleTryStatement(n)
	case *CallExpr:
		a.walkNode(n.Func)
		for _, arg := range n.Arguments {
			a.walkNode(arg)
		}
		for _, arg := range n.NamedArgs {
			a.walkNode(arg)
		}
	case *BinaryExpr:
		a.walkNode(n.Left)
		a.walkNode(n.Right)
	case *UnaryExpr:
		a.walkNode(n.Operand)
	case *TernaryExpr:
		a.walkNode(n.Condition)
		a.walkNode(n.TrueExpr)
		a.walkNode(n.FalseExpr)
	case *IndexExpr:
		a.walkNode(n.Object)
		a.walkNode(n.Index)
	case *PropertyAccess:
		a.walkNode(n.Object)
	case *ArrayLiteral:
		for _, elem := range n.Elements {
			a.walkNode(elem)
		}
	case *ObjectLiteral:
		for _, val := range n.Pairs {
			a.walkNode(val)
		}
	case *TemplateLiteral:
		for _, part := range n.Parts {
			a.walkNode(part)
		}
	case *FunctionExpr:
		a.handleFunctionExpr(n)
	case *ReturnStatement:
		if n.Value != nil {
			a.walkNode(n.Value)
		}
	case *CompoundAssignStatement:
		a.handleCompoundAssignStatement(n)
	case *PostIncrementStatement:
		a.handlePostIncrementStatement(n)
	}
}

// handleFunctionDef handles function definition
func (a *LintAnalyzer) handleFunctionDef(n *FunctionDef) {
	// Define the function in current scope
	a.defineSymbol(n.Name, "function", n.Pos)

	// Create new scope for function body
	oldScope := a.currentScope
	a.currentScope = &LintScope{
		Parent:     oldScope,
		Symbols:    make(map[string]*SymbolInfo),
		IsFunction: true,
	}

	// Add parameters to function scope
	for _, param := range n.Parameters {
		a.defineSymbol(param.Name, "parameter", n.Pos)
		if param.Default != nil {
			a.walkNode(param.Default)
		}
	}

	// Walk function body
	a.walkNodes(n.Body)

	// Report unused definitions in function scope
	a.reportUnusedDefinitions(a.currentScope)

	// Restore scope
	a.currentScope = oldScope
}

// handleFunctionExpr handles anonymous function expression
func (a *LintAnalyzer) handleFunctionExpr(n *FunctionExpr) {
	// Create new scope for function body
	oldScope := a.currentScope
	a.currentScope = &LintScope{
		Parent:     oldScope,
		Symbols:    make(map[string]*SymbolInfo),
		IsFunction: true,
	}

	// Add parameters to function scope
	for _, param := range n.Parameters {
		a.defineSymbol(param.Name, "parameter", Position{Line: 0, Column: 0})
		if param.Default != nil {
			a.walkNode(param.Default)
		}
	}

	// Walk function body
	a.walkNodes(n.Body)

	// Report unused definitions in function scope
	a.reportUnusedDefinitions(a.currentScope)

	// Restore scope
	a.currentScope = oldScope
}

// handleAssignStatement handles variable assignment
func (a *LintAnalyzer) handleAssignStatement(n *AssignStatement) {
	// If this is a declaration or simple assignment to identifier, define the variable
	if ident, ok := n.Target.(*Identifier); ok {
		if n.IsVarDeclaration || !a.symbolExists(ident.Name) {
			a.defineSymbol(ident.Name, "variable", ident.Pos)
		}
		// Note: assigning to a variable doesn't count as "using" it - only reading does
	} else {
		// Assignment to array element or property - walk it
		a.walkNode(n.Target)
	}

	// Walk the value
	a.walkNode(n.Value)
}

// handleCompoundAssignStatement handles +=, -=, etc.
func (a *LintAnalyzer) handleCompoundAssignStatement(n *CompoundAssignStatement) {
	if ident, ok := n.Target.(*Identifier); ok {
		// Note: compound assignment both reads and writes to the variable, so mark as used
		a.useSymbol(ident.Name)
	} else {
		a.walkNode(n.Target)
	}
	a.walkNode(n.Value)
}

// handlePostIncrementStatement handles ++ and --
func (a *LintAnalyzer) handlePostIncrementStatement(n *PostIncrementStatement) {
	if ident, ok := n.Target.(*Identifier); ok {
		// Note: ++ and -- both read and write to the variable, so mark as used
		a.useSymbol(ident.Name)
	} else {
		a.walkNode(n.Target)
	}
}

// handleIdentifier handles identifier reference
func (a *LintAnalyzer) handleIdentifier(n *Identifier) {
	a.useSymbolAt(n.Name, n.Pos)
}

// handleForStatement handles for loops
func (a *LintAnalyzer) handleForStatement(n *ForStatement) {
	if n.IsNumeric {
		// Numeric for: define loop variable
		a.walkNode(n.Start)
		a.walkNode(n.End)
		if n.Step != nil {
			a.walkNode(n.Step)
		}

		// Create new scope for loop
		oldScope := a.currentScope
		a.currentScope = &LintScope{
			Parent:     oldScope,
			Symbols:    make(map[string]*SymbolInfo),
			IsFunction: false,
		}

		// Define loop variable in loop scope
		a.defineSymbol(n.Var, "variable", n.Pos)

		// Walk body
		a.walkNodes(n.Body)

		// Restore scope
		a.currentScope = oldScope
	} else {
		// Iterator for: for item in array
		a.walkNode(n.Iterator)

		// Create new scope for loop
		oldScope := a.currentScope
		a.currentScope = &LintScope{
			Parent:     oldScope,
			Symbols:    make(map[string]*SymbolInfo),
			IsFunction: false,
		}

		// Define loop variable
		a.defineSymbol(n.Var, "variable", n.Pos)

		// Walk body
		a.walkNodes(n.Body)

		// Restore scope
		a.currentScope = oldScope
	}
}

// handleIfStatement handles if statements
func (a *LintAnalyzer) handleIfStatement(n *IfStatement) {
	a.walkNode(n.Condition)
	a.walkNodes(n.Then)
	for _, elseif := range n.Elseifs {
		a.walkNode(elseif.Condition)
		a.walkNodes(elseif.Then)
	}
	a.walkNodes(n.Else)
}

// handleWhileStatement handles while loops
func (a *LintAnalyzer) handleWhileStatement(n *WhileStatement) {
	a.walkNode(n.Condition)
	a.walkNodes(n.Body)
}

// handleTryStatement handles try/catch
func (a *LintAnalyzer) handleTryStatement(n *TryStatement) {
	a.walkNodes(n.Block)

	// Create scope for catch block with catch variable
	oldScope := a.currentScope
	a.currentScope = &LintScope{
		Parent:     oldScope,
		Symbols:    make(map[string]*SymbolInfo),
		IsFunction: false,
	}

	a.defineSymbol(n.CatchVar, "variable", n.Pos)
	a.walkNodes(n.CatchBlock)

	a.currentScope = oldScope
}

// defineSymbol defines a symbol in the current scope
func (a *LintAnalyzer) defineSymbol(name string, kind string, pos Position) {
	a.currentScope.Symbols[name] = &SymbolInfo{
		Name:     name,
		Scope:    a.currentScope,
		Position: pos,
		Used:     false,
		Kind:     kind,
	}
}

// useSymbol marks a symbol as used
func (a *LintAnalyzer) useSymbol(name string) {
	a.useSymbolAt(name, Position{Line: 0, Column: 0})
}

// useSymbolAt marks a symbol as used, with position info
func (a *LintAnalyzer) useSymbolAt(name string, pos Position) {
	sym := a.findSymbol(name)
	if sym != nil {
		sym.Used = true
	} else {
		// Undefined variable (warning, not error - may be from parent scope or object property)
		a.addDiagnosticAt("undefined variable: "+name, 1, pos)
	}
}

// symbolExists checks if a symbol exists in the current scope or parent scopes
func (a *LintAnalyzer) symbolExists(name string) bool {
	return a.findSymbol(name) != nil
}

// findSymbol finds a symbol in the current scope or parent scopes
func (a *LintAnalyzer) findSymbol(name string) *SymbolInfo {
	scope := a.currentScope
	for scope != nil {
		if sym, ok := scope.Symbols[name]; ok {
			return sym
		}
		scope = scope.Parent
	}
	return nil
}

// reportUnusedDefinitions reports unused definitions in a scope
func (a *LintAnalyzer) reportUnusedDefinitions(scope *LintScope) {
	for _, sym := range scope.Symbols {
		if !sym.Used && sym.Kind != "builtin" && sym.Kind != "parameter" {
			a.addDiagnosticAt("unused "+sym.Kind+": "+sym.Name, 1, sym.Position)
		}
	}
}

// addDiagnostic adds a diagnostic at default position
func (a *LintAnalyzer) addDiagnostic(message string, severity int) {
	diag := &LintDiagnostic{
		Message:  message,
		Severity: severity,
		Line:     1,
		Column:   0,
	}
	a.diagnostics = append(a.diagnostics, diag)
}

// addDiagnosticAt adds a diagnostic at a specific position
func (a *LintAnalyzer) addDiagnosticAt(message string, severity int, pos Position) {
	diag := &LintDiagnostic{
		Message:  message,
		Severity: severity,
		Line:     pos.Line,
		Column:   pos.Column,
	}
	a.diagnostics = append(a.diagnostics, diag)
}

// alwaysExits checks if a statement always exits (return/break/continue)
// or if it's an if/else that covers all paths with exits
func (a *LintAnalyzer) alwaysExits(node Node) bool {
	switch n := node.(type) {
	case *ReturnStatement:
		return true
	case *BreakStatement:
		return true
	case *ContinueStatement:
		return true
	case *IfStatement:
		// If has no else, it doesn't always exit
		if len(n.Else) == 0 {
			return false
		}
		// Then branch must exit
		if !a.blockAlwaysExits(n.Then) {
			return false
		}
		// All elseifs must have then blocks that exit
		for _, elseif := range n.Elseifs {
			if !a.blockAlwaysExits(elseif.Then) {
				return false
			}
		}
		// Else branch must exit
		return a.blockAlwaysExits(n.Else)
	}
	return false
}

// blockAlwaysExits checks if a statement block always exits
func (a *LintAnalyzer) blockAlwaysExits(stmts []Node) bool {
	if len(stmts) == 0 {
		return false
	}
	// Check if last statement always exits
	return a.alwaysExits(stmts[len(stmts)-1])
}

// getNodePos returns the position of a node, or zero if not available
func (a *LintAnalyzer) getNodePos(node Node) Position {
	switch n := node.(type) {
	case *IfStatement:
		return n.Pos
	case *WhileStatement:
		return n.Pos
	case *ForStatement:
		return n.Pos
	case *FunctionDef:
		return n.Pos
	case *TryStatement:
		return n.Pos
	case *ReturnStatement:
		return n.Pos
	case *BreakStatement:
		return n.Pos
	case *ContinueStatement:
		return n.Pos
	case *AssignStatement:
		return n.Pos
	case *CompoundAssignStatement:
		return n.Pos
	case *PostIncrementStatement:
		return n.Pos
	case *BinaryExpr:
		return n.Pos
	case *UnaryExpr:
		return n.Pos
	case *TernaryExpr:
		return n.Pos
	case *CallExpr:
		return n.Pos
	case *IndexExpr:
		return n.Pos
	case *PropertyAccess:
		return n.Pos
	case *Identifier:
		return n.Pos
	case *TemplateLiteral:
		return n.Pos
	}
	return Position{Line: 0, Column: 0}
}
