package runtime

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// BusySpinner manages a spinning busy cursor
type BusySpinner struct {
	mu        sync.Mutex
	stopChan  chan struct{}
	doneChan  chan struct{}
	running   bool
	message   string // Current message being displayed
	messageLen int   // Length of message for backspacing
}

var (
	globalBusySpinner *BusySpinner
	busySpinnerMu     sync.Mutex
)

// NewBusyFunction creates a busy() builtin that displays a spinning cursor with messages.
//
// busy("message")  - Print message to stderr and start animated spinner
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
func NewBusyFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		// Get the first argument (message or empty)
		message := ""
		if val, ok := args["0"]; ok {
			// Convert to string
			switch v := val.(type) {
			case string:
				message = v
			}
		}

		busySpinnerMu.Lock()
		if globalBusySpinner == nil {
			globalBusySpinner = &BusySpinner{}
		}
		spinner := globalBusySpinner
		busySpinnerMu.Unlock()

		spinner.mu.Lock()

		if message != "" {
			// Stop any existing animation and clear previous message
			if spinner.running {
				spinner.running = false
				close(spinner.stopChan)
				spinner.mu.Unlock()
				<-spinner.doneChan
				spinner.mu.Lock()

				// Backspace out the previous message + 1 character (the animated frame)
				backspaceCount := spinner.messageLen + 1
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
			}

			// Print message to stderr and start new spinner
			fmt.Fprint(os.Stderr, message)
			spinner.message = message
			spinner.messageLen = len(message)
			spinner.running = true
			spinner.stopChan = make(chan struct{})
			spinner.doneChan = make(chan struct{})

			go func() {
				defer close(spinner.doneChan)

				frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
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
			// Clear existing spinner
			if spinner.running {
				spinner.running = false
				close(spinner.stopChan)

				// Wait for goroutine to finish
				spinner.mu.Unlock()
				<-spinner.doneChan
				spinner.mu.Lock()

				// Backspace out the message + 1 character (the animated frame)
				backspaceCount := spinner.messageLen + 1
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

				spinner.message = ""
				spinner.messageLen = 0
			}

			spinner.mu.Unlock()
		}

		return nil, nil
	}
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
	if spinner.running {
		spinner.running = false
		close(spinner.stopChan)

		// Wait for goroutine to finish
		spinner.mu.Unlock()
		<-spinner.doneChan
		spinner.mu.Lock()

		// Backspace out the message + 1 character (the animated frame)
		backspaceCount := spinner.messageLen + 1
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

		spinner.message = ""
		spinner.messageLen = 0
	}
	spinner.mu.Unlock()
}
