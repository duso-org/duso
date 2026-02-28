package runtime

import "github.com/duso-org/duso/pkg/script"

// globalInterpreter is set by the host (CLI or embedded) for builtins that need it
// This is set in cmd/duso/main.go before any scripts execute
var globalInterpreter *script.Interpreter

// SetInterpreter sets the global interpreter instance for use by builtins
func SetInterpreter(interp *script.Interpreter) {
	globalInterpreter = interp
}

// RegisterBuiltins registers all builtin functions in the global script registry.
// This is called once at startup before any scripts are executed.
func RegisterBuiltins() {
	// Console I/O
	RegisterBuiltin("print", builtinPrint)
	RegisterBuiltin("input", builtinInput)

	// Environment
	RegisterBuiltin("env", builtinEnv)

	// Data operations
	RegisterBuiltin("deep_copy", builtinDeepCopy)

	// JSON operations
	RegisterBuiltin("parse_json", builtinParseJSON)
	RegisterBuiltin("format_json", builtinFormatJSON)

	// String operations
	RegisterBuiltin("upper", builtinUpper)
	RegisterBuiltin("lower", builtinLower)
	RegisterBuiltin("repeat", builtinRepeat)
	RegisterBuiltin("substr", builtinSubstr)
	RegisterBuiltin("trim", builtinTrim)

	// Math operations - basic
	RegisterBuiltin("floor", builtinFloor)
	RegisterBuiltin("ceil", builtinCeil)
	RegisterBuiltin("round", builtinRound)
	RegisterBuiltin("abs", builtinAbs)
	RegisterBuiltin("min", builtinMin)
	RegisterBuiltin("max", builtinMax)
	RegisterBuiltin("sqrt", builtinSqrt)
	RegisterBuiltin("pow", builtinPow)
	RegisterBuiltin("clamp", builtinClamp)

	// Math operations - trigonometric
	RegisterBuiltin("sin", builtinSin)
	RegisterBuiltin("cos", builtinCos)
	RegisterBuiltin("tan", builtinTan)
	RegisterBuiltin("asin", builtinAsin)
	RegisterBuiltin("acos", builtinAcos)
	RegisterBuiltin("atan", builtinAtan)
	RegisterBuiltin("atan2", builtinAtan2)

	// Math operations - exponential and logarithmic
	RegisterBuiltin("exp", builtinExp)
	RegisterBuiltin("log", builtinLog)
	RegisterBuiltin("ln", builtinLn)
	RegisterBuiltin("pi", builtinPi)
	RegisterBuiltin("random", builtinRandom)

	// Date/time operations
	RegisterBuiltin("now", builtinNow)
	RegisterBuiltin("format_time", builtinFormatTime)
	RegisterBuiltin("parse_time", builtinParseTime)

	// Type operations
	RegisterBuiltin("len", builtinLen)
	RegisterBuiltin("type", builtinType)
	RegisterBuiltin("tonumber", builtinToNumber)
	RegisterBuiltin("tostring", builtinToString)
	RegisterBuiltin("tobool", builtinToBool)

	// Code operations
	RegisterBuiltin("parse", builtinParse)

	// Array/Object operations
	RegisterBuiltin("keys", builtinKeys)
	RegisterBuiltin("values", builtinValues)
	RegisterBuiltin("push", builtinPush)
	RegisterBuiltin("pop", builtinPop)
	RegisterBuiltin("shift", builtinShift)
	RegisterBuiltin("unshift", builtinUnshift)
	RegisterBuiltin("split", builtinSplit)
	RegisterBuiltin("join", builtinJoin)
	RegisterBuiltin("range", builtinRange)

	// Functional operations
	RegisterBuiltin("map", builtinMap)
	RegisterBuiltin("filter", builtinFilter)
	RegisterBuiltin("reduce", builtinReduce)
	RegisterBuiltin("sort", builtinSort)

	// Regex operations
	RegisterBuiltin("contains", builtinContains)
	RegisterBuiltin("starts_with", builtinStartsWith)
	RegisterBuiltin("ends_with", builtinEndsWith)
	RegisterBuiltin("find", builtinFind)
	RegisterBuiltin("replace", builtinReplace)

	// Template operations
	RegisterBuiltin("template", builtinTemplate)

	// System operations
	RegisterBuiltin("exit", builtinExit)
	RegisterBuiltin("sleep", builtinSleep)
	RegisterBuiltin("uuid", builtinUUID)

	// HTTP operations
	RegisterBuiltin("fetch", builtinFetch)

	// Data storage operations
	RegisterBuiltin("datastore", builtinDatastore)

	// Debug/Error operations
	RegisterBuiltin("throw", builtinThrow)
	RegisterBuiltin("breakpoint", builtinBreakpoint)
	RegisterBuiltin("watch", builtinWatch)

	// Spawning/Execution operations
	RegisterBuiltin("spawn", builtinSpawn)
	RegisterBuiltin("run", builtinRun)
	RegisterBuiltin("context", builtinContext)

	// HTTP server
	RegisterBuiltin("http_server", builtinHTTPServer)

	// Parallel execution
	RegisterBuiltin("parallel", builtinParallel)
}
