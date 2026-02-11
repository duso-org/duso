package runtime

import (
	"fmt"
	"sync"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// BusySpinner manages a spinning busy cursor
type BusySpinner struct {
	mu       sync.Mutex
	stopChan chan struct{}
	doneChan chan struct{}
	running  bool
}

var (
	globalBusySpinner *BusySpinner
	busySpinnerMu     sync.Mutex
)

// NewBusyFunction creates a busy() builtin that displays a spinning cursor.
//
// busy(true)  - Start the spinner
// busy(false) - Stop the spinner
//
// Uses beautiful Braille pattern animation: ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
// It overwrites the same line using carriage returns.
//
// Example:
//
//	busy(true)
//	// do some work...
//	sleep(2)
//	busy(false)
//	print("Done!")
func NewBusyFunction(interp *script.Interpreter) func(*script.Evaluator, map[string]any) (any, error) {
	return func(evaluator *script.Evaluator, args map[string]any) (any, error) {
		// Get the first argument (enabled flag)
		enabled := false
		if val, ok := args["0"]; ok {
			// Convert to boolean
			switch v := val.(type) {
			case bool:
				enabled = v
			case float64:
				enabled = v != 0
			case string:
				enabled = v != "" && v != "false" && v != "0"
			}
		}

		busySpinnerMu.Lock()
		if globalBusySpinner == nil {
			globalBusySpinner = &BusySpinner{}
		}
		spinner := globalBusySpinner
		busySpinnerMu.Unlock()

		spinner.mu.Lock()
		defer spinner.mu.Unlock()

		if enabled {
			// Start spinner if not already running
			if !spinner.running {
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
							fmt.Print("\b \b")
							return
						}

						select {
						case <-spinner.stopChan:
							fmt.Print("\b \b")
							return
						default:
							frame := frames[i%len(frames)]
							if first {
								fmt.Print(frame)
								first = false
							} else {
								fmt.Print("\b" + frame)
							}
							i++
							time.Sleep(80 * time.Millisecond)
						}
					}
				}()
			}
		} else {
			// Stop spinner if running
			if spinner.running {
				spinner.running = false
				close(spinner.stopChan)

				// Wait for goroutine to finish cleaning up
				spinner.mu.Unlock()
				<-spinner.doneChan
				spinner.mu.Lock()
			}
		}

		return nil, nil
	}
}
