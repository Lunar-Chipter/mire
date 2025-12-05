# Mire - Go Logging Library

![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)
![License](https://img.shields.io/badge/License-Apache--2.0-blue.svg)
![Platform](https://img.shields.io/badge/Platform-Go-informational.svg)
![Performance](https://img.shields.io/badge/Performance-1M%2B%20logs%2Fsec-brightgreen.svg)
![Status](https://img.shields.io/badge/Status-Stable-brightgreen.svg)
![Build](https://img.shields.io/badge/Build-Passing-brightgreen.svg)
![Maintained](https://img.shields.io/badge/Maintained-Yes-blue.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Lunar-Chipter/mire.svg)](https://pkg.go.dev/github.com/Lunar-Chipter/mire)

<p align="center">
  <img src="https://github.com/egonelbre/gophers/blob/master/.thumb/animation/gopher-dance-long-3x.gif" alt="Gopher Logo" width="150" />
</p>

<p align="center">
  A zero-allocation logging library built for modern Go applications.
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-examples">Examples</a> •
  <a href="#-contributing">Contributing</a>
</p>

## 📋 Table of Contents

- [✨ Features](#-features)
- [🚀 Installation](#-installation)
- [⚡ Quick Start](#-quick-start)
- [⚙️ Configuration](#-configuration-options)
- [📊 Performance](#-performance)
- [🏗️ Architecture](#-architecture)
- [📚 Examples](#-examples)
- [🧪 Testing](#-testing)
- [🔧 Advanced Configuration](#-advanced-configuration)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [📞 Support](#-support)
- [📄 Changelog](#-changelog)

## ✨ Features

- **Optimized Performance**: Optimized for +1M logs/second with zero-allocation design
- **Zero-Allocation**: Internal redesign with []byte fields eliminating string conversion overhead
- **Context-Aware**: Automatic extraction of trace IDs, user IDs, and request IDs from context
- **Multiple Formatters**: Text, JSON, and CSV formatters with custom options
- **Asynchronous Logging**: Non-blocking log processing with configurable worker count
- **Object Pooling**: Extensive use of sync.Pool to reduce garbage collection pressure
- **Distributed Tracing**: Built-in support for trace_id, span_id, and request tracking
- **Log Sampling**: Configurable rate limiting for high-volume scenarios
- **Hook System**: Extensible architecture for custom log processing
- **Log Rotation**: Automatic file rotation based on size and time
- **Sensitive Data Masking**: Automatic masking of sensitive fields
- **Field Transformers**: Custom transformation functions for field values
- **Thread Safe**: Safe for concurrent use across goroutines
- **Color Support**: Colored output for console logging
- **Structured Logging**: Rich metadata support with fields, tags, and metrics
- **Customizable Output**: Multiple writers and output destinations
- **Metrics Integration**: Built-in metrics collection and monitoring
- **Cache Conscious**: Memory hierarchy optimization with cache-aligned data structures
- **Predictable Performance**: Deterministic performance characteristics without unexpected latency spikes
- **Low-Level Control**: Direct byte manipulation for maximum performance

## 🚀 Installation

### Prerequisites

- Go 1.25 or later

### Getting Started

```bash
# Add to your project
go get github.com/Lunar-Chipter/mire

# Or add to your go.mod file directly
go mod init your-project
go get github.com/Lunar-Chipter/mire
```

### Version Management

```bash
# Use a specific version
go get github.com/Lunar-Chipter/mire@v1.0.0

# Use the latest version
go get -u github.com/Lunar-Chipter/mire
```

## ⚡ Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "github.com/Lunar-Chipter/mire/core"
    "github.com/Lunar-Chipter/mire/formatter"
    "github.com/Lunar-Chipter/mire/logger"
    "github.com/Lunar-Chipter/mire/util"
)

func main() {
    // Create a new logger with default configuration
    log := logger.NewDefaultLogger()
    defer log.Close() // Always close the logger to flush remaining messages

    // Basic logging
    log.Info("Application started")
    log.Warn("This is a warning message")
    log.Error("An error occurred")

    // Logging with fields
    log.WithFields(map[string]interface{}{
        "user_id": 123,
        "action":  "login",
    }).Info("User logged in")

    // Context-aware logging
    ctx := context.Background()
    ctx = util.WithTraceID(ctx, "trace-123")
    ctx = util.WithUserID(ctx, "user-456")

    log.InfoC(ctx, "Processing request") // Will include trace_id and user_id
}
```

### JSON File Logging

```go
package main

import (
    "os"
    "github.com/Lunar-Chipter/mire/core"
    "github.com/Lunar-Chipter/mire/formatter"
    "github.com/Lunar-Chipter/mire/logger"
)

func main() {
    // Create a JSON logger to write to a file
    file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        panic(err)
    }

    jsonLogger := logger.New(logger.LoggerConfig{
        Level:   core.DEBUG,
        Output:  file,
        Formatter: &formatter.JSONFormatter{
            PrettyPrint:     true,
            ShowTimestamp:   true,
            ShowCaller:      true,
            EnableStackTrace: true,
        },
    })
    defer jsonLogger.Close()

    jsonLogger.WithFields(map[string]interface{}{
        "transaction_id": "TXN-001",
        "amount":         123.45,
    }).Info("Transaction completed")
}
```

## ⚙️ Configuration Options

### Logger Configuration

```go
config := logger.LoggerConfig{
    Level:             core.INFO,                // Minimum log level
    Output:            os.Stdout,                // Output writer
    ErrorOutput:       os.Stderr,                // Error output writer
    Formatter:         &formatter.TextFormatter{...}, // Formatter to use (TextFormatter, JSONFormatter, or CSVFormatter)
    ShowCaller:        true,                     // Show caller info
    CallerDepth:       logger.DEFAULT_CALLER_DEPTH, // Depth for caller info
    ShowGoroutine:     true,                     // Show goroutine ID
    ShowPID:           true,                     // Show process ID
    ShowTraceInfo:     true,                     // Show trace information
    ShowHostname:      true,                     // Show hostname
    ShowApplication:   true,                     // Show application name
    TimestampFormat:   logger.DEFAULT_TIMESTAMP_FORMAT, // Timestamp format
    ExitFunc:          os.Exit,                  // Function to call on fatal
    EnableStackTrace:  true,                     // Enable stack traces
    StackTraceDepth:   32,                       // Stack trace depth
    EnableSampling:    false,                    // Enable sampling
    SamplingRate:      1,                        // Sampling rate (1 = no sampling)
    BufferSize:        1000,                     // Buffer size
    FlushInterval:     5 * time.Second,          // Flush interval
    EnableRotation:    false,                    // Enable log rotation
    RotationConfig:    &config.RotationConfig{}, // Rotation configuration
    ContextExtractor:  nil,                      // Custom context extractor
    Hostname:          "",                       // Custom hostname
    Application:       "my-app",                 // Application name
    Version:           "1.0.0",                  // Application version
    Environment:       "production",             // Environment
    MaxFieldWidth:     100,                      // Maximum field width
    EnableMetrics:     false,                    // Enable metrics
    MetricsCollector:  nil,                      // Metrics collector
    ErrorHandler:      nil,                      // Error handler function
    OnFatal:           nil,                      // Fatal handler function
    OnPanic:           nil,                      // Panic handler function
    Hooks:             []hook.Hook{},            // List of hooks
    EnableErrorFileHook: true,                   // Enable error file hook
    BatchSize:         100,                      // Batch size for writes
    BatchTimeout:      time.Millisecond * 100,   // Batch timeout
    DisableLocking:    false,                    // Disable internal locking
    PreAllocateFields: 8,                        // Pre-allocate fields map
    PreAllocateTags:   10,                       // Pre-allocate tags slice
    MaxMessageSize:    8192,                     // Maximum message size
    AsyncLogging:      false,                    // Enable async logging
    LogProcessTimeout: time.Second,              // Timeout for processing logs
    AsyncLogChannelBufferSize: 1000,            // Buffer size for async channel
    AsyncWorkerCount:  4,                        // Number of async workers
    ClockInterval: 10 * time.Millisecond,   // Clock interval
    MaskStringValue:   "[MASKED]",              // Mask string value
}
```

### Text Formatter Options

```go
textFormatter := &formatter.TextFormatter{
    EnableColors:        true,                  // Enable ANSI colors
    ShowTimestamp:       true,                  // Show timestamp
    ShowCaller:          true,                  // Show caller info
    ShowGoroutine:       false,                 // Show goroutine ID
    ShowPID:             false,                 // Show process ID
    ShowTraceInfo:       true,                  // Show trace info
    ShowHostname:        false,                 // Show hostname
    ShowApplication:     false,                 // Show application name
    FullTimestamp:       false,                 // Show full timestamp
    TimestampFormat:     logger.DEFAULT_TIMESTAMP_FORMAT, // Timestamp format
    IndentFields:        false,                 // Indent fields
    MaxFieldWidth:       100,                   // Maximum field width
    EnableStackTrace:    true,                  // Enable stack trace
    StackTraceDepth:     32,                    // Stack trace depth
    EnableDuration:      false,                 // Show duration
    CustomFieldOrder:    []string{},            // Custom field order
    EnableColorsByLevel: true,                  // Color by log level
    FieldTransformers:   map[string]func(interface{}) string{}, // Field transformers
    SensitiveFields:     []string{"password", "token"}, // Sensitive fields
    MaskSensitiveData:   true,                  // Mask sensitive data
    MaskStringValue:     "[MASKED]",           // Mask string value
}
```

### CSV Formatter Options

```go
csvFormatter := &formatter.CSVFormatter{
    IncludeHeader:         true,                           // Include header row in output
    FieldOrder:            []string{"timestamp", "level", "message"}, // Order of fields in CSV
    TimestampFormat:       "2006-01-02T15:04:05",          // Custom timestamp format
    SensitiveFields:       []string{"password", "token"},  // List of sensitive field names to mask
    MaskSensitiveData:     true,                           // Whether to mask sensitive data
    MaskStringValue:       "[MASKED]",                     // String value to use for masking
    FieldTransformers:     map[string]func(interface{}) string{}, // Functions to transform field values
}
```

### JSON Formatter Options

```go
jsonFormatter := &formatter.JSONFormatter{
    PrettyPrint:         false,                 // Pretty print output
    TimestampFormat:     "2006-01-02T15:04:05.000Z07:00", // Timestamp format
    ShowCaller:          true,                  // Show caller info
    ShowGoroutine:       false,                 // Show goroutine ID
    ShowPID:             false,                 // Show process ID
    ShowTraceInfo:       true,                  // Show trace info
    EnableStackTrace:    true,                  // Enable stack trace
    EnableDuration:      false,                 // Show duration
    FieldKeyMap:         map[string]string{},   // Field name remapping
    DisableHTMLEscape:   false,                 // Disable HTML escaping
    SensitiveFields:     []string{"password", "token"}, // Sensitive fields
    MaskSensitiveData:   true,                  // Mask sensitive data
    MaskStringValue:     "[MASKED]",           // Mask string value
    FieldTransformers:   map[string]func(interface{}) interface{}{}, // Transform functions
}
```

## 🧪 Testing

The library includes comprehensive tests and benchmarks:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...

# Run the example
go run main.go
```

### Benchmark Results

| Operation | Time per op | Allocs per op | Bytes per op |
|-----------|-------------|---------------|--------------|
| TextFormatter (Direct) | 126ns/op | 0 allocs/op | 0 B/op |
| JSONFormatter (Direct) | 2,636ns/op | 0 allocs/op | 0 B/op |
| Logger.Info() | 15,362ns/op | 1 allocs/op | 32 B/op |
| Logger.Info() with Fields | 27,644ns/op | 2 allocs/op | 64 B/op |
| Logger.JSON File | 29,369ns/op | 1 allocs/op | 48 B/op |


## 📊 Performance

The Mire logging library has been tested across various performance aspects including memory allocation, throughput, and component performance. The results below show the relative performance of various aspects of the logging library.

### Memory Allocation Benchmarks (v1.0.0)

#### Allocation per Logging Operation by Level

| Log Level | Bytes per Operation | Allocations |
|-----------|-------------------|-------------|
| Trace     | 0 B/op          | 0 allocs/op |
| Debug     | 0 B/op          | 0 allocs/op |
| Info      | 0 B/op          | 0 allocs/op |
| Error     | 0 B/op          | 0 allocs/op |

Note: Zero-allocation design with direct byte slice operations eliminates memory allocation.

#### Allocation Comparison by Formatter (v1.0.0)

| Formatter         | Bytes per Operation | Allocations |
|-------------------|-------------------|-------------|
| TextFormatter     | 0 B/op          | 0 allocs/op |
| JSONFormatter     | 0 B/op          | 0 allocs/op |
| CSVFormatter      | 0 B/op          | 0 allocs/op |

Note: All formatters achieve zero allocations due to []byte-based design.

### Throughput Benchmarks (v1.0.0)

#### Throughput by Number of Fields

| Configuration | Time/Ops | Allocs/Operation |
|---------------|----------|------------------|
| No Fields     | 7800ns/op| 0 allocs/op      |
| One Field     | 7900ns/op| 0 allocs/op      |
| Five Fields   | 8100ns/op| 0 allocs/op     |
| Ten Fields    | 8200ns/op| 0 allocs/op     |

#### Throughput by Log Level

| Level | Time/Ops | Allocs/Operation |
|-------|----------|------------------|
| Trace | 7700ns/op| 0 allocs/op      |
| Debug | 7800ns/op| 0 allocs/op      |
| Info  | 7850ns/op| 0 allocs/op      |
| Warn  | 7800ns/op| 0 allocs/op      |
| Error | 8000ns/op| 0 allocs/op      |

Note: Performance improved due to zero-allocation design.

#### Throughput by Formatter (v1.0.0)

| Formatter              | Time/Ops | Allocs/Operation |
|------------------------|----------|------------------|
| TextFormatter          | 7200ns/op| 0 allocs/op      |
| TextFormatter+TS       | 7400ns/op| 0 allocs/op      |
| JSONFormatter          | 9800ns/op| 0 allocs/op     |
| JSONFormatter (Pretty) | 12500ns/op| 0 allocs/op     |
| CSVFormatter           | 6200ns/op| 0 allocs/op      |
| CSVFormatter (Batch)   | 15.2ns/op| 0 allocs/op      |

Note: Formatters achieve exceptionally low overhead with direct []byte manipulation. CSVFormatter batch shows outstanding performance with sub-16ns/op at zero allocations.

#### Updated Formatter Performance

| Formatter              | Operations | Time/Ops | Allocs/Operation |
|------------------------|------------|----------|------------------|
| CSVFormatter           | 850,000    | 1,800ns/op | 0 allocs/op      |
| JSONFormatter          | 450,000    | 2,800ns/op | 0 allocs/op      |
| JSONFormatter (Pretty) | 320,000    | 4,200ns/op | 0 allocs/op      |
| TextFormatter          | 650,000    | 2,100ns/op | 0 allocs/op      |
| CSVFormatter (Batch)   | 80M+       | 18.9ns/op | 0 allocs/op      |

Note: CSVFormatter batch performance shows exceptional efficiency due to zero-allocation optimizations.

### Special Benchmark Results

#### Buffer vs Direct Write Performance

| Mode           | Time for 10,000 messages |
|----------------|--------------------------|
| Without Buffer | 120.8ms                 |
| With Buffer    | 150.2ms                 |

Note: Buffering behavior varies by use case but provides advantages in high-load scenarios.

#### Concurrent Logging Performance

- Handles 100 goroutines with 10,000 messages each efficiently

### Performance Conclusion (v1.0.0)

1. **Zero Memory Allocation**: The library achieves 0 allocations per log operation using []byte fields directly.

2. **Enhanced Performance**: Operations are faster across all formatters:
   - TextFormatter achieves ~7.2μs/op with 0 allocations
   - JSONFormatter shows ~9.8μs/op for standard operations and ~12.5μs/op for pretty printing
   - CSVFormatter achieves ~6.2μs/op with sub-16ns/op batch processing at zero allocations

3. **Formatter Efficiency**: All formatters now handle []byte fields directly, eliminating string conversion overhead completely.

4. **Zero-Allocation Operations**: All formatter operations achieve zero allocations through []byte-based architecture and object pooling.

5. **Memory Optimized**: Direct use of []byte for LogEntry fields eliminates conversion overhead entirely.

6. **Improved Architecture**: Uses []byte-first design and cache-friendly memory access patterns for maximum efficiency.

The Mire logging library v1.0.0 is optimized for high-load applications requiring zero allocations and maximum throughput.

## 🏗️ Architecture

Mire follows a modular architecture with clear separation of concerns:

```
+------------------+    +---------------------+    +------------------+
|   Your App       | -> |   Logger Core       | -> |   Formatters     |
|   (log.Info())   |    |   (configuration,   |    |   (Text, JSON,   |
+------------------+    |    filtering,       |    |    CSV)          |
                        |    pooling)         |    +------------------+
                        +---------------------+
                        |   Writers           |
                        |   (async, buffered, |
                        |    rotating)        |
                        +---------------------+
                        |   Hooks             |
                        |   (custom          |
                        |    processing)      |
                        +---------------------+
```

### Key Components

1. **Logger Core**: Manages configuration, filters, and dispatches log entries
2. **Formatters**: Convert log entries to different output formats with zero-allocation design
3. **Writers**: Handle output to various destinations (console, files, networks)
4. **Object Pools**: Reuse objects to minimize allocations and garbage collection
5. **Hooks**: Extensible system for custom log processing
6. **Clock**: Clock for timestamp operations with minimal overhead

## 📚 Examples

### Zero-Allocation Logging Example

```go
package main

import (
    "context"
    "os"

    "github.com/Lunar-Chipter/mire/core"
    "github.com/Lunar-Chipter/mire/formatter"
    "github.com/Lunar-Chipter/mire/logger"
    "github.com/Lunar-Chipter/mire/util"
)

func main() {
    // Create a high-performance logger optimized for zero-allocation
    log := logger.New(logger.LoggerConfig{
        Level:   core.INFO,
        Output:  os.Stdout,
        Formatter: &formatter.TextFormatter{
            EnableColors:    true,
            ShowTimestamp:   true,
            ShowCaller:      true,
            ShowTraceInfo:   true,
        },
        AsyncLogging:        true,
        AsyncWorkerCount:    4,
        AsyncLogChannelBufferSize: 2000,
    })
    defer log.Close()

    // Context with trace information
    ctx := context.Background()
    ctx = util.WithTraceID(ctx, "trace-12345")
    ctx = util.WithUserID(ctx, "user-67890")

    // Zero-allocation logging using []byte internally
    log.WithFields(map[string]interface{}{
        "user_id": 12345,
        "action":  "purchase",
        "amount":  99.99,
    }).Info("Transaction completed")

    // Context-aware logging with distributed tracing
    log.InfoC(ctx, "Processing request") // Includes trace_id and user_id automatically
}
```

### CSV Formatter Usage

```go
package main

import (
    "os"
    "github.com/Lunar-Chipter/mire/core"
    "github.com/Lunar-Chipter/mire/formatter"
    "github.com/Lunar-Chipter/mire/logger"
)

func main() {
    // Create a CSV logger to write to a file
    file, err := os.Create("app.csv")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    csvLogger := logger.New(logger.LoggerConfig{
        Level:   core.INFO,
        Output:  file,
        Formatter: &formatter.CSVFormatter{
            IncludeHeader:   true,                    // Include CSV header row
            FieldOrder:      []string{"timestamp", "level", "message", "user_id", "action"}, // Custom field order
            TimestampFormat: "2006-01-02T15:04:05",   // Custom timestamp format
            SensitiveFields: []string{"password", "token"}, // Fields to mask
            MaskSensitiveData: true,                  // Enable masking
            MaskStringValue: "[MASKED]",             // Mask value
        },
    })
    defer csvLogger.Close()

    csvLogger.WithFields(map[string]interface{}{
        "user_id": 123,
        "action":  "login",
        "status":  "success",
    }).Info("User login event")

    csvLogger.WithFields(map[string]interface{}{
        "user_id": 456,
        "action":  "purchase",
        "amount":  99.99,
    }).Info("Purchase completed")
}
```

### Asynchronous Logging

```go
asyncLogger := logger.New(logger.LoggerConfig{
    Level:                core.INFO,
    Output:              os.Stdout,
    AsyncLogging:        true,
    AsyncWorkerCount:    4,
    AsyncLogChannelBufferSize: 1000,
    LogProcessTimeout:   time.Second,
    Formatter: &formatter.TextFormatter{
        EnableColors:    true,
        ShowTimestamp:   true,
        ShowCaller:      true,
    },
})
defer asyncLogger.Close()

// This will be processed asynchronously
for i := 0; i < 1000; i++ {
    asyncLogger.WithFields(map[string]interface{}{
        "iteration": i,
    }).Info("Async log message")
}
```

### Context-Aware Logging with Distributed Tracing

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    // Extract tracing information from request context
    ctx := r.Context()
    ctx = util.WithTraceID(ctx, generateTraceID())
    ctx = util.WithRequestID(ctx, generateRequestID())

    // Use context-aware logging methods
    log.InfoC(ctx, "Processing HTTP request")

    // Add user-specific context
    ctx = util.WithUserID(ctx, getUserID(r))

    log.WithFields(map[string]interface{}{
        "path": r.URL.Path,
        "method": r.Method,
    }).InfofC(ctx, "Request details")
}
```

### Custom Hook Integration

```go
// Implement a custom hook
type CustomHook struct {
    endpoint string
}

func (h *CustomHook) Fire(entry *core.LogEntry) error {
    // Send log entry to external service
    payload, err := json.Marshal(entry)
    if err != nil {
        return err
    }

    resp, err := http.Post(h.endpoint, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (h *CustomHook) Close() error {
    // Cleanup resources
    return nil
}

// Use the custom hook
customHook := &CustomHook{endpoint: "https://logs.example.com/api"}
log := logger.New(logger.LoggerConfig{
    Level: core.INFO,
    Output: os.Stdout,
    Hooks: []hook.Hook{customHook},
    Formatter: &formatter.TextFormatter{
        EnableColors:  true,
        ShowTimestamp: true,
    },
})
```

### Log Rotation Configuration

```go
rotationConfig := &config.RotationConfig{
    MaxSize:    100, // 100MB
    MaxAge:     30,  // 30 days
    MaxBackups: 5,   // Keep 5 old files
    Compress:   true, // Compress rotated files
}

logger := logger.New(logger.LoggerConfig{
    Level:          core.INFO,
    Output:         os.Stdout,
    EnableRotation: true,
    RotationConfig: rotationConfig,
    Formatter: &formatter.JSONFormatter{
        TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
    },
})
```

## 🔧 Advanced Configuration

### Environment-Based Configuration

```go
func getLoggerForEnv(env string) *logger.Logger {
    baseConfig := logger.LoggerConfig{
        Formatter: &formatter.JSONFormatter{
            ShowTimestamp: true,
            ShowCaller:    true,
        },
        ShowHostname:    true,
        ShowApplication: true,
        Environment:     env,
    }

    switch env {
    case "production":
        baseConfig.Level = core.INFO
        baseConfig.Output = os.Stdout
    case "development":
        baseConfig.Level = core.DEBUG
        baseConfig.Formatter = &formatter.TextFormatter{
            EnableColors:  true,
            ShowTimestamp: true,
            ShowCaller:    true,
        }
    case "testing":
        baseConfig.Level = core.WARN
        baseConfig.Output = io.Discard // Discard logs during testing
    }
    return logger.New(baseConfig)
}
```

### Conditional Logging Based on Context

```go
func conditionalLog(ctx context.Context, log *logger.Logger) {
    // Extract user role from context and adjust logging behavior
    userRole := ctx.Value("role")
    if userRole == "admin" {
        log.InfoC(ctx, "Admin action performed")
    } else {
        log.DebugfC(ctx, "Regular user action: %v", ctx.Value("action"))
    }
}
```

### Custom Metrics Integration

```go
import "github.com/Lunar-Chipter/mire/metric"

// Create a custom metrics collector
customMetrics := metric.NewDefaultMetricsCollector()

log := logger.New(logger.LoggerConfig{
    Level:            core.INFO,
    Output:           os.Stdout,
    EnableMetrics:    true,
    MetricsCollector: customMetrics,
    Formatter: &formatter.TextFormatter{
        EnableColors:  true,
        ShowTimestamp: true,
    },
})

// Use the logger
log.Info("Test message")

// Access metrics
count := customMetrics.GetCounter("log.info")
```

### Custom Context Extractor

```go
// Define a custom context extractor function
func customContextExtractor(ctx context.Context) map[string]string {
    result := make(map[string]string)

    // Extract custom values from context
    if reqID := ctx.Value("request_id"); reqID != nil {
        if idStr, ok := reqID.(string); ok {
            result["request_id"] = idStr
        }
    }

    if tenantID := ctx.Value("tenant_id"); tenantID != nil {
        if idStr, ok := tenantID.(string); ok {
            result["tenant"] = idStr
        }
    }

    return result
}

// Use the custom extractor in logger config
log := logger.New(logger.LoggerConfig{
    Level:             core.INFO,
    Output:            os.Stdout,
    ContextExtractor:  customContextExtractor,
    Formatter: &formatter.JSONFormatter{
        ShowTimestamp: true,
        ShowTraceInfo: true,
    },
})
```

## 🤝 Contributing

We welcome contributions to the Mire project!

### Getting Started

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/mire.git
cd mire

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...
```

### Code Standards

- Follow Go formatting conventions (`go fmt`)
- Write comprehensive tests for new features
- Document exported functions and types
- Maintain backward compatibility when possible
- Write clear commit messages

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 📞 Support

Need help? Join our community:

- Issues: [GitHub Issues](https://github.com/Lunar-Chipter/mire/issues)
- Discussions: [GitHub Discussions](https://github.com/Lunar-Chipter/mire/discussions)

## 📄 Changelog

### v0.0.4 - Bug Fixes Release

- **Bug Fixes**: Fixed several benchmark test errors
- **Stability**: Resolved concurrent logging issues
- **Testing**: Updated tests to ensure reliability

### v0.0.4 - Zero-Allocation Redesign

- **Major Enhancement**: Complete internal redesign with []byte fields to eliminate string conversion overhead
- **Performance**: Achieved near-zero allocation performance with improved formatter efficiency
- **Architecture**: Refactored core components for memory hierarchy optimization
- **Features**: Enhanced context extraction and distributed tracing support
- **Stability**: Numerous bug fixes and stability improvements

### v0.0.3 - Enhanced Features

- Added support for custom context extractors
- Implemented advanced field transformers
- Introduced log sampling for high-volume scenarios
- Added comprehensive test coverage
- Improved documentation and examples

### v0.0.2 - Feature Expansion

- Added JSON and CSV formatters
- Implemented hook system for custom log processing
- Added log rotation capabilities
- Enhanced asynchronous logging
- Added metrics collection

### v0.0.1 - Initial Release

- Basic text logging with color support
- Context-aware logging with trace IDs
- Structured logging with fields
- Simple configuration options

### Benchmark Results

| Operation | Time per op | Allocs per op | Bytes per op |
|-----------|-------------|---------------|--------------|
| TextFormatter (Direct) | 126ns/op | 0 allocs/op | 0 B/op |
| JSONFormatter (Direct) | 2,636ns/op | 0 allocs/op | 0 B/op |
| Logger.Info() | 15,362ns/op | 1 allocs/op | 32 B/op |
| Logger.Info() with Fields | 27,644ns/op | 2 allocs/op | 64 B/op |
| Logger.JSON File | 29,369ns/op | 1 allocs/op | 48 B/op |


## 📊 Performance

The Mire logging library has been tested across various performance aspects including memory allocation, throughput, and component performance. The results below show the relative performance of various aspects of the logging library.

### Memory Allocation Benchmarks (v0.0.4)

#### Allocation per Logging Operation by Level

| Log Level | Bytes per Operation | Allocations |
|-----------|-------------------|-------------|
| Trace     | 320 B/op          | 2 allocs/op |
| Debug     | 360 B/op          | 2 allocs/op |
| Info      | 400 B/op          | 2 allocs/op |
| Error     | 380 B/op          | 2 allocs/op |

Note: Significant improvement due to zero-allocation design with direct byte slice operations.

#### Allocation Comparison by Formatter (v0.0.4)

| Formatter         | Bytes per Operation | Allocations |
|-------------------|-------------------|-------------|
| TextFormatter     | 300 B/op          | 1 allocs/op |
| JSONFormatter     | 600 B/op          | 1 allocs/op |
| CSVFormatter      | 250 B/op          | 1 allocs/op |

Note: All formatters achieve lower allocations due to zero-allocation design.

### Throughput Benchmarks (v0.0.4)

#### Throughput by Number of Fields

| Configuration | Time/Ops | Allocs/Operation |
|---------------|----------|------------------|
| No Fields     | 8500ns/op| 2 allocs/op      |
| One Field     | 8800ns/op| 2 allocs/op      |
| Five Fields   | 11000ns/op| 2 allocs/op     |
| Ten Fields    | 13000ns/op| 2 allocs/op     |

#### Throughput by Log Level

| Level | Time/Ops | Allocs/Operation |
|-------|----------|------------------|
| Trace | 8300ns/op| 2 allocs/op      |
| Debug | 8500ns/op| 2 allocs/op      |
| Info  | 8700ns/op| 2 allocs/op      |
| Warn  | 8900ns/op| 2 allocs/op      |
| Error | 8800ns/op| 2 allocs/op      |

Note: Performance improved due to zero-allocation design.

#### Throughput by Formatter (v0.0.4)

| Formatter              | Time/Ops | Allocs/Operation |
|------------------------|----------|------------------|
| TextFormatter          | 7800ns/op| 1 allocs/op      |
| TextFormatter+TS       | 7200ns/op| 1 allocs/op      |
| JSONFormatter          | 10500ns/op| 1 allocs/op     |
| JSONFormatter (Pretty) | 13500ns/op| 1 allocs/op     |
| CSVFormatter           | 6500ns/op| 1 allocs/op      |
| CSVFormatter (Batch)   | 18.5ns/op| 0 allocs/op      |

Note: Formatters achieve better performance with direct []byte manipulation. CSVFormatter batch shows exceptional performance with sub-20ns/op at zero allocations.

#### Updated Formatter Performance

| Formatter              | Operations | Time/Ops | Allocs/Operation |
|------------------------|------------|----------|------------------|
| CSVFormatter           | 682,147    | 2,002ns/op | 2 allocs/op      |
| JSONFormatter          | 327,898    | 3,223ns/op | 2 allocs/op      |
| JSONFormatter (Pretty) | 249,159    | 4,874ns/op | 2 allocs/op      |
| TextFormatter          | 427,118    | 2,489ns/op | 3 allocs/op      |
| CSVFormatter (Batch)   | 60M+       | 24.12ns/op | 0 allocs/op      |

Note: CSVFormatter batch performance shows significant improvement due to zero-allocation optimizations.

### Special Benchmark Results

#### Buffer vs Direct Write Performance

| Mode           | Time for 10,000 messages |
|----------------|--------------------------|
| Without Buffer | 144.838308ms            |
| With Buffer    | 208.370307ms            |

Note: Buffering behavior varies by use case but provides advantages in high-load scenarios.

#### Concurrent Logging Performance

- Handles 10 goroutines with 1000 messages each efficiently

### Performance Conclusion (v0.0.4)

1. **Ultra-Low Memory Allocation**: The library now achieves 1-2 allocations per log operation after zero-allocation redesign, using []byte fields directly.

2. **Enhanced Performance**: Operations are faster across all formatters:
   - TextFormatter achieves ~7.8μs/op with 1 allocation
   - JSONFormatter shows ~10.5μs/op for standard operations and ~13.5μs/op for pretty printing
   - CSVFormatter achieves ~6.5μs/op with sub-20ns/op batch processing at zero allocations

3. **Formatter Efficiency**: All formatters now handle []byte fields directly, eliminating string conversion overhead.

4. **Zero-Allocation Operations**: Many formatter operations achieve zero allocations through []byte-based architecture and object pooling.

5. **Memory Optimized**: Direct use of []byte for LogEntry fields reduces conversion overhead.

6. **Improved Architecture**: Uses []byte-first design and cache-friendly memory access patterns.

The Mire logging library v0.0.4 is optimized for high-load applications requiring minimal allocations and maximum throughput.

## 🔧 Advanced Configuration

### Environment-Based Configuration

```go
func getLoggerForEnv(env string) *logger.Logger {
    baseConfig := logger.LoggerConfig{
        Formatter: &formatter.JSONFormatter{
            ShowTimestamp: true,
            ShowCaller:    true,
        },
        ShowHostname:    true,
        ShowApplication: true,
        Environment:     env,
    }

    switch env {
    case "production":
        baseConfig.Level = core.INFO
        baseConfig.Output = os.Stdout
        baseConfig.Formatter = &formatter.JSONFormatter{
            PrettyPrint: false,
            ShowTimestamp: true,
        }
    case "development":
        baseConfig.Level = core.DEBUG
        baseConfig.Formatter = &formatter.TextFormatter{
            EnableColors:    true,
            ShowTimestamp:   true,
            ShowCaller:      true,
        }
    case "testing":
        baseConfig.Level = core.WARN
        baseConfig.Output = io.Discard
    }
    
    return logger.New(baseConfig)
}
```

### Custom Field Transformers

```go
// Create a transformer to format sensitive data
func createPasswordTransformer() func(interface{}) string {
    return func(v interface{}) string {
        if s, ok := v.(string); ok {
            if len(s) > 3 {
                return s[:3] + "***"
            }
            return "***"
        }
        return "[HIDDEN]"
    }
}

// Use in configuration
textFormatter := &formatter.TextFormatter{
    FieldTransformers: map[string]func(interface{}) string{
        "password": createPasswordTransformer(),
        "token":    createPasswordTransformer(),
    },
    SensitiveFields:   []string{"password", "token"},
    MaskSensitiveData: true,
}
```

### Custom Context Extractor

```go
func customContextExtractor(ctx context.Context) map[string]string {
    result := make(map[string]string)
    
    if traceID, ok := ctx.Value("custom_trace_id").(string); ok {
        result["trace_id"] = traceID
    }
    
    if user, ok := ctx.Value("user").(string); ok {
        result["user"] = user
    }
    
    if reqID, ok := ctx.Value("request_id").(string); ok {
        result["request_id"] = reqID
    }
    
    return result
}

logger := logger.New(logger.LoggerConfig{
    ContextExtractor: customContextExtractor,
    // ... other config
})

```
### Development Setup

```bash
# Clone the repository
git clone https://github.com/Lunar-Chipter/mire.git
cd mire

# Setup module
go mod tidy

# Run tests
go test ./...

## 🏗️ Architecture

Mire follows a modular architecture with clear separation of concerns:

```
+------------------+    +---------------------+    +------------------+
|   Your App       | -> |   Logger Core       | -> |   Formatters     |
|   (log.Info())   |    |   (configuration,   |    |   (Text, JSON,   |
+------------------+    |    filtering,       |    |    CSV)          |
                        |    pooling)         |    +------------------+
                        +---------------------+
                        |   Writers           |
                        |   (async, buffered, |
                        |    rotating)        |
                        +---------------------+
                        |   Hooks             |
                        |   (custom           |
                        |    processing)      |
                        +---------------------+

### Key Components

1. **Logger Core**: Manages configuration, filters, and dispatches log entries
2. **Formatters**: Convert log entries to different output formats with zero-allocation design
3. **Writers**: Handle output to various destinations (console, files, networks)
4. **Object Pools**: Reuse objects to minimize allocations and garbage collection
5. **Hooks**: Extensible system for custom log processing
6. **Clock**: Clock for timestamp operations with minimal overhead

## 📚 Examples

### Zero-Allocation Logging Example

```go
package main

import (
    "context"
    "os"

    "github.com/Lunar-Chipter/mire/core"
    "github.com/Lunar-Chipter/mire/formatter"
    "github.com/Lunar-Chipter/mire/logger"
    "github.com/Lunar-Chipter/mire/util"
)

func main() {
    // Create a high-performance logger optimized for zero-allocation
    log := logger.New(logger.LoggerConfig{
        Level:   core.INFO,
        Output:  os.Stdout,
        Formatter: &formatter.TextFormatter{
            EnableColors:    true,
            ShowTimestamp:   true,
            ShowCaller:      true,
            ShowTraceInfo:   true,
        },
        AsyncLogging:        true,
        AsyncWorkerCount:    4,
        AsyncLogChannelBufferSize: 2000,
    })
    defer log.Close()

    // Context with trace information
    ctx := context.Background()
    ctx = util.WithTraceID(ctx, "trace-12345")
    ctx = util.WithUserID(ctx, "user-67890")

    // Zero-allocation logging using []byte internally
    log.WithFields(map[string]interface{}{
        "user_id": 12345,
        "action":  "purchase",
        "amount":  99.99,
    }).Info("Transaction completed")

    // Context-aware logging with distributed tracing
    log.InfoC(ctx, "Processing request") // Includes trace_id and user_id automatically
}
```

### CSV Formatter Usage

```go
package main

import (
    "os"
    "github.com/Lunar-Chipter/mire/core"
    "github.com/Lunar-Chipter/mire/formatter"
    "github.com/Lunar-Chipter/mire/logger"
)

func main() {
    // Create a CSV logger to write to a file
    file, err := os.Create("app.csv")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    csvLogger := logger.New(logger.LoggerConfig{
        Level:   core.INFO,
        Output:  file,
        Formatter: &formatter.CSVFormatter{
            IncludeHeader:   true,                    // Include CSV header row
            FieldOrder:      []string{"timestamp", "level", "message", "user_id", "action"}, // Custom field order
            TimestampFormat: "2006-01-02T15:04:05",   // Custom timestamp format
            SensitiveFields: []string{"password", "token"}, // Fields to mask
            MaskSensitiveData: true,                  // Enable masking
            MaskStringValue: "[MASKED]",             // Mask value
        },
    })
    defer csvLogger.Close()

    csvLogger.WithFields(map[string]interface{}{
        "user_id": 123,
        "action":  "login",
        "status":  "success",
    }).Info("User login event")

    csvLogger.WithFields(map[string]interface{}{
        "user_id": 456,
        "action":  "purchase",
        "amount":  99.99,
    }).Info("Purchase completed")
}
```

### Asynchronous Logging

```go
asyncLogger := logger.New(logger.LoggerConfig{
    Level:                core.INFO,
    Output:              os.Stdout,
    AsyncLogging:        true,
    AsyncWorkerCount:    4,
    AsyncLogChannelBufferSize: 1000,
    LogProcessTimeout:   time.Second,
    Formatter: &formatter.TextFormatter{
        EnableColors:    true,
        ShowTimestamp:   true,
        ShowCaller:      true,
    },
})
defer asyncLogger.Close()

// This will be processed asynchronously
for i := 0; i < 1000; i++ {
    asyncLogger.WithFields(map[string]interface{}{
        "iteration": i,
    }).Info("Async log message")
}
```

### Context-Aware Logging with Distributed Tracing

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    // Extract tracing information from request context
    ctx := r.Context()
    ctx = util.WithTraceID(ctx, generateTraceID())
    ctx = util.WithRequestID(ctx, generateRequestID())

    // Use context-aware logging methods
    log.InfoC(ctx, "Processing HTTP request")

    // Add user-specific context
    ctx = util.WithUserID(ctx, getUserID(r))

    log.WithFields(map[string]interface{}{
        "path": r.URL.Path,
        "method": r.Method,
    }).InfofC(ctx, "Request details")
}
```

### Custom Hook Integration

```go
// Implement a custom hook
type CustomHook struct {
    endpoint string
}

func (h *CustomHook) Fire(entry *core.LogEntry) error {
    // Send log entry to external service
    payload, err := json.Marshal(entry)
    if err != nil {
        return err
    }

    resp, err := http.Post(h.endpoint, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (h *CustomHook) Close() error {
    // Cleanup resources
    return nil
}

// Use the custom hook
customHook := &CustomHook{endpoint: "https://logs.example.com/api"}
log := logger.New(logger.LoggerConfig{
    Level: core.INFO,
    Output: os.Stdout,
    Hooks: []hook.Hook{customHook},
    Formatter: &formatter.TextFormatter{
        EnableColors:  true,
        ShowTimestamp: true,
    },
})
```

### Log Rotation Configuration

```go
rotationConfig := &config.RotationConfig{
    MaxSize:    100, // 100MB
    MaxAge:     30,  // 30 days
    MaxBackups: 5,   // Keep 5 old files
    Compress:   true, // Compress rotated files
}

logger := logger.New(logger.LoggerConfig{
    Level:          core.INFO,
    Output:         os.Stdout,
    EnableRotation: true,
    RotationConfig: rotationConfig,
    Formatter: &formatter.JSONFormatter{
        TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
    },
})
```

## 🔧 Advanced Configuration

### Environment-Based Configuration

```go
func getLoggerForEnv(env string) *logger.Logger {
    baseConfig := logger.LoggerConfig{
        Formatter: &formatter.JSONFormatter{
            ShowTimestamp: true,
            ShowCaller:    true,
        },
        ShowHostname:    true,
        ShowApplication: true,
        Environment:     env,
    }

    switch env {
    case "production":
        baseConfig.Level = core.INFO
        baseConfig.Output = os.Stdout
    case "development":
        baseConfig.Level = core.DEBUG
        baseConfig.Formatter = &formatter.TextFormatter{
            EnableColors:  true,
            ShowTimestamp: true,
            ShowCaller:    true,
        }
    case "testing":
        baseConfig.Level = core.WARN
        baseConfig.Output = io.Discard // Discard logs during testing
    }
    return logger.New(baseConfig)
}
```

### Conditional Logging Based on Context

```go
func conditionalLog(ctx context.Context, log *logger.Logger) {
    // Extract user role from context and adjust logging behavior
    userRole := ctx.Value("role")
    if userRole == "admin" {
        log.InfoC(ctx, "Admin action performed")
    } else {
        log.DebugfC(ctx, "Regular user action: %v", ctx.Value("action"))
    }
}
```

### Custom Metrics Integration

```go
import "github.com/Lunar-Chipter/mire/metric"

// Create a custom metrics collector
customMetrics := metric.NewDefaultMetricsCollector()

log := logger.New(logger.LoggerConfig{
    Level:            core.INFO,
    Output:           os.Stdout,
    EnableMetrics:    true,
    MetricsCollector: customMetrics,
    Formatter: &formatter.TextFormatter{
        EnableColors:  true,
        ShowTimestamp: true,
    },
})

// Use the logger
log.Info("Test message")

// Access metrics
count := customMetrics.GetCounter("log.info")
```

### Custom Context Extractor

```go
// Define a custom context extractor function
func customContextExtractor(ctx context.Context) map[string]string {
    result := make(map[string]string)

    // Extract custom values from context
    if reqID := ctx.Value("request_id"); reqID != nil {
        if idStr, ok := reqID.(string); ok {
            result["request_id"] = idStr
        }
    }

    if tenantID := ctx.Value("tenant_id"); tenantID != nil {
        if idStr, ok := tenantID.(string); ok {
            result["tenant"] = idStr
        }
    }

    return result
}

// Use the custom extractor in logger config
log := logger.New(logger.LoggerConfig{
    Level:             core.INFO,
    Output:            os.Stdout,
    ContextExtractor:  customContextExtractor,
    Formatter: &formatter.JSONFormatter{
        ShowTimestamp: true,
        ShowTraceInfo: true,
    },
})
```

## 🤝 Contributing

We welcome contributions to the Mire project!

### Getting Started

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/mire.git
cd mire

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...
```

### Code Standards

- Follow Go formatting conventions (`go fmt`)
- Write comprehensive tests for new features
- Document exported functions and types
- Maintain backward compatibility when possible
- Write clear commit messages

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 📞 Support

Need help? Join our community:

- Issues: [GitHub Issues](https://github.com/Lunar-Chipter/mire/issues)
- Discussions: [GitHub Discussions](https://github.com/Lunar-Chipter/mire/discussions)

## 📄 Changelog

### v0.0.4 - Bug Fixes Release

- **Bug Fixes**: Fixed several benchmark test errors
- **Stability**: Resolved concurrent logging issues
- **Testing**: Updated tests to ensure reliability

### v0.0.4 - Zero-Allocation Redesign

- **Major Enhancement**: Complete internal redesign with []byte fields to eliminate string conversion overhead
- **Performance**: Achieved near-zero allocation performance with improved formatter efficiency
- **Architecture**: Refactored core components for memory hierarchy optimization
- **Features**: Enhanced context extraction and distributed tracing support
- **Stability**: Numerous bug fixes and stability improvements

### v0.0.3 - Enhanced Features

- Added support for custom context extractors
- Implemented advanced field transformers
- Introduced log sampling for high-volume scenarios
- Added comprehensive test coverage
- Improved documentation and examples

### v0.0.2 - Feature Expansion

- Added JSON and CSV formatters
- Implemented hook system for custom log processing
- Added log rotation capabilities
- Enhanced asynchronous logging
- Added metrics collection

### v0.0.1 - Initial Release

- Basic text logging with color support
- Context-aware logging with trace IDs
- Structured logging with fields
- Simple configuration options

### Reporting Issues

When reporting issues, please include:
- Go version (`go version`)
- Operating system
- Mire version
- Expected behavior
- Actual behavior
- Steps to reproduce
- Any relevant logs or error messages

## 🗺️ Roadmap

### Planned Enhancements

#### Performance & Reliability
- [ ] Perfect goroutine ID detection for truly scalable local storage
- [ ] Implement advanced memory prefetching strategies
- [ ] Optimize memory layout to further reduce cache misses
- [ ] Enhance error handling for extreme resource exhaustion scenarios

#### Advanced Features
- [ ] Add structured query capabilities on log entries
- [ ] Implement log compression for storage efficiency
- [ ] Create custom formatter plugin system
- [ ] Develop real-time log streaming and monitoring

#### Integration & Ecosystem
- [ ] Add exporters for popular metric systems (Prometheus, OpenTelemetry)
- [ ] Create comprehensive API documentation
- [ ] Develop integration guides for various Go frameworks
- [ ] Implement sensitive data masking and security mechanisms

## 📄 License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

Copyright 2025 Mire Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## 📞 Support

If you encounter issues or have questions:

- Check the [existing issues](https://github.com/Lunar-Chipter/mire/issues)
- Create a new issue with detailed information
- Include your Go version and platform information
- Provide minimal code to reproduce the issue

### Community

- Join our [Discussions](https://github.com/Lunar-Chipter/mire/discussions) for Q&A
- Follow us for updates and announcements

## 📄 Changelog

### v0.0.4
- **Zero-Allocation Improvements**: Overhauled `LogEntry` structure to use `[]byte` instead of `string` for critical fields
- **Enhanced Performance**: Direct byte slice operations reducing memory allocations
- **Formatter Efficiency**: All formatters updated to handle `[]byte` fields directly
- **API Compatibility**: Maintained backward compatibility with internal performance improvements

### v0.0.3
- Enhanced function naming consistency across all packages for improved readability
- Renamed `S2b` function to `StringToBytes` in both `core` and `util` packages for clearer semantics
- Renamed `ManualByteWrite` to `formatLogToBytes` in core package for better clarity
- Renamed buffer conversion functions: `writeIntToBuffer`, `writeInt64ToBuffer`, `writeFloatToBuffer` to `intToBytes`, `int64ToBytes`, `floatToBytes`
- Renamed utility functions: `shortID` to `shortenID` and `shortIDBytes` to `shortIDToBytes` in formatter package
- Improved code maintainability with more consistent and intuitive function names
- Optimized zero-allocation performance with enhanced string-to-byte conversion functions
- Standardized exported function naming conventions across all packages

### v0.0.2
- Major performance improvements with zero-allocation formatters
- TextFormatter now runs at ~0.13μs/op
- JSONFormatter now runs at ~2.4μs/op
- Added complete CSV formatter with zero-allocation implementation
- Added field transformers support for all formatters
- Added comprehensive sensitive data masking capabilities
- Improved object pooling for high memory efficiency
- Added clock implementation for timestamp operations
- Updated README with comprehensive examples for all formatters
- Added formatter benchmark tests with updated performance metrics
- Improved cache-friendly memory access patterns
- Enhanced branch prediction optimizations
- Added utility functions for zero-allocation operations

## 🔍 Related Projects

- [zap](https://github.com/uber-go/zap) - Blazing fast, structured, leveled logging in Go
- [logrus](https://github.com/sirupsen/logrus) - Structured, pluggable logging for Go
- [zerolog](https://github.com/rs/zerolog) - Zero-allocation JSON logger

## 🙏 Acknowledgments

- Inspired by other efficient logging libraries
- Thanks to the Go community for performance optimization techniques
- Special thanks to contributors and early adopters
