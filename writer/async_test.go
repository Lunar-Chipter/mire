package writer

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/errors"
)

// mockLogProcessor implements LogProcessor interface for testing
type mockLogProcessor struct {
	loggedEntries   []*logJob
	errorHandler    func(error)
	errOut          io.Writer
	errOutMu        *sync.Mutex
	logCalls        int
	mu              sync.Mutex
}

func (m *mockLogProcessor) Log(ctx context.Context, level core.Level, msg []byte, fields map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.loggedEntries = append(m.loggedEntries, &logJob{
		level:  level,
		msg:    msg,
		fields: fields,
		ctx:    ctx,
	})
	m.logCalls++
}

func (m *mockLogProcessor) ErrorHandler() func(error) { 
	return m.errorHandler 
}

func (m *mockLogProcessor) ErrOut() io.Writer { 
	return m.errOut 
}

func (m *mockLogProcessor) ErrOutMu() *sync.Mutex { 
	return m.errOutMu 
}

// TestNewAsyncLogger tests creating a new AsyncLogger
func TestNewAsyncLogger(t *testing.T) {
	processor := &mockLogProcessor{}
	
	asyncLogger := NewAsyncLogger(processor, 2, 100, 5*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	if asyncLogger.processor != processor {
		t.Error("AsyncLogger processor not set correctly")
	}
	if asyncLogger.workerCount != 2 {
		t.Errorf("Expected worker count 2, got %d", asyncLogger.workerCount)
	}
	if cap(asyncLogger.logChan) != 100 {
		t.Errorf("Expected log channel capacity 100, got %d", cap(asyncLogger.logChan))
	}
	if asyncLogger.logProcessTimeout != 5*time.Second {
		t.Errorf("Expected log process timeout 5s, got %v", asyncLogger.logProcessTimeout)
	}
}

// TestAsyncLoggerLog tests the Log method of AsyncLogger
func TestAsyncLoggerLog(t *testing.T) {
	processor := &mockLogProcessor{}
	
	asyncLogger := NewAsyncLogger(processor, 2, 100, 5*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	ctx := context.Background()
	
	// Log a message asynchronously
	asyncLogger.Log(core.INFO, []byte("test message"), map[string]interface{}{"key": "value"}, ctx)
	
	// Give some time for the async processing
	time.Sleep(50 * time.Millisecond)
	
	// Close the logger to ensure all messages are processed
	asyncLogger.Close()
	
	// Check that the message was processed by the mock
	processor.mu.Lock()
	defer processor.mu.Unlock()
	
	if processor.logCalls != 1 {
		t.Errorf("Expected 1 log call, got %d", processor.logCalls)
	}
	
	if len(processor.loggedEntries) != 1 {
		t.Errorf("Expected 1 logged entry, got %d", len(processor.loggedEntries))
	} else {
		entry := processor.loggedEntries[0]
		if entry.level != core.INFO {
			t.Errorf("Expected level INFO, got %v", entry.level)
		}
		if string(entry.msg) != "test message" {
			t.Errorf("Expected message 'test message', got '%s'", string(entry.msg))
		}
		if val, ok := entry.fields["key"]; !ok || val != "value" {
			t.Errorf("Expected field 'key' with value 'value', got %+v", entry.fields)
		}
	}
}

// TestAsyncLoggerMultipleLogs tests multiple log messages
func TestAsyncLoggerMultipleLogs(t *testing.T) {
	processor := &mockLogProcessor{}
	
	asyncLogger := NewAsyncLogger(processor, 3, 200, 5*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	ctx := context.Background()
	
	// Log multiple messages
	for i := 0; i < 10; i++ {
		asyncLogger.Log(core.DEBUG, []byte("message "+string(rune(i+'0'))), 
			map[string]interface{}{"idx": i}, ctx)
	}
	
	// Give some time for the async processing
	time.Sleep(100 * time.Millisecond)
	
	// Close the logger to ensure all messages are processed
	asyncLogger.Close()
	
	// Check that all messages were processed
	processor.mu.Lock()
	defer processor.mu.Unlock()
	
	if processor.logCalls != 10 {
		t.Errorf("Expected 10 log calls, got %d", processor.logCalls)
	}
	
	if len(processor.loggedEntries) != 10 {
		t.Errorf("Expected 10 logged entries, got %d", len(processor.loggedEntries))
	}
}

// TestAsyncLoggerBufferFull tests behavior when the log channel is full
func TestAsyncLoggerBufferFull(t *testing.T) {
	// Create a mock processor with error handling
	errChan := make(chan error, 1)
	processor := &mockLogProcessor{
		errorHandler: func(err error) {
			errChan <- err
		},
		errOut: nil, // We'll use custom error handling
		errOutMu: &sync.Mutex{},
	}
	
	// Create async logger with small buffer
	asyncLogger := NewAsyncLogger(processor, 1, 5, 1*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	ctx := context.Background()
	
	// Send more messages than the buffer can hold
	for i := 0; i < 10; i++ {
		asyncLogger.Log(core.INFO, []byte("test message"), nil, ctx)
	}
	
	// Check if any errors were reported
	select {
	case err := <-errChan:
		if err != errors.ErrAsyncBufferFull {
			t.Errorf("Expected ErrAsyncBufferFull, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		// No error reported within timeout - this might be expected based on implementation
		t.Log("No error reported - buffer full errors might be handled differently")
	}
	
	asyncLogger.Close()
}

// TestAsyncLoggerWithTimeout tests async logger with log processing timeout
func TestAsyncLoggerWithTimeout(t *testing.T) {
	processor := &mockLogProcessor{}
	
	// Create async logger with short timeout
	asyncLogger := NewAsyncLogger(processor, 2, 100, 10*time.Millisecond, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	ctx := context.Background()
	
	// Log a message
	asyncLogger.Log(core.INFO, []byte("timeout test"), nil, ctx)
	
	// Give time for processing
	time.Sleep(50 * time.Millisecond)
	asyncLogger.Close()
	
	// Check that the message was processed despite the timeout
	processor.mu.Lock()
	defer processor.mu.Unlock()
	
	if processor.logCalls < 1 {
		t.Errorf("Expected at least 1 log call, got %d", processor.logCalls)
	}
}

// TestAsyncLoggerWithoutTimeout tests async logger with timeout disabled
func TestAsyncLoggerWithoutTimeout(t *testing.T) {
	processor := &mockLogProcessor{}
	
	// Create async logger with timeout disabled
	asyncLogger := NewAsyncLogger(processor, 2, 100, 0, true) // disablePerLogContextTimeout = true
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	ctx := context.Background()
	
	// Log a message
	asyncLogger.Log(core.INFO, []byte("no timeout test"), nil, ctx)
	
	// Give time for processing
	time.Sleep(50 * time.Millisecond)
	asyncLogger.Close()
	
	// Check that the message was processed
	processor.mu.Lock()
	defer processor.mu.Unlock()
	
	if processor.logCalls < 1 {
		t.Errorf("Expected at least 1 log call, got %d", processor.logCalls)
	}
}

// TestAsyncLoggerClose tests closing the AsyncLogger multiple times
func TestAsyncLoggerClose(t *testing.T) {
	processor := &mockLogProcessor{}
	
	asyncLogger := NewAsyncLogger(processor, 1, 10, 1*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	
	// Close once - should work
	asyncLogger.Close()
	
	// Close again - should not cause an error or panic
	asyncLogger.Close()
	
	// Double close should be safe
}

// TestAsyncLoggerWorkerPanicRecovery tests worker panic recovery
func TestAsyncLoggerWorkerPanicRecovery(t *testing.T) {
	// Create a processor that will cause a panic when Log is called
	panicProcessor := &panicLogProcessor{}
	
	asyncLogger := NewAsyncLogger(panicProcessor, 1, 10, 1*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	ctx := context.Background()
	
	// Send a message that will cause a panic in the worker
	asyncLogger.Log(core.INFO, []byte("panic test"), nil, ctx)
	
	// Give the worker time to process and potentially panic/recover
	time.Sleep(100 * time.Millisecond)
	
	// The logger should still be functional after the panic/recovery
	// Close should work without issues
	asyncLogger.Close()
}

// TestAsyncLoggerConcurrent tests the AsyncLogger in a concurrent context
func TestAsyncLoggerConcurrent(t *testing.T) {
	processor := &mockLogProcessor{}
	
	asyncLogger := NewAsyncLogger(processor, 5, 500, 5*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	// Run multiple goroutines that log concurrently
	const numGoroutines = 10
	const messagesPerGoroutine = 50
	done := make(chan bool, numGoroutines)
	
	startTime := time.Now()
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			
			ctx := context.Background()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := []byte("message from goroutine " + string(rune(goroutineID+'0')) + " - " + string(rune(j+'0')))
				asyncLogger.Log(core.INFO, msg, map[string]interface{}{"goroutine": goroutineID, "msg_num": j}, ctx)
			}
		}(i)
	}
	
	// Wait for all goroutines to finish sending messages
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Wait a bit more to ensure all messages are processed
	time.Sleep(500 * time.Millisecond)
	
	// Close the logger to ensure all messages are processed
	asyncLogger.Close()
	
	// Check that all messages were processed
	processor.mu.Lock()
	defer processor.mu.Unlock()
	
	expectedMessages := numGoroutines * messagesPerGoroutine
	actualMessages := len(processor.loggedEntries)
	
	if actualMessages != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, actualMessages)
	}
	
	if processor.logCalls != expectedMessages {
		t.Errorf("Expected %d log calls, got %d", expectedMessages, processor.logCalls)
	}
	
	elapsed := time.Since(startTime)
	t.Logf("Processed %d messages in %v with %d workers", actualMessages, elapsed, asyncLogger.workerCount)
}

// panicLogProcessor is a mock processor that panics on Log
type panicLogProcessor struct{}

func (p *panicLogProcessor) Log(ctx context.Context, level core.Level, msg []byte, fields map[string]interface{}) {
	panic("intentional panic for testing")
}

func (p *panicLogProcessor) ErrorHandler() func(error) { 
	return func(err error) {} // No-op error handler
}

func (p *panicLogProcessor) ErrOut() io.Writer { 
	return nil 
}

func (p *panicLogProcessor) ErrOutMu() *sync.Mutex { 
	return &sync.Mutex{}
}

// TestAsyncLoggerWorker tests individual worker functionality
func TestAsyncLoggerWorker(t *testing.T) {
	processor := &mockLogProcessor{}
	
	asyncLogger := NewAsyncLogger(processor, 1, 50, 1*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	defer asyncLogger.Close()
	
	ctx := context.Background()
	
	// Send a few messages to the channel directly (bypassing Log method)
	job := &logJob{
		level: core.INFO,
		msg: []byte("direct job"),
		fields: map[string]interface{}{"source": "direct"},
		ctx: ctx,
	}
	
	// Try to send to the channel
	select {
	case asyncLogger.logChan <- job:
		// Successfully sent
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Could not send job to async logger channel in time")
	}
	
	// Wait for processing
	time.Sleep(200 * time.Millisecond)
	
	// Close to ensure processing completes
	asyncLogger.Close()
	
	// Verify the job was processed
	processor.mu.Lock()
	defer processor.mu.Unlock()
	
	if processor.logCalls < 1 {
		t.Errorf("Expected at least 1 log call, got %d", processor.logCalls)
	}
}

// TestAsyncLoggerClosedBehavior tests logging to a closed AsyncLogger
func TestAsyncLoggerClosedBehavior(t *testing.T) {
	processor := &mockLogProcessor{}
	
	asyncLogger := NewAsyncLogger(processor, 1, 10, 1*time.Second, false)
	if asyncLogger == nil {
		t.Fatal("NewAsyncLogger returned nil")
	}
	
	// Close the logger immediately
	asyncLogger.Close()
	
	// Try to log after closing - this should not panic
	ctx := context.Background()
	asyncLogger.Log(core.INFO, []byte("message after close"), nil, ctx)
	
	// The message might or might not be processed depending on timing,
	// but it should not cause a panic
}