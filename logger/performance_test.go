package logger

import (
	"context"
	"io"
	"testing"
	"time"

	"mire/core"
	"mire/formatter"
)

// TestMemoryAllocations tests memory allocations for different log levels
func TestMemoryAllocations(t *testing.T) {
	t.Run("TraceLevelAllocations", func(t *testing.T) {
		allocs := testing.AllocsPerRun(1000, func() {
			logger := New(LoggerConfig{
				Level:  core.TRACE,
				Output: io.Discard,
				Formatter: &formatter.TextFormatter{
					EnableColors:    false,
					ShowTimestamp:   false,
					ShowCaller:      false,
				},
			})
			logger.Trace("test message")
			logger.Close()
		})
		t.Logf("Allocations per Trace call: %.2f", allocs)
	})

	t.Run("DebugLevelAllocations", func(t *testing.T) {
		allocs := testing.AllocsPerRun(1000, func() {
			logger := New(LoggerConfig{
				Level:  core.DEBUG,
				Output: io.Discard,
				Formatter: &formatter.TextFormatter{
					EnableColors:    false,
					ShowTimestamp:   false,
					ShowCaller:      false,
				},
			})
			logger.Debug("test message")
			logger.Close()
		})
		t.Logf("Allocations per Debug call: %.2f", allocs)
	})

	t.Run("InfoLevelAllocations", func(t *testing.T) {
		allocs := testing.AllocsPerRun(1000, func() {
			logger := New(LoggerConfig{
				Level:  core.INFO,
				Output: io.Discard,
				Formatter: &formatter.TextFormatter{
					EnableColors:    false,
					ShowTimestamp:   false,
					ShowCaller:      false,
				},
			})
			logger.Info("test message")
			logger.Close()
		})
		t.Logf("Allocations per Info call: %.2f", allocs)
	})

	t.Run("ErrorLevelAllocations", func(t *testing.T) {
		allocs := testing.AllocsPerRun(1000, func() {
			logger := New(LoggerConfig{
				Level:  core.ERROR,
				Output: io.Discard,
				Formatter: &formatter.TextFormatter{
					EnableColors:    false,
					ShowTimestamp:   false,
					ShowCaller:      false,
				},
			})
			logger.Error("test message")
			logger.Close()
		})
		t.Logf("Allocations per Error call: %.2f", allocs)
	})
}

// TestThroughput benchmarks the throughput of the logger
func TestThroughput(t *testing.T) {
	t.Run("ThroughputWithoutFields", func(t *testing.T) {
		logger := NewDefaultLogger()
		defer logger.Close()

		start := time.Now()
		n := 100000
		for i := 0; i < n; i++ {
			logger.Info("test message")
		}
		duration := time.Since(start)

		opsPerSec := float64(n) / duration.Seconds()
		t.Logf("Throughput without fields: %.0f ops/sec", opsPerSec)
		t.Logf("Time per operation: %v", duration/time.Duration(n))
	})

	t.Run("ThroughputWithFields", func(t *testing.T) {
		logger := NewDefaultLogger()
		defer logger.Close()

		start := time.Now()
		n := 100000
		for i := 0; i < n; i++ {
			logger.WithFields(map[string]interface{}{
				"field1": "value1",
				"field2": i,
				"field3": true,
			}).Info("test message")
		}
		duration := time.Since(start)

		opsPerSec := float64(n) / duration.Seconds()
		t.Logf("Throughput with fields: %.0f ops/sec", opsPerSec)
		t.Logf("Time per operation: %v", duration/time.Duration(n))
	})
}

// BenchmarkMemoryAllocations benchmarks memory allocations for different log levels
func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("BenchmarkTraceAlloc", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.TRACE,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Trace("test message")
		}
	})

	b.Run("BenchmarkDebugAlloc", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.DEBUG,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Debug("test message")
		}
	})

	b.Run("BenchmarkInfoAlloc", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})

	b.Run("BenchmarkErrorAlloc", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.ERROR,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Error("test message")
		}
	})
}

// BenchmarkFormatterAllocations compares memory allocations between formatters
func BenchmarkFormatterAllocations(b *testing.B) {
	b.Run("BenchmarkTextFormatterAllocs", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false, // Disable timestamp for benchmark consistency
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})

	b.Run("BenchmarkJSONFormatterAllocs", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.JSONFormatter{
				PrettyPrint: false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})
}

// BenchmarkFieldAllocations tests allocations with different numbers of fields
func BenchmarkFieldAllocations(b *testing.B) {
	b.Run("BenchmarkWithNoFields", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})

	b.Run("BenchmarkWithOneField", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.WithFields(map[string]interface{}{
				"field1": "value1",
			}).Info("test message")
		}
	})

	b.Run("BenchmarkWithFiveFields", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.WithFields(map[string]interface{}{
				"field1": "value1",
				"field2": i,
				"field3": true,
				"field4": 3.14,
				"field5": []string{"a", "b", "c"},
			}).Info("test message")
		}
	})

	b.Run("BenchmarkWithTenFields", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.WithFields(map[string]interface{}{
				"field1":  "value1",
				"field2":  i,
				"field3":  true,
				"field4":  3.14,
				"field5":  []string{"a", "b", "c"},
				"field6":  "value6",
				"field7":  42,
				"field8":  false,
				"field9":  2.71,
				"field10": []int{1, 2, 3, 4, 5},
			}).Info("test message")
		}
	})
}

// BenchmarkThroughput benchmarks the throughput of the logger
func BenchmarkThroughput(b *testing.B) {
	b.Run("BenchmarkThroughputNoFields", func(b *testing.B) {
		logger := NewDefaultLogger()
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})

	b.Run("BenchmarkThroughputWithFields", func(b *testing.B) {
		logger := NewDefaultLogger()
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.WithFields(map[string]interface{}{
				"field1": "value1",
				"field2": i,
				"field3": true,
			}).Info("test message")
		}
	})

	b.Run("BenchmarkThroughputWithFieldsAndFormatting", func(b *testing.B) {
		logger := NewDefaultLogger()
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.WithFields(map[string]interface{}{
				"field1": "value1",
				"field2": i,
				"field3": true,
			}).Infof("test message with %d", i)
		}
	})
}

// BenchmarkThroughputLevels tests throughput at different log levels
func BenchmarkThroughputLevels(b *testing.B) {
	levels := []struct {
		name  string
		level core.Level
		logFn func(*Logger, string)
	}{
		{"Trace", core.TRACE, func(l *Logger, msg string) { l.Trace(msg) }},
		{"Debug", core.DEBUG, func(l *Logger, msg string) { l.Debug(msg) }},
		{"Info", core.INFO, func(l *Logger, msg string) { l.Info(msg) }},
		{"Warn", core.WARN, func(l *Logger, msg string) { l.Warn(msg) }},
		{"Error", core.ERROR, func(l *Logger, msg string) { l.Error(msg) }},
	}

	for _, lvl := range levels {
		b.Run("BenchmarkThroughput"+lvl.name, func(b *testing.B) {
			logger := New(LoggerConfig{
				Level:  core.TRACE, // Set to lowest level to ensure all logs pass through
				Output: io.Discard,
				Formatter: &formatter.TextFormatter{
					EnableColors:    false,
					ShowTimestamp:   false,
					ShowCaller:      false,
				},
			})
			defer logger.Close()

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				lvl.logFn(logger, "test message")
			}
		})
	}
}

// BenchmarkThroughputFormatters compares throughput of different formatters
func BenchmarkThroughputFormatters(b *testing.B) {
	formatters := []struct {
		name string
		fmt  formatter.Formatter
	}{
		{"TextFormatter", &formatter.TextFormatter{EnableColors: false, ShowTimestamp: false, ShowCaller: false}},
		{"TextFormatterWithTS", &formatter.TextFormatter{EnableColors: false, ShowTimestamp: true, ShowCaller: false}},
		{"TextFormatterWithTSAndCaller", &formatter.TextFormatter{EnableColors: false, ShowTimestamp: true, ShowCaller: true}},
		{"JSONFormatter", &formatter.JSONFormatter{PrettyPrint: false}},
		{"JSONFormatterPretty", &formatter.JSONFormatter{PrettyPrint: true}},
	}

	for _, f := range formatters {
		b.Run("BenchmarkThroughput"+f.name, func(b *testing.B) {
			logger := New(LoggerConfig{
				Level:     core.INFO,
				Output:    io.Discard,
				Formatter: f.fmt,
			})
			defer logger.Close()

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				logger.Info("test message")
			}
		})
	}
}

// BenchmarkFormatters compares different formatters
func BenchmarkFormatters(b *testing.B) {
	b.Run("BenchmarkTextFormatter", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   true,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})

	b.Run("BenchmarkJSONFormatter", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.JSONFormatter{
				PrettyPrint: false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})
}

// BenchmarkAsyncLogging tests the async logging performance
func BenchmarkAsyncLogging(b *testing.B) {
	b.Run("BenchmarkAsyncLogging", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:             core.INFO,
			Output:            io.Discard,
			AsyncLogging:      true,
			AsyncWorkerCount:  4,
			AsyncLogChannelBufferSize: 1000,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   true,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})

	b.Run("BenchmarkSyncLogging", func(b *testing.B) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   true,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})
}

// BenchmarkContextLogging tests performance with context
func BenchmarkContextLogging(b *testing.B) {
	ctx := context.Background()

	b.Run("BenchmarkContextLogging", func(b *testing.B) {
		logger := NewDefaultLogger()
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.InfoC(ctx, "test message with context")
		}
	})

	b.Run("BenchmarkStandardLogging", func(b *testing.B) {
		logger := NewDefaultLogger()
		defer logger.Close()

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			logger.Info("test message without context")
		}
	})
}

// TestLogEntryPoolEffectiveness tests the effectiveness of the object pool
func TestLogEntryPoolEffectiveness(t *testing.T) {
	// First, create many log entries to populate the pool
	logger := NewDefaultLogger()
	for i := 0; i < 10000; i++ {
		logger.Info("initial message")
	}
	logger.Close()

	// Now test allocations with a warmed-up pool
	allocs := testing.AllocsPerRun(1000, func() {
		logger := NewDefaultLogger()
		logger.Info("test message")
		logger.Close()
	})

	t.Logf("Allocations per log with warmed pool: %.2f", allocs)

	// Compare with initial metrics to show pool effectiveness
	initialMetrics := core.GetEntryMetrics()
	t.Logf("Entry metrics - Created: %d, Reused: %d, Pool Misses: %d",
		initialMetrics.CreatedCount(),
		initialMetrics.ReusedCount(),
		initialMetrics.PoolMissCount())
}

// TestConcurrentLoggingPerformance tests performance under concurrent load
func TestConcurrentLoggingPerformance(t *testing.T) {
	logger := NewDefaultLogger()
	defer logger.Close()

	// Run concurrent logging from multiple goroutines
	done := make(chan bool, 10)
	start := time.Now()

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 1000; j++ {
				logger.WithFields(map[string]interface{}{
					"goroutine": id,
					"iteration": j,
				}).Info("concurrent log message")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	duration := time.Since(start)
	t.Logf("Concurrent logging (10 goroutines, 1000 messages each): %v", duration)
}

// TestBufferedWriterPerformance tests performance with buffered writer
func TestBufferedWriterPerformance(t *testing.T) {
	t.Run("WithoutBuffer", func(t *testing.T) {
		logger := New(LoggerConfig{
			Level:  core.INFO,
			Output: io.Discard,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		start := time.Now()
		for i := 0; i < 10000; i++ {
			logger.Info("test message")
		}
		duration := time.Since(start)

		t.Logf("Without buffer: %v for 10000 messages", duration)
	})

	t.Run("WithBuffer", func(t *testing.T) {
		logger := New(LoggerConfig{
			Level:      core.INFO,
			Output:     io.Discard,
			BufferSize: 1000,
			Formatter: &formatter.TextFormatter{
				EnableColors:    false,
				ShowTimestamp:   false,
				ShowCaller:      false,
			},
		})
		defer logger.Close()

		start := time.Now()
		for i := 0; i < 10000; i++ {
			logger.Info("test message")
		}
		// Close logger to flush buffer
		logger.Close()
		duration := time.Since(start)

		t.Logf("With buffer: %v for 10000 messages", duration)
	})
}

// TestFormatterPerformance compares performance of different formatters
func TestFormatterPerformance(t *testing.T) {
	formatters := []struct {
		name string
		fmt  formatter.Formatter
	}{
		{"TextFormatter", &formatter.TextFormatter{EnableColors: false, ShowTimestamp: false, ShowCaller: false}},
		{"TextFormatterWithTS", &formatter.TextFormatter{EnableColors: false, ShowTimestamp: true, ShowCaller: false}},
		{"JSONFormatter", &formatter.JSONFormatter{PrettyPrint: false}},
	}

	for _, f := range formatters {
		t.Run("Performance"+f.name, func(t *testing.T) {
			logger := New(LoggerConfig{
				Level:     core.INFO,
				Output:    io.Discard,
				Formatter: f.fmt,
			})
			defer logger.Close()

			start := time.Now()
			n := 10000
			for i := 0; i < n; i++ {
				logger.Info("test message")
			}
			duration := time.Since(start)

			opsPerSec := float64(n) / duration.Seconds()
			t.Logf("%s: %.0f ops/sec", f.name, opsPerSec)
		})
	}
}

// Example usage for documentation
func ExampleLogger_performance() {
	// Create a high-performance logger configuration
	logger := New(LoggerConfig{
		Level:             core.INFO,
		Output:            io.Discard, // Use io.Discard for benchmarks
		AsyncLogging:      true,       // Enable async logging
		AsyncWorkerCount:  4,          // Use 4 async workers
		AsyncLogChannelBufferSize: 1000, // Buffer up to 1000 log messages
		Formatter: &formatter.TextFormatter{
			EnableColors:    false,      // Disable colors for performance
			ShowTimestamp:   true,       // Include timestamps
			ShowCaller:      false,      // Disable caller info for performance
		},
	})

	// Use the logger
	logger.Info("This is a high-performance log message")

	// Always close the logger when done
	logger.Close()

	// Output:
}