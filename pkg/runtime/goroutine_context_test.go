package runtime

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestGetGoroutineID_MainThread tests getting goroutine ID on main thread
func TestGetGoroutineID_MainThread(t *testing.T) {
	id := GetGoroutineID()

	if id == 0 {
		t.Errorf("Expected non-zero goroutine ID, got 0")
	}
}

// TestGetGoroutineID_Uniqueness tests that different goroutines have different IDs
func TestGetGoroutineID_Uniqueness(t *testing.T) {
	mainID := GetGoroutineID()
	var goroutineID uint64
	done := make(chan bool)

	go func() {
		goroutineID = GetGoroutineID()
		done <- true
	}()

	<-done

	if mainID == goroutineID {
		t.Errorf("Different goroutines should have different IDs: main=%d, goroutine=%d", mainID, goroutineID)
	}

	if goroutineID == 0 {
		t.Errorf("Goroutine ID should not be 0")
	}
}

// TestGetGoroutineID_Multiple tests multiple goroutines have unique IDs
func TestGetGoroutineID_Multiple(t *testing.T) {
	numGoroutines := 10
	ids := make(map[uint64]bool)
	var mutex sync.Mutex
	wg := sync.WaitGroup{}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := GetGoroutineID()

			mutex.Lock()
			ids[id] = true
			mutex.Unlock()
		}()
	}

	wg.Wait()

	if len(ids) != numGoroutines {
		t.Errorf("Expected %d unique IDs, got %d", numGoroutines, len(ids))
	}

	for id := range ids {
		if id == 0 {
			t.Errorf("Got zero ID in set")
		}
	}
}

// TestSetRequestContext_Basic tests storing request context
func TestSetRequestContext_Basic(t *testing.T) {
	gid := GetGoroutineID()
	ctx := &RequestContext{
		Data: "test_data",
	}

	setRequestContext(gid, ctx)

	retrieved, ok := GetRequestContext(gid)
	if !ok {
		t.Errorf("Context not found after setting")
	}

	if retrieved.Data != "test_data" {
		t.Errorf("Retrieved context has wrong data: %v", retrieved.Data)
	}

	// Cleanup
	ClearRequestContext(gid)
}

// TestSetRequestContext_Overwrite tests overwriting context
func TestSetRequestContext_Overwrite(t *testing.T) {
	gid := GetGoroutineID()

	ctx1 := &RequestContext{Data: "first"}
	setRequestContext(gid, ctx1)

	ctx2 := &RequestContext{Data: "second"}
	setRequestContext(gid, ctx2)

	retrieved, _ := GetRequestContext(gid)
	if retrieved.Data != "second" {
		t.Errorf("Expected second context, got %v", retrieved.Data)
	}

	ClearRequestContext(gid)
}

// TestGetRequestContext_NotFound tests getting nonexistent context
func TestGetRequestContext_NotFound(t *testing.T) {
	gid := uint64(999999999) // Unlikely to be a real goroutine ID

	_, ok := GetRequestContext(gid)
	if ok {
		t.Errorf("Should not find context for nonexistent goroutine")
	}
}

// TestClearRequestContext_Success tests clearing context
func TestClearRequestContext_Success(t *testing.T) {
	gid := GetGoroutineID()
	ctx := &RequestContext{Data: "data"}

	setRequestContext(gid, ctx)
	ClearRequestContext(gid)

	_, ok := GetRequestContext(gid)
	if ok {
		t.Errorf("Context should be cleared")
	}
}

// TestClearRequestContext_NonexistentKey tests clearing nonexistent context
func TestClearRequestContext_NonexistentKey(t *testing.T) {
	gid := uint64(999999999)

	// Should not panic
	ClearRequestContext(gid)
}

// TestSetRequestContextWithData_Basic tests setting context with data
func TestSetRequestContextWithData_Basic(t *testing.T) {
	gid := GetGoroutineID()
	ctx := &RequestContext{}
	data := map[string]any{"key": "value"}

	SetRequestContextWithData(gid, ctx, data)

	retrieved, ok := GetRequestContext(gid)
	if !ok {
		t.Errorf("Context not found")
	}

	if retrievedMap, ok := retrieved.Data.(map[string]any); ok {
		if retrievedMap["key"] != "value" {
			t.Errorf("Data not set correctly")
		}
	} else {
		t.Errorf("Data not a map")
	}

	ClearRequestContext(gid)
}

// TestSetContextGetter_Basic tests storing context getter
func TestSetContextGetter_Basic(t *testing.T) {
	gid := GetGoroutineID()

	getter := func() any {
		return "context_value"
	}

	SetContextGetter(gid, getter)

	retrieved, ok := GetContextGetter(gid)
	if !ok {
		t.Errorf("Context getter not found")
	}

	result := retrieved()
	if result != "context_value" {
		t.Errorf("Getter returned wrong value: %v", result)
	}

	ClearContextGetter(gid)
}

// TestGetContextGetter_NotFound tests getting nonexistent getter
func TestGetContextGetter_NotFound(t *testing.T) {
	gid := uint64(999999999)

	_, ok := GetContextGetter(gid)
	if ok {
		t.Errorf("Should not find getter for nonexistent goroutine")
	}
}

// TestClearContextGetter_Success tests clearing getter
func TestClearContextGetter_Success(t *testing.T) {
	gid := GetGoroutineID()

	SetContextGetter(gid, func() any { return "value" })
	ClearContextGetter(gid)

	_, ok := GetContextGetter(gid)
	if ok {
		t.Errorf("Getter should be cleared")
	}
}

// TestGetContext_WithGetter tests GetContext with available getter
func TestGetContext_WithGetter(t *testing.T) {
	gid := GetGoroutineID()

	setter := func() any {
		return "test_context"
	}

	SetContextGetter(gid, setter)

	result := GetContext(gid)
	if result != "test_context" {
		t.Errorf("GetContext returned wrong value: %v", result)
	}

	ClearContextGetter(gid)
}

// TestGetContext_WithoutGetter tests GetContext without available getter
func TestGetContext_WithoutGetter(t *testing.T) {
	gid := uint64(999999999)

	result := GetContext(gid)
	if result != nil {
		t.Errorf("GetContext should return nil when no getter available, got %v", result)
	}
}

// TestContextIsolation_TwoGoroutines tests that contexts are isolated between goroutines
func TestContextIsolation_TwoGoroutines(t *testing.T) {
	mainID := GetGoroutineID()
	var goroutineID uint64
	done := make(chan bool)

	// Set context in main goroutine
	mainCtx := &RequestContext{Data: "main_data"}
	setRequestContext(mainID, mainCtx)
	defer ClearRequestContext(mainID)

	// Set context in child goroutine
	go func() {
		goroutineID = GetGoroutineID()
		childCtx := &RequestContext{Data: "child_data"}
		setRequestContext(goroutineID, childCtx)
		defer ClearRequestContext(goroutineID)

		// Verify child sees its own context
		retrieved, ok := GetRequestContext(goroutineID)
		if !ok || retrieved.Data != "child_data" {
			t.Errorf("Child goroutine context incorrect")
		}

		done <- true
	}()

	<-done

	// Verify main still sees its context
	retrieved, ok := GetRequestContext(mainID)
	if !ok || retrieved.Data != "main_data" {
		t.Errorf("Main goroutine context corrupted")
	}
}

// TestContextIsolation_MultipleGoroutines tests isolation with many goroutines
func TestContextIsolation_MultipleGoroutines(t *testing.T) {
	numGoroutines := 10
	wg := sync.WaitGroup{}
	errors := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			gid := GetGoroutineID()
			data := fmt.Sprintf("data_%d", index)
			ctx := &RequestContext{Data: data}

			setRequestContext(gid, ctx)
			defer ClearRequestContext(gid)

			// Verify own context
			retrieved, ok := GetRequestContext(gid)
			if !ok {
				errors <- fmt.Sprintf("Goroutine %d: context not found", index)
				return
			}

			if retrieved.Data != data {
				errors <- fmt.Sprintf("Goroutine %d: wrong data", index)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for errMsg := range errors {
		t.Errorf(errMsg)
	}
}

// TestGetResponse_JSON tests GetResponse json helper
func TestGetResponse_JSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	rc := &RequestContext{
		Request: req,
		Writer:  recorder,
		closed:  false,
	}

	respObj := rc.GetResponse()

	if _, ok := respObj["json"]; !ok {
		t.Errorf("GetResponse should have 'json' method")
	}
}

// TestGetResponse_Text tests GetResponse text helper
func TestGetResponse_Text(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	rc := &RequestContext{
		Request: req,
		Writer:  recorder,
		closed:  false,
	}

	respObj := rc.GetResponse()

	if _, ok := respObj["text"]; !ok {
		t.Errorf("GetResponse should have 'text' method")
	}
}

// TestGetResponse_HTML tests GetResponse html helper
func TestGetResponse_HTML(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	rc := &RequestContext{
		Request: req,
		Writer:  recorder,
		closed:  false,
	}

	respObj := rc.GetResponse()

	if _, ok := respObj["html"]; !ok {
		t.Errorf("GetResponse should have 'html' method")
	}
}

// TestGetResponse_Error tests GetResponse error helper
func TestGetResponse_Error(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	rc := &RequestContext{
		Request: req,
		Writer:  recorder,
		closed:  false,
	}

	respObj := rc.GetResponse()

	if _, ok := respObj["error"]; !ok {
		t.Errorf("GetResponse should have 'error' method")
	}
}

// TestGetResponse_Redirect tests GetResponse redirect helper
func TestGetResponse_Redirect(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	rc := &RequestContext{
		Request: req,
		Writer:  recorder,
		closed:  false,
	}

	respObj := rc.GetResponse()

	if _, ok := respObj["redirect"]; !ok {
		t.Errorf("GetResponse should have 'redirect' method")
	}
}

// TestGetResponse_File tests GetResponse file helper
func TestGetResponse_File(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	rc := &RequestContext{
		Request: req,
		Writer:  recorder,
		closed:  false,
	}

	respObj := rc.GetResponse()

	if _, ok := respObj["file"]; !ok {
		t.Errorf("GetResponse should have 'file' method")
	}
}

// TestGetResponse_Response tests GetResponse generic response helper
func TestGetResponse_Response(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	rc := &RequestContext{
		Request: req,
		Writer:  recorder,
		closed:  false,
	}

	respObj := rc.GetResponse()

	if _, ok := respObj["response"]; !ok {
		t.Errorf("GetResponse should have 'response' method")
	}
}

// TestContextGetter_Concurrency tests context getters with concurrent access
func TestContextGetter_Concurrency(t *testing.T) {
	gid := GetGoroutineID()
	numOps := 100
	wg := sync.WaitGroup{}

	// Set getter
	SetContextGetter(gid, func() any { return "concurrent_value" })
	defer ClearContextGetter(gid)

	// Concurrent reads
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			getter, ok := GetContextGetter(gid)
			if !ok {
				t.Errorf("Getter not found")
				return
			}

			result := getter()
			if result != "concurrent_value" {
				t.Errorf("Got wrong value: %v", result)
			}
		}()
	}

	wg.Wait()
}

// TestRequestContext_BodyCaching tests that request body is cached
func TestRequestContext_BodyCaching(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("test body"))

	rc := &RequestContext{
		Request: req,
	}

	// First call should read and cache body
	result1 := rc.GetRequest()
	data1 := result1.(map[string]any)

	if data1["body"] != "test body" {
		t.Errorf("First read: expected 'test body', got %v", data1["body"])
	}

	// Second call should use cached body (since HTTP body can only be read once)
	result2 := rc.GetRequest()
	data2 := result2.(map[string]any)

	if data2["body"] != "test body" {
		t.Errorf("Second read should use cache, got %v", data2["body"])
	}
}

// TestRequestContext_MultiValueHeaders tests multi-value header parsing
func TestRequestContext_MultiValueHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/html")
	req.Header.Add("Accept", "application/json")

	rc := &RequestContext{
		Request: req,
	}

	result := rc.GetRequest()
	data := result.(map[string]any)
	headers := data["headers"].(map[string]any)

	acceptHeader := headers["Accept"]
	if arr, ok := acceptHeader.([]any); ok {
		if len(arr) != 2 {
			t.Errorf("Expected 2 Accept values, got %d", len(arr))
		}
	} else {
		t.Errorf("Expected Accept to be array for multi-value header")
	}
}
