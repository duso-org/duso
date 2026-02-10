package main

import (
	"bufio"
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/duso-org/duso/pkg/cli"
	"github.com/duso-org/duso/pkg/lsp"
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
func setupInterpreter(scriptPath string, verbose, debug, noStdin, noFiles bool, configStr string) (*script.Interpreter, error) {
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
		NoFiles:   noFiles,
	}, nil); err != nil {
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
func runScript(scriptPath string, source []byte, verbose, debug, noStdin, noFiles bool, configStr string) (string, error) {
	interp, err := setupInterpreter(scriptPath, verbose, debug, noStdin, noFiles, configStr)
	if err != nil {
		return "", err
	}

	// If in debug mode, start background listener for debug events from child scripts
	if debug {
		go func() {
			for event := range interp.GetDebugEventChan() {
				if event != nil {
					// Handle the debug event (opens REPL)
					handleDebugEvent(interp, event, false)
					// After REPL closes, send resume signal so child can continue
					if event.ResumeChan != nil {
						event.ResumeChan <- true
					}
				}
			}
		}()
	}

	// Execute the script
	return interp.Execute(string(source))
}

func printLogo(noColor bool) {
	if noColor {
		// Plain text version - keep the pretty box, no colors
		fmt.Fprintf(os.Stderr, "\n ┌───────────────┐\n")
		fmt.Fprintf(os.Stderr, " │               │\n")
		fmt.Fprintf(os.Stderr, " │          █    │\n")
		fmt.Fprintf(os.Stderr, " │      ▄ ▄ █    │   Duso %s\n", Version)
		fmt.Fprintf(os.Stderr, " │    █ █ █ █    │   Scripted Intelligence\n")
		fmt.Fprintf(os.Stderr, " │        ▄ █    │   ©2026 Ludonode LLC\n")
		fmt.Fprintf(os.Stderr, " │        █ ▀    │   \n")
		fmt.Fprintf(os.Stderr, " │               │\n")
		fmt.Fprintf(os.Stderr, " └───────────────┘\n")
		fmt.Fprintf(os.Stderr, "\n")
		return
	}

	// ANSI code for bright white foreground on blue background
	styled := "\033[97;44m"
	bold := "\033[1;97m"
	gray := "\033[90m"
	reset := "\033[0m"

	// Print logo with title to the right
	fmt.Fprintf(os.Stderr, "\n  %s               %s\n", styled, reset)
	fmt.Fprintf(os.Stderr, "  %s          █    %s\n", styled, reset)
	fmt.Fprintf(os.Stderr, "  %s      ▄ ▄ █    %s    %sDuso%s %s%s%s\n", styled, reset, bold, reset, gray, Version, reset)
	fmt.Fprintf(os.Stderr, "  %s    █ █ █ █    %s    Scripted Intelligence\n", styled, reset)
	fmt.Fprintf(os.Stderr, "  %s        ▄ █    %s    %s©2026 Ludonode%s\n", styled, reset, gray, reset)
	fmt.Fprintf(os.Stderr, "  %s        █ ▀    %s   \n", styled, reset)
	fmt.Fprintf(os.Stderr, "  %s               %s\n", styled, reset)
	fmt.Fprintf(os.Stderr, "\n")
}

// printFormattedHelp executes a duso script to render help.md through the markdown module
func printFormattedHelp(noColor bool) error {
	interp := script.NewInterpreter(false)

	// Register CLI functions
	opts := cli.RegisterOptions{ScriptDir: "."}
	if err := cli.RegisterFunctions(interp, opts, nil); err != nil {
		return fmt.Errorf("failed to register CLI functions: %w", err)
	}

	// Build inline script that loads markdown module and renders help
	// If noColor is set, skip the markdown formatting
	var dusoScript string
	if noColor {
		dusoScript = `print(require("markdown").text(load("/EMBED/docs/cli/help.md")))`
	} else {
		dusoScript = `print(require("markdown").ansi(load("/EMBED/docs/cli/help.md")))`
	}

	// Execute the inline script
	_, err := interp.Execute(dusoScript)
	if err != nil {
		return fmt.Errorf("failed to render help: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	return nil
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
	}, nil); err != nil {
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

// initProject handles the -init flag to create a new project
func initProject(projectName string, noColor bool) error {
	if projectName == "" {
		return fmt.Errorf("project name is required: duso -init my-project")
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current directory: %v", err)
	}

	projectPath := filepath.Join(cwd, projectName)

	// List available templates
	templates, err := listTemplates()
	if err != nil {
		return fmt.Errorf("could not list templates: %v", err)
	}

	if len(templates) == 0 {
		return fmt.Errorf("no templates found")
	}

	// Show templates and let user choose
	fmt.Println("Available templates:")
	for i, tmpl := range templates {
		fmt.Printf("\n  %d. %s", i+1, tmpl)
	}
	fmt.Print("\n\nSelect a template (1-" + fmt.Sprintf("%d", len(templates)) + "): ")

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Parse selection
	choice := 0
	fmt.Sscanf(input, "%d", &choice)

	if choice < 1 || choice > len(templates) {
		return fmt.Errorf("invalid selection")
	}

	selectedTemplate := templates[choice-1]

	// Confirm creation
	fmt.Printf("\nCreate new project at:\n%s? [y/N]: ", projectPath)
	input, _ = reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(input)) != "y" {
		fmt.Println("Cancelled.")
		return nil
	}

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("could not create project directory: %v", err)
	}

	// Copy template files
	if err := copyTemplate(selectedTemplate, projectPath); err != nil {
		// Clean up on failure
		os.RemoveAll(projectPath)
		return fmt.Errorf("could not copy template: %v", err)
	}

	fmt.Printf("\n✅ Project created at: %s\n", projectPath)
	fmt.Printf("Run with: cd %s && duso %s.du\n", projectName, selectedTemplate)

	return nil
}

// listTemplates returns a list of available template names
func listTemplates() ([]string, error) {
	entries, err := embeddedFS.ReadDir("examples/init")
	if err != nil {
		return nil, err
	}

	var templates []string
	for _, entry := range entries {
		if entry.IsDir() {
			templates = append(templates, entry.Name())
		}
	}

	return templates, nil
}

// copyTemplate copies a template from embedded FS to the target directory
func copyTemplate(templateName, targetPath string) error {
	templatePath := filepath.Join("examples/init", templateName)

	// Walk through template directory
	entries, err := embeddedFS.ReadDir(templatePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(templatePath, entry.Name())
		dstPath := filepath.Join(targetPath, entry.Name())

		if entry.IsDir() {
			// Create subdirectory
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			// Recursively copy contents
			if err := copyTemplateDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			data, err := embeddedFS.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyTemplateDir recursively copies a directory from embedded FS
func copyTemplateDir(srcPath, dstPath string) error {
	entries, err := embeddedFS.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		src := filepath.Join(srcPath, entry.Name())
		dst := filepath.Join(dstPath, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dst, 0755); err != nil {
				return err
			}
			if err := copyTemplateDir(src, dst); err != nil {
				return err
			}
		} else {
			data, err := embeddedFS.ReadFile(src)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dst, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	verbose := flag.Bool("v", false, "Enable verbose output")
	showDoc := flag.Bool("doc", false, "Display documentation for a module (defaults to 'index' if no module specified)")
	code := flag.String("c", "", "Execute inline code")
	noColor := flag.Bool("no-color", false, "Disable ANSI color output")
	noStdin := flag.Bool("no-stdin", false, "Disable stdin (input() returns empty, breakpoint/watch skip REPL)")
	repl := flag.Bool("repl", false, "Start interactive REPL mode")
	debug := flag.Bool("debug", false, "Enable debug mode (breakpoint() pauses execution)")
	stdinPort := flag.Int("stdin-port", 0, "Port for HTTP stdin/stdout transport (enables remote interaction with input() calls)")
	_ = flag.String("debug-bind", "localhost", "Bind address for HTTP debug server (deprecated)")
	docserver := flag.Bool("docserver", false, "Launch documentation server and open browser")
	noFiles := flag.Bool("no-files", false, "Restrict to /STORE/ and /EMBED/ only (disable filesystem access)")
	showVersion := flag.Bool("version", false, "Show version and exit")
	showHelp := flag.Bool("help", false, "Show help and exit")
	libPath := flag.String("lib-path", "", "Add directory to module search path (prepends to DUSO_LIB)")
	configStr := flag.String("config", "", "Pass configuration as key=value pairs (e.g., -config \"port=8080, debug=true\")")
	doInit := flag.Bool("init", false, "Initialize a new Duso project")
	lspStdio := flag.Bool("lsp", false, "Start LSP server on stdio")
	lspTCP := flag.String("lsp-tcp", "", "Start LSP server on TCP port (e.g., -lsp-tcp 9999)")
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
		if err := printFormattedHelp(*noColor); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not display help: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle -init flag
	if *doInit {
		args := flag.Args()
		projName := ""
		if len(args) > 0 {
			projName = args[0]
		}
		if err := initProject(projName, *noColor); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
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
		_, _ = runScript(scriptPath, source, *verbose, *debug, *noStdin, *noFiles, *configStr)
		os.Exit(0)
	}

	// Handle LSP mode (before REPL mode so it takes priority)
	if *lspStdio || *lspTCP != "" {
		// Create a minimal interpreter for LSP
		interp := script.NewInterpreter(*verbose)
		interp.SetDebugMode(*debug)
		interp.SetNoStdin(*noStdin)

		// Register CLI functions for LSP
		if err := cli.RegisterFunctions(interp, cli.RegisterOptions{
			ScriptDir: ".",
			DebugMode: *debug,
			NoFiles:   *noFiles,
		}, nil); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not register CLI functions: %v\n", err)
			os.Exit(1)
		}

		// Create LSP server
		server := lsp.NewServer(interp, embeddedFS)

		// Start transport
		var transport lsp.Transport
		if *lspTCP != "" {
			transport = lsp.NewTCPTransport(*lspTCP)
		} else {
			transport = lsp.NewStdioTransport()
		}

		if err := transport.Start(server); err != nil {
			fmt.Fprintf(os.Stderr, "LSP server error: %v\n", err)
			os.Exit(1)
		}
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
		output, err := runScript("<inline>", []byte(*code), *verbose, *debug, *noStdin, *noFiles, *configStr)
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
		_, err = runScript(scriptPath, source, *verbose, *debug, *noStdin, *noFiles, *configStr)
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
		if err := printFormattedHelp(*noColor); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not display help: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
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

	// Create HTTP stdin/stdout server if -stdin-port is specified
	var stdinServer *cli.StdinHTTPServer
	if *stdinPort > 0 {
		stdinServer = cli.NewStdinHTTPServer(*stdinPort, "localhost")
		go func() {
			if err := stdinServer.Start(); err != nil && err.Error() != "http: Server closed" {
				fmt.Fprintf(os.Stderr, "Error starting stdin/stdout server: %v\n", err)
			}
		}()
		defer stdinServer.Stop()

		// Print usage instructions for HTTP stdin/stdout transport
		fmt.Fprintf(os.Stderr, "HTTP stdin/stdout transport listening on http://localhost:%d\n", *stdinPort)
		fmt.Fprintf(os.Stderr, "  GET /        - Read accumulated output\n")
		fmt.Fprintf(os.Stderr, "  GET /input   - Block until input() is called, returns accumulated output\n")
		fmt.Fprintf(os.Stderr, "  POST /input  - Send input in request body to waiting input() call\n\n")
	}

	// Set up the interpreter
	interp, err := setupInterpreter(scriptPath, *verbose, *debug, *noStdin, *noFiles, *configStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If HTTP stdin/stdout is enabled, override the output/input writers
	if stdinServer != nil {
		interp.OutputWriter = stdinServer.GetOutputWriter()
		interp.InputReader = stdinServer.GetInputReader()
	}

	// Execute the script
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

		// Use unified ExecuteScript for statement-by-statement execution with breakpoint handling
		frame := &script.InvocationFrame{
			Filename: scriptPath,
			Reason:   "main",
			Details:  map[string]any{},
		}
		ctx := &script.RequestContext{
			Frame:    frame,
			ExitChan: make(chan any),
		}
		result := script.ExecuteScript(program, interp, frame, ctx, context.Background())
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", result.Error)
			os.Exit(1)
		}

	} else {
		// Normal mode: fast path execution
		var err error
		_, err = interp.Execute(string(source))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
