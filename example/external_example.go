package example

import (
	"context"
	"os"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/logger"
)

// ExternalServiceExample demonstrates how to use the external hook
func ExternalServiceExample() {
	// Create an external hook for a logging service
	// In a real application, this would be your actual logging service endpoint
	// For this example, we'll just show how to set it up

	// externalHook, err := hook.NewExternalHook("https://logs.your-service.com/api/logs", 3, 10*time.Second)
	// if err != nil {
	// 	panic(err)
	// }

	// Since we can't actually connect to an external service in this example,
	// we'll just show how the code would be structured

	// Create a logger with the external hook
	// This logger will send all ERROR level and above logs to the external service
	log := logger.New(logger.LoggerConfig{
		Level:  core.INFO,
		Output: os.Stdout,
		Formatter: &formatter.TextFormatter{
			EnableColors:  true,
			ShowTimestamp: true,
			ShowCaller:    true,
		},
		// In real usage, uncomment the next lines:
		// Hooks: []hook.Hook{externalHook},
	})
	defer log.Close()

	// Log some messages - these would also be sent to the external service
	log.Info("Application started")
	log.Info("Connecting to external service")
	log.Warn("This is a warning that might be sent to external service")
	log.Error("This error would be sent to external service")

	// Simulate some processing
	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "trace-123")

	log.InfoC(ctx, "Processing request with context")

	// The external hook handles retries, timeouts, and graceful degradation
	// If the external service is down, logs will continue to work normally
	// and errors will be handled gracefully without crashing the application
}
