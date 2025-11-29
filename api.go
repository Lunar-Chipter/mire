// Package api mendefinisikan interface dan kontrak antar modul dalam Mire
package api

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/config"
	"github.com/Lunar-Chipter/mire/formatter"
	"github.com/Lunar-Chipter/mire/hook"
	"github.com/Lunar-Chipter/mire/metric"
	"github.com/Lunar-Chipter/mire/writer"
)

// LoggerInterface adalah interface utama untuk operasi logging
type LoggerInterface interface {
	// Metode logging tingkat dasar
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Notice(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	// Metode logging dengan format
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Noticef(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	// Metode logging dengan context
	TraceC(ctx context.Context, args ...interface{})
	DebugC(ctx context.Context, args ...interface{})
	InfoC(ctx context.Context, args ...interface{})
	NoticeC(ctx context.Context, args ...interface{})
	WarnC(ctx context.Context, args ...interface{})
	ErrorC(ctx context.Context, args ...interface{})
	FatalC(ctx context.Context, args ...interface{})
	PanicC(ctx context.Context, args ...interface{})

	// Metode logging dengan format dan context
	TracefC(ctx context.Context, format string, args ...interface{})
	DebugfC(ctx context.Context, format string, args ...interface{})
	InfofC(ctx context.Context, format string, args ...interface{})
	NoticefC(ctx context.Context, format string, args ...interface{})
	WarnfC(ctx context.Context, format string, args ...interface{})
	ErrorfC(ctx context.Context, format string, args ...interface{})
	FatalfC(ctx context.Context, format string, args ...interface{})
	PanicfC(ctx context.Context, format string, args ...interface{})

	// Metode tambahan
	WithFields(fields map[string]interface{}) *LoggerInterface
	Close()
}

// WriterInterface adalah interface untuk komponen penulisan log
type WriterInterface interface {
	writer.Writer
}

// FormatterInterface adalah interface untuk komponen format log
type FormatterInterface interface {
	formatter.Formatter
}

// HookInterface adalah interface untuk komponen hook
type HookInterface interface {
	hook.Hook
}

// ConfigInterface adalah interface untuk konfigurasi
type ConfigInterface interface {
	GetLevel() core.Level
	GetOutput() io.Writer
	GetFormatter() FormatterInterface
	GetTimestampFormat() string
	GetRotationConfig() *config.RotationConfig
}

// MetricsInterface adalah interface untuk kolektor metrik
type MetricsInterface interface {
	metric.MetricsCollector
}

// EntryBuilderInterface adalah interface untuk pembuatan log entry
type EntryBuilderInterface interface {
	BuildEntry(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) *core.LogEntry
}

// ProcessorInterface adalah interface untuk pemrosesan log
type ProcessorInterface interface {
	ProcessLog(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) error
}

// PoolInterface adalah interface untuk manajemen pool
type PoolInterface interface {
	GetEntryFromPool() *core.LogEntry
	PutEntryToPool(entry *core.LogEntry)
	GetBufferFromPool() interface{} // Using interface{} to avoid direct dependency
	PutBufferToPool(buf interface{})
}

// ClockInterface adalah interface untuk manajemen waktu
type ClockInterface interface {
	Now() time.Time
	Stop()
}