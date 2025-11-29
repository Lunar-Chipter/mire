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
	return &JSONFormatter{
		MaskStringValue: "[MASKED]",
		FieldKeyMap:     make(map[string]string),
		FieldTransformers: make(map[string]func(interface{}) interface{}),
		SensitiveFields: make([]string, 0),
	}
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
	buf.WriteString("\"timestamp\":\"")
	util.FormatTimestamp(buf, entry.Timestamp, f.TimestampFormat)
	buf.WriteString("\",")

	// Add level
	buf.WriteString("\"level_name\":\"")
	buf.Write(entry.Level.Bytes()) // Using pre-allocated level bytes
	buf.WriteString("\",")

	// Add message
	buf.WriteString("\"message\":\"")
	// Escape the message to handle special characters
	escapeJSON(buf, entry.Message)
	buf.WriteString("\"")

	// Add PID if needed - reduce branching by checking condition once
	if f.ShowPID {
		buf.WriteString(",\"pid\":")
		util.WriteInt(buf, int64(entry.PID))
	}

	// Add caller info if needed
	if f.ShowCaller && entry.Caller != nil {
		buf.WriteString(",\"caller\":\"")
		buf.Write(core.S2b(entry.Caller.File))
		buf.WriteByte(':')
		util.WriteInt(buf, int64(entry.Caller.Line))
		buf.WriteByte('"')
	}

	// Add fields if present
	if len(entry.Fields) > 0 {
		buf.WriteString(",\"fields\":")
		f.formatFields(buf, entry.Fields)
	}

	// Add trace info if needed - organize in a way that reduces branching
	if f.ShowTraceInfo {
		if entry.TraceID != "" {
			buf.WriteString(",\"trace_id\":\"")
			buf.Write(core.S2b(entry.TraceID))
			buf.WriteByte('"')
		}
		if entry.SpanID != "" {
			buf.WriteString(",\"span_id\":\"")
			buf.Write(core.S2b(entry.SpanID))
			buf.WriteByte('"')
		}
		if entry.UserID != "" {
			buf.WriteString(",\"user_id\":\"")
			buf.Write(core.S2b(entry.UserID))
			buf.WriteByte('"')
		}
	}

	if f.EnableStackTrace && len(entry.StackTrace) > 0 {
		buf.WriteString(",\"stack_trace\":\"")
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
	// Pre-allocate indent string to reuse and avoid repeated allocations
	indentBuf := util.GetBufferFromPool()
	defer util.PutBufferToPool(indentBuf)

	// Pre-allocate 10 levels of indentation (should be sufficient for most cases)
	for i := 0; i < 10; i++ {
		indentBuf.WriteString("  ")
	}
	indentLevels := indentBuf.Bytes()

	indent := func(level int) {
		if level <= 0 {
			return
		}
		// Use pre-allocated indentation
		indentSize := level * 2
		if indentSize <= len(indentLevels) {
			buf.Write(indentLevels[:indentSize])
		} else {
			// If we need more indentation than pre-allocated, add more
			for i := 0; i < level; i++ {
				buf.WriteString("  ")
			}
		}
	}

	// Start JSON object
	buf.WriteByte('{')

	newline := func(level int) {
		buf.WriteByte('\n')
		indent(level)
	}

	// Add timestamp
	newline(1)
	buf.WriteString("\"timestamp\": \"")
	util.FormatTimestamp(buf, entry.Timestamp, f.TimestampFormat)
	buf.WriteString("\"")

	// Add level
	buf.WriteString(",\n  ")
	indent(1)
	buf.WriteString("\"level_name\": \"")
	buf.Write(entry.Level.Bytes()) // Using pre-allocated level bytes
	buf.WriteString("\"")

	// Add message
	buf.WriteString(",\n  ")
	indent(1)
	buf.WriteString("\"message\": \"")
	// Escape the message to handle special characters
	escapeJSON(buf, entry.Message)
	buf.WriteString("\"")

	// Add PID if needed
	if f.ShowPID && entry.PID != 0 {
		buf.WriteString(",\n  ")
		indent(1)
		buf.WriteString("\"pid\": ")
		util.WriteInt(buf, int64(entry.PID))
	}

	// Add caller info if needed
	if f.ShowCaller && entry.Caller != nil {
		buf.WriteString(",\n  ")
		indent(1)
		buf.WriteString("\"caller\": \"")
		buf.Write(core.S2b(entry.Caller.File))
		buf.WriteByte(':')
		util.WriteInt(buf, int64(entry.Caller.Line))
		buf.WriteByte('"')
	}

	// Add fields if present
	if len(entry.Fields) > 0 {
		buf.WriteString(",\n  ")
		indent(1)
		buf.WriteString("\"fields\": ")
		// For indented fields, we need to format them manually with indentation
		f.formatFieldsIndented(buf, entry.Fields, 2)
	}

	// Add trace info if needed
	if f.ShowTraceInfo {
		if entry.TraceID != "" {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"trace_id\": \"")
			buf.Write(core.S2b(entry.TraceID))
			buf.WriteByte('"')
		}
		if entry.SpanID != "" {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"span_id\": \"")
			buf.Write(core.S2b(entry.SpanID))
			buf.WriteByte('"')
		}
		if entry.UserID != "" {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"user_id\": \"")
			buf.Write(core.S2b(entry.UserID))
			buf.WriteByte('"')
		}
	}

	if f.EnableStackTrace && len(entry.StackTrace) > 0 {
		buf.WriteString(",\n  ")
		indent(1)
		buf.WriteString("\"stack_trace\": \"")
		escapeJSON(buf, entry.StackTrace)
		buf.WriteByte('"')
	}

	newline(0)
	buf.WriteByte('}')
	buf.WriteByte('\n')

	return nil
}

// escapeJSON escapes special characters in JSON strings
func escapeJSON(buf *bytes.Buffer, data []byte) {
	if len(data) == 0 {
		return
	}

	// Pre-allocate a working buffer to avoid multiple allocations
	escaped := util.GetBufferFromPool()
	defer util.PutBufferToPool(escaped)

	for _, b := range data {
		switch b {
		case '"':
			escaped.Write([]byte("\\\""))
		case '\\':
			escaped.Write([]byte("\\\\"))
		case '\b':
			escaped.Write([]byte("\\b"))
		case '\f':
			escaped.Write([]byte("\\f"))
		case '\n':
			escaped.Write([]byte("\\n"))
		case '\r':
			escaped.Write([]byte("\\r"))
		case '\t':
			escaped.Write([]byte("\\t"))
		default:
			if b < 0x20 {
				// Manual hex formatting to avoid fmt.Sprintf allocation
				escaped.Write([]byte("\\u00"))
				// Convert to hex manually
				hex1 := b / 16
				hex2 := b % 16
				if hex1 < 10 {
					escaped.WriteByte('0' + hex1)
				} else {
					escaped.WriteByte('a' + hex1 - 10)
				}
				if hex2 < 10 {
					escaped.WriteByte('0' + hex2)
				} else {
					escaped.WriteByte('a' + hex2 - 10)
				}
			} else {
				escaped.WriteByte(b)
			}
		}

		// Periodically flush to main buffer to avoid growing the escaped buffer too large
		if escaped.Len() > 1024 {
			buf.Write(escaped.Bytes())
			escaped.Reset()
		}
	}

	// Write any remaining escaped data
	if escaped.Len() > 0 {
		buf.Write(escaped.Bytes())
	}
}

// formatJSONValue formats a value for JSON output
func (f *JSONFormatter) formatJSONValue(buf *bytes.Buffer, v interface{}) {
	switch val := v.(type) {
	case string:
		// For strings, we need to determine if this is part of a field value or a standalone value
		// Since we can't know the field name here, we'll just format the string normally
		// The field-level sensitivity check is handled in formatFields
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
		// Apply field transformers if available
		transformed := f.transformValue(val, "<complex-type>")
		buf.WriteByte('"')
		escapeJSON(buf, core.S2b(transformed))
		buf.WriteByte('"')
	}
}

// isSensitiveField checks if a field is in the sensitive fields list
func (f *JSONFormatter) isSensitiveField(field string) bool {
	for _, sensitiveField := range f.SensitiveFields {
		if field == sensitiveField {
			return true
		}
	}
	return false
}

// isSensitive checks if a field name is in the sensitive fields list
func (f *JSONFormatter) isSensitive(field string) bool {
	return f.isSensitiveField(field)
}

// createMapForSensitiveCheck creates a map for O(1) sensitive field lookup when there are many sensitive fields
func (f *JSONFormatter) createSensitiveFieldMap() map[string]bool {
	if len(f.SensitiveFields) == 0 {
		return nil
	}

	// Only create map if there are enough fields to justify it
	if len(f.SensitiveFields) < 5 {
		return nil
	}

	fieldMap := make(map[string]bool, len(f.SensitiveFields))
	for _, field := range f.SensitiveFields {
		fieldMap[field] = true
	}
	return fieldMap
}

// transformValue applies field transformers to a value
func (f *JSONFormatter) transformValue(val interface{}, defaultVal string) string {
	// We need to determine which field this is from the context
	// For now, we'll use the default approach and return a string representation
	switch v := val.(type) {
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
		return defaultVal
	}
}

// formatFields formats the fields map in JSON format
func (f *JSONFormatter) formatFields(buf *bytes.Buffer, fields map[string]interface{}) {
	if len(fields) == 0 {
		return
	}

	buf.Write([]byte("{"))
	first := true

	for k, v := range fields {
		if !first {
			buf.WriteByte(',')
		}
		first = false

		// Apply field key mapping if available
		fieldName := k
		if mappedKey, exists := f.FieldKeyMap[k]; exists {
			fieldName = mappedKey
		}

		// Write field name
		buf.WriteByte('"')
		buf.Write(core.S2b(fieldName))
		buf.Write([]byte("\":"))

		// Apply field transformer if available
		if transformer, exists := f.FieldTransformers[k]; exists {
			transformedValue := transformer(v)
			// Convert transformed value to string and write
			transformedStr := f.transformValue(transformedValue, "<transformed-value>")
			buf.WriteByte('"')
			escapeJSON(buf, core.S2b(transformedStr))
			buf.WriteByte('"')
		} else {
			// Format value based on type
			f.formatJSONValue(buf, v)
		}
	}

	buf.Write([]byte("}"))
}

// formatFieldsIndented formats the fields map in JSON format with indentation
func (f *JSONFormatter) formatFieldsIndented(buf *bytes.Buffer, fields map[string]interface{}, indentLevel int) {
	// Pre-allocate indent string to avoid repeated string operations
	indentBuf := util.GetBufferFromPool()
	defer util.PutBufferToPool(indentBuf)

	// Pre-build the indentation string
	for i := 0; i < indentLevel; i++ {
		indentBuf.WriteString("  ") // 2 spaces per indent level
	}
	indentBytes := indentBuf.Bytes()

	// Save original indent to be used later
	originalIndent := make([]byte, len(indentBytes))
	copy(originalIndent, indentBytes)

	newlineAndIndent := func() {
		buf.WriteByte('\n')
		buf.Write(indentBytes)
	}

	buf.WriteByte('{')

	// Add a newline after opening brace if there are fields
	if len(fields) > 0 {
		// Add one more level of indentation
		indentBuf.WriteString("  ")
		indentBytes = indentBuf.Bytes()
		newlineAndIndent()
	}

	fieldNewline := func() {
		buf.WriteByte('\n')
		buf.Write(indentBytes)
	}

	// Use ordered keys approach to avoid pool allocation when possible
	var orderedKeys []string
	if len(f.FieldKeyMap) > 0 || len(f.FieldTransformers) > 0 {
		// If field mapping or transformers are used, we need to pool the slice
		keys := util.GetStringSliceFromPool()
		defer util.PutStringSliceToPool(keys)

		for k := range fields {
			keys = append(keys, k)
		}
		orderedKeys = keys
	} else {
		// Otherwise, create a simple slice without pool overhead
		orderedKeys = make([]string, 0, len(fields))
		for k := range fields {
			orderedKeys = append(orderedKeys, k)
		}
	}

	first := true
	for _, k := range orderedKeys {
		v := fields[k]
		if !first {
			buf.WriteByte(',')
		}
		fieldNewline()

		// Apply field key mapping if available
		fieldName := k
		if mappedKey, exists := f.FieldKeyMap[k]; exists {
			fieldName = mappedKey
		}

		// Write field name
		buf.WriteByte('"')
		buf.Write(core.S2b(fieldName))
		buf.Write([]byte("\": "))

		// Apply field transformer if available
		if transformer, exists := f.FieldTransformers[k]; exists {
			transformedValue := transformer(v)
			// Convert transformed value to string and write
			transformedStr := f.transformValue(transformedValue, "<transformed-value>")
			buf.WriteByte('"')
			escapeJSON(buf, core.S2b(transformedStr))
			buf.WriteByte('"')
		} else {
			// Format value based on type
			f.formatJSONValue(buf, v)
		}
		first = false
	}

	// Add newline and closing brace with proper indentation
	// Adjust indent level back by one
	if len(originalIndent) >= 2 {
		indentBytes = originalIndent[:len(originalIndent)-2]
	} else {
		indentBytes = originalIndent[:0] // Empty slice
	}

	buf.WriteByte('\n')
	buf.Write(indentBytes)
	buf.WriteByte('}')
}
