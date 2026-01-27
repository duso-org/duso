package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
//go:embed stdlib docs contrib examples
var embeddedFS embed.FS

// Version is set at build time via -ldflags
var Version = "dev"

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
		fmt.Fprintf(os.Stderr, "\n%s[Debug] Breakpoint hit at %s%s\n", brightRed, loc, reset)
	} else {
		fmt.Fprintf(os.Stderr, "\n[Debug] Breakpoint hit at %s\n", loc)
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
	showVersion := flag.Bool("version", false, "Show version and exit")
	showHelp := flag.Bool("help", false, "Show help and exit")
	libPath := flag.String("lib-path", "", "Add directory to module search path (prepends to DUSO_LIB)")
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

	// Handle REPL mode
	if *repl {
		runREPL(*verbose, *noColor, *debug, *noStdin)
		os.Exit(0)
	}

	args := flag.Args()

	// Handle -c flag (execute inline code)
	if *code != "" {
		// Create interpreter
		interp := script.NewInterpreter(*verbose)
		interp.SetDebugMode(*debug)
		interp.SetNoStdin(*noStdin)
		interp.SetScriptDir(".")

		// Register all CLI-specific functions with current directory as script dir
		if err := cli.RegisterFunctions(interp, cli.RegisterOptions{
			ScriptDir: ".",
			DebugMode: *debug,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not register CLI functions: %v\n", err)
			os.Exit(1)
		}

		// Execute the code
		output, err := interp.Execute(*code)
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
		// For -doc, we use current directory as script dir
		resolver := cli.NewModuleResolver(cli.RegisterOptions{ScriptDir: "."})
		docFn := cli.NewDocFunction(resolver)
		markdownFn := cli.NewMarkdownFunctionWithOptions(*noColor)

		// Build args map - only include "0" if a module name was provided
		docArgs := make(map[string]any)
		if len(args) > 0 {
			docArgs["0"] = args[0]
		}

		result, err := docFn(docArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if result != nil {
			// Format the result (including path header) with markdown rendering
			formatted, err := markdownFn(map[string]any{"0": result.(string)})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting: %v\n", err)
				os.Exit(1)
			}
			fmt.Println()
			fmt.Print(formatted)
			fmt.Print("\n\n")
		} else {
			docName := "index"
			if len(args) > 0 {
				docName = args[0]
			}
			fmt.Fprintf(os.Stderr, "Module not found: %s\n", docName)
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

	// Create interpreter
	interp := script.NewInterpreter(*verbose)
	interp.SetDebugMode(*debug)
	interp.SetNoStdin(*noStdin)

	// Set the file path for error reporting
	interp.SetFilePath(scriptPath)

	// Get the directory of the script for file operations
	scriptDir := filepath.Dir(scriptPath)
	interp.SetScriptDir(scriptDir)

	// Register all CLI-specific functions (load, save, include, require)
	// This is a single call that registers all optional CLI features
	if err := cli.RegisterFunctions(interp, cli.RegisterOptions{
		ScriptDir: scriptDir,
		DebugMode: *debug,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not register CLI functions: %v\n", err)
		os.Exit(1)
	}

	// Execute the script
	var output string

	if *debug {
		// Debug mode: parse and execute statement-by-statement
		lexer := script.NewLexer(string(source))
		tokens := lexer.Tokenize()

		parser := script.NewParser(tokens)
		program, parseErr := parser.Parse()
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", parseErr)
			os.Exit(1)
		}

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

				// Other errors
				fmt.Fprintf(os.Stderr, "Error: %v\n", execErr)
				os.Exit(1)
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
