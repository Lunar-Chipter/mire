package main

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/logger"
	"github.com/Lunar-Chipter/mire/util"
	// "github.com/Lunar-Chipter/mire/hook" // Removed import for hook package
)

// wrappedError wraps an error with a message
type wrappedError struct {
	msg   string
	cause error
}

func (e *wrappedError) Error() string {
	if e.cause != nil {
		return e.msg + ": " + e.cause.Error()
	}
	return e.msg
}

func (e *wrappedError) Unwrap() error {
	return e.cause
}

// printLine is a helper function to print lines without fmt
func printLine(s string) {
	os.Stdout.Write([]byte(s))
	os.Stdout.Write([]byte("\n"))
}

func main() {
	printLine("===================================================")
	printLine("  MIRE LOGGING LIBRARY DEMONSTRATION")
	printLine("===================================================")

	// Example 1: Default Logger
	printLine("### 1. Default Logger ###")
	defaultConfig := logger.Config{
		Level:            core.WARN, // Only show WARN and above in console
		Output:           os.Stdout,
		ErrorOutput:      io.Discard, // Discard internal logger error messages
		CallerDepth:      logger.DEFAULT_CALLER_DEPTH,
		TimestampFormat:  logger.DEFAULT_TIMESTAMP_FORMAT,
		BufferSize:       logger.DEFAULT_BUFFER_SIZE,
		FlushInterval:    logger.DEFAULT_FLUSH_INTERVAL,
		ClockInterval:    10 * time.Millisecond,
		Formatter: &formatter.TextFormatter{
			UseColors:    true,
			ShowTimestamp:   true,
			ShowCaller:      true,
			TimestampFormat: logger.DEFAULT_TIMESTAMP_FORMAT,
		},
	}
	logDefault := logger.New(defaultConfig)
	defer logDefault.Close() // Ensure logger is closed cleanly.

	logDefault.Info("This is an INFORMATION message from the default logger. (Will not appear in console due to WARN level).")
	logDefault.Warnf("There are %d warnings in the system.", 2)
	logDefault.Debug("This debug message will NOT appear because default level is WARN.") // Will not appear
	logDefault.Trace("This trace message will also NOT appear.")                          // Will not appear
	logDefault.Error("A simple error occurred.")
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond) // Give time to flush buffer if any

	// --- Example 2: Logger with Fields and Context ---
	// Logger allows adding fields (key-value pairs) to each log entry.
	// We can also add contextual information like TraceID, SpanID.
	printLine("### 2. Logger with Fields & Context ###")
	ctx := context.Background()
	// Adding TraceID, SpanID, UserID to context.
	// These will be automatically extracted by the logger if enabled.
	ctx = util.WithTraceID(ctx, "trace-xyz-987")
	ctx = util.WithSpanID(ctx, "span-123")
	ctx = util.WithUserID(ctx, "user-alice")

	// Using default logger to demonstrate fields and context.
	logWithContext := logDefault.WithFields(map[string]interface{}{
		"service": "auth-service",
		"version": "1.0.0",
	})
	logWithContext.WithFields(map[string]interface{}{
		"username":   "alice",
		"ip_address": "192.168.1.100",
	}).Info("User successfully logged in.")

	// Log with explicit context using context-aware methods.
	logWithContext.InfofC(ctx, "Processing authorization request for %s.", "token-ABC")
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// --- Example 3: Error Logging with Stack Trace ---
	// Logger can record errors and include stack traces for debugging.
	printLine("### 3. Error Logging with Stack Trace ###")
	errSample := errors.New("failed to read database configuration")
	logDefault.WithFields(map[string]interface{}{
		"error_code": 500,
		"component":  "database-connector",
	}).Error("Error during initialization:", errSample.Error())
	// Default logger already enables ShowStackTrace for ERROR level and above.
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// --- Example 4: JSON Logger to File (app.log) ---
	// Configure logger to write logs in JSON format to file.
	printLine("### 4. JSON Logger to File (app.log) ###")
	printLine("JSON logs will be written to 'app.log'. Check its contents after the program completes.")
	jsonFileLogger, err := setupJSONFileLogger("app.log")
	if err != nil {
		logDefault.Fatalf("Failed to setup JSON file logger: %v", err) // Use logDefault to fatal here
	}
	defer jsonFileLogger.Close() // Important: Close logger to flush buffers to file!

	jsonFileLogger.Debug("Debug message for JSON file logger.")
	jsonFileLogger.WithFields(map[string]interface{}{
		"trans_id": "TXN-001",
		"amount":   123.45,
		"currency": "IDR",
	}).Info("Transaction processed successfully.")
	jsonFileLogger.WithFields(map[string]interface{}{
		"user_id":   "user-bob",
		"cache_key": "user:bob",
	}).Error("Failed to save user data to cache.")
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// --- Example 5: Custom Text Logger (Without Timestamp & Caller) ---
	// Create logger with custom text format, hiding some metadata.
	printLine("### 5. Custom Text Logger ###")
	customTextLogger := setupCustomTextLogger()
	customTextLogger.Info("This is an 'INFO' message from custom logger (without timestamp/caller).")
	customTextLogger.Infof("Lowest level: %s", core.TRACE.String())
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// --- Example 6: Hooks Demonstration (errors.log) ---
	// Shows how to configure and use hooks.
	// ERROR level and above logs will be written to 'errors.log'.
	// printLine("### 6. Hooks Demonstration ###") // Commented out
	// demonstrateHooks() // Commented out
	printLine("---------------------------------------------------")
	time.Sleep(10 * time.Millisecond)

	// Ensure all buffers are flushed before program ends.
	// Especially important for buffered writers and async loggers.
	// logger.New().Close() // If NewDefaultLogger is called multiple times, only need to close the used instance.
	// jsonFileLogger.Close() // Make sure to close if not deferred
	// Usually, the main logger will be closed at the end of the application.
	// For this demo, we don't explicitly close logDefault here,
	// because it writes directly to os.Stdout, but if it has a buffered writer,
	// then it should be closed.

	printLine("===================================================")
	printLine("  DEMONSTRATION COMPLETED                          ")
	printLine("===================================================")
	printLine("Check 'app.log' and 'errors.log' files to see the log output.")
}

func setupJSONFileLogger(filePath string) (*logger.Logger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		return nil, &wrappedError{
			msg:   "failed to open log file " + filePath,
			cause: err,
		}
	}

	jsonConfig := logger.Config{
		Level:       core.DEBUG,
		Output:      file,
		ErrorOutput: io.Discard, // Discard internal logger error messages
		BufferSize:  1024,       // Enable buffered writer
		Formatter: &formatter.JSONFormatter{
			PrettyPrint:      true,
			ShowCaller:       true,
			IncludeStackTrace: true,
			TimestampFormat:  logger.DEFAULT_TIMESTAMP_FORMAT,
		},
	}
	return logger.New(jsonConfig), nil
}

// setupCustomTextLogger creates a logger with simplified text format.
func setupCustomTextLogger() *logger.Logger {
	customConfig := logger.Config{
		Level:       core.TRACE, // Display all logs, even trace
		ErrorOutput: io.Discard, // Discard internal logger error messages
		Formatter: &formatter.TextFormatter{
			UseColors:  true,
			ShowTimestamp: false, // Hide timestamp
			ShowCaller:    false, // Hide caller info
			ShowPID:       true,  // Show Process ID
			ShowGoroutine: true,  // Show Goroutine ID
		},
	}
	return logger.New(customConfig)
}

// demonstrateHooks shows how to configure and use hooks.
// This hook will write all ERROR level and above logs to 'errors.log'.
// func demonstrateHooks() {
// 	fmt.Println("ERROR level and above logs will be written to 'errors.log'.")

// 	// Create a file hook for errors.log
// 	errorFileHook, err := hook.NewFileHook("errors.log") // Use mire/hook/NewFileHook
// 	if err != nil {
// 		fmt.Printf("Failed to create error file hook: %v\n", err)
// 		return
// 	}
// 	defer errorFileHook.Close() // Ensure the file hook is closed

// 	// 2. Configure logger to use hook.
// 	logWithHook := logger.New(logger.Config{
// 		Level:       core.INFO, // This logger will display INFO and above to console
// 		Output:      os.Stdout, // Console output
// 		ErrorOutput: io.Discard, // Discard internal logger errors
// 		ErrorHook: false, // Disable the built-in error file hook
// 		Hooks: []hook.Hook{
// 			errorFileHook, // Add the manual file hook
// 		},
// 		Formatter: &formatter.TextFormatter{ // This formatter is for console output
// 			UseColors:    true,
// 			TimestampFormat: logger.DEFAULT_TIMESTAMP_FORMAT,
// 			ShowCaller:      true,
// 		},
// 	})
// 	defer logWithHook.Close() // Ensure this logger instance is closed

// 	// 3. Use logger as usual.
// 	logWithHook.Info("This is an INFO message, will appear in console, but not in 'errors.log'.")
// 	logWithHook.Warn("This is a WARN message, will appear in console, but not in 'errors.log'.")

// 	// This message will go to console AND to errors.log file because of hook.
// 	logWithHook.WithFields(map[string]interface{}{"db_host": "10.0.0.5"}).Error("Failed to connect to database.")

// 	// This message will also trigger the hook.
// 	logWithHook.WithFields(map[string]interface{}{"service": "payment-gateway"}).Error("Transaction timeout.")
// }
