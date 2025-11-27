package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"mire/core"
	"mire/formatter"
	"mire/logger"
)

// BenchmarkTextFormatterNoAlloc measures allocation for text formatter
func BenchmarkTextFormatterNoAlloc(b *testing.B) {
	f := &formatter.TextFormatter{
		EnableColors:    false,
		ShowTimestamp:   false,
		ShowCaller:      false,
	}
	
	entry := core.GetEntryFromPool()
	entry.Message = core.S2b("Test message for benchmark")
	entry.Level = core.INFO
	entry.LevelName = core.INFO.String()
	defer core.PutEntryToPool(entry)
	
	var buf bytes.Buffer
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		f.Format(&buf, entry)
	}
}

// BenchmarkJSONFormatterNoAlloc measures allocation for JSON formatter (optimized version)
func BenchmarkJSONFormatterNoAlloc(b *testing.B) {
	f := &formatter.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	}
	
	entry := core.GetEntryFromPool()
	entry.Message = core.S2b("Test message for benchmark")
	entry.Level = core.INFO
	entry.LevelName = core.INFO.String()
	defer core.PutEntryToPool(entry)
	
	var buf bytes.Buffer
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		f.Format(&buf, entry)  // This will use our optimized manual formatter
	}
}

// BenchmarkLoggerInfoNoAlloc measures allocation for logger Info calls
func BenchmarkLoggerInfoNoAlloc(b *testing.B) {
	cfg := logger.LoggerConfig{
		Level:   core.INFO,
		Output:  io.Discard, // Discard output to measure only logging overhead
		Formatter: &formatter.TextFormatter{
			EnableColors:    false,
			ShowTimestamp:   false,
			ShowCaller:      false,
		},
	}
	log := logger.New(cfg)
	defer log.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("Benchmark message")
	}
}

// BenchmarkLoggerWithFieldsNoAlloc measures allocation for logger with fields
func BenchmarkLoggerWithFieldsNoAlloc(b *testing.B) {
	cfg := logger.LoggerConfig{
		Level:   core.INFO,
		Output:  io.Discard,
		Formatter: &formatter.TextFormatter{
			EnableColors:    false,
			ShowTimestamp:   false,
			ShowCaller:      false,
		},
	}
	log := logger.New(cfg)
	defer log.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.WithFields(map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}).Info("Benchmark message with fields")
	}
}

// BenchmarkLoggerJSONFileNoAlloc measures allocation for JSON file logging
func BenchmarkLoggerJSONFileNoAlloc(b *testing.B) {
	// Create a temporary file for benchmark
	tmpFile, err := os.CreateTemp("", "benchmark_*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()
	
	cfg := logger.LoggerConfig{
		Level:   core.INFO,
		Output:  io.Discard, // Use io.Discard to avoid actual file I/O in benchmark
		Formatter: &formatter.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		},
	}
	log := logger.New(cfg)
	defer log.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.WithFields(map[string]interface{}{
			"timestamp": "2023-01-01T00:00:00.000Z",
			"level":     "INFO",
			"message":   "Benchmark JSON message",
		}).Info("Benchmark message")
	}
}

// Example output expected:
// BenchmarkTextFormatterNoAlloc-8         	10000000	       150 ns/op	       0 B/op	       0 allocs/op
// BenchmarkJSONFormatterNoAlloc-8         	 5000000	       300 ns/op	       0 B/op	       0 allocs/op  
// BenchmarkLoggerInfoNoAlloc-8            	 2000000	       800 ns/op	      32 B/op	       1 allocs/op
// BenchmarkLoggerWithFieldsNoAlloc-8      	 1000000	      1500 ns/op	      64 B/op	       2 allocs/op
// BenchmarkLoggerJSONFileNoAlloc-8        	 1000000	      2000 ns/op	      48 B/op	       1 allocs/op