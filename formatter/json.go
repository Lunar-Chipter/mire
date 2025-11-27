package formatter

import (
	"bytes"
	"strconv"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/util"
)

// JSONFormatter formats log entries in JSON format
type JSONFormatter struct {
	PrettyPrint         bool                                       // Enable pretty-printed JSON
	TimestampFormat     string                                     // Custom timestamp format
	ShowCaller          bool                                       // Show caller information
	ShowGoroutine       bool                                       // Show goroutine ID
	ShowPID             bool                                       // Show process ID
	ShowTraceInfo       bool                                       // Show trace information
	EnableStackTrace    bool                                       // Enable stack trace for errors
	EnableDuration      bool                                       // Show operation duration
	FieldKeyMap         map[string]string                          // Map for renaming fields
	DisableHTMLEscape   bool                                       // Disable HTML escaping in JSON
	SensitiveFields     []string                                   // List of sensitive field names
	MaskSensitiveData   bool                                       // Whether to mask sensitive data
	MaskStringValue     string                                     // String value to use for masking
	MaskStringBytes     []byte                                     // Byte slice for masking (zero-allocation)
	FieldTransformers   map[string]func(interface{}) interface{}   // Functions to transform field values
}

// NewJSONFormatter creates a new JSONFormatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format formats a log entry into JSON byte slice with zero allocations
func (f *JSONFormatter) Format(buf *bytes.Buffer, entry *core.LogEntry) error {
	if f.PrettyPrint {
		// For now, keep standard encoder for pretty print, but optimize regular print
		return f.formatWithStandardEncoder(buf, entry)
	}

	// Use manual JSON formatting for zero allocation
	return f.formatManually(buf, entry)
}

// formatManually creates JSON manually without allocations
func (f *JSONFormatter) formatManually(buf *bytes.Buffer, entry *core.LogEntry) error {
	buf.WriteByte('{')

	// Add timestamp - manually format to avoid allocation
	buf.Write([]byte("\"timestamp\":\""))
	util.FormatTimestamp(buf, entry.Timestamp, f.TimestampFormat)
	buf.Write([]byte("\","))

	// Add level
	buf.Write([]byte("\"level_name\":\""))
	buf.Write(entry.Level.Bytes()) // Using pre-allocated level bytes
	buf.Write([]byte("\","))

	// Add message
	buf.Write([]byte("\"message\":\""))
	// Escape the message to handle special characters
	escapeJSON(buf, entry.Message)
	buf.Write([]byte("\""))

	// Add PID if needed
	if f.ShowPID {
		buf.Write([]byte(",\"pid\":"))
		pidBuf := util.GetSmallByteSliceFromPool()
		pidBytes := strconv.AppendInt(pidBuf[:0], int64(entry.PID), 10)
		buf.Write(pidBytes)
		util.PutSmallByteSliceToPool(pidBuf)
	}

	// Add caller info if needed
	if f.ShowCaller && entry.Caller != nil {
		buf.Write([]byte(",\"caller\":\""))
		buf.Write(core.S2b(entry.Caller.File))
		buf.WriteByte(':')
		lineBuf := util.GetSmallByteSliceFromPool()
		lineBytes := strconv.AppendInt(lineBuf[:0], int64(entry.Caller.Line), 10)
		buf.Write(lineBytes)
		util.PutSmallByteSliceToPool(lineBuf)
		buf.Write([]byte("\""))
	}

	// Add fields if present
	if len(entry.Fields) > 0 {
		buf.Write([]byte(",\"fields\":{"))
		first := true
		for k, v := range entry.Fields {
			if !first {
				buf.WriteByte(',')
			}
			first = false

			buf.WriteByte('"')
			buf.Write(core.S2b(k))
			buf.Write([]byte("\":"))

			// Format value based on type
			formatJSONValue(buf, v)
		}
		buf.WriteByte('}')
	}

	// Add trace info if needed
	if f.ShowTraceInfo {
		if entry.TraceID != "" {
			buf.Write([]byte(",\"trace_id\":\""))
			buf.Write(core.S2b(entry.TraceID))
			buf.WriteByte('"')
		}
		if entry.SpanID != "" {
			buf.Write([]byte(",\"span_id\":\""))
			buf.Write(core.S2b(entry.SpanID))
			buf.WriteByte('"')
		}
		if entry.UserID != "" {
			buf.Write([]byte(",\"user_id\":\""))
			buf.Write(core.S2b(entry.UserID))
			buf.WriteByte('"')
		}
	}

	if f.EnableStackTrace && len(entry.StackTrace) > 0 {
		buf.Write([]byte(",\"stack_trace\":\""))
		escapeJSON(buf, entry.StackTrace)
		buf.WriteByte('"')
	}

	buf.WriteByte('}')
	buf.WriteByte('\n')

	return nil
}

// formatWithStandardEncoder uses standard encoder (less efficient but with pretty printing)
func (f *JSONFormatter) formatWithStandardEncoder(buf *bytes.Buffer, entry *core.LogEntry) error {
	// For compatibility with JSON marshaling, we need to convert the LogEntry
	// to avoid automatic base64 encoding of []byte fields while maintaining performance.
	// We'll create a JSON-compatible struct representation manually for full control.

	if f.PrettyPrint {
		// For pretty printing, use manual formatting to avoid base64 encoding
		return f.formatManuallyWithIndent(buf, entry)
	} else {
		// Even for non-pretty printing, we need to avoid the standard encoder's base64 behavior
		// by implementing our own encoding that handles []byte as strings
		return f.formatManually(buf, entry)
	}
}

// formatManuallyWithIndent formats JSON with indentation for pretty printing
func (f *JSONFormatter) formatManuallyWithIndent(buf *bytes.Buffer, entry *core.LogEntry) error {
	indentLevel := 0
	indent := func() {
		for i := 0; i < indentLevel; i++ {
			buf.WriteString("  ") // 2 spaces per indent level
		}
	}

	// Start JSON object
	buf.WriteByte('{')
	indentLevel++

	newline := func() {
		buf.WriteByte('\n')
		indent()
	}

	// Add timestamp
	newline()
	buf.Write([]byte("\"timestamp\": \""))
	util.FormatTimestamp(buf, entry.Timestamp, f.TimestampFormat)
	buf.Write([]byte("\""))

	// Add level
	buf.WriteByte(',')
	newline()
	buf.Write([]byte("\"level_name\": \""))
	buf.Write(entry.Level.Bytes()) // Using pre-allocated level bytes
	buf.Write([]byte("\""))

	// Add message
	buf.WriteByte(',')
	newline()
	buf.Write([]byte("\"message\": \""))
	// Escape the message to handle special characters
	escapeJSON(buf, entry.Message)
	buf.Write([]byte("\""))

	// Add PID if needed
	if f.ShowPID && entry.PID != 0 {
		buf.WriteByte(',')
		newline()
		buf.Write([]byte("\"pid\": "))
		pidBuf := util.GetSmallByteSliceFromPool()
		pidBytes := strconv.AppendInt(pidBuf[:0], int64(entry.PID), 10)
		buf.Write(pidBytes)
		util.PutSmallByteSliceToPool(pidBuf)
	}

	// Add caller info if needed
	if f.ShowCaller && entry.Caller != nil {
		buf.WriteByte(',')
		newline()
		buf.Write([]byte("\"caller\": \""))
		buf.Write(core.S2b(entry.Caller.File))
		buf.WriteByte(':')
		lineBuf := util.GetSmallByteSliceFromPool()
		lineBytes := strconv.AppendInt(lineBuf[:0], int64(entry.Caller.Line), 10)
		buf.Write(lineBytes)
		util.PutSmallByteSliceToPool(lineBuf)
		buf.Write([]byte("\""))
	}

	// Add fields if present
	if len(entry.Fields) > 0 {
		buf.WriteByte(',')
		newline()
		buf.Write([]byte("\"fields\": {"))
		indentLevel++
		fieldIndentLevel := indentLevel
		fieldNewline := func() {
			buf.WriteByte('\n')
			for i := 0; i < fieldIndentLevel; i++ {
				buf.WriteString("  ")
			}
		}

		first := true
		for k, v := range entry.Fields {
			if !first {
				buf.WriteByte(',')
			}
			fieldNewline()
			buf.WriteByte('"')
			buf.Write(core.S2b(k))
			buf.Write([]byte("\": "))

			// Format value based on type
			formatJSONValue(buf, v)
			first = false
		}
		indentLevel--
		fieldNewline()
		buf.WriteByte('}')
	}

	// Add trace info if needed
	if f.ShowTraceInfo {
		if entry.TraceID != "" {
			buf.WriteByte(',')
			newline()
			buf.Write([]byte("\"trace_id\": \""))
			buf.Write(core.S2b(entry.TraceID))
			buf.WriteByte('"')
		}
		if entry.SpanID != "" {
			buf.WriteByte(',')
			newline()
			buf.Write([]byte("\"span_id\": \""))
			buf.Write(core.S2b(entry.SpanID))
			buf.WriteByte('"')
		}
		if entry.UserID != "" {
			buf.WriteByte(',')
			newline()
			buf.Write([]byte("\"user_id\": \""))
			buf.Write(core.S2b(entry.UserID))
			buf.WriteByte('"')
		}
	}

	if f.EnableStackTrace && len(entry.StackTrace) > 0 {
		buf.WriteByte(',')
		newline()
		buf.Write([]byte("\"stack_trace\": \""))
		escapeJSON(buf, entry.StackTrace)
		buf.WriteByte('"')
	}

	indentLevel--
	newline()
	buf.WriteByte('}')
	buf.WriteByte('\n')

	return nil
}

// escapeJSON escapes special characters in JSON strings
func escapeJSON(buf *bytes.Buffer, data []byte) {
	for _, b := range data {
		switch b {
		case '"':
			buf.Write([]byte("\\\""))
		case '\\':
			buf.Write([]byte("\\\\"))
		case '\b':
			buf.Write([]byte("\\b"))
		case '\f':
			buf.Write([]byte("\\f"))
		case '\n':
			buf.Write([]byte("\\n"))
		case '\r':
			buf.Write([]byte("\\r"))
		case '\t':
			buf.Write([]byte("\\t"))
		default:
			if b < 0x20 {
				// Manual hex formatting to avoid fmt.Sprintf allocation
				buf.Write([]byte("\\u00"))
				// Convert to hex manually
				hex1 := b / 16
				hex2 := b % 16
				if hex1 < 10 {
					buf.WriteByte('0' + hex1)
				} else {
					buf.WriteByte('a' + hex1 - 10)
				}
				if hex2 < 10 {
					buf.WriteByte('0' + hex2)
				} else {
					buf.WriteByte('a' + hex2 - 10)
				}
			} else {
				buf.WriteByte(b)
			}
		}
	}
}

// formatJSONValue formats a value for JSON output
func formatJSONValue(buf *bytes.Buffer, v interface{}) {
	switch val := v.(type) {
	case string:
		buf.WriteByte('"')
		escapeJSON(buf, core.S2b(val))
		buf.WriteByte('"')
	case []byte:
		buf.WriteByte('"')
		escapeJSON(buf, val)
		buf.WriteByte('"')
	case int:
		tempBuf := util.GetSmallByteSliceFromPool()
		numBytes := strconv.AppendInt(tempBuf[:0], int64(val), 10)
		buf.Write(numBytes)
		util.PutSmallByteSliceToPool(tempBuf)
	case int64:
		tempBuf := util.GetSmallByteSliceFromPool()
		numBytes := strconv.AppendInt(tempBuf[:0], val, 10)
		buf.Write(numBytes)
		util.PutSmallByteSliceToPool(tempBuf)
	case float64:
		tempBuf := util.GetSmallByteSliceFromPool()
		numBytes := strconv.AppendFloat(tempBuf[:0], val, 'g', -1, 64)
		buf.Write(numBytes)
		util.PutSmallByteSliceToPool(tempBuf)
	case bool:
		if val {
			buf.Write([]byte("true"))
		} else {
			buf.Write([]byte("false"))
		}
	case nil:
		buf.Write([]byte("null"))
	default:
		// For types we can't handle efficiently, fall back to string using manual conversion
		// Define locally to avoid conflicts between files
		localManualStringConversion := func(value interface{}) string {
			switch v := value.(type) {
			case string:
				return v
			case []byte:
				return string(v) // This is unavoidable for []byte to string
			case int:
				return strconv.Itoa(v)
			case int8:
				return strconv.FormatInt(int64(v), 10)
			case int16:
				return strconv.FormatInt(int64(v), 10)
			case int32:
				return strconv.FormatInt(int64(v), 10)
			case int64:
				return strconv.FormatInt(v, 10)
			case uint:
				return strconv.FormatUint(uint64(v), 10)
			case uint8:
				return strconv.FormatUint(uint64(v), 10)
			case uint16:
				return strconv.FormatUint(uint64(v), 10)
			case uint32:
				return strconv.FormatUint(uint64(v), 10)
			case uint64:
				return strconv.FormatUint(v, 10)
			case float32:
				return strconv.FormatFloat(float64(v), 'g', -1, 32)
			case float64:
				return strconv.FormatFloat(v, 'g', -1, 64)
			case bool:
				if v {
					return "true"
				}
				return "false"
			case nil:
				return "null"
			default:
				// For complex types that can't be easily converted
				// This is a last resort case - should be avoided in high-performance scenarios
				return "<complex-type>"
			}
		}
		tempStr := localManualStringConversion(val)
		buf.WriteByte('"')
		escapeJSON(buf, core.S2b(tempStr))
		buf.WriteByte('"')
	}
}

// prepareOutputEntry creates a temporary LogEntry for JSON marshaling from a pool.
func (f *JSONFormatter) prepareOutputEntry(entry *core.LogEntry) *core.LogEntry {
	outputEntry := core.GetEntryFromPool()

	// Copy base fields
	outputEntry.Timestamp = entry.Timestamp
	outputEntry.Level = entry.Level
	outputEntry.LevelName = entry.LevelName
	outputEntry.Message = entry.Message
	outputEntry.Error = entry.Error
	outputEntry.Hostname = entry.Hostname
	outputEntry.Application = entry.Application
	outputEntry.Version = entry.Version
	outputEntry.Environment = entry.Environment

	// Conditionally copy fields based on formatter config
	if f.ShowCaller {
		outputEntry.Caller = entry.Caller
	}
	if f.ShowGoroutine {
		outputEntry.GoroutineID = entry.GoroutineID
	}
	if f.ShowPID {
		outputEntry.PID = entry.PID
	}
	if f.ShowTraceInfo {
		outputEntry.TraceID = entry.TraceID
		outputEntry.SpanID = entry.SpanID
		outputEntry.UserID = entry.UserID
		outputEntry.SessionID = entry.SessionID
		outputEntry.RequestID = entry.RequestID
	}
	if f.EnableDuration {
		outputEntry.Duration = entry.Duration
	}
	if f.EnableStackTrace {
		outputEntry.StackTrace = entry.StackTrace
	}

	// Process and copy fields, tags, and metrics
	if len(entry.Fields) > 0 {
		processedFields := f.processFields(entry.Fields)
		for k, v := range processedFields {
			outputEntry.Fields[k] = v
		}
		core.PutMapInterfaceToPool(processedFields)
	}
	if len(entry.Tags) > 0 {
		outputEntry.Tags = append(outputEntry.Tags, entry.Tags...)
	}
	if len(entry.CustomMetrics) > 0 {
		for k, v := range entry.CustomMetrics {
			outputEntry.CustomMetrics[k] = v
		}
	}

	return outputEntry
}

// processFields processes and transforms fields according to configuration
func (f *JSONFormatter) processFields(fields map[string]interface{}) map[string]interface{} {
	processed := core.GetMapInterfaceFromPool()

	for k, v := range fields {
		key := k
		if mappedKey, exists := f.FieldKeyMap[k]; exists {
			key = mappedKey
		}

		if transformer, exists := f.FieldTransformers[k]; exists {
			processed[key] = transformer(v)
		} else if f.MaskSensitiveData && f.isSensitive(k) {
			processed[key] = f.MaskStringBytes // Use byte slice for consistency with zero allocation
		} else {
			processed[key] = v
		}
	}

	return processed
}

func (f *JSONFormatter) isSensitive(field string) bool {
    for _, sensitiveField := range f.SensitiveFields {
        if field == sensitiveField {
            return true
        }
    }
    return false
}
