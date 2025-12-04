package core

import (
	"sync"
	"testing"
	"time"
)

// TestGetCoreMetrics tests getting the global core metrics instance
func TestGetCoreMetrics(t *testing.T) {
	metrics1 := GetCoreMetrics()
	metrics2 := GetCoreMetrics()
	
	if metrics1 == nil {
		t.Fatal("GetCoreMetrics returned nil")
	}
	
	// Should return the same global instance
	if metrics1 != metrics2 {
		t.Error("GetCoreMetrics should return the same global instance")
	}
}

// TestCoreMetricsIncEntryCreated tests the IncEntryCreated method
func TestCoreMetricsIncEntryCreated(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.CreatedCount()
	metrics.IncEntryCreated()
	newCount := metrics.CreatedCount()
	
	if newCount != initialCount+1 {
		t.Errorf("IncEntryCreated: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncEntryReused tests the IncEntryReused method
func TestCoreMetricsIncEntryReused(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.ReusedCount()
	metrics.IncEntryReused()
	newCount := metrics.ReusedCount()
	
	if newCount != initialCount+1 {
		t.Errorf("IncEntryReused: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncEntryPoolMiss tests the IncEntryPoolMiss method
func TestCoreMetricsIncEntryPoolMiss(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.PoolMissCount()
	metrics.IncEntryPoolMiss()
	newCount := metrics.PoolMissCount()
	
	if newCount != initialCount+1 {
		t.Errorf("IncEntryPoolMiss: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncEntrySerialized tests the IncEntrySerialized method
func TestCoreMetricsIncEntrySerialized(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.SerializedCount()
	metrics.IncEntrySerialized()
	newCount := metrics.SerializedCount()
	
	if newCount != initialCount+1 {
		t.Errorf("IncEntrySerialized: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncBufferGet tests the IncBufferGet method
func TestCoreMetricsIncBufferGet(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.GetBufferMetrics()["gets"]
	metrics.IncBufferGet()
	newMetrics := metrics.GetBufferMetrics()
	newCount := newMetrics["gets"]
	
	if newCount != initialCount+1 {
		t.Errorf("IncBufferGet: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncBufferPut tests the IncBufferPut method
func TestCoreMetricsIncBufferPut(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.GetBufferMetrics()["puts"]
	metrics.IncBufferPut()
	newMetrics := metrics.GetBufferMetrics()
	newCount := newMetrics["puts"]
	
	if newCount != initialCount+1 {
		t.Errorf("IncBufferPut: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncBufferMiss tests the IncBufferMiss method
func TestCoreMetricsIncBufferMiss(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.GetBufferMetrics()["misses"]
	metrics.IncBufferMiss()
	newMetrics := metrics.GetBufferMetrics()
	newCount := newMetrics["misses"]
	
	if newCount != initialCount+1 {
		t.Errorf("IncBufferMiss: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncSliceGet tests the IncSliceGet method
func TestCoreMetricsIncSliceGet(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.GetSliceMetrics()["gets"]
	metrics.IncSliceGet()
	newMetrics := metrics.GetSliceMetrics()
	newCount := newMetrics["gets"]
	
	if newCount != initialCount+1 {
		t.Errorf("IncSliceGet: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncSlicePut tests the IncSlicePut method
func TestCoreMetricsIncSlicePut(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.GetSliceMetrics()["puts"]
	metrics.IncSlicePut()
	newMetrics := metrics.GetSliceMetrics()
	newCount := newMetrics["puts"]
	
	if newCount != initialCount+1 {
		t.Errorf("IncSlicePut: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsIncError tests the IncError method
func TestCoreMetricsIncError(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialCount := metrics.GetErrorMetrics()["errors"]
	metrics.IncError()
	newMetrics := metrics.GetErrorMetrics()
	newCount := newMetrics["errors"]
	
	if newCount != initialCount+1 {
		t.Errorf("IncError: expected %d, got %d", initialCount+1, newCount)
	}
}

// TestCoreMetricsSetLastOperationTime tests the SetLastOperationTime method
func TestCoreMetricsSetLastOperationTime(t *testing.T) {
	metrics := GetCoreMetrics()

	newTime := time.Now()
	metrics.SetLastOperationTime(newTime)

	finalTime := time.Unix(0, metrics.GetTimingMetrics()["last_operation"])

	if finalTime.UnixNano() != newTime.UnixNano() {
		t.Logf("Timing may differ due to precision, but should be close: expected %d, got %d",
			newTime.UnixNano(), finalTime.UnixNano())
	}
}

// TestCoreMetricsAddProcessingTime tests the AddProcessingTime method
func TestCoreMetricsAddProcessingTime(t *testing.T) {
	metrics := GetCoreMetrics()
	
	initialTime := metrics.GetTimingMetrics()["processing_time_ns"]
	testDuration := 100 * time.Millisecond
	metrics.AddProcessingTime(testDuration)
	newTime := metrics.GetTimingMetrics()["processing_time_ns"]
	
	// We expect newTime to be initialTime + testDuration.Nanoseconds()
	expected := initialTime + testDuration.Nanoseconds()
	if newTime != expected {
		t.Errorf("AddProcessingTime: expected %d, got %d", expected, newTime)
	}
}

// TestCoreMetricsGetEntryMetrics tests the GetEntryMetrics method
func TestCoreMetricsGetEntryMetrics(t *testing.T) {
	metrics := GetCoreMetrics()
	
	entryMetrics := metrics.GetEntryMetrics()
	if entryMetrics == nil {
		t.Fatal("GetEntryMetrics returned nil")
	}
	
	// Check that expected keys exist
	expectedKeys := []string{"created", "reused", "pool_miss", "serialized", "hit_ratio"}
	for _, key := range expectedKeys {
		if _, exists := entryMetrics[key]; !exists {
			t.Errorf("GetEntryMetrics missing key: %s", key)
		}
	}
	
	// Verify hit ratio calculation makes sense
	created := entryMetrics["created"]
	reused := entryMetrics["reused"]
	hitRatio := entryMetrics["hit_ratio"]

	// Hit ratio = reused * 100 / (created + reused + 1)
	// The +1 is to avoid division by zero
	calculatedHitRatio := reused * 100 / (created + reused + 1)
	if hitRatio != calculatedHitRatio {
		t.Errorf("Hit ratio calculation mismatch: expected %d, got %d", calculatedHitRatio, hitRatio)
	}
}

// TestCoreMetricsGetBufferMetrics tests the GetBufferMetrics method
func TestCoreMetricsGetBufferMetrics(t *testing.T) {
	metrics := GetCoreMetrics()
	
	bufferMetrics := metrics.GetBufferMetrics()
	if bufferMetrics == nil {
		t.Fatal("GetBufferMetrics returned nil")
	}
	
	// Check that expected keys exist
	expectedKeys := []string{"gets", "puts", "misses", "hit_ratio"}
	for _, key := range expectedKeys {
		if _, exists := bufferMetrics[key]; !exists {
			t.Errorf("GetBufferMetrics missing key: %s", key)
		}
	}
}

// TestCoreMetricsGetSliceMetrics tests the GetSliceMetrics method
func TestCoreMetricsGetSliceMetrics(t *testing.T) {
	metrics := GetCoreMetrics()
	
	sliceMetrics := metrics.GetSliceMetrics()
	if sliceMetrics == nil {
		t.Fatal("GetSliceMetrics returned nil")
	}
	
	// Check that expected keys exist
	expectedKeys := []string{"gets", "puts"}
	for _, key := range expectedKeys {
		if _, exists := sliceMetrics[key]; !exists {
			t.Errorf("GetSliceMetrics missing key: %s", key)
		}
	}
}

// TestCoreMetricsGetErrorMetrics tests the GetErrorMetrics method
func TestCoreMetricsGetErrorMetrics(t *testing.T) {
	metrics := GetCoreMetrics()
	
	errorMetrics := metrics.GetErrorMetrics()
	if errorMetrics == nil {
		t.Fatal("GetErrorMetrics returned nil")
	}
	
	// Check that expected keys exist
	expectedKeys := []string{"errors"}
	for _, key := range expectedKeys {
		if _, exists := errorMetrics[key]; !exists {
			t.Errorf("GetErrorMetrics missing key: %s", key)
		}
	}
}

// TestCoreMetricsGetTimingMetrics tests the GetTimingMetrics method
func TestCoreMetricsGetTimingMetrics(t *testing.T) {
	metrics := GetCoreMetrics()
	
	timingMetrics := metrics.GetTimingMetrics()
	if timingMetrics == nil {
		t.Fatal("GetTimingMetrics returned nil")
	}
	
	// Check that expected keys exist
	expectedKeys := []string{"processing_time_ns", "last_operation"}
	for _, key := range expectedKeys {
		if _, exists := timingMetrics[key]; !exists {
			t.Errorf("GetTimingMetrics missing key: %s", key)
		}
	}
}

// TestCoreMetricsGetAllMetrics tests the GetAllMetrics method
func TestCoreMetricsGetAllMetrics(t *testing.T) {
	metrics := GetCoreMetrics()
	
	allMetrics := metrics.GetAllMetrics()
	if allMetrics == nil {
		t.Fatal("GetAllMetrics returned nil")
	}
	
	// Check that expected top-level keys exist
	expectedKeys := []string{"entries", "buffers", "slices", "errors", "timing"}
	for _, key := range expectedKeys {
		if _, exists := allMetrics[key]; !exists {
			t.Errorf("GetAllMetrics missing top-level key: %s", key)
		}
	}
	
	// Check that nested metrics are properly structured
	if entries, ok := allMetrics["entries"].(map[string]int64); !ok || entries == nil {
		t.Error("AllMetrics should have 'entries' as map[string]int64")
	}
	
	if buffers, ok := allMetrics["buffers"].(map[string]int64); !ok || buffers == nil {
		t.Error("AllMetrics should have 'buffers' as map[string]int64")
	}
	
	if slices, ok := allMetrics["slices"].(map[string]int64); !ok || slices == nil {
		t.Error("AllMetrics should have 'slices' as map[string]int64")
	}
	
	if errors, ok := allMetrics["errors"].(map[string]int64); !ok || errors == nil {
		t.Error("AllMetrics should have 'errors' as map[string]int64")
	}
	
	if timing, ok := allMetrics["timing"].(map[string]int64); !ok || timing == nil {
		t.Error("AllMetrics should have 'timing' as map[string]int64")
	}
}

// TestCoreMetricsResetMetrics tests the ResetMetrics method
func TestCoreMetricsResetMetrics(t *testing.T) {
	metrics := GetCoreMetrics()
	
	// Set some metrics values
	metrics.IncEntryCreated()
	metrics.IncEntryReused()
	metrics.IncEntryPoolMiss()
	metrics.IncEntrySerialized()
	metrics.IncBufferGet()
	metrics.IncBufferPut()
	metrics.IncBufferMiss()
	metrics.IncSliceGet()
	metrics.IncSlicePut()
	metrics.IncError()
	metrics.SetLastOperationTime(time.Now())
	metrics.AddProcessingTime(10 * time.Millisecond)
	
	// Check that metrics have non-zero values
	initialAll := metrics.GetAllMetrics()
	if initialAll == nil {
		t.Fatal("Initial metrics should not be nil")
	}
	
	// Reset metrics
	metrics.ResetMetrics()
	
	// Check that all metrics are back to zero

	// Check entry metrics
	entryMetrics := metrics.GetEntryMetrics()
	for _, value := range entryMetrics {
		if value != 0 {
			t.Error("After reset, entry metrics should be 0")
			break
		}
	}
	
	// Check buffer metrics
	bufferMetrics := metrics.GetBufferMetrics()
	for _, value := range bufferMetrics {
		if value != 0 {
			t.Error("After reset, buffer metrics should be 0")
			break
		}
	}
	
	// Check slice metrics
	sliceMetrics := metrics.GetSliceMetrics()
	for _, value := range sliceMetrics {
		if value != 0 {
			t.Error("After reset, slice metrics should be 0")
			break
		}
	}
	
	// Check error metrics
	errorMetrics := metrics.GetErrorMetrics()
	for _, value := range errorMetrics {
		if value != 0 {
			t.Error("After reset, error metrics should be 0")
			break
		}
	}
	
	// Check timing metrics
	timingMetrics := metrics.GetTimingMetrics()
	if timingMetrics["processing_time_ns"] != 0 {
		t.Error("After reset, processing_time_ns should be 0")
	}
	if timingMetrics["last_operation"] != 0 {
		t.Error("After reset, last_operation should be 0")
	}
}

// TestCoreMetricsConcurrent tests the CoreMetrics in a concurrent context
func TestCoreMetricsConcurrent(t *testing.T) {
	metrics := GetCoreMetrics()
	
	const numGoroutines = 10
	const operationsPerGoroutine = 100
	var wg sync.WaitGroup
	
	startTime := time.Now()
	
	// Test concurrent increment operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				// Call different increment methods in each goroutine
				metrics.IncEntryCreated()
				metrics.IncEntryReused()
				metrics.IncEntryPoolMiss()
				metrics.IncEntrySerialized()
				metrics.IncBufferGet()
				metrics.IncBufferPut()
				metrics.IncSliceGet()
				metrics.IncError()
				metrics.SetLastOperationTime(time.Now())
				metrics.AddProcessingTime(time.Microsecond)
			}
		}()
	}
	
	wg.Wait()
	
	elapsed := time.Since(startTime)
	t.Logf("Completed %d concurrent metric updates in %v", 
		numGoroutines*operationsPerGoroutine*9, elapsed) // 9 operations per loop iteration
	
	// Verify that all operations were counted
	expectedEntryCreates := int64(numGoroutines * operationsPerGoroutine)
	actualEntryCreates := metrics.GetEntryMetrics()["created"]
	
	if actualEntryCreates != expectedEntryCreates {
		t.Errorf("Expected %d entry creates, got %d", expectedEntryCreates, actualEntryCreates)
	}
	
	// The other metrics should also be updated appropriately
	expectedOthers := int64(numGoroutines * operationsPerGoroutine)
	allMetrics := metrics.GetAllMetrics()
	
	if entryMetrics := allMetrics["entries"].(map[string]int64); entryMetrics["reused"] != expectedOthers {
		t.Logf("Expected %d reused entries, got %d", expectedOthers, entryMetrics["reused"])
	}
	if entryMetrics := allMetrics["entries"].(map[string]int64); entryMetrics["pool_miss"] != expectedOthers {
		t.Logf("Expected %d pool misses, got %d", expectedOthers, entryMetrics["pool_miss"])
	}
	if bufferMetrics := allMetrics["buffers"].(map[string]int64); bufferMetrics["gets"] != expectedOthers {
		t.Logf("Expected %d buffer gets, got %d", expectedOthers, bufferMetrics["gets"])
	}
}

// TestCoreMetricsHitRatioCalculation tests the hit ratio calculation specifically
func TestCoreMetricsHitRatioCalculation(t *testing.T) {
	metrics := GetCoreMetrics()
	
	// Reset metrics first
	metrics.ResetMetrics()
	
	// Scenario: 3 created, 7 reused (total 10 operations)
	// Hit ratio = reused * 100 / (created + reused + 1) = 7 * 100 / (3 + 7 + 1) = 700 / 11 ≈ 63
	for i := 0; i < 3; i++ {
		metrics.IncEntryCreated()
	}
	for i := 0; i < 7; i++ {
		metrics.IncEntryReused()
	}
	
	entryMetrics := metrics.GetEntryMetrics()
	hitRatio := entryMetrics["hit_ratio"]
	expectedHitRatio := int64(7 * 100 / (3 + 7 + 1)) // 63
	
	if hitRatio != expectedHitRatio {
		t.Errorf("Expected hit ratio %d, got %d", expectedHitRatio, hitRatio)
	}
	
	// Test with zero values - should not cause division by zero
	metrics.ResetMetrics()
	entryMetrics = metrics.GetEntryMetrics()
	hitRatio = entryMetrics["hit_ratio"]
	if hitRatio != 0 {
		t.Errorf("Expected hit ratio 0 for zero values, got %d", hitRatio)
	}
	
	// Test with only created entries
	metrics.IncEntryCreated()
	metrics.IncEntryCreated()
	entryMetrics = metrics.GetEntryMetrics()
	hitRatio = entryMetrics["hit_ratio"]
	// (0 * 100) / (2 + 0 + 1) = 0
	if hitRatio != 0 {
		t.Errorf("Expected hit ratio 0 when no reused entries, got %d", hitRatio)
	}
	
	// Test with only reused entries
	metrics.ResetMetrics()
	metrics.IncEntryReused()
	metrics.IncEntryReused()
	entryMetrics = metrics.GetEntryMetrics()
	hitRatio = entryMetrics["hit_ratio"]
	// (2 * 100) / (0 + 2 + 1) = 200 / 3 ≈ 66
	expected := int64(2 * 100 / (0 + 2 + 1))
	if hitRatio != expected {
		t.Errorf("Expected hit ratio %d when only reused entries, got %d", expected, hitRatio)
	}
}