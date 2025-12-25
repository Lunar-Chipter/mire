package formatter

import (
	"bytes"
	"strconv"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/util"
)

// to
var (
	jsonTimestampKey = []byte("\"timestamp\":\"")
	jsonLevelKey     = []byte("\"level_name\":\"")
	jsonMessageKey   = []byte("\"message\":\"")
	jsonPidKey       = []byte(",\"pid\":")
	jsonCallerKey    = []byte(",\"caller\":\"")
	jsonTraceKey     = []byte(",\"trace_id\":\"")
	jsonSpanKey      = []byte(",\"span_id\":\"")
	jsonUserKey      = []byte(",\"user_id\":\"")
	jsonFieldsKey    = []byte(",\"fields\":")
	jsonStackKey     = []byte(",\"stack_trace\":")
	jsonQuote        = []byte("\"")
	jsonComma        = []byte(",")
	jsonColon        = []byte(":")
	jsonBraceOpen    = []byte("{")
)

// JSONFormatter formats log entries in JSON format
type JSONFormatter struct {
	PrettyPrint       bool                                     // Enable pretty-printed JSON
	TimestampFormat   string                                   // Custom timestamp format
	ShowCaller        bool                                     // Show caller information
	ShowGoroutine     bool                                     // Show goroutine ID
	ShowPID           bool                                     // Show process ID
	ShowTrace     bool                                     // Show trace information
	IncludeStackTrace  bool                                     // Enable stack trace for errors
	ShowDuration    bool                                     // Show operation duration
	FieldKeyMap       map[string]string                        // Map for renaming fields
	DisableHTMLEscape bool                                     // Disable HTML escaping in JSON
	SensitiveFields   []string                                 // List of sensitive field names
	MaskSensitiveData bool                                     // Whether to mask sensitive data
	MaskValue   string                                   // String value to use for masking
	MaskStringBytes   []byte                                   // Byte slice for masking (zero-allocation)
	FieldTransformers map[string]func(interface{}) interface{} // Functions to transform field values
}

// NewJSONFormatter creates a new JSONFormatter
func NewJSON() *JSONFormatter {
	return &JSONFormatter{
		MaskValue:   "[MASKED]",
		MaskStringBytes: []byte("[MASKED]"),
		FieldKeyMap:       make(map[string]string),
		FieldTransformers: make(map[string]func(interface{}) interface{}),
		SensitiveFields:   make([]string, 0),
	}
}

// Format formats a log entry into JSON
func (f *JSONFormatter) Format(buf *bytes.Buffer, entry *core.LogEntry) error {
	if f.PrettyPrint {
		return f.formatWithStandardEncoder(buf, entry)
	}

	return f.formatManually(buf, entry)
}

// formatManually creates JSON manually
func (f *JSONFormatter) formatManually(buf *bytes.Buffer, entry *core.LogEntry) error {
	buf.Write(jsonBraceOpen)

	// Add timestamp - manually format to avoid allocation
	buf.Write(jsonTimestampKey)
	util.FormatTimestamp(buf, entry.Timestamp, f.TimestampFormat)
	buf.Write(jsonQuote)
	buf.Write(jsonComma)

	// Add level
	buf.Write(jsonLevelKey)
	buf.Write(entry.Level.Bytes()) // Using pre-allocated level bytes
	buf.Write(jsonQuote)
	buf.Write(jsonComma)

	// Add message
	buf.Write(jsonMessageKey)
	// Escape the message to handle special characters
	escapeJSON(buf, entry.Message)
	buf.Write(jsonQuote)

	// at
	if f.ShowPID {
		buf.Write(jsonPidKey)
		util.WriteInt(buf, int64(entry.PID))
	}

	// Add caller info if needed
	if f.ShowCaller && entry.Caller != nil {
		buf.Write(jsonCallerKey)
		buf.Write(core.StringToBytes(entry.Caller.File))
		buf.Write(jsonColon)
		util.WriteInt(buf, int64(entry.Caller.Line))
		buf.WriteByte('"')
	}

	// Add fields if present
	if len(entry.Fields) > 0 {
		buf.Write(jsonFieldsKey)
		f.formatFields(buf, entry.Fields)
	}

	// Add trace info if needed - organize in a way that reduces branching
	if f.ShowTrace {
		if entry.TraceID != nil {
			buf.Write(jsonTraceKey)
			buf.Write(entry.TraceID)
			buf.Write(jsonQuote)
		}
		if entry.SpanID != nil {
			buf.Write(jsonSpanKey)
			buf.Write(entry.SpanID)
			buf.Write(jsonQuote)
		}
		if entry.UserID != nil {
			buf.Write(jsonUserKey)
			buf.Write(entry.UserID)
			buf.Write(jsonQuote)
		}
	}

	if f.IncludeStackTrace && len(entry.StackTrace) > 0 {
		buf.Write(jsonStackKey)
		escapeJSON(buf, entry.StackTrace)
		buf.WriteByte('"')
	}

	buf.WriteByte('}')
	buf.WriteByte('\n')

	return nil
}

// formatWithStandardEncoder handles pretty printing
func (f *JSONFormatter) formatWithStandardEncoder(buf *bytes.Buffer, entry *core.LogEntry) error {
	if f.PrettyPrint {
		return f.formatManuallyWithIndent(buf, entry)
	}
	return f.formatManually(buf, entry)
}

// formatManuallyWithIndent formats JSON with indentation
func (f *JSONFormatter) formatManuallyWithIndent(buf *bytes.Buffer, entry *core.LogEntry) error {
	indentBuf := util.GetBuffer()
	defer util.PutBuffer(indentBuf)

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
		buf.Write(core.StringToBytes(entry.Caller.File))
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
	if f.ShowTrace {
		if entry.TraceID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"trace_id\": \"")
			buf.Write(entry.TraceID)
			buf.WriteByte('"')
		}
		if entry.SpanID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"span_id\": \"")
			buf.Write(entry.SpanID)
			buf.WriteByte('"')
		}
		if entry.UserID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"user_id\": \"")
			buf.Write(entry.UserID)
			buf.WriteByte('"')
		}
	}

	if f.IncludeStackTrace && len(entry.StackTrace) > 0 {
		buf.WriteString(",\n  ")
		indent(1)
		buf.WriteString("\"stack_trace\": \"")
		escapeJSON(buf, entry.StackTrace)
		buf.WriteByte('"')
	}

	// Add caller info if needed
	if f.ShowCaller && entry.Caller != nil {
		buf.WriteString(",\n  ")
		indent(1)
		buf.WriteString("\"caller\": \"")
		buf.Write(core.StringToBytes(entry.Caller.File))
		buf.WriteByte(':')
		util.WriteInt(buf, int64(entry.Caller.Line))
		buf.WriteByte('"')
	}

	// Add trace info if needed - organize in a way that reduces branching
	if f.ShowTrace {
		if entry.TraceID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"trace_id\": \"")
			buf.Write(entry.TraceID)
			buf.WriteByte('"')
		}
		if entry.SpanID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"span_id\": \"")
			buf.Write(entry.SpanID)
			buf.WriteByte('"')
		}
		if entry.UserID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"user_id\": \"")
			buf.Write(entry.UserID)
			buf.WriteByte('"')
		}
	}

	if f.IncludeStackTrace && len(entry.StackTrace) > 0 {
		buf.WriteString(",\n  ")
		indent(1)
		buf.WriteString("\"stack_trace\": \"")
		escapeJSON(buf, entry.StackTrace)
		buf.WriteByte('"')
	}

	// Add caller info if needed
	if f.ShowCaller && entry.Caller != nil {
		buf.WriteString(",\n  ")
		indent(1)
		buf.WriteString("\"caller\": \"")
		buf.Write(core.StringToBytes(entry.Caller.File))
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
	if f.ShowTrace {
		if entry.TraceID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"trace_id\": \"")
			buf.Write(entry.TraceID)
			buf.WriteByte('"')
		}
		if entry.SpanID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"span_id\": \"")
			buf.Write(entry.SpanID)
			buf.WriteByte('"')
		}
		if entry.UserID != nil {
			buf.WriteString(",\n  ")
			indent(1)
			buf.WriteString("\"user_id\": \"")
			buf.Write(entry.UserID)
			buf.WriteByte('"')
		}
	}

	if f.IncludeStackTrace && len(entry.StackTrace) > 0 {
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

	escaped := util.GetBuffer()
	defer util.PutBuffer(escaped)

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
				escaped.Write([]byte("\\u00"))
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

		if escaped.Len() > 1024 {
			buf.Write(escaped.Bytes())
			escaped.Reset()
		}
	}

	if escaped.Len() > 0 {
		buf.Write(escaped.Bytes())
	}
}

// formatJSONValue formats a value for JSON output
func (f *JSONFormatter) formatJSONValue(buf *bytes.Buffer, v interface{}) {
	switch val := v.(type) {
	case string:
		buf.WriteByte('"')
		escapeJSON(buf, core.StringToBytes(val))
		buf.WriteByte('"')
	case []byte:
		buf.WriteByte('"')
		escapeJSON(buf, val)
		buf.WriteByte('"')
	case int:
		tempBuf := util.GetSmallBuf()
		numBytes := strconv.AppendInt(tempBuf[:0], int64(val), 10)
		buf.Write(numBytes)
		util.PutSmallBuf(tempBuf)
	case int64:
		tempBuf := util.GetSmallBuf()
		numBytes := strconv.AppendInt(tempBuf[:0], val, 10)
		buf.Write(numBytes)
		util.PutSmallBuf(tempBuf)
	case float64:
		tempBuf := util.GetSmallBuf()
		numBytes := strconv.AppendFloat(tempBuf[:0], val, 'g', -1, 64)
		buf.Write(numBytes)
		util.PutSmallBuf(tempBuf)
	case bool:
		if val {
			buf.Write([]byte("true"))
		} else {
			buf.Write([]byte("false"))
		}
	case nil:
		buf.Write([]byte("null"))
	default:
		transformed := f.transformValue(val, "<complex-type>")
		buf.WriteByte('"')
		escapeJSON(buf, core.StringToBytes(transformed))
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

// transformValue converts value to string
func (f *JSONFormatter) transformValue(val interface{}, defaultVal string) string {
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
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
		return defaultVal
	}
}

// formatFields formats the fields map in JSON format
func (f *JSONFormatter) formatFields(buf *bytes.Buffer, fields map[string][]byte) {
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

		buf.WriteByte('"')
		buf.Write(core.StringToBytes(k))
		buf.Write([]byte("\":"))

		buf.WriteByte('"')
		if f.MaskSensitiveData && f.isSensitiveField(k) {
			buf.Write(core.StringToBytes(f.MaskValue))
		} else {
			escapeJSON(buf, v)
		}
		buf.WriteByte('"')
	}

	buf.Write([]byte("}"))
}

// formatFieldsIndented formats the fields map in JSON format with indentation
func (f *JSONFormatter) formatFieldsIndented(buf *bytes.Buffer, fields map[string][]byte, indentLevel int) {
	indentBuf := util.GetBuffer()
	defer util.PutBuffer(indentBuf)

	for i := 0; i < indentLevel; i++ {
		indentBuf.WriteString("  ")
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

	if len(fields) > 0 {
		indentBuf.WriteString("  ")
		indentBytes = indentBuf.Bytes()
		newlineAndIndent()
	}

	fieldNewline := func() {
		buf.WriteByte('\n')
		buf.Write(indentBytes)
	}

	// to
	orderedKeys := make([]string, 0, len(fields))
	for k := range fields {
		orderedKeys = append(orderedKeys, k)
	}

	first := true
	for _, k := range orderedKeys {
		v := fields[k]
		if !first {
			buf.WriteByte(',')
		}
		fieldNewline()

		buf.WriteByte('"')
		buf.Write(core.StringToBytes(k))
		buf.Write([]byte("\": "))

		buf.WriteByte('"')
		if f.MaskSensitiveData && f.isSensitiveField(k) {
			buf.Write(f.MaskStringBytes)
		} else {
			escapeJSON(buf, v)
		}
		buf.WriteByte('"')
		first = false
	}

	if len(originalIndent) >= 2 {
		indentBytes = originalIndent[:len(originalIndent)-2]
	} else {
		indentBytes = originalIndent[:0]
	}

	buf.WriteByte('\n')
	buf.Write(indentBytes)
	buf.WriteByte('}')
}
