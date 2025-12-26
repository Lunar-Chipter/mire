// Package example demonstrates best practices for using Mire logging library
package example

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/logger"
	"github.com/Lunar-Chipter/mire/util"
)

// ZeroAllocExample demonstrates high-performance, zero-allocation logging
func ZeroAllocExample() {
	// Create logger with zero-allocation formatter
	log := logger.New(logger.LoggerConfig{
		Level:  core.INFO,
		Output: os.Stdout,
		Formatter: &formatter.TextFormatter{
			EnableColors:  true,
			ShowTimestamp: true,
			ShowCaller:    true,
		},
	})
	defer log.Close()

	// Unified API (Recommended)

	// Basic logging
	log.Log(context.Background(), core.INFO, []byte("Application started"), nil)
	log.Log(context.Background(), core.DEBUG, []byte("Processing request"), nil)

	// Logging with fields
	userLogger := log.WithFields(map[string]interface{}{
		"service": "user-api",
		"version": "1.0.0",
	})
	userLogger.Log(context.Background(), core.INFO, []byte("User authenticated"), nil)

	// Context-aware logging
	ctx := context.Background()
	ctx = util.WithTraceID(ctx, "trace-12345")

	log.LogC(ctx, core.INFO, []byte("Payment processed"))

	// Legacy API (still works but allocates):
	log.Info("User logged in with legacy API")
}

// AppLogger demonstrates how to create a well-configured logger for an application
type AppLogger struct {
	logger *logger.Logger
}

// NewAppLogger creates a new instance of AppLogger with recommended configuration
func NewAppLogger(output io.Writer, level core.Level, environment string) *AppLogger {
	config := logger.LoggerConfig{
		Level:            level,
		Output:           output,
		ErrorOutput:      os.Stderr,
		CallerDepth:      logger.DEFAULT_CALLER_DEPTH,
		TimestampFormat:  logger.DEFAULT_TIMESTAMP_FORMAT,
		BufferSize:       logger.DEFAULT_BUFFER_SIZE,
		FlushInterval:    logger.DEFAULT_FLUSH_INTERVAL,
		ClockInterval:    10 * time.Millisecond,
		MaskValue:  "[MASKED]",
		Environment:      environment,
		Hostname:         os.Getenv("HOSTNAME"),
		Application:      "mire-example-app",
		Version:          "1.0.0",
	}

	// Set formatter based on output type
	if output == os.Stdout || output == os.Stderr {
		config.Formatter = &formatter.TextFormatter{
			EnableColors:      true,
			ShowTimestamp:     true,
			ShowCaller:        true,
			ShowTraceInfo:     true,
			ShowPID:           true,
			TimestampFormat:   logger.DEFAULT_TIMESTAMP_FORMAT,
			SensitiveFields:   []string{"password", "token", "secret"},
			MaskSensitiveData: true,
		}
	} else {
		config.Formatter = &formatter.JSONFormatter{
			PrettyPrint:       false,
			TimestampFormat:   logger.DEFAULT_TIMESTAMP_FORMAT,
			ShowCaller:        true,
			ShowTrace:     true,
			IncludeStackTrace:  true,
			SensitiveFields:   []string{"password", "token", "secret"},
			MaskSensitiveData: true,
		}
	}

	return &AppLogger{
		logger: logger.New(config),
	}
}

// Close closes the application logger
func (al *AppLogger) Close() error {
	if al.logger != nil {
		al.logger.Close()
	}
	return nil
}

// LogRequest logs request-related information with context
func (al *AppLogger) LogRequest(ctx context.Context, method, path string, duration time.Duration) {
	al.logger.WithFields(map[string]interface{}{
		"method":   method,
		"path":     path,
		"duration": duration.Milliseconds(),
	}).InfoC(ctx, "HTTP request processed")
}

// LogError logs error with context and stack trace
func (al *AppLogger) LogError(ctx context.Context, operation string, err error) {
	al.logger.WithFields(map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
	}).ErrorC(ctx, "Operation failed")
}

// LogUserAction logs user-specific actions with user context
func (al *AppLogger) LogUserAction(ctx context.Context, action string, userID string) {
	ctx = util.WithUserID(ctx, userID)
	al.logger.WithFields(map[string]interface{}{
		"action":  action,
		"user_id": userID,
	}).InfoC(ctx, "User action recorded")
}

// LogTransaction logs financial or business transaction
func (al *AppLogger) LogTransaction(ctx context.Context, transactionID string, amount float64, currency string) {
	al.logger.WithFields(map[string]interface{}{
		"transaction_id": transactionID,
		"amount":         amount,
		"currency":       currency,
	}).InfoC(ctx, "Transaction processed")
}

// LogSecurityEvent logs security-related events
func (al *AppLogger) LogSecurityEvent(ctx context.Context, eventType, description string, userID string) {
	ctx = util.WithUserID(ctx, userID)
	al.logger.WithFields(map[string]interface{}{
		"event_type":  eventType,
		"description": description,
		"user_id":     userID,
		"severity":    "high",
	}).WarnC(ctx, "Security event detected")
}

// Example usage of the AppLogger
func ExampleUsage() {
	// Create application logger
	appLogger := NewAppLogger(os.Stdout, core.INFO, "production")
	defer func() {
	_ = appLogger.Close()
}()

	// Simulate context with trace ID
	ctx := context.Background()
	ctx = util.WithTraceID(ctx, "trace-12345")
	ctx = util.WithRequestID(ctx, "req-67890")

	// at
	appLogger.LogRequest(ctx, "GET", "/api/users/123", 150*time.Millisecond)
	appLogger.LogUserAction(ctx, "view_profile", "user-123")
	appLogger.LogTransaction(ctx, "txn-98765", 99.99, "USD")

	// Simulate and log an error
	simulatedError := &os.PathError{Op: "open", Path: "/invalid/path", Err: os.ErrNotExist}
	appLogger.LogError(ctx, "file_operation", simulatedError)

	// Log a security event
	appLogger.LogSecurityEvent(ctx, "failed_login_attempt", "Multiple failed attempts from same IP", "user-456")
}

// NewFastLogger creates a logger optimized for high throughput
func NewFastLogger() *logger.Logger {
	config := logger.LoggerConfig{
		Level:             core.INFO,
		Output:            os.Stdout,
		AsyncMode:             true,
		ChannelSize:       10000,
		ProcessTimeout:  5 * time.Second,
		NoTimeout:      true,
		BufferSize:        8192,
		FlushInterval:     100 * time.Millisecond,
		NoLocking:            true,
		IncludeStackTrace:        false,
		Formatter: &formatter.CSVFormatter{
			IncludeHeader:     false,
			SensitiveFields:   []string{"password", "token"},
			MaskSensitiveData: true,
		},
	}

	return logger.New(config)
}
