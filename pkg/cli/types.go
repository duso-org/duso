package cli

import (
	"github.com/duso-org/duso/pkg/script"
)

// Type aliases to avoid script. prefix in CLI builtins
type (
	Evaluator = script.Evaluator
	Value = script.Value
	ValueRef = script.ValueRef
	GoFunction = script.GoFunction
)

// Registry functions
var (
	RegisterBuiltin = script.RegisterBuiltin
)
