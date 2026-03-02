package runtime

import (
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// initIOQueueing registers the datastore queue appender callback with the script package.
// This is called by RegisterBuiltins() to enable I/O routing functionality.
func initIOQueueing() {
	script.SetDatastoreQueueAppender(appendToIOQueue)
}

// appendToIOQueue appends an I/O event to a datastore queue.
// Each event is an object with: {pid = number, eventType = value}
// Example: {pid = 123, out = "message"} or {pid = 123, err = "error text"}
func appendToIOQueue(datastore, queue, eventType string, data any, pid int) error {
	// Get the datastore
	ds := GetDatastore(datastore, nil)
	if ds == nil {
		return fmt.Errorf("datastore '%s' not found", datastore)
	}

	// Get the current queue (should be an array)
	currentValue, _ := ds.Get(queue)
	var queueArray []any

	// Convert existing value to array (should be empty on first call)
	if currentValue != nil {
		if arr, ok := currentValue.([]any); ok {
			queueArray = arr
		} else {
			return fmt.Errorf("queue '%s' is not an array", queue)
		}
	}

	// Create the event object
	event := map[string]any{
		"pid":       float64(pid), // Convert to float64 for consistency with Duso numbers
		eventType:   data,
	}

	// Append to queue
	queueArray = append(queueArray, event)

	// Store back in datastore (deep copy happens automatically)
	ds.Set(queue, queueArray)

	return nil
}
