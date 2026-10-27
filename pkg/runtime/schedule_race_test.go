package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// TestScheduleDueJobDoesNotRace covers a job that is due the moment it is
// registered. schedule() used to build the job's datastore record after
// putting the job on the heap, while the scheduler goroutine was already
// firing it and writing lastPID and nextFire. Run with -race.
func TestScheduleDueJobDoesNotRace(t *testing.T) {
	registerOnce.Do(RegisterBuiltins)

	// Set once and left in place, as the CLI does: scheduled runs keep reading
	// it in the background after this test returns.
	interp := script.NewInterpreter()
	interp.ScriptLoader = os.ReadFile
	SetInterpreter(interp)

	path := filepath.Join(t.TempDir(), "job.du")
	if err := os.WriteFile(path, []byte("x = 1\n"), 0644); err != nil {
		t.Fatalf("test setup: %v", err)
	}

	var ids []string
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("race_due_%d", i)
		// A one-shot in the past fires at once
		if _, err := builtinSchedule(nil, map[string]any{"0": "2020-01-01 00:00", "1": path, "id": id}); err != nil {
			t.Fatalf("schedule one-shot: %v", err)
		}
		// An anchored recurring job in the past fast-forwards and is rescheduled after each firing
		id = fmt.Sprintf("race_recur_%d", i)
		ids = append(ids, id)
		if _, err := builtinSchedule(nil, map[string]any{"0": "00:00 1s", "1": path, "id": id}); err != nil {
			t.Fatalf("schedule recurring: %v", err)
		}
	}

	time.Sleep(1500 * time.Millisecond)

	for _, id := range ids {
		if _, err := builtinUnschedule(nil, map[string]any{"0": id}); err != nil {
			t.Fatalf("unschedule %s: %v", id, err)
		}
	}
}
