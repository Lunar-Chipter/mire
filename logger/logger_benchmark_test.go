package logger

import (
	"context"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/util"
)

// BenchmarkBasicLogging benchmarks basic logging operations
func BenchmarkBasicLogging(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.TextFormatter{
			EnableColors:     false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message for benchmark")
	}
}

// at
func BenchmarkLoggingWithFields(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.TextFormatter{
			EnableColors:     false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(map[string]interface{}{
			"user_id": i,
			"action":  "test_action",
			"value":   i * 10,
		}).Info("test message with fields")
	}
}

// BenchmarkContextAwareLogging benchmarks context-aware logging
func BenchmarkContextAwareLogging(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.TextFormatter{
			EnableColors:     false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
	})
	defer logger.Close()

	ctx := context.Background()
	ctx = util.WithTraceID(ctx, "trace-12345")
	ctx = util.WithUserID(ctx, "user-67890")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.InfoC(ctx, "test message with context")
	}
}

// BenchmarkAsyncLoggingOps benchmarks asynchronous logging operations
func BenchmarkAsyncLoggingOps(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.TextFormatter{
			EnableColors:     false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		AsyncMode:      true,
		WorkerCount:    4,
		ChannelSize:    10000,
		ProcessTimeout: 5 * time.Second,
		NoTimeout:      true,
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message for async benchmark")
	}
}

// BenchmarkConcurrentLogging benchmarks concurrent logging operations
func BenchmarkConcurrentLogging(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.TextFormatter{
			EnableColors:     false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
	})
	defer logger.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.WithFields(map[string]interface{}{
				"goroutine_id": pb,
				"iteration":    time.Now().UnixNano(),
			}).Info("concurrent test message")
		}
	})
}

// BenchmarkSamplingLogging benchmarks logging with sampling
func BenchmarkSamplingLogging(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.TextFormatter{
			EnableColors:     false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
		EnableSampling: true,
		SamplingRate:   5, // Log every 5th message
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message with sampling", i)
	}
}

// BenchmarkFormattedLogging benchmarks logging with text formatter
func BenchmarkFormattedLogging(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.TextFormatter{
			EnableColors:       false,
			ShowTimestamp:   true,
			ShowCaller:      true,
			ShowGoroutine:   true,
			ShowPID:         true,
			ShowTraceInfo:   true,
			ShowHostname:    true,
			ShowApplication: true,
		},
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(map[string]interface{}{
			"user_id": i,
			"action":  "formatted_action",
			"value":   i * 10,
		}).Info("formatted test message")
	}
}

// BenchmarkJSONLogging benchmarks logging with JSON formatter
func BenchmarkJSONLogging(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.JSONFormatter{
			PrettyPrint:       false,
			TimestampFormat:   "2006-01-02 15:04:05.000",
			ShowCaller:        true,
			ShowDuration:      false,
			ShowTrace:         true,
			IncludeStackTrace: false,
		},
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(map[string]interface{}{
			"user_id": i,
			"action":  "json_action",
			"value":   i * 10,
		}).Info("JSON formatted test message")
	}
}

// BenchmarkCSVLogging benchmarks logging with CSV formatter
func BenchmarkCSVLogging(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.CSVFormatter{
			IncludeHeader: false,
			FieldOrder:    []string{"timestamp", "level_name", "message", "user_id", "action"},
		},
	})
	defer logger.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(map[string]interface{}{
			"user_id": i,
			"action":  "csv_action",
		}).Info("CSV formatted test message")
	}
}

// at
func BenchmarkMultipleLevels(b *testing.B) {
	logger := New(LoggerConfig{
		Level:  core.INFO,
		Output: &testWriteSyncer{},
		Formatter: &formatter.TextFormatter{
			EnableColors:     false,
			ShowTimestamp: false,
			ShowCaller:    false,
		},
	})
	defer logger.Close()

	levels := []func(args ...interface{}){
		logger.Trace, logger.Debug, logger.Info,
		logger.Warn, logger.Error,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		level := levels[i%len(levels)]
		level("test message at different levels", i)
	}
}

// testWriteSyncer is a mock WriteSyncer for testing
type testWriteSyncer struct{}

func (tws *testWriteSyncer) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (tws *testWriteSyncer) Sync() error {
	return nil
}
