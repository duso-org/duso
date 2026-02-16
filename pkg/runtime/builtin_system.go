package runtime

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// System functions

// builtinExit stops execution and returns values to host
func builtinExit(evaluator *Evaluator, args map[string]any) (any, error) {
	// Collect all arguments as return values
	values := make([]any, 0)
	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		if val, ok := args[key]; ok {
			// Deep copy to isolate return values from parent scope
			values = append(values, DeepCopyAny(val))
		} else {
			break
		}
	}

	return nil, &ExitExecution{Values: values}
}

// builtinSleep pauses execution for the specified duration in seconds (default: 1)
func builtinSleep(evaluator *Evaluator, args map[string]any) (any, error) {
	seconds := 1.0 // Default to 1 second
	if arg, ok := args["0"]; ok {
		num, ok := arg.(float64)
		if !ok {
			return nil, fmt.Errorf("sleep() requires a number (seconds)")
		}
		if num < 0 {
			return nil, fmt.Errorf("sleep() duration cannot be negative")
		}
		seconds = num
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
	return nil, nil
}

// builtinUUID generates a UUID v7 (RFC 9562)
// UUID v7 is time-sorted with 48-bit Unix timestamp in milliseconds followed by random data
func builtinUUID(evaluator *Evaluator, args map[string]any) (any, error) {
	buf := make([]byte, 16)

	// 48-bit timestamp (Unix epoch in milliseconds)
	binary.BigEndian.PutUint64(buf[0:8], uint64(time.Now().UnixMilli()))

	// Truncate timestamp to 6 bytes, shifting because PutUint64 writes 8 bytes
	copy(buf[0:6], buf[2:8])

	// 10 bytes random data
	if _, err := rand.Read(buf[6:16]); err != nil {
		return nil, fmt.Errorf("uuid() failed to generate random bytes: %v", err)
	}

	// Version 7: set version bits to 0111 in the 7th byte
	buf[6] = (buf[6] & 0x0f) | 0x70

	// Variant: set variant bits to 10 in the 9th byte
	buf[8] = (buf[8] & 0x3f) | 0x80

	// Format as UUID string: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16]), nil
}

