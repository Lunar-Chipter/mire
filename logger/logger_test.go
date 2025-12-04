package logger

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/hook"
)

// TestNewDefaultLogger tests creating a default logger
func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	if logger == nil {
		t.Fatal("NewDefaultLogger returned nil")
	}
	defer logger.Close()

	// Check default configuration
	if logger.Config.Level != core.INFO {
		t.Errorf("Default logger level should be INFO, got %v", logger.Config.Level)
	}
	if logger.Config.Output != os.Stdout {
		t.Error("Default logger output should be os.Stdout")
	}
	if logger.Config.ErrorOutput != os.Stderr {
		t.Error("Default logger error output should be os.Stderr")
	}
	if logger.Config.CallerDepth != DEFAULT_CALLER_DEPTH {
		t.Errorf("Default logger caller depth should be %d, got %d", DEFAULT_CALLER_DEPTH, logger.Config.CallerDepth)
	}
	if logger.Config.TimestampFormat != DEFAULT_TIMESTAMP_FORMAT {
		t.Errorf("Default logger timestamp format should be %s, got %s", DEFAULT_TIMESTAMP_FORMAT, logger.Config.TimestampFormat)
	}
	if logger.Config.BufferSize != DEFAULT_BUFFER_SIZE {
		t.Errorf("Default logger buffer size should be %d, got %d", DEFAULT_BUFFER_SIZE, logger.Config.BufferSize)
	}
	if logger.Config.FlushInterval != DEFAULT_FLUSH_INTERVAL {
		t.Errorf("Default logger flush interval should be %v, got %v", DEFAULT_FLUSH_INTERVAL, logger.Config.FlushInterval)
	}

	// Check that formatter is properly configured
	textFormatter, ok := logger.formatter.(*formatter.TextFormatter)
	if !ok {
		t.Error("Default logger should use TextFormatter")
	} else {
		if !textFormatter.EnableColors {
			t.Error("Default TextFormatter should have colors enabled")
		}
		if !textFormatter.ShowTimestamp {
			t.Error("Default TextFormatter should show timestamp")
		}
		if !textFormatter.ShowCaller {
			t.Error("Default TextFormatter should show caller")
		}
		if textFormatter.TimestampFormat != DEFAULT_TIMESTAMP_FORMAT {
			t.Errorf("Default TextFormatter timestamp format should be %s, got %s", DEFAULT_TIMESTAMP_FORMAT, textFormatter.TimestampFormat)
		}
	}
}

// TestNewLogger tests creating a logger with custom configuration
func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	config := LoggerConfig{
		Level:             core.DEBUG,
		Output:            &buf,
		ErrorOutput:       &buf,
		Formatter:         &formatter.TextFormatter{},
		ShowCaller:        true,
		CallerDepth:       2,
		ShowGoroutine:     true,
		ShowPID:           true,
		ShowTraceInfo:     true,
		ShowHostname:      true,
		ShowApplication:   true,
		TimestampFormat:   "2006-01-02",
		EnableStackTrace:  true,
		StackTraceDepth:   10,
		EnableSampling:    true,
		SamplingRate:      5,
		BufferSize:        500,
		FlushInterval:     2 * time.Second,
		MaxFieldSize:      1000,
		EnableMetrics:     true,
		MaxMessageSize:    2000,
		AsyncLogging:      true,
		AsyncWorkerCount:  2,
		AsyncLogChannelBufferSize: 100,
		ClockInterval:     5 * time.Millisecond,
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	defer logger.Close()

	// Verify configuration
	if logger.Config.Level != core.DEBUG {
		t.Errorf("Logger level should be DEBUG, got %v", logger.Config.Level)
	}
	if logger.Config.Output != &buf {
		t.Error("Logger output should be the provided buffer")
	}
	if logger.Config.ErrorOutput != &buf {
		t.Error("Logger error output should be the provided buffer")
	}
	if logger.Config.ShowCaller != true {
		t.Error("Logger should show caller")
	}
	if logger.Config.CallerDepth != 2 {
		t.Errorf("Logger caller depth should be 2, got %d", logger.Config.CallerDepth)
	}
	if logger.Config.TimestampFormat != "2006-01-02" {
		t.Errorf("Logger timestamp format should be '2006-01-02', got %s", logger.Config.TimestampFormat)
	}
	if logger.Config.EnableStackTrace != true {
		t.Error("Logger should enable stack trace")
	}
	if logger.Config.StackTraceDepth != 10 {
		t.Errorf("Logger stack trace depth should be 10, got %d", logger.Config.StackTraceDepth)
	}
	if logger.Config.EnableSampling != true {
		t.Error("Logger should enable sampling")
	}
	if logger.Config.SamplingRate != 5 {
		t.Errorf("Logger sampling rate should be 5, got %d", logger.Config.SamplingRate)
	}
	if logger.Config.BufferSize != 500 {
		t.Errorf("Logger buffer size should be 500, got %d", logger.Config.BufferSize)
	}
	if logger.Config.FlushInterval != 2*time.Second {
		t.Errorf("Logger flush interval should be 2s, got %v", logger.Config.FlushInterval)
	}
	if logger.Config.MaxFieldSize != 1000 {
		t.Errorf("Logger max field size should be 1000, got %d", logger.Config.MaxFieldSize)
	}
	if logger.Config.EnableMetrics != true {
		t.Error("Logger should enable metrics")
	}
	if logger.Config.MaxMessageSize != 2000 {
		t.Errorf("Logger max message size should be 2000, got %d", logger.Config.MaxMessageSize)
	}
	if logger.Config.AsyncLogging != true {
		t.Error("Logger should enable async logging")
	}
	if logger.Config.AsyncWorkerCount != 2 {
		t.Errorf("Logger async worker count should be 2, got %d", logger.Config.AsyncWorkerCount)
	}
	if logger.Config.AsyncLogChannelBufferSize != 100 {
		t.Errorf("Logger async channel buffer size should be 100, got %d", logger.Config.AsyncLogChannelBufferSize)
	}
	if logger.Config.ClockInterval != 5*time.Millisecond {
		t.Errorf("Logger clock interval should be 5ms, got %v", logger.Config.ClockInterval)
	}

	// Verify that sampler was created
	if logger.sampler == nil {
		t.Error("Sampler should be created when EnableSampling=true and SamplingRate > 1")
	}

	// Verify that async logger was created
	if logger.asyncLogger == nil {
		t.Error("Async logger should be created when AsyncLogging=true")
	}
}

// TestLoggerWithFields tests adding fields to a logger
func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	// Add fields
	fields := map[string]interface{}{"key1": "value1", "key2": 123}
	newLogger := logger.WithFields(fields)

	// Verify that the new logger has the additional fields
	if len(newLogger.fields) != 2 {
		t.Errorf("New logger should have 2 fields, got %d", len(newLogger.fields))
	}
	if newLogger.fields["key1"] != "value1" {
		t.Errorf("New logger field key1 should be 'value1', got %v", newLogger.fields["key1"])
	}
	if newLogger.fields["key2"] != 123 {
		t.Errorf("New logger field key2 should be 123, got %v", newLogger.fields["key2"])
	}

	// Verify that the original logger is not affected
	if len(logger.fields) != 0 {
		t.Errorf("Original logger should have 0 fields, got %d", len(logger.fields))
	}
}

// TestLoggerLogMethods tests all the logging methods
func TestLoggerLogMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.TRACE,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	// Test all log methods
	logger.Trace("trace", "message")
	logger.Debug("debug", "message")
	logger.Info("info", "message") 
	logger.Notice("notice", "message")
	logger.Warn("warn", "message")
	logger.Error("error", "message")
	logger.Fatal("fatal", "message")
	logger.Panic("panic", "message")

	// At this point, we expect the buffer to have been written to
	// (though the exact content depends on the formatter)
	output := buf.String()
	if len(output) == 0 {
		t.Error("No output was written to buffer")
	}
}

// TestLoggerLogMethodsFormatted tests all the formatted logging methods
func TestLoggerLogMethodsFormatted(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.TRACE,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	// Test all formatted log methods
	logger.Tracef("trace: %s", "formatted")
	logger.Debugf("debug: %s", "formatted")
	logger.Infof("info: %s", "formatted")
	logger.Noticef("notice: %s", "formatted")
	logger.Warnf("warn: %s", "formatted")
	logger.Errorf("error: %s", "formatted")
	logger.Fatalf("fatal: %s", "formatted")
	logger.Panicf("panic: %s", "formatted")

	// Check that output was written
	output := buf.String()
	if len(output) == 0 {
		t.Error("No output was written to buffer")
	}
}

// TestLoggerContextMethods tests all the context-aware logging methods
func TestLoggerContextMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.TRACE,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	ctx := context.Background()

	// Test all context-aware log methods
	logger.TraceC(ctx, "trace", "context", "message")
	logger.DebugC(ctx, "debug", "context", "message")
	logger.InfoC(ctx, "info", "context", "message")
	logger.NoticeC(ctx, "notice", "context", "message")
	logger.WarnC(ctx, "warn", "context", "message")
	logger.ErrorC(ctx, "error", "context", "message")
	logger.FatalC(ctx, "fatal", "context", "message")
	logger.PanicC(ctx, "panic", "context", "message")

	// Check that output was written
	output := buf.String()
	if len(output) == 0 {
		t.Error("No output was written to buffer")
	}
}

// TestLoggerContextMethodsFormatted tests all the context-aware formatted logging methods
func TestLoggerContextMethodsFormatted(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.TRACE,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	ctx := context.Background()

	// Test all context-aware formatted log methods
	logger.TracefC(ctx, "trace: %s with context", "formatted")
	logger.DebugfC(ctx, "debug: %s with context", "formatted")
	logger.InfofC(ctx, "info: %s with context", "formatted")
	logger.NoticefC(ctx, "notice: %s with context", "formatted")
	logger.WarnfC(ctx, "warn: %s with context", "formatted")
	logger.ErrorfC(ctx, "error: %s with context", "formatted")
	logger.FatalfC(ctx, "fatal: %s with context", "formatted")
	logger.PanicfC(ctx, "panic: %s with context", "formatted")

	// Check that output was written
	output := buf.String()
	if len(output) == 0 {
		t.Error("No output was written to buffer")
	}
}

// TestLoggerLevelFiltering tests that logs are filtered by level
func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.WARN,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	// Log messages at different levels
	logger.Info("This should not appear")
	logger.Warn("This should appear")
	logger.Error("This should appear")

	output := buf.String()
	
	// Should contain "WARN" and "ERROR" messages but not "INFO"
	if !strings.Contains(output, "This should appear") {
		t.Error("Expected 'This should appear' messages were not found")
	}
	if strings.Contains(output, "This should not appear") {
		t.Error("INFO message was not filtered out as expected")
	}
}

// TestLoggerWithHooks tests logger with hooks
func TestLoggerWithHooks(t *testing.T) {
	var buf bytes.Buffer
	mockHook := &mockHook{}
	
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
		Hooks:   []hook.Hook{mockHook},
	})
	defer logger.Close()

	// Log a message that should trigger the hook
	logger.Info("test message with hook")

	// Verify that the hook was called
	if !mockHook.called {
		t.Error("Hook was not called")
	}
	if mockHook.lastEntry == nil {
		t.Error("Hook was called but received nil entry")
	} else if string(mockHook.lastEntry.Message) != "test message with hook" {
		t.Errorf("Hook received message '%s', expected 'test message with hook'", string(mockHook.lastEntry.Message))
	}
}

// TestAddHook tests adding a hook after logger creation
func TestAddHook(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	mockHook := &mockHook{}
	
	// Add hook after creation
	logger.AddHook(mockHook)

	// Log a message
	logger.Info("test with added hook")

	// Verify that the hook was called
	if !mockHook.called {
		t.Error("Added hook was not called")
	}
}

// TestLoggerClose tests closing the logger
func TestLoggerClose(t *testing.T) {
	logger := NewDefaultLogger()
	
	// Close the logger
	logger.Close() // Close doesn't return a value

	// Try to close again - should not cause issues
	logger.Close() // Close doesn't return a value
}

// TestLoggerClosedBehavior tests that logging after closing doesn't panic
func TestLoggerClosedBehavior(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	
	logger.Close()
	
	// These calls should not panic even though logger is closed
	logger.Info("message after close")
	logger.Error("error after close")
	
	// Output buffer should still be empty
	if buf.Len() > 0 {
		t.Error("Output buffer should be empty after logger is closed")
	}
}

// TestFormatArgsToBytes tests the formatArgsToBytes method
func TestFormatArgsToBytes(t *testing.T) {
	logger := NewDefaultLogger()
	defer logger.Close()

	// Test with string arguments
	result := logger.formatArgsToBytes("hello", "world")
	expected := "hello world"
	if string(result) != expected {
		t.Errorf("formatArgsToBytes('hello', 'world') = %s, want %s", string(result), expected)
	}

	// Test with mixed types
	result = logger.formatArgsToBytes("number:", 42, "bool:", true)
	expected = "number: 42 bool: true"
	if string(result) != expected {
		t.Errorf("formatArgsToBytes with mixed types = %s, want %s", string(result), expected)
	}

	// Test with single argument
	result = logger.formatArgsToBytes("single")
	expected = "single"
	if string(result) != expected {
		t.Errorf("formatArgsToBytes('single') = %s, want %s", string(result), expected)
	}

	// Test with empty arguments
	result = logger.formatArgsToBytes()
	expected = ""
	if string(result) != expected {
		t.Errorf("formatArgsToBytes() = %s, want %s", string(result), expected)
	}
}

// TestFormatfArgsToBytes tests the formatfArgsToBytes method
func TestFormatfArgsToBytes(t *testing.T) {
	logger := NewDefaultLogger()
	defer logger.Close()

	// Test with format string and arguments
	result := logger.formatfArgsToBytes("Hello %s, you are %d years old", "Alice", 30)
	expected := "Hello Alice, you are 30 years old"
	if string(result) != expected {
		t.Errorf("formatfArgsToBytes = %s, want %s", string(result), expected)
	}

	// Test with percent sign
	result = logger.formatfArgsToBytes("Discount is %d%%", 20)
	expected = "Discount is 20%%"
	if string(result) != expected {
		t.Errorf("formatfArgsToBytes with percent = %s, want %s", string(result), expected)
	}

	// Test with no arguments
	result = logger.formatfArgsToBytes("Simple string without args")
	expected = "Simple string without args"
	if string(result) != expected {
		t.Errorf("formatfArgsToBytes without args = %s, want %s", string(result), expected)
	}
}

// TestLoggerStats tests the logger statistics
func TestLoggerStats(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	// Log a few messages to increment stats
	logger.Info("message 1")
	logger.Warn("message 2")
	logger.Error("message 3")

	// Get stats
	stats := logger.stats.GetStats()
	if stats == nil {
		t.Fatal("GetStats returned nil")
	}

	logCounts, ok := stats["log_counts"].(map[string]interface{})
	if !ok {
		t.Fatal("log_counts not found in stats or not a map")
	}

	if infoCount, ok := logCounts["INFO"].(int64); !ok || infoCount != 1 {
		t.Errorf("INFO count should be 1, got %v", infoCount)
	}
	if warnCount, ok := logCounts["WARN"].(int64); !ok || warnCount != 1 {
		t.Errorf("WARN count should be 1, got %v", warnCount)
	}
	if errorCount, ok := logCounts["ERROR"].(int64); !ok || errorCount != 1 {
		t.Errorf("ERROR count should be 1, got %v", errorCount)
	}
}

// TestLoggerErrorHandler tests custom error handling
func TestLoggerErrorHandler(t *testing.T) {
	var buf bytes.Buffer
	var errorHandled error
	
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
		ErrorHandler: func(err error) {
			errorHandled = err
		},
	})
	defer logger.Close()

	// Trigger an error by using an invalid formatter
	logger.formatter = &failingFormatter{}
	
	// This should trigger the error handler
	logger.Info("test message with failing formatter")
	
	// Check if error was handled
	if errorHandled == nil {
		t.Error("Error was not handled by custom error handler")
	}
}

// TestLoggerWithSampling tests logger with sampling enabled
func TestLoggerWithSampling(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
		EnableSampling: true,
		SamplingRate: 3, // Log every 3rd message
	})
	defer logger.Close()

	// Log 10 messages, we expect approximately 3-4 to pass through
	for i := 0; i < 10; i++ {
		logger.Info("sampled message", i)
	}

	// Output should have fewer than 10 messages due to sampling
	output := buf.String()
	messageCount := strings.Count(output, "sampled message")
	if messageCount > 5 { // Allow some variance due to sampling
		t.Errorf("Expected fewer sampled messages, got %d", messageCount)
	}
}

// TestLoggerClone tests the clone functionality
func TestLoggerClone(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	// Add a field to the original logger
	logger.fields["original"] = "value"

	// Clone the logger
	cloned := logger.clone()

	// Verify the clone has the same field
	if cloned.fields["original"] != "value" {
		t.Error("Clone does not have the original field")
	}

	// Add a field to the clone
	cloned.fields["cloned"] = "value"

	// Verify the original logger is not affected
	if logger.fields["cloned"] != nil {
		t.Error("Original logger was affected by clone changes")
	}

	// Verify the clone has its own field
	if cloned.fields["cloned"] != "value" {
		t.Error("Clone does not have its own field")
	}
}

// TestLoggerConcurrent tests the logger in a concurrent context
func TestLoggerConcurrent(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
	})
	defer logger.Close()

	// Run multiple goroutines that log concurrently
	const numGoroutines = 10
	const messagesPerGoroutine = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("Concurrent message", goroutineID, j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify that all messages were logged without data races
	output := buf.String()
	if strings.Count(output, "Concurrent message") != numGoroutines*messagesPerGoroutine {
		t.Errorf("Expected %d messages, got %d", numGoroutines*messagesPerGoroutine, strings.Count(output, "Concurrent message"))
	}
}

// TestLoggerWithClock tests the logger with clock functionality
func TestLoggerWithClock(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
		ClockInterval: 10 * time.Millisecond,
	})
	defer logger.Close()

	// Verify that clock was created
	if logger.clock == nil {
		t.Error("Logger should have created a clock when ClockInterval > 0")
	}

	// Log a message to verify it works with clock
	logger.Info("Message with clock")
	
	output := buf.String()
	if !strings.Contains(output, "Message with clock") {
		t.Error("Message was not logged when using clock")
	}
}

// TestNewDefaultLoggerWithMaskString tests default logger with mask string
func TestNewDefaultLoggerWithMaskString(t *testing.T) {
	logger := NewDefaultLogger()
	defer logger.Close()

	// Check that the mask string is set correctly in the default formatter
	if textFormatter, ok := logger.formatter.(*formatter.TextFormatter); ok {
		expectedMask := []byte("[MASKED]")
		if !bytes.Equal(textFormatter.MaskStringBytes, expectedMask) {
			t.Errorf("Default TextFormatter MaskStringBytes should be %v, got %v", expectedMask, textFormatter.MaskStringBytes)
		}
	} else {
		t.Error("Default logger should have a TextFormatter")
	}
}

// TestLoggerWithCustomMaskString tests logger with custom mask string
func TestLoggerWithCustomMaskString(t *testing.T) {
	var buf bytes.Buffer
	customMask := "XXX"
	logger := New(LoggerConfig{
		Level:   core.INFO,
		Output:  &buf,
		Formatter: &formatter.TextFormatter{},
		MaskStringValue: customMask,
	})
	defer logger.Close()

	// Check that the mask string is set correctly
	if textFormatter, ok := logger.formatter.(*formatter.TextFormatter); ok {
		expectedMask := []byte(customMask)
		if !bytes.Equal(textFormatter.MaskStringBytes, expectedMask) {
			t.Errorf("TextFormatter MaskStringBytes should be %v, got %v", expectedMask, textFormatter.MaskStringBytes)
		}
	} else {
		t.Error("Logger should have a TextFormatter")
	}
}

// mockHook is a mock implementation of the Hook interface for testing
type mockHook struct {
	called     bool
	lastEntry  *core.LogEntry
	fireErrors []error
}

func (h *mockHook) Fire(entry *core.LogEntry) error {
	h.called = true
	h.lastEntry = entry
	
	if len(h.fireErrors) > 0 {
		err := h.fireErrors[0]
		h.fireErrors = h.fireErrors[1:]
		return err
	}
	
	return nil
}

func (h *mockHook) Close() error {
	return nil
}

// failingFormatter is a mock formatter that always returns an error
type failingFormatter struct{}

func (f *failingFormatter) Format(buf *bytes.Buffer, entry *core.LogEntry) error {
	return errors.New("formatter failed")
}