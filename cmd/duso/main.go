package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/duso-org/duso/pkg/cli"
	"github.com/duso-org/duso/pkg/script"
)

// Embed stdlib, docs, contrib, and examples directories into the binary.
// Before building, run: go generate ./cmd/duso
// This copies stdlib/, docs/, contrib/, and examples/ from repo root into this directory for embedding.
// See embed/main.go for details.
//
//go:generate go run ./embed ../../stdlib ./stdlib
//go:generate go run ./embed ../../docs ./docs
//go:generate go run ./embed ../../contrib ./contrib
//go:generate go run ./embed ../../examples ./examples
//go:generate go run ./embed ../../README.md ./README.md
//go:generate go run ./embed ../../CONTRIBUTING.md ./CONTRIBUTING.md
//go:generate go run ./embed ../../LICENSE ./LICENSE
//go:embed stdlib docs contrib examples README.md CONTRIBUTING.md LICENSE
var embeddedFS embed.FS

// Version is set at build time via -ldflags
var Version = "dev"

// setupInterpreter creates and configures a Duso interpreter
func setupInterpreter(scriptPath string, verbose, debug, noStdin bool, configStr string) (*script.Interpreter, error) {
	// Create interpreter
	interp := script.NewInterpreter(verbose)
	interp.SetDebugMode(debug)
	interp.SetNoStdin(noStdin)

	// Set the file path for error reporting
	interp.SetFilePath(scriptPath)

	// Get the directory of the script for file operations
	scriptDir := filepath.Dir(scriptPath)
	if scriptDir == "" {
		scriptDir = "."
	}
	interp.SetScriptDir(scriptDir)

	// Register CLI functions
	if err := cli.RegisterFunctions(interp, cli.RegisterOptions{
		ScriptDir: scriptDir,
		DebugMode: debug,
	}); err != nil {
		return nil, fmt.Errorf("could not register CLI functions: %w", err)
	}

	// Initialize sys datastore with config
	script.InitSystemMetrics()
	sysDs := script.GetDatastore("sys", nil)
	if configStr != "" {
		config, err := parseConfigString(configStr)
		if err != nil {
			return nil, err
		}
		if config != nil {
			sysDs.Set("config", config)
		}
	}

	return interp, nil
}

// runScript executes a Duso script with the given configuration
func runScript(scriptPath string, source []byte, verbose, debug, noStdin bool, configStr string) (string, error) {
	interp, err := setupInterpreter(scriptPath, verbose, debug, noStdin, configStr)
	if err != nil {
		return "", err
	}

	// Execute the script
	return interp.Execute(string(source))
}

func printLogo(noColor bool) {
	if noColor {
		// Plain text version - keep the pretty box, no colors
		fmt.Fprintf(os.Stderr, "\n             \n")
		fmt.Fprintf(os.Stderr, "         █   \n")
		fmt.Fprintf(os.Stderr, "     ▄ ▄ █      Duso\n")
		fmt.Fprintf(os.Stderr, "   █ █ █ █      Embeddable scripting language\n")
		fmt.Fprintf(os.Stderr, "       ▄ █      %s\n", Version)
		fmt.Fprintf(os.Stderr, "       █ ▀   \n")
		fmt.Fprintf(os.Stderr, "             \n")
		fmt.Fprintf(os.Stderr, "\n")
		return
	}

	// ANSI code for bright white foreground on blue background
	styled := "\033[97;44m"
	bold := "\033[1;97m"
	gray := "\033[90m"
	reset := "\033[0m"

	// Print logo with title to the right
	fmt.Fprintf(os.Stderr, "\n  %s             %s\n", styled, reset)
	fmt.Fprintf(os.Stderr, "  %s         █   %s\n", styled, reset)
	fmt.Fprintf(os.Stderr, "  %s     ▄ ▄ █   %s   %sDuso%s\n", styled, reset, bold, reset)
	fmt.Fprintf(os.Stderr, "  %s   █ █ █ █   %s   Embeddable scripting language\n", styled, reset)
	fmt.Fprintf(os.Stderr, "  %s       ▄ █   %s   %s%s%s\n", styled, reset, gray, Version, reset)
	fmt.Fprintf(os.Stderr, "  %s       █ ▀   %s\n", styled, reset)
	fmt.Fprintf(os.Stderr, "  %s             %s\n", styled, reset)
	fmt.Fprintf(os.Stderr, "\n")
}

// runREPLLoop executes a REPL loop with the given interpreter, prompt, and exit behavior.
// If exitOnC is true, the 'c' command will exit the loop (for debug REPL).
// Otherwise, only 'exit' command exits (for normal REPL).
// If useContext is true, uses EvalInContext to preserve scope (for debug REPL inside nested scopes).
// If env is provided, evaluates expressions in that specific environment (for breakpoint scope).
func runREPLLoop(interp *script.Interpreter, prompt string, exitOnC bool, useContext bool, env *script.Environment) error {
	scanner := bufio.NewScanner(os.Stdin)
	var input strings.Builder

	for {
		// Determine current prompt
		currentPrompt := prompt
		if input.Len() > 0 {
			currentPrompt = strings.Repeat(" ", len(prompt)-2) + "> "
		}

		// Print prompt
		fmt.Fprint(os.Stderr, currentPrompt)

		// Read line
		if !scanner.Scan() {
			// EOF
			return nil
		}

		line := scanner.Text()

		// Handle line continuation
		if strings.HasSuffix(line, "\\") {
			// Remove trailing backslash and newline, append to input
			input.WriteString(strings.TrimSuffix(line, "\\"))
			input.WriteString("\n")
			continue
		}

		// Append final line
		input.WriteString(line)
		code := input.String()
		input.Reset()

		// Check for exit command in debug mode
		if exitOnC && code == "c" {
			return nil
		}

		// Execute code
		var output string
		var err error
		if env != nil {
			// Evaluate in specific environment (breakpoint scope)
			output, err = interp.EvalInEnvironment(code, env)
		} else if useContext {
			output, err = interp.EvalInContext(code)
		} else {
			output, err = interp.Execute(code)
		}
		if err != nil {
			// Check for exit() call (specific error message)
			if strings.Contains(err.Error(), "exit") {
				return err
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		// Print output if any
		if output != "" {
			fmt.Print(output)
		}
	}
}

// handleDebugEvent processes a debug event from a child script
// Displays the full invocation stack and enters debug REPL
func handleDebugEvent(interp *script.Interpreter, event *script.DebugEvent, noColor bool) {
	// Format location info
	loc := event.FilePath
	if event.Position.Line > 0 {
		loc = fmt.Sprintf("%s:%d", loc, event.Position.Line)
		if event.Position.Column > 0 {
			loc = fmt.Sprintf("%s:%d", loc, event.Position.Column)
		}
	}

	// Print invocation stack (how we got here)
	if !noColor {
		brightRed := "\033[91m"
		reset := "\033[0m"
		if loc != "" {
			fmt.Fprintf(os.Stderr, "\n%s[Debug] Error in child script at %s%s\n", brightRed, loc, reset)
		} else {
			fmt.Fprintf(os.Stderr, "\n%s[Debug] Error in child script%s\n", brightRed, reset)
		}
	} else {
		if loc != "" {
			fmt.Fprintf(os.Stderr, "\n[Debug] Error in child script at %s\n", loc)
		} else {
			fmt.Fprintf(os.Stderr, "\n[Debug] Error in child script\n")
		}
	}

	// Print invocation stack (chain of run() calls that led to this error)
	if event.InvocationStack != nil {
		fmt.Fprintf(os.Stderr, "\nInvocation stack:\n")
		frame := event.InvocationStack
		for frame != nil {
			fmt.Fprintf(os.Stderr, "  %s", frame.Reason)
			if frame.Filename != "" {
				fmt.Fprintf(os.Stderr, " (%s:%d:%d)", frame.Filename, frame.Line, frame.Col)
			}
			fmt.Fprintf(os.Stderr, "\n")
			frame = frame.Parent
		}
	}

	// Show the error message
	if event.Message != "" {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", event.Message)
	}

	// Create a synthetic breakpoint error for the REPL
	bpErr := &script.BreakpointError{
		FilePath:  event.FilePath,
		Position:  event.Position,
		CallStack: event.CallStack,
		Env:       event.Env,
	}

	// Open debug REPL
	debugREPL(interp, bpErr, noColor)

	// Signal the child script to resume
	select {
	case event.ResumeChan <- true:
	default:
		// Channel closed or already received, skip
	}
}

// debugREPL enters a debug REPL at a breakpoint, allowing variable inspection.
func debugREPL(interp *script.Interpreter, bpErr *script.BreakpointError, noColor bool) error {
	// Format breakpoint location with position info
	loc := bpErr.FilePath
	if bpErr.Position.Line > 0 {
		loc = fmt.Sprintf("%s:%d", loc, bpErr.Position.Line)
		if bpErr.Position.Column > 0 {
			loc = fmt.Sprintf("%s:%d", loc, bpErr.Position.Column)
		}
	}

	// Add bright red color if colors are enabled
	if !noColor {
		brightRed := "\033[91m"
		reset := "\033[0m"
		if loc != "" {
			fmt.Fprintf(os.Stderr, "\n%s[Debug] Breakpoint hit at %s%s\n", brightRed, loc, reset)
		} else {
			fmt.Fprintf(os.Stderr, "\n%s[Debug] Error%s\n", brightRed, reset)
		}
	} else {
		if loc != "" {
			fmt.Fprintf(os.Stderr, "\n[Debug] Breakpoint hit at %s\n", loc)
		} else {
			fmt.Fprintf(os.Stderr, "\n[Debug] Error\n")
		}
	}

	// Show source code context around the breakpoint
	if bpErr.Position.Line > 0 {
		showSourceContext(bpErr.FilePath, bpErr.Position.Line, bpErr.Position.Column, noColor)
	}

	// Print call stack from the breakpoint error
	if len(bpErr.CallStack) > 0 {
		fmt.Fprintf(os.Stderr, "\nCall stack:\n")
		for i := len(bpErr.CallStack) - 1; i >= 0; i-- {
			frame := bpErr.CallStack[i]
			fmt.Fprintf(os.Stderr, "  at %s", frame.FunctionName)
			if frame.FilePath != "" {
				fmt.Fprintf(os.Stderr, " (%s:%d", frame.FilePath, frame.Position.Line)
				if frame.Position.Column > 0 {
					fmt.Fprintf(os.Stderr, ":%d", frame.Position.Column)
				}
				fmt.Fprintf(os.Stderr, ")")
			}
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	// If stdin is disabled, print warning and skip REPL
	if interp.GetEvaluator().NoStdin {
		fmt.Fprintf(os.Stderr, "\nwarning: stdin disabled, assuming 'c' to continue\n")
		return nil
	}

	fmt.Fprintf(os.Stderr, "\nType 'c' to continue, or inspect variables.\n")
	return runREPLLoop(interp, "debug> ", true, true, bpErr.Env)
}

// showSourceContext displays the source code around a breakpoint
func showSourceContext(filePath string, line int, col int, noColor bool) {
	// Read the file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return // Silently fail if we can't read the file
	}

	lines := strings.Split(string(source), "\n")
	if line < 1 || line > len(lines) {
		return
	}

	// Show 2 lines before, the line itself, and 3 lines after
	start := line - 2
	if start < 1 {
		start = 1
	}
	end := line + 3
	if end > len(lines) {
		end = len(lines)
	}

	fmt.Fprintf(os.Stderr, "\n")

	// Calculate width needed for line numbers
	lineNumWidth := len(fmt.Sprintf("%d", end))

	for i := start; i <= end; i++ {
		lineContent := ""
		if i <= len(lines) {
			lineContent = lines[i-1]
		}

		// Blank line before the breakpoint line
		if i == line {
			fmt.Fprintf(os.Stderr, "\n")
		}

		// Highlight the line with the breakpoint
		if i == line {
			if !noColor {
				fmt.Fprintf(os.Stderr, "\033[93m%*d | %s\033[0m\n", lineNumWidth, i, lineContent)
			} else {
				fmt.Fprintf(os.Stderr, "%*d | %s\n", lineNumWidth, i, lineContent)
			}

			// Show column marker if column is specified
			if col > 0 {
				// Account for line number width + " | " separator
				marker := strings.Repeat(" ", lineNumWidth+3+col-1) + "^"
				if !noColor {
					fmt.Fprintf(os.Stderr, "\033[91m%s\033[0m\n", marker) // Bright red for caret
				} else {
					fmt.Fprintf(os.Stderr, "%s\n", marker)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "%*d | %s\n", lineNumWidth, i, lineContent)
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
}

// parseConfigString parses a config string like "key=value, key=value" into a map
func parseConfigString(configStr string) (map[string]any, error) {
	if configStr == "" {
		return nil, nil
	}

	// Create temp interpreter to parse the config as Duso code
	interp := script.NewInterpreter(false)

	// Execute assignment to parse the object
	_, err := interp.Execute("__cfg__ = {" + configStr + "}")
	if err != nil {
		return nil, fmt.Errorf("invalid -config format: %v", err)
	}

	// Get the value from environment
	cfgVal, err := interp.GetEvaluator().GetEnv().Get("__cfg__")
	if err != nil {
		return nil, err
	}

	// Convert to map[string]any by extracting .Data from each Value
	objMap := cfgVal.AsObject()
	result := make(map[string]any)
	for k, v := range objMap {
		result[k] = v.Data
	}

	return result, nil
}

func runREPL(verbose, noColor, debugMode, noStdin bool) {
	printLogo(noColor)
	fmt.Fprintf(os.Stderr, "Duso REPL (type 'exit' to quit, use \\ for line continuation)\n\n")

	// Create interpreter with persistent state (ack)
	interp := script.NewInterpreter(verbose)
	interp.SetDebugMode(debugMode)
	interp.SetNoStdin(noStdin)
	interp.SetScriptDir(".")

	// Register CLI functions
	if err := cli.RegisterFunctions(interp, cli.RegisterOptions{
		ScriptDir: ".",
		DebugMode: debugMode,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not register CLI functions: %v\n", err)
		os.Exit(1)
	}

	if err := runREPLLoop(interp, "duso> ", false, false, nil); err != nil {
		// Check for exit() call
		if !strings.Contains(err.Error(), "exit") {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	fmt.Fprintf(os.Stderr, "\nGoodbye!\n")
}

func main() {
	verbose := flag.Bool("v", false, "Enable verbose output")
	showDoc := flag.Bool("doc", false, "Display documentation for a module (defaults to 'index' if no module specified)")
	code := flag.String("c", "", "Execute inline code")
	noColor := flag.Bool("no-color", false, "Disable ANSI color output")
	noStdin := flag.Bool("no-stdin", false, "Disable stdin (input() returns empty, breakpoint/watch skip REPL)")
	repl := flag.Bool("repl", false, "Start interactive REPL mode")
	debug := flag.Bool("debug", false, "Enable debug mode (breakpoint() pauses execution)")
	docserver := flag.Bool("docserver", false, "Launch documentation server and open browser")
	showVersion := flag.Bool("version", false, "Show version and exit")
	showHelp := flag.Bool("help", false, "Show help and exit")
	libPath := flag.String("lib-path", "", "Add directory to module search path (prepends to DUSO_LIB)")
	configStr := flag.String("config", "", "Pass configuration as key=value pairs (e.g., -config \"port=8080, debug=true\")")
	flag.Parse()

	// Initialize embedded filesystem for file I/O operations (needed before --help)
	cli.SetEmbeddedFS(embeddedFS)

	// Handle --version
	if *showVersion {
		fmt.Printf("Duso %s\n", Version)
		os.Exit(0)
	}

	// Handle --help
	if *showHelp {
		printLogo(*noColor)
		markdownFn := cli.NewMarkdownFunctionWithOptions(*noColor)
		helpContent, err := cli.ReadEmbeddedFile("/EMBED/docs/cli/help.md")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not read help: %v\n", err)
			os.Exit(1)
		}
		formatted, err := markdownFn(map[string]any{"0": string(helpContent)})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not format help: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(os.Stderr, formatted)
		fmt.Fprint(os.Stderr, "\n\n")
		os.Exit(0)
	}

	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		*noColor = true
	}

	// Prepend --lib-path to DUSO_LIB env var if provided
	if *libPath != "" {
		existing := os.Getenv("DUSO_LIB")
		if existing != "" {
			os.Setenv("DUSO_LIB", *libPath+string(os.PathListSeparator)+existing)
		} else {
			os.Setenv("DUSO_LIB", *libPath)
		}
	}

	// Handle -docserver flag
	if *docserver {
		scriptPath := "stdlib/docserver/docserver.du"

		// Read the script file (try local first, then embedded)
		source, err := os.ReadFile(scriptPath)
		if err != nil {
			// Try embedded files if local read failed
			source, err = cli.ReadEmbeddedFile("/EMBED/" + scriptPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not read docserver: %v\n", err)
				os.Exit(1)
			}
		}

		// Copy URL to clipboard
		url := "http://localhost:5150"
		cmd := exec.Command("bash", "-c", fmt.Sprintf("echo -n '%s' | pbcopy", url))
		if err := cmd.Run(); err != nil {
			// pbcopy failed (macOS only), try other methods
			switch runtime.GOOS {
			case "linux":
				cmd = exec.Command("bash", "-c", fmt.Sprintf("echo -n '%s' | xclip -selection clipboard", url))
				_ = cmd.Run()
			case "windows":
				cmd = exec.Command("powershell", "-Command", fmt.Sprintf("'%s' | Set-Clipboard", url))
				_ = cmd.Run()
			}
		}

		fmt.Printf("URL copied to clipboard: %s\n", url)

		// Run the server script (blocks on server.start())
		_, _ = runScript(scriptPath, source, *verbose, *debug, *noStdin, *configStr)
		os.Exit(0)
	}

	// Handle REPL mode
	if *repl {
		script.InitSystemMetrics()
		sysDs := script.GetDatastore("sys", nil)
		if *configStr != "" {
			config, err := parseConfigString(*configStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if config != nil {
				sysDs.Set("config", config)
			}
		}
		runREPL(*verbose, *noColor, *debug, *noStdin)
		os.Exit(0)
	}

	args := flag.Args()

	// Handle -c flag (execute inline code)
	if *code != "" {
		output, err := runScript("<inline>", []byte(*code), *verbose, *debug, *noStdin, *configStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Output the result
		if output != "" {
			fmt.Print(output)
		}
		os.Exit(0)
	}

	// Handle -doc flag (show module documentation and exit)
	if *showDoc {
		scriptPath := "stdlib/doccli/doccli.du"

		// Read the script file (try local first, then embedded)
		source, err := os.ReadFile(scriptPath)
		if err != nil {
			// Try embedded files if local read failed
			source, err = cli.ReadEmbeddedFile("/EMBED/" + scriptPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not read doccli script: %v\n", err)
				os.Exit(1)
			}
		}

		// Determine the topic (defaults to "index" if not specified)
		topic := "index"
		if len(args) > 0 {
			topic = args[0]
		}

		// Initialize sys datastore and set up config and doc_topic separately
		script.InitSystemMetrics()
		sysDs := script.GetDatastore("sys", nil)
		if *configStr != "" {
			config, err := parseConfigString(*configStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if config != nil {
				sysDs.Set("config", config)
			}
		}
		// Set doc_topic separately, so it doesn't interfere with user's config
		sysDs.Set("doc_topic", topic)

		// Run the doccli script
		_, err = runScript(scriptPath, source, *verbose, *debug, *noStdin, *configStr)
		if err != nil {
			if !strings.Contains(err.Error(), "exit") {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(args) == 0 {
		printLogo(*noColor)
		// Load and render help from embedded markdown
		markdownFn := cli.NewMarkdownFunctionWithOptions(*noColor)
		helpContent, err := cli.ReadEmbeddedFile("/EMBED/docs/cli/help.md")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not read help: %v\n", err)
			os.Exit(1)
		}
		formatted, err := markdownFn(map[string]any{"0": string(helpContent)})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not format help: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(os.Stderr, formatted)
		fmt.Fprint(os.Stderr, "\n\n")
		os.Exit(1)
	}

	scriptPath := args[0]

	// Read the script file (try local first, then embedded)
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		// Try embedded files if local read failed
		source, err = cli.ReadEmbeddedFile("/EMBED/" + scriptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not read script '%s': %v\n", scriptPath, err)
			os.Exit(1)
		}
	}

	// Set up the interpreter
	interp, err := setupInterpreter(scriptPath, *verbose, *debug, *noStdin, *configStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Execute the script
	var output string

	if *debug {
		// Debug mode: parse and execute statement-by-statement
		lexer := script.NewLexer(string(source))
		tokens := lexer.Tokenize()

		parser := script.NewParserWithFile(tokens, scriptPath)
		program, parseErr := parser.Parse()
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", parseErr)
			os.Exit(1)
		}

		// Start background listener for debug events from child scripts
		go func() {
			for event := range interp.GetDebugEventChan() {
				if event != nil {
					// Handle the debug event (opens REPL)
					handleDebugEvent(interp, event, *noColor)
					// After REPL closes, send resume signal so child can continue
					if event.ResumeChan != nil {
						event.ResumeChan <- true
					}
				}
			}
		}()

		// Execute each statement
		for _, stmt := range program.Statements {
			execErr := interp.ExecuteNode(stmt)
			if execErr != nil {
				// Check for BreakpointError
				if bpErr, ok := execErr.(*script.BreakpointError); ok {
					// Enter debug REPL
					if debugErr := debugREPL(interp, bpErr, *noColor); debugErr != nil {
						if strings.Contains(debugErr.Error(), "exit") {
							break
						}
						fmt.Fprintf(os.Stderr, "Error in debug REPL: %v\n", debugErr)
						os.Exit(1)
					}
					continue
				}

				// Check for ExitExecution
				if _, ok := execErr.(*script.ExitExecution); ok {
					break
				}

				// In debug mode, any other error triggers debug REPL
				// Build a breakpoint error with available position info
				bpErr := &script.BreakpointError{
					Env: interp.GetEvaluator().GetEnv(),
				}

				// Extract position info if available
				if dusoErr, ok := execErr.(*script.DusoError); ok {
					bpErr.FilePath = dusoErr.FilePath
					bpErr.Position = dusoErr.Position
					bpErr.CallStack = dusoErr.CallStack
				}

				// Show the error message before entering REPL
				fmt.Fprintf(os.Stderr, "Error: %v\n", execErr)

				if debugErr := debugREPL(interp, bpErr, *noColor); debugErr != nil {
					if strings.Contains(debugErr.Error(), "exit") {
						break
					}
					fmt.Fprintf(os.Stderr, "Error in debug REPL: %v\n", debugErr)
					os.Exit(1)
				}
				continue
			}
		}

		output = interp.GetOutput()
	} else {
		// Normal mode: fast path execution
		var err error
		output, err = interp.Execute(string(source))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Output the result
	if output != "" {
		fmt.Print(output)
	}
}
