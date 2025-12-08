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
	formatter.IncludeHeader = false
	formatter.FieldOrder = []string{"timestamp", "level", "message", "pid", "trace_id"}

	entry := createBenchmarkEntry()
	defer core.PutEntryToPool(entry)

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

	entry := createBenchmarkEntry()
	defer core.PutEntryToPool(entry)

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

	entry := createBenchmarkEntry()
	defer core.PutEntryToPool(entry)

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

	entry := createBenchmarkEntry()
	defer core.PutEntryToPool(entry)

	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		formatter.Format(&buf, entry)
	}
}

// BenchmarkTextFormatterWithColors benchmarks the Text formatter with colors
func BenchmarkTextFormatterWithColors(b *testing.B) {
	formatter := NewTextFormatter()
	formatter.ShowTimestamp = true
	formatter.ShowCaller = true
	formatter.EnableColors = true

	entry := createBenchmarkEntry()
	defer core.PutEntryToPool(entry)

	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		formatter.Format(&buf, entry)
	}
}

// BenchmarkCSVFormatterBatch benchmarks the CSV formatter in batch mode
func BenchmarkCSVFormatterBatch(b *testing.B) {
	formatter := NewCSVFormatter()
	formatter.IncludeHeader = false

	entry := createBenchmarkEntry()
	defer core.PutEntryToPool(entry)

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
		name      string
		formatter Formatter
	}{
		{"CSV", NewCSVFormatter()},
		{"JSON", NewJSONFormatter()},
		{"JSON-Pretty", &JSONFormatter{PrettyPrint: true}},
		{"Text", NewTextFormatter()},
		{"Text-With-Colors", &TextFormatter{EnableColors: true}},
	}

	entry := createBenchmarkEntry()
	defer core.PutEntryToPool(entry)

	for _, tt := range formatters {
		b.Run(tt.name, func(b *testing.B) {
			var buf bytes.Buffer
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf.Reset()
				tt.formatter.Format(&buf, entry)
			}
		})
	}
}

// BenchmarkFormatterWithFields benchmarks formatters with fields
func BenchmarkFormatterWithFields(b *testing.B) {
	formatters := []struct {
		name      string
		formatter Formatter
	}{
		{"CSV", NewCSVFormatter()},
		{"JSON", NewJSONFormatter()},
		{"Text", NewTextFormatter()},
	}

	entry := core.GetEntryFromPool()
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.LevelName = core.StringToBytes("INFO")
	entry.Message = core.StringToBytes("Test message for benchmark")
	entry.Fields = map[string][]byte{
		"user_id":    []byte("123"),
		"action":     []byte("login"),
		"session_id": []byte("sess-456"),
		"source":     []byte("web"),
	}

	for _, tt := range formatters {
		b.Run(tt.name+"_with_fields", func(b *testing.B) {
			var buf bytes.Buffer
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf.Reset()
				tt.formatter.Format(&buf, entry)
			}
		})
	}

	core.PutEntryToPool(entry)
}

// BenchmarkFormatterWithSensitiveData benchmarks formatters with sensitive data masking
func BenchmarkFormatterWithSensitiveData(b *testing.B) {
	csvFormatter := NewCSVFormatter()
	csvFormatter.SensitiveFields = []string{"password", "token", "ssn"}
	csvFormatter.MaskSensitiveData = true
	csvFormatter.MaskStringValue = "[MASKED]"

	jsonFormatter := NewJSONFormatter()
	jsonFormatter.SensitiveFields = []string{"password", "token", "ssn"}
	jsonFormatter.MaskSensitiveData = true
	jsonFormatter.MaskStringValue = "[MASKED]"

	textFormatter := NewTextFormatter()
	textFormatter.SensitiveFields = []string{"password", "token", "ssn"}
	textFormatter.MaskSensitiveData = true
	textFormatter.MaskStringValue = "[MASKED]"

	entry := core.GetEntryFromPool()
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.LevelName = core.StringToBytes("INFO")
	entry.Message = core.StringToBytes("Test message with sensitive data")
	entry.Fields = map[string][]byte{
		"user_id":  []byte("123"),
		"action":   []byte("login"),
		"password": []byte("very_secret_password"),
		"token":    []byte("auth_token_12345"),
		"ssn":      []byte("123-45-6789"),
	}

	formatters := []struct {
		name      string
		formatter Formatter
	}{
		{"CSV-sensitive", csvFormatter},
		{"JSON-sensitive", jsonFormatter},
		{"Text-sensitive", textFormatter},
	}

	for _, tt := range formatters {
		b.Run(tt.name, func(b *testing.B) {
			var buf bytes.Buffer
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf.Reset()
				tt.formatter.Format(&buf, entry)
			}
		})
	}

	core.PutEntryToPool(entry)
}

// createBenchmarkEntry creates a standardized entry for benchmarking
func createBenchmarkEntry() *core.LogEntry {
	entry := core.GetEntryFromPool()
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.LevelName = core.StringToBytes("INFO")
	entry.Message = core.StringToBytes("Test message for benchmark")
	entry.PID = 12345
	entry.TraceID = core.StringToBytes("trace-12345")
	entry.Fields = map[string][]byte{
		"user_id": []byte("123"),
		"action":  []byte("login"),
		"status":  []byte("success"),
	}

	return entry
}

// BenchmarkFormatterConcurrent benchmarks concurrent formatter usage
func BenchmarkFormatterConcurrent(b *testing.B) {
	formatter := NewTextFormatter()
	entry := createBenchmarkEntry()
	defer core.PutEntryToPool(entry)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		for pb.Next() {
			buf.Reset()
			formatter.Format(&buf, entry)
		}
	})
}