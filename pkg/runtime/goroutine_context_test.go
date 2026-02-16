package runtime

import (
	"fmt"
	"sync"
	"testing"
)

// TestGetGoroutineID tests goroutine ID extraction
func TestGetGoroutineID(t *testing.T) {
	id := GetGoroutineID()
	if id == 0 {
		t.Error("GetGoroutineID returned 0")
	}

	// Different goroutine should have different ID
	var otherId uint64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		otherId = GetGoroutineID()
	}()
	wg.Wait()

	if otherId == 0 {
		t.Error("goroutine ID is 0")
	}
	if id == otherId {
		t.Errorf("same ID for different goroutines: %d", id)
	}
}

// TestGetGoroutineIDConsistent tests same goroutine always returns same ID
func TestGetGoroutineIDConsistent(t *testing.T) {
	id1 := GetGoroutineID()
	id2 := GetGoroutineID()
	id3 := GetGoroutineID()

	if id1 != id2 || id2 != id3 {
		t.Errorf("inconsistent IDs: %d, %d, %d", id1, id2, id3)
	}
}

// TestRequestContextStorage tests setRequestContext/GetRequestContext
func TestRequestContextStorage(t *testing.T) {
	gid := uint64(12345)
	ctx := &RequestContext{
		Data: "test data",
	}

	setRequestContext(gid, ctx)

	retrieved, ok := GetRequestContext(gid)
	if !ok {
		t.Error("context not found after setting")
		return
	}

	if retrieved.Data != "test data" {
		t.Errorf("data mismatch: got %v, want 'test data'", retrieved.Data)
	}
}

// TestSetRequestContextWithData tests SetRequestContextWithData
func TestSetRequestContextWithData(t *testing.T) {
	gid := uint64(54321)
	ctx := &RequestContext{}
	data := "spawned data"

	SetRequestContextWithData(gid, ctx, data)

	retrieved, ok := GetRequestContext(gid)
	if !ok {
		t.Error("context not found")
		return
	}

	if retrieved.Data != data {
		t.Errorf("data = %v, want %q", retrieved.Data, data)
	}
}

// TestClearRequestContext tests clearing context
func TestClearRequestContext(t *testing.T) {
	gid := uint64(99999)
	ctx := &RequestContext{Data: "will be cleared"}

	setRequestContext(gid, ctx)

	// Verify it's there
	_, ok := GetRequestContext(gid)
	if !ok {
		t.Fatal("context not stored")
	}

	// Clear it
	ClearRequestContext(gid)

	// Verify it's gone
	_, ok = GetRequestContext(gid)
	if ok {
		t.Error("context still exists after clear")
	}
}

// TestRequestContextPerGoroutine tests each goroutine can have its own context
func TestRequestContextPerGoroutine(t *testing.T) {
	mainID := GetGoroutineID()
	mainCtx := &RequestContext{Data: "main"}
	setRequestContext(mainID, mainCtx)

	var otherID uint64
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		otherID = GetGoroutineID()
		ctx := &RequestContext{Data: "other"}
		setRequestContext(otherID, ctx)

		// In this goroutine, we should see our context
		retrieved, ok := GetRequestContext(otherID)
		if !ok || retrieved.Data != "other" {
			t.Error("goroutine cannot see its own context")
		}
	}()
	wg.Wait()

	// Main goroutine should still see its context
	retrieved, ok := GetRequestContext(mainID)
	if !ok || retrieved.Data != "main" {
		t.Error("main context was affected by other goroutine")
	}
}

// TestContextGetterStorage tests context getter registration
func TestContextGetterStorage(t *testing.T) {
	gid := uint64(11111)
	called := false
	getter := func() any {
		called = true
		return "context value"
	}

	SetContextGetter(gid, getter)

	retrieved, ok := GetContextGetter(gid)
	if !ok {
		t.Error("getter not found")
		return
	}

	result := retrieved()
	if !called {
		t.Error("getter was not called")
	}
	if result != "context value" {
		t.Errorf("getter returned %v, want 'context value'", result)
	}
}

// TestClearContextGetter tests clearing getter
func TestClearContextGetter(t *testing.T) {
	gid := uint64(22222)
	getter := func() any { return "data" }

	SetContextGetter(gid, getter)

	_, ok := GetContextGetter(gid)
	if !ok {
		t.Fatal("getter not stored")
	}

	ClearContextGetter(gid)

	_, ok = GetContextGetter(gid)
	if ok {
		t.Error("getter still exists after clear")
	}
}

// TestGetContextCallsGetter tests GetContext calls the getter
func TestGetContextCallsGetter(t *testing.T) {
	gid := uint64(33333)
	expectedVal := "from getter"
	getter := func() any {
		return expectedVal
	}

	SetContextGetter(gid, getter)

	result := GetContext(gid)
	if result != expectedVal {
		t.Errorf("GetContext returned %v, want %q", result, expectedVal)
	}
}

// TestGetContextNoGetter tests GetContext when no getter set
func TestGetContextNoGetter(t *testing.T) {
	gid := uint64(44444)
	// Don't set any getter

	result := GetContext(gid)
	if result != nil {
		t.Errorf("GetContext returned %v, want nil", result)
	}
}

// TestRequestContextFields tests RequestContext field initialization
func TestRequestContextFields(t *testing.T) {
	ctx := &RequestContext{
		Data:     "test",
		closed:   false,
		PathParams: map[string]any{"id": "123"},
	}

	if ctx.Data != "test" {
		t.Error("Data field incorrect")
	}
	if ctx.closed {
		t.Error("closed should be false")
	}
	if ctx.PathParams["id"] != "123" {
		t.Error("PathParams incorrect")
	}
}

// TestConcurrentContextAccess tests thread-safe context access
func TestConcurrentContextAccess(t *testing.T) {
	// Create multiple goroutines that each set/get their own context
	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			gid := GetGoroutineID()
			dataStr := fmt.Sprintf("data-%d", index)
			ctx := &RequestContext{Data: dataStr}

			setRequestContext(gid, ctx)

			// Retrieve and verify
			retrieved, ok := GetRequestContext(gid)
			if !ok {
				t.Errorf("goroutine %d: context not found", index)
				return
			}

			if retrieved.Data != dataStr {
				t.Errorf("goroutine %d: data mismatch: got %v, want %q",
					index, retrieved.Data, dataStr)
			}

			ClearRequestContext(gid)
		}(i)
	}

	wg.Wait()
}

// TestConcurrentGetterAccess tests thread-safe getter access
func TestConcurrentGetterAccess(t *testing.T) {
	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			gid := GetGoroutineID()
			expectedVal := fmt.Sprintf("value-%d", index)
			getter := func() any {
				return expectedVal
			}

			SetContextGetter(gid, getter)

			retrieved, ok := GetContextGetter(gid)
			if !ok {
				t.Errorf("goroutine %d: getter not found", index)
				return
			}

			result := retrieved()
			if result != expectedVal {
				t.Errorf("goroutine %d: got %v, want %q",
					index, result, expectedVal)
			}

			ClearContextGetter(gid)
		}(i)
	}

	wg.Wait()
}

// TestSetRequestContextOverwrite tests overwriting existing context
func TestSetRequestContextOverwrite(t *testing.T) {
	gid := uint64(55555)

	// Set initial context
	ctx1 := &RequestContext{Data: "first"}
	setRequestContext(gid, ctx1)

	retrieved, _ := GetRequestContext(gid)
	if retrieved.Data != "first" {
		t.Error("first context not set correctly")
	}

	// Overwrite with new context
	ctx2 := &RequestContext{Data: "second"}
	setRequestContext(gid, ctx2)

	retrieved, _ = GetRequestContext(gid)
	if retrieved.Data != "second" {
		t.Error("context not overwritten")
	}
}

// TestContextGetterOverwrite tests overwriting existing getter
func TestContextGetterOverwrite(t *testing.T) {
	gid := uint64(66666)

	getter1 := func() any { return "first" }
	SetContextGetter(gid, getter1)

	result1 := GetContext(gid)
	if result1 != "first" {
		t.Error("first getter not set correctly")
	}

	getter2 := func() any { return "second" }
	SetContextGetter(gid, getter2)

	result2 := GetContext(gid)
	if result2 != "second" {
		t.Error("getter not overwritten")
	}
}

// TestGetContextGetterMissing tests GetContextGetter for missing getter
func TestGetContextGetterMissing(t *testing.T) {
	gid := uint64(77777)
	// Don't set any getter

	_, ok := GetContextGetter(gid)
	if ok {
		t.Error("expected getter not found, but got one")
	}
}

// TestRequestContextGetMissing tests GetRequestContext for missing context
func TestRequestContextGetMissing(t *testing.T) {
	gid := uint64(88888)
	// Don't set any context

	_, ok := GetRequestContext(gid)
	if ok {
		t.Error("expected context not found, but got one")
	}
}

// TestMultipleContextsInSameGoroutine tests setting multiple contexts for same goroutine
func TestMultipleContextsInSameGoroutine(t *testing.T) {
	gid := GetGoroutineID()

	// Set context
	ctx1 := &RequestContext{Data: "ctx1"}
	setRequestContext(gid, ctx1)

	retrieved1, _ := GetRequestContext(gid)
	if retrieved1.Data != "ctx1" {
		t.Error("context 1 not set")
	}

	// Set different context (overwrites)
	ctx2 := &RequestContext{Data: "ctx2"}
	setRequestContext(gid, ctx2)

	retrieved2, _ := GetRequestContext(gid)
	if retrieved2.Data != "ctx2" {
		t.Error("context 2 not set")
	}

	// Clear it
	ClearRequestContext(gid)

	_, ok := GetRequestContext(gid)
	if ok {
		t.Error("context still exists after clear")
	}
}
