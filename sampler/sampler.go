package sampler

import (
	"context"
	"github.com/Lunar-Chipter/mire/core"
	"sync/atomic"
)

// Sampler defines the interface for a logger that can be sampled.
type Sampler interface {
	Log(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte)
}

// LogSampler provides log sampling to reduce volume
type LogSampler struct {
	processor Sampler
	rate      int
	counter   int64
}

// NewSampler creates a new LogSampler
func NewSampler(processor Sampler, rate int) *LogSampler {
	return &LogSampler{
		processor: processor,
		rate:      rate,
	}
}

// ShouldLog determines if a log should be recorded based on sampling rate
func (ls *LogSampler) ShouldLog() bool {
	if ls.rate <=1 {
		return true
	}
	counter := atomic.AddInt64(&ls.counter, 1)
	return counter%int64(ls.rate) == 0
}

// Log logs a message if it passes sampling rate.
func (ls *LogSampler) Log(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte) {
	if ls.ShouldLog() {
		ls.processor.Log(ctx, level, msg, fields)
	}
}

// SamplingLogger provides log sampling to reduce volume
type SamplingLogger struct {
	processor LogSampler
	rate      int
	counter   int64
}

// NewSamplingLogger creates a new SamplingLogger
func NewSampler(processor LogSampler, rate int) *LogSampler {
	return &SamplingLogger{
		processor: processor,
		rate:      rate,
	}
}

// ShouldLog determines if a log should be recorded based on sampling rate
func (sl *LogSampler) ShouldLog() bool {
	if sl.rate <= 1 {
		return true
	}
	counter := atomic.AddInt64(&sl.counter, 1)
	return counter%int64(sl.rate) == 0
}

// Log logs a message if it passes the sampling rate.
func (sl *LogSampler) Log(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte) {
	if sl.ShouldLog() {
		sl.processor.Log(ctx, level, msg, fields)
	}
}
