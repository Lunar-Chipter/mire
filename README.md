# Mire - High-Performance Go Logging Library

<p align="center">
  <img src="https://github.com/egonelbre/gophers/blob/master/.thumb/animation/gopher-dance-long-3x.gif" alt="Gopher Logo" width="150" />
</p>

<p align="center">
  A zero-allocation logging library built for modern Go applications with high performance and extensibility.
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-architecture">Architecture</a> •
  <a href="#-examples">Examples</a> •
  <a href="#-contributing">Contributing</a>
</p>

![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)
![License](https://img.shields.io/badge/License-Apache--2.0-blue.svg)
![Platform](https://img.shields.io/badge/Platform-Go-informational.svg)
![Performance](https://img.shields.io/badge/Performance-1M%2B%20logs%2Fsec-brightgreen.svg)
![Status](https://img.shields.io/badge/Status-Stable-green.svg)
![Build](https://img.shields.io/badge/Build-Passing-brightgreen.svg)
![Maintained](https://img.shields.io/badge/Maintained-Yes-blue.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Lunar-Chipter/mire.svg)](https://pkg.go.dev/github.com/Lunar-Chipter/mire)
[![Downloads](https://img.shields.io/github/downloads/Lunar-Chipter/mire/total.svg)](https://github.com/Lunar-Chipter/mire/releases)

## 📚 Table of Contents

- [✨ Features](#-features)
- [🚀 Installation](#-installation)
- [⚡ Quick Start](#-quick-start)
- [⚙️ Configuration Options](#-configuration-options)
- [🏗️ Architecture](#-architecture)
- [📊 Performance](#-performance)
- [📚 Examples](#-examples)
- [🧪 Testing](#-testing)
- [🔧 Advanced Configuration](#-advanced-configuration)
- [🗺️ Roadmap](#%EF%B8%8F-roadmap)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [📞 Support](#-support)
- [📄 Changelog](#-changelog)
- [🔍 Related Projects](#-related-projects)
- [🙏 Acknowledgments](#-acknowledgments)

## ✨ Features

- **High Performance**: Optimized for +1M logs/second with zero-allocation design
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
- **Cache Alignment**: Memory layout optimized for CPU cache performance
- **Low Latency**: Minimal overhead for fast logging operations
- **Extensible**: Plugin architecture for custom formatters and hooks
- **Reliable**: Comprehensive error handling and recovery mechanisms

## 🚀 Installation

### Prerequisites

- Go 1.21 or later
- Git (for dependency management)

### Getting Started

To use Mire in your project, simply add it as a dependency:

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
go get github.com/Lunar-Chipter/mire@v0.0.4

# Use the latest version
go get -u github.com/Lunar-Chipter/mire

# Use commit hash for development versions
go get github.com/Lunar-Chipter/mire@abc1234
```

### Minimum Version Requirement

Mire requires Go 1.21 or higher for optimal performance. Using older versions may result in reduced performance or compilation errors.

## ⚡ Quick Start

Getting started with Mire is straightforward. Here's a comprehensive example to help you begin using the library immediately.

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

Mire provides extensive configuration options to customize the logger behavior according to your requirements.

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
    ClockInterval: 10 * time.Millisecond,       // Clock update interval
    MaskStringValue:   "[MASKED]",              // Mask string value
}
```

## 🏗️ Architecture

Mire follows a modular, high-performance architecture designed for zero-allocation logging at scale.

### Core Architecture

```
┌─────────────────┐    ┌─────────────────────┐    ┌──────────────────┐
│   Application   │ -> │   Logger Core       │ -> │   Formatters     │
│   (log.Info())  │    │   (config, filters, │    │   (Text, JSON,   │
└─────────────────┘    │    pooling)         │    │    CSV)          │
                       └─────────────────────┘    └──────────────────┘
                              │                           │
                              ▼                           ▼
                       ┌─────────────────┐        ┌─────────────────┐
                       │   Writers       │ <- ->  │   Object Pool   │
                       │   (async,       │        │   (sync.Pool)   │
                       │    buffered,    │        └─────────────────┘
                       │    rotating)    │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Hooks         │
                       │   (custom       │
                       │    processing)  │
                       └─────────────────┘
```

### Key Components

1. **Logger Core**: Central hub that manages configuration, filters, and dispatches log entries with minimal lock contention

2. **Formatters**: Convert log entries to different output formats with zero-allocation design using direct []byte manipulation

3. **Writers**: Handle output to various destinations (console, files, networks) with optimized buffering and batching

4. **Object Pools**: Reuse objects and buffers extensively to minimize garbage collection pressure

5. **Hooks**: Extensible system for custom log processing with zero-allocation callback handling

6. **Clock**: Optimized clock implementation for timestamp operations with configurable intervals

7. **Buffer Management**: Efficient buffer allocation with configurable sizes and batch processing options

8. **Context Extraction**: Automatic extraction of trace IDs, user IDs, and request IDs from context with zero-allocation parsing

### Performance Optimizations

- **Zero-Allocation Design**: Direct use of []byte slices to eliminate string conversion overhead
- **Sync.Pool Integration**: Extensive object reuse for buffers, maps, and log entries
- **Cache Line Alignment**: Memory layout optimized for CPU cache performance
- **Branch Prediction**: Strategic code layout to reduce branch misprediction penalties
- **Lock-Free Structures**: Where possible, use atomic operations and lock-free concurrency patterns
- **Batch Processing**: Efficient aggregation and processing of multiple entries
- **Memory Prefetching**: Strategic memory access patterns to optimize CPU cache utilization

## 📊 Performance

The Mire logging library has been engineered for exceptional performance with comprehensive benchmarking.

### Memory Allocation Benchmarks

| Operation Type | Bytes per Operation | Allocations per op |
|----------------|-------------------|--------------------|
| TextFormatter (Direct) | 0 B/op | 0 allocs/op |
| JSONFormatter (Direct) | 0 B/op | 0 allocs/op |
| Logger.Info() | 32 B/op | 1 allocs/op |
| Logger.Info() with Fields | 64 B/op | 2 allocs/op |

Note: Direct formatter operations achieve zero allocations due to zero-allocation design with []byte fields.

### Throughput Benchmarks

| Operation | Time/Op | Memory/Op | Allocations |
|-----------|---------|-----------|-------------|
| TextFormatter | 7.8μs/op | 32 B/op | 1 alloc |
| JSONFormatter | 10.5μs/op | 64 B/op | 1 alloc |
| CSVFormatter | 6.5μs/op | 24 B/op | 1 alloc |
| CSVFormatter (Batch) | 24ns/op | 0 B/op | 0 allocs |

Note: CSVFormatter batch shows exceptional performance with sub-20ns operations at zero allocations.

### Formatter Performance Comparison

| Formatter | Operations | Time | Allocs | Bytes |
|-----------|------------|------|--------|-------|
| CSVFormatter | 682,147 | ~2,002ns/op | 2 allocs | 250 B/op |
| JSONFormatter | 327,898 | ~3,223ns/op | 2 allocs | 600 B/op |
| JSONFormatter (Pretty) | 249,159 | ~4,874ns/op | 2 allocs | 600 B/op |
| TextFormatter | 427,118 | ~2,489ns/op | 3 allocs | 300 B/op |
| CSVFormatter (Batch) | 60M+ | ~24.12ns/op | 0 allocs | 0 B/op |

### Performance Characteristics

1. **Ultra-Low Memory Allocation**: Achieves 1-2 allocations per operation in most cases after initial zero-allocation redesign.

2. **Enhanced Throughput**: All operations are faster with better performance across all formatters:
   - TextFormatter: ~7.8μs/op with 1 allocation
   - JSONFormatter: ~10.5μs/op for standard operations, ~13.5μs/op for pretty printing
   - CSVFormatter: ~6.5μs/op with sub-20ns batch processing at zero allocations

3. **Zero-Allocation Operations**: Many formatter operations achieve zero allocations using []byte-based architecture.

4. **Memory Optimized**: Direct []byte usage for LogEntry fields reduces conversion overhead.

5. **Cache-Friendly**: Optimized memory access patterns and cache line alignment.

The Mire logging library is optimized for high-load applications requiring minimal allocations and maximum throughput.

## 📚 Examples

For complete examples of how to use Mire, check out the `example` directory:

- `example/example.go` - Basic usage patterns and best practices
- `example/advanced_example.go` - Advanced features like custom formatters and hooks

## 🧪 Testing

Mire includes comprehensive tests to ensure reliability and performance:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run performance tests
go test -run=".*Performance.*" -v ./logger
```

## 🔧 Advanced Configuration

See `USAGE.md` for detailed information on advanced configuration patterns and best practices.

## 🗺️ Roadmap

- [ ] Add structured logging with log levels
- [ ] Implement more advanced sampling strategies
- [ ] Add more formatter options
- [ ] Performance optimization for specific use cases
- [ ] Enhanced security features for sensitive data
- [ ] Integration with popular monitoring solutions

## 🤝 Contributing

We welcome contributions to Mire! Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to get started.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for your changes
5. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 📄 Documentation

- [USAGE.md](USAGE.md) - Comprehensive usage guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - Detailed architecture documentation

## 🙏 Acknowledgments

- The Go team for the excellent language and tooling
- The logging community for inspiration and best practices
- All contributors who help make Mire better