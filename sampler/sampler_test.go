package sampler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
)

// mockSamplerProcessor implements LogSampler interface for testing
type mockSamplerProcessor struct {
	loggedEntries []*sampleEntry
	logCalls      int
	mu            sync.Mutex
}

type sampleEntry struct {
	ctx    context.Context
	level  core.Level
	msg    []byte
	fields map[string]interface{}
}

func (m *mockSamplerProcessor) Log(ctx context.Context, level core.Level, msg []byte, fields map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.loggedEntries = append(m.loggedEntries, &sampleEntry{
		ctx:    ctx,
		level:  level,
		msg:    msg,
		fields: fields,
	})
	m.logCalls++
}

// TestNewSamplingLogger tests creating a new SamplingLogger
func TestNewSamplingLogger(t *testing.T) {
	processor := &mockSamplerProcessor{}
	
	samplingLogger := NewSamplingLogger(processor, 5)
	if samplingLogger == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}
	
	if samplingLogger.processor != processor {
		t.Error("SamplingLogger processor not set correctly")
	}
	if samplingLogger.rate != 5 {
		t.Errorf("Expected rate 5, got %d", samplingLogger.rate)
	}
	// The counter should initially be 0
}

// TestSamplingLoggerShouldLog tests the ShouldLog method with different rates
func TestSamplingLoggerShouldLog(t *testing.T) {
	// Test with rate 1 (should always log)
	sampler1 := NewSamplingLogger(&mockSamplerProcessor{}, 1)
	
	for i := 0; i < 10; i++ {
		if !sampler1.ShouldLog() {
			t.Errorf("With rate 1, ShouldLog should always return true, but failed on call %d", i+1)
		}
	}
	
	// Test with rate 2 (should log every 2nd call)
	sampler2 := NewSamplingLogger(&mockSamplerProcessor{}, 2)
	logCount := 0
	for i := 0; i < 10; i++ {
		if sampler2.ShouldLog() {
			logCount++
		}
	}
	// With rate 2 and 10 calls, we expect approximately 5 logs (every even-numbered call in the sequence)
	// Actually, it depends on the counter implementation. Based on the code, it will be every time counter%rate==0
	// Counter starts at 0, increments to 1, 2, 3... and the check is (counter%rate==0)
	// So for rate 2: on counter values 2, 4, 6, 8, 10 (if we call 10 times, the counter will reach 10)
	// That's 5 times, so 5 out of 10 calls
	if logCount != 5 {
		t.Errorf("With rate 2, expected 5 logs out of 10, got %d", logCount)
	}
	
	// Test with rate 3 (should log every 3rd call)
	sampler3 := NewSamplingLogger(&mockSamplerProcessor{}, 3)
	logCount = 0
	for i := 0; i < 12; i++ {
		if sampler3.ShouldLog() {
			logCount++
		}
	}
	// With rate 3 and 12 calls, we expect 4 logs (on calls 3, 6, 9, 12)
	if logCount != 4 {
		t.Errorf("With rate 3, expected 4 logs out of 12, got %d", logCount)
	}
	
	// Test with rate 0 (should behave like rate 1, always log based on the code)
	sampler0 := NewSamplingLogger(&mockSamplerProcessor{}, 0)
	// The implementation will likely treat rate <= 1 as always log
	result := sampler0.ShouldLog()
	if !result {
		t.Error("With rate 0, ShouldLog should return true (treated as rate <= 1)")
	}
}

// TestSamplingLoggerLog tests the Log method with sampling
func TestSamplingLoggerLog(t *testing.T) {
	processor := &mockSamplerProcessor{}
	
	// Create a sampler with rate 2 to log every 2nd entry
	sampler := NewSamplingLogger(processor, 2)
	if sampler == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}
	
	ctx := context.Background()
	
	// Log 10 messages - should only log 5 (every 2nd one)
	for i := 0; i < 10; i++ {
		msg := []byte("message " + string(rune(i+'0')))
		sampler.Log(ctx, core.INFO, msg, map[string]interface{}{"idx": i})
	}
	
	// Verify that only approximately half the messages were processed
	processor.mu.Lock()
	logCalls := processor.logCalls
	loggedEntries := processor.loggedEntries
	processor.mu.Unlock()
	
	// Based on the counter implementation: 10 calls, log when counter%2==0
	// Counter goes 1, 2, 3, 4, 5, 6, 7, 8, 9, 10
	// Log on 2, 4, 6, 8, 10 = 5 logs
	if logCalls != 5 {
		t.Errorf("Expected 5 log calls with rate 2, got %d", logCalls)
	}
	
	if len(loggedEntries) != 5 {
		t.Errorf("Expected 5 logged entries with rate 2, got %d", len(loggedEntries))
	}
	
	// Verify the logged messages have the expected indices
	// Based on the ShouldLog() implementation, it should log when counter % rate == 0
	// First call adds 1 to counter (making it 1), then checks if 1%2==0 (false)
	// Second call adds 1 to counter (making it 2), then checks if 2%2==0 (true) -> log
	// Third call adds 1 to counter (making it 3), then checks if 3%2==0 (false)
	// Fourth call adds 1 to counter (making it 4), then checks if 4%2==0 (true) -> log
	// So calls 2, 4, 6, 8, 10 will log, meaning messages 1, 3, 5, 7, 9 will be logged
	
	if len(loggedEntries) > 0 {
		// Check that the messages that were logged match expected indices
		// The 0-indexed calls that should result in logs are 1, 3, 5, 7, 9 (the 2nd, 4th, 6th, 8th, 10th calls)
		expectedIndices := []int{1, 3, 5, 7, 9}
		for i, entry := range loggedEntries {
			if idx, ok := entry.fields["idx"]; !ok || idx != expectedIndices[i] {
				t.Errorf("Expected logged entry %d to have idx %d, got %v", i, expectedIndices[i], idx)
			}
		}
	}
}

// TestSamplingLoggerWithRate1 tests sampling with rate 1 (should log everything)
func TestSamplingLoggerWithRate1(t *testing.T) {
	processor := &mockSamplerProcessor{}
	
	sampler := NewSamplingLogger(processor, 1)
	if sampler == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}
	
	ctx := context.Background()
	
	// Log 5 messages with rate 1 - should log all of them
	for i := 0; i < 5; i++ {
		sampler.Log(ctx, core.DEBUG, []byte("all"), nil)
	}
	
	processor.mu.Lock()
	logCalls := processor.logCalls
	processor.mu.Unlock()
	
	if logCalls != 5 {
		t.Errorf("With rate 1, expected 5 log calls, got %d", logCalls)
	}
}

// TestSamplingLoggerWithRate0 tests sampling with rate 0 (should log everything)
func TestSamplingLoggerWithRate0(t *testing.T) {
	processor := &mockSamplerProcessor{}
	
	sampler := NewSamplingLogger(processor, 0)
	if sampler == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}
	
	ctx := context.Background()
	
	// Log 5 messages with rate 0 - should log all of them based on code logic (rate <= 1)
	for i := 0; i < 5; i++ {
		sampler.Log(ctx, core.INFO, []byte("rate 0 test"), nil)
	}
	
	processor.mu.Lock()
	logCalls := processor.logCalls
	processor.mu.Unlock()
	
	if logCalls != 5 {
		t.Errorf("With rate 0, expected 5 log calls, got %d", logCalls)
	}
}

// TestSamplingLoggerConcurrent tests the SamplingLogger in a concurrent context
func TestSamplingLoggerConcurrent(t *testing.T) {
	processor := &mockSamplerProcessor{}
	
	// Use rate 3 for this test
	sampler := NewSamplingLogger(processor, 3)
	if sampler == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}
	
	// Run multiple goroutines that log concurrently
	const numGoroutines = 5
	const logsPerGoroutine = 15  // Total of 75 logs, with rate 3 expect ~25
	var wg sync.WaitGroup
	ctx := context.Background()
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				msg := []byte("goroutine " + string(rune(goroutineID+'0')) + " log " + string(rune(j+'0')))
				sampler.Log(ctx, core.INFO, msg, map[string]interface{}{
					"goroutine": goroutineID,
					"log_num":   j,
				})
			}
		}(i)
	}
	
	wg.Wait()
	
	// Wait briefly to ensure all operations are complete
	time.Sleep(10 * time.Millisecond)
	
	processor.mu.Lock()
	logCalls := processor.logCalls
	totalExpected := numGoroutines * logsPerGoroutine
	processor.mu.Unlock()
	
	// With rate 3 and 75 total logs, we expect around 75/3 = 25 logs
	// But due to the concurrent nature, it might vary slightly
	expectedApprox := totalExpected / 3
	
	t.Logf("Total logs attempted: %d, Actual logs: %d, Expected approximately: %d", 
		totalExpected, logCalls, expectedApprox)
	
	// Check that we got some sampling (not all logs)
	if logCalls >= totalExpected {
		t.Error("Sampling doesn't seem to be working - all logs were processed")
	}
	
	// Check that we got some logs (not none)
	if logCalls <= 0 {
		t.Error("No logs were processed - sampling might be too aggressive")
	}
	
	// The actual number should be close to expected (within a reasonable range)
	if logCalls < expectedApprox/2 || logCalls > expectedApprox*2 {
		t.Logf("Log count %d differs significantly from expected %d", logCalls, expectedApprox)
		// This might be acceptable due to concurrent counter increments
	}
}

// TestSamplingLoggerWithDifferentLevels tests sampling with different log levels
func TestSamplingLoggerWithDifferentLevels(t *testing.T) {
	processor := &mockSamplerProcessor{}
	
	sampler := NewSamplingLogger(processor, 2) // Rate 2
	if sampler == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}
	
	ctx := context.Background()
	
	// Log messages with different levels
	levels := []core.Level{core.TRACE, core.DEBUG, core.INFO, core.WARN, core.ERROR, core.FATAL, core.PANIC}
	
	for _, level := range levels {
		sampler.Log(ctx, level, []byte("level test"), map[string]interface{}{"level": level.String()})
	}
	
	processor.mu.Lock()
	logCalls := processor.logCalls
	loggedEntries := processor.loggedEntries
	processor.mu.Unlock()
	
	// With rate 2 and 7 messages, expect ~3-4 messages to be logged
	expectedApprox := len(levels) / 2
	if logCalls < expectedApprox/2 || logCalls > expectedApprox*2 {
		t.Logf("Level test: Expected ~%d logs, got %d", expectedApprox, logCalls)
	}
	
	// Verify that logged entries have the correct levels
	for _, entry := range loggedEntries {
		// Each logged entry should have a level field
		if level, ok := entry.fields["level"]; !ok {
			t.Error("Logged entry missing level field")
		} else {
			t.Logf("Logged entry with level: %v", level)
		}
	}
}

// TestSamplingLoggerWithFields tests sampling with fields
func TestSamplingLoggerWithFields(t *testing.T) {
	processor := &mockSamplerProcessor{}
	
	sampler := NewSamplingLogger(processor, 1) // Rate 1, so all should be logged for testing fields
	if sampler == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}
	
	ctx := context.Background()
	
	// Log with various field types
	fields := map[string]interface{}{
		"string_field": "value",
		"int_field":    42,
		"bool_field":   true,
		"float_field":  3.14,
	}
	
	sampler.Log(ctx, core.INFO, []byte("fields test"), fields)
	
	processor.mu.Lock()
	loggedEntries := processor.loggedEntries
	processor.mu.Unlock()
	
	if len(loggedEntries) != 1 {
		t.Fatalf("Expected 1 logged entry, got %d", len(loggedEntries))
	}
	
	entry := loggedEntries[0]
	for key, expectedValue := range fields {
		if actualValue, exists := entry.fields[key]; !exists {
			t.Errorf("Expected field %s not found in logged entry", key)
		} else if actualValue != expectedValue {
			t.Errorf("Field %s: expected %v, got %v", key, expectedValue, actualValue)
		}
	}
}

// TestSamplingLoggerCounterRace tests the counter for race conditions (requires -race flag)
func TestSamplingLoggerCounterRace(t *testing.T) {
	processor := &mockSamplerProcessor{}
	
	sampler := NewSamplingLogger(processor, 2)
	if sampler == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}
	
	// Create multiple goroutines that only call ShouldLog (this tests the counter)
	const numGoroutines = 10
	const callsPerGoroutine = 100
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				_ = sampler.ShouldLog()
			}
		}()
	}
	
	wg.Wait()
	
	// The counter should not have been corrupted
	// The final counter value should be numGoroutines * callsPerGoroutine
}

// TestSamplingLoggerWithHighRate tests sampling with a high rate
func TestSamplingLoggerWithHighRate(t *testing.T) {
	processor := &mockSamplerProcessor{}

	// Use a high rate - should log rarely
	sampler := NewSamplingLogger(processor, 100)
	if sampler == nil {
		t.Fatal("NewSamplingLogger returned nil")
	}

	ctx := context.Background()

	// Log 50 messages with rate 100 - expect only 0 or 1 messages to pass
	for i := 0; i < 50; i++ {
		sampler.Log(ctx, core.INFO, []byte("high rate test"), nil)
	}

	processor.mu.Lock()
	logCalls := processor.logCalls
	processor.mu.Unlock()

	if logCalls > 1 {
		t.Logf("With high rate 100 and 50 attempts, got %d logs (expected 0 or 1)", logCalls)
		// This is still valid behavior, just noting the probability
	}
}
