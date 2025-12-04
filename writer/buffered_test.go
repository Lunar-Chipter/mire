package writer

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockWriteError implements io.Writer that returns an error on write
type mockWriteError struct{}

func (m *mockWriteError) Write(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}

// mockWriteCounter implements io.Writer that counts writes
type mockWriteCounter struct {
	writeCount int
	mu         sync.Mutex
	data       []byte
}

func (m *mockWriteCounter) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.writeCount++
	m.data = append(m.data, p...)
	return len(p), nil
}

func (m *mockWriteCounter) GetWriteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.writeCount
}

func (m *mockWriteCounter) GetData() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data
}

// TestNewBufferedWriter tests creating a new BufferedWriter
func TestNewBufferedWriter(t *testing.T) {
	var output bytes.Buffer
	errorHandler := func(err error) {}
	
	bufferedWriter := NewBufferedWriter(&output, 10, 100*time.Millisecond, errorHandler, 5, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Check initial state
	if bufferedWriter.writer != &output {
		t.Error("BufferedWriter writer not set correctly")
	}
	if bufferedWriter.bufferSize != 10 {
		t.Errorf("Expected buffer size 10, got %d", bufferedWriter.bufferSize)
	}
	if bufferedWriter.flushInterval != 100*time.Millisecond {
		t.Errorf("Expected flush interval 100ms, got %v", bufferedWriter.flushInterval)
	}
	if bufferedWriter.batchSize != 5 {
		t.Errorf("Expected batch size 5, got %d", bufferedWriter.batchSize)
	}
	if bufferedWriter.batchTimeout != 1*time.Second {
		t.Errorf("Expected batch timeout 1s, got %v", bufferedWriter.batchTimeout)
	}
}

// TestBufferedWriterWrite tests writing to the BufferedWriter
func TestBufferedWriterWrite(t *testing.T) {
	var output bytes.Buffer
	errorHandler := func(err error) {}
	
	bufferedWriter := NewBufferedWriter(&output, 10, 100*time.Millisecond, errorHandler, 5, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	data := []byte("test data")
	n, err := bufferedWriter.Write(data)
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned %d, expected %d", n, len(data))
	}
	
	// Wait a bit for the flush
	time.Sleep(150 * time.Millisecond)
	
	result := output.String()
	if !strings.Contains(result, string(data)) {
		t.Errorf("Output should contain written data, got: %s", result)
	}
}

// TestBufferedWriterFullBuffer tests behavior when buffer is full
func TestBufferedWriterFullBuffer(t *testing.T) {
	counter := &mockWriteCounter{}
	
	// Create a buffered writer with a small channel size
	bufferedWriter := NewBufferedWriter(counter, 2, 100*time.Millisecond, func(err error) {}, 5, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Write data to fill up the internal channel
	for i := 0; i < 10; i++ {
		data := []byte("data " + string(rune(i+'0')))
		_, err := bufferedWriter.Write(data)
		if err != nil {
			t.Errorf("Write %d returned error: %v", i, err)
		}
	}
	
	// Wait for processing
	time.Sleep(200 * time.Millisecond)
	
	// Check statistics - some logs might have been dropped if the buffer filled up
	stats := bufferedWriter.Stats()
	totalLogs := stats["total_logs"].(int64)
	droppedLogs := stats["dropped_logs"].(int64)
	
	// Should have processed 10 total logs, with some potentially dropped
	if totalLogs != 10 {
		t.Errorf("Expected total logs 10, got %d", totalLogs)
	}
	
	// We expect some dropped logs because we wrote more than buffer size
	// But the exact number depends on the implementation
	t.Logf("Total logs: %d, Dropped logs: %d", totalLogs, droppedLogs)
}

// TestBufferedWriterWriteError tests behavior when the underlying writer returns an error
func TestBufferedWriterWriteError(t *testing.T) {
	var capturedError error
	errorHandler := func(err error) {
		capturedError = err
	}
	
	bufferedWriter := NewBufferedWriter(&mockWriteError{}, 10, 10*time.Millisecond, errorHandler, 5, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Write some data
	_, err := bufferedWriter.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	
	// Wait for flush to occur
	time.Sleep(50 * time.Millisecond)
	
	// Check if error was handled
	if capturedError == nil {
		t.Error("Expected error to be captured by error handler")
	}
	
	// Close should complete without blocking
	bufferedWriter.Close()
}

// TestBufferedWriterClose tests closing the BufferedWriter
func TestBufferedWriterClose(t *testing.T) {
	var output bytes.Buffer
	errorHandler := func(err error) {}
	
	bufferedWriter := NewBufferedWriter(&output, 10, 100*time.Millisecond, errorHandler, 5, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	
	// Write some data
	_, err := bufferedWriter.Write([]byte("before close"))
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	
	// Close the writer - this should flush remaining data
	err = bufferedWriter.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
	
	// Data should have been written
	result := output.String()
	if !strings.Contains(result, "before close") {
		t.Error("Data written before close should be in output")
	}
	
	// Try closing again - should not cause an error
	err2 := bufferedWriter.Close()
	if err2 != nil {
		t.Errorf("Closing already closed writer returned error: %v", err2)
	}
}

// TestBufferedWriterWithBatches tests batching functionality
func TestBufferedWriterWithBatches(t *testing.T) {
	counter := &mockWriteCounter{}
	
	// Create buffered writer with batching enabled
	bufferedWriter := NewBufferedWriter(counter, 20, 100*time.Millisecond, func(err error) {}, 3, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Write multiple small chunks that should be batched together
	for i := 0; i < 5; i++ {
		data := []byte("chunk " + string(rune(i+'0')) + "\n")
		_, err := bufferedWriter.Write(data)
		if err != nil {
			t.Errorf("Write %d returned error: %v", i, err)
		}
	}
	
	// Wait for batching to occur
	time.Sleep(200 * time.Millisecond)
	
	// Check the output - verify batching occurred
	output := string(counter.GetData())
	if !strings.Contains(output, "chunk 0") || !strings.Contains(output, "chunk 1") || !strings.Contains(output, "chunk 2") {
		t.Error("Batched data should contain all chunks")
	}
	
	// The write counter should be low if batching worked correctly
	writeCount := counter.GetWriteCount()
	t.Logf("Number of actual writes: %d", writeCount)
	// With batchSize=3, we expect at most 2 writes (3 items in first batch, 2 in second)
	if writeCount > 2 {
		t.Log("More writes than expected - batching might not be working optimally")
	}
}

// TestBufferedWriterWithTimeout tests batch timeout functionality
func TestBufferedWriterWithTimeout(t *testing.T) {
	counter := &mockWriteCounter{}
	
	// Create buffered writer with small batch size and timeout
	bufferedWriter := NewBufferedWriter(counter, 20, 10*time.Millisecond, func(err error) {}, 10, 50*time.Millisecond)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Write one chunk - it should be flushed by timeout since batch size won't be reached
	_, err := bufferedWriter.Write([]byte("single chunk\n"))
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	
	// Wait for timeout to occur and flush
	time.Sleep(100 * time.Millisecond)
	
	// Check that data was written despite not reaching batch size
	output := string(counter.GetData())
	if !strings.Contains(output, "single chunk") {
		t.Error("Data should be flushed by timeout even if batch size not reached")
	}
}

// TestBufferedWriterStats tests the Stats method
func TestBufferedWriterStats(t *testing.T) {
	var output bytes.Buffer
	errorHandler := func(err error) {}
	
	bufferedWriter := NewBufferedWriter(&output, 10, 100*time.Millisecond, errorHandler, 5, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Write some data to populate stats
	for i := 0; i < 3; i++ {
		_, _ = bufferedWriter.Write([]byte("stat test"))
	}
	
	// Get stats
	stats := bufferedWriter.Stats()
	if stats == nil {
		t.Fatal("Stats returned nil")
	}
	
	// Verify expected keys exist
	if _, exists := stats["buffer_size"]; !exists {
		t.Error("Stats should contain buffer_size")
	}
	
	if _, exists := stats["current_queue"]; !exists {
		t.Error("Stats should contain current_queue")
	}
	
	if _, exists := stats["dropped_logs"]; !exists {
		t.Error("Stats should contain dropped_logs")
	}
	
	if _, exists := stats["total_logs"]; !exists {
		t.Error("Stats should contain total_logs")
	}
	
	if _, exists := stats["last_flush"]; !exists {
		t.Error("Stats should contain last_flush")
	}
	
	// Verify values are reasonable
	if size, ok := stats["buffer_size"].(int); ok && size != 10 {
		t.Errorf("Expected buffer_size 10, got %d", size)
	}
	
	if total, ok := stats["total_logs"].(int64); ok && total < 3 {
		t.Errorf("Expected total_logs >= 3, got %d", total)
	}
}

// TestBufferedWriterConcurrent tests the BufferedWriter in a concurrent context
func TestBufferedWriterConcurrent(t *testing.T) {
	counter := &mockWriteCounter{}
	
	bufferedWriter := NewBufferedWriter(counter, 50, 50*time.Millisecond, func(err error) {}, 10, 100*time.Millisecond)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Run multiple goroutines that write concurrently
	const numGoroutines = 5
	const writesPerGoroutine = 20
	var wg sync.WaitGroup
	
	startTime := time.Now()
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				data := []byte("goroutine " + string(rune(goroutineID+'0')) + " message " + string(rune(j+'0')) + "\n")
				_, err := bufferedWriter.Write(data)
				if err != nil {
					t.Errorf("Goroutine %d write %d returned error: %v", goroutineID, j, err)
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	// Wait for all writes to be processed
	time.Sleep(200 * time.Millisecond)
	
	// Close to ensure all data is flushed
	bufferedWriter.Close()
	
	// Check results
	totalExpected := numGoroutines * writesPerGoroutine
	output := string(counter.GetData())
	actualCount := strings.Count(output, "goroutine")
	
	if actualCount != totalExpected {
		t.Errorf("Expected %d messages in output, got %d", totalExpected, actualCount)
	}
	
	elapsed := time.Since(startTime)
	t.Logf("Processed %d concurrent writes in %v", totalExpected, elapsed)
}

// TestBufferedWriterWriteAfterClose tests writing after closing
func TestBufferedWriterWriteAfterClose(t *testing.T) {
	var output bytes.Buffer
	errorHandler := func(err error) {}
	
	bufferedWriter := NewBufferedWriter(&output, 10, 100*time.Millisecond, errorHandler, 5, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	
	// Close the writer
	bufferedWriter.Close()
	
	// Try to write after closing - should not panic
	_, err := bufferedWriter.Write([]byte("after close"))
	if err != nil {
		t.Logf("Write after close returned error (expected): %v", err)
	}
}

// TestBufferedWriterLargeWrite tests writing large data
func TestBufferedWriterLargeWrite(t *testing.T) {
	var output bytes.Buffer
	errorHandler := func(err error) {}
	
	bufferedWriter := NewBufferedWriter(&output, 5, 100*time.Millisecond, errorHandler, 3, 1*time.Second)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Write a large chunk of data
	largeData := make([]byte, 1000)
	for i := range largeData {
		largeData[i] = byte('A' + (i % 26))
	}
	
	n, err := bufferedWriter.Write(largeData)
	if err != nil {
		t.Errorf("Write large data returned error: %v", err)
	}
	if n != len(largeData) {
		t.Errorf("Write large data returned %d, expected %d", n, len(largeData))
	}
	
	// Wait for processing
	time.Sleep(150 * time.Millisecond)
	
	result := output.String()
	if len(result) < 900 { // Check that most of the data was written
		t.Errorf("Output seems truncated, got %d chars", len(result))
	}
}

// TestBufferedWriterWithRealFile tests with an in-memory representation
func TestBufferedWriterWithBytesBuffer(t *testing.T) {
	var buf bytes.Buffer
	errorHandler := func(err error) {}
	
	bufferedWriter := NewBufferedWriter(&buf, 20, 20*time.Millisecond, errorHandler, 5, 50*time.Millisecond)
	if bufferedWriter == nil {
		t.Fatal("NewBufferedWriter returned nil")
	}
	defer bufferedWriter.Close()
	
	// Write several pieces of data
	for i := 0; i < 7; i++ {
		_, _ = bufferedWriter.Write([]byte("line " + string(rune(i+'0')) + "\n"))
	}
	
	// Wait for flush
	time.Sleep(100 * time.Millisecond)
	
	// Verify data was written
	output := buf.String()
	if !strings.Contains(output, "line 0") || !strings.Contains(output, "line 6") {
		t.Error("Not all lines were written to buffer")
	}
}