package cli

import (
	"fmt"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/duso-org/duso/pkg/script"
)

// BusySpinner manages a spinning busy cursor
type BusySpinner struct {
	mu         sync.Mutex
	stopChan   chan struct{}
	doneChan   chan struct{}
	running    bool
	message    string // Current message being displayed
	messageLen int    // Length of message for backspacing
}

var (
	globalBusySpinner *BusySpinner
	busySpinnerMu     sync.Mutex
)

// builtinBusy displays a spinning cursor with optional message.
//
// busy("message")  - Print message to stderr and start animated spinner
// busy("")         - Start animated spinner with no message
// busy()           - Stop spinner, clear message with backspaces, return immediately
//
// Uses beautiful Braille pattern animation: ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
// Spinner output goes to stderr to avoid polluting stdout redirection.
//
// Example:
//
//	busy("Processing data")
//	sleep(2)
//	busy()
//	print("Done!")
func builtinBusy(evaluator *script.Evaluator, args map[string]any) (any, error) {
	// Check if an argument was provided
	message := ""

	busySpinnerMu.Lock()
	if globalBusySpinner == nil {
		globalBusySpinner = &BusySpinner{}
	}
	spinner := globalBusySpinner
	busySpinnerMu.Unlock()

	spinner.mu.Lock()

	if val, ok := args["0"]; ok {
		// Stop any existing animation and clear previous message
		if spinner.running {
			clearSpinnerLocked(spinner)
		}

		// Convert to string
		switch v := val.(type) {
		case string:
			message = v
		}

		// Hide cursor and print message to stderr with space before spinner (if message exists)
		fmt.Fprint(os.Stderr, "\033[?25l")
		fmt.Fprint(os.Stderr, message)
		if len(message) > 0 {
			fmt.Fprint(os.Stderr, " ")
		}
		spinner.message = message
		spinner.messageLen = utf8.RuneCountInString(message)
		spinner.running = true
		spinner.stopChan = make(chan struct{})
		spinner.doneChan = make(chan struct{})

		go func() {
			defer close(spinner.doneChan)

			// indecision? maybe. bouncy braille wins. seriously what is wrong with me.
			//frames := []string{"◴", "◷", "◶", "◵"}
			//frames := []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "▊", "▋", "▌", "▍", "▎", "▏"}
			//frames := []string{" ", "▃", "▄", "▅", "▆", "▇", "▆", "▅", "▄", "▃"}
			//frames := []string{" ", "▃", "▄", "▅", "▆", "▉", "▊", "▋", "▌", "▍", "▎", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "▆", "▅", "▄", "▃"}
			//frames := []string{"▀", "▜", "▐", "▙", "▌", "▛"}
			//frames := []string{" ", "░", "▒", "▓", "█", "▓", "▒", "░"}
			//frames := []string{"⣄", "⣦", "⣧", "⣨", "⣵", "⣰", "⣱", "⣐", "⣷", "⣮", "⣎", "⣌", "⣊", "⣄", "⣈"}
			//frames := []string{"⣉", "⣝", "⡾", "⣿", "⢷", "⣫"}
			//frames := []string{"⠉", "⠙", "⠸", "⢰", "⣠", "⣀", "⣄", "⡆", "⠇", "⠋"}
			//frames := []string{" ", "⠁", "⠉", "⠙", "⠹", "⢹", "⣹", "⣽", "⣿", "⣷", "⣧", "⣇", "⡇", "⠇", "⠃", "⠁"}
			//frames := []string{"⡀", "⡄", "⡆", "⡇", "⡏", "⡟", "⡿", "⣿", "⢿", "⢻", "⢹", "⢸", "⢰", "⢠", "⢀", " "}
			//frames := []string{"⠂", "⠃", "⠋", "⠛", "⠻", "⢻", "⣻", "⣿", "⣾", "⣶", "⣦", "⣆", "⡆", "⠆"}
			//frames := []string{"⡀", "⡄", "⡆", "⡇", "⡏", "⡟", "⡿", "⣿", "⣻", "⣹", "⣸", "⣰", "⣠", "⣀", "⡀", " "}
			//frames := []string{"⣉", "⡜", "⠶", "⢣"}
			//frames := []string{"□", "◫", "▤", "▨", "▦", "▩", "■", "▩", "▦", "▨", "▤", "◫"}
			frames := []string{"⣀", "⣤", "⣶", "⠿", "⠛", "⠛", "⠛", "⠶", "⣤", "⣀"}
			i := 0
			first := true

			for {
				// Check if we should stop before entering select
				spinner.mu.Lock()
				shouldStop := !spinner.running
				spinner.mu.Unlock()

				if shouldStop {
					return
				}

				select {
				case <-spinner.stopChan:
					return
				default:
					frame := frames[i%len(frames)]
					if first {
						fmt.Fprint(os.Stderr, frame)
						first = false
					} else {
						fmt.Fprint(os.Stderr, "\b"+frame)
					}
					i++
					time.Sleep(80 * time.Millisecond)
				}
			}
		}()

		spinner.mu.Unlock()
	} else {
		clearSpinnerLocked(spinner)
		spinner.mu.Unlock()
	}

	return nil, nil
}

// clearSpinnerLocked stops the spinner and clears it from the terminal.
// Assumes spinner.mu is already held by the caller.
func clearSpinnerLocked(spinner *BusySpinner) {
	if !spinner.running {
		return
	}

	spinner.running = false
	close(spinner.stopChan)

	// Wait for goroutine to finish
	spinner.mu.Unlock()
	<-spinner.doneChan
	spinner.mu.Lock()

	// Backspace out the message + optional space + 1 character (the animated frame)
	backspaceCount := spinner.messageLen + 1 // frame
	if spinner.messageLen > 0 {
		backspaceCount++ // add space if there's a message
	}
	for i := 0; i < backspaceCount; i++ {
		fmt.Fprint(os.Stderr, "\b")
	}
	// Clear with spaces
	for i := 0; i < backspaceCount; i++ {
		fmt.Fprint(os.Stderr, " ")
	}
	// Backspace again to return to start
	for i := 0; i < backspaceCount; i++ {
		fmt.Fprint(os.Stderr, "\b")
	}

	// Show cursor again
	fmt.Fprint(os.Stderr, "\033[?25h")

	spinner.message = ""
	spinner.messageLen = 0
}

// ClearBusySpinner clears any active spinner without the mutex lock
// Used internally by OutputWriter to auto-clear before printing
func ClearBusySpinner() {
	busySpinnerMu.Lock()
	spinner := globalBusySpinner
	busySpinnerMu.Unlock()

	if spinner == nil {
		return
	}

	spinner.mu.Lock()
	clearSpinnerLocked(spinner)
	spinner.mu.Unlock()
}
