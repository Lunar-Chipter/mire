package formatter

import (
	"bytes"
	"encoding/csv"
	"strconv"

	"mire/core"
	"mire/util"
)

// CSVFormatter - Optimized version - Returns byte slice instead of string to reduce memory allocation
// CSVFormatter - Versi optimal - Mengembalikan slice byte bukan string untuk mengurangi alokasi memori
type CSVFormatter struct {
	IncludeHeader   bool     // Include header row in output
	FieldOrder      []string // Order of fields in CSV
	TimestampFormat string   // Custom timestamp format
}

// NewCSVFormatter creates a new CSVFormatter
// NewCSVFormatter membuat CSVFormatter baru
func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{}
}

// Format formats a log entry into CSV byte slice
// Format memformat entri log menjadi slice byte CSV
func (f *CSVFormatter) Format(buf *bytes.Buffer, entry *core.LogEntry) error {
	writer := csv.NewWriter(buf)

	record := util.GetStringSliceFromPool()
	defer util.PutStringSliceToPool(record)
	
	valBuf := util.GetBufferFromPool()
	defer util.PutBufferToPool(valBuf)

	for _, field := range f.FieldOrder {
		valBuf.Reset()
		f.formatField(valBuf, field, entry)
		record = append(record, valBuf.String())
	}

	if err := writer.Write(record); err != nil {
		return err
	}

	writer.Flush()
	return writer.Error()
}

func (f *CSVFormatter) formatField(valBuf *bytes.Buffer, field string, entry *core.LogEntry) {
	switch field {
	case "timestamp":
		valBuf.Write(entry.Timestamp.AppendFormat(nil, f.TimestampFormat))
	case "level":
		valBuf.Write(entry.Level.Bytes())
	case "message":
		valBuf.Write(entry.Message)
	case "pid":
		pidBuf := util.GetSmallByteSliceFromPool()
		defer util.PutSmallByteSliceToPool(pidBuf)
		pidBuf = strconv.AppendInt(pidBuf, int64(entry.PID), 10)
		valBuf.Write(pidBuf)
	case "goroutine_id":
		valBuf.WriteString(entry.GoroutineID)
	case "trace_id":
		valBuf.WriteString(entry.TraceID)
	case "file":
		if entry.Caller != nil {
			valBuf.WriteString(entry.Caller.File)
		}
	case "line":
		if entry.Caller != nil {
			lineBuf := util.GetSmallByteSliceFromPool()
			defer util.PutSmallByteSliceToPool(lineBuf)
			lineBuf = strconv.AppendInt(lineBuf, int64(entry.Caller.Line), 10)
			valBuf.Write(lineBuf)
		}
	case "error":
		if entry.Error != nil {
			// Check if the error implements ErrorAppender for zero-allocation
			if appender, ok := entry.Error.(core.ErrorAppender); ok {
				appender.AppendError(valBuf)
			} else {
				valBuf.WriteString(entry.Error.Error()) // Fallback to standard Error()
			}
		}
	default:
		if val, exists := entry.Fields[field]; exists {
			util.FormatValue(valBuf, val, 0)
		}
	}
}

