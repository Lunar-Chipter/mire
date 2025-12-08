package core

import (
	"testing"
	"time"
)

// TestLogEntryPoolOperations tests the complete lifecycle of LogEntry pooling
func TestLogEntryPoolOperations(t *testing.T) {
	// Test getting an entry from pool
	entry := GetEntryFromPool()
	if entry == nil {
		t.Fatal("GetEntryFromPool returned nil")
	}

	// Verify initial state is reset
	if !entry.Timestamp.IsZero() {
		t.Error("Entry timestamp should be zero after pool retrieval")
	}
	if entry.Level != INFO {
		t.Errorf("Entry level should be INFO after pool retrieval, got %v", entry.Level)
	}
	if entry.Message != nil {
		t.Error("Entry message should be nil after pool retrieval")
	}
	if len(entry.Fields) != 0 {
		t.Error("Entry fields map should be empty after pool retrieval")
	}

	// Set some values
	entry.Timestamp = time.Now()
	entry.Level = ERROR
	entry.Message = []byte("test message")
	entry.Fields["test"] = []byte("value")

	// Put back to pool
	PutEntryToPool(entry)

	// Get another entry and verify it was reset
	entry2 := GetEntryFromPool()
	if !entry2.Timestamp.IsZero() {
		t.Error("Entry timestamp should be reset after pool retrieval")
	}
	if entry2.Level != INFO {
		t.Errorf("Entry level should be reset to INFO, got %v", entry2.Level)
	}
	if entry2.Message != nil {
		t.Error("Entry message should be reset to nil")
	}
	if len(entry2.Fields) != 0 {
		t.Error("Entry fields map should be reset to empty")
	}
}

// TestLogEntryZeroAllocSerialization tests zero-allocation serialization
func TestLogEntryZeroAllocSerialization(t *testing.T) {
	entry := GetEntryFromPool()
	defer PutEntryToPool(entry)

	entry.Timestamp = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	entry.Level = INFO
	entry.LevelName = []byte("INFO")
	entry.Message = []byte("test message")

	result := entry.ZeroAllocJSONSerialize()
	if len(result) == 0 {
		t.Error("ZeroAllocJSONSerialize returned empty result")
	}

	// Check that result contains expected elements
	resultStr := string(result)
	if !contains(resultStr, "INFO") {
		t.Error("Result should contain level info")
	}
	if !contains(resultStr, "test message") {
		t.Error("Result should contain message")
	}
}

// TestCallerInfoPoolOperations tests the CallerInfo pool operations
func TestCallerInfoPoolOperations(t *testing.T) {
	// Get from pool
	ci := GetCallerInfoFromPool()
	if ci == nil {
		t.Fatal("GetCallerInfoFromPool returned nil")
	}

	// Verify initial state
	if ci.File != "" {
		t.Error("Initial caller info file should be empty")
	}
	if ci.Line != 0 {
		t.Error("Initial caller info line should be 0")
	}
	if ci.Function != "" {
		t.Error("Initial caller info function should be empty")
	}
	if ci.Package != "" {
		t.Error("Initial caller info package should be empty")
	}

	// Set some values
	ci.File = "test.go"
	ci.Line = 42
	ci.Function = "TestFunction"
	ci.Package = "main"

	// Put back to pool
	PutCallerInfoToPool(ci)

	// Get another and verify reset
	ci2 := GetCallerInfoFromPool()
	defer PutCallerInfoToPool(ci2)

	if ci2.File != "" {
		t.Error("Caller info file should be reset")
	}
	if ci2.Line != 0 {
		t.Error("Caller info line should be reset")
	}
	if ci2.Function != "" {
		t.Error("Caller info function should be reset")
	}
	if ci2.Package != "" {
		t.Error("Caller info package should be reset")
	}
}

// TestMapPools tests the map pools functionality
func TestMapPools(t *testing.T) {
	// Test map byte pool
	map1 := GetMapByteFromPool()
	if map1 == nil {
		t.Fatal("GetMapByteFromPool returned nil")
	}
	if len(map1) != 0 {
		t.Error("Initial map should be empty")
	}

	// Add some data
	map1["key1"] = []byte("value1")
	map1["key2"] = []byte("42")

	// Put back to pool
	PutMapByteToPool(map1)

	// Get another and verify it's clean
	map2 := GetMapByteFromPool()
	defer PutMapByteToPool(map2)

	if len(map2) != 0 {
		t.Error("Pool should return clean map after put")
	}

	// Test map float pool
	floatMap1 := GetMapFloatFromPool()
	if floatMap1 == nil {
		t.Fatal("GetMapFloatFromPool returned nil")
	}
	if len(floatMap1) != 0 {
		t.Error("Initial float map should be empty")
	}

	// Add some data
	floatMap1["metric1"] = 1.5
	floatMap1["metric2"] = 2.7

	// Put back to pool
	PutMapFloatToPool(floatMap1)

	// Get another and verify it's clean
	floatMap2 := GetMapFloatFromPool()
	defer PutMapFloatToPool(floatMap2)

	if len(floatMap2) != 0 {
		t.Error("Float pool should return clean map after put")
	}
}

// TestBufferPoolOperations tests the buffer pool operations
func TestBufferPoolOperations(t *testing.T) {
	// Get from pool
	buf := GetBufferFromPool()
	if buf == nil {
		t.Fatal("GetBufferFromPool returned nil")
	}

	if len(*buf) != 0 {
		t.Error("Initial buffer length should be 0")
	}

	// Add some data
	*buf = append(*buf, []byte("test data")...)
	if len(*buf) != 9 {
		t.Errorf("Buffer should have 9 bytes after append, got %d", len(*buf))
	}

	// Put back to pool
	PutBufferToPool(buf)

	// Get another and verify it's reset but keeps capacity
	buf2 := GetBufferFromPool()
	defer PutBufferToPool(buf2)

	if len(*buf2) != 0 {
		t.Error("Buffer should be reset to length 0 after pool return")
	}
}

// TestLevelBytesConversion tests the byte conversion functionality for levels
func TestLevelBytesConversion(t *testing.T) {
	levels := []Level{TRACE, DEBUG, INFO, NOTICE, WARN, ERROR, FATAL, PANIC}
	expected := [][]byte{
		[]byte("TRACE"),
		[]byte("DEBUG"),
		[]byte("INFO"),
		[]byte("NOTICE"),
		[]byte("WARN"),
		[]byte("ERROR"),
		[]byte("FATAL"),
		[]byte("PANIC"),
	}

	for i, level := range levels {
		bytes := level.Bytes()
		if string(bytes) != string(expected[i]) {
			t.Errorf("Level %v should convert to %s, got %s", level, string(expected[i]), string(bytes))
		}

		// Test ToBytes method as well
		bytes2 := level.ToBytes()
		if string(bytes2) != string(expected[i]) {
			t.Errorf("Level %v ToBytes should convert to %s, got %s", level, string(expected[i]), string(bytes2))
		}
	}

	// Test unknown level
	unknownLevel := Level(99)
	bytes := unknownLevel.Bytes()
	if string(bytes) != "UNKNOWN" {
		t.Errorf("Unknown level should convert to UNKNOWN, got %s", string(bytes))
	}
}

// TestCoreMetricsOperations tests the core metrics functionality
func TestCoreMetricsOperations(t *testing.T) {
	metrics := GetCoreMetrics()

	// Reset metrics for test
	metrics.EntryCreatedCount.Store(0)
	metrics.EntryReusedCount.Store(0)
	metrics.EntryPoolMissCount.Store(0)
	metrics.EntrySerializedCount.Store(0)

	// Test increment operations
	metrics.IncEntryCreated()
	metrics.IncEntryCreated()
	if metrics.CreatedCount() != 2 {
		t.Errorf("Created count should be 2, got %d", metrics.CreatedCount())
	}

	metrics.IncEntryReused()
	metrics.IncEntryReused()
	metrics.IncEntryReused()
	if metrics.ReusedCount() != 3 {
		t.Errorf("Reused count should be 3, got %d", metrics.ReusedCount())
	}

	metrics.IncEntryPoolMiss()
	if metrics.PoolMissCount() != 1 {
		t.Errorf("Pool miss count should be 1, got %d", metrics.PoolMissCount())
	}

	metrics.IncEntrySerialized()
	if metrics.SerializedCount() != 1 {
		t.Errorf("Serialized count should be 1, got %d", metrics.SerializedCount())
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}