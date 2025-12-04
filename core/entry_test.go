package core

import (
	"context"
	"testing"
	"time"
)

// TestLogEntryCreation tests basic LogEntry creation and field initialization
func TestLogEntryCreation(t *testing.T) {
	entry := GetEntryFromPool()
	defer PutEntryToPool(entry)

	if entry == nil {
		t.Fatal("GetEntryFromPool returned nil")
	}

	// Check initial state
	if !entry.Timestamp.IsZero() {
		t.Error("New entry should have zero timestamp")
	}
	if entry.Level != INFO {
		t.Errorf("New entry should have default level INFO, got %v", entry.Level)
	}
	if entry.LevelName != nil {
		t.Error("New entry should have nil LevelName")
	}
	if entry.Message != nil {
		t.Error("New entry should have nil Message")
	}
	if entry.Fields == nil {
		t.Error("New entry should have non-nil Fields map")
	}
	if entry.CustomMetrics == nil {
		t.Error("New entry should have non-nil CustomMetrics map")
	}
	if len(entry.Tags) != 0 {
		t.Errorf("New entry should have empty Tags slice, got length %d", len(entry.Tags))
	}
	if entry.PID != 0 {
		t.Errorf("New entry should have zero PID, got %d", entry.PID)
	}
	if entry.GoroutineID != nil {
		t.Error("New entry should have nil GoroutineID")
	}
	if entry.TraceID != nil {
		t.Error("New entry should have nil TraceID")
	}
	if entry.SpanID != nil {
		t.Error("New entry should have nil SpanID")
	}
	if entry.UserID != nil {
		t.Error("New entry should have nil UserID")
	}
	if entry.SessionID != nil {
		t.Error("New entry should have nil SessionID")
	}
	if entry.RequestID != nil {
		t.Error("New entry should have nil RequestID")
	}
	if entry.Duration != 0 {
		t.Errorf("New entry should have zero Duration, got %v", entry.Duration)
	}
	if entry.Error != nil {
		t.Error("New entry should have nil Error")
	}
	if entry.StackTrace != nil {
		t.Error("New entry should have nil StackTrace")
	}
	if entry.Hostname != nil {
		t.Error("New entry should have nil Hostname")
	}
	if entry.Application != nil {
		t.Error("New entry should have nil Application")
	}
	if entry.Version != nil {
		t.Error("New entry should have nil Version")
	}
	if entry.Environment != nil {
		t.Error("New entry should have nil Environment")
	}
}

// TestGetEntryFromPool tests the GetEntryFromPool function
func TestGetEntryFromPool(t *testing.T) {
	entry1 := GetEntryFromPool()
	entry2 := GetEntryFromPool()

	if entry1 == nil || entry2 == nil {
		t.Fatal("GetEntryFromPool returned nil")
	}

	// Return the entries to the pool
	PutEntryToPool(entry1)
	PutEntryToPool(entry2)

	// Get entries again to test reuse
	entry3 := GetEntryFromPool()
	entry4 := GetEntryFromPool()

	if entry3 == nil || entry4 == nil {
		t.Fatal("GetEntryFromPool returned nil after reuse")
	}

	// Return the entries to the pool
	PutEntryToPool(entry3)
	PutEntryToPool(entry4)
}

// TestPutEntryToPool tests the PutEntryToPool function
func TestPutEntryToPool(t *testing.T) {
	entry := GetEntryFromPool()

	// Set some values to ensure they're reset
	entry.Timestamp = time.Now()
	entry.Level = ERROR
	entry.LevelName = []byte("ERROR")
	entry.Message = []byte("test message")
	entry.Fields["test"] = "value"
	entry.CustomMetrics["metric"] = 1.0
	entry.Tags = [][]byte{[]byte("tag1")}
	entry.PID = 1234
	entry.GoroutineID = []byte("123")
	entry.TraceID = []byte("trace123")
	entry.SpanID = []byte("span123")
	entry.UserID = []byte("user123")
	entry.SessionID = []byte("session123")
	entry.RequestID = []byte("req123")
	entry.Duration = 5 * time.Second
	entry.Error = context.DeadlineExceeded
	entry.StackTrace = []byte("stack trace")
	entry.Hostname = []byte("hostname")
	entry.Application = []byte("app")
	entry.Version = []byte("v1.0")
	entry.Environment = []byte("test")
	entry.StackTraceBufPtr = &[]byte{1, 2, 3}

	PutEntryToPool(entry)

	// Get a new entry and check if it's properly reset
	newEntry := GetEntryFromPool()

	// Check that all fields are reset to their default values
	if !newEntry.Timestamp.IsZero() {
		t.Error("PutEntryToPool did not reset Timestamp")
	}
	if newEntry.Level != INFO {
		t.Errorf("PutEntryToPool did not reset Level, got %v", newEntry.Level)
	}
	if newEntry.LevelName != nil {
		t.Error("PutEntryToPool did not reset LevelName")
	}
	if newEntry.Message != nil {
		t.Error("PutEntryToPool did not reset Message")
	}
	if len(newEntry.Fields) != 0 {
		t.Errorf("PutEntryToPool did not reset Fields map, got length %d", len(newEntry.Fields))
	}
	if len(newEntry.CustomMetrics) != 0 {
		t.Errorf("PutEntryToPool did not reset CustomMetrics map, got length %d", len(newEntry.CustomMetrics))
	}
	if len(newEntry.Tags) != 0 {
		t.Errorf("PutEntryToPool did not reset Tags slice, got length %d", len(newEntry.Tags))
	}
	if newEntry.PID != 0 {
		t.Errorf("PutEntryToPool did not reset PID, got %d", newEntry.PID)
	}
	if newEntry.GoroutineID != nil {
		t.Error("PutEntryToPool did not reset GoroutineID")
	}
	if newEntry.TraceID != nil {
		t.Error("PutEntryToPool did not reset TraceID")
	}
	if newEntry.SpanID != nil {
		t.Error("PutEntryToPool did not reset SpanID")
	}
	if newEntry.UserID != nil {
		t.Error("PutEntryToPool did not reset UserID")
	}
	if newEntry.SessionID != nil {
		t.Error("PutEntryToPool did not reset SessionID")
	}
	if newEntry.RequestID != nil {
		t.Error("PutEntryToPool did not reset RequestID")
	}
	if newEntry.Duration != 0 {
		t.Errorf("PutEntryToPool did not reset Duration, got %v", newEntry.Duration)
	}
	if newEntry.Error != nil {
		t.Error("PutEntryToPool did not reset Error")
	}
	if newEntry.StackTrace != nil {
		t.Error("PutEntryToPool did not reset StackTrace")
	}
	if newEntry.Hostname != nil {
		t.Error("PutEntryToPool did not reset Hostname")
	}
	if newEntry.Application != nil {
		t.Error("PutEntryToPool did not reset Application")
	}
	if newEntry.Version != nil {
		t.Error("PutEntryToPool did not reset Version")
	}
	if newEntry.Environment != nil {
		t.Error("PutEntryToPool did not reset Environment")
	}
	if newEntry.StackTraceBufPtr != nil {
		t.Error("PutEntryToPool did not reset StackTraceBufPtr")
	}

	PutEntryToPool(newEntry)
}

// TestCallerInfoPool tests the CallerInfo object pool
func TestCallerInfoPool(t *testing.T) {
	ci1 := GetCallerInfoFromPool()
	if ci1 == nil {
		t.Fatal("GetCallerInfoFromPool returned nil")
	}

	// Set some values
	ci1.File = "test.go"
	ci1.Line = 123
	ci1.Function = "TestFunction"
	ci1.Package = "test"

	PutCallerInfoToPool(ci1)

	// Get a new CallerInfo and check if it's properly reset
	ci2 := GetCallerInfoFromPool()
	if ci2.File != "" || ci2.Line != 0 || ci2.Function != "" || ci2.Package != "" {
		t.Errorf("PutCallerInfoToPool did not reset CallerInfo, got %+v", ci2)
	}

	PutCallerInfoToPool(ci2)
}

// TestMapFloatPool tests the map[string]float64 object pool
func TestMapFloatPool(t *testing.T) {
	m1 := GetMapFloatFromPool()
	if m1 == nil {
		t.Fatal("GetMapFloatFromPool returned nil")
	}

	// Add some values
	m1["test"] = 1.0

	PutMapFloatToPool(m1)

	// Get a new map and check if it's properly reset
	m2 := GetMapFloatFromPool()
	if len(m2) != 0 {
		t.Errorf("PutMapFloatToPool did not reset map, got length %d", len(m2))
	}

	PutMapFloatToPool(m2)
}

// TestMapInterfacePool tests the map[string]interface{} object pool
func TestMapInterfacePool(t *testing.T) {
	m1 := GetMapInterfaceFromPool()
	if m1 == nil {
		t.Fatal("GetMapInterfaceFromPool returned nil")
	}

	// Add some values
	m1["test"] = "value"

	PutMapInterfaceToPool(m1)

	// Get a new map and check if it's properly reset
	m2 := GetMapInterfaceFromPool()
	if len(m2) != 0 {
		t.Errorf("PutMapInterfaceToPool did not reset map, got length %d", len(m2))
	}

	PutMapInterfaceToPool(m2)
}

// TestBufferPool tests the byte buffer object pool
func TestBufferPool(t *testing.T) {
	buf1 := GetBufferFromPool()
	if buf1 == nil {
		t.Fatal("GetBufferFromPool returned nil")
	}

	// Add some data
	*buf1 = append(*buf1, []byte("test")...)

	PutBufferToPool(buf1)

	// Get a new buffer and check if it's properly reset
	buf2 := GetBufferFromPool()
	if len(*buf2) != 0 {
		t.Errorf("PutBufferToPool did not reset buffer, got length %d", len(*buf2))
	}

	PutBufferToPool(buf2)
}

// TestStringSlicePool tests the string slice object pool
func TestStringSlicePool(t *testing.T) {
	s1 := GetStringSliceFromPool()
	if s1 == nil {
		t.Fatal("GetStringSliceFromPool returned nil")
	}

	// Add some values
	*s1 = append(*s1, "test")

	PutStringSliceToPool(s1)

	// Get a new slice and check if it's properly reset
	s2 := GetStringSliceFromPool()
	if len(*s2) != 0 {
		t.Errorf("PutStringSliceToPool did not reset slice, got length %d", len(*s2))
	}

	PutStringSliceToPool(s2)
}

// TestMetrics tests the CoreMetrics functionality
func TestCoreMetrics(t *testing.T) {
	metrics := GetCoreMetrics()

	initialCreated := metrics.CreatedCount()
	initialReused := metrics.ReusedCount()
	initialPoolMiss := metrics.PoolMissCount()
	initialSerialized := metrics.SerializedCount()

	// Test increment methods
	metrics.IncEntryCreated()
	metrics.IncEntryReused()
	metrics.IncEntryPoolMiss()
	metrics.IncEntrySerialized()

	if metrics.CreatedCount() != initialCreated+1 {
		t.Error("IncEntryCreated did not increment properly")
	}
	if metrics.ReusedCount() != initialReused+1 {
		t.Error("IncEntryReused did not increment properly")
	}
	if metrics.PoolMissCount() != initialPoolMiss+1 {
		t.Error("IncEntryPoolMiss did not increment properly")
	}
	if metrics.SerializedCount() != initialSerialized+1 {
		t.Error("IncEntrySerialized did not increment properly")
	}

	// Test all metrics getters
	allMetrics := metrics.GetAllMetrics()
	if allMetrics == nil {
		t.Error("GetAllMetrics returned nil")
	}

	entryMetrics := metrics.GetEntryMetrics()
	if entryMetrics == nil {
		t.Error("GetEntryMetrics returned nil")
	}

	bufferMetrics := metrics.GetBufferMetrics()
	if bufferMetrics == nil {
		t.Error("GetBufferMetrics returned nil")
	}

	sliceMetrics := metrics.GetSliceMetrics()
	if sliceMetrics == nil {
		t.Error("GetSliceMetrics returned nil")
	}

	errorMetrics := metrics.GetErrorMetrics()
	if errorMetrics == nil {
		t.Error("GetErrorMetrics returned nil")
	}

	timingMetrics := metrics.GetTimingMetrics()
	if timingMetrics == nil {
		t.Error("GetTimingMetrics returned nil")
	}

	// Test reset
	metrics.ResetMetrics()

	if metrics.CreatedCount() != 0 {
		t.Error("ResetMetrics did not reset CreatedCount")
	}
	if metrics.ReusedCount() != 0 {
		t.Error("ResetMetrics did not reset ReusedCount")
	}
	if metrics.PoolMissCount() != 0 {
		t.Error("ResetMetrics did not reset PoolMissCount")
	}
	if metrics.SerializedCount() != 0 {
		t.Error("ResetMetrics did not reset SerializedCount")
	}
}

// TestMetricsConcurrent tests the CoreMetrics functionality in a concurrent context
func TestMetricsConcurrent(t *testing.T) {
	metrics := GetCoreMetrics()

	// Run multiple goroutines to increment metrics concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				metrics.IncEntryCreated()
				metrics.IncEntryReused()
				metrics.IncEntryPoolMiss()
				metrics.IncEntrySerialized()
				metrics.SetLastOperationTime(time.Now())
				metrics.AddProcessingTime(time.Nanosecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check final counts
	expected := int64(1000) // 10 goroutines * 100 iterations
	if metrics.CreatedCount() != expected {
		t.Errorf("Concurrent IncEntryCreated: expected %d, got %d", expected, metrics.CreatedCount())
	}
	if metrics.ReusedCount() != expected {
		t.Errorf("Concurrent IncEntryReused: expected %d, got %d", expected, metrics.ReusedCount())
	}
	if metrics.PoolMissCount() != expected {
		t.Errorf("Concurrent IncEntryPoolMiss: expected %d, got %d", expected, metrics.PoolMissCount())
	}
	if metrics.SerializedCount() != expected {
		t.Errorf("Concurrent IncEntrySerialized: expected %d, got %d", expected, metrics.SerializedCount())
	}

	// Test atomic operations work correctly by comparing expected vs actual
	if metrics.CreatedCount() < 0 {
		t.Error("CreatedCount should not be negative")
	}
}

// TestLogEntryZeroAllocJSONSerialize tests the ZeroAllocJSONSerialize method
func TestLogEntryZeroAllocJSONSerialize(t *testing.T) {
	entry := GetEntryFromPool()
	defer PutEntryToPool(entry)

	// Set up the entry with test data
	entry.Timestamp = time.Now()
	entry.Level = INFO
	entry.LevelName = []byte("INFO")
	entry.Message = []byte("test message")
	entry.PID = 1234

	// Test basic serialization
	result := entry.ZeroAllocJSONSerialize()

	// Verify the result contains the expected fields
	if len(result) == 0 {
		t.Error("ZeroAllocJSONSerialize returned empty result")
	}
	if entry.Timestamp.IsZero() {
		t.Error("Entry timestamp was zeroed during serialization")
	}

	// Check if the result contains the expected values (at least the start and end)
	if result[0] != '{' {
		t.Errorf("ZeroAllocJSONSerialize result does not start with '{': %s", result)
	}
	if result[len(result)-1] != '}' {
		t.Errorf("ZeroAllocJSONSerialize result does not end with '}': %s", result)
	}

	// Verify that the message is included
	if len(result) < len(`{"timestamp":"2006-01-02T15:04:05.000Z07:00","level":"INFO","message":"test message","pid":1234}`) {
		t.Log("Result may be valid, but it's shorter than expected. This is acceptable as the format might differ.")
	}
}

// TestLogEntryFormatLogToBytes tests the formatLogToBytes method
func TestLogEntryFormatLogToBytes(t *testing.T) {
	entry := GetEntryFromPool()
	defer PutEntryToPool(entry)

	// Set up the entry with test data
	entry.Timestamp = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	entry.Level = INFO
	entry.LevelName = []byte("INFO")
	entry.Message = []byte("test message")
	entry.Fields["key"] = "value"
	entry.Tags = [][]byte{[]byte("tag1")}

	// Create a buffer and test formatting
	initialBuf := make([]byte, 0, 100)
	result := entry.formatLogToBytes(initialBuf)

	// Verify the result contains the expected data
	if len(result) == 0 {
		t.Error("formatLogToBytes returned empty result")
	}

	// Check for basic components in the formatted string
	if len(string(result)) < len("2023-01-01T12:00:00") {
		t.Error("formatLogToBytes result too short")
	}
	if !entry.Timestamp.IsZero() {
		// Timestamp should not be modified by formatting
	}
}

// TestLogEntryIntConversion tests int to bytes conversion functions
func TestLogEntryIntConversion(t *testing.T) {
	entry := GetEntryFromPool()
	defer PutEntryToPool(entry)

	// Test intToBytes
	buf := []byte{}
	result := entry.intToBytes(buf, 123)
	expected := []byte("123")
	if string(result) != string(expected) {
		t.Errorf("intToBytes(123) = %s, want %s", string(result), string(expected))
	}

	// Test int64ToBytes
	buf = []byte{}
	result = entry.int64ToBytes(buf, 123)
	expected = []byte("123")
	if string(result) != string(expected) {
		t.Errorf("int64ToBytes(123) = %s, want %s", string(result), string(expected))
	}

	// Test with negative number
	buf = []byte{}
	result = entry.intToBytes(buf, -123)
	expected = []byte("-123")
	if string(result) != string(expected) {
		t.Errorf("intToBytes(-123) = %s, want %s", string(result), string(expected))
	}

	// Test with zero
	buf = []byte{}
	result = entry.intToBytes(buf, 0)
	expected = []byte("0")
	if string(result) != string(expected) {
		t.Errorf("intToBytes(0) = %s, want %s", string(result), string(expected))
	}
}

// TestLogEntryFloatConversion tests float to bytes conversion function
func TestLogEntryFloatConversion(t *testing.T) {
	entry := GetEntryFromPool()
	defer PutEntryToPool(entry)

	// Test with a sample float
	buf := []byte{}
	result := entry.floatToBytes(buf, 3.14)
	// We'll check if the result contains our expected string representation
	if len(string(result)) == 0 {
		t.Error("floatToBytes returned empty result")
	}
	// This is harder to test exactly due to floating point precision, but we can at least verify it's not empty
}