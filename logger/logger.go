package logger

import (
	"bytes"
	"context"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/hook"
	"github.com/Lunar-Chipter/mire/metric"
	"github.com/Lunar-Chipter/mire/sampler"
	"github.com/Lunar-Chipter/mire/util"
	"github.com/Lunar-Chipter/mire/writer"
	"github.com/Lunar-Chipter/mire/config" // Add this import
)
const (
	DEFAULT_TIMESTAMP_FORMAT = "2006-01-02 15:04:05.000"
	DEFAULT_CALLER_DEPTH     = 3
	DEFAULT_BUFFER_SIZE      = 1000
	DEFAULT_FLUSH_INTERVAL   = 5 * time.Second
	
	// Buffer sizes dikonfigurasi saat inisialisasi - aligned with zero-allocation philosophy
	SmallBufferSize          = 512   // Untuk perf-critical
	MediumBufferSize         = 2048  // Untuk standard logs  
	LargeBufferSize          = 8192  // Untuk verbose debugging
)


// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level             core.Level                      // Minimum level to log
	EnableColors      bool                            // Enable ANSI colors in output
	Output            io.Writer                       // Output writer for logs
	ErrorOutput       io.Writer                       // Output writer for internal logger errors
	Formatter         formatter.Formatter             // Formatter to use for log entries
	ShowCaller        bool                            // Show caller information (file, line)
	CallerDepth       int                             // Depth to look for caller info
	ShowGoroutine     bool                            // Show goroutine ID
	ShowPID           bool                            // Show process ID
	ShowTraceInfo     bool                            // Show trace information (trace_id, span_id, etc.)
	ShowHostname      bool                            // Show hostname
	ShowApplication   bool                            // Show application name
	TimestampFormat   string                          // Format for timestamps
	ExitFunc          func(int)                       // Function to call on fatal/panic (defaults to os.Exit)
	EnableStackTrace  bool                            // Enable stack trace for errors
	StackTraceDepth   int                             // Maximum depth for stack trace
	EnableSampling    bool                            // Enable log sampling
	SamplingRate      int                             // Sampling rate (log every Nth message)
	BufferSize        int                             // Size of buffer for buffered writer
	FlushInterval     time.Duration                   // Interval to flush buffered logs
	EnableRotation    bool                            // Enable log rotation
	RotationConfig    *config.RotationConfig          // Configuration for log rotation
	ContextExtractor  func(context.Context) map[string]string // Function to extract fields from context
	Hostname          string                          // Hostname to include in logs
	Application       string                          // Application name to include in logs
	Version           string                          // Application version to include in logs
	Environment       string                          // Environment (dev, prod, etc.)
	MaxFieldSize      int                             // Maximum size for field values
	EnableMetrics     bool                            // Enable metrics collection
	MetricsCollector  metric.MetricsCollector         // Metrics collector to use
	ErrorHandler      func(error)                     // Function to handle internal logger errors
	OnFatal           func(*core.LogEntry)            // Function to call when a fatal log occurs
	OnPanic           func(*core.LogEntry)            // Function to call when a panic log occurs
	Hooks             []hook.Hook                     // Hooks to execute for each log entry
	EnableErrorFileHook bool                          // Enable built-in error file hook for ERROR+ levels
	BatchSize         int                             // Size of batch for batched writes
	BatchTimeout      time.Duration                   // Timeout for batched writes
	DisableLocking    bool                            // Disable internal locking (for performance, use with caution)
	PreAllocateFields int                             // Pre-allocate map capacity for fields
	PreAllocateTags   int                             // Pre-allocate slice capacity for tags
	MaxMessageSize    int                             // Maximum size for log messages
	AsyncLogging      bool                            // Enable asynchronous logging
	LogProcessTimeout time.Duration                   // Timeout for processing log in async worker
	AsyncLogChannelBufferSize int                     // Buffer size for async log channel
	AsyncWorkerCount          int                     // Number of async worker goroutines
	DisablePerLogContextTimeout bool                  // Disable context timeout per log in async mode
	ClockInterval time.Duration                   // Interval for clock (for timestamp optimization)
	MaskStringValue   string                          // String value to use for masking sensitive data
}
// validate ensures the logger configuration has sane defaults.
func validate(c *LoggerConfig) {
	if c.Output == nil {
		c.Output = os.Stdout
	}
	if c.ErrorOutput == nil {
		c.ErrorOutput = os.Stderr
	}
	if c.Formatter == nil {
		// Use default mask string for the default formatter
		defaultFormatter := &formatter.TextFormatter{TimestampFormat: DEFAULT_TIMESTAMP_FORMAT}
		if c.MaskStringValue == "" {
			defaultFormatter.MaskStringBytes = []byte("[MASKED]") // Default mask
		} else {
			defaultFormatter.MaskStringBytes = []byte(c.MaskStringValue)
		}
		c.Formatter = defaultFormatter
	} else {
		// If a formatter is provided, check its type and set MaskStringBytes
		if tf, ok := c.Formatter.(*formatter.TextFormatter); ok {
			if c.MaskStringValue == "" {
				tf.MaskStringBytes = []byte("[MASKED]") // Default mask
			} else {
				tf.MaskStringBytes = []byte(c.MaskStringValue)
			}
		} else if jf, ok := c.Formatter.(*formatter.JSONFormatter); ok {
			if c.MaskStringValue == "" {
				jf.MaskStringBytes = []byte("[MASKED]") // Default mask
			} else {
				jf.MaskStringBytes = []byte(c.MaskStringValue)
			}
		}
	}

	if c.CallerDepth <= 0 {
		c.CallerDepth = DEFAULT_CALLER_DEPTH
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = DEFAULT_FLUSH_INTERVAL
	}
    if c.TimestampFormat == "" {
        c.TimestampFormat = DEFAULT_TIMESTAMP_FORMAT
    }
}
// Logger is the main logging structure
type Logger struct {
	Config           LoggerConfig                    // Configuration for the logger
	formatter        formatter.Formatter             // Formatter to use for log entries
	out              io.Writer                       // Output writer for logs
	errOut           io.Writer                       // Output writer for internal logger errors
	errOutMu         sync.Mutex                      // Mutex for protecting errOut
	mu               *sync.RWMutex                   // Mutex for protecting internal state (changed to pointer to allow safe cloning)
	hooks            []hook.Hook                     // Hooks to execute for each log entry
	exitFunc         func(int)                       // Function to call on fatal/panic
	fields           map[string]interface{}          // Default fields to include in all logs
	sampler          *sampler.SamplingLogger         // Sampler for log sampling
	buffer           *writer.BufferedWriter          // Buffered writer for performance
	rotation         *writer.RotatingFileWriter      // Rotating file writer for log rotation
	contextExtractor func(context.Context) map[string]string // Function to extract fields from context
	metrics          metric.MetricsCollector         // Metrics collector
	onFatal          func(*core.LogEntry)            // Function to call when a fatal log occurs
	onPanic          func(*core.LogEntry)            // Function to call when a panic log occurs
	stats            *LoggerStats                    // Statistics for the logger
	asyncLogger      *writer.AsyncLogger             // Async logger for non-blocking logging
	errorFileHook    *hook.SimpleFileHook            // Built-in error file hook for ERROR+ levels
	closed           atomic.Bool                     // Flag to indicate if logger is closed
	pid              int                             // Process ID
	clock            *util.Clock                 // Clock for timestamp optimization
}

// LoggerStats tracks logger statistics
type LoggerStats struct {
	LogCounts    map[core.Level]int64
	BytesWritten int64
	StartTime    time.Time
	mu           sync.RWMutex
}

// NewLoggerStats creates a new LoggerStats
func NewLoggerStats() *LoggerStats {
	return &LoggerStats{
		LogCounts:    make(map[core.Level]int64),
		BytesWritten: 0,
		StartTime:    time.Now(),
	}
}

// Increment increments the statistics for a log level
func (ls *LoggerStats) Increment(level core.Level, bytes int) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.LogCounts[level]++
	ls.BytesWritten += int64(bytes)
}

// GetStats returns the current statistics
func (ls *LoggerStats) GetStats() map[string]interface{} {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	
	stats := make(map[string]interface{})
	stats["start_time"] = ls.StartTime
	stats["bytes_written"] = ls.BytesWritten
	stats["uptime"] = time.Since(ls.StartTime).String()
	
	counts := make(map[string]int64)
	for level, count := range ls.LogCounts {
		counts[level.String()] = count
	}
	stats["log_counts"] = counts
	
	return stats
}

// NewDefaultLogger creates a logger with default configuration
// This logger is configured with standard settings suitable for most applications
func NewDefaultLogger() *Logger {
	cfg := LoggerConfig{
		Level:             core.INFO,
		Output:            os.Stdout,
		ErrorOutput:       os.Stderr,
		CallerDepth:       DEFAULT_CALLER_DEPTH,
		TimestampFormat:   DEFAULT_TIMESTAMP_FORMAT,
		BufferSize:        DEFAULT_BUFFER_SIZE,
		FlushInterval:     DEFAULT_FLUSH_INTERVAL,
		AsyncWorkerCount:  4,
		ClockInterval: 10 * time.Millisecond,
		MaskStringValue:   "[MASKED]", // Set default mask string value here
		Formatter: &formatter.TextFormatter{
			EnableColors:      true,
			ShowTimestamp:     true,
			ShowCaller:        true,
			TimestampFormat:   DEFAULT_TIMESTAMP_FORMAT,
		},
	}
	return New(cfg)
}

// New creates a new logger with the given configuration
// Optimized for 1M+ logs/second with early filtering
func New(config LoggerConfig) *Logger {
	validate(&config)

	l := &Logger{
		Config:           config,
		formatter:        config.Formatter,
		out:              config.Output,
		errOut:           config.ErrorOutput,
		mu:               new(sync.RWMutex), // Initialize the mutex pointer
		exitFunc:         config.ExitFunc,
		fields:           make(map[string]interface{}),
		hooks:            config.Hooks, // Initialize hooks from config
		contextExtractor: config.ContextExtractor,
		metrics:          config.MetricsCollector,
		onFatal:          config.OnFatal,
		onPanic:          config.OnPanic,
		stats:            NewLoggerStats(),
		pid:              os.Getpid(),
	}

	if config.EnableErrorFileHook {
		errorHook, err := hook.NewFileHook("errors.log") // Use NewFileHook from hook package
		if err != nil {
			l.handleError(newErrorf("failed to create error file hook: %v", err))
		} else {
			l.errorFileHook = errorHook
			l.hooks = append(l.hooks, errorHook)
		}
	}

	if config.ClockInterval > 0 {
		l.clock = util.NewClock(config.ClockInterval)
	}

	if l.exitFunc == nil {
		l.exitFunc = os.Exit
	}
	
l.setupWriters()
	
	if config.EnableSampling && config.SamplingRate > 1 {
		l.sampler = sampler.NewSamplingLogger(l, config.SamplingRate)
	}

	if config.AsyncLogging {
		l.asyncLogger = writer.NewAsyncLogger(l, config.AsyncWorkerCount, config.AsyncLogChannelBufferSize, config.LogProcessTimeout, config.DisablePerLogContextTimeout)
	}

	return l
}

func (l *Logger) setupWriters() {
	currentWriter := l.Config.Output

	if l.Config.EnableRotation && l.Config.RotationConfig != nil {
		if file, ok := currentWriter.(*os.File); ok {
			var err error
			l.rotation, err = writer.NewRotatingFileWriter(file.Name(), l.Config.RotationConfig)
			if err == nil {
				currentWriter = l.rotation
			} else {
				l.handleError(newErrorf("failed to setup rotation: %v", err))
			}
		}
	}

	if l.Config.BufferSize > 0 {
		l.buffer = writer.NewBufferedWriter(currentWriter, l.Config.BufferSize, l.Config.FlushInterval, l.handleError, l.Config.BatchSize, l.Config.BatchTimeout)
		currentWriter = l.buffer
	}

	l.out = currentWriter
}


// These methods are to satisfy interfaces for async/sampler writers
func (l *Logger) Log(ctx context.Context, level core.Level, msg []byte, fields map[string]interface{}) {
    l.write(ctx, level, msg, fields)
}
func (l *Logger) ErrorHandler() func(error) { return l.handleError }
func (l *Logger) ErrOut() io.Writer { return l.errOut }
func (l *Logger) ErrOutMu() *sync.Mutex { return &l.errOutMu }


// internal logging method optimized for 1M+ logs/second
// Early filtering to avoid unnecessary work
func (l *Logger) log(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) {
	// Early return if logger is closed
	if l.closed.Load() {
		return
	}

	// Early filtering to avoid unnecessary work - branch prediction optimized
	if level < l.Config.Level {
		return
	}

    // Sampling if enabled
    if l.sampler != nil && !l.sampler.ShouldLog() {
        return
	}

	// Optimized path for non-blocking scenarios using atomic operations
	if l.asyncLogger != nil {
		// Use lock-free async logging for high throughput
		l.asyncLogger.Log(level, message, fields, ctx)
		return
	}

	// Hot path is efficient
	l.write(ctx, level, message, fields)
}

	// final write to output dengan zero-allocation optimizations
func (l *Logger) write(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) {
	entry := l.buildEntry(ctx, level, message, fields)
	
	// Gunakan buffer yang efisien untuk zero-allocation
	buf := util.GetBufferFromPool()
	defer util.PutBufferToPool(buf)

	if err := l.formatter.Format(buf, entry); err != nil {
		l.handleError(err)
		core.PutEntryToPool(entry)
		return
	}
    
    bytesToWrite := buf.Bytes()

	// Optimized write with minimal locking - only lock when actually writing
	if l.Config.DisableLocking {
		// Direct path: no locking at all
		if n, err := l.out.Write(bytesToWrite); err != nil {
			l.handleError(err)
		} else {
			l.stats.Increment(level, n)
		}
	} else {
		// Standard path with proper locking
		l.mu.Lock()
		if n, err := l.out.Write(bytesToWrite); err != nil {
			l.handleError(err)
		} else {
			l.stats.Increment(level, n)
		}
		l.mu.Unlock()
	}

	l.runHooks(entry)

    // must be done after hooks and writing, but before PutEntryToPool
	l.handleLevelActions(level, entry)

	core.PutEntryToPool(entry)
}

// formatArgsToBytes formats variadic arguments into a byte slice with minimal allocations.
// This implementation aims for at most 1 allocation per call.
func (l *Logger) formatArgsToBytes(args ...interface{}) []byte {
	// Use an efficient approach by pre-calculating the total size if possible
	// to reduce allocations during concatenation
	totalLen := 0
	spaceCount := len(args) - 1
	if spaceCount < 0 {
		spaceCount = 0
	}

	// Pre-calculate total length for efficient allocation
	for i, arg := range args {
		if i > 0 {
			totalLen++ // space between args
		}
		switch v := arg.(type) {
		case string:
			totalLen += len(v)
		case []byte:
			totalLen += len(v)
		case int:
			totalLen += len(strconv.FormatInt(int64(v), 10))
		case int64:
			totalLen += len(strconv.FormatInt(v, 10))
		case float64:
			totalLen += len(strconv.FormatFloat(v, 'g', -1, 64))
		case bool:
			if v {
				totalLen += 4 // "true"
			} else {
				totalLen += 5 // "false"
			}
		default:
			totalLen += 15 // estimate for unknown types
		}
	}

	// We only allow one allocation per call: the final byte slice with pre-calculated capacity
	buf := make([]byte, 0, totalLen)

	// Format arguments directly to the pre-allocated slice
	for i, arg := range args {
		if i > 0 {
			buf = append(buf, ' ') // Add space between arguments
		}
		switch v := arg.(type) {
		case string:
			buf = append(buf, core.StringToBytes(v)...) // Use zero-allocation string to byte conversion
		case []byte:
			buf = append(buf, v...)
		case int:
			tempBuf := util.GetSmallByteSliceFromPool()
			result := strconv.AppendInt(tempBuf[:0], int64(v), 10)
			buf = append(buf, result...)
			util.PutSmallByteSliceToPool(tempBuf)
		case int64:
			tempBuf := util.GetSmallByteSliceFromPool()
			result := strconv.AppendInt(tempBuf[:0], v, 10)
			buf = append(buf, result...)
			util.PutSmallByteSliceToPool(tempBuf)
		case float64:
			tempBuf := util.GetSmallByteSliceFromPool()
			result := strconv.AppendFloat(tempBuf[:0], v, 'g', -1, 64)
			buf = append(buf, result...)
			util.PutSmallByteSliceToPool(tempBuf)
		case bool:
			if v {
				buf = append(buf, "true"...)
			} else {
				buf = append(buf, "false"...)
			}
		default:
			// For other types, we fallback to manual conversion to avoid fmt
			// Use a temporary buffer to avoid multiple allocations
			tempBuf := util.GetBufferFromPool()
			manualFormatValue(tempBuf, v)
			buf = append(buf, tempBuf.Bytes()...)
			util.PutBufferToPool(tempBuf)
		}
	}

	return buf
}

// formatfArgsToBytes formats variadic arguments with a format string into a byte slice with minimal allocations.
// This implementation aims for at most 1 allocation per call.
func (l *Logger) formatfArgsToBytes(format string, args ...interface{}) []byte {
	// We only allow one allocation per call: the final byte slice
	buf := util.GetBufferFromPool()
	defer util.PutBufferToPool(buf)

	// Use manual formatting to avoid fmt dependency
	manualFormatWithArgs(buf, format, args...)

	// Final single allocation: copy the buffer content to a new byte slice
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}

// buildEntry creates a log entry with minimal allocations
// Principle: "Control over every byte" - Manual byte manipulation
func (l *Logger) buildEntry(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) *core.LogEntry {
    entry := core.GetEntryFromPool()

    // Use clock if available to avoid allocation
    if l.clock != nil {
        entry.Timestamp = l.clock.Now()
    } else {
        entry.Timestamp = time.Now()
    }

	entry.Level = level
	entry.LevelName = level.String()
	entry.Message = message
	entry.PID = l.pid

	// Copy fields with minimal allocations
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	if fields != nil {
		for k, v := range fields {
			entry.Fields[k] = v
		}
	}

	// Extract context with zero allocation if possible
	if l.contextExtractor != nil {
		contextFields := l.contextExtractor(ctx)
		for k, v := range contextFields {
			entry.Fields[k] = v
		}
	} else if ctx != nil {
        contextData := util.ExtractFromContext(ctx)
        for k, v := range contextData {
            switch k {
            case "trace_id": entry.TraceID = v
            case "span_id": entry.SpanID = v
            case "user_id": entry.UserID = v
            case "session_id": entry.SessionID = v
            case "request_id": entry.RequestID = v
            }
        }
        util.PutMapStringToPool(contextData)
    }

	// Caller info only if required to avoid overhead
	if l.Config.ShowCaller {
		entry.Caller = util.GetCallerInfo(l.Config.CallerDepth)
	}

    // Stack trace only for ERROR level and above
    if l.Config.EnableStackTrace && level >= core.ERROR {
        stackTraceBytes, stackTraceBufPtr := util.GetStackTrace(l.Config.StackTraceDepth)
        entry.StackTrace = stackTraceBytes
        entry.StackTraceBufPtr = stackTraceBufPtr
    }

	return entry
}

// runHooks executes hooks with zero lock contention
// Zero lock contention dengan RLock
func (l *Logger) runHooks(entry *core.LogEntry) {
	// Gunakan RLock untuk read-only access dan zero lock contention
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Early return jika tidak ada hooks
	if len(l.hooks) == 0 {
		return
	}
	
	// Execute hooks dengan graceful error handling
	// Never let logging crash your application
	for _, h := range l.hooks {
		if err := h.Fire(entry); err != nil {
			l.handleError(newErrorf("hook error: %v", err))
		}
	}
}

// AddHook adds a hook to the logger.
func (l *Logger) AddHook(h hook.Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, h)
}

func (l *Logger) handleLevelActions(level core.Level, entry *core.LogEntry) {
	switch level {
	case core.FATAL:
		if l.onFatal != nil {
			l.onFatal(entry)
		}
		l.exitFunc(1)
	case core.PANIC:
		if l.onPanic != nil {
			l.onPanic(entry)
		}
		// BUKAN: panic(entry.Message) // Yang menyebabkan crash aplikasi
		// Handle buffer full dengan graceful degradation
		if l.out != nil {
			// Use a temporary byte buffer to avoid string concatenation allocation
			var msgBuf bytes.Buffer
			msgBuf.WriteString("PANIC: ")
			msgBuf.Write(entry.Message)
			msgBuf.WriteByte('\n')
			l.out.Write(msgBuf.Bytes())
		}
		l.exitFunc(1)
	}
}

// handleError handles errors gracefully without panicking
// Prinsip: "Never let logging crash your application"
func (l *Logger) handleError(err error) {
	// Graceful error handling tanpa panic
	if l.Config.ErrorHandler != nil {
		l.Config.ErrorHandler(err)
	} else {
		// Ignore errors gracefully, jangan crash aplikasi
		l.errOutMu.Lock()
		defer l.errOutMu.Unlock()

		// Use manual formatting with zero-allocation approach
		buf := util.GetBufferFromPool()
		defer util.PutBufferToPool(buf)
		buf.Write([]byte("logger error: "))
		buf.Write(core.StringToBytes(err.Error())) // Use zero-allocation string to byte conversion
		buf.Write([]byte("\n"))
		l.errOut.Write(buf.Bytes())
	}
}

// Close gracefully closes the logger and its writers.
// Handles all edge cases with graceful degradation
func (l *Logger) Close() {
	// Ensure it's only closed once
	if l.closed.CompareAndSwap(false, true) {
		// Close async logger if present
		if l.asyncLogger != nil {
			l.asyncLogger.Close()
		}

		// Close buffered writer if present
		if l.buffer != nil {
			// Graceful degradation during closing
			if err := l.buffer.Close(); err != nil {
				l.handleError(newErrorf("error closing buffered writer: %v", err))
			}
		}

		// Close rotating file writer if present
		if l.rotation != nil {
			// Graceful degradation during closing
			if err := l.rotation.Close(); err != nil {
				l.handleError(newErrorf("error closing rotating file writer: %v", err))
			}
		}

		// Stop clock if present
		if l.clock != nil {
			l.clock.Stop()
		}

		// Close error file hook if present
		if l.errorFileHook != nil {
			// Graceful degradation during closing
			if err := l.errorFileHook.Close(); err != nil {
				l.handleError(newErrorf("error closing error file hook: %v", err))
			}
		}
	}
	// If already closed, do nothing (graceful)
}

// WithFields creates a new logger with additional fields
// These fields will be included in all log entries made with the returned logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := l.clone()
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// clone creates a copy of the logger with shared resources
// Share nothing, own everything for zero lock contention
func (l *Logger) clone() *Logger {
    l.mu.RLock()
    defer l.mu.RUnlock()

    // Perform a shallow copy of the logger. This is critical.
    // We want clone to share the same writers (async, buffer)
    // and their goroutines, not create new ones.
    cloned := *l

    // Create a new map for cloned fields and copy parent fields.
    // Pre-allocate with sufficient capacity to avoid reallocation
    cloned.fields = make(map[string]interface{}, len(l.fields)+10)
    for k, v := range l.fields {
        cloned.fields[k] = v
    }

    return &cloned
}

// --- Level-based logging methods ---

// Trace logs a message with TRACE level
func (l *Logger) Trace(args ...interface{}) { l.log(context.Background(), core.TRACE, l.formatArgsToBytes(args...), nil) }

// Debug logs a message with DEBUG level
func (l *Logger) Debug(args ...interface{}) { l.log(context.Background(), core.DEBUG, l.formatArgsToBytes(args...), nil) }

// Info logs a message with INFO level
func (l *Logger) Info(args ...interface{})  { l.log(context.Background(), core.INFO, l.formatArgsToBytes(args...), nil) }

// Notice logs a message with NOTICE level
func (l *Logger) Notice(args ...interface{}){ l.log(context.Background(), core.NOTICE, l.formatArgsToBytes(args...), nil) }

// Warn logs a message with WARN level
func (l *Logger) Warn(args ...interface{})  { l.log(context.Background(), core.WARN, l.formatArgsToBytes(args...), nil) }

// Error logs a message with ERROR level
func (l *Logger) Error(args ...interface{}) { l.log(context.Background(), core.ERROR, l.formatArgsToBytes(args...), nil) }

// Fatal logs a message with FATAL level and exits the application
func (l *Logger) Fatal(args ...interface{}) { l.log(context.Background(), core.FATAL, l.formatArgsToBytes(args...), nil) }

// Panic logs a message with PANIC level and panics
func (l *Logger) Panic(args ...interface{}) { l.log(context.Background(), core.PANIC, l.formatArgsToBytes(args...), nil) }

// Tracef logs a formatted message with TRACE level
func (l *Logger) Tracef(format string, args ...interface{}) { l.log(context.Background(), core.TRACE, l.formatfArgsToBytes(format, args...), nil) }

// Debugf logs a formatted message with DEBUG level
func (l *Logger) Debugf(format string, args ...interface{}) { l.log(context.Background(), core.DEBUG, l.formatfArgsToBytes(format, args...), nil) }

// Infof logs a formatted message with INFO level
func (l *Logger) Infof(format string, args ...interface{})  { l.log(context.Background(), core.INFO, l.formatfArgsToBytes(format, args...), nil) }

// Noticef logs a formatted message with NOTICE level
func (l *Logger) Noticef(format string, args ...interface{}){ l.log(context.Background(), core.NOTICE, l.formatfArgsToBytes(format, args...), nil) }

// Warnf logs a formatted message with WARN level
func (l *Logger) Warnf(format string, args ...interface{})  { l.log(context.Background(), core.WARN, l.formatfArgsToBytes(format, args...), nil) }

// Errorf logs a formatted message with ERROR level
func (l *Logger) Errorf(format string, args ...interface{}) { l.log(context.Background(), core.ERROR, l.formatfArgsToBytes(format, args...), nil) }

// Fatalf logs a formatted message with FATAL level and exits the application
func (l *Logger) Fatalf(format string, args ...interface{}) { l.log(context.Background(), core.FATAL, l.formatfArgsToBytes(format, args...), nil) }

// Panicf logs a formatted message with PANIC level and panics
func (l *Logger) Panicf(format string, args ...interface{}) { l.log(context.Background(), core.PANIC, l.formatfArgsToBytes(format, args...), nil) }

// Context-aware logging methods

// TraceC logs a message with TRACE level and extracts context information
func (l *Logger) TraceC(ctx context.Context, args ...interface{}) { l.log(ctx, core.TRACE, l.formatArgsToBytes(args...), nil) }

// DebugC logs a message with DEBUG level and extracts context information
func (l *Logger) DebugC(ctx context.Context, args ...interface{}) { l.log(ctx, core.DEBUG, l.formatArgsToBytes(args...), nil) }

// InfoC logs a message with INFO level and extracts context information
func (l *Logger) InfoC(ctx context.Context, args ...interface{})  { l.log(ctx, core.INFO, l.formatArgsToBytes(args...), nil) }

// NoticeC logs a message with NOTICE level and extracts context information
func (l *Logger) NoticeC(ctx context.Context, args ...interface{}){ l.log(ctx, core.NOTICE, l.formatArgsToBytes(args...), nil) }

// WarnC logs a message with WARN level and extracts context information
func (l *Logger) WarnC(ctx context.Context, args ...interface{})  { l.log(ctx, core.WARN, l.formatArgsToBytes(args...), nil) }

// ErrorC logs a message with ERROR level and extracts context information
func (l *Logger) ErrorC(ctx context.Context, args ...interface{}) { l.log(ctx, core.ERROR, l.formatArgsToBytes(args...), nil) }

// FatalC logs a message with FATAL level and extracts context information, then exits the application
func (l *Logger) FatalC(ctx context.Context, args ...interface{}) { l.log(ctx, core.FATAL, l.formatArgsToBytes(args...), nil) }

// PanicC logs a message with PANIC level and extracts context information, then panics
func (l *Logger) PanicC(ctx context.Context, args ...interface{}) { l.log(ctx, core.PANIC, l.formatArgsToBytes(args...), nil) }

// TracefC logs a formatted message with TRACE level and extracts context information
func (l *Logger) TracefC(ctx context.Context, format string, args ...interface{}) { l.log(ctx, core.TRACE, l.formatfArgsToBytes(format, args...), nil) }

// DebugfC logs a formatted message with DEBUG level and extracts context information
func (l *Logger) DebugfC(ctx context.Context, format string, args ...interface{}) { l.log(ctx, core.DEBUG, l.formatfArgsToBytes(format, args...), nil) }

// InfofC logs a formatted message with INFO level and extracts context information
func (l *Logger) InfofC(ctx context.Context, format string, args ...interface{})  { l.log(ctx, core.INFO, l.formatfArgsToBytes(format, args...), nil) }

// NoticefC logs a formatted message with NOTICE level and extracts context information
func (l *Logger) NoticefC(ctx context.Context, format string, args ...interface{}){ l.log(ctx, core.NOTICE, l.formatfArgsToBytes(format, args...), nil) }

// WarnfC logs a formatted message with WARN level and extracts context information
func (l *Logger) WarnfC(ctx context.Context, format string, args ...interface{})  { l.log(ctx, core.WARN, l.formatfArgsToBytes(format, args...), nil) }

// ErrorfC logs a formatted message with ERROR level and extracts context information
func (l *Logger) ErrorfC(ctx context.Context, format string, args ...interface{}) { l.log(ctx, core.ERROR, l.formatfArgsToBytes(format, args...), nil) }

// FatalfC logs a formatted message with FATAL level and extracts context information, then exits the application
func (l *Logger) FatalfC(ctx context.Context, format string, args ...interface{}) { l.log(ctx, core.FATAL, l.formatfArgsToBytes(format, args...), nil) }

// PanicfC logs a formatted message with PANIC level and extracts context information, then panics
func (l *Logger) PanicfC(ctx context.Context, format string, args ...interface{}) { l.log(ctx, core.PANIC, l.formatfArgsToBytes(format, args...), nil) }

// manualFormatValue formats a value without using fmt package
func manualFormatValue(buf *bytes.Buffer, v interface{}) {
	switch val := v.(type) {
	case string:
		buf.Write(core.StringToBytes(val)) // Use zero-allocation conversion
	case []byte:
		buf.Write(val)
	case int:
		tempBuf := util.GetSmallByteSliceFromPool()
		defer util.PutSmallByteSliceToPool(tempBuf)
		result := strconv.AppendInt(tempBuf[:0], int64(val), 10)
		buf.Write(result)
	case int64:
		tempBuf := util.GetSmallByteSliceFromPool()
		defer util.PutSmallByteSliceToPool(tempBuf)
		result := strconv.AppendInt(tempBuf[:0], val, 10)
		buf.Write(result)
	case float64:
		tempBuf := util.GetSmallByteSliceFromPool()
		defer util.PutSmallByteSliceToPool(tempBuf)
		result := strconv.AppendFloat(tempBuf[:0], val, 'g', -1, 64)
		buf.Write(result)
	case bool:
		if val {
			buf.Write([]byte("true"))
		} else {
			buf.Write([]byte("false"))
		}
	case nil:
		buf.Write([]byte("<nil>"))
	default:
		// For remaining types, we'll use a simple representation
		buf.Write([]byte("<unknown-type>"))
	}
}

// manualFormatWithArgs formats with a format string and arguments without using fmt
func manualFormatWithArgs(buf *bytes.Buffer, format string, args ...interface{}) {
	argIndex := 0
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) {
			if format[i+1] == '%' {
				buf.WriteByte('%')
				i++ // Skip the next character
				continue
			}

			// Process format specifier
			if argIndex < len(args) {
				arg := args[argIndex]
				spec := format[i+1] // Simple format specifier handling
				switch spec {
				case 's':
					if s, ok := arg.(string); ok {
						buf.Write(core.StringToBytes(s)) // Use zero-allocation string to byte conversion
					} else {
						manualFormatValue(buf, arg)
					}
				case 'd':
					if d, ok := arg.(int); ok {
						tempBuf := util.GetSmallByteSliceFromPool()
						defer util.PutSmallByteSliceToPool(tempBuf)
						result := strconv.AppendInt(tempBuf[:0], int64(d), 10)
						buf.Write(result)
					} else {
						manualFormatValue(buf, arg)
					}
				case 'f':
					if f, ok := arg.(float64); ok {
						tempBuf := util.GetSmallByteSliceFromPool()
						defer util.PutSmallByteSliceToPool(tempBuf)
						result := strconv.AppendFloat(tempBuf[:0], f, 'g', -1, 64)
						buf.Write(result)
					} else {
						manualFormatValue(buf, arg)
					}
				case 'v':
					manualFormatValue(buf, arg)
				default:
					manualFormatValue(buf, arg)
				}
				argIndex++
				i++ // Skip the format specifier
			} else {
				// No more arguments, just output the %
				buf.WriteByte('%')
			}
		} else {
			buf.WriteByte(format[i])
		}
	}
}

// errorString creates an error with string content without fmt dependency
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// newErrorf creates a formatted error without fmt dependency
func newErrorf(format string, args ...interface{}) error {
	buf := util.GetBufferFromPool()
	defer util.PutBufferToPool(buf)

	manualFormatWithArgs(buf, format, args...)
	return &errorString{s: buf.String()}
}
