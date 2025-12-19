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

// ApplicationLogger demonstrates how to create a well-configured logger for an application
type ApplicationLogger struct {
	logger *logger.Logger
}

// NewApplicationLogger creates a new instance of ApplicationLogger with recommended configuration
func NewApplicationLogger(output io.Writer, level core.Level, environment string) *ApplicationLogger {
	config := logger.LoggerConfig{
		Level:           level,
		Output:          output,
		ErrorOutput:     os.Stderr,
		CallerDepth:     logger.DEFAULT_CALLER_DEPTH,
		TimestampFormat: logger.DEFAULT_TIMESTAMP_FORMAT,
		BufferSize:      logger.DEFAULT_BUFFER_SIZE,
		FlushInterval:   logger.DEFAULT_FLUSH_INTERVAL,
		AsyncWorkerCount: 4,
		ClockInterval:   10 * time.Millisecond,
		MaskStringValue: "[MASKED]",
		Environment:     environment,
		Hostname:        os.Getenv("HOSTNAME"),
		Application:     "mire-example-app",
		Version:         "1.0.0",
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
			ShowTraceInfo:     true,
			EnableStackTrace:  true,
			SensitiveFields:   []string{"password", "token", "secret"},
			MaskSensitiveData: true,
		}
	}

	return &ApplicationLogger{
		logger: logger.New(config),
	}
}

// Close closes the application logger
func (al *ApplicationLogger) Close() error {
	if al.logger != nil {
		al.logger.Close()
	}
	return nil
}

// LogRequest logs request-related information with context
func (al *ApplicationLogger) LogRequest(ctx context.Context, method, path string, duration time.Duration) {
	al.logger.WithFields(map[string]interface{}{
		"method":   method,
		"path":     path,
		"duration": duration.Milliseconds(),
	}).InfoC(ctx, "HTTP request processed")
}

// LogError logs error with context and stack trace
func (al *ApplicationLogger) LogError(ctx context.Context, operation string, err error) {
	al.logger.WithFields(map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
	}).ErrorC(ctx, "Operation failed")
}

// LogUserAction logs user-specific actions with user context
func (al *ApplicationLogger) LogUserAction(ctx context.Context, action string, userID string) {
	ctx = util.WithUserID(ctx, userID)
	al.logger.WithFields(map[string]interface{}{
		"action":  action,
		"user_id": userID,
	}).InfoC(ctx, "User action recorded")
}

// LogTransaction logs financial or business transaction
func (al *ApplicationLogger) LogTransaction(ctx context.Context, transactionID string, amount float64, currency string) {
	al.logger.WithFields(map[string]interface{}{
		"transaction_id": transactionID,
		"amount":         amount,
		"currency":       currency,
	}).InfoC(ctx, "Transaction processed")
}

// LogSecurityEvent logs security-related events
func (al *ApplicationLogger) LogSecurityEvent(ctx context.Context, eventType, description string, userID string) {
	ctx = util.WithUserID(ctx, userID)
	al.logger.WithFields(map[string]interface{}{
		"event_type":  eventType,
		"description": description,
		"user_id":     userID,
		"severity":    "high",
	}).WarnC(ctx, "Security event detected")
}

// Example usage of the ApplicationLogger
func ExampleUsage() {
	// Create application logger
	appLogger := NewApplicationLogger(os.Stdout, core.INFO, "production")
	defer func() { appLogger.Close() }()

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

// HighPerformanceLogger demonstrates configuration for high-volume logging
type HighPerformanceLogger struct {
	logger *logger.Logger
}

// NewHighPerformanceLogger creates a logger optimized for high throughput
func NewHighPerformanceLogger() *HighPerformanceLogger {
	config := logger.LoggerConfig{
		Level:                         core.INFO,
		Output:                        os.Stdout,
		AsyncLogging:                  true,
		AsyncWorkerCount:              8,
		AsyncLogChannelBufferSize:     10000,
		LogProcessTimeout:             5 * time.Second,
		DisablePerLogContextTimeout:   true,
		BufferSize:                    8192,
		FlushInterval:                 100 * time.Millisecond,
		DisableLocking:                true,
		EnableStackTrace:              false, // Disable for performance
		Formatter: &formatter.CSVFormatter{
			IncludeHeader:   false,
			SensitiveFields: []string{"password", "token"},
			MaskSensitiveData: true,
		},
	}

	return &HighPerformanceLogger{
		logger: logger.New(config),
	}
}

// LogEvent logs an event using the high-performance logger
func (hpl *HighPerformanceLogger) LogEvent(eventType, message string, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"event_type": eventType,
		"message":    message,
	}
	for k, v := range fields {
		allFields[k] = v
	}

	hpl.logger.WithFields(allFields).Info()
}

// Close closes the high-performance logger
func (hpl *HighPerformanceLogger) Close() error {
	if hpl.logger != nil {
		hpl.logger.Close()
	}
	return nil
}