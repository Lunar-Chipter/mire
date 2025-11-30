package formatter

import (
	"bytes"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
)

// BenchmarkCSVFormatter benchmarks the CSV formatter
func BenchmarkCSVFormatter(b *testing.B) {
	formatter := NewCSVFormatter()
	formatter.FieldOrder = []string{"timestamp", "level", "message", "pid", "trace_id"}
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.LevelName = core.StringToBytes("INFO")
	entry.Message = core.StringToBytes("Test message for benchmark")
	entry.PID = 12345
	entry.TraceID = core.StringToBytes("trace-12345")
	entry.Fields = map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}
	
	var buf bytes.Buffer
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		buf.Reset()
		formatter.Format(&buf, entry)
	}
}

// BenchmarkJSONFormatter benchmarks the JSON formatter (non-pretty)
func BenchmarkJSONFormatter(b *testing.B) {
	formatter := NewJSONFormatter()
	formatter.PrettyPrint = false
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.LevelName = core.StringToBytes("INFO")
	entry.Message = core.StringToBytes("Test message for benchmark")
	entry.PID = 12345
	entry.TraceID = core.StringToBytes("trace-12345")
	entry.Fields = map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}
	
	var buf bytes.Buffer
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		buf.Reset()
		formatter.Format(&buf, entry)
	}
}

// BenchmarkJSONFormatterPretty benchmarks the JSON formatter (pretty-printed)
func BenchmarkJSONFormatterPretty(b *testing.B) {
	formatter := NewJSONFormatter()
	formatter.PrettyPrint = true
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.LevelName = core.StringToBytes("INFO")
	entry.Message = core.StringToBytes("Test message for benchmark")
	entry.PID = 12345
	entry.TraceID = core.StringToBytes("trace-12345")
	entry.Fields = map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}
	
	var buf bytes.Buffer
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		buf.Reset()
		formatter.Format(&buf, entry)
	}
}

// BenchmarkTextFormatter benchmarks the Text formatter
func BenchmarkTextFormatter(b *testing.B) {
	formatter := NewTextFormatter()
	formatter.ShowTimestamp = true
	formatter.ShowCaller = true
	formatter.EnableColors = false
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.LevelName = core.StringToBytes("INFO")
	entry.Message = core.StringToBytes("Test message for benchmark")
	entry.PID = 12345
	entry.TraceID = core.StringToBytes("trace-12345")
	entry.Fields = map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}
	
	var buf bytes.Buffer
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		buf.Reset()
		formatter.Format(&buf, entry)
	}
}

// BenchmarkAllFormatters benchmarks all formatters together
func BenchmarkAllFormatters(b *testing.B) {
	formatters := []struct {
		name string
		formatter Formatter
	}{
		{"CSV", NewCSVFormatter()},
		{"JSON", NewJSONFormatter()},
		{"Text", NewTextFormatter()},
	}

	for _, tt := range formatters {
		b.Run(tt.name, func(b *testing.B) {
			entry := core.GetEntryFromPool()
			defer core.PutEntryToPool(entry)
			
			entry.Timestamp = time.Now()
			entry.Level = core.INFO
			entry.LevelName = core.StringToBytes("INFO")
			entry.Message = core.StringToBytes("Test message for benchmark")
			entry.PID = 12345
			entry.TraceID = core.StringToBytes("trace-12345")
			entry.Fields = map[string]interface{}{
				"user_id": 123,
				"action":  "login",
			}
			
			var buf bytes.Buffer
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				buf.Reset()
				tt.formatter.Format(&buf, entry)
			}
		})
	}
}