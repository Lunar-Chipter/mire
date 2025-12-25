package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Lunar-Chipter/mire/config"
	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/hook"
	"github.com/Lunar-Chipter/mire/metric"
	"github.com/Lunar-Chipter/mire/sampler"
	"github.com/Lunar-Chipter/mire/util"
	"github.com/Lunar-Chipter/mire/writer"
)

const (
	DEFAULT_TIMESTAMP_FORMAT = "2006-01-02 15:04:05.000"
	DEFAULT_CALLER_DEPTH     = 3
	DEFAULT_BUFFER_SIZE      = 1000
	DEFAULT_FLUSH_INTERVAL   = 5 * time.Second

	SmallBufferSize  = 512
	MediumBufferSize = 2048
	LargeBufferSize  = 8192
)

// Config holds configuration for the logger
type Config struct {
	Level                   core.Level                              // Minimum level to log
	UseColors               bool                                    // Use ANSI colors in output
	Output                  io.Writer                               // Output writer for logs
	ErrorOutput             io.Writer                               // Output writer for internal logger errors
	Formatter               formatter.Formatter                     // Formatter to use for log entries
	ShowCaller              bool                                    // Show caller information (file, line)
	CallerDepth             int                                     // Depth to look for caller info
	ShowGoroutine          bool                                    // Show goroutine ID
	ShowPID                 bool                                    // Show process ID
	ShowTrace               bool                                    // Show trace information (trace_id, span_id, etc.)
	ShowHostname            bool                                    // Show hostname
	ShowApp                 bool                                    // Show application name
	TimestampFormat         string                                  // Format for timestamps
	ExitFunc                func(int)                               // Function to call on fatal/panic (defaults to os.Exit)
	IncludeStackTrace       bool                                    // Include stack trace for errors
	StackTraceDepth         int                                     // Maximum depth for stack trace
	EnableSampling         bool                                    // Enable log sampling
	SamplingRate            int                                     // Sampling rate (log every Nth message)
	BufferSize              int                                     // Size of buffer for buffered writer
	FlushInterval           time.Duration                           // Interval to flush buffered logs
	EnableRotation         bool                                    // Enable log rotation
	RotationConfig               *config.RotationConfig                  // Configuration for log rotation
	ExtractContext          func(context.Context) map[string][]byte // Function to extract fields from context as []byte for zero allocation
	Hostname                string                                  // Hostname to include in logs
	Application             string                                  // Application name to include in logs
	Version                 string                                  // Application version to include in logs
	Environment             string                                  // Environment (dev, prod, etc.)
	MaxFieldSize            int                                     // Maximum size for field values
	EnableMetrics          bool                                    // Enable metrics collection
	Collector               metric.Collector                 // Metrics collector to use
	ErrorHandler            func(error)                             // Function to handle internal logger errors
	OnFatal                 func(*core.LogEntry)                    // Function to call when a fatal log occurs
	OnPanic                 func(*core.LogEntry)                    // Function to call when a panic log occurs
	Hooks                   []hook.Hook                             // Hooks to execute for each log entry
	LogErrors               bool                                    // Log errors to error file (ERROR+ levels)
	BatchSize               int                                     // Size of batch for batched writes
	BatchTimeout            time.Duration                           // Timeout for batched writes
	NoLocking               bool                                    // Disable internal locking (for performance, use with caution)
	FieldCapacity           int                                     // Pre-allocate map capacity for fields
	TagCapacity             int                                     // Pre-allocate slice capacity for tags
	MaxMessageSize          int                                     // Maximum size for log messages
	AsyncMode               bool                                    // Enable asynchronous logging
	ProcessTimeout         time.Duration                           // Timeout for processing log in async worker
	ChannelSize             int                                     // Buffer size for async log channel
	WorkerCount             int                                     // Number of async worker goroutines
	NoTimeout              bool                                    // Disable context timeout per log in async mode
	ClockInterval           time.Duration                           // Interval for clock (for timestamp optimization)
	MaskValue               string                                  // String value to use for masking sensitive data
}

// validate ensures the logger configuration has sane defaults
func validate(c *Config) {
	if c.Output == nil {
		c.Output = os.Stdout
	}
	if c.ErrorOutput == nil {
		c.ErrorOutput = os.Stderr
	}
	if c.Formatter == nil {
		// Use default mask string for the default formatter
		defaultFormatter := &formatter.TextFormatter{TimestampFormat: DEFAULT_TIMESTAMP_FORMAT}
		if c.MaskValue == "" {
			defaultFormatter.MaskStringBytes = []byte("[MASKED]") // Default mask
		} else {
			defaultFormatter.MaskStringBytes = []byte(c.MaskValue)
		}
		c.Formatter = defaultFormatter
	} else {
		// If a formatter is provided, check its type and set MaskStringBytes
		if tf, ok := c.Formatter.(*formatter.TextFormatter); ok {
			if c.MaskValue == "" {
				tf.MaskStringBytes = []byte("[MASKED]") // Default mask
			} else {
				tf.MaskStringBytes = []byte(c.MaskValue)
			}
		} else if jf, ok := c.Formatter.(*formatter.JSONFormatter); ok {
			if c.MaskValue == "" {
				jf.MaskStringBytes = []byte("[MASKED]") // Default mask
			} else {
				jf.MaskStringBytes = []byte(c.MaskValue)
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
	Config           Config                            // Configuration for the logger
	formatter        formatter.Formatter                     // Formatter to use for log entries
	out              io.Writer                               // Output writer for logs
	errOut           io.Writer                               // Output writer for internal logger errors
	errOutMu         *sync.Mutex                             // Mutex for protecting errOut
	mu               *sync.RWMutex                           // Mutex for protecting internal state (changed to pointer to allow safe cloning)
	hooks            []hook.Hook                             // Hooks to execute for each log entry
	exitFunc         func(int)                               // Function to call on fatal/panic
	fields           map[string][]byte                       // Default fields to include in all logs as []byte for zero allocation
	sampler          *sampler.LogSampler                    // Sampler for log sampling
	buffer           *writer.Buffered                       // Buffered writer for performance
	rotation         *writer.Rotator                         // Rotating file writer for log rotation
	contextExtractor func(context.Context) map[string][]byte // Function to extract fields from context
	metrics          metric.Collector                        // Metrics collector to use
	onFatal          func(*core.LogEntry)                    // Function to call when a fatal log occurs
	onPanic          func(*core.LogEntry)                    // Function to call when a panic log occurs
	stats            *LoggerStats                            // Statistics for logger
	asyncLogger      *writer.AsyncLogger                      // Async logger for non-blocking logging
	errorFileHook    *hook.FileHook                          // Built-in error file hook for ERROR+ levels
	closed           *atomic.Bool                            // Flag to indicate if logger is closed
	pid              int                                     // Process ID
	clock            *util.Clock                             // Clock for timestamp optimization
	entryPool        *sync.Pool                              // Pool for LogEntry objects
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
func NewDefault() *Logger {
	cfg := Config{
		Level:           core.INFO,
		Output:          os.Stdout,
		ErrorOutput:     os.Stderr,
		CallerDepth:     DEFAULT_CALLER_DEPTH,
		TimestampFormat: DEFAULT_TIMESTAMP_FORMAT,
		BufferSize:      DEFAULT_BUFFER_SIZE,
		FlushInterval:   DEFAULT_FLUSH_INTERVAL,
		ClockInterval:   10 * time.Millisecond,
		MaskValue:       "[MASKED]",
		Formatter: &formatter.TextFormatter{
			UseColors:        true,
			ShowTimestamp:    true,
			ShowCaller:       true,
			TimestampFormat:  DEFAULT_TIMESTAMP_FORMAT,
		},
	}
	return New(cfg)
}

// New creates a new logger with the given configuration
// Optimized for 1M+ logs/second with early filtering
func New(config Config) *Logger {
	validate(&config)

	l := &Logger{
		Config:           config,
		formatter:        config.Formatter,
		out:              config.Output,
		errOut:           config.ErrorOutput,
		errOutMu:         &sync.Mutex{},     // Initialize the mutex pointer
		mu:               new(sync.RWMutex), // Initialize the mutex pointer
		exitFunc:         config.ExitFunc,
		fields:           make(map[string][]byte),
		hooks:            config.Hooks, // Initialize hooks from config
		contextExtractor: config.ExtractContext,
		metrics:          config.Collector,
		onFatal:          config.OnFatal,
		onPanic:          config.OnPanic,
		stats:            NewLoggerStats(),
		closed:           &atomic.Bool{},
		pid:              os.Getpid(),
		entryPool: &sync.Pool{
			New: func() interface{} {
				return &core.LogEntry{}
			},
		},
	}

	if config.LogErrors {
		errorHook, err := hook.NewFileHook("errors.log")
		if err != nil {
			l.handleError(newErrorf("failed to create error file hook: %v", err))
		} else {
			l.errorFileHook = errorHook
			l.hooks = append(l.hooks, errorHook)
		}
	}

	if config.ClockInterval > 0 {
		l.clock = util.NewClock(config.ClockInterval)
	} else {
		l.clock = util.NewClock(10 * time.Millisecond) // Default clock interval
	}

	if l.exitFunc == nil {
		l.exitFunc = os.Exit
	}

	l.setupWriters()

	if config.EnableSampling && config.EnableSamplingRate > 1 {
		samplerWrapper := &samplerLoggerWrapper{logger: l}
		l.sampler = sampler.NewSamplingLogger(samplerWrapper, config.EnableSamplingRate)
	}

	if config.AsyncMode {
		l.asyncLogger = writer.NewAsyncLogger(l, config.WorkerCount, config.ChannelSize, config.ProcessTimeout, config.NoTimeout)
	}

	if config.AsyncMode {
		l.asyncLogger = writer.NewAsyncLogger(l, config.WorkerCount, config.ChannelSize, config.ProcessTimeout, config.NoTimeout)
	}

	return l
}

// samplerLoggerWrapper wraps Logger to provide old interface for sampler
type samplerLoggerWrapper struct {
	logger *Logger
}

func (w *samplerLoggerWrapper) Log(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte) {
	w.logger.LogLegacy(ctx, level, msg, fields)
}

func (l *Logger) setupWriters() {
	currentWriter := l.Config.Output

	if l.Config.EnableRotation && l.Config.RotationConfig != nil {
		if file, ok := currentWriter.(*os.File); ok {
			var err error
			l.rotation, err = writer.NewRotator(file.Name(), l.Config.RotationConfig)
			if err == nil {
				currentWriter = l.rotation
			} else {
				l.handleError(newErrorf("failed to setup rotation: %v", err))
			}
		}
	}

	if l.Config.BufSize > 0 {
		l.buffer = writer.NewBuffered(currentWriter, l.Config.BufSize, l.Config.Flush, l.handleError, l.Config.Batch, l.Config.BatchTime)
		currentWriter = l.buffer
	}

	l.out = currentWriter
}

// These methods are to satisfy interfaces for async/sampler writers
func (l *Logger) ErrorHandler() func(error) { return l.handleError }
func (l *Logger) ErrOut() io.Writer         { return l.errOut }
func (l *Logger) ErrOutMu() *sync.Mutex     { return l.errOutMu }

// internal logging method optimized for 1M+ logs/second with interface{} fields (for backward compatibility)
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

	// Convert interface{} fields to []byte fields for zero-allocation processing
	byteFields := make(map[string][]byte)
	for k, v := range fields {
		switch value := v.(type) {
		case string:
			byteFields[k] = core.StringToBytes(value)
		case []byte:
			byteFields[k] = value
		default:
			// For unsupported types, convert to string representation
			byteFields[k] = core.StringToBytes(fmt.Sprintf("%v", v))
		}
	}

	// Optimized path for non-blocking scenarios using atomic operations
	if l.asyncLogger != nil {
		// Use lock-free async logging for high throughput
		l.asyncLogger.Log(level, message, byteFields, ctx)
		return
	}

	// Hot path is efficient
	l.writeByte(ctx, level, message, byteFields)
}

// internal logging method optimized for 1M+ logs/second with []byte fields (zero-allocation)
// logZero handles zero-allocation logging with variadic key-value pairs
func (l *Logger) logZero(ctx context.Context, level core.Level, message []byte, keyvals ...[]byte) {
	if l.sampler != nil && !l.sampler.ShouldLog() {
		return
	}

	// For zero-allocation, we pass keyvals directly to formatter
	if l.asyncLogger != nil {
		l.asyncLogger.LogZero(level, message, ctx, keyvals...)
		return
	}

	l.writeZero(ctx, level, message, keyvals...)
}

// logBytes handles logging with map fields (legacy)
func (l *Logger) logBytes(ctx context.Context, level core.Level, message []byte, fields map[string][]byte) {
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

	// Hot path is efficient - no field conversion needed
	l.writeByte(ctx, level, message, fields)
}

// final write to output with zero-allocation optimizations for interface{} fields (backward compatibility)
//nolint:unused
func (l *Logger) write(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) {
	entry := l.buildEntry(ctx, level, message, fields)

	// Use efficient buffer for zero-allocation
	buf := util.GetBuffer()
	defer util.PutBuffer(buf)

	if err := l.formatter.Format(buf, entry); err != nil {
		l.handleError(err)
		core.PutEntryToPool(entry)
		return
	}

	bytesToWrite := buf.Bytes()

	// Optimized write with minimal locking - only lock when actually writing
	if l.Config.NoLocking {
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

// writeZero writes log entry with zero allocations using variadic key-value pairs
func (l *Logger) writeZero(ctx context.Context, level core.Level, message []byte, keyvals ...[]byte) {
	entry := l.entryPool.Get().(*core.LogEntry)
	defer l.entryPool.Put(entry)

	entry.Reset()
	entry.Level = level
	entry.Message = message
	entry.Timestamp = l.clock.Now()

	// Set keyvals directly without map allocation
	if len(keyvals)%2 == 0 {
		entry.KeyVals = keyvals
	} else {
		// Odd number of keyvals, ignore the last one
		entry.KeyVals = keyvals[:len(keyvals)-1]
	}

	if l.Config.ShowCaller {
		entry.ShowCaller = util.GetCaller(l.Config.ShowCallerDepth)
	}

	buf := util.GetBuffer()
	defer util.PutBuffer(buf)
	
	l.Config.Formatter.Format(buf, entry)
	l.mu.Lock()
	l.Config.Output.Write(buf.Bytes())
	l.mu.Unlock()
}

// final write to output with zero-allocation optimizations for []byte fields (true zero-allocation)
func (l *Logger) writeByte(ctx context.Context, level core.Level, message []byte, fields map[string][]byte) {
	entry := l.buildEntryByte(ctx, level, message, fields)

	// Use efficient buffer for zero-allocation
	buf := util.GetBuffer()
	defer util.PutBuffer(buf)

	if err := l.formatter.Format(buf, entry); err != nil {
		l.handleError(err)
		core.PutEntryToPool(entry)
		return
	}

	bytesToWrite := buf.Bytes()

	// Optimized write with minimal locking - only lock when actually writing
	if l.Config.NoLocking {
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
			tempBuf := util.GetSmallBuf()
			result := strconv.ShowAppendInt(tempBuf[:0], int64(v), 10)
			buf = append(buf, result...)
			util.PutSmallBuf(tempBuf)
		case int64:
			tempBuf := util.GetSmallBuf()
			result := strconv.ShowAppendInt(tempBuf[:0], v, 10)
			buf = append(buf, result...)
			util.PutSmallBuf(tempBuf)
		case float64:
			tempBuf := util.GetSmallBuf()
			result := strconv.ShowAppendFloat(tempBuf[:0], v, 'g', -1, 64)
			buf = append(buf, result...)
			util.PutSmallBuf(tempBuf)
		case bool:
			if v {
				buf = append(buf, "true"...)
			} else {
				buf = append(buf, "false"...)
			}
		default:
			// For other types, we fallback to manual conversion to avoid fmt
			// Use a temporary buffer to avoid multiple allocations
			tempBuf := util.GetBuffer()
			manualFormatValue(tempBuf, v)
			buf = append(buf, tempBuf.Bytes()...)
			util.PutBuffer(tempBuf)
		}
	}

	return buf
}

// formatfArgsToBytes formats variadic arguments with a format string into a byte slice with minimal allocations.
// This implementation aims for at most 1 allocation per call.
func (l *Logger) formatfArgsToBytes(format string, args ...interface{}) []byte {
	// We only allow one allocation per call: the final byte slice
	buf := util.GetBuffer()
	defer util.PutBuffer(buf)

	// Use manual formatting to avoid fmt dependency
	manualFormatWithArgs(buf, format, args...)

	// Final single allocation: copy the buffer content to a new byte slice
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}

// buildEntry creates a log entry with minimal allocations
//
//nolint:unused
func (l *Logger) buildEntry(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) *core.LogEntry {
	entry := core.GetEntryFromPool()

	// Use clock if available to avoid allocation
	if l.clock != nil {
		entry.Timestamp = l.clock.Now()
	} else {
		entry.Timestamp = time.Now()
	}

	entry.Level = level
	entry.LevelName = level.ToBytes()
	entry.Message = message
	entry.ShowPID = l.pid

	// Copy fields with minimal allocations - l.fields is now []byte
	for k, v := range l.fields {
		entry.Fields[k] = v // v is already []byte
	}
	for k, v := range fields {
		// Convert interface{} values to []byte when copying to entry.Fields
		switch val := v.(type) {
		case string:
			entry.Fields[k] = core.StringToBytes(val)
		case []byte:
			entry.Fields[k] = val
		default:
			entry.Fields[k] = core.StringToBytes(fmt.Sprintf("%v", val))
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
			case "trace_id":
				entry.ShowTraceID = core.StringToBytes(v)
			case "span_id":
				entry.SpanID = core.StringToBytes(v)
			case "user_id":
				entry.UserID = core.StringToBytes(v)
			case "session_id":
				entry.SessionID = core.StringToBytes(v)
			case "request_id":
				entry.RequestID = core.StringToBytes(v)
			}
		}
		util.PutMapStr(contextData)
	}

	// Caller info only if required to avoid overhead
	if l.Config.ShowCaller {
		entry.ShowCaller = util.GetCaller(l.Config.ShowCallerDepth)
	}

	// Stack trace only for ERROR level and above
	if l.Config.IncludeStackTrace && level >= core.ERROR {
		stackTraceBytes, stackTraceBufPtr := util.GetStackTrace(l.Config.StackDepth)
		entry.IncludeStackTrace = stackTraceBytes
		entry.IncludeStackTraceBufPtr = stackTraceBufPtr
	}

	return entry
}

// buildEntryByte creates a log entry with minimal allocations using []byte fields (true zero-allocation)
func (l *Logger) buildEntryByte(ctx context.Context, level core.Level, message []byte, fields map[string][]byte) *core.LogEntry {
	entry := core.GetEntryFromPool()

	// Use clock if available to avoid allocation
	if l.clock != nil {
		entry.Timestamp = l.clock.Now()
	} else {
		entry.Timestamp = time.Now()
	}

	entry.Level = level
	entry.LevelName = level.ToBytes()
	entry.Message = message
	entry.ShowPID = l.pid

	// Copy fields with minimal allocations - these are already []byte
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	for k, v := range fields {
		entry.Fields[k] = v
	}

	// Extract context with zero allocation if possible
	if l.contextExtractor != nil {
		contextFields := l.contextExtractor(ctx)
		for k, v := range contextFields {
			entry.Fields[k] = v // v is already []byte
		}
	} else if ctx != nil {
		contextData := util.ExtractFromContext(ctx)
		for k, v := range contextData {
			switch k {
			case "trace_id":
				entry.ShowTraceID = core.StringToBytes(v)
			case "span_id":
				entry.SpanID = core.StringToBytes(v)
			case "user_id":
				entry.UserID = core.StringToBytes(v)
			case "session_id":
				entry.SessionID = core.StringToBytes(v)
			case "request_id":
				entry.RequestID = core.StringToBytes(v)
			}
		}
		util.PutMapStr(contextData)
	}

	// Caller info only if required to avoid overhead
	if l.Config.ShowCaller {
		entry.ShowCaller = util.GetCaller(l.Config.ShowCallerDepth)
	}

	// Stack trace only for ERROR level and above
	if l.Config.IncludeStackTrace && level >= core.ERROR {
		stackTraceBytes, stackTraceBufPtr := util.GetStackTrace(l.Config.StackDepth)
		entry.IncludeStackTrace = stackTraceBytes
		entry.IncludeStackTraceBufPtr = stackTraceBufPtr
	}

	return entry
}

// runHooks executes hooks with minimal lock contention
func (l *Logger) runHooks(entry *core.LogEntry) {
	// Use RLock for read-only access
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Early return if no hooks
	if len(l.hooks) == 0 {
		return
	}

	// Execute hooks with graceful error handling
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
		// NOT: panic(entry.Message) // Which would cause application crash
		// Handle buffer full with graceful degradation
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
func (l *Logger) handleError(err error) {
	// Graceful error handling tanpa panic
	if l.Config.ErrorHandler != nil {
		l.Config.ErrorHandler(err)
	} else {
		// Ignore errors gracefully, jangan crash aplikasi
		l.errOutMu.Lock()
		defer l.errOutMu.Unlock()

		// Use manual formatting with zero-allocation approach
		buf := util.GetBuffer()
		defer util.PutBuffer(buf)
		buf.Write([]byte("logger error: "))
		buf.Write(core.StringToBytes(err.Error())) // Use zero-allocation string to byte conversion
		buf.Write([]byte("\n"))
		_, _ = l.errOut.Write(buf.Bytes())
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
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	if len(fields) == 0 {
		return l
	}
	newLogger := l.clone()
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			newLogger.fields[k] = []byte(val)
		case []byte:
			newLogger.fields[k] = val
		case int:
			newLogger.fields[k] = I2B(val)
		case float64:
			newLogger.fields[k] = F2B(val)
		case bool:
			newLogger.fields[k] = B2B(val)
		default:
			newLogger.fields[k] = []byte(fmt.Sprintf("%v", val))
		}
	}
	return newLogger
}

// clone creates a copy of the logger with shared resources
func (l *Logger) clone() *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Create a new logger with shared resources to avoid copying noCopy fields.
	// We want clone to share the same writers (async, buffer)
	// and their goroutines, not create new ones.
	cloned := &Logger{
		Config:           l.Config,
		formatter:        l.formatter,
		out:              l.out,
		errOut:           l.errOut,
		errOutMu:         l.errOutMu,
		mu:               l.mu,
		hooks:            make([]hook.Hook, len(l.hooks)),
		exitFunc:         l.exitFunc,
		fields:           make(map[string][]byte, len(l.fields)+10),
		sampler:          l.sampler,
		buffer:           l.buffer,
		rotation:         l.rotation,
		contextExtractor: l.contextExtractor,
		metrics:          l.metrics,
		onFatal:          l.onFatal,
		onPanic:          l.onPanic,
		stats:            l.stats,
		asyncLogger:      l.asyncLogger,
		errorFileHook:    l.errorFileHook,
		closed:           &atomic.Bool{},
		pid:              l.pid,
		clock:            l.clock,
	}
	copy(cloned.hooks, l.hooks)

	// Copy parent fields.
	for k, v := range l.fields {
		cloned.fields[k] = v
	}

	return cloned
}

// --- Unified Public API ---

// Log logs with level and message
// Zero-allocation logging API using variadic parameters
// LogZ logs with zero allocations using key-value pairs
func (l *Logger) LogZ(ctx context.Context, level core.Level, msg []byte, keyvals ...[]byte) {
	if level < l.Config.Level {
		return
	}
	l.logZero(ctx, level, msg, keyvals...)
}

// LogZC logs with context only (zero allocation)
func (l *Logger) LogZC(ctx context.Context, level core.Level, msg []byte) {
	if level < l.Config.Level {
		return
	}
	l.logZero(ctx, level, msg)
}

// Legacy API - Log with map (may allocate)
// LogLegacy for sampler compatibility (legacy map interface)
func (l *Logger) LogLegacy(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte) {
	if level < l.Config.Level {
		return
	}
	l.logBytes(ctx, level, msg, fields)
}

// Log using variadic key-value pairs
func (l *Logger) Log(ctx context.Context, level core.Level, msg []byte, keyvals ...[]byte) {
	if level < l.Config.Level {
		return
	}
	l.logZero(ctx, level, msg, keyvals...)
}

// LogC with context
func (l *Logger) LogC(ctx context.Context, level core.Level, msg []byte) {
	if level < l.Config.Level {
		return
	}
	l.log(ctx, level, msg, nil)
}

// LogCF with context and fields (complete API)
func (l *Logger) LogCF(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte) {
	if level < l.Config.Level {
		return
	}
	l.logBytes(ctx, level, msg, fields)
}

// Level constants for convenience
const (
	TRACE = core.TRACE
	DEBUG = core.DEBUG
	INFO  = core.INFO
	WARN  = core.WARN
	ERROR = core.ERROR
	FATAL = core.FATAL
)

// Optimized formatted logging methods
func (l *Logger) Tracef(format string, args ...interface{}) {
	if core.TRACE < l.Config.Level {
		return
	}
	l.log(context.Background(), core.TRACE, l.formatfArgsToBytes(format, args...), nil)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if core.DEBUG < l.Config.Level {
		return
	}
	l.log(context.Background(), core.DEBUG, l.formatfArgsToBytes(format, args...), nil)
}

// Infof logs a formatted message with INFO level
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(context.Background(), core.INFO, l.formatfArgsToBytes(format, args...), nil)
}

// Noticef logs a formatted message with NOTICE level
func (l *Logger) Noticef(format string, args ...interface{}) {
	l.log(context.Background(), core.NOTICE, l.formatfArgsToBytes(format, args...), nil)
}

// Warnf logs a formatted message with WARN level
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(context.Background(), core.WARN, l.formatfArgsToBytes(format, args...), nil)
}

// Errorf logs a formatted message with ERROR level
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(context.Background(), core.ERROR, l.formatfArgsToBytes(format, args...), nil)
}

// Fatalf logs a formatted message with FATAL level and exits the application
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(context.Background(), core.FATAL, l.formatfArgsToBytes(format, args...), nil)
}

// Panicf logs a formatted message with PANIC level and panics
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.log(context.Background(), core.PANIC, l.formatfArgsToBytes(format, args...), nil)
}

// Efficient helper functions for common conversions
func IntB(i int) []byte {
	return []byte(strconv.Itoa(i))
}

func FloatB(f float64) []byte {
	return []byte(strconv.FormatFloat(f, 'g', -1, 64))
}

func BoolB(b bool) []byte {
	if b {
		return []byte("true")
	}
	return []byte("false")
}

// Context-aware logging methods with []byte (zero allocation)

// TraceCB logs a message with TRACE level and extracts context information using []byte (zero-allocation)
func (l *Logger) TraceCB(ctx context.Context, message []byte) {
	if core.TRACE >= l.Config.Level {
		l.log(ctx, core.TRACE, message, nil)
	}
}

// DebugCB logs a message with DEBUG level and extracts context information using []byte (zero-allocation)
func (l *Logger) DebugCB(ctx context.Context, message []byte) {
	if core.DEBUG >= l.Config.Level {
		l.log(ctx, core.DEBUG, message, nil)
	}
}

// InfoCB logs a message with INFO level and extracts context information using []byte (zero-allocation)
func (l *Logger) InfoCB(ctx context.Context, message []byte) { 
	if core.INFO >= l.Config.Level {
		l.log(ctx, core.INFO, message, nil) 
	}
}

// WarnCB logs a message with WARN level and extracts context information using []byte (zero-allocation)
func (l *Logger) WarnCB(ctx context.Context, message []byte) { 
	if core.WARN >= l.Config.Level {
		l.log(ctx, core.WARN, message, nil) 
	}
}

// ErrorCB logs a message with ERROR level and extracts context information using []byte (zero-allocation)
func (l *Logger) ErrorCB(ctx context.Context, message []byte) {
	if core.ERROR >= l.Config.Level {
		l.log(ctx, core.ERROR, message, nil)
	}
}

// FatalCB logs a message with FATAL level and extracts context information, then exits the application using []byte (zero-allocation)
func (l *Logger) FatalCB(ctx context.Context, message []byte) {
	l.log(ctx, core.FATAL, message, nil)
}

// Context-aware logging methods (interface{} args)
func (l *Logger) TraceC(ctx context.Context, args ...interface{}) {
	if core.TRACE >= l.Config.Level {
		l.log(ctx, core.TRACE, l.formatArgsToBytes(args...), nil)
	}
}

func (l *Logger) DebugC(ctx context.Context, args ...interface{}) {
	if core.DEBUG >= l.Config.Level {
		l.log(ctx, core.DEBUG, l.formatArgsToBytes(args...), nil)
	}
}

func (l *Logger) InfoC(ctx context.Context, args ...interface{}) {
	if core.INFO >= l.Config.Level {
		l.log(ctx, core.INFO, l.formatArgsToBytes(args...), nil)
	}
}

func (l *Logger) WarnC(ctx context.Context, args ...interface{}) {
	if core.WARN >= l.Config.Level {
		l.log(ctx, core.WARN, l.formatArgsToBytes(args...), nil)
	}
}

func (l *Logger) ErrorC(ctx context.Context, args ...interface{}) {
	if core.ERROR >= l.Config.Level {
		l.log(ctx, core.ERROR, l.formatArgsToBytes(args...), nil)
	}
}

// Utility functions for []byte conversion
func S2B(s string) []byte { return []byte(s) }
func I2B(i int) []byte { return []byte(strconv.Itoa(i)) }
func F2B(f float64) []byte { return []byte(strconv.FormatFloat(f, 'g', -1, 64)) }
func B2B(b bool) []byte { 
	if b { return []byte("true") }
	return []byte("false")
}

// Legacy API for backward compatibility
func (l *Logger) Trace(args ...interface{}) {
	l.LogCF(context.Background(), TRACE, l.formatArgsToBytes(args...), nil)
}

func (l *Logger) Info(args ...interface{}) {
	l.LogCF(context.Background(), INFO, l.formatArgsToBytes(args...), nil)
}

func (l *Logger) Debug(args ...interface{}) {
	l.LogCF(context.Background(), DEBUG, l.formatArgsToBytes(args...), nil)
}

func (l *Logger) Warn(args ...interface{}) {
	l.LogCF(context.Background(), WARN, l.formatArgsToBytes(args...), nil)
}

func (l *Logger) Error(args ...interface{}) {
	l.LogCF(context.Background(), ERROR, l.formatArgsToBytes(args...), nil)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.LogCF(context.Background(), FATAL, l.formatArgsToBytes(args...), nil)
}

// Context-aware legacy methods
func (l *Logger) InfofC(ctx context.Context, format string, args ...interface{}) {
	l.LogCF(ctx, INFO, l.formatfArgsToBytes(format, args...), nil)
}

// Helper functions
func newErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func manualFormatValue(buf *bytes.Buffer, v interface{}) {
	fmt.Fprintf(buf, "%v", v)
}

func manualFormatWithArgs(buf *bytes.Buffer, format string, args ...interface{}) {
	fmt.Fprintf(buf, format, args...)
}


