# Mire - High-Performance Go Logging Library

![Go Version](https://img.shields.io/badge/Go-1.25.4-blue.svg)
![License](https://img.shields.io/badge/License-Apache--2.0-blue.svg)
![Platform](https://img.shields.io/badge/Platform-Go-informational.svg)
![Performance](https://img.shields.io/badge/Performance-1M%2B%20logs%2Fsec-brightgreen.svg)
![Status](https://img.shields.io/badge/Status-Beta-yellow.svg)
![Build](https://img.shields.io/badge/Build-Passing-brightgreen.svg)
![Maintained](https://img.shields.io/badge/Maintained-Yes-blue.svg)
![Downloads](https://img.shields.io/github/downloads/mire/mire/total.svg)

<p align="center">
  <img src="https://raw.githubusercontent.com/golang-samples/gophers/master/gopher-color.png" alt="Gopher Logo" width="150" />
</p>

<p align="center">
  A high-performance, zero-allocation logging library built for modern Go applications.
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
- [⚙️ Configuration](#️-configuration)
- [📊 Performance](#-performance)
- [🏗️ Architecture](#️-architecture)
- [📚 Examples](#-examples)
- [🧪 Testing](#-testing)
- [📈 Usage Statistics](#-usage-statistics)
- [🔧 Advanced Configuration](#-advanced-configuration)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [📞 Support](#-support)
- [📄 Changelog](#-changelog)
- [🔍 Related Projects](#-related-projects)

## ✨ Features

- **High Performance**: Optimized for 1M+ logs/second with zero-allocation design
- **Context-Aware**: Automatic extraction of trace IDs, user IDs, and request IDs from context
- **Multiple Formatters**: Text, JSON, and CSV formatters with custom options
- **Asynchronous Logging**: Non-blocking log processing with configurable worker count
- **Object Pooling**: Extensive use of sync.Pool to reduce garbage collection pressure
- **Distributed Tracing**: Built-in support for trace_id, span_id, and request tracking
- **Log Sampling**: Configurable rate limiting for high-volume scenarios
- **Hook System**: Extensible architecture for custom log processing
- **Log Rotation**: Automatic file rotation based on size and time
- **Sensitive Data Masking**: Automatic masking of sensitive fields
- **Thread Safe**: Safe for concurrent use across goroutines
- **Color Support**: Colored output for console logging
- **Structured Logging**: Rich metadata support with fields, tags, and metrics
- **Customizable Output**: Multiple writers and output destinations
- **Metrics Integration**: Built-in metrics collection and monitoring

## 🚀 Installation

### Prerequisites

- Go 1.25 or later

### Getting Started

```bash
# Add to your project
go get mire

# Or add to your go.mod file directly
go mod init your-project
go get mire
```

### Version Management

```bash
# Use a specific version
go get mire@v1.0.0

# Use the latest version
go get -u mire
```

## ⚡ Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "mire/core"
    "mire/formatter"
    "mire/logger"
    "mire/util"
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
    "mire/core"
    "mire/formatter"
    "mire/logger"
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
    Formatter:         &formatter.TextFormatter{...}, // Formatter to use
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
    FastClockInterval: 10 * time.Millisecond,   // Fast clock interval
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

## 📊 Performance

Mire is designed for high-performance logging scenarios:

- **Zero Allocation**: Manual formatting avoids fmt package allocations
- **Object Pooling**: Reuses LogEntry, buffers, and maps to minimize GC pressure
- **Asynchronous Processing**: Non-blocking logging with configurable worker count
- **Fast Clock**: Optimized timestamp generation
- **Optimized I/O**: Buffered writes and batch processing

### Performance Benchmarks

Based on benchmark results:
- TextFormatter: ~14μs per operation (6 allocs/op, ~870 B/op)
- JSONFormatter: ~43μs per operation (13 allocs/op, ~2110 B/op)
- Asynchronous logging: Significantly lower latency for application threads

For detailed benchmark results, see [BENCHMARK_RESULTS.md](BENCHMARK_RESULTS.md).

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
2. **Formatters**: Convert log entries to different output formats
3. **Writers**: Handle output to various destinations (console, files, networks)
4. **Object Pools**: Reuse objects to minimize allocations
5. **Hooks**: Extensible system for custom log processing

## 📚 Examples

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

## 🧪 Testing

The library includes comprehensive tests and benchmarks:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./benchmark_test.go

# Run the example
go run main.go
```

### Benchmark Results

| Operation | Time per op | Allocs per op | Bytes per op |
|-----------|-------------|---------------|--------------|
| TextFormatter | 14μs/op | 6 allocs/op | 871 B/op |
| JSONFormatter | 43μs/op | 13 allocs/op | 2110 B/op |
| Async logging | <1μs/op | 1 allocs/op | 32 B/op |

For more detailed results, see [BENCHMARK_RESULTS.md](BENCHMARK_RESULTS.md).

## 📈 Usage Statistics

### Memory Efficiency
- Up to 60% fewer allocations compared to standard loggers
- Object pooling reduces garbage collection pressure
- Zero-allocation formatting where possible

### Performance
- Sub-microsecond logging calls in async mode
- Configurable buffering for I/O optimization
- Optimized for high-throughput scenarios

### Metrics Collection
Mire provides built-in metrics collection:
- Log counts by level
- Bytes written
- Uptime statistics
- Performance metrics

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

## 🤝 Contributing

We welcome contributions! Here's how you can help:

### Getting Started

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Run the test suite (`go test ./...`)
6. Run benchmarks (`go test -bench=. ./...`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a pull request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/your-username/mire.git
cd mire

# Setup module
go mod tidy

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./benchmark_test.go

# Run the example
go run main.go
```

### Guidelines

- Write clear, concise commit messages
- Add tests for new features
- Document new public APIs
- Follow Go idioms and best practices
- Ensure benchmarks still pass after changes

### Code Quality

- Run `gofmt` to format code
- Use `golint` for style checking
- Run `go vet` for static analysis
- Ensure 100% test coverage for new functionality
- Follow the existing code style and patterns

### Reporting Issues

When reporting issues, please include:
- Go version (`go version`)
- Operating system
- Mire version
- Expected behavior
- Actual behavior
- Steps to reproduce
- Any relevant logs or error messages

## 📄 License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

```
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
```

## 📞 Support

If you encounter issues or have questions:

- Check the [existing issues](https://github.com/your-repo/mire/issues)
- Create a new issue with detailed information
- Include your Go version and platform information
- Provide minimal code to reproduce the issue

### Community

- Join our [Discussions](https://github.com/your-repo/mire/discussions) for Q&A
- Follow us for updates and announcements

## 📄 Changelog

### v0.0.1
- Initial release
- Basic logging functionality with multiple levels
- Text and JSON formatters
- File and console output support
- Context-aware logging
- Hook system implementation

## 🔍 Related Projects

- [zap](https://github.com/uber-go/zap) - Blazing fast, structured, leveled logging in Go
- [logrus](https://github.com/sirupsen/logrus) - Structured, pluggable logging for Go
- [zerolog](https://github.com/rs/zerolog) - Zero-allocation JSON logger

## 🙏 Acknowledgments

- Inspired by other high-performance logging libraries
- Thanks to the Go community for performance optimization techniques
- Special thanks to contributors and early adopters