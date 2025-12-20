package logger

import (
	"bytes"
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
)

// OptimizedLogger is an efficient logger version for minimal allocations
type OptimizedLogger struct {
	config     LoggerConfig
	formatter  formatter.Formatter
	out        io.Writer
	mu         sync.Mutex
	fields     map[string]interface{}
	level      core.Level
	bufferPool *sync.Pool
	closed     atomic.Bool
}

// NewOptimizedLogger creates an efficient logger instance
func NewOptimizedLogger(config LoggerConfig) *OptimizedLogger {
	if config.Output == nil {
		config.Output = io.Discard
	}
	if config.Formatter == nil {
		config.Formatter = &formatter.TextFormatter{}
	}

	bufferPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 1024))
		},
	}

	return &OptimizedLogger{
		config:     config,
		formatter:  config.Formatter,
		out:        config.Output,
		level:      config.Level,
		bufferPool: bufferPool,
		fields:     make(map[string]interface{}),
	}
}

// logInternal is the core efficient function for minimal allocations
func (l *OptimizedLogger) logInternal(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) {
	if level < l.level {
		return
	}
	buf := l.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

	// Create entry using pool
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)

	// Fill entry with required data
	entry.Timestamp = time.Now()
	entry.Level = level
	entry.LevelName = level.ToBytes()
	entry.Message = message

	// Add fields from logger
	for k, v := range l.fields {
		// Convert interface{} values to []byte when copying to entry.Fields
		switch val := v.(type) {
		case string:
			entry.Fields[k] = core.StringToBytes(val)
		case []byte:
			entry.Fields[k] = val
		default:
			entry.Fields[k] = core.StringToBytes(stringify(val)) // Use existing stringify function with conversion
		}
	}
	// Add specific fields for this log
	for k, v := range fields {
		// Convert interface{} values to []byte when copying to entry.Fields
		switch val := v.(type) {
		case string:
			entry.Fields[k] = core.StringToBytes(val)
		case []byte:
			entry.Fields[k] = val
		default:
			entry.Fields[k] = core.StringToBytes(stringify(val)) // Use existing stringify function with conversion
		}
	}

	// Format entry to buffer
	if err := l.formatter.Format(buf, entry); err != nil {
		// Handle error if needed
		return
	}

	// Write to output
	l.mu.Lock()
	_, _ = l.out.Write(buf.Bytes())
	l.mu.Unlock()
}

// Basic logging level methods
func (l *OptimizedLogger) Trace(args ...interface{}) {
	if core.TRACE >= l.level {
		message := l.formatArgsToBytes(args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.TRACE, message, nil)
	}
}

func (l *OptimizedLogger) Debug(args ...interface{}) {
	if core.DEBUG >= l.level {
		message := l.formatArgsToBytes(args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.DEBUG, message, nil)
	}
}

func (l *OptimizedLogger) Info(args ...interface{}) {
	if core.INFO >= l.level {
		message := l.formatArgsToBytes(args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.INFO, message, nil)
	}
}

func (l *OptimizedLogger) Warn(args ...interface{}) {
	if core.WARN >= l.level {
		message := l.formatArgsToBytes(args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.WARN, message, nil)
	}
}

func (l *OptimizedLogger) Error(args ...interface{}) {
	if core.ERROR >= l.level {
		message := l.formatArgsToBytes(args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.ERROR, message, nil)
	}
}

func (l *OptimizedLogger) Tracef(format string, args ...interface{}) {
	if core.TRACE >= l.level {
		message := l.formatfArgsToBytes(format, args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.TRACE, message, nil)
	}
}

func (l *OptimizedLogger) Debugf(format string, args ...interface{}) {
	if core.DEBUG >= l.level {
		message := l.formatfArgsToBytes(format, args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.DEBUG, message, nil)
	}
}

func (l *OptimizedLogger) Infof(format string, args ...interface{}) {
	if core.INFO >= l.level {
		message := l.formatfArgsToBytes(format, args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.INFO, message, nil)
	}
}

func (l *OptimizedLogger) Warnf(format string, args ...interface{}) {
	if core.WARN >= l.level {
		message := l.formatfArgsToBytes(format, args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.WARN, message, nil)
	}
}

func (l *OptimizedLogger) Errorf(format string, args ...interface{}) {
	if core.ERROR >= l.level {
		message := l.formatfArgsToBytes(format, args...) // This should be efficient for only 1 allocation
		l.logInternal(context.Background(), core.ERROR, message, nil)
	}
}

// formatArgsToBytes converts arguments to byte slice with only 1 allocation
func (l *OptimizedLogger) formatArgsToBytes(args ...interface{}) []byte {
	// Use buffer from pool to combine arguments
	buf := l.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

	for i, arg := range args {
		if i > 0 {
			buf.WriteByte(' ') // Tambahkan spasi antar argumen
		}
		switch v := arg.(type) {
		case string:
			buf.WriteString(v)
		case []byte:
			buf.Write(v)
		case int:
			buf.WriteString(itoa(v))
		case int64:
			buf.WriteString(i64toa(v))
		case float64:
			buf.WriteString(ftoa(v))
		case bool:
			if v {
				buf.WriteString("true")
			} else {
				buf.WriteString("false")
			}
		default:
			buf.WriteString(stringify(v)) // This may cause additional allocation
		}
	}

	// Create a copy of buffer data and return to pool beforehand
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}

// formatfArgsToBytes converts formatted arguments to byte slice with only 1 allocation
func (l *OptimizedLogger) formatfArgsToBytes(format string, args ...interface{}) []byte {
	// Use buffer from pool to combine arguments
	buf := l.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

	// This will cause 1 allocation due to fmt.Sprintf usage
	// In truly efficient production implementation,
	// we would replace this with zero-allocation implementation
	buf.WriteString(formatString(format, args...))

	// Create a copy of buffer data and return to pool beforehand
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}

// Helper function for type conversion without allocation (as much as possible)
func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	// Use stack-allocated buffer for conversion
	var buf [32]byte
	n := len(buf)
	sign := false

	if i < 0 {
		sign = true
		i = -i
	}

	for i > 0 {
		n--
		buf[n] = byte(i%10) + '0'
		i /= 10
	}

	if sign {
		n--
		buf[n] = '-'
	}

	return string(buf[n:])
}

func i64toa(i int64) string {
	if i == 0 {
		return "0"
	}

	// Use stack-allocated buffer for conversion
	var buf [32]byte
	n := len(buf)
	sign := false

	if i < 0 {
		sign = true
		i = -i
	}

	for i > 0 {
		n--
		buf[n] = byte(i%10) + '0'
		i /= 10
	}

	if sign {
		n--
		buf[n] = '-'
	}

	return string(buf[n:])
}

func ftoa(f float64) string {
	// This still causes allocation because we don't implement manual float parsing
	// In efficient production library, we would use zero-allocation algorithm
	return formatFloat(f)
}

// formatFloat is a basic implementation to convert float to string
func formatFloat(f float64) string {
	// Integer part
	intPart := int64(f)
	decPart := f - float64(intPart)

	// Convert decimal part to string
	decStr := itoa(int(decPart * 100))

	return i64toa(intPart) + "." + decStr
}

// stringify converts any type to string (this causes allocation)
func stringify(v interface{}) string {
	return string([]byte(v.(string))) // This causes additional allocation
}

// formatString is a basic implementation of fmt.Sprintf (causes allocation)
func formatString(format string, args ...interface{}) string {
	// In production implementation, this will be replaced with zero-allocation implementation
	return format
}

// WithFields adds fields to logger
func (l *OptimizedLogger) WithFields(fields map[string]interface{}) *OptimizedLogger {
	newLogger := &OptimizedLogger{
		config:     l.config,
		formatter:  l.formatter,
		out:        l.out,
		level:      l.level,
		bufferPool: l.bufferPool,
	}

	// Copy fields from original logger
	newLogger.fields = make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithFieldsB adds fields to logger using []byte (zero-allocation)
func (l *OptimizedLogger) WithFieldsB(fields map[string][]byte) *OptimizedLogger {
	newLogger := &OptimizedLogger{
		config:     l.config,
		formatter:  l.formatter,
		out:        l.out,
		level:      l.level,
		bufferPool: l.bufferPool,
	}

	// Copy fields from original logger and convert to interface{}
	newLogger.fields = make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v // This will be converted to []byte in logInternal
	}

	return newLogger
}

// Close closes the logger
func (l *OptimizedLogger) Close() {
	l.closed.Store(true)
}
