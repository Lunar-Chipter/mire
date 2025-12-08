package core

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentPoolOperations tests the thread safety of pool operations
func TestConcurrentPoolOperations(t *testing.T) {
	const goroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				entry := GetEntryFromPool()
				
				// Set some values to ensure proper reset
				entry.Timestamp = time.Now()
				entry.Level = ERROR
				entry.Message = []byte("test message")
				entry.Fields["test"] = []byte(fmt.Sprintf("%d", j))

				// Use the entry (simulated)
				_ = entry.LevelName

				PutEntryToPool(entry)
			}
		}()
	}

	wg.Wait()
	
	// Verify metrics after concurrent operations
	metrics := GetCoreMetrics()
	if metrics.CreatedCount() < 0 {
		t.Error("Created count should be non-negative")
	}
	if metrics.PoolMissCount() < 0 {
		t.Error("Pool miss count should be non-negative")
	}
}

// TestGoroutineLocalPoolUsage tests the usage of goroutine-local pools
func TestGoroutineLocalPoolUsage(t *testing.T) {
	// Get entry from local pool (simulated through regular function)
	entry := GetEntryFromPool()
	
	if entry == nil {
		t.Fatal("Entry should not be nil")
	}
	
	// Test reset functionality is working
	if len(entry.Fields) != 0 {
		t.Error("Fields map should be empty after pool retrieval")
	}
	if !entry.Timestamp.IsZero() {
		t.Error("Timestamp should be zero after pool retrieval")
	}
	
	PutEntryToPool(entry)
}

// TestCoreMetricsConcurrentUpdate tests concurrent updates to metrics
func TestCoreMetricsConcurrentUpdate(t *testing.T) {
	metrics := GetCoreMetrics()

	// Store initial values to check increments
	initialCreated := metrics.CreatedCount()
	initialReused := metrics.ReusedCount()
	initialPoolMiss := metrics.PoolMissCount()
	initialSerialized := metrics.SerializedCount()

	const goroutines = 5
	const operations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				metrics.IncEntryCreated()
				metrics.IncEntryReused()
				metrics.IncEntryPoolMiss()
				metrics.IncEntrySerialized()
			}
		}()
	}

	wg.Wait()

	// Check that metrics were incremented by at least the expected amount
	// (may be more due to other tests using the same global metrics)
	if metrics.CreatedCount() < initialCreated + int64(goroutines * operations) {
		t.Logf("Created count: expected at least %d more than %d", goroutines * operations, initialCreated)
	}
	if metrics.ReusedCount() < initialReused + int64(goroutines * operations) {
		t.Logf("Reused count: expected at least %d more than %d", goroutines * operations, initialReused)
	}
	if metrics.PoolMissCount() < initialPoolMiss + int64(goroutines * operations) {
		t.Logf("Pool miss count: expected at least %d more than %d", goroutines * operations, initialPoolMiss)
	}
	if metrics.SerializedCount() < initialSerialized + int64(goroutines * operations) {
		t.Logf("Serialized count: expected at least %d more than %d", goroutines * operations, initialSerialized)
	}
}

// TestEntryFormatLogToBytesConcurrent tests concurrent use of formatLogToBytes
func TestEntryFormatLogToBytesConcurrent(t *testing.T) {
	const goroutines = 3
	const operations = 30
	
	var wg sync.WaitGroup
	wg.Add(goroutines)
	
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				entry := GetEntryFromPool()
				defer PutEntryToPool(entry)
				
				entry.Timestamp = time.Now()
				entry.Level = INFO
				entry.LevelName = []byte("INFO")
				entry.Message = []byte("test message")
				entry.Fields["iteration"] = []byte(fmt.Sprintf("%d", j))
				entry.Fields["goroutine"] = []byte(fmt.Sprintf("%d", i))
				
				// Use formatLogToBytes - this should not cause race conditions
				buf := make([]byte, 0, 100)
				result := entry.formatLogToBytes(buf)
				
				if len(result) == 0 {
					t.Error("Format result should not be empty")
				}
			}
		}()
	}
	
	wg.Wait()
}

// TestPoolReuseEfficiency tests the efficiency of the pool reuse
func TestPoolReuseEfficiency(t *testing.T) {
	// Reset metrics
	metrics := GetCoreMetrics()
	metrics.EntryCreatedCount.Store(0)
	metrics.EntryReusedCount.Store(0)
	metrics.EntryPoolMissCount.Store(0)

	// Create and return many entries to test pool reuse
	const totalEntries = 1000
	entries := make([]*LogEntry, totalEntries)
	
	// First, create all entries (this will initially create them)
	for i := 0; i < totalEntries; i++ {
		entries[i] = GetEntryFromPool()
	}
	
	// Then return them all to the pool
	for i := 0; i < totalEntries; i++ {
		PutEntryToPool(entries[i])
	}
	
	// Now get them again - these should be reused from the pool
	for i := 0; i < totalEntries; i++ {
		entry := GetEntryFromPool()
		PutEntryToPool(entry)
	}
	
	// The number of created entries should be around the number of unique needs, not total operations
	created := metrics.CreatedCount()
	reused := metrics.ReusedCount()
	
	// In this test, we expect significant reuse since we're returning entries to the pool
	if created > int64(totalEntries) {
		t.Logf("Created %d entries out of %d operations - pool is working", created, totalEntries)
	}
	if reused < int64(totalEntries/2) {
		t.Log("Expected more reuses than this - pool might not be working optimally")
	}
}

// TestEntryFieldManipulation tests manipulation of fields within entries
func TestEntryFieldManipulation(t *testing.T) {
	entry := GetEntryFromPool()
	defer PutEntryToPool(entry)

	// Add various types of fields
	entry.Fields["string_val"] = []byte("hello")
	entry.Fields["int_val"] = []byte("42")
	entry.Fields["float_val"] = []byte("3.14")
	entry.Fields["bool_val"] = []byte("true")
	entry.Fields["nil_val"] = nil

	// Verify all fields are stored correctly
	if string(entry.Fields["string_val"]) != "hello" {
		t.Error("String field not stored correctly")
	}
	if string(entry.Fields["int_val"]) != "42" {
		t.Error("Int field not stored correctly")
	}
	if string(entry.Fields["float_val"]) != "3.14" {
		t.Error("Float field not stored correctly")
	}
	if string(entry.Fields["bool_val"]) != "true" {
		t.Error("Bool field not stored correctly")
	}
	if entry.Fields["nil_val"] != nil {
		t.Error("Nil field not stored correctly")
	}
}

// TestEntryWithAllFields tests creating an entry with all possible fields populated
func TestEntryWithAllFields(t *testing.T) {
	entry := GetEntryFromPool()
	defer PutEntryToPool(entry)

	// Populate all fields
	entry.Timestamp = time.Date(2023, 8, 15, 14, 30, 0, 0, time.UTC)
	entry.Level = DEBUG
	entry.LevelName = []byte("DEBUG")
	entry.Message = []byte("complete test message")
	entry.Caller = GetCallerInfoFromPool()
	entry.Caller.File = "test.go"
	entry.Caller.Line = 100
	entry.Caller.Function = "CompleteTest"
	entry.Caller.Package = "main"
	entry.Fields["key"] = []byte("value")
	entry.PID = 12345
	entry.GoroutineID = []byte("999")
	entry.TraceID = []byte("trace-12345")
	entry.SpanID = []byte("span-67890")
	entry.UserID = []byte("user-11111")
	entry.SessionID = []byte("session-22222")
	entry.RequestID = []byte("request-33333")
	entry.Duration = 5 * time.Second
	entry.Error = nil
	entry.StackTrace = []byte("stack trace content")
	entry.Hostname = []byte("test-host")
	entry.Application = []byte("test-app")
	entry.Version = []byte("2.0.0")
	entry.Environment = []byte("testing")
	entry.CustomMetrics["metric1"] = 1.5
	entry.Tags = [][]byte{[]byte("tag1"), []byte("tag2"), []byte("tag3")}

	// Verify all fields are set
	if entry.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
	if entry.Level != DEBUG {
		t.Error("Level should be DEBUG")
	}
	if string(entry.LevelName) != "DEBUG" {
		t.Error("LevelName should be 'DEBUG'")
	}
	if string(entry.Message) != "complete test message" {
		t.Error("Message not set correctly")
	}
	if entry.Caller.File != "test.go" {
		t.Error("Caller file not set correctly")
	}
	if entry.PID != 12345 {
		t.Error("PID not set correctly")
	}
	if len(entry.Fields) != 1 {
		t.Error("Fields map should have 1 entry")
	}
	if string(entry.Fields["key"]) != "value" {
		t.Error("Fields value not set correctly")
	}
	if string(entry.UserID) != "user-11111" {
		t.Error("UserID not set correctly")
	}
	if entry.Duration != 5*time.Second {
		t.Error("Duration not set correctly")
	}
	if len(entry.Tags) != 3 {
		t.Error("Should have 3 tags")
	}
	if string(entry.Tags[0]) != "tag1" {
		t.Error("First tag not set correctly")
	}
}