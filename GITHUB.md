# Mire - GitHub Deployment Information

## Project Overview

Mire is a high-performance logging library for Go that focuses on zero-allocation design and maximum throughput. This document provides information specifically for GitHub deployment and usage.

## Deployment to GitHub

### Repository Structure
```
mire/
├── .github/              # GitHub configuration (workflows, issue templates)
├── config/              # Configuration related code
├── core/                # Core data structures and types
├── errors/              # Error handling utilities
├── example/             # Example implementations
├── formatter/           # Format implementations (text, JSON, CSV)
├── hook/                # Hook system for extending functionality
├── logger/              # Main logger implementation
├── metric/              # Metrics collection
├── sampler/             # Log sampling functionality
├── util/                # Utility functions
├── writer/              # Output writers (buffered, async, etc.)
├── main.go              # Example main application
├── main_test.go         # Main application tests
├── README.md            # Main documentation
├── USAGE.md             # Usage guide
├── ARCHITECTURE.md      # Architecture documentation
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── LICENSE              # License information
```

## GitHub Actions Configuration

To properly deploy and test Mire in a GitHub Actions environment, use the following workflow:

### .github/workflows/test.yml
```yaml
name: Test

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22, 1.23]

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: Verify Go version
      run: go version

    - name: Get dependencies
      run: go mod download

    - name: Build
      run: go build -v ./...

    - name: Test
      run: go test -v ./...

    - name: Race Condition Tests
      run: go test -race -v ./...
```

### .github/workflows/lint.yml
```yaml
name: Lint

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  golangci:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: 1.23
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
```

## Usage in GitHub Projects

### Importing Mire
```go
import "github.com/Lunar-Chipter/mire/logger"
```

### Basic GitHub CI/CD Integration
```go
// Example of using Mire in a service that runs in CI/CD
func setupLogger(environment string) *logger.Logger {
    config := logger.LoggerConfig{
        Level:   core.INFO,
        Output:  os.Stdout,
        Formatter: &formatter.JSONFormatter{
            TimestampFormat: time.RFC3339,
        },
        Environment: environment,
        Application: "github-action-service",
    }
    
    return logger.New(config)
}
```

## Performance on GitHub Infrastructure

Mire is optimized to perform well in containerized environments like those used by GitHub Actions:

- Minimal memory allocations reduce GC pressure
- Async logging ensures non-blocking operations
- Configurable buffering for different I/O patterns
- Context-aware logging for distributed tracing

## Best Practices for GitHub Usage

1. **Use structured logging**: Mire's JSON formatter is ideal for log aggregation in CI/CD systems
2. **Configure appropriate log levels**: Use DEBUG for development, INFO/WARN/ERROR for production
3. **Leverage hooks**: Use hooks to send logs to external services for analysis
4. **Use context extraction**: Enable distributed tracing with request IDs and user IDs

## Troubleshooting GitHub Deployments

### Common Issues and Solutions:

1. **Build failures**: Ensure Go 1.21+ is used
   ```yaml
   - name: Set up Go
     uses: actions/setup-go@v4
     with:
       go-version: '1.21'  # or higher
   ```

2. **Test failures**: Check that all tests pass in different environments
   ```bash
   go test -v ./...
   ```

3. **Performance issues**: Use async logging configuration for high-throughput scenarios
   ```go
   config.AsyncLogging = true
   config.AsyncWorkerCount = 4
   ```

## Releases and Versioning

Mire follows semantic versioning. For production use, reference specific versions in your go.mod:

```
require github.com/Lunar-Chipter/mire v0.0.4
```

## Security Considerations

- Mire implements field masking for sensitive data
- Context extraction is configurable to prevent information leakage
- All logging operations are designed to be safe and not crash applications

## Support

For issues and support, please file an issue in the GitHub repository. Include:
- Go version
- Operating system
- Mire version
- Relevant code snippets
- Expected vs actual behavior

For performance issues, please include benchmark results where possible.