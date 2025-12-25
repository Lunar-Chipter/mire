package formatter

import (
	"bytes"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/util"
)

// CSVFormatter formats log entries in CSV format
type CSVFormatter struct {
	IncludeHeader     bool                                // Include header row in output
	FieldOrder        []string                            // Order of fields in CSV
	TimestampFormat   string                              // Custom timestamp format
	SensitiveFields   []string                            // List of sensitive field names to mask
	MaskSensitiveData bool                                // Whether to mask sensitive data
	MaskValue   string                              // String value to use for masking
	FieldTransformers map[string]func(interface{}) string // Functions to transform field values
}

// NewCSVFormatter creates a new CSVFormatter
func NewCSV() *CSVFormatter {
	return &CSVFormatter{
		MaskValue:       "[MASKED]",
		FieldTransformers: make(map[string]func(interface{}) string),
	}
}

// Format formats a log entry into CSV byte slice with zero allocations
func (f *CSVFormatter) Format(buf *bytes.Buffer, entry *core.LogEntry) error {
	estimatedSize := 256
	estimatedSize += len(entry.Message)
	if len(entry.Fields) > 0 {
		estimatedSize += len(entry.Fields) * 32
	}

	currentCap := buf.Cap()
	if estimatedSize > currentCap {
		buf.Grow(estimatedSize - currentCap)
	}

	if f.IncludeHeader {
		headerWritten := buf.Len() == 0
		if headerWritten {
			for i, field := range f.FieldOrder {
				if i > 0 {
					buf.WriteByte(',')
				}
				f.writeCSVValue(buf, field)
			}
			buf.WriteByte('\n')
		}
	}

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

func (f *CSVFormatter) writeCSVValue(buf *bytes.Buffer, value string) {
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
				buf.WriteString(`""`)
			} else {
				buf.WriteByte(b)
			}
		}
		buf.WriteByte('"')
	} else {
		buf.WriteString(value)
	}
}

func (f *CSVFormatter) writeCSVValueBytes(buf *bytes.Buffer, value []byte) {
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
				buf.WriteString(`""`)
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
		timestamp := util.GetBuffer()
		format := f.TimestampFormat
		if format == "" {
			format = "2006-01-02 15:04:05.000"
		}
		util.FormatTimestamp(timestamp, entry.Timestamp, format)
		f.writeCSVValueBytes(buf, timestamp.Bytes())
		util.PutBuffer(timestamp)
	case "level":
		f.writeCSVValueBytes(buf, entry.Level.Bytes())
	case "message":
		f.writeCSVValueBytes(buf, entry.Message)
	case "pid":
		buf.WriteByte('"')
		util.WriteInt(buf, int64(entry.PID))
		buf.WriteByte('"')
	case "goroutine_id":
		f.writeCSVValueBytes(buf, entry.GoroutineID)
	case "trace_id":
		f.writeCSVValueBytes(buf, entry.TraceID)
	case "span_id":
		f.writeCSVValueBytes(buf, entry.SpanID)
	case "user_id":
		f.writeCSVValueBytes(buf, entry.UserID)
	case "request_id":
		f.writeCSVValueBytes(buf, entry.RequestID)
	case "file":
		if entry.Caller != nil {
			f.writeCSVValue(buf, entry.Caller.File)
		} else {
			buf.WriteByte('"')
			buf.WriteByte('"')
		}
	case "line":
		if entry.Caller != nil {
			buf.WriteByte('"')
			util.WriteInt(buf, int64(entry.Caller.Line))
			buf.WriteByte('"')
		} else {
			buf.WriteByte('"')
			buf.WriteByte('"')
		}
	case "error":
		if entry.Error != nil {
			if appender, ok := entry.Error.(core.ErrAppend); ok {
				buf.WriteByte('"')
				appender.AppendError(buf)
				buf.WriteByte('"')
			} else {
				f.writeCSVValue(buf, entry.Error.Error())
			}
		} else {
			buf.WriteByte('"')
			buf.WriteByte('"')
		}
	default:
		if val, exists := entry.Fields[field]; exists {
			if f.MaskSensitiveData && f.isSensitiveField(field) {
				f.writeCSVValue(buf, f.MaskValue)
				return nil
			}

			if transformer, exists := f.FieldTransformers[field]; exists {
				transformed := transformer(val)
				buf.WriteByte('"')
				util.FormatValue(buf, transformed, 0)
				buf.WriteByte('"')
			} else {
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

func (f *CSVFormatter) isSensitiveField(field string) bool {
	for _, sensitiveField := range f.SensitiveFields {
		if field == sensitiveField {
			return true
		}
	}
	return false
}
