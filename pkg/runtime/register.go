package runtime

import "github.com/duso-org/duso/pkg/script"

// globalInterpreter is set by the host (CLI or embedded) for builtins that need it
// This is set in cmd/duso/main.go before any scripts execute
//
// READ-ONLY AFTER INIT: Never set IOConfig, OutputWriter, or call SetFilePath on this.
// Each execution (spawn, run, HTTP handler) gets a fresh evaluator and IOConfig in its RequestContext.
// Builtins read capabilities (ScriptLoader, FileReader, etc.) from this shared template.
var globalInterpreter *script.Interpreter

// ResolvePath is set by the host (CLI) to resolve special path prefixes.
// Handles /EMBED/, /STORE/, /HERE/, /CWD/, bare paths, and absolute paths.
// Set by cli.RegisterFunctions() during initialization.
var ResolvePath func(string) string

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

	// CSV operations
	RegisterBuiltin("parse_csv", builtinParseCSV)
	RegisterBuiltin("format_csv", builtinFormatCSV)

	// Base64 operations
	RegisterBuiltin("encode_base64", builtinEncodeBase64)
	RegisterBuiltin("decode_base64", builtinDecodeBase64)

	// Hash operations
	RegisterBuiltin("hash", builtinHash)

	// Password operations
	RegisterBuiltin("hash_password", builtinHashPassword)
	RegisterBuiltin("verify_password", builtinVerifyPassword)

	// RSA operations
	RegisterBuiltin("sign_rsa", builtinSignRSA)
	RegisterBuiltin("verify_rsa", builtinVerifyRSA)
	RegisterBuiltin("rsa_from_jwk", builtinRSAFromJWK)

	// EC (elliptic curve) operations
	RegisterBuiltin("sign_ec", builtinSignEC)
	RegisterBuiltin("verify_ec", builtinVerifyEC)
	RegisterBuiltin("ec_from_jwk", builtinECFromJWK)

	// Ed25519 operations
	RegisterBuiltin("verify_ed25519", builtinVerifyEd25519)

	// HMAC operations
	RegisterBuiltin("hmac", builtinHMAC)

	// WebSocket operations
	RegisterBuiltin("websocket", builtinWebSocket)
	RegisterBuiltin("send_websocket", builtinSendWebSocket)

	// String operations
	RegisterBuiltin("upper", builtinUpper)
	RegisterBuiltin("lower", builtinLower)
	RegisterBuiltin("repeat", builtinRepeat)
	RegisterBuiltin("substr", builtinSubstr)
	RegisterBuiltin("trim", builtinTrim)
	RegisterBuiltin("pad_left", builtinPadLeft)
	RegisterBuiltin("pad_right", builtinPadRight)

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

	// Math operations - special functions
	RegisterBuiltin("fibonacci", builtinFibonacci)

	// Date/time operations
	RegisterBuiltin("now", builtinNow)
	RegisterBuiltin("timestamp", builtinTimestamp)
	RegisterBuiltin("timer", builtinTimer)
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
	RegisterBuiltin("toregex", builtinToRegex)
	RegisterBuiltin("contains", builtinContains)
	RegisterBuiltin("starts_with", builtinStartsWith)
	RegisterBuiltin("ends_with", builtinEndsWith)
	RegisterBuiltin("find", builtinFind)
	RegisterBuiltin("replace", builtinReplace)

	// Template operations
	RegisterBuiltin("template", builtinTemplate)

	// Markdown operations
	RegisterBuiltin("markdown_html", builtinMarkdownHTML)
	RegisterBuiltin("markdown_ansi", builtinMarkdownANSI)
	RegisterBuiltin("markdown_text", builtinMarkdownText)

	// System operations
	RegisterBuiltin("exit", builtinExit)
	RegisterBuiltin("sleep", builtinSleep)
	RegisterBuiltin("uuid", builtinUUID)

	// HTTP operations
	RegisterBuiltin("fetch", builtinFetch)

	// Data storage operations
	RegisterBuiltin("datastore", builtinDatastore)
	RegisterBuiltin("sql", builtinSQL)

	// Debug/Error operations
	RegisterBuiltin("throw", builtinThrow)
	RegisterBuiltin("assert", builtinAssert)
	RegisterBuiltin("breakpoint", builtinBreakpoint)
	RegisterBuiltin("watch", builtinWatch)

	// Spawning/Execution operations
	RegisterBuiltin("spawn", builtinSpawn)
	RegisterBuiltin("run", builtinRun)
	RegisterBuiltin("kill", builtinKill)
	RegisterBuiltin("context", builtinContext)

	// HTTP server
	RegisterBuiltin("http_server", builtinHTTPServer)

	// Parallel execution
	RegisterBuiltin("parallel", builtinParallel)

	// Image operations
	RegisterBuiltin("scale_image", builtinScaleImage)
	RegisterBuiltin("crop_image", builtinCropImage)
	RegisterBuiltin("convert_image", builtinConvertImage)
	RegisterBuiltin("rotate_image", builtinRotateImage)
	RegisterBuiltin("flip_image_x", builtinFlipImageX)
	RegisterBuiltin("flip_image_y", builtinFlipImageY)
	RegisterBuiltin("grayscale_image", builtinGrayscaleImage)
	RegisterBuiltin("composite_image", builtinCompositeImage)
	RegisterBuiltin("set_image_opacity", builtinSetImageOpacity)
	RegisterBuiltin("adjust_image_opacity", builtinAdjustImageOpacity)

	// Initialize I/O queueing support (for I/O routing in spawned processes)
	initIOQueueing()
}
