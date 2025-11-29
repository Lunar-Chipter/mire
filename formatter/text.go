package formatter

import (
	"bytes"
	"strconv"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/util"
)

// TextFormatter formats log entries in a human-readable text format
type TextFormatter struct {
	EnableColors        bool                                       // Enable ANSI colors in output
	ShowTimestamp       bool                                       // Show timestamp in output
	ShowCaller          bool                                       // Show caller information
	ShowGoroutine       bool                                       // Show goroutine ID
	ShowPID             bool                                       // Show process ID
	ShowTraceInfo       bool                                       // Show trace information
	ShowHostname        bool                                       // Show hostname
	ShowApplication     bool                                       // Show application name
	FullTimestamp       bool                                       // Show full timestamp with nanoseconds
	TimestampFormat     string                                     // Custom timestamp format
	IndentFields        bool                                       // Indent fields for better readability
	MaxFieldWidth       int                                        // Maximum width for field values
	EnableStackTrace    bool                                       // Enable stack trace for errors
	StackTraceDepth     int                                        // Maximum stack trace depth
	EnableDuration      bool                                       // Show operation duration
	CustomFieldOrder    []string                                   // Custom order for fields
	EnableColorsByLevel bool                                       // Enable colors based on log level
	FieldTransformers   map[string]func(interface{}) string        // Functions to transform field values
	SensitiveFields     []string                                   // List of sensitive field names
	MaskSensitiveData   bool                                       // Whether to mask sensitive data
	MaskStringValue     string                                     // String value to use for masking
	MaskStringBytes     []byte                                     // Byte slice for masking (zero-allocation)
	DisableHTMLEscape   bool                                       // Disable HTML escaping in text
}

var ResetColorBytes = []byte("\033[0m")
var metaColorBytes = []byte("\033[38;5;245m")      // Gray for meta info
var callerColorBytes = []byte("\033[38;5;246m")    // Gray for caller info
var durationColorBytes = []byte("\033[38;5;155m")  // Light green for duration
var traceColorBytes = []byte("\033[38;5;141m")     // Purple for trace info
var errorColorBytes = []byte("\033[38;5;196m")     // Bright red for errors
var stackTraceColorBytes = []byte("\033[38;5;240m") // Dark gray for stack trace
var fieldsWrapperColorBytes = []byte("\033[38;5;243m") // Light gray for field wrappers
var fieldKeyColorBytes = []byte("\033[38;5;228m")  // Light yellow for field keys
var fieldValueColorBytes = []byte("\033[38;5;159m")// Light cyan for field values
var tagsColorBytes = []byte("\033[38;5;135m")      // Purple for tags
var metricsColorBytes = []byte("\033[38;5;85m")    // Green for metrics

// NewTextFormatter creates a new TextFormatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		MaskStringValue: "[MASKED]",
		FieldTransformers: make(map[string]func(interface{}) string),
		SensitiveFields: make([]string, 0),
	}
}

// Format formats a log entry into a byte slice
func (f *TextFormatter) Format(buf *bytes.Buffer, entry *core.LogEntry) error {
	// Pre-calculate required buffer space to minimize reallocations
	estimatedSize := 64 // Base size for timestamp, level, spacing
	estimatedSize += len(entry.Message)
	if f.ShowTimestamp {
		estimatedSize += len(f.TimestampFormat) + 10 // Extra for formatting overhead
	}
	estimatedSize += len(entry.Level.Bytes())
	if len(entry.Fields) > 0 {
		estimatedSize += 64 // Estimate for fields formatting
	}

	// Pre-grow buffer if possible to reduce reallocations
	currentCap := buf.Cap()
	if estimatedSize > currentCap {
		buf.Grow(estimatedSize - currentCap)
	}

	// Write timestamp - using more efficient manual formatting
	if f.ShowTimestamp {
		buf.WriteByte('[')
		// Use manual timestamp formatting to avoid allocation
		manualFormatTimestamp(buf, entry.Timestamp, f.TimestampFormat)
		buf.WriteByte(']')
		buf.WriteByte(' ')
	}

	// Write level with background and padding - manual byte manipulation
	levelBytes := entry.Level.Bytes()
	if f.EnableColors {
		buf.Write(core.LevelBackgroundBytes[entry.Level])
		buf.Write(core.LevelColorBytes[entry.Level])
		buf.WriteByte(' ')
		buf.Write(levelBytes)
		buf.WriteByte(' ')
		buf.Write(ResetColorBytes)
	} else {
		buf.WriteByte('[')
		buf.Write(levelBytes)
		buf.WriteByte(']')
	}
	buf.WriteByte(' ')

	// Write other metadata
	f.writeMeta(buf, entry)

	// Write message - already using []byte which is efficient
	if f.EnableColors && f.EnableColorsByLevel {
		buf.Write(core.LevelColorBytes[entry.Level])
		if entry.Level >= core.ERROR {
			buf.Write([]byte("\033[1m")) // Bold for important messages
		}
	}
	buf.Write(entry.Message) // Message is []byte, efficient
	if f.EnableColors && f.EnableColorsByLevel {
		buf.Write(ResetColorBytes)
	}

	// Write error, fields, tags, metrics, and stack trace
	f.writePostMessage(buf, entry)

	buf.WriteByte('\n')

	return nil
}

func (f *TextFormatter) writeMeta(buf *bytes.Buffer, entry *core.LogEntry) {
	if f.ShowHostname && entry.Hostname != "" {
		f.writeMetaPart(buf, entry.Hostname)
	}
	if f.ShowApplication && entry.Application != "" {
		f.writeMetaPart(buf, entry.Application)
	}
	if f.ShowPID {
		// Use pooled byte slice for AppendInt
		pidBuf := util.GetSmallByteSliceFromPool()
		defer util.PutSmallByteSliceToPool(pidBuf)
		pidBytes := strconv.AppendInt(pidBuf[:0], int64(entry.PID), 10)

		if f.EnableColors {
			buf.Write(metaColorBytes)
		}
		buf.Write([]byte("PID:"))
		buf.Write(pidBytes)
		if f.EnableColors {
			buf.Write(ResetColorBytes)
		}
		buf.WriteByte(' ')
	}
	if f.ShowGoroutine && entry.GoroutineID != "" {
		// Create GID string more efficiently using a temporary buffer
		gidPrefix := []byte("GID:")
		buf.Write(gidPrefix)
		buf.Write(core.StringToBytes(entry.GoroutineID))
		buf.WriteByte(' ')
	}
	if f.ShowTraceInfo {
		f.writeTraceInfo(buf, entry)
	}
	if f.ShowCaller && entry.Caller != nil {
		if f.EnableColors {
			buf.Write(callerColorBytes)
		}
		buf.Write(core.StringToBytes(entry.Caller.File))
		buf.WriteByte(':')

		// Use pooled byte slice for AppendInt
		lineBuf := util.GetSmallByteSliceFromPool()
		defer util.PutSmallByteSliceToPool(lineBuf)
		lineBytes := strconv.AppendInt(lineBuf[:0], int64(entry.Caller.Line), 10)
		buf.Write(lineBytes)

		if f.EnableColors {
			buf.Write(ResetColorBytes)
		}
		buf.WriteByte(' ')
	}
	if f.EnableDuration && entry.Duration > 0 {
		// Use pooled byte slice for AppendInt
		durationBuf := util.GetSmallByteSliceFromPool()
		defer util.PutSmallByteSliceToPool(durationBuf)
		durationBytes := strconv.AppendInt(durationBuf[:0], entry.Duration.Milliseconds(), 10)

		if f.EnableColors {
			buf.Write(durationColorBytes)
		}
		buf.WriteByte('(')
		buf.Write(durationBytes)
		buf.Write([]byte("ms)"))
		if f.EnableColors {
			buf.Write(ResetColorBytes)
		}
		buf.WriteByte(' ')
	}
}

func (f *TextFormatter) writeMetaPart(buf *bytes.Buffer, part string) {
	if f.EnableColors {
		buf.Write(metaColorBytes)
	}
	buf.Write(core.StringToBytes(part)) // Use zero-allocation string to byte conversion
	if f.EnableColors {
		buf.Write(ResetColorBytes)
	}
	buf.WriteByte(' ')
}

func (f *TextFormatter) writeTraceInfo(buf *bytes.Buffer, entry *core.LogEntry) {
	if entry.TraceID != "" {
		f.writeTracePart(buf, []byte("TRACE"), shortenID(entry.TraceID))
	}
	if entry.SpanID != "" {
		f.writeTracePart(buf, []byte("SPAN"), shortenID(entry.SpanID))
	}
	if entry.RequestID != "" {
		f.writeTracePart(buf, []byte("REQ"), shortenID(entry.RequestID))
	}
}

func (f *TextFormatter) writeTracePart(buf *bytes.Buffer, key []byte, value string) {
	if f.EnableColors {
		buf.Write(traceColorBytes)
	}
	buf.Write(key)
	buf.WriteByte(':')
	buf.Write(core.StringToBytes(value)) // Use zero-allocation string to byte conversion
	if f.EnableColors {
		buf.Write(ResetColorBytes)
	}
	buf.WriteByte(' ')
}

func (f *TextFormatter) writePostMessage(buf *bytes.Buffer, entry *core.LogEntry) {
	if entry.Error != nil {
		buf.WriteByte(' ')
		if f.EnableColors {
			buf.Write(errorColorBytes)
		}
		buf.Write([]byte("error="))
		if appender, ok := entry.Error.(core.ErrorAppender); ok {
			appender.AppendError(buf) // Use zero-allocation append
		} else {
			buf.WriteString(entry.Error.Error()) // Fallback to standard Error() which allocates
		}
		// Removed the '"' as it was inconsistent
		if f.EnableColors {
			buf.Write(ResetColorBytes)
		}
	}

	if len(entry.Fields) > 0 {
		buf.WriteByte(' ')
		f.formatFields(buf, entry.Fields)
	}
	if len(entry.Tags) > 0 {
		buf.WriteByte(' ')
		f.formatTags(buf, entry.Tags)
	}
	if len(entry.CustomMetrics) > 0 {
		buf.WriteByte(' ')
		f.formatMetrics(buf, entry.CustomMetrics)
	}
	if f.EnableStackTrace && len(entry.StackTrace) > 0 { // Check len() for []byte
		buf.WriteByte('\n')
		if f.EnableColors {
			buf.Write(stackTraceColorBytes)
		}
		buf.Write(entry.StackTrace) // Now []byte
		if f.EnableColors {
			buf.Write(ResetColorBytes)
		}
	}
}

func (f *TextFormatter) formatFields(buf *bytes.Buffer, fields map[string]interface{}) {
	if f.EnableColors {
		buf.Write(fieldsWrapperColorBytes)
	}
	buf.WriteByte('{')

	// Determine field order: either custom order or map order
	var orderedKeys []string
	if len(f.CustomFieldOrder) > 0 {
		// Use custom order for fields that exist in the entry
		// This avoids string slice allocation if custom order is not used
		orderedKeys = make([]string, 0, len(fields))
		// Add fields in custom order
		for _, field := range f.CustomFieldOrder {
			if _, exists := fields[field]; exists {
				orderedKeys = append(orderedKeys, field)
			}
		}

		// Add any remaining fields that weren't in the custom order
		for field := range fields {
			existsInCustom := false
			for _, customField := range orderedKeys {
				if field == customField {
					existsInCustom = true
					break
				}
			}
			if !existsInCustom {
				orderedKeys = append(orderedKeys, field)
			}
		}
	} else {
		// Use natural map order - avoid extra slice allocation where possible
		orderedKeys = make([]string, 0, len(fields))
		for k := range fields {
			orderedKeys = append(orderedKeys, k)
		}
	}

	for i, k := range orderedKeys {
		v := fields[k]
		if i > 0 {
			buf.WriteByte(' ')
		}

		if f.EnableColors {
			buf.Write(fieldKeyColorBytes)
		}
		// Use manual byte writing for key to avoid allocation
		buf.Write(core.StringToBytes(k))
		buf.WriteByte('=')
		if f.EnableColors {
			buf.Write(fieldValueColorBytes)
		}

		// Apply field transformer or mask sensitive data
		if f.MaskSensitiveData && f.isSensitiveField(k) {
			// Use byte slice for mask value to avoid string allocation
			buf.Write(core.StringToBytes(f.MaskStringValue)) // Use string to byte conversion
		} else if transformer, exists := f.FieldTransformers[k]; exists {
			// Apply field transformer
			transformedValue := transformer(v)
			buf.Write(core.StringToBytes(transformedValue)) // Convert string to byte slice without allocation
		} else {
			// Use optimized FormatValue that minimizes allocations
			util.FormatValue(buf, v, f.MaxFieldWidth)
		}
	}

	if f.EnableColors {
		buf.Write(fieldsWrapperColorBytes)
	}
	buf.WriteByte('}')
	if f.EnableColors {
		buf.Write(ResetColorBytes)
	}
}

func (f *TextFormatter) formatTags(buf *bytes.Buffer, tags []string) {
	if f.EnableColors {
		buf.Write(tagsColorBytes)
	}
	buf.WriteByte('[')
	for i, tag := range tags {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.Write(core.StringToBytes(tag)) // Use zero-allocation string to byte conversion
	}
	buf.WriteByte(']')
	if f.EnableColors {
		buf.Write(ResetColorBytes)
	}
}

func (f *TextFormatter) formatMetrics(buf *bytes.Buffer, metrics map[string]float64) {
	if f.EnableColors {
		buf.Write(metricsColorBytes)
	}
	buf.WriteByte('<')
	first := true
	for k, v := range metrics {
		if !first {
			buf.WriteByte(' ')
		}
		first = false
		buf.Write(core.StringToBytes(k)) // Use zero-allocation string to byte conversion
		buf.WriteByte('=')

		// Use pooled byte slice for AppendFloat
		floatBuf := util.GetSmallByteSliceFromPool()
		defer util.PutSmallByteSliceToPool(floatBuf)
		floatBytes := strconv.AppendFloat(floatBuf[:0], v, 'f', 2, 64) // Reset length before appending
		buf.Write(floatBytes)
	}
	buf.WriteByte('>')
	if f.EnableColors {
		buf.Write(ResetColorBytes)
	}
}

// --- Helper functions ---

// manualFormatTimestamp formats timestamp manually to avoid allocation
func manualFormatTimestamp(buf *bytes.Buffer, t time.Time, format string) {
	// Use the utility function that's optimized for timestamp formatting
	util.FormatTimestamp(buf, t, format)
}

func shortenID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// shortIDToBytes returns a byte slice with short ID without allocation
func shortIDToBytes(id string) []byte {
	if len(id) > 8 {
		return core.StringToBytes(id[:8])
	}
	return core.StringToBytes(id)
}

func (f *TextFormatter) isSensitiveField(field string) bool {
	for _, sensitiveField := range f.SensitiveFields {
		if field == sensitiveField {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

