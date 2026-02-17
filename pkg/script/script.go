package script

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// DebugHandler is a callback function that handles debug events (breakpoints, errors).
// It receives the debug event and is responsible for:
// - Displaying the event to the user (via chosen I/O mechanism)
// - Opening a debug session (REPL, HTTP interface, etc.)
// - Sending a resume signal when the user is done debugging
type DebugHandler func(*DebugEvent)

// ParseCacheEntry holds a cached parsed AST with its modification time
type ParseCacheEntry struct {
	ast   *Program
	mtime int64 // File modification time at parse time
}

// Interpreter is the public API for executing Duso scripts.
//
// CORE INTERPRETER - This is suitable for both embedded Go applications and CLI usage.
// It uses only the core language runtime with no external dependencies.
//
// To extend with CLI features (file I/O, module loading), see pkg/cli/register.go
type Interpreter struct {
	evaluator       *Evaluator
	scriptDir       string                      // Directory of the main script (for relative path resolution in run/spawn)
	moduleCache     map[string]Value            // Cache for require() results, keyed by absolute path
	moduleCacheMu   sync.RWMutex                // Protects moduleCache
	parseCache      map[string]*ParseCacheEntry // Cache for parsed ASTs, keyed by absolute path
	parseMutex      sync.RWMutex                // Protects parseCache
	debugEventChan  chan *DebugEvent            // Channel for debug events from child scripts
	debugHandler    DebugHandler                // Handler for debug events (set by CLI or other integrations)
	debugHandlerMu  sync.Mutex                  // Protects debugHandler
	debugSessionMu  sync.Mutex                  // Serializes debug REPL access (only one session at a time)

	// Host-provided capabilities (for builtins that need host services)
	ScriptLoader func(path string) ([]byte, error) // Loads scripts for spawn/run (required for those builtins)
	FileReader   func(path string) ([]byte, error) // Reads files for load/readfile (required for those builtins)
	FileWriter   func(path, content string) error  // Writes files for save/writefile (required for those builtins)
	FileStatter  func(path string) int64           // Gets file modification time for caching (used by http_server)
	OutputWriter func(msg string) error            // Outputs messages for print/error/debug (required for those builtins)
	InputReader  func(prompt string) (string, error) // Reads input from user (required for input() builtin)
	EnvReader    func(varname string) string       // Reads environment variables (used by env() builtin)
}

// NewInterpreter creates a new interpreter instance.
//
// This creates a minimal interpreter with only the core Duso language features.
// Use this in embedded Go applications, then optionally register custom functions
// with RegisterFunction() or CLI features with pkg/cli.RegisterFunctions().
func NewInterpreter() *Interpreter {
	return &Interpreter{
		moduleCache:    make(map[string]Value),
		parseCache:     make(map[string]*ParseCacheEntry),
		debugEventChan: make(chan *DebugEvent, 1000), // Buffered to allow many debug events to queue
	}
}

// SetScriptDir sets the directory of the main script for relative path resolution.
// Used by run() and spawn() to resolve relative script paths when loading from embedded files.
func (i *Interpreter) SetScriptDir(dir string) {
	i.scriptDir = dir
}

// GetScriptDir returns the directory of the main script.
func (i *Interpreter) GetScriptDir() string {
	return i.scriptDir
}

// RegisterFunction registers a custom Go function callable from Duso scripts.
//
// This is how embedded applications extend Duso with domain-specific functionality.
// For CLI-specific functions (load, save, include), see pkg/cli.
func (i *Interpreter) RegisterFunction(name string, fn GoFunction) error {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}
	i.evaluator.RegisterFunction(name, fn)
	return nil
}

// RegisterObject registers an object with methods (e.g., "agents" with methods like "classify")
func (i *Interpreter) RegisterObject(name string, methods map[string]GoFunction) error {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}
	i.evaluator.RegisterObject(name, methods)

	// Create a wrapper object that allows method calls
	objMethods := make(map[string]Value)
	for methodName, fn := range methods {
		objMethods[methodName] = NewGoFunction(fn)
	}

	// Register as an object in the environment
	objVal := NewObject(make(map[string]Value))
	i.evaluator.env.Define(name, objVal)

	// Actually, we need to handle object method calls differently
	// For now, register each method as "object.method"
	for methodName, fn := range methods {
		fullName := name + "." + methodName
		i.evaluator.RegisterFunction(fullName, fn)
	}

	return nil
}

// Execute executes script source code
func (i *Interpreter) Execute(source string) (string, error) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}

	// Tokenize
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()

	// Parse
	filePath := i.GetFilePath()
	parser := NewParserWithFile(tokens, filePath)
	program, err := parser.Parse()
	if err != nil {
		return "", err
	}

	// Set up invocation frame and request context for the main script
	// This allows spawn/run/load to resolve relative paths correctly
	// Convert filePath to absolute to ensure relative path resolution works
	absFilePath := filePath
	if filePath != "" && !filepath.IsAbs(filePath) {
		if abs, err := filepath.Abs(filePath); err == nil {
			absFilePath = abs
		}
	}

	gid := GetGoroutineID()
	frame := &InvocationFrame{
		Filename: absFilePath,
		Line:     1,
		Col:      1,
		Reason:   "main",
	}
	ctx := &RequestContext{Frame: frame}
	SetRequestContextWithData(gid, ctx, nil)
	defer ClearRequestContext(gid)

	// Evaluate
	_, err = i.evaluator.Eval(program)
	if err != nil {
		return "", err
	}

	return "", nil
}

// ExecuteNode executes a single AST node.
// Used by debugger for statement-by-statement execution.
// Maintains evaluator state between calls.
func (i *Interpreter) ExecuteNode(node Node) error {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}
	_, err := i.evaluator.Eval(node)
	return err
}

// EvalInContext evaluates code in the current evaluator context.
// Used by the debug REPL to maintain variable scope and evaluator state.
// Unlike Execute(), this preserves all evaluator state without reinitializing.
func (i *Interpreter) EvalInContext(source string) (string, error) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}

	// Tokenize
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()

	// Parse
	parser := NewParserWithFile(tokens, i.GetFilePath())
	program, err := parser.Parse()
	if err != nil {
		return "", err
	}

	// Evaluate statements individually in current context
	for _, stmt := range program.Statements {
		_, err := i.evaluator.Eval(stmt)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

// EvalInEnvironment evaluates code in a specific environment context.
// This is used by the debug REPL to evaluate expressions in the scope where the breakpoint occurred.
func (i *Interpreter) EvalInEnvironment(source string, env *Environment) (string, error) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}

	// Tokenize
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()

	// Parse
	parser := NewParserWithFile(tokens, i.GetFilePath())
	program, err := parser.Parse()
	if err != nil {
		return "", err
	}

	// Save current environment and switch to the breakpoint's environment
	prevEnv := i.evaluator.env
	i.evaluator.env = env

	// Evaluate statements individually in the provided environment
	for _, stmt := range program.Statements {
		_, err := i.evaluator.Eval(stmt)
		if err != nil {
			i.evaluator.env = prevEnv
			return "", err
		}
	}

	i.evaluator.env = prevEnv
	return "", nil
}

// ExecuteFile executes a script file
func (i *Interpreter) ExecuteFile(path string) (string, error) {
	// Note: We don't have file I/O here - that's handled by the caller
	// This is a placeholder for future implementation
	return "", nil
}

// SetFilePath sets the current file path for error reporting
func (i *Interpreter) SetFilePath(path string) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}
	i.evaluator.ctx.FilePath = path
}

// GetFilePath returns the current file path for error reporting
func (i *Interpreter) GetFilePath() string {
	if i.evaluator == nil {
		return "<stdin>"
	}
	return i.evaluator.ctx.FilePath
}

// GetCallStack returns the current call stack for debugging
func (i *Interpreter) GetCallStack() []CallFrame {
	if i.evaluator == nil || i.evaluator.ctx == nil {
		return nil
	}
	return i.evaluator.ctx.CallStack
}

// GetEvaluator returns the internal evaluator instance (for advanced use).
// This is primarily used by CLI functions that need access to registered Go functions.
func (i *Interpreter) GetEvaluator() *Evaluator {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}
	return i.evaluator
}

// GetDebugEventChan returns the channel for receiving debug events from child scripts
func (i *Interpreter) GetDebugEventChan() chan *DebugEvent {
	return i.debugEventChan
}

// QueueDebugEvent sends a debug event to the main process (non-blocking due to buffered channel)
func (i *Interpreter) QueueDebugEvent(event *DebugEvent) {
	select {
	case i.debugEventChan <- event:
	default:
		// Buffer full, skip (shouldn't happen with size 1, but fail-safe)
	}
}

// RegisterDebugHandler registers a handler function to be called when debug events occur.
// The handler is responsible for displaying the event to the user and managing the debug session.
// This allows the runtime to be I/O-agnostic while delegating user interaction to the handler.
//
// Example (console debugging):
//
//	interp.RegisterDebugHandler(func(event *DebugEvent) {
//	    handleConsoleDebugEvent(interp, event)
//	})
//
// The handler will be called from the debug event listener goroutine.
func (i *Interpreter) RegisterDebugHandler(handler DebugHandler) {
	i.debugHandlerMu.Lock()
	defer i.debugHandlerMu.Unlock()
	i.debugHandler = handler
}

// GetDebugHandler retrieves the currently registered debug handler.
func (i *Interpreter) GetDebugHandler() DebugHandler {
	i.debugHandlerMu.Lock()
	defer i.debugHandlerMu.Unlock()
	return i.debugHandler
}

// GetDebugSessionMutex returns the mutex that serializes debug REPL sessions.
// Only one debug REPL should be active at a time to prevent multiple readers on stdin.
func (i *Interpreter) GetDebugSessionMutex() *sync.Mutex {
	return &i.debugSessionMu
}

// ExecuteModule executes script source in an isolated module scope and returns the result value.
// This is used by require() to load modules in isolation. The module's variables
// don't leak into the caller's scope. The last expression value (or explicit return) is the export.
func (i *Interpreter) ExecuteModule(source string) (Value, error) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}

	// Tokenize
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()

	// Parse
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		return NewNil(), err
	}

	// Evaluate in isolated scope
	return i.evaluator.EvalModule(program)
}

// ExecuteModuleProgram executes a pre-parsed program in an isolated module scope.
// This is used by require() when the AST is already cached.
// The module's variables don't leak into the caller's scope.
// The last expression value (or explicit return) is the export.
func (i *Interpreter) ExecuteModuleProgram(program *Program) (Value, error) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}
	return i.evaluator.EvalModule(program)
}

// EvalProgram evaluates a pre-parsed program in the current scope.
// This is used by include() when the AST is already cached.
// Unlike ExecuteModuleProgram, this executes in the current environment
// so variables and functions are available after execution.
func (i *Interpreter) EvalProgram(program *Program) (Value, error) {
	if i.evaluator == nil {
		i.evaluator = NewEvaluator()
	}
	return i.evaluator.Eval(program)
}

// ParseScriptFile reads and parses a script file with AST caching and mtime checking.
// This is the centralized script loader used by require(), include(), and main script execution.
//
// The cache is validated using file modification time:
// - If the file hasn't changed since caching, the cached AST is returned
// - If the file is newer, it's re-parsed and the cache is updated
// - For /EMBED/ files, the cached AST is always returned (embedded files don't change)
//
// This function requires a FileReadFunc to be provided for reading files.
// It's typically called from pkg/cli with an appropriate file reader.
func (i *Interpreter) ParseScriptFile(path string, readFile func(string) ([]byte, error), getMtime func(string) int64) (*Program, error) {
	// Check cache with mtime validation
	i.parseMutex.RLock()
	cached, ok := i.parseCache[path]
	i.parseMutex.RUnlock()

	if ok {
		// For embedded files, always use cache
		if strings.HasPrefix(path, "/EMBED/") {
			return cached.ast, nil
		}
		// For regular files, validate mtime
		currentMtime := getMtime(path)
		if currentMtime > 0 && currentMtime == cached.mtime {
			return cached.ast, nil // Cache is valid
		}
	}

	// Not in cache or cache is invalid - read and parse
	source, err := readFile(path)
	if err != nil {
		return nil, err
	}

	// Tokenize
	lexer := NewLexer(string(source))
	tokens := lexer.Tokenize()

	// Parse
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	// Store in cache with mtime
	entry := &ParseCacheEntry{
		ast:   program,
		mtime: getMtime(path),
	}
	i.parseMutex.Lock()
	i.parseCache[path] = entry
	i.parseMutex.Unlock()

	return program, nil
}

// ParseScript parses a script file with AST caching, using the interpreter's ScriptLoader.
// This is used by spawn(), run(), and HTTP handlers to avoid re-parsing the same script.
//
// The cache is validated using file modification time:
// - If the file hasn't changed, the cached AST is returned
// - If the file is newer, it's re-parsed and the cache is updated
// - For /EMBED/ files, the cached AST is always returned
//
// This requires ScriptLoader and FileStatter to be set on the interpreter.
func (i *Interpreter) ParseScript(path string) (*Program, error) {
	// Check cache with mtime validation
	i.parseMutex.RLock()
	cached, ok := i.parseCache[path]
	i.parseMutex.RUnlock()

	if ok {
		// For embedded files, always use cache
		if strings.HasPrefix(path, "/EMBED/") {
			return cached.ast, nil
		}
		// For regular files, validate mtime
		if i.FileStatter != nil {
			currentMtime := i.FileStatter(path)
			if currentMtime > 0 && currentMtime == cached.mtime {
				return cached.ast, nil // Cache is valid
			}
		}
	}

	// Not in cache or cache is invalid - read and parse
	if i.ScriptLoader == nil {
		return nil, fmt.Errorf("ParseScript: ScriptLoader not configured")
	}

	source, err := i.ScriptLoader(path)
	if err != nil {
		return nil, err
	}

	// Tokenize and parse
	lexer := NewLexer(string(source))
	tokens := lexer.Tokenize()
	parser := NewParserWithFile(tokens, path)
	program, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	// Store in cache with mtime (initialize map if needed)
	mtime := int64(0)
	if i.FileStatter != nil {
		mtime = i.FileStatter(path)
	}
	entry := &ParseCacheEntry{
		ast:   program,
		mtime: mtime,
	}
	i.parseMutex.Lock()
	if i.parseCache == nil {
		i.parseCache = make(map[string]*ParseCacheEntry)
	}
	i.parseCache[path] = entry
	i.parseMutex.Unlock()

	return program, nil
}

// GetModuleCache retrieves a cached module value by absolute path.
// Used by require() to implement module caching.
func (i *Interpreter) GetModuleCache(path string) (Value, bool) {
	i.moduleCacheMu.RLock()
	defer i.moduleCacheMu.RUnlock()
	val, ok := i.moduleCache[path]
	return val, ok
}

// SetModuleCache stores a module value in the cache by absolute path.
// Used by require() to cache module results so they're only loaded once.
func (i *Interpreter) SetModuleCache(path string, value Value) {
	i.moduleCacheMu.Lock()
	defer i.moduleCacheMu.Unlock()
	i.moduleCache[path] = value
}

// Reset resets the environment
func (i *Interpreter) Reset() {
	i.evaluator = nil
}

// ResolveScriptPath resolves a script path relative to a calling script's directory.
// If the path is absolute or special (/EMBED/, /STORE/), returns it unchanged.
// If the path is relative, resolves it relative to the calling script's directory.
// Example: ResolveScriptPath("./worker.du", "/path/to/bees/bees.du")
//          returns "/path/to/bees/worker.du"
func ResolveScriptPath(requestedPath, callingScriptFilename string) string {
	// If path is absolute, special prefix, or empty, return as-is
	if filepath.IsAbs(requestedPath) ||
		strings.HasPrefix(requestedPath, "/EMBED/") ||
		strings.HasPrefix(requestedPath, "/STORE/") ||
		requestedPath == "" {
		return requestedPath
	}

	// Get the directory of the calling script
	scriptDir := filepath.Dir(callingScriptFilename)
	if scriptDir == "" || scriptDir == "." {
		// If no directory info, return path as-is (will be relative to CWD)
		return requestedPath
	}

	// Resolve relative path from script's directory
	return filepath.Join(scriptDir, requestedPath)
}
