// resolver.go - Compile-time slot resolution for function parameters
//
// After parsing, resolveProgram walks the AST and annotates Identifier nodes
// that provably refer to a parameter of the enclosing function with a slot
// index. Parameters occupy the first inline storage slots of the function
// environment in declaration order (both call paths guarantee this), so an
// annotated identifier reads e.env.fnScope.vals[slot] directly instead of
// walking the scope chain with string compares.
//
// CORE LANGUAGE COMPONENT: This is part of the minimal core runtime.
//
// The resolver is deliberately conservative: it annotates a use only when the
// dynamic path would provably produce the same result, because local lookup
// wins over self properties and parent scopes for both reads and writes
// (see Environment.Get/Set). Anything ambiguous stays un-annotated and takes
// the existing name-based path. Punt rules:
//
//   - only parameters are slotted; they are bound before the body runs no
//     matter how they were supplied (positional, named, or default)
//   - a parameter shadowed anywhere in the function body — by a var
//     declaration, a for-loop variable, a catch variable, or a nested
//     function definition name — is not slotted at all
//   - identifiers inside named-argument expressions are never slotted:
//     object constructors evaluate them in a temp scope where earlier named
//     args are visible (see callObject)
//   - parameter default expressions are never slotted (evaluated in the
//     closure environment, not the function environment)
//   - nested function bodies get their own scope; outer parameters are not
//     slottable inside them (closure capture stays name-based)
//   - a parameter named "self" is never slotted (Get special-cases the name)
//   - only the first smallScopeSize parameters get slots (inline storage)
package script

// resolveProgram annotates parameter references in all function bodies.
// Called once at the end of Parser.Parse; annotations are read-only afterward.
func resolveProgram(p *Program) {
	r := resolver{}
	r.walkAll(p.Statements)
}

type resolver struct {
	// slottable params of the innermost function being walked: name -> slot+1
	// (0 means dynamic); nil outside any function body
	slots map[string]uint8
}

// containsFunction reports whether any node in the tree defines a function.
// A nested function's closure chains through the enclosing call env, so its
// presence anywhere in a body (even transitively) makes that env non-poolable.
func containsFunction(nodes []Node) bool {
	for _, n := range nodes {
		switch x := n.(type) {
		case *FunctionDef, *FunctionExpr:
			return true
		case *IfStatement:
			if containsFunction([]Node{x.Condition}) || containsFunction(x.Then) || containsFunction(x.Else) {
				return true
			}
			for _, ei := range x.Elseifs {
				if containsFunction([]Node{ei.Condition}) || containsFunction(ei.Then) {
					return true
				}
			}
		case *WhileStatement:
			if containsFunction([]Node{x.Condition}) || containsFunction(x.Body) {
				return true
			}
		case *ForStatement:
			if containsFunction(x.Body) {
				return true
			}
			for _, sub := range []Node{x.Start, x.End, x.Step, x.Iterator} {
				if sub != nil && containsFunction([]Node{sub}) {
					return true
				}
			}
		case *TryStatement:
			if containsFunction(x.Block) || containsFunction(x.CatchBlock) {
				return true
			}
		case *ReturnStatement:
			if x.Value != nil && containsFunction([]Node{x.Value}) {
				return true
			}
		case *AssignStatement:
			if containsFunction([]Node{x.Value}) || containsFunction([]Node{x.Target}) {
				return true
			}
		case *CompoundAssignStatement:
			if containsFunction([]Node{x.Value}) || containsFunction([]Node{x.Target}) {
				return true
			}
		case *PostIncrementStatement:
			if containsFunction([]Node{x.Target}) {
				return true
			}
		case *BinaryExpr:
			if containsFunction([]Node{x.Left, x.Right}) {
				return true
			}
		case *TernaryExpr:
			if containsFunction([]Node{x.Condition, x.TrueExpr, x.FalseExpr}) {
				return true
			}
		case *CallExpr:
			if containsFunction([]Node{x.Func}) || containsFunction(x.Arguments) {
				return true
			}
			for _, v := range x.NamedArgs {
				if containsFunction([]Node{v}) {
					return true
				}
			}
		case *IndexExpr:
			if containsFunction([]Node{x.Object, x.Index}) {
				return true
			}
		case *PropertyAccess:
			if containsFunction([]Node{x.Object}) {
				return true
			}
		case *ArrayLiteral:
			if containsFunction(x.Elements) {
				return true
			}
		case *ObjectLiteral:
			for _, v := range x.StaticPairs {
				if containsFunction([]Node{v}) {
					return true
				}
			}
			for _, cp := range x.ComputedPairs {
				if containsFunction([]Node{cp.KeyExpr, cp.ValueExpr}) {
					return true
				}
			}
		case *TemplateLiteral:
			if containsFunction(x.Parts) {
				return true
			}
		}
	}
	return false
}

// resolveFunction computes the slottable parameters for one function, then
// walks its body with them in scope.
func (r *resolver) resolveFunction(params []*Parameter, body []Node) {
	shadowed := make(map[string]bool)
	collectShadows(body, shadowed)

	slots := make(map[string]uint8, len(params))
	for i, p := range params {
		if i >= smallScopeSize {
			break
		}
		if p.Name == "self" || shadowed[p.Name] {
			continue
		}
		slots[p.Name] = uint8(i + 1)
	}

	prev := r.slots
	r.slots = slots
	r.walkAll(body)
	r.slots = prev
}

// collectShadows records every name a function body could bind in the
// function env or a child env at runtime. Binding constructs only occur at
// statement level, so expressions need no scan; nested function bodies bind
// into their own scopes and are excluded.
func collectShadows(nodes []Node, out map[string]bool) {
	for _, n := range nodes {
		switch s := n.(type) {
		case *IfStatement:
			collectShadows(s.Then, out)
			for _, ei := range s.Elseifs {
				collectShadows(ei.Then, out)
			}
			collectShadows(s.Else, out)
		case *WhileStatement:
			collectShadows(s.Body, out)
		case *ForStatement:
			out[s.Var] = true
			collectShadows(s.Body, out)
		case *TryStatement:
			out[s.CatchVar] = true
			collectShadows(s.Block, out)
			collectShadows(s.CatchBlock, out)
		case *FunctionDef:
			out[s.Name] = true
		case *AssignStatement:
			if s.IsVarDeclaration {
				if id, ok := s.Target.(*Identifier); ok {
					out[id.Name] = true
				}
			}
		}
	}
}

func (r *resolver) walkAll(nodes []Node) {
	for _, n := range nodes {
		r.walk(n)
	}
}

func (r *resolver) walk(n Node) {
	switch x := n.(type) {
	case *Identifier:
		if r.slots != nil {
			x.slot = r.slots[x.Name]
		}
	case *IfStatement:
		r.walk(x.Condition)
		r.walkAll(x.Then)
		for _, ei := range x.Elseifs {
			r.walk(ei.Condition)
			r.walkAll(ei.Then)
		}
		r.walkAll(x.Else)
	case *WhileStatement:
		r.walk(x.Condition)
		r.walkAll(x.Body)
	case *ForStatement:
		// Start/End/Step/Iterator evaluate before the loop env exists, so
		// only the body can capture it
		x.noCapture = !containsFunction(x.Body)
		if x.IsNumeric {
			r.walk(x.Start)
			r.walk(x.End)
			if x.Step != nil {
				r.walk(x.Step)
			}
		} else {
			r.walk(x.Iterator)
		}
		r.walkAll(x.Body)
	case *FunctionDef:
		x.noCapture = !containsFunction(x.Body)
		r.resolveFunction(x.Parameters, x.Body)
	case *FunctionExpr:
		x.noCapture = !containsFunction(x.Body)
		r.resolveFunction(x.Parameters, x.Body)
	case *TryStatement:
		r.walkAll(x.Block)
		r.walkAll(x.CatchBlock)
	case *ReturnStatement:
		if x.Value != nil {
			r.walk(x.Value)
		}
	case *AssignStatement:
		r.walk(x.Target)
		r.walk(x.Value)
	case *CompoundAssignStatement:
		r.walk(x.Target)
		r.walk(x.Value)
	case *PostIncrementStatement:
		r.walk(x.Target)
	case *BinaryExpr:
		r.walk(x.Left)
		r.walk(x.Right)
	case *TernaryExpr:
		r.walk(x.Condition)
		r.walk(x.TrueExpr)
		r.walk(x.FalseExpr)
	case *CallExpr:
		r.walk(x.Func)
		r.walkAll(x.Arguments)
		// NamedArgs deliberately not walked (see punt rules above)
	case *IndexExpr:
		r.walk(x.Object)
		r.walk(x.Index)
	case *PropertyAccess:
		r.walk(x.Object)
	case *ArrayLiteral:
		r.walkAll(x.Elements)
	case *ObjectLiteral:
		for _, v := range x.StaticPairs {
			r.walk(v)
		}
		for _, cp := range x.ComputedPairs {
			r.walk(cp.KeyExpr)
			r.walk(cp.ValueExpr)
		}
	case *TemplateLiteral:
		for _, part := range x.Parts {
			if _, ok := part.(*TextPart); !ok {
				r.walk(part)
			}
		}
	}
}
