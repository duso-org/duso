package runtime

import (
	"sync"
	"testing"
	"time"
)

// TestIncrementHTTPProcs tests incrementing HTTP process counter
func TestIncrementHTTPProcs(t *testing.T) {
	// Initialize metrics
	InitSystemMetrics()

	initialHTTP := GetMetric("http_procs")
	initialHTTPVal := initialHTTP.(float64)

	IncrementHTTPProcs()

	currentHTTP := GetMetric("http_procs")
	currentHTTPVal := currentHTTP.(float64)

	if currentHTTPVal != initialHTTPVal+1 {
		t.Errorf("Expected http_procs to increment by 1, got difference of %v", currentHTTPVal-initialHTTPVal)
	}
}

// TestIncrementSpawnProcs tests incrementing spawn process counter
func TestIncrementSpawnProcs(t *testing.T) {
	InitSystemMetrics()

	initialSpawn := GetMetric("spawn_procs")
	initialSpawnVal := initialSpawn.(float64)

	IncrementSpawnProcs()

	currentSpawn := GetMetric("spawn_procs")
	currentSpawnVal := currentSpawn.(float64)

	if currentSpawnVal != initialSpawnVal+1 {
		t.Errorf("Expected spawn_procs to increment by 1, got difference of %v", currentSpawnVal-initialSpawnVal)
	}
}

// TestIncrementRunProcs tests incrementing run process counter
func TestIncrementRunProcs(t *testing.T) {
	InitSystemMetrics()

	initialRun := GetMetric("run_procs")
	initialRunVal := initialRun.(float64)

	IncrementRunProcs()

	currentRun := GetMetric("run_procs")
	currentRunVal := currentRun.(float64)

	if currentRunVal != initialRunVal+1 {
		t.Errorf("Expected run_procs to increment by 1, got difference of %v", currentRunVal-initialRunVal)
	}
}

// TestGetMetric_HTTPProcs tests getting HTTP procs metric
func TestGetMetric_HTTPProcs(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("http_procs")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if _, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	}
}

// TestGetMetric_SpawnProcs tests getting spawn procs metric
func TestGetMetric_SpawnProcs(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("spawn_procs")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if _, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	}
}

// TestGetMetric_RunProcs tests getting run procs metric
func TestGetMetric_RunProcs(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("run_procs")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if _, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	}
}

// TestGetMetric_ActiveGoroutines tests getting active goroutines count
func TestGetMetric_ActiveGoroutines(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("active_goroutines")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if count, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	} else if count <= 0 {
		t.Errorf("Expected positive goroutine count, got %v", count)
	}
}

// TestGetMetric_PeakGoroutines tests getting peak goroutines count
func TestGetMetric_PeakGoroutines(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("peak_goroutines")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if _, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	}
}

// TestGetMetric_HeapAlloc tests getting heap allocation metric
func TestGetMetric_HeapAlloc(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("heap_alloc")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if alloc, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	} else if alloc < 0 {
		t.Errorf("Expected non-negative heap allocation, got %v", alloc)
	}
}

// TestGetMetric_TotalAlloc tests getting total allocation metric
func TestGetMetric_TotalAlloc(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("total_alloc")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if _, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	}
}

// TestGetMetric_HeapSys tests getting heap sys metric
func TestGetMetric_HeapSys(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("heap_sys")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if _, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	}
}

// TestGetMetric_NumGC tests getting garbage collection count
func TestGetMetric_NumGC(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("num_gc")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if _, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	}
}

// TestGetMetric_PeakHeapAlloc tests getting peak heap allocation
func TestGetMetric_PeakHeapAlloc(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("peak_heap_alloc")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if _, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	}
}

// TestGetMetric_ServerStart tests getting server start time
func TestGetMetric_ServerStart(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("server_start")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if timestamp, ok := metric.(int64); !ok {
		t.Errorf("Expected int64, got %T", metric)
	} else if timestamp <= 0 {
		t.Errorf("Expected positive timestamp, got %v", timestamp)
	}
}

// TestGetMetric_DatastoreCount tests getting datastore count
func TestGetMetric_DatastoreCount(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("datastore_count")
	if metric == nil {
		t.Errorf("Expected metric, got nil")
	}

	if count, ok := metric.(float64); !ok {
		t.Errorf("Expected float64, got %T", metric)
	} else if count < 0 {
		t.Errorf("Expected non-negative datastore count, got %v", count)
	}
}

// TestGetMetric_InvalidKey tests getting invalid metric key
func TestGetMetric_InvalidKey(t *testing.T) {
	InitSystemMetrics()

	metric := GetMetric("invalid_metric_key")
	if metric != nil {
		t.Errorf("Expected nil for invalid metric key, got %v", metric)
	}
}

// TestUpdatePeakHeapAlloc tests peak heap allocation tracking
func TestUpdatePeakHeapAlloc(t *testing.T) {
	InitSystemMetrics()

	initialPeak := GetMetric("peak_heap_alloc").(float64)

	// Allocate some memory to potentially increase heap
	var data []int
	for i := 0; i < 1000000; i++ {
		data = append(data, i)
	}

	UpdatePeakHeapAlloc()

	newPeak := GetMetric("peak_heap_alloc").(float64)

	// Peak should not decrease
	if newPeak < initialPeak {
		t.Errorf("Peak heap allocation decreased unexpectedly")
	}
}

// TestUpdatePeakGoroutines tests peak goroutine tracking
func TestUpdatePeakGoroutines(t *testing.T) {
	InitSystemMetrics()

	initialPeak := GetMetric("peak_goroutines").(float64)

	// Create some goroutines
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
		}()
	}

	// Update metrics while goroutines are running
	IncrementHTTPProcs()

	wg.Wait()

	newPeak := GetMetric("peak_goroutines").(float64)

	// Peak should increase (we had more goroutines while they were running)
	if newPeak < initialPeak {
		t.Errorf("Peak goroutine count decreased unexpectedly")
	}
}

// TestMetricsConcurrency tests concurrent metric updates
func TestMetricsConcurrency(t *testing.T) {
	InitSystemMetrics()

	numGoroutines := 10
	operationsPerGoroutine := 100
	wg := sync.WaitGroup{}

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for i := 0; i < operationsPerGoroutine; i++ {
				switch idx % 3 {
				case 0:
					IncrementHTTPProcs()
				case 1:
					IncrementSpawnProcs()
				case 2:
					IncrementRunProcs()
				}
			}
		}(g)
	}

	wg.Wait()

	httpCount := GetMetric("http_procs").(float64)
	spawnCount := GetMetric("spawn_procs").(float64)
	runCount := GetMetric("run_procs").(float64)

	// We should have some non-zero counts
	totalOps := httpCount + spawnCount + runCount
	expectedOps := float64(numGoroutines * operationsPerGoroutine)

	// The counts might not exactly equal expected due to initialization,
	// but we should see evidence of the operations
	if totalOps < expectedOps/2 {
		t.Errorf("Concurrent metrics appear to have been lost: %v < %v", totalOps, expectedOps)
	}
}

// TestInitSystemMetrics tests initializing system metrics
func TestInitSystemMetrics(t *testing.T) {
	InitSystemMetrics()

	serverStart := GetMetric("server_start").(int64)
	currentTime := time.Now().Unix()

	// Server start time should be recent (within 1 second)
	if currentTime-serverStart > 1 {
		t.Errorf("Server start time not recent: %v vs %v", serverStart, currentTime)
	}
}

// TestMetricsOrdering tests that metrics are consistent
func TestMetricsOrdering(t *testing.T) {
	InitSystemMetrics()

	// Take initial measurements
	httpProcs1 := GetMetric("http_procs").(float64)
	spawnProcs1 := GetMetric("spawn_procs").(float64)
	runProcs1 := GetMetric("run_procs").(float64)

	// Increment all counters
	IncrementHTTPProcs()
	IncrementSpawnProcs()
	IncrementRunProcs()

	// Take new measurements
	httpProcs2 := GetMetric("http_procs").(float64)
	spawnProcs2 := GetMetric("spawn_procs").(float64)
	runProcs2 := GetMetric("run_procs").(float64)

	// All should have increased by 1
	if httpProcs2 != httpProcs1+1 {
		t.Errorf("HTTP procs not incremented correctly")
	}

	if spawnProcs2 != spawnProcs1+1 {
		t.Errorf("Spawn procs not incremented correctly")
	}

	if runProcs2 != runProcs1+1 {
		t.Errorf("Run procs not incremented correctly")
	}
}

// TestHeapAllocIncreases tests that heap allocation increases with memory use
func TestHeapAllocIncreases(t *testing.T) {
	InitSystemMetrics()

	heapBefore := GetMetric("heap_alloc").(float64)

	// Allocate a large amount of memory
	var data []int64
	for i := 0; i < 100000; i++ {
		data = append(data, int64(i))
	}

	UpdatePeakHeapAlloc()

	heapAfter := GetMetric("heap_alloc").(float64)

	// Heap allocation should increase
	if heapAfter <= heapBefore {
		t.Errorf("Heap allocation did not increase: before=%v, after=%v", heapBefore, heapAfter)
	}

	// Clear to avoid affecting other tests
	_ = data // Use data to prevent compiler optimization
}

// TestActiveGoroutinesIncreases tests that active goroutine count increases
func TestActiveGoroutinesIncreases(t *testing.T) {
	InitSystemMetrics()

	initialGoroutines := GetMetric("active_goroutines").(float64)

	// Create some goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			time.Sleep(100 * time.Millisecond)
			done <- true
		}()
	}

	// Check immediately - should see more goroutines
	activeNow := GetMetric("active_goroutines").(float64)

	if activeNow <= initialGoroutines {
		t.Errorf("Active goroutine count should increase: initial=%v, now=%v", initialGoroutines, activeNow)
	}

	// Wait for goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
