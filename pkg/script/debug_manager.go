package script

import (
	"sync"
)

// DebugManager handles debug events sequentially.
// Scripts call Wait() synchronously and block until the user responds.
// The manager processes each event from its queue one-by-one, opening
// the REPL and waiting for user input before resuming the caller.
type DebugManager struct {
	eventQueue  chan *debugQueueItem
	once        sync.Once
	debugServer *DebugServer
}

type debugQueueItem struct {
	event       *DebugEvent
	resumeChan  chan struct{}
	interpreter *Interpreter
}

var globalDebugManager *DebugManager
var debugManagerOnce sync.Once

// GetDebugManager returns the global debug manager instance
func GetDebugManager() *DebugManager {
	debugManagerOnce.Do(func() {
		globalDebugManager = &DebugManager{
			eventQueue: make(chan *debugQueueItem, 100),
		}
		globalDebugManager.startProcessor()
	})
	return globalDebugManager
}

// SetDebugServer sets the HTTP debug server for HTTP debug mode
func (dm *DebugManager) SetDebugServer(server *DebugServer) {
	dm.debugServer = server
}

// startProcessor starts the background goroutine that processes debug events
func (dm *DebugManager) startProcessor() {
	go func() {
		for item := range dm.eventQueue {
			if item == nil {
				continue
			}

			// Check if we're in HTTP debug mode
			if dm.debugServer != nil {
				// HTTP debug mode: expose the event and wait for HTTP input
				dm.debugServer.SetEvent(item.event)
				// Wait for HTTP client to send input via POST
				<-dm.debugServer.GetInputChannel()
			} else {
				// Console debug mode: call the debug handler which opens REPL
				handler := item.interpreter.GetDebugHandler()
				if handler != nil {
					handler(item.event)
				}
			}

			// Signal the waiting caller to resume
			select {
			case item.resumeChan <- struct{}{}:
			default:
			}
		}
	}()
}

// Wait blocks until the user responds to the debug event.
// This is called synchronously by ExecuteScript when a breakpoint is hit.
func (dm *DebugManager) Wait(event *DebugEvent, interpreter *Interpreter) {
	resumeChan := make(chan struct{}, 1)
	item := &debugQueueItem{
		event:       event,
		resumeChan:  resumeChan,
		interpreter: interpreter,
	}
	// Queue the event
	dm.eventQueue <- item
	// Block until the event is processed and user responds
	<-resumeChan
}
