package runtime

import (
	"container/heap"
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

// popForFiring takes a job off the heap the way fireDueJobs does at the start
// of a batch, leaving it mid-firing for the test.
func popForFiring(t *testing.T, id string) *scheduleJob {
	t.Helper()
	schedMu.Lock()
	defer schedMu.Unlock()
	job, ok := schedIndex[id]
	if !ok || job.index < 0 {
		t.Fatalf("job %q is not on the heap", id)
	}
	heap.Remove(&schedHeap, job.index)
	return job
}

func onHeap(id string) int {
	schedMu.Lock()
	defer schedMu.Unlock()
	n := 0
	for _, j := range schedHeap {
		if j.id == id {
			n++
		}
	}
	return n
}

// A job is off the heap while its batch fires. unschedule() used to miss it
// there - returning false - and the scheduler then put it back, so it kept
// running and its record reappeared.
func TestUnscheduleWhileFiring(t *testing.T) {
	registerOnce.Do(RegisterBuiltins)
	const id = "unschedule_while_firing"

	if _, err := builtinSchedule(nil, map[string]any{"0": "1h", "1": "job.du", "id": id}); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	job := popForFiring(t, id)

	found, err := builtinUnschedule(nil, map[string]any{"0": id})
	if err != nil || found != true {
		t.Fatalf("unschedule of a firing job = %v, %v; want true", found, err)
	}

	finishFiring(job, 42)

	if n := onHeap(id); n != 0 {
		t.Errorf("unscheduled job is back on the heap %d time(s)", n)
	}
	schedMu.Lock()
	_, indexed := schedIndex[id]
	schedMu.Unlock()
	if indexed {
		t.Error("unscheduled job is back in schedIndex")
	}
	if rec, _ := GetDatastore("duso_schedule", nil).Get(id); rec != nil {
		t.Errorf("unscheduled job's record came back: %v", rec)
	}
}

// schedule() with the id of a job that is firing replaces it: only the new job
// may remain, and the record must be the new job's.
func TestRescheduleWhileFiring(t *testing.T) {
	registerOnce.Do(RegisterBuiltins)
	const id = "reschedule_while_firing"

	if _, err := builtinSchedule(nil, map[string]any{"0": "1h", "1": "old.du", "id": id}); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	old := popForFiring(t, id)

	if _, err := builtinSchedule(nil, map[string]any{"0": "2h", "1": "new.du", "id": id}); err != nil {
		t.Fatalf("re-schedule: %v", err)
	}

	finishFiring(old, 42)

	if n := onHeap(id); n != 1 {
		t.Fatalf("job %q is on the heap %d times, want once", id, n)
	}
	schedMu.Lock()
	live := schedIndex[id]
	schedMu.Unlock()
	if live == old || live == nil || live.scriptPath == old.scriptPath {
		t.Errorf("schedIndex should hold the replacement job, got %+v", live)
	}
	rec, _ := GetDatastore("duso_schedule", nil).Get(id)
	if r, ok := rec.(map[string]any); !ok || r["script"] != live.scriptPath {
		t.Errorf("record should be the replacement job's, got %v", rec)
	}

	if _, err := builtinUnschedule(nil, map[string]any{"0": id}); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
}
