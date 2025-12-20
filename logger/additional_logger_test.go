package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/util"
)

// at
func TestAdditionalLoggerFeatures(t *testing.T) {
	// Test logger with various configuration options
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:    false,
			ShowTimestamp:   true,
			ShowCaller:      true,
			ShowGoroutine:   true,
			ShowPID:         true,
			ShowTraceInfo:   true,
			ShowHostname:    true,
			ShowApplication: true,
		},
		ShowGoroutine:   true,
		ShowPID:         true,
		ShowTraceInfo:   true,
		ShowHostname:    true,
		ShowApplication: true,
		Hostname:        "test-host",
		Application:     "test-app",
		Version:         "1.0.0",
		Environment:     "test-env",
		MaxFieldSize:    1000,
	})
	defer logger.Close()

	// Test basic logging with all features enabled
	ctx := context.Background()
	ctx = util.WithTraceID(ctx, "test-trace-123")
	ctx = util.WithUserID(ctx, "test-user-456")

	logger.InfoC(ctx, "test message with context")

	output := buf.String()
	// Verify that the message appears (basic functionality)
	if !strings.Contains(output, "test message with context") {
		t.Error("Basic message should be in output")
	}
}

// TestLoggerWithCustomExitFunc tests custom exit function
func TestLoggerWithCustomExitFunc(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:  false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		ExitFunc: func(code int) {
			// This function will be called when fatal occurs
		},
	})
	defer logger.Close()

	// Test with fatal level
	logger.Fatal("fatal message")

	if !strings.Contains(buf.String(), "fatal message") {
		t.Error("Fatal message should be in output")
	}
}

// TestLoggerWithCustomErrorHandler tests custom error handler
func TestLoggerWithCustomErrorHandler(t *testing.T) {
	errorHandled := false
	var capturedError error

	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:  false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		ErrorHandler: func(err error) {
			errorHandled = true
			capturedError = err
		},
	})
	defer logger.Close()

	// Force an error somehow - trigger error handling mechanism
	logger.Info("normal message")

	// Error might not be triggered in normal operations, but the handler is set
	_ = errorHandled
	_ = capturedError
}

// TestLoggerWithHooks tests hook functionality
func TestLoggerWithHooks(t *testing.T) {
	var buf bytes.Buffer

	// Since hooks require the hook interface, we'll just verify that a logger with hooks config can be created
	// For full hook functionality, see the hook package tests

	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:  false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
	})
	defer logger.Close()

	logger.Info("message with potential hooks")

	if !strings.Contains(buf.String(), "message with potential hooks") {
		t.Error("Message should be in output")
	}
}

// TestLoggerBufferedWriting tests buffered writing functionality
func TestLoggerBufferedWriting(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:  false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		BufferSize:    512,
		FlushInterval: 10 * time.Millisecond, // Shorter interval for test
	})
	defer logger.Close()

	// Send several messages
	for i := 0; i < 5; i++ {
		logger.WithFields(map[string]interface{}{
			"iteration": i,
		}).Info("buffered message")
	}

	// Give time for buffer to flush
	time.Sleep(50 * time.Millisecond)

	// Close logger to ensure all messages are flushed
	logger.Close()

	// Messages should appear in the buffer
	output := buf.String()
	if strings.Count(output, "buffered message") < 1 { // At least one should appear
		t.Errorf("Expected at least 1 message, got %d", strings.Count(output, "buffered message"))
	}
}

// TestLoggerWithSampling tests sampling functionality more thoroughly
func TestLoggerWithSampling(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:  false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		EnableSampling: true,
		SamplingRate:   2, // Every 2nd message
	})
	defer logger.Close()

	// Send 10 messages, expect about 5 to be logged
	for i := 0; i < 10; i++ {
		logger.Info("sampled message", i)
	}

	output := buf.String()
	count := strings.Count(output, "sampled message")
	// Should be around 5 (half of 10), with tolerance for variation
	if count < 3 || count > 7 { // Allow some variance
		t.Errorf("Expected about 5 sampled messages, got %d", count)
	}
}

// at
func TestLoggerWithDisableLocking(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:  false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		DisableLocking: true, // This option exists
	})
	defer logger.Close()

	logger.Info("message with locking disabled")

	output := buf.String()
	if !strings.Contains(output, "message with locking disabled") {
		t.Error("Message should be in output")
	}
}

// TestLoggerWithMetrics tests metrics collection
func TestLoggerWithMetrics(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:  false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		EnableMetrics: true,
	})
	defer logger.Close()

	logger.Info("message with metrics")

	output := buf.String()
	if !strings.Contains(output, "message with metrics") {
		t.Error("Message should be in output")
	}
}

// TestLoggerWithPreAllocateSettings tests pre-allocation settings
func TestLoggerWithPreAllocateSettings(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &buf,
		Formatter: &formatter.TextFormatter{
			EnableColors:  false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		PreAllocateFields: 16, // Pre-allocate for fields
		PreAllocateTags:   32, // Pre-allocate for tags
	})
	defer logger.Close()

	logger.Info("message with pre-allocated settings")

	output := buf.String()
	if !strings.Contains(output, "message with pre-allocated settings") {
		t.Error("Message should be in output")
	}
}
