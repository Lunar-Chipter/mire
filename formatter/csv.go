package formatter

import (
	"bytes"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/util"
)

// CSVFormatter formats log entries in CSV format
// CSVFormatter memformat entri log dalam format CSV
type CSVFormatter struct {
	IncludeHeader     bool                                       // Include header row in output
	FieldOrder        []string                                   // Order of fields in CSV
	TimestampFormat   string                                     // Custom timestamp format
	SensitiveFields   []string                                   // List of sensitive field names to mask
	MaskSensitiveData bool                                       // Whether to mask sensitive data
	MaskStringValue   string                                     // String value to use for masking
	FieldTransformers map[string]func(interface{}) string        // Functions to transform field values
}

// NewCSVFormatter creates a new CSVFormatter
// NewCSVFormatter membuat CSVFormatter baru
func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{
		MaskStringValue: "[MASKED]",
		FieldTransformers: make(map[string]func(interface{}) string),
	}
}

// Format formats a log entry into CSV byte slice with zero allocations
// Format memformat entri log menjadi slice byte CSV tanpa alokasi
func (f *CSVFormatter) Format(buf *bytes.Buffer, entry *core.LogEntry) error {
	// Pre-allocate buffer space if possible
	estimatedSize := 256 // Base size for common fields
	estimatedSize += len(entry.Message)
	if len(entry.Fields) > 0 {
		estimatedSize += len(entry.Fields) * 32 // Estimate for fields
	}

	currentCap := buf.Cap()
	if estimatedSize > currentCap {
		buf.Grow(estimatedSize - currentCap)
	}

	// Write header if needed
	if f.IncludeHeader {
		headerWritten := buf.Len() == 0 // Only write header if buffer is empty
		if headerWritten {
			for i, field := range f.FieldOrder {
				if i > 0 {
					buf.WriteByte(',')
				}
				// Escape field name if needed
				f.writeCSVValue(buf, field)
			}
			buf.WriteByte('\n')
		}
	}

	// Write CSV values
	for i, field := range f.FieldOrder {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := f.formatCSVField(buf, field, entry); err != nil {
			return err
		}
	}
	buf.WriteByte('\n')

	return nil
}

// writeCSVValue writes a value to CSV format, escaping if necessary
// This is a manual implementation to avoid allocations from encoding/csv
func (f *CSVFormatter) writeCSVValue(buf *bytes.Buffer, value string) {
	// Check if value needs escaping by scanning for special characters
	needsEscaping := false
	for i := 0; i < len(value); i++ {
		b := value[i]
		if b == '"' || b == ',' || b == '\n' || b == '\r' {
			needsEscaping = true
			break
		}
	}

	if needsEscaping {
		buf.WriteByte('"')
		for i := 0; i < len(value); i++ {
			b := value[i]
			if b == '"' {
				buf.WriteString(`""`) // Double quotes to escape
			} else {
				buf.WriteByte(b)
			}
		}
		buf.WriteByte('"')
	} else {
		buf.WriteString(value)
	}
}

// writeCSVValueBytes writes a byte slice to CSV format, escaping if necessary
func (f *CSVFormatter) writeCSVValueBytes(buf *bytes.Buffer, value []byte) {
	// Check if value needs escaping by scanning for special characters directly in bytes
	needsEscaping := false
	for _, b := range value {
		if b == '"' || b == ',' || b == '\n' || b == '\r' {
			needsEscaping = true
			break
		}
	}

	if needsEscaping {
		buf.WriteByte('"')
		for _, b := range value {
			if b == '"' {
				buf.WriteString(`""`) // Double quotes to escape
			} else {
				buf.WriteByte(b)
			}
		}
		buf.WriteByte('"')
	} else {
		buf.Write(value)
	}
}

func (f *CSVFormatter) formatCSVField(buf *bytes.Buffer, field string, entry *core.LogEntry) error {
	switch field {
	case "timestamp":
		timestamp := util.GetBufferFromPool()
		util.FormatTimestamp(timestamp, entry.Timestamp, f.TimestampFormat)
		f.writeCSVValueBytes(buf, timestamp.Bytes())
		util.PutBufferToPool(timestamp)
	case "level":
		f.writeCSVValueBytes(buf, entry.Level.Bytes())
	case "message":
		f.writeCSVValueBytes(buf, entry.Message)
	case "pid":
		// Write integer directly to the main buffer to avoid an extra copy
		buf.WriteByte('"')
		util.WriteInt(buf, int64(entry.PID))
		buf.WriteByte('"')
	case "goroutine_id":
		f.writeCSVValue(buf, entry.GoroutineID)
	case "trace_id":
		f.writeCSVValue(buf, entry.TraceID)
	case "span_id":
		f.writeCSVValue(buf, entry.SpanID)
	case "user_id":
		f.writeCSVValue(buf, entry.UserID)
	case "request_id":
		f.writeCSVValue(buf, entry.RequestID)
	case "file":
		if entry.Caller != nil {
			f.writeCSVValue(buf, entry.Caller.File)
		} else {
			buf.WriteByte('"')
			buf.WriteByte('"')
		}
	case "line":
		if entry.Caller != nil {
			// Write integer directly to the main buffer to avoid an extra copy
			buf.WriteByte('"')
			util.WriteInt(buf, int64(entry.Caller.Line))
			buf.WriteByte('"')
		} else {
			buf.WriteByte('"')
			buf.WriteByte('"')
		}
	case "error":
		if entry.Error != nil {
			// Check if the error implements ErrorAppender for zero-allocation
			if appender, ok := entry.Error.(core.ErrorAppender); ok {
				// Use the buffer directly for the appender to avoid extra copy
				buf.WriteByte('"')
				appender.AppendError(buf)
				buf.WriteByte('"')
			} else {
				f.writeCSVValue(buf, entry.Error.Error()) // Fallback to standard Error()
			}
		} else {
			buf.WriteByte('"')
			buf.WriteByte('"')
		}
	default:
		if val, exists := entry.Fields[field]; exists {
			// Check for sensitive fields that need masking
			if f.MaskSensitiveData && f.isSensitiveField(field) {
				f.writeCSVValue(buf, f.MaskStringValue)
				return nil
			}

			// Apply field transformer if available
			if transformer, exists := f.FieldTransformers[field]; exists {
				transformed := transformer(val)
				f.writeCSVValue(buf, transformed)
			} else {
				// Use a temporary buffer for formatting the value
				buf.WriteByte('"')
				util.FormatValue(buf, val, 0)
				buf.WriteByte('"')
			}
		} else {
			buf.WriteByte('"')
			buf.WriteByte('"')
		}
	}
	return nil
}

// isSensitiveField checks if a field is in the sensitive fields list
func (f *CSVFormatter) isSensitiveField(field string) bool {
	for _, sensitiveField := range f.SensitiveFields {
		if field == sensitiveField {
			return true
		}
	}
	return false
}

