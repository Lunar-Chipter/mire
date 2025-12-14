package example

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/hook"
	"github.com/Lunar-Chipter/mire/logger"
	"github.com/Lunar-Chipter/mire/util"
)

// AdvancedExample demonstrates more complex usage patterns of Mire
func AdvancedExample() {
	fmt.Println("=== Advanced Mire Usage Examples ===")

	// at
	contextAwareExample()

	// Example 2: Custom formatter with field transformers
	customFormatterExample()

	// Example 3: Hook implementation for external services
	hookExample()

	// Example 4: Performance optimization techniques
	performanceExample()

	fmt.Println("Advanced examples completed")
}

func contextAwareExample() {
	fmt.Println("\n--- Context-Aware Logging Example ---")

	// Create a context with various identifiers
	ctx := context.Background()
	ctx = util.WithTraceID(ctx, "trace-abc-123")
	ctx = util.WithSpanID(ctx, "span-xyz-789")
	ctx = util.WithUserID(ctx, "user-john-doe")
	ctx = util.WithRequestID(ctx, "req-456-def")

	// Create logger with context-aware capabilities
	log := logger.New(logger.LoggerConfig{
		Level:   core.DEBUG,
		Output:  os.Stdout,
		Formatter: &formatter.JSONFormatter{
			TimestampFormat:  logger.DEFAULT_TIMESTAMP_FORMAT,
			ShowTraceInfo:    true,
			EnableStackTrace: true,
		},
	})
	defer log.Close()

	log.InfoC(ctx, "Processing user request") // Will include trace_id, user_id, etc.

	// Add more fields
	log.WithFields(map[string]interface{}{
		"service": "auth-service",
		"action":  "login",
	}).InfofC(ctx, "User %s authenticated successfully", "john-doe")
}

func customFormatterExample() {
	fmt.Println("\n--- Custom Formatter with Field Transformers Example ---")

	// Create a formatter with custom field transformers
	jsonFormatter := &formatter.JSONFormatter{
		TimestampFormat: logger.DEFAULT_TIMESTAMP_FORMAT,
		ShowCaller:    true,
		FieldTransformers: map[string]func(interface{}) interface{}{
			"credit_card": func(v interface{}) interface{} {
				if cc, ok := v.(string); ok && len(cc) > 4 {
					return cc[:4] + "****" + cc[len(cc)-4:]
				}
				return "***MASKED***"
			},
			"email": func(v interface{}) interface{} {
				if email, ok := v.(string); ok {
					return MaskEmail(email)
				}
				return "***MASKED***"
			},
		},
		SensitiveFields:   []string{"password", "token"},
		MaskSensitiveData: true,
	}

	log := logger.New(logger.LoggerConfig{
		Level:     core.INFO,
		Output:    os.Stdout,
		Formatter: jsonFormatter,
	})
	defer log.Close()

	// Log with sensitive data that will be transformed
	log.WithFields(map[string]interface{}{
		"user":        "jane.doe@example.com",
		"credit_card": "1234567890123456",
		"password":    "supersecret123",
		"action":      "purchase",
		"amount":      199.99,
	}).Info("Secure transaction processed")
}

func hookExample() {
	fmt.Println("\n--- Hook Implementation Example ---")

	// Create a custom hook that sends logs to an external service
	customHook := &CustomHTTPHook{
		endpoint: "https://logs.example.com/api/logs",
	}

	log := logger.New(logger.LoggerConfig{
		Level:   core.WARN,
		Output:  os.Stdout,
		Formatter: &formatter.JSONFormatter{
			TimestampFormat: logger.DEFAULT_TIMESTAMP_FORMAT,
		},
		Hooks: []hook.Hook{customHook},
	})
	defer log.Close()

	// This will trigger the hook since it's WARN level
	log.WithFields(map[string]interface{}{
		"error_code": "E500",
		"component":  "database",
	}).Warn("Database connection failed, retrying...")

	// Close hook to flush any remaining data
	customHook.Close()
}

func performanceExample() {
	fmt.Println("\n--- Performance Optimization Example ---")

	// High-performance async logger configuration
	perfLog := logger.New(logger.LoggerConfig{
		Level:                         core.INFO,
		Output:                        os.Stdout,
		AsyncLogging:                  true,
		AsyncWorkerCount:              6,
		AsyncLogChannelBufferSize:     5000,
		DisablePerLogContextTimeout:   true,
		BufferSize:                    4096,
		FlushInterval:                 50 * time.Millisecond,
		DisableLocking:                true,
		EnableStackTrace:              false, // Disable for performance
		Formatter: &formatter.CSVFormatter{
			IncludeHeader: false,
		},
	})
	defer perfLog.Close()

	// Log many messages to demonstrate performance
	start := time.Now()
	for i := 0; i < 1000; i++ {
		perfLog.WithFields(map[string]interface{}{
			"iteration": i,
			"worker":    "perf-worker",
		}).Info("Performance test message")
	}
	duration := time.Since(start)

	fmt.Printf("Logged 1000 messages in %v using async logging\n", duration)

	// Close to ensure all messages are processed
	perfLog.Close()
}

// CustomHTTPHook is a custom hook implementation
type CustomHTTPHook struct {
	endpoint string
}

// Fire implements the Hook interface
func (h *CustomHTTPHook) Fire(entry *core.LogEntry) error {
	// In a real implementation, this would send the log entry to an HTTP endpoint
	// For this example, we'll just simulate the operation
	_ = entry
	fmt.Printf("Simulating sending log to %s\n", h.endpoint)
	return nil
}

// Close implements the Hook interface
func (h *CustomHTTPHook) Close() error {
	fmt.Println("CustomHTTPHook closed")
	return nil
}

// at
func MaskEmail(email string) string {
	if idx := strings.Index(email, "@"); idx != -1 {
		localPart := email[:idx]
		if len(localPart) > 2 {
			return localPart[:2] + "***@" + email[idx+1:]
		}
		return "***@" + email[idx+1:]
	}
	return email
}